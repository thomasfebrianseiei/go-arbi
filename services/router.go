// services/router.go
package services

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"arbitrage-bot/config"
	"arbitrage-bot/contracts"
)

// RouterService handles operations related to DEX routers
type RouterService struct {
	Client       *EthClient
	TokenService *TokenService
	Config       *config.Config
	RouterABI    abi.ABI
}

// NewRouterService creates a new RouterService
func NewRouterService(client *EthClient, tokenService *TokenService, cfg *config.Config) *RouterService {
	return &RouterService{
		Client:       client,
		TokenService: tokenService,
		Config:       cfg,
		RouterABI:    contracts.RouterABI,
	}
}

// GetAmountsOut returns the expected output amounts for a given input amount and path
func (s *RouterService) GetAmountsOut(router common.Address, amountIn *big.Int, path []common.Address) ([]*big.Int, error) {
	if len(path) < 2 {
		return nil, fmt.Errorf("path must contain at least 2 tokens")
	}

	if amountIn == nil || amountIn.Cmp(big.NewInt(0)) <= 0 {
		return nil, fmt.Errorf("invalid input amount")
	}

	// Validate path addresses
	for i, addr := range path {
		if addr == (common.Address{}) {
			return nil, fmt.Errorf("invalid token address at path index %d", i)
		}

		// Check for identical consecutive addresses
		if i > 0 && path[i-1] == addr {
			return nil, fmt.Errorf("identical consecutive addresses in path at indices %d and %d", i-1, i)
		}
	}

	// Pack the function call
	callData, err := s.RouterABI.Pack("getAmountsOut", amountIn, path)
	if err != nil {
		return nil, fmt.Errorf("failed to pack getAmountsOut: %v", err)
	}

	// Call the contract with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := s.Client.Client.CallContract(ctx, ethereum.CallMsg{
		To:   &router,
		Data: callData,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call getAmountsOut on router %s: %v", router.Hex(), err)
	}

	// Unpack the result
	var amounts []*big.Int
	err = s.RouterABI.UnpackIntoInterface(&amounts, "getAmountsOut", result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack getAmountsOut result: %v", err)
	}

	// Validate results
	if len(amounts) != len(path) {
		return nil, fmt.Errorf("unexpected result length: got %d, expected %d", len(amounts), len(path))
	}

	// Check for zero amounts (indicates liquidity issues)
	for i, amount := range amounts {
		if amount == nil || amount.Cmp(big.NewInt(0)) <= 0 {
			return nil, fmt.Errorf("zero output amount at index %d, possible liquidity issue", i)
		}
	}

	return amounts, nil
}

// GetAmountOutSingle returns the expected output amount for a single swap
func (s *RouterService) GetAmountOutSingle(router common.Address, amountIn *big.Int, path []common.Address) (*big.Int, error) {
	amounts, err := s.GetAmountsOut(router, amountIn, path)
	if err != nil {
		return nil, err
	}

	if len(amounts) < 2 {
		return nil, fmt.Errorf("insufficient amounts returned")
	}

	return amounts[len(amounts)-1], nil
}

// SwapExactTokensForTokens executes a token swap
func (s *RouterService) SwapExactTokensForTokens(
	router common.Address,
	amountIn *big.Int,
	amountOutMin *big.Int,
	path []common.Address,
) (*common.Hash, error) {
	if len(path) < 2 {
		return nil, fmt.Errorf("path must contain at least 2 tokens")
	}

	// Validate input parameters
	if amountIn == nil || amountIn.Cmp(big.NewInt(0)) <= 0 {
		return nil, fmt.Errorf("invalid input amount")
	}
	if amountOutMin == nil || amountOutMin.Cmp(big.NewInt(0)) < 0 {
		return nil, fmt.Errorf("invalid minimum output amount")
	}

	// Get nonce
	nonce, err := s.Client.Client.PendingNonceAt(context.Background(), s.Client.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %v", err)
	}

	// Get gas price
	gasPrice, err := s.Client.Client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %v", err)
	}

	// Add 20% buffer to gas price for faster execution
	gasPrice = new(big.Int).Mul(gasPrice, big.NewInt(120))
	gasPrice = new(big.Int).Div(gasPrice, big.NewInt(100))

	// Calculate deadline (5 minutes from now)
	deadline := big.NewInt(time.Now().Unix() + 300)

	// Pack function call
	callData, err := s.RouterABI.Pack(
		"swapExactTokensForTokens",
		amountIn,
		amountOutMin,
		path,
		s.Client.Address,
		deadline,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to pack swap function: %v", err)
	}

	// Create transaction
	tx := types.NewTransaction(
		nonce,
		router,
		big.NewInt(0), // no ether value for token swaps
		s.Config.GasLimit,
		gasPrice,
		callData,
	)

	// Sign transaction
	chainID := big.NewInt(56) // BSC chain ID
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), s.Client.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %v", err)
	}

	// Send transaction
	err = s.Client.Client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %v", err)
	}

	hash := signedTx.Hash()
	log.Printf("Swap transaction sent: %s", hash.Hex())

	return &hash, nil
}

