// config/config.go
package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Mainnet contract addresses for BSC
const (
	// Wrapped BNB
	WBNB = "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c"

	// PancakeSwap V2 addresses
	PancakeswapRouter  = "0x10ED43C718714eb63d5aA57B78B54704E256024E"
	PancakeswapFactory = "0xcA143Ce32Fe78f1f7019d7d551a6402fC5350c73"

	// Biswap addresses
	BiswapRouter  = "0x3a6d8cA21D1CF76F653A67577FA0D27453350dD8"
	BiswapFactory = "0x858E3312ed3A876947EA49d572A7C42DE08af7EE"

	// Common stablecoins and tokens on BSC
	USDT = "0x55d398326f99059fF775485246999027B3197955"
	BUSD = "0xe9e7CEA3DedcA5984780Bafc599bD69ADd087D56"
	USDC = "0x8AC76a51cc950d9822D68b83fE1Ad97B32Cd580d"
	BTCB = "0x7130d2A12B9BCbFAe4f2634d864A1Ee1Ce3Ead9c"
	ETH  = "0x2170Ed0880ac9A755fd29B2688956BD959F933F8"
	BSW  = "0x965F527D9159dCe6288a2219DB51fc6Eef120dD1"
)

// Config holds the application configuration
type Config struct {
	PrivateKey       string
	RpcURLs          []string
	FlashArbContract string
	GasLimit         uint64
	GasPrice         uint64
	MinProfit        float64
	CooldownPeriod   int
	MaxSlippage      float64
	Debug            bool
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	// Load .env file if it exists
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Validate required environment variables
	privateKey := getEnvRequired("PRIVATE_KEY")
	if len(privateKey) == 0 {
		log.Fatal("PRIVATE_KEY is required")
	}

	// Remove '0x' prefix if present
	if strings.HasPrefix(privateKey, "0x") {
		privateKey = privateKey[2:]
	}

	// Validate private key length
	if len(privateKey) != 64 {
		log.Fatal("PRIVATE_KEY must be 64 characters (32 bytes) hex string")
	}

	// Load RPC URLs
	rpcURLs := getRpcURLs()
	if len(rpcURLs) == 0 {
		log.Fatal("At least one BSC RPC URL is required")
	}

	// Parse optional configuration with defaults
	gasLimit := parseUint64Env("GAS_LIMIT", 600000)
	gasPrice := parseUint64Env("GAS_PRICE", 5000000000)  // 5 Gwei default
	minProfit := parseFloat64Env("MIN_PROFIT", 0.005)    // 0.5% default
	cooldownPeriod := parseIntEnv("COOLDOWN_PERIOD", 30) // 30 seconds default
	maxSlippage := parseFloat64Env("MAX_SLIPPAGE", 0.01) // 1% default
	debug := parseBoolEnv("DEBUG", false)

	// Flash arbitrage contract (optional)
	flashArbContract := getEnv("FLASH_ARB_CONTRACT", "")

	// Log configuration (without sensitive data)
	log.Printf("Configuration loaded:")
	log.Printf("- RPC URLs: %d endpoints", len(rpcURLs))
	log.Printf("- Gas Limit: %d", gasLimit)
	log.Printf("- Gas Price: %d wei (%.2f Gwei)", gasPrice, float64(gasPrice)/1e9)
	log.Printf("- Min Profit: %.2f%%", minProfit*100)
	log.Printf("- Cooldown Period: %d seconds", cooldownPeriod)
	log.Printf("- Max Slippage: %.2f%%", maxSlippage*100)
	log.Printf("- Debug Mode: %t", debug)
	if flashArbContract != "" {
		log.Printf("- Flash Arbitrage Contract: %s", flashArbContract)
	}

	return &Config{
		PrivateKey:       privateKey,
		RpcURLs:          rpcURLs,
		FlashArbContract: flashArbContract,
		GasLimit:         gasLimit,
		GasPrice:         gasPrice,
		MinProfit:        minProfit,
		CooldownPeriod:   cooldownPeriod,
		MaxSlippage:      maxSlippage,
		Debug:            debug,
	}
}

