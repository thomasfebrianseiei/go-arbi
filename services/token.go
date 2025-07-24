// services/token.go
package services

import (
	"context"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"arbitrage-bot/contracts"
)

// TokenService handles operations related to ERC20 tokenas
type TokenService struct {
	Client *EthClient
}

// NewTokenService creates a new TokenService
func NewTokenService(client *EthClient) *TokenService {
	return &TokenService{
		Client: client,
	}
}

// GetTokenDecimals returns the decimals of a token
func (s *TokenService) GetTokenDecimals(tokenAddress common.Address) (uint8, error) {
	callData, err := contracts.ERC20ABI.Pack("decimals")
	if err != nil {
		return 0, err
	}

	result, err := s.Client.Client.CallContract(context.Background(),
		ethereum.CallMsg{
			To:   &tokenAddress,
			Data: callData,
		},
		nil, // latest block
	)

	if err != nil {
		return 0, err
	}

	decimals := new(uint8)
	err = contracts.ERC20ABI.UnpackIntoInterface(decimals, "decimals", result)
	if err != nil {
		return 0, err
	}

	return *decimals, nil
}

// GetTokenBalance returns the balance of a token for a specific address
func (s *TokenService) GetTokenBalance(tokenAddress, ownerAddress common.Address) (*big.Int, error) {
	callData, err := contracts.ERC20ABI.Pack("balanceOf", ownerAddress)
	if err != nil {
		return nil, err
	}

	result, err := s.Client.Client.CallContract(context.Background(),
		ethereum.CallMsg{
			To:   &tokenAddress,
			Data: callData,
		},
		nil, // latest block
	)

	if err != nil {
		return nil, err
	}

	var balance *big.Int
	err = contracts.ERC20ABI.UnpackIntoInterface(&balance, "balanceOf", result)
	if err != nil {
		return nil, err
	}

	return balance, nil
}

// ApproveToken approves a spender to spend tokens
func (s *TokenService) ApproveToken(tokenAddress, spenderAddress common.Address, amount *big.Int) (*common.Hash, error) {
	nonce, err := s.Client.Client.PendingNonceAt(context.Background(), s.Client.Address)
	if err != nil {
		return nil, err
	}

	gasPrice, err := s.Client.Client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(s.Client.PrivateKey, big.NewInt(56)) // BSC chain ID
	if err != nil {
		return nil, err
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // no value
	auth.GasLimit = uint64(100000) // gas limit for approve
	auth.GasPrice = gasPrice

	callData, err := contracts.ERC20ABI.Pack("approve", spenderAddress, amount)
	if err != nil {
		return nil, err
	}

	tx := types.NewTransaction(
		auth.Nonce.Uint64(),
		tokenAddress,
		auth.Value,
		auth.GasLimit,
		auth.GasPrice,
		callData,
	)

	// Sign transaction
	chainID := big.NewInt(56) // BSC chain ID
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), s.Client.PrivateKey)
	if err != nil {
		return nil, err
	}

	// Send transaction
	err = s.Client.Client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, err
	}

	hash := signedTx.Hash()
	return &hash, nil
}

// FormatTokenAmount formats a token amount with the correct number of decimals
func (s *TokenService) FormatTokenAmount(amount float64, decimals uint8) *big.Int {
	// Convert float to string with high precision
	amountStr := strconv.FormatFloat(amount, 'f', int(decimals), 64)

	// Create big.Float for precise conversion
	amountFloat, _ := new(big.Float).SetString(amountStr)

	// Multiply by 10^decimals
	multiplier := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil))
	amountFloat.Mul(amountFloat, multiplier)

	// Convert to big.Int
	amountInt, _ := amountFloat.Int(nil)
	return amountInt
}

// ConvertToReadable converts a token amount from wei to a readable format
func (s *TokenService) ConvertToReadable(amount *big.Int, decimals uint8) float64 {
	if amount == nil {
		return 0
	}

	// Convert to big.Float
	fAmount := new(big.Float).SetInt(amount)

	// Divide by 10^decimals
	divisor := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil))
	fAmount.Quo(fAmount, divisor)

	// Convert to float64
	result, _ := fAmount.Float64()
	return result
}
