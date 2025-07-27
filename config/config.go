// config/config.go - Enhanced config with multiple RPC support
package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all configuration values
type Config struct {
	// Wallet
	PrivateKey string

	// RPC URLs (multiple for failover)
	BSCRPCURL  string
	BSCRPCURL1 string
	BSCRPCURL2 string
	BSCRPCURL3 string
	BSCRPCURL4 string
	BSCRPCURL5 string
	BSCRPCURL6 string
	BSCRPCURL7 string
	BSCRPCURL8 string

	// Contracts
	FlashArbContract string

	// Gas settings
	GasLimit uint64
	GasPrice int64

	// Trading parameters
	MinProfit      float64
	MaxSlippage    float64
	CooldownPeriod int

	// Debug mode
	Debug bool
}

// Token addresses (constants)
const (
	WBNB = "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c"
	USDT = "0x55d398326f99059fF775485246999027B3197955"
	BUSD = "0xe9e7cea3dedca5984780bafc599bd69add087d56"
	CAKE = "0x0E09FaBB73Bd3Ade0a17ECC321fD13a19e81cE82"
	DOGE = "0xbA2aE424d960c26247Dd6c32edC70B295c744C43"
	SHIB = "0x2859e4544C4bB03966803b044A93563Bd2D0DD4D"

	// DEX Routers
	PancakeswapRouter = "0x10ED43C718714eb63d5aA57B78B54704E256024E"
	BiswapRouter      = "0x3a6d8cA21D1CF76F653A67577FA0D27453350dD8"

	// DEX Factories
	PancakeswapFactory = "0xcA143Ce32Fe78f1f7019d7d551a6402fC5350c73"
	BiswapFactory      = "0x858E3312ed3A876947EA49d572A7C42DE08af7EE"
)

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("📝 No .env file found, using environment variables")
	}

	// Create config with defaults
	cfg := &Config{
		// Default values
		GasLimit:       600000,
		GasPrice:       5000000000, // 5 Gwei
		MinProfit:      0.005,      // 0.5%
		MaxSlippage:    0.02,       // 2%
		CooldownPeriod: 30,         // 30 seconds
		Debug:          false,
	}

	// Load required values
	cfg.PrivateKey = getEnvRequired("PRIVATE_KEY")

	// Load RPC URLs - check both BSC_RPC_URL and BSCRPCURL for compatibility
	cfg.BSCRPCURL = getEnv("BSC_RPC_URL", getEnv("BSCRPCURL", ""))
	cfg.BSCRPCURL1 = getEnv("BSC_RPC_URL1", getEnv("BSCRPCURL1", ""))
	cfg.BSCRPCURL2 = getEnv("BSC_RPC_URL2", getEnv("BSCRPCURL2", ""))
	cfg.BSCRPCURL3 = getEnv("BSC_RPC_URL3", getEnv("BSCRPCURL3", ""))
	cfg.BSCRPCURL4 = getEnv("BSC_RPC_URL4", getEnv("BSCRPCURL4", ""))
	cfg.BSCRPCURL5 = getEnv("BSC_RPC_URL5", getEnv("BSCRPCURL5", ""))
	cfg.BSCRPCURL6 = getEnv("BSC_RPC_URL6", getEnv("BSCRPCURL6", ""))
	cfg.BSCRPCURL7 = getEnv("BSC_RPC_URL7", getEnv("BSCRPCURL7", ""))
	cfg.BSCRPCURL8 = getEnv("BSC_RPC_URL8", getEnv("BSCRPCURL8", ""))

	// Log configured RPC count
	rpcCount := cfg.countConfiguredRPCs()
	log.Printf("🌐 Configured %d RPC endpoints for failover", rpcCount)

	// Load optional contract
	cfg.FlashArbContract = getEnv("FLASH_ARB_CONTRACT", "")

	// Load gas settings
	if gasLimit := getEnv("GAS_LIMIT", ""); gasLimit != "" {
		if parsed, err := strconv.ParseUint(gasLimit, 10, 64); err == nil {
			cfg.GasLimit = parsed
		}
	}

	if gasPrice := getEnv("GAS_PRICE", ""); gasPrice != "" {
		if parsed, err := strconv.ParseInt(gasPrice, 10, 64); err == nil {
			cfg.GasPrice = parsed
		}
	}

	// Load trading parameters
	if minProfit := getEnv("MIN_PROFIT", ""); minProfit != "" {
		if parsed, err := strconv.ParseFloat(minProfit, 64); err == nil {
			cfg.MinProfit = parsed
		}
	}

	if maxSlippage := getEnv("MAX_SLIPPAGE", ""); maxSlippage != "" {
		if parsed, err := strconv.ParseFloat(maxSlippage, 64); err == nil {
			cfg.MaxSlippage = parsed
		}
	}

	if cooldown := getEnv("COOLDOWN_PERIOD", ""); cooldown != "" {
		if parsed, err := strconv.Atoi(cooldown); err == nil {
			cfg.CooldownPeriod = parsed
		}
	}

	// Load debug flag
	if debug := getEnv("DEBUG", ""); debug != "" {
		cfg.Debug = strings.ToLower(debug) == "true"
	}

	return cfg
}

