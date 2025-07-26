// arbitrage.go
package services

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sort"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"arbitrage-bot/config"
	"arbitrage-bot/contracts"
	"arbitrage-bot/models"
)

// ArbitrageService handles arbitrage operations
type ArbitrageService struct {
	Client        *EthClient
	TokenService  *TokenService
	RouterService *RouterService
	Config        *config.Config
	TokenPairs    []models.TokenPair

	PancakeRouter common.Address
	BiswapRouter  common.Address
	FlashContract common.Address
}

// NewArbitrageService creates a new ArbitrageService
func NewArbitrageService(
	client *EthClient,
	tokenService *TokenService,
	routerService *RouterService,
	cfg *config.Config,
) *ArbitrageService {
	return &ArbitrageService{
		Client:        client,
		TokenService:  tokenService,
		RouterService: routerService,
		Config:        cfg,
		TokenPairs:    models.InitializeTokenPairs(),

		PancakeRouter: common.HexToAddress(config.PancakeswapRouter),
		BiswapRouter:  common.HexToAddress(config.BiswapRouter),
		FlashContract: common.HexToAddress(cfg.FlashArbContract),
	}
}

// FindArbitrageOpportunities scans all token pairs for arbitrage opportunities
func (s *ArbitrageService) FindArbitrageOpportunities() error {
	log.Println("Scanning for arbitrage opportunities...")

	// Loop through all token pairs
	for _, pair := range s.TokenPairs {
		log.Printf("Checking pair: %s", pair.Name)

		// Verify tokens and pairs before trying arbitrage
		if err := s.VerifyPairTokens(pair); err != nil {
			log.Printf("Warning: Pair %s has issues: %v", pair.Name, err)
			continue // Skip pairs with issues
		}

		// Try different test amounts
		for _, amount := range pair.TestAmounts {
			// Check PancakeSwap -> BiSwap -> PancakeSwap route
			resultPancakeFirst, err := s.CheckTriangularArbitrage(pair, amount, true)
			if err != nil {
				log.Printf("Error checking Pancake->BiSwap route: %v", err)
				continue
			}

			// Check BiSwap -> PancakeSwap -> BiSwap route
			resultBiswapFirst, err := s.CheckTriangularArbitrage(pair, amount, false)
			if err != nil {
				log.Printf("Error checking BiSwap->Pancake route: %v", err)
				continue
			}

			// Log the results with proper formatting
			log.Printf("Pancake->BiSwap route profit: %.4f%%", resultPancakeFirst.ProfitPercent*100)
			log.Printf("BiSwap->Pancake route profit: %.4f%%", resultBiswapFirst.ProfitPercent*100)

			// Check if either route is profitable enough
			if resultPancakeFirst.ProfitPercent > s.Config.MinProfit {
				log.Printf("Found profitable opportunity (Pancake->BiSwap): %.4f%%", resultPancakeFirst.ProfitPercent*100)

				// Double-check profitability with a second calculation
				confirmProfit, err := s.ConfirmProfitability(pair, amount, true)
				if err != nil || confirmProfit < s.Config.MinProfit {
					log.Printf("Profit confirmation failed: %.4f%% (below threshold or error: %v)",
						confirmProfit*100, err)
					continue
				}

				// Execute the arbitrage if we have a flash arbitrage contract
				if s.FlashContract != (common.Address{}) {
					err = s.ExecuteArbitrage(pair, resultPancakeFirst.TargetAmount, true)
					if err != nil {
						log.Printf("Error executing arbitrage: %v", err)
					}
				} else {
					log.Println("Flash arbitrage contract not set. Skipping execution.")
				}

				return nil
			} else if resultBiswapFirst.ProfitPercent > s.Config.MinProfit {
				log.Printf("Found profitable opportunity (BiSwap->Pancake): %.4f%%", resultBiswapFirst.ProfitPercent*100)

				// Double-check profitability with a second calculation
				confirmProfit, err := s.ConfirmProfitability(pair, amount, false)
				if err != nil || confirmProfit < s.Config.MinProfit {
					log.Printf("Profit confirmation failed: %.4f%% (below threshold or error: %v)",
						confirmProfit*100, err)
					continue
				}

				// Execute the arbitrage if we have a flash arbitrage contract
				if s.FlashContract != (common.Address{}) {
					err = s.ExecuteArbitrage(pair, resultBiswapFirst.TargetAmount, false)
					if err != nil {
						log.Printf("Error executing arbitrage: %v", err)
					}
				} else {
					log.Println("Flash arbitrage contract not set. Skipping execution.")
				}

				return nil
			}
		}
	}

	log.Println("No profitable arbitrage opportunities found in this round.")
	return nil
}