// GetReserves gets the reserves of a liquidity pair
func (s *RouterService) GetReserves(pairAddress common.Address) (reserve0, reserve1 *big.Int, blockTimestampLast uint32, err error) {
	// Pair contract ABI for getReserves function
	pairABI := `[{"inputs":[],"name":"getReserves","outputs":[{"internalType":"uint112","name":"_reserve0","type":"uint112"},{"internalType":"uint112","name":"_reserve1","type":"uint112"},{"internalType":"uint32","name":"_blockTimestampLast","type":"uint32"}],"stateMutability":"view","type":"function"}]`

	parsed, err := abi.JSON(strings.NewReader(pairABI))
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to parse pair ABI: %v", err)
	}

	// Pack the function call
	callData, err := parsed.Pack("getReserves")
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to pack getReserves: %v", err)
	}

	// Call the contract
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := s.Client.Client.CallContract(ctx, ethereum.CallMsg{
		To:   &pairAddress,
		Data: callData,
	}, nil)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to call getReserves: %v", err)
	}

	// Unpack the result
	var reserves struct {
		Reserve0           *big.Int
		Reserve1           *big.Int
		BlockTimestampLast uint32
	}

	err = parsed.UnpackIntoInterface(&reserves, "getReserves", result)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to unpack getReserves result: %v", err)
	}

	return reserves.Reserve0, reserves.Reserve1, reserves.BlockTimestampLast, nil
}

// ValidateSwapPath validates that a swap path is valid
func (s *RouterService) ValidateSwapPath(path []common.Address) error {
	if len(path) < 2 {
		return fmt.Errorf("path must contain at least 2 tokens")
	}

	if len(path) > 4 {
		return fmt.Errorf("path too long, maximum 4 hops supported")
	}

	// Check for duplicate addresses
	seen := make(map[common.Address]bool)
	for i, addr := range path {
		if addr == (common.Address{}) {
			return fmt.Errorf("zero address at index %d", i)
		}

		if seen[addr] {
			return fmt.Errorf("duplicate address %s in path", addr.Hex())
		}
		seen[addr] = true
	}

	return nil
}

// EstimateGasForSwap estimates gas required for a swap transaction
func (s *RouterService) EstimateGasForSwap(
	router common.Address,
	amountIn *big.Int,
	amountOutMin *big.Int,
	path []common.Address,
) (uint64, error) {
	// Validate inputs
	if err := s.ValidateSwapPath(path); err != nil {
		return 0, err
	}

	// Calculate deadline
	deadline := big.NewInt(time.Now().Unix() + 300)

	// Pack function call
	callData, err := s.RouterABI.Pack(
		"swapExactTokensForTokens",
		amountIn,
		amountOutMin,
		path,
		s.Client.Address,
		deadline,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to pack swap function: %v", err)
	}

	// Estimate gas
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	gasLimit, err := s.Client.Client.EstimateGas(ctx, ethereum.CallMsg{
		From: s.Client.Address,
		To:   &router,
		Data: callData,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to estimate gas: %v", err)
	}

	// Add 20% buffer
	gasLimit = gasLimit * 120 / 100

	return gasLimit, nil
}

// CheckLiquidity checks if there's sufficient liquidity for a trade
func (s *RouterService) CheckLiquidity(router common.Address, amountIn *big.Int, path []common.Address) error {
	if len(path) < 2 {
		return fmt.Errorf("path must contain at least 2 tokens")
	}

	// Try to get amounts out - if this fails, liquidity is likely insufficient
	amounts, err := s.GetAmountsOut(router, amountIn, path)
	if err != nil {
		return fmt.Errorf("insufficient liquidity: %v", err)
	}

	// Check that output amount is reasonable (not too small)
	finalAmount := amounts[len(amounts)-1]
	if finalAmount.Cmp(big.NewInt(1000)) < 0 { // Less than 1000 wei
		return fmt.Errorf("output amount too small, possible liquidity issue")
	}

	return nil
}

// GetPriceImpact calculates the price impact of a trade
func (s *RouterService) GetPriceImpact(router common.Address, amountIn *big.Int, path []common.Address) (float64, error) {
	if len(path) != 2 {
		return 0, fmt.Errorf("price impact calculation only supports direct pairs")
	}

	// Get small amount for reference price
	smallAmount := new(big.Int).Div(amountIn, big.NewInt(1000))
	if smallAmount.Cmp(big.NewInt(1)) < 0 {
		smallAmount = big.NewInt(1)
	}

	// Get reference price with small amount
	refAmounts, err := s.GetAmountsOut(router, smallAmount, path)
	if err != nil {
		return 0, err
	}

	// Get actual price with full amount
	actualAmounts, err := s.GetAmountsOut(router, amountIn, path)
	if err != nil {
		return 0, err
	}

	// Calculate price per unit
	refPrice := new(big.Float).Quo(new(big.Float).SetInt(refAmounts[1]), new(big.Float).SetInt(smallAmount))
	actualPrice := new(big.Float).Quo(new(big.Float).SetInt(actualAmounts[1]), new(big.Float).SetInt(amountIn))

	// Calculate price impact
	priceDiff := new(big.Float).Sub(refPrice, actualPrice)
	priceImpact := new(big.Float).Quo(priceDiff, refPrice)

	impact, _ := priceImpact.Float64()
	return impact * 100, nil // Return as percentage
}