// ValidateConfig validates the configuration
func (c *Config) ValidateConfig() error {
	var errors []string

	// Validate private key
	if c.PrivateKey == "" {
		errors = append(errors, "PRIVATE_KEY is required")
	} else if len(c.PrivateKey) != 64 {
		errors = append(errors, "PRIVATE_KEY must be 64 characters (without 0x prefix)")
	}

	// Validate at least one RPC URL
	if c.countConfiguredRPCs() == 0 {
		errors = append(errors, "at least one BSC_RPC_URL must be configured")
	}

	// Validate gas settings
	if c.GasLimit < 21000 {
		errors = append(errors, "GAS_LIMIT must be at least 21000")
	}

	if c.GasPrice < 1000000000 { // 1 Gwei minimum
		errors = append(errors, "GAS_PRICE must be at least 1 Gwei (1000000000)")
	}

	// Validate trading parameters
	if c.MinProfit < 0.001 || c.MinProfit > 0.1 {
		errors = append(errors, "MIN_PROFIT must be between 0.001 (0.1%) and 0.1 (10%)")
	}

	if c.MaxSlippage < 0.005 || c.MaxSlippage > 0.1 {
		errors = append(errors, "MAX_SLIPPAGE must be between 0.005 (0.5%) and 0.1 (10%)")
	}

	if c.CooldownPeriod < 5 || c.CooldownPeriod > 300 {
		errors = append(errors, "COOLDOWN_PERIOD must be between 5 and 300 seconds")
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

// countConfiguredRPCs counts how many RPC URLs are configured
func (c *Config) countConfiguredRPCs() int {
	count := 0
	rpcs := []string{
		c.BSCRPCURL, c.BSCRPCURL1, c.BSCRPCURL2, c.BSCRPCURL3,
		c.BSCRPCURL4, c.BSCRPCURL5, c.BSCRPCURL6, c.BSCRPCURL7, c.BSCRPCURL8,
	}

	for _, rpc := range rpcs {
		if rpc != "" {
			count++
		}
	}

	return count
}

// GetAllRPCURLs returns all configured RPC URLs
func (c *Config) GetAllRPCURLs() []string {
	var urls []string
	rpcs := []string{
		c.BSCRPCURL, c.BSCRPCURL1, c.BSCRPCURL2, c.BSCRPCURL3,
		c.BSCRPCURL4, c.BSCRPCURL5, c.BSCRPCURL6, c.BSCRPCURL7, c.BSCRPCURL8,
	}

	for _, rpc := range rpcs {
		if rpc != "" {
			urls = append(urls, rpc)
		}
	}

	return urls
}

// LogConfiguration prints the current configuration (without sensitive data)
func (c *Config) LogConfiguration() {
	log.Println("======================================")
	log.Println("⚙️ Configuration Summary")
	log.Println("======================================")
	log.Printf("🌐 RPC endpoints: %d configured", c.countConfiguredRPCs())
	log.Printf("⛽ Gas limit: %d", c.GasLimit)
	log.Printf("💰 Gas price: %.2f Gwei", float64(c.GasPrice)/1e9)
	log.Printf("📊 Min profit: %.2f%%", c.MinProfit*100)
	log.Printf("🎯 Max slippage: %.2f%%", c.MaxSlippage*100)
	log.Printf("⏰ Scan interval: %d seconds", c.CooldownPeriod)
	log.Printf("🔍 Debug mode: %v", c.Debug)

	if c.FlashArbContract != "" {
		log.Printf("⚡ Flash contract: %s", c.FlashArbContract)
	} else {
		log.Printf("⚡ Flash contract: Not configured (manual arbitrage only)")
	}

	log.Println("======================================")
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvRequired(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("❌ Required environment variable %s is not set", key)
	}
	return value
}