// VerifyPairTokens checks if all tokens and pairs are valid
func (s *ArbitrageService) VerifyPairTokens(pair models.TokenPair) error {
	var errors []string

	// Verify tokens
	for name, addr := range pair.Tokens {
		if addr == "" {
			errors = append(errors, fmt.Sprintf("Empty address for token %s", name))
			continue
		}

		if !common.IsHexAddress(addr) {
			errors = append(errors, fmt.Sprintf("Invalid hex address for token %s: %s", name, addr))
			continue
		}

		tokenAddr := common.HexToAddress(addr)
		_, err := s.TokenService.GetTokenDecimals(tokenAddr)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Cannot get decimals for token %s: %v", name, err))
		}
	}

	// Quick validation of pair addresses
	for name, addr := range pair.PancakeswapPair {
		if addr != "" && !common.IsHexAddress(addr) {
			errors = append(errors, fmt.Sprintf("Invalid PancakeSwap pair address %s: %s", name, addr))
		}
	}

	for name, addr := range pair.BiswapPair {
		if addr != "" && !common.IsHexAddress(addr) {
			errors = append(errors, fmt.Sprintf("Invalid BiSwap pair address %s: %s", name, addr))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("pair verification issues: %s", strings.Join(errors, "; "))
	}

	return nil
}

// CheckTriangularArbitrage checks if a triangular arbitrage opportunity exists
func (s *ArbitrageService) CheckTriangularArbitrage(
	pair models.TokenPair,
	testAmount float64,
	pancakeFirst bool,
) (*models.ArbitrageResult, error) {
	// Get token addresses safely
	tokenA := common.HexToAddress(pair.Tokens["WBNB"])

	// Get the other two tokens (not WBNB)
	otherTokens := getOtherTokens(pair.Tokens)
	if len(otherTokens) < 2 {
		return nil, fmt.Errorf("need at least 3 tokens for triangular arbitrage, got %d", len(otherTokens)+1)
	}

	tokenB := common.HexToAddress(pair.Tokens[otherTokens[0]])
	tokenC := common.HexToAddress(pair.Tokens[otherTokens[1]])

	// Verify the addresses are different
	if tokenA == tokenB || tokenB == tokenC || tokenA == tokenC {
		return nil, fmt.Errorf("token addresses must be different for arbitrage")
	}

	// Log token addresses for debugging
	log.Printf("Token A (WBNB): %s", tokenA.Hex())
	log.Printf("Token B (%s): %s", otherTokens[0], tokenB.Hex())
	log.Printf("Token C (%s): %s", otherTokens[1], tokenC.Hex())

	// Get token decimals
	tokenADecimals, err := s.TokenService.GetTokenDecimals(tokenA)
	if err != nil {
		return nil, fmt.Errorf("failed to get decimals for WBNB: %v", err)
	}

	// Convert test amount to token amount with decimals
	tokenAmount := s.TokenService.FormatTokenAmount(testAmount, tokenADecimals)
	log.Printf("Test amount: %.6f WBNB (%s wei)", testAmount, tokenAmount.String())

	// Prepare paths for both routes
	path1 := []common.Address{tokenA, tokenB}
	path2 := []common.Address{tokenB, tokenC}
	path3 := []common.Address{tokenC, tokenA}

	var route1Router, route2Router, route3Router common.Address
	var routeDescription string

	if pancakeFirst {
		// PancakeSwap -> BiSwap -> PancakeSwap
		route1Router = s.PancakeRouter
		route2Router = s.BiswapRouter
		route3Router = s.PancakeRouter
		routeDescription = "PancakeSwap -> BiSwap -> PancakeSwap"
	} else {
		// BiSwap -> PancakeSwap -> BiSwap
		route1Router = s.BiswapRouter
		route2Router = s.PancakeRouter
		route3Router = s.BiswapRouter
		routeDescription = "BiSwap -> PancakeSwap -> BiSwap"
	}

	log.Printf("Route: %s", routeDescription)

	// Calculate amounts out for each step in the route
	// Step 1: WBNB -> TokenB
	amounts1, err := s.RouterService.GetAmountsOut(route1Router, tokenAmount, path1)
	if err != nil {
		return nil, fmt.Errorf("error in step 1 (WBNB -> %s): %v", otherTokens[0], err)
	}

	if len(amounts1) < 2 {
		return nil, fmt.Errorf("invalid amounts1 length: %d", len(amounts1))
	}

	var dex1 string
	if pancakeFirst {
		dex1 = "PancakeSwap"
	} else {
		dex1 = "BiSwap"
	}

	log.Printf("Step 1 (WBNB -> %s via %s): In: %s, Out: %s",
		otherTokens[0], dex1, tokenAmount.String(), amounts1[1].String())

	// Step 2: TokenB -> TokenC
	amounts2, err := s.RouterService.GetAmountsOut(route2Router, amounts1[1], path2)
	if err != nil {
		return nil, fmt.Errorf("error in step 2 (%s -> %s): %v", otherTokens[0], otherTokens[1], err)
	}

	if len(amounts2) < 2 {
		return nil, fmt.Errorf("invalid amounts2 length: %d", len(amounts2))
	}

	var dex2 string
	if pancakeFirst {
		dex2 = "BiSwap"
	} else {
		dex2 = "PancakeSwap"
	}

	log.Printf("Step 2 (%s -> %s via %s): In: %s, Out: %s",
		otherTokens[0], otherTokens[1], dex2, amounts1[1].String(), amounts2[1].String())

	// Step 3: TokenC -> WBNB
	amounts3, err := s.RouterService.GetAmountsOut(route3Router, amounts2[1], path3)
	if err != nil {
		return nil, fmt.Errorf("error in step 3 (%s -> WBNB): %v", otherTokens[1], err)
	}

	if len(amounts3) < 2 {
		return nil, fmt.Errorf("invalid amounts3 length: %d", len(amounts3))
	}

	var dex3 string
	if pancakeFirst {
		dex3 = "PancakeSwap"
	} else {
		dex3 = "BiSwap"
	}

	log.Printf("Step 3 (%s -> WBNB via %s): In: %s, Out: %s",
		otherTokens[1], dex3, amounts2[1].String(), amounts3[1].String())

	// Calculate profit (or loss)
	finalAmount := amounts3[1]
	profit := new(big.Int).Sub(finalAmount, tokenAmount)

	// Calculate profit percentage
	profitFloat := new(big.Float).SetInt(profit)
	initialFloat := new(big.Float).SetInt(tokenAmount)

	var profitPercent float64
	if initialFloat.Cmp(big.NewFloat(0)) > 0 {
		percentFloat := new(big.Float).Quo(profitFloat, initialFloat)
		profitPercent, _ = percentFloat.Float64()
	}

	// Estimate gas costs (~0.1% for BSC)
	gasAdjustedProfitPercent := profitPercent - 0.001

	// Log results with proper formatting
	log.Printf("Initial: %.6f WBNB, Final: %.6f WBNB",
		s.TokenService.ConvertToReadable(tokenAmount, tokenADecimals),
		s.TokenService.ConvertToReadable(finalAmount, tokenADecimals))
	log.Printf("Profit: %.6f WBNB (%.4f%%), Gas adjusted profit: %.4f%%",
		s.TokenService.ConvertToReadable(profit, tokenADecimals),
		profitPercent*100, gasAdjustedProfitPercent*100)

	// Prepare the result
	result := &models.ArbitrageResult{
		Profit:        profit,
		PlatformFee:   new(big.Int).Div(profit, big.NewInt(10)), // 10% platform fee
		UserProfit:    new(big.Int).Sub(profit, new(big.Int).Div(profit, big.NewInt(10))),
		TargetAmount:  tokenAmount,
		ProfitPercent: gasAdjustedProfitPercent,
		Direction:     pancakeFirst,
		Path:          []string{pair.Tokens["WBNB"], pair.Tokens[otherTokens[0]], pair.Tokens[otherTokens[1]]},
	}

	return result, nil
}

