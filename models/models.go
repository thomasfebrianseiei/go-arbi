// models/models.go
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

// Initialize token pairs with CORRECTED addresses
func InitializeTokenPairs() []TokenPair {
	return []TokenPair{
		{
			Name: "WBNB-BTCB-USDT",
			Tokens: map[string]string{
				"WBNB": "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c",
				"BTCB": "0x7130d2A12B9BCbFAe4f2634d864A1Ee1Ce3Ead9c",
				"USDT": "0x55d398326f99059fF775485246999027B3197955",
			},
			PancakeswapPair: map[string]string{
				"WBNB-BTCB": "0x61EB789d75A95CAa3fF50ed7E47b96c132fEc082",
				"BTCB-USDT": "0xD171B26E4484402de70e3Ea256bE5A2630d7e88D", // CORRECTED
				"USDT-WBNB": "0x16b9a82891338f9bA80E2D6970FddA79D1eb0daE",
			},
			BiswapPair: map[string]string{
				"WBNB-BTCB": "0xC7e9d76ba11099AF3F330ff829c5F442d571e057",
				"BTCB-USDT": "0xa987f0b7098585c735cD943ee07544a84e923d1D",
				"USDT-WBNB": "0x8840C6252e2e86e545deFb6da98B2a0E26d8C1BA",
			},
			Priority:    2,
			TestAmounts: []float64{0.01, 0.05, 0.1}, // WBNB amounts for testing
		},
		{
			Name: "WBNB-ETH-USDT",
			Tokens: map[string]string{
				"WBNB": "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c",
				"ETH":  "0x2170Ed0880ac9A755fd29B2688956BD959F933F8",
				"USDT": "0x55d398326f99059fF775485246999027B3197955",
			},
			PancakeswapPair: map[string]string{
				"WBNB-ETH":  "0x74E4716E431f45807DCF19f284c7aA99F18a4fbc",
				"ETH-USDT":  "0x531FEbfeb9a61D948c384ACFBe6dCc51057AEa7e",
				"USDT-WBNB": "0x16b9a82891338f9bA80E2D6970FddA79D1eb0daE",
			},
			BiswapPair: map[string]string{
				"WBNB-ETH":  "0x5bf6941f029424674bb93A43b79fc46bF4A67c21",
				"ETH-USDT":  "0x63b30de1A998e9E64FD58A21F68D323B9BcD8F85",
				"USDT-WBNB": "0x8840C6252e2e86e545deFb6da98B2a0E26d8C1BA",
			},
			Priority:    2,
			TestAmounts: []float64{0.01, 0.05, 0.1}, // WBNB amounts for testing
		},
		{
			Name: "WBNB-BSW-USDT",
			Tokens: map[string]string{
				"WBNB": "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c",
				"BSW":  "0x965F527D9159dCe6288a2219DB51fc6Eef120dD1", // CORRECT BSW address
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
			Priority:    1,
			TestAmounts: []float64{0.01, 0.05, 0.1}, // WBNB amounts for testing
		},
		// Add a simpler pair for testing
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
			Priority:    3,
			TestAmounts: []float64{0.005, 0.01, 0.02}, // Smaller amounts for stable coins
		},
	}
}