// getRpcURLs collects all RPC URLs from environment variables
func getRpcURLs() []string {
	var rpcURLs []string

	// Primary RPC URL
	if url := getEnv("BSC_RPC_URL", ""); url != "" {
		rpcURLs = append(rpcURLs, url)
	}

	// Additional RPC URLs (numbered)
	for i := 1; i <= 10; i++ {
		envKey := "BSC_RPC_URL" + strconv.Itoa(i)
		if url := getEnv(envKey, ""); url != "" {
			rpcURLs = append(rpcURLs, url)
		}
	}

	// Fallback to public RPC if no URLs provided
	if len(rpcURLs) == 0 {
		log.Println("Warning: No RPC URLs provided, using public BSC RPC (may be rate limited)")
		rpcURLs = append(rpcURLs, "https://bsc-dataseed.binance.org/")
	}

	return rpcURLs
}

// Helper functions for parsing environment variables

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvRequired(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Required environment variable %s is not set", key)
	}
	return value
}

func parseUint64Env(key string, defaultValue uint64) uint64 {
	str := getEnv(key, "")
	if str == "" {
		return defaultValue
	}

	value, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		log.Printf("Warning: Invalid value for %s (%s), using default: %d", key, str, defaultValue)
		return defaultValue
	}

	return value
}

func parseIntEnv(key string, defaultValue int) int {
	str := getEnv(key, "")
	if str == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(str)
	if err != nil {
		log.Printf("Warning: Invalid value for %s (%s), using default: %d", key, str, defaultValue)
		return defaultValue
	}

	return value
}

func parseFloat64Env(key string, defaultValue float64) float64 {
	str := getEnv(key, "")
	if str == "" {
		return defaultValue
	}

	value, err := strconv.ParseFloat(str, 64)
	if err != nil {
		log.Printf("Warning: Invalid value for %s (%s), using default: %.6f", key, str, defaultValue)
		return defaultValue
	}

	// Validate reasonable ranges
	switch key {
	case "MIN_PROFIT":
		if value < 0.001 || value > 0.1 { // 0.1% to 10%
			log.Printf("Warning: MIN_PROFIT should be between 0.001 and 0.1, got %.6f, using default", value)
			return defaultValue
		}
	case "MAX_SLIPPAGE":
		if value < 0.005 || value > 0.05 { // 0.5% to 5%
			log.Printf("Warning: MAX_SLIPPAGE should be between 0.005 and 0.05, got %.6f, using default", value)
			return defaultValue
		}
	}

	return value
}

func parseBoolEnv(key string, defaultValue bool) bool {
	str := getEnv(key, "")
	if str == "" {
		return defaultValue
	}

	value, err := strconv.ParseBool(str)
	if err != nil {
		log.Printf("Warning: Invalid boolean value for %s (%s), using default: %t", key, str, defaultValue)
		return defaultValue
	}

	return value
}

// ValidateConfig validates the configuration values
func (c *Config) ValidateConfig() error {
	if c.PrivateKey == "" {
		return fmt.Errorf("private key is required")
	}

	if len(c.RpcURLs) == 0 {
		return fmt.Errorf("at least one RPC URL is required")
	}

	if c.GasLimit < 100000 || c.GasLimit > 10000000 {
		return fmt.Errorf("gas limit should be between 100,000 and 10,000,000")
	}

	if c.MinProfit < 0.001 || c.MinProfit > 0.1 {
		return fmt.Errorf("minimum profit should be between 0.1%% and 10%%")
	}

	if c.MaxSlippage < 0.005 || c.MaxSlippage > 0.05 {
		return fmt.Errorf("max slippage should be between 0.5%% and 5%%")
	}

	if c.CooldownPeriod < 1 || c.CooldownPeriod > 300 {
		return fmt.Errorf("cooldown period should be between 1 and 300 seconds")
	}

	return nil
}