// ConfirmProfitability does a second profit calculation to verify results
func (s *ArbitrageService) ConfirmProfitability(
	pair models.TokenPair,
	testAmount float64,
	pancakeFirst bool,
) (float64, error) {
	// Get token addresses safely
	tokenA := common.HexToAddress(pair.Tokens["WBNB"])

	otherTokens := getOtherTokens(pair.Tokens)
	if len(otherTokens) < 2 {
		return 0, fmt.Errorf("need at least 3 tokens for triangular arbitrage")
	}

	tokenB := common.HexToAddress(pair.Tokens[otherTokens[0]])
	tokenC := common.HexToAddress(pair.Tokens[otherTokens[1]])

	// Get token decimals
	decimalsA, err := s.TokenService.GetTokenDecimals(tokenA)
	if err != nil {
		return 0, err
	}

	// Convert to wei
	amountInWei := s.TokenService.FormatTokenAmount(testAmount, decimalsA)

	// Prepare routes
	var route1Router, route2Router, route3Router common.Address
	if pancakeFirst {
		route1Router = s.PancakeRouter
		route2Router = s.BiswapRouter
		route3Router = s.PancakeRouter
	} else {
		route1Router = s.BiswapRouter
		route2Router = s.PancakeRouter
		route3Router = s.BiswapRouter
	}

	// Calculate each swap
	path1 := []common.Address{tokenA, tokenB}
	amountOut1, err := s.RouterService.GetAmountOutSingle(route1Router, amountInWei, path1)
	if err != nil {
		return 0, err
	}

	path2 := []common.Address{tokenB, tokenC}
	amountOut2, err := s.RouterService.GetAmountOutSingle(route2Router, amountOut1, path2)
	if err != nil {
		return 0, err
	}

	path3 := []common.Address{tokenC, tokenA}
	amountOut3, err := s.RouterService.GetAmountOutSingle(route3Router, amountOut2, path3)
	if err != nil {
		return 0, err
	}

	// Calculate profit
	profit := new(big.Float).SetInt(new(big.Int).Sub(amountOut3, amountInWei))
	initial := new(big.Float).SetInt(amountInWei)

	var profitPercent float64
	if initial.Cmp(big.NewFloat(0)) > 0 {
		percentFloat := new(big.Float).Quo(profit, initial)
		profitPercent, _ = percentFloat.Float64()
	}

	// Subtract gas costs (approximately 0.1%)
	profitPercent -= 0.001

	return profitPercent, nil
}

