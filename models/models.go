// models/models.go - High Volume Trading Pairs Only
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

// Initialize token pairs with HIGH VOLUME FOCUS - coins with consistent trading activity
func InitializeTokenPairs() []TokenPair {
	return []TokenPair{
		// PRIORITY 1: HIGH VOLUME STABLE PAIRS - Most liquid and active
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
			Priority:    1,
			TestAmounts: []float64{0.05, 0.1, 0.2, 0.5}, // Conservative amounts for stable pairs
		},

		// PRIORITY 2: BSW - BiSwap native token with high trading volume
		{
			Name: "WBNB-BSW-USDT",
			Tokens: map[string]string{
				"WBNB": "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c",
				"BSW":  "0x965F527D9159dCe6288a2219DB51fc6Eef120dD1", // BiSwap Token - high volume
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
			TestAmounts: []float64{0.1, 0.3, 0.5, 1.0},
		},

		// PRIORITY 3: CAKE - PancakeSwap native token, very high volume
		{
			Name: "WBNB-CAKE-USDT",
			Tokens: map[string]string{
				"WBNB": "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c",
				"CAKE": "0x0E09FaBB73Bd3Ade0a17ECC321fD13a19e81cE82", // PancakeSwap token - huge volume
				"USDT": "0x55d398326f99059fF775485246999027B3197955",
			},
			PancakeswapPair: map[string]string{
				"WBNB-CAKE": "0x0eD7e52944161450477ee417DE9Cd3a859b14fD0",
				"CAKE-USDT": "0xA39Af17CE4a8eb807E076805Da1e2B8EA7D0755b", // High volume pair
				"USDT-WBNB": "0x16b9a82891338f9bA80E2D6970FddA79D1eb0daE",
			},
			BiswapPair: map[string]string{
				"WBNB-CAKE": "0x2bF2dEB40639201C9A94c9e33b4852D9AEa5fd2D", // Also available on BiSwap
				"CAKE-USDT": "0x7b3ae32eE8C532016f3E31C8941D937c59e055B9",
				"USDT-WBNB": "0x8840C6252e2e86e545deFb6da98B2a0E26d8C1BA",
			},
			Priority:    3,
			TestAmounts: []float64{0.1, 0.5, 1.0, 2.0},
		},

		// PRIORITY 4: BTC - Wrapped Bitcoin, extremely high volume and liquidity
		{
			Name: "WBNB-BTCB-USDT",
			Tokens: map[string]string{
				"WBNB": "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c",
				"BTCB": "0x7130d2A12B9BCbFAe4f2634d864A1Ee1Ce3Ead9c", // Bitcoin BEP20 - massive volume
				"USDT": "0x55d398326f99059fF775485246999027B3197955",
			},
			PancakeswapPair: map[string]string{
				"WBNB-BTCB": "0x61EB789d75A95CAa3fF50ed7E47b96c132fEc082",
				"BTCB-USDT": "0xD171B26E4484402de70e3944CC6B9301014B0C1e",
				"USDT-WBNB": "0x16b9a82891338f9bA80E2D6970FddA79D1eb0daE",
			},
			BiswapPair: map[string]string{
				"WBNB-BTCB": "0xF45cd219aEF8618A92BAa7aD848364a158a24F33",
				"BTCB-USDT": "0x54c1EC2F543966953F2f7564692606EA7d5a184E",
				"USDT-WBNB": "0x8840C6252e2e86e545deFb6da98B2a0E26d8C1BA",
			},
			Priority:    4,
			TestAmounts: []float64{0.1, 0.5, 1.0},
		},

		// PRIORITY 5: ETH - Wrapped Ethereum, very high volume
		{
			Name: "WBNB-ETH-USDT",
			Tokens: map[string]string{
				"WBNB": "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c",
				"ETH":  "0x2170Ed0880ac9A755fd29B2688956BD959F933F8", // Ethereum BEP20 - high volume
				"USDT": "0x55d398326f99059fF775485246999027B3197955",
			},
			PancakeswapPair: map[string]string{
				"WBNB-ETH":  "0x74E4716E431f45807DCF19f284c7aA99F18a4fbc",
				"ETH-USDT":  "0x531FEbA4f95504c4B9883Eb2b59C30Bd7a10aBfB",
				"USDT-WBNB": "0x16b9a82891338f9bA80E2D6970FddA79D1eb0daE",
			},
			BiswapPair: map[string]string{
				"WBNB-ETH":  "0x70D8929d04b60Af4fb9B58713eBcf18765aDE422",
				"ETH-USDT":  "0x6A445cEB72c8B1751755819D4B1e50897bC03F91",
				"USDT-WBNB": "0x8840C6252e2e86e545deFb6da98B2a0E26d8C1BA",
			},
			Priority:    5,
			TestAmounts: []float64{0.1, 0.5, 1.0},
		},

		// PRIORITY 6: ADA - Cardano, good trading volume
		{
			Name: "WBNB-ADA-USDT",
			Tokens: map[string]string{
				"WBNB": "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c",
				"ADA":  "0x3EE2200Efb3400fAbB9AacF31297cBdD1d435D47", // Cardano BEP20 - good volume
				"USDT": "0x55d398326f99059fF775485246999027B3197955",
			},
			PancakeswapPair: map[string]string{
				"WBNB-ADA":  "0x28415ff2C35b65B9E5c7de82126b4015ab9d031F",
				"ADA-USDT":  "0x8CEd2C5F5FFFd7d6a27ac0c7b5c0Ed84E3e8b19A",
				"USDT-WBNB": "0x16b9a82891338f9bA80E2D6970FddA79D1eb0daE",
			},
			BiswapPair: map[string]string{
				"WBNB-ADA":  "0x5610837fa26e1Fa17e73B8E9F5bCb7eBfA4E4d6A",
				"ADA-USDT":  "0x2BCa7cEf78e5F2fccaEBC8aED5A29Cc5AE90e23E",
				"USDT-WBNB": "0x8840C6252e2e86e545deFb6da98B2a0E26d8C1BA",
			},
			Priority:    6,
			TestAmounts: []float64{0.1, 0.3, 0.5},
		},

		// PRIORITY 7: DOT - Polkadot, consistent trading volume
		{
			Name: "WBNB-DOT-USDT",
			Tokens: map[string]string{
				"WBNB": "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c",
				"DOT":  "0x7083609fCE4d1d8Dc0C979AAb8c869Ea2C873402", // Polkadot BEP20 - steady volume
				"USDT": "0x55d398326f99059fF775485246999027B3197955",
			},
			PancakeswapPair: map[string]string{
				"WBNB-DOT":  "0xDd5bAd8f8b360d76d12FdA230F8BAF42fe0022CF",
				"DOT-USDT":  "0x54aFF400858Dcac39797a81894D9920f16972D1D",
				"USDT-WBNB": "0x16b9a82891338f9bA80E2D6970FddA79D1eb0daE",
			},
			BiswapPair: map[string]string{
				"WBNB-DOT":  "0xF9045866e7b372DeF1EFf3712CE55FAc1A9C01E3",
				"DOT-USDT":  "0x1B96B92314C44b159149f7E0303511fB2Fc4774f",
				"USDT-WBNB": "0x8840C6252e2e86e545deFb6da98B2a0E26d8C1BA",
			},
			Priority:    7,
			TestAmounts: []float64{0.1, 0.3, 0.5},
		},

		// PRIORITY 8: MATIC - Polygon, active trading
		{
			Name: "WBNB-MATIC-USDT",
			Tokens: map[string]string{
				"WBNB":  "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c",
				"MATIC": "0xCC42724C6683B7E57334c4E856f4c9965ED682bD", // Polygon BEP20 - active trading
				"USDT":  "0x55d398326f99059fF775485246999027B3197955",
			},
			PancakeswapPair: map[string]string{
				"WBNB-MATIC": "0x7Eb5D86FD78f3852a3e0e064f2842d45a3dB6EA2",
				"MATIC-USDT": "0x3578B7cc2aFD09bE5bFAaED8c1d89C40dFbf4b3A",
				"USDT-WBNB":  "0x16b9a82891338f9bA80E2D6970FddA79D1eb0daE",
			},
			BiswapPair: map[string]string{
				"WBNB-MATIC": "0x8EE2c34311E7ae84af96bfd5C3De1D72Ec4d3d0c",
				"MATIC-USDT": "0x1a1Ba8C3F6f4EEe7EB9B7D5b3c1Ab1e6D4D8f5Ac",
				"USDT-WBNB":  "0x8840C6252e2e86e545deFb6da98B2a0E26d8C1BA",
			},
			Priority:    8,
			TestAmounts: []float64{0.1, 0.3, 0.5},
		},

		// PRIORITY 9: LINK - Chainlink, reliable volume
		{
			Name: "WBNB-LINK-USDT",
			Tokens: map[string]string{
				"WBNB": "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c",
				"LINK": "0xF8A0BF9cF54Bb92F17374d9e9A321E6a111a51bD", // Chainlink BEP20 - reliable volume
				"USDT": "0x55d398326f99059fF775485246999027B3197955",
			},
			PancakeswapPair: map[string]string{
				"WBNB-LINK": "0x824eb9faDFb377394430d2744fa7C42916DE3eCe",
				"LINK-USDT": "0x88f4BdE4a94cbE8B9318B2CAc59Fe55c68b7ff97",
				"USDT-WBNB": "0x16b9a82891338f9bA80E2D6970FddA79D1eb0daE",
			},
			BiswapPair: map[string]string{
				"WBNB-LINK": "0xaB4aCd6AD51dd665B0C0a7dE13d93b4650c5D4C5",
				"LINK-USDT": "0x42c92A3A24E3b7D03F0B9E2e8F0E8c1b6d9A3bCd",
				"USDT-WBNB": "0x8840C6252e2e86e545deFb6da98B2a0E26d8C1BA",
			},
			Priority:    9,
			TestAmounts: []float64{0.1, 0.3, 0.5},
		},

		// PRIORITY 10: XRP - Ripple, consistent volume
		{
			Name: "WBNB-XRP-USDT",
			Tokens: map[string]string{
				"WBNB": "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c",
				"XRP":  "0x1D2F0da169ceB9fC7B3144628dB156f3F6c60dBE", // XRP BEP20 - consistent volume
				"USDT": "0x55d398326f99059fF775485246999027B3197955",
			},
			PancakeswapPair: map[string]string{
				"WBNB-XRP":  "0x03F18135c44C64ebFdCBad8297fe5bDafdBbdd86",
				"XRP-USDT":  "0x8b303d5BbfBbf46F1a4d9741E491e06986894e18",
				"USDT-WBNB": "0x16b9a82891338f9bA80E2D6970FddA79D1eb0daE",
			},
			BiswapPair: map[string]string{
				"WBNB-XRP":  "0xC2eDBAF4Cb78dE86A6F6D3bAa6C11A8f92b07321",
				"XRP-USDT":  "0x5Bc9A2a8c37B9C12d4F52A6A4a9D8c7C8F0d9E3b",
				"USDT-WBNB": "0x8840C6252e2e86e545deFb6da98B2a0E26d8C1BA",
			},
			Priority:    10,
			TestAmounts: []float64{0.1, 0.3, 0.5},
		},
	}
}
