package contracts

import (
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// ABI definitions for various contracts
var (
	RouterABI abi.ABI
	ERC20ABI  abi.ABI
	PairABI   abi.ABI
	FlashABI  abi.ABI
)

// Initialize loads all the required ABIs
func Initialize() error {
	var err error
	
	// Router ABI (minimum required functions)
	routerAbiJson := `[
		{"inputs":[{"internalType":"uint256","name":"amountIn","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"}],"name":"getAmountsOut","outputs":[{"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],"stateMutability":"view","type":"function"},
		{"inputs":[{"internalType":"uint256","name":"amountIn","type":"uint256"},{"internalType":"uint256","name":"amountOutMin","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"swapExactTokensForTokens","outputs":[{"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],"stateMutability":"nonpayable","type":"function"}
	]`
	
	// ERC20 ABI (minimum required functions)
	erc20AbiJson := `[
		{"inputs":[],"name":"decimals","outputs":[{"internalType":"uint8","name":"","type":"uint8"}],"stateMutability":"view","type":"function"},
		{"inputs":[{"internalType":"address","name":"account","type":"address"}],"name":"balanceOf","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},
		{"inputs":[{"internalType":"address","name":"spender","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"approve","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"}
	]`
	
	// Pair ABI (minimum required functions)
	pairAbiJson := `[
		{"inputs":[],"name":"token0","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},
		{"inputs":[],"name":"token1","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},
		{"inputs":[],"name":"getReserves","outputs":[{"internalType":"uint112","name":"reserve0","type":"uint112"},{"internalType":"uint112","name":"reserve1","type":"uint112"},{"internalType":"uint32","name":"blockTimestampLast","type":"uint32"}],"stateMutability":"view","type":"function"}
	]`
	
	// Flash arbitrage contract ABI (key functions only)
	flashAbiJson := `[
		{"inputs":[{"components":[{"internalType":"address[]","name":"path1","type":"address[]"},{"internalType":"address[]","name":"path2","type":"address[]"},{"internalType":"address[]","name":"path3","type":"address[]"},{"internalType":"uint256[]","name":"minAmountsOut","type":"uint256[]"},{"internalType":"bool","name":"direction","type":"bool"}],"internalType":"struct FlashTriangularArbitrage.ArbitrageData","name":"data","type":"tuple"},{"internalType":"uint256","name":"loanAmount","type":"uint256"},{"internalType":"bool","name":"fromPancake","type":"bool"}],"name":"checkArbitrageProfitability","outputs":[{"internalType":"uint256","name":"expectedProfit","type":"uint256"},{"internalType":"uint256","name":"expectedPlatformFee","type":"uint256"},{"internalType":"uint256","name":"expectedUserProfit","type":"uint256"}],"stateMutability":"view","type":"function"},
		{"inputs":[{"internalType":"address","name":"pairAddress","type":"address"},{"internalType":"uint256","name":"borrowAmount","type":"uint256"},{"components":[{"internalType":"address[]","name":"path1","type":"address[]"},{"internalType":"address[]","name":"path2","type":"address[]"},{"internalType":"address[]","name":"path3","type":"address[]"},{"internalType":"uint256[]","name":"minAmountsOut","type":"uint256[]"},{"internalType":"bool","name":"direction","type":"bool"}],"internalType":"struct FlashTriangularArbitrage.ArbitrageData","name":"data","type":"tuple"},{"internalType":"bool","name":"fromPancake","type":"bool"}],"name":"executeFlashLoan","outputs":[],"stateMutability":"nonpayable","type":"function"}
	]`
	
	RouterABI, err = abi.JSON(strings.NewReader(routerAbiJson))
	if err != nil {
		return err
	}
	
	ERC20ABI, err = abi.JSON(strings.NewReader(erc20AbiJson))
	if err != nil {
		return err
	}
	
	PairABI, err = abi.JSON(strings.NewReader(pairAbiJson))
	if err != nil {
		return err
	}
	
	FlashABI, err = abi.JSON(strings.NewReader(flashAbiJson))
	if err != nil {
		return err
	}
	
	return nil
}