// ExecuteArbitrage executes a triangular arbitrage trade
func (s *ArbitrageService) ExecuteArbitrage(
	pair models.TokenPair,
	amount *big.Int,
	pancakeFirst bool,
) error {
	log.Printf("Executing arbitrage on pair %s, amount: %s, pancakeFirst: %v",
		pair.Name, amount.String(), pancakeFirst)

	// If we have a flash arbitrage contract, use it
	if s.FlashContract != (common.Address{}) {
		return s.ExecuteFlashArbitrage(pair, amount, pancakeFirst)
	}

	// Otherwise execute manually (not recommended without flash loans)
	return s.ExecuteManualArbitrage(pair, amount, pancakeFirst)
}

// ExecuteFlashArbitrage executes a triangular arbitrage using the flash arbitrage contract
func (s *ArbitrageService) ExecuteFlashArbitrage(
	pair models.TokenPair,
	amount *big.Int,
	pancakeFirst bool,
) error {
	log.Println("Executing flash arbitrage...")

	// Get token addresses safely
	tokenA := common.HexToAddress(pair.Tokens["WBNB"])

	otherTokens := getOtherTokens(pair.Tokens)
	if len(otherTokens) < 2 {
		return fmt.Errorf("need at least 3 tokens for triangular arbitrage")
	}

	tokenB := common.HexToAddress(pair.Tokens[otherTokens[0]])
	tokenC := common.HexToAddress(pair.Tokens[otherTokens[1]])

	// Prepare paths
	path1 := []common.Address{tokenA, tokenB}
	path2 := []common.Address{tokenB, tokenC}
	path3 := []common.Address{tokenC, tokenA}

	// Calculate min amounts out with 1% slippage tolerance
	minOutA := new(big.Int).Div(new(big.Int).Mul(amount, big.NewInt(99)), big.NewInt(100))
	minOutB := new(big.Int).Div(new(big.Int).Mul(amount, big.NewInt(99)), big.NewInt(100))
	minOutC := new(big.Int).Div(new(big.Int).Mul(amount, big.NewInt(100)), big.NewInt(99))

	minAmountsOut := []*big.Int{minOutA, minOutB, minOutC}

	// Define pair address to borrow from
	var pairAddress common.Address
	pairKey := fmt.Sprintf("WBNB-%s", otherTokens[0])

	if pancakeFirst {
		if addr, exists := pair.PancakeswapPair[pairKey]; exists && addr != "" {
			pairAddress = common.HexToAddress(addr)
		}
	} else {
		if addr, exists := pair.BiswapPair[pairKey]; exists && addr != "" {
			pairAddress = common.HexToAddress(addr)
		}
	}

	if pairAddress == (common.Address{}) {
		return fmt.Errorf("pair address not found for flash loan")
	}

	log.Printf("Using pair address for flash loan: %s", pairAddress.Hex())

	// Prepare arbitrage data
	arbData := models.ArbitrageData{
		Path1:         path1,
		Path2:         path2,
		Path3:         path3,
		MinAmountsOut: minAmountsOut,
		Direction:     pancakeFirst,
	}

	// Get nonce
	nonce, err := s.Client.Client.PendingNonceAt(context.Background(), s.Client.Address)
	if err != nil {
		return err
	}

	// Get gas price
	gasPrice, err := s.Client.Client.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}

	// Pack function call
	callData, err := contracts.FlashABI.Pack(
		"executeFlashLoan",
		pairAddress,
		amount,
		arbData,
		pancakeFirst,
	)
	if err != nil {
		return err
	}

	// Create transaction
	tx := types.NewTransaction(
		nonce,
		s.FlashContract,
		big.NewInt(0), // no ether value
		s.Config.GasLimit,
		gasPrice,
		callData,
	)

	// Sign the transaction
	chainID := big.NewInt(56) // BSC chain ID
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), s.Client.PrivateKey)
	if err != nil {
		return err
	}

	// Send the transaction
	err = s.Client.Client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return err
	}

	log.Printf("Arbitrage transaction sent: %s", signedTx.Hash().Hex())

	// Wait for transaction to be mined
	receipt, err := bind.WaitMined(context.Background(), s.Client.Client, signedTx)
	if err != nil {
		return err
	}

	// Check if transaction was successful
	if receipt.Status == 0 {
		return fmt.Errorf("transaction failed")
	}

	log.Printf("Arbitrage transaction successful, gas used: %d", receipt.GasUsed)

	return nil
}

