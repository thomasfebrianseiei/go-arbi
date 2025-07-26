// models/models.go - Complete replacement
package models

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// TokenPair represents a token pair for arbitrage
type TokenPair struct {
	Name            string
	Tokens          map[string]string
	PancakeswapPair map[string]string
	BiswapPair      map[string]string
	Priority        int
	TestAmounts     []float64
}

// ArbitrageData represents the data structure for arbitrage execution
type ArbitrageData struct {
	Path1         []common.Address
	Path2         []common.Address
	Path3         []common.Address
	MinAmountsOut []*big.Int
	Direction     bool
}

// ArbitrageResult represents the result of an arbitrage operation
type ArbitrageResult struct {
	Profit        *big.Int
	PlatformFee   *big.Int
	UserProfit    *big.Int
	TargetAmount  *big.Int
	ProfitPercent float64
	Direction     bool
	Path          []string
}

// PairReserves represents the reserves of a token pair
type PairReserves struct {
	Reserve0 *big.Int
	Reserve1 *big.Int
	Token0   common.Address
	Token1   common.Address
}

// Initialize token pairs with ENHANCED MEME COIN FOCUS
func InitializeTokenPairs() []TokenPair {
	return []TokenPair{
		// PRIORITY 1: MEME COINS - Higher volatility and spreads
		{
			Name: "WBNB-SHIB-USDT",
			Tokens: map[string]string{
				"WBNB": "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c",
				"SHIB": "0x2859e4544C4bB03966803b044A93563Bd2D0DD4D", // SHIB on BSC
				"USDT": "0x55d398326f99059fF775485246999027B3197955",
			},
			PancakeswapPair: map[string]string{
				"WBNB-SHIB": "0x...", // Will be fetched dynamically
				"SHIB-USDT": "0x...",
				"USDT-WBNB": "0x16b9a82891338f9bA80E2D6970FddA79D1eb0daE",
			},
			BiswapPair: map[string]string{
				"WBNB-SHIB": "0x...", // Will be fetched dynamically
				"SHIB-USDT": "0x...",
				"USDT-WBNB": "0x8840C6252e2e86e545deFb6da98B2a0E26d8C1BA",
			},
			Priority:    1,
			TestAmounts: []float64{0.1, 0.5, 1.0, 2.0}, // Larger amounts for meme coins
		},
		{
			Name: "WBNB-DOGE-USDT",
			Tokens: map[string]string{
				"WBNB": "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c",
				"DOGE": "0xbA2aE424d960c26247Dd6c32edC70B295c744C43", // DOGE on BSC
				"USDT": "0x55d398326f99059fF775485246999027B3197955",
			},
			PancakeswapPair: map[string]string{
				"WBNB-DOGE": "0x...",
				"DOGE-USDT": "0x...",
				"USDT-WBNB": "0x16b9a82891338f9bA80E2D6970FddA79D1eb0daE",
			},
			BiswapPair: map[string]string{
				"WBNB-DOGE": "0x...",
				"DOGE-USDT": "0x...",
				"USDT-WBNB": "0x8840C6252e2e86e545deFb6da98B2a0E26d8C1BA",
			},
			Priority:    1,
			TestAmounts: []float64{0.1, 0.5, 1.0, 2.0},
		},
		{
			Name: "WBNB-FLOKI-USDT",
			Tokens: map[string]string{
				"WBNB":  "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c",
				"FLOKI": "0xfb5B838b6cfEEdC2873aB27866079AC55363D37E", // FLOKI on BSC
				"USDT":  "0x55d398326f99059fF775485246999027B3197955",
			},
			PancakeswapPair: map[string]string{
				"WBNB-FLOKI": "0x...",
				"FLOKI-USDT": "0x...",
				"USDT-WBNB":  "0x16b9a82891338f9bA80E2D6970FddA79D1eb0daE",
			},
			BiswapPair: map[string]string{
				"WBNB-FLOKI": "0x...",
				"FLOKI-USDT": "0x...",
				"USDT-WBNB":  "0x8840C6252e2e86e545deFb6da98B2a0E26d8C1BA",
			},
			Priority:    1,
			TestAmounts: []float64{0.1, 0.5, 1.0, 2.0},
		},
		{
			Name: "WBNB-SAFEMOON-USDT",
			Tokens: map[string]string{
				"WBNB":     "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c",
				"SAFEMOON": "0x42981d0bfbAf196529376EE702F2a9Eb9092fcB5", // SafeMoon V2
				"USDT":     "0x55d398326f99059fF775485246999027B3197955",
			},
			PancakeswapPair: map[string]string{
				"WBNB-SAFEMOON": "0x...",
				"SAFEMOON-USDT": "0x...",
				"USDT-WBNB":     "0x16b9a82891338f9bA80E2D6970FddA79D1eb0daE",
			},
			BiswapPair: map[string]string{
				"WBNB-SAFEMOON": "0x...",
				"SAFEMOON-USDT": "0x...",
				"USDT-WBNB":     "0x8840C6252e2e86e545deFb6da98B2a0E26d8C1BA",
			},
			Priority:    1,
			TestAmounts: []float64{0.05, 0.1, 0.2}, // Smaller amounts for high risk
		},

		// PRIORITY 2: BSW (BiSwap native token - often has arbitrage opportunities)
		{
			Name: "WBNB-BSW-USDT",
			Tokens: map[string]string{
				"WBNB": "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c",
				"BSW":  "0x965F527D9159dCe6288a2219DB51fc6Eef120dD1",
				"USDT": "0x55d398326f99059fF775485246999027B3197955",
			},
			PancakeswapPair: map[string]string{
				"WBNB-BSW":  "0x8CA3fF14A52b080C54A6d1a405eecA02959d39fE",
				"BSW-USDT":  "0x4344b51cf44f4f182a8e25239cf4d81b315331c3",
				"USDT-WBNB": "0x16b9a82891338f9bA80E2D6970FddA79D1eb0daE",
			},
			BiswapPair: map[string]string{
				"WBNB-BSW":  "0x46492B26639Df0cda9b2769429845cb991591E0A",
				"BSW-USDT":  "0x2b30c317ceDFb554Ec525F85E79538D59970BEb0",
				"USDT-WBNB": "0x8840C6252e2e86e545deFb6da98B2a0E26d8C1BA",
			},
			Priority:    2,
			TestAmounts: []float64{0.1, 0.5, 1.0},
		},

		// PRIORITY 3: High volume established pairs (backup)
		{
			Name: "WBNB-CAKE-USDT",
			Tokens: map[string]string{
				"WBNB": "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c",
				"CAKE": "0x0E09FaBB73Bd3Ade0a17ECC321fD13a19e81cE82", // PancakeSwap token
				"USDT": "0x55d398326f99059fF775485246999027B3197955",
			},
			PancakeswapPair: map[string]string{
				"WBNB-CAKE": "0x0eD7e52944161450477ee417DE9Cd3a859b14fD0",
				"CAKE-USDT": "0x...",
				"USDT-WBNB": "0x16b9a82891338f9bA80E2D6970FddA79D1eb0daE",
			},
			BiswapPair: map[string]string{
				"WBNB-CAKE": "0x...",
				"CAKE-USDT": "0x...",
				"USDT-WBNB": "0x8840C6252e2e86e545deFb6da98B2a0E26d8C1BA",
			},
			Priority:    3,
			TestAmounts: []float64{0.1, 0.5, 1.0, 2.0},
		},

		// PRIORITY 4: Stable pairs (fallback - keep some for safety)
		{
			Name: "WBNB-USDT-BUSD",
			Tokens: map[string]string{
				"WBNB": "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c",
				"USDT": "0x55d398326f99059fF775485246999027B3197955",
				"BUSD": "0xe9e7CEA3DedcA5984780Bafc599bD69ADd087D56",
			},
			PancakeswapPair: map[string]string{
				"WBNB-USDT": "0x16b9a82891338f9bA80E2D6970FddA79D1eb0daE",
				"USDT-BUSD": "0x7EFaEf62fDdCCa950418312c6C91Aef321375A00",
				"BUSD-WBNB": "0x58F876857a02D6762E0101bb5C46A8c1ED44Dc16",
			},
			BiswapPair: map[string]string{
				"WBNB-USDT": "0x8840C6252e2e86e545deFb6da98B2a0E26d8C1BA",
				"USDT-BUSD": "0xDA8ceb724A06819c0A5cDb4304ea0cB27F8304cF",
				"BUSD-WBNB": "0x2b30c317ceDFb554Ec525F85E79538D59970BEb0",
			},
			Priority:    4,
			TestAmounts: []float64{0.05, 0.1, 0.2}, // Smaller amounts for stable pairs
		},
	}
}