// ExecuteManualArbitrage executes a triangular arbitrage manually (without flash loans)
func (s *ArbitrageService) ExecuteManualArbitrage(
	pair models.TokenPair,
	amount *big.Int,
	pancakeFirst bool,
) error {
	log.Println("Executing manual arbitrage (warning: not using flash loans)...")

	// Get token addresses safely
	tokenA := common.HexToAddress(pair.Tokens["WBNB"])

	otherTokens := getOtherTokens(pair.Tokens)
	if len(otherTokens) < 2 {
		return fmt.Errorf("need at least 3 tokens for triangular arbitrage")
	}

	tokenB := common.HexToAddress(pair.Tokens[otherTokens[0]])
	tokenC := common.HexToAddress(pair.Tokens[otherTokens[1]])

	// Get decimals for logging
	decimalsA, err := s.TokenService.GetTokenDecimals(tokenA)
	if err != nil {
		return fmt.Errorf("failed to get WBNB decimals: %v", err)
	}

	log.Printf("Initial amount: %.6f WBNB",
		s.TokenService.ConvertToReadable(amount, decimalsA))

	// Prepare paths
	path1 := []common.Address{tokenA, tokenB}
	path2 := []common.Address{tokenB, tokenC}
	path3 := []common.Address{tokenC, tokenA}

	var route1Router, route2Router, route3Router common.Address
	var routeDescription string

	if pancakeFirst {
		// PancakeSwap -> BiSwap -> PancakeSwap
		route1Router = s.PancakeRouter
		route2Router = s.BiswapRouter
		route3Router = s.PancakeRouter
		routeDescription = "PancakeSwap -> BiSwap -> PancakeSwap"
	} else {
		// BiSwap -> PancakeSwap -> BiSwap
		route1Router = s.BiswapRouter
		route2Router = s.PancakeRouter
		route3Router = s.BiswapRouter
		routeDescription = "BiSwap -> PancakeSwap -> BiSwap"
	}

	log.Printf("Executing route: %s", routeDescription)

	// Step 1: Calculate min amounts out with 1% slippage tolerance
	amountsOut1, err := s.RouterService.GetAmountsOut(route1Router, amount, path1)
	if err != nil {
		return fmt.Errorf("error calculating amounts for step 1: %v", err)
	}
	minOut1 := new(big.Int).Div(new(big.Int).Mul(amountsOut1[1], big.NewInt(99)), big.NewInt(100))

	// Step 1: WBNB -> TokenB
	log.Printf("Step 1: Swapping %.6f WBNB for %s",
		s.TokenService.ConvertToReadable(amount, decimalsA), otherTokens[0])

	hash1, err := s.RouterService.SwapExactTokensForTokens(
		route1Router,
		amount,
		minOut1,
		path1,
	)
	if err != nil {
		return fmt.Errorf("error executing step 1 swap: %v", err)
	}

	log.Printf("Step 1 transaction sent: %s", hash1.Hex())

	// Wait for transaction confirmation
	log.Println("Waiting for step 1 confirmation...")
	time.Sleep(15 * time.Second)

	// Get TokenB balance after step 1
	balanceB, err := s.TokenService.GetTokenBalance(tokenB, s.Client.Address)
	if err != nil {
		return fmt.Errorf("error getting %s balance: %v", otherTokens[0], err)
	}

	decimalsB, err := s.TokenService.GetTokenDecimals(tokenB)
	if err != nil {
		return fmt.Errorf("error getting %s decimals: %v", otherTokens[0], err)
	}

	log.Printf("Received: %.6f %s",
		s.TokenService.ConvertToReadable(balanceB, decimalsB), otherTokens[0])

	// Step 2: Calculate min amounts out for TokenB -> TokenC
	amountsOut2, err := s.RouterService.GetAmountsOut(route2Router, balanceB, path2)
	if err != nil {
		return fmt.Errorf("error calculating amounts for step 2: %v", err)
	}
	minOut2 := new(big.Int).Div(new(big.Int).Mul(amountsOut2[1], big.NewInt(99)), big.NewInt(100))

	// Step 2: TokenB -> TokenC
	log.Printf("Step 2: Swapping %.6f %s for %s",
		s.TokenService.ConvertToReadable(balanceB, decimalsB),
		otherTokens[0], otherTokens[1])

	hash2, err := s.RouterService.SwapExactTokensForTokens(
		route2Router,
		balanceB,
		minOut2,
		path2,
	)
	if err != nil {
		return fmt.Errorf("error executing step 2 swap: %v", err)
	}

	log.Printf("Step 2 transaction sent: %s", hash2.Hex())

	// Wait for transaction confirmation
	log.Println("Waiting for step 2 confirmation...")
	time.Sleep(15 * time.Second)

	// Get TokenC balance after step 2
	balanceC, err := s.TokenService.GetTokenBalance(tokenC, s.Client.Address)
	if err != nil {
		return fmt.Errorf("error getting %s balance: %v", otherTokens[1], err)
	}

	decimalsC, err := s.TokenService.GetTokenDecimals(tokenC)
	if err != nil {
		return fmt.Errorf("error getting %s decimals: %v", otherTokens[1], err)
	}

	log.Printf("Received: %.6f %s",
		s.TokenService.ConvertToReadable(balanceC, decimalsC), otherTokens[1])

	// Step 3: Calculate min amounts out for TokenC -> WBNB
	amountsOut3, err := s.RouterService.GetAmountsOut(route3Router, balanceC, path3)
	if err != nil {
		return fmt.Errorf("error calculating amounts for step 3: %v", err)
	}
	minOut3 := new(big.Int).Div(new(big.Int).Mul(amountsOut3[1], big.NewInt(99)), big.NewInt(100))

	// Step 3: TokenC -> WBNB
	log.Printf("Step 3: Swapping %.6f %s for WBNB",
		s.TokenService.ConvertToReadable(balanceC, decimalsC), otherTokens[1])

	hash3, err := s.RouterService.SwapExactTokensForTokens(
		route3Router,
		balanceC,
		minOut3,
		path3,
	)
	if err != nil {
		return fmt.Errorf("error executing step 3 swap: %v", err)
	}

	log.Printf("Step 3 transaction sent: %s", hash3.Hex())

	// Wait for final transaction confirmation
	log.Println("Waiting for step 3 confirmation...")
	time.Sleep(15 * time.Second)

	// Get final WBNB balance
	finalBalance, err := s.TokenService.GetTokenBalance(tokenA, s.Client.Address)
	if err != nil {
		return fmt.Errorf("error getting final WBNB balance: %v", err)
	}

	// Calculate profit/loss
	profit := new(big.Int).Sub(finalBalance, amount)
	profitReadable := s.TokenService.ConvertToReadable(profit, decimalsA)
	initialReadable := s.TokenService.ConvertToReadable(amount, decimalsA)
	finalReadable := s.TokenService.ConvertToReadable(finalBalance, decimalsA)

	// Calculate profit percentage
	var profitPercent float64
	if amount.Cmp(big.NewInt(0)) > 0 {
		profitFloat := new(big.Float).SetInt(profit)
		initialFloat := new(big.Float).SetInt(amount)
		percentFloat := new(big.Float).Quo(profitFloat, initialFloat)
		profitPercent, _ = percentFloat.Float64()
	}

	// Display results
	log.Println("========================================")
	log.Println("Manual Arbitrage Execution Complete")
	log.Println("========================================")
	log.Printf("Route: %s", routeDescription)
	log.Printf("Initial WBNB: %.6f", initialReadable)
	log.Printf("Final WBNB: %.6f", finalReadable)
	log.Printf("Profit/Loss: %.6f WBNB", profitReadable)
	log.Printf("Profit Percentage: %.4f%%", profitPercent*100)
	log.Println("========================================")
	log.Println("Transaction Hashes:")
	log.Printf("Step 1 (WBNB -> %s): %s", otherTokens[0], hash1.Hex())
	log.Printf("Step 2 (%s -> %s): %s", otherTokens[0], otherTokens[1], hash2.Hex())
	log.Printf("Step 3 (%s -> WBNB): %s", otherTokens[1], hash3.Hex())
	log.Println("========================================")

	// Check if profitable
	if profit.Cmp(big.NewInt(0)) > 0 {
		log.Printf("✅ Arbitrage successful! Profit: %.6f WBNB (%.4f%%)",
			profitReadable, profitPercent*100)
	} else {
		log.Printf("❌ Arbitrage resulted in loss: %.6f WBNB (%.4f%%)",
			profitReadable, profitPercent*100)
	}

	return nil
}

// VerifyAndUpdatePairs verifies all pairs and dynamically updates addresses
func (s *ArbitrageService) VerifyAndUpdatePairs() error {
	log.Println("Verifying and updating pair addresses...")

	pancakeFactory := common.HexToAddress(config.PancakeswapFactory)
	biswapFactory := common.HexToAddress(config.BiswapFactory)

	for i, pair := range s.TokenPairs {
		log.Printf("Verifying pair: %s", pair.Name)

		tokenAAddr := common.HexToAddress(pair.Tokens["WBNB"])
		otherTokens := getOtherTokens(pair.Tokens)

		if len(otherTokens) < 2 {
			log.Printf("Skipping pair %s: insufficient tokens", pair.Name)
			continue
		}

		tokenBAddr := common.HexToAddress(pair.Tokens[otherTokens[0]])
		tokenCAddr := common.HexToAddress(pair.Tokens[otherTokens[1]])

		// Update pair addresses for both exchanges
		s.updatePairAddresses(&s.TokenPairs[i], pancakeFactory, biswapFactory,
			tokenAAddr, tokenBAddr, tokenCAddr, otherTokens)
	}

	return nil
}

// updatePairAddresses updates pair addresses for a given token pair
func (s *ArbitrageService) updatePairAddresses(
	pair *models.TokenPair,
	pancakeFactory, biswapFactory, tokenA, tokenB, tokenC common.Address,
	otherTokens []string,
) {
	// Update PancakeSwap pairs
	if pairAB, err := s.GetPairAddressFromFactory(pancakeFactory, tokenA, tokenB); err == nil {
		pair.PancakeswapPair["WBNB-"+otherTokens[0]] = pairAB.Hex()
		log.Printf("Updated PancakeSwap pair WBNB-%s: %s", otherTokens[0], pairAB.Hex())
	}

	if pairBC, err := s.GetPairAddressFromFactory(pancakeFactory, tokenB, tokenC); err == nil {
		pair.PancakeswapPair[otherTokens[0]+"-"+otherTokens[1]] = pairBC.Hex()
		log.Printf("Updated PancakeSwap pair %s-%s: %s", otherTokens[0], otherTokens[1], pairBC.Hex())
	}

	if pairCA, err := s.GetPairAddressFromFactory(pancakeFactory, tokenC, tokenA); err == nil {
		pair.PancakeswapPair[otherTokens[1]+"-WBNB"] = pairCA.Hex()
		log.Printf("Updated PancakeSwap pair %s-WBNB: %s", otherTokens[1], pairCA.Hex())
	}

	// Update BiSwap pairs
	if pairAB, err := s.GetPairAddressFromFactory(biswapFactory, tokenA, tokenB); err == nil {
		pair.BiswapPair["WBNB-"+otherTokens[0]] = pairAB.Hex()
		log.Printf("Updated BiSwap pair WBNB-%s: %s", otherTokens[0], pairAB.Hex())
	}

	if pairBC, err := s.GetPairAddressFromFactory(biswapFactory, tokenB, tokenC); err == nil {
		pair.BiswapPair[otherTokens[0]+"-"+otherTokens[1]] = pairBC.Hex()
		log.Printf("Updated BiSwap pair %s-%s: %s", otherTokens[0], otherTokens[1], pairBC.Hex())
	}

	if pairCA, err := s.GetPairAddressFromFactory(biswapFactory, tokenC, tokenA); err == nil {
		pair.BiswapPair[otherTokens[1]+"-WBNB"] = pairCA.Hex()
		log.Printf("Updated BiSwap pair %s-WBNB: %s", otherTokens[1], pairCA.Hex())
	}
}

// GetPairAddressFromFactory gets pair address from factory contract
func (s *ArbitrageService) GetPairAddressFromFactory(factoryAddress, tokenA, tokenB common.Address) (common.Address, error) {
	// Factory ABI
	factoryABI := `[{"inputs":[{"internalType":"address","name":"tokenA","type":"address"},{"internalType":"address","name":"tokenB","type":"address"}],"name":"getPair","outputs":[{"internalType":"address","name":"pair","type":"address"}],"stateMutability":"view","type":"function"}]`

	parsed, err := abi.JSON(strings.NewReader(factoryABI))
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to parse factory ABI: %v", err)
	}

	// Pack parameters
	callData, err := parsed.Pack("getPair", tokenA, tokenB)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to pack getPair: %v", err)
	}

	// Call contract
	result, err := s.Client.Client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &factoryAddress,
		Data: callData,
	}, nil)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to call getPair: %v", err)
	}

	// Unpack result
	var pairAddress common.Address
	err = parsed.UnpackIntoInterface(&pairAddress, "getPair", result)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to unpack getPair result: %v", err)
	}

	// Check if pair exists
	if pairAddress == (common.Address{}) {
		return common.Address{}, fmt.Errorf("pair does not exist")
	}

	return pairAddress, nil
}

// Helper function to get other tokens (non-WBNB tokens) from a pair
func getOtherTokens(tokens map[string]string) []string {
	var otherTokens []string
	for key := range tokens {
		if key != "WBNB" {
			otherTokens = append(otherTokens, key)
		}
	}

	// Sort for consistency
	sort.Strings(otherTokens)
	return otherTokens
}

/// arbitrage.go - FIXED VERSION - Replace line 754 onwards with this
// (Keep everything above line 754, replace everything after)

func (s *ArbitrageService) FindEnhancedArbitrageOpportunities() error {
	log.Println("🎯 Enhanced Arbitrage: Targeting meme coins for higher spreads...")

	// Check if we're in peak trading hours
	hour := time.Now().UTC().Hour()
	isPeakHour := (hour >= 13 && hour <= 16) || (hour >= 21 && hour <= 23)

	if isPeakHour {
		log.Println("🔥 PEAK HOURS - High meme coin volatility expected!")
	} else if hour >= 2 && hour <= 6 {
		log.Println("😴 Low activity hours - reduced opportunities expected")
	}

	// Get all pairs but prioritize meme coins
	pairs := s.TokenPairs
	foundOpportunity := false

	for _, pair := range pairs {
		// Determine pair category and settings
		category := getMemeCategory(pair.Name)
		minProfit := getMinProfitForCategory(category)
		gasAdjustment := getGasAdjustmentForCategory(category)

		log.Printf("🎯 Checking %s: %s (min profit: %.2f%%)", category, pair.Name, minProfit*100)

		// Try enhanced test amounts
		for _, amount := range pair.TestAmounts {
			// Check triangular arbitrage opportunities
			result1, err1 := s.CheckTriangularArbitrage(pair, amount, true)
			result2, err2 := s.CheckTriangularArbitrage(pair, amount, false)

			if err1 != nil && err2 != nil {
				log.Printf("⚠️ Both routes failed for %s: %v", pair.Name, err1)
				continue
			}

			var bestResult *models.ArbitrageResult
			var pancakeFirst bool
			var adjustedProfit float64

			// Evaluate Pancake->Biswap route
			if err1 == nil {
				adjustedProfit1 := result1.ProfitPercent - gasAdjustment
				log.Printf("📊 Pancake->Biswap: %.4f%% (Gas adj: %.4f%%)",
					result1.ProfitPercent*100, adjustedProfit1*100)

				if adjustedProfit1 >= minProfit {
					bestResult = result1
					pancakeFirst = true
					adjustedProfit = adjustedProfit1
				}
			}

			// Evaluate Biswap->Pancake route
			if err2 == nil {
				adjustedProfit2 := result2.ProfitPercent - gasAdjustment
				log.Printf("📊 Biswap->Pancake: %.4f%% (Gas adj: %.4f%%)",
					result2.ProfitPercent*100, adjustedProfit2*100)

				if adjustedProfit2 >= minProfit && (bestResult == nil || adjustedProfit2 > adjustedProfit) {
					bestResult = result2
					pancakeFirst = false
					adjustedProfit = adjustedProfit2
				}
			}

			// Execute if profitable
			if bestResult != nil {
				log.Printf("💰 ENHANCED OPPORTUNITY FOUND!")
				log.Printf("🚀 %s: %.4f%% profit (%.6f WBNB)", pair.Name, adjustedProfit*100, amount)
				log.Printf("📈 Category: %s, Route: %s", category, getRouteDescription(pancakeFirst))

				// Execute the arbitrage
				err := s.ExecuteArbitrage(pair, bestResult.TargetAmount, pancakeFirst)
				if err != nil {
					log.Printf("❌ Enhanced execution failed: %v", err)
				} else {
					foundOpportunity = true
					log.Printf("✅ Enhanced trade executed successfully!")
					recordEnhancedTrade(pair.Name, adjustedProfit, amount, category)
				}
				break // Move to next pair after execution
			}
		}

		if foundOpportunity {
			break // Focus on one opportunity at a time
		}
	}

	if !foundOpportunity {
		log.Println("😞 No enhanced opportunities found this round")
		suggestEnhancedOptimizations(isPeakHour)
	}

	return nil
}

// Helper functions for enhanced arbitrage
func getMemeCategory(pairName string) string {
	switch {
	case strings.Contains(pairName, "SHIB") || strings.Contains(pairName, "DOGE") ||
		strings.Contains(pairName, "FLOKI") || strings.Contains(pairName, "SAFEMOON"):
		return "meme"
	case strings.Contains(pairName, "BSW"):
		return "volatile"
	case strings.Contains(pairName, "CAKE"):
		return "established"
	case strings.Contains(pairName, "BUSD"):
		return "stable"
	default:
		return "unknown"
	}
}

func getMinProfitForCategory(category string) float64 {
	switch category {
	case "meme":
		return 0.005 // 0.5% for meme coins (higher volatility expected)
	case "volatile":
		return 0.003 // 0.3% for volatile tokens
	case "established":
		return 0.002 // 0.2% for established tokens
	case "stable":
		return 0.001 // 0.1% for stable pairs
	default:
		return 0.002
	}
}

func getGasAdjustmentForCategory(category string) float64 {
	switch category {
	case "meme":
		return 0.0015 // 0.15% - meme coins may have higher gas costs
	case "volatile":
		return 0.0012 // 0.12%
	case "established":
		return 0.0010 // 0.10%
	case "stable":
		return 0.0008 // 0.08%
	default:
		return 0.0010
	}
}

func getRouteDescription(pancakeFirst bool) string {
	if pancakeFirst {
		return "Pancake→Biswap→Pancake"
	}
	return "Biswap→Pancake→Biswap"
}

// Enhanced statistics tracking
var enhancedStats = struct {
	TotalTrades   int
	MemeTrades    int
	TotalProfit   float64
	BestTrade     float64
	CategoryStats map[string]int
}{
	CategoryStats: make(map[string]int),
}

func recordEnhancedTrade(pairName string, profit, amount float64, category string) {
	enhancedStats.TotalTrades++
	enhancedStats.CategoryStats[category]++

	tradeProfit := profit * amount
	enhancedStats.TotalProfit += tradeProfit

	if category == "meme" {
		enhancedStats.MemeTrades++
	}

	if tradeProfit > enhancedStats.BestTrade {
		enhancedStats.BestTrade = tradeProfit
	}

	log.Printf("📊 Enhanced Stats: %d total trades, %d meme trades, %.6f WBNB profit",
		enhancedStats.TotalTrades, enhancedStats.MemeTrades, enhancedStats.TotalProfit)
}

func suggestEnhancedOptimizations(isPeakHour bool) {
	if !isPeakHour {
		log.Println("💡 Not in peak hours - meme coins typically less volatile")
		log.Println("   Peak hours: 13-16 UTC (Asia), 21-23 UTC (US)")
	}

	if enhancedStats.TotalTrades > 3 && enhancedStats.MemeTrades == 0 {
		log.Println("💡 No meme trades yet - consider:")
		log.Println("   • Checking if SHIB/DOGE are actively traded")
		log.Println("   • Lowering meme coin threshold to 0.3%")
		log.Println("   • Waiting for market volatility")
	}
}
