// main.go - SIMPLE FIXED VERSION
package main

import (
	"context"
	"log"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"arbitrage-bot/config"
	"arbitrage-bot/contracts"
	"arbitrage-bot/services"
)

func main() {
	// Enhanced log format
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("======================================")
	log.Println("🚀 BSC Enhanced Arbitrage Bot v2.0")
	log.Println("🎯 MEME COIN FOCUS")
	log.Println("======================================")
	log.Println("✨ Targeting: SHIB, DOGE, FLOKI, SAFEMOON")
	log.Println("💰 Higher profit thresholds for volatiles")
	log.Println("⏰ Peak hour optimization")
	log.Println("======================================")

	// Load and validate configuration
	cfg := config.LoadConfig()
	if err := cfg.ValidateConfig(); err != nil {
		log.Fatalf("❌ Invalid configuration: %v", err)
	}

	// Initialize contract ABIs
	log.Println("🔧 Initializing contract ABIs...")
	err := contracts.Initialize()
	if err != nil {
		log.Fatalf("❌ Failed to initialize contract ABIs: %v", err)
	}
	log.Println("✅ Contract ABIs initialized successfully")

	// Create Ethereum client with retry logic
	log.Println("🌐 Connecting to BSC network...")
	var client *services.EthClient
	var retries = 3
	for i := 0; i < retries; i++ {
		client, err = services.NewEthClient(cfg)
		if err != nil {
			log.Printf("⚠️ Failed to create Ethereum client (attempt %d/%d): %v", i+1, retries, err)
			if i < retries-1 {
				time.Sleep(5 * time.Second)
				continue
			}
			log.Fatalf("❌ Failed to connect after %d attempts", retries)
		}
		break
	}
	defer client.Close()
	log.Println("✅ Connected to BSC network successfully")

	// Create services
	log.Println("🔧 Initializing services...")
	tokenService := services.NewTokenService(client)
	routerService := services.NewRouterService(client, tokenService, cfg)
	arbitrageService := services.NewArbitrageService(client, tokenService, routerService, cfg)
	log.Println("✅ Services initialized successfully")

	// Print enhanced wallet information
	printEnhancedWalletInfo(client, tokenService)

	// Print configuration
	printEnhancedConfig(cfg)

	// Verify and update pair addresses
	log.Println("🔍 Verifying and updating pair addresses...")
	err = arbitrageService.VerifyAndUpdatePairs()
	if err != nil {
		log.Printf("⚠️ Warning: Error verifying pairs: %v", err)
		log.Println("📝 Continuing with manually configured addresses...")
	} else {
		log.Println("✅ Pair addresses verified successfully")
	}

	// Setup graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start enhanced arbitrage loop
	done := make(chan bool, 1)

	log.Println("======================================")
	log.Println("🎯 Starting Enhanced Meme Coin Arbitrage...")
	log.Println("======================================")

	go runEnhancedArbitrageLoop(arbitrageService, cfg, done, stop)

	// Wait for termination signal
	<-stop
	log.Println("\n======================================")
	log.Println("🛑 Shutdown signal received...")
	log.Println("======================================")

	// Signal the arbitrage loop to stop
	done <- true

	// Wait a bit for graceful shutdown
	time.Sleep(2 * time.Second)

	log.Println("✅ Enhanced Bot stopped gracefully")
	log.Println("🙏 Thank you for using BSC Enhanced Arbitrage Bot!")
}

func printEnhancedWalletInfo(client *services.EthClient, tokenService *services.TokenService) {
	log.Println("======================================")
	log.Println("💼 Enhanced Wallet Information")
	log.Println("======================================")
	log.Printf("📍 Address: %s", client.Address.Hex())

	// Get native BNB balance
	bnbBalance, err := client.Client.BalanceAt(context.Background(), client.Address, nil)
	if err != nil {
		log.Printf("❌ Error getting BNB balance: %v", err)
	} else {
		bnbFloat := new(big.Float).SetInt(bnbBalance)
		bnbFloat.Quo(bnbFloat, big.NewFloat(1e18))
		bnbBalanceFloat, _ := bnbFloat.Float64()
		log.Printf("🪙 Native BNB Balance: %.6f BNB", bnbBalanceFloat)

		if bnbBalanceFloat < 0.01 {
			log.Println("⚠️ WARNING: Low BNB balance for gas fees!")
		}
	}

	// Get WBNB balance with enhanced warnings
	wbnbAddr := common.HexToAddress(config.WBNB)
	wbnbBalance, err := tokenService.GetTokenBalance(wbnbAddr, client.Address)
	if err != nil {
		log.Printf("❌ Error getting WBNB balance: %v", err)
	} else {
		decimals, err := tokenService.GetTokenDecimals(wbnbAddr)
		if err != nil {
			log.Printf("❌ Error getting WBNB decimals: %v", err)
		} else {
			readableBalance := tokenService.ConvertToReadable(wbnbBalance, decimals)
			log.Printf("💰 WBNB Balance: %.6f WBNB", readableBalance)

			if readableBalance < 0.1 {
				log.Println("🚨 CRITICAL: Very low WBNB balance!")
				log.Println("💡 Meme coin arbitrage needs at least 1 WBNB")
			} else if readableBalance < 0.5 {
				log.Println("⚠️ WARNING: Low WBNB for meme coin arbitrage")
			} else if readableBalance >= 1.0 {
				log.Println("✅ Good WBNB balance for meme coin opportunities!")
			}
		}
	}

	// Get USDT balance
	usdtAddr := common.HexToAddress(config.USDT)
	usdtBalance, err := tokenService.GetTokenBalance(usdtAddr, client.Address)
	if err != nil {
		log.Printf("💵 USDT Balance: Unable to fetch (%v)", err)
	} else {
		decimals, err := tokenService.GetTokenDecimals(usdtAddr)
		if err == nil {
			readableBalance := tokenService.ConvertToReadable(usdtBalance, decimals)
			log.Printf("💵 USDT Balance: %.2f USDT", readableBalance)
		}
	}

	log.Println("======================================")
}

func printEnhancedConfig(cfg *config.Config) {
	log.Println("======================================")
	log.Println("⚙️ Enhanced Configuration")
	log.Println("======================================")
	log.Printf("💰 Base min profit: %.2f%%", cfg.MinProfit*100)
	log.Printf("⏰ Base scan interval: %d seconds", cfg.CooldownPeriod)
	log.Println("🎯 Enhanced profit targets:")
	log.Println("   • Meme coins (SHIB, DOGE): 0.5% minimum")
	log.Println("   • Volatile (BSW): 0.3% minimum")
	log.Println("   • Established (CAKE): 0.2% minimum")
	log.Println("   • Stable (BUSD): 0.1% minimum")
	log.Println("⏰ Peak hours: 13-16 UTC (Asia), 21-23 UTC (US)")
	log.Println("💡 Strategy: Target meme coin volatility for higher profits")
	log.Println("======================================")
}

func runEnhancedArbitrageLoop(arbitrageService *services.ArbitrageService, cfg *config.Config, done chan bool, stop chan os.Signal) {
	// Enhanced timing
	baseScanInterval := time.Duration(cfg.CooldownPeriod) * time.Second
	if baseScanInterval < 15*time.Second {
		baseScanInterval = 15 * time.Second // Minimum for meme coin scanning
	}

	ticker := time.NewTicker(baseScanInterval)
	defer ticker.Stop()

	// Enhanced statistics
	var totalScans int
	var successfulScans int
	var errorCount int
	var consecutiveErrors int
	startTime := time.Now()

	log.Printf("🔄 Enhanced monitoring started (interval: %v)", baseScanInterval)
	log.Println("🎯 Focusing on meme coins for higher profit opportunities")

	// Run initial enhanced scan
	log.Println("🔍 Running initial enhanced scan...")
	if err := performEnhancedScan(arbitrageService, "initial"); err != nil {
		log.Printf("❌ Initial scan error: %v", err)
		errorCount++
		consecutiveErrors++
	} else {
		successfulScans++
		consecutiveErrors = 0
	}
	totalScans++

	for {
		select {
		case <-ticker.C:
			// Determine scan type based on current time
			scanType := getScanType()

			// Perform enhanced scan
			if err := performEnhancedScan(arbitrageService, scanType); err != nil {
				log.Printf("❌ Enhanced scan error: %v", err)
				errorCount++
				consecutiveErrors++

				// Enhanced error recovery
				if consecutiveErrors >= 3 {
					log.Println("⚠️ Multiple errors, attempting recovery...")
					if switchErr := arbitrageService.Client.SwitchRPC(); switchErr != nil {
						log.Printf("❌ RPC switch failed: %v", switchErr)
					} else {
						log.Println("✅ Successfully switched to backup RPC")
						consecutiveErrors = 0
					}
				}
			} else {
				successfulScans++
				consecutiveErrors = 0
			}
			totalScans++

			// Print enhanced statistics every 5 scans
			if totalScans%5 == 0 {
				printEnhancedStats(totalScans, successfulScans, errorCount, startTime)
			}

			// Adaptive scan interval based on time and performance
			newInterval := calculateAdaptiveInterval(baseScanInterval, consecutiveErrors)
			if newInterval != baseScanInterval {
				log.Printf("⚡ Adjusting scan interval: %v → %v", baseScanInterval, newInterval)
				ticker.Reset(newInterval)
				baseScanInterval = newInterval
			}

		case <-done:
			log.Println("🛑 Stopping enhanced arbitrage loop...")
			printFinalEnhancedStats(totalScans, successfulScans, errorCount, startTime)
			return
		}
	}
}

func performEnhancedScan(arbitrageService *services.ArbitrageService, scanType string) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("❌ Panic in %s scan: %v", scanType, r)
		}
	}()

	startTime := time.Now()

	// Use enhanced arbitrage method
	log.Printf("🎯 %s enhanced scan (meme focus)...", scanType)
	err := arbitrageService.FindEnhancedArbitrageOpportunities()

	scanDuration := time.Since(startTime)

	if err != nil {
		log.Printf("❌ Enhanced scan failed in %v: %v", scanDuration, err)
		return err
	}

	log.Printf("✅ Enhanced scan completed in %v", scanDuration.Truncate(time.Millisecond))
	return nil
}

func getScanType() string {
	hour := time.Now().UTC().Hour()

	if (hour >= 13 && hour <= 16) || (hour >= 21 && hour <= 23) {
		return "peak_hours"
	} else if hour >= 2 && hour <= 6 {
		return "low_activity"
	} else {
		return "standard"
	}
}

func calculateAdaptiveInterval(baseInterval time.Duration, consecutiveErrors int) time.Duration {
	hour := time.Now().UTC().Hour()

	// Peak hours - scan faster for meme coin opportunities
	if (hour >= 13 && hour <= 16) || (hour >= 21 && hour <= 23) {
		return time.Duration(float64(baseInterval) * 0.7) // 30% faster
	}

	// Low activity hours - scan slower
	if hour >= 2 && hour <= 6 {
		return time.Duration(float64(baseInterval) * 2.0) // 100% slower
	}

	// If errors, slow down
	if consecutiveErrors > 1 {
		return time.Duration(float64(baseInterval) * 1.5) // 50% slower
	}

	return baseInterval
}

func printEnhancedStats(totalScans, successfulScans, errorCount int, startTime time.Time) {
	uptime := time.Since(startTime)
	successRate := float64(successfulScans) / float64(totalScans) * 100

	log.Println("📊 === Enhanced Statistics ===")
	log.Printf("⏰ Uptime: %v", uptime.Round(time.Second))
	log.Printf("🔍 Total scans: %d", totalScans)
	log.Printf("✅ Successful: %d (%.1f%%)", successfulScans, successRate)
	log.Printf("❌ Errors: %d", errorCount)
	log.Printf("⚡ Avg scan time: %.1fs", uptime.Seconds()/float64(totalScans))

	// Time-based insights
	hour := time.Now().UTC().Hour()
	if (hour >= 13 && hour <= 16) || (hour >= 21 && hour <= 23) {
		log.Printf("🔥 PEAK HOURS - Prime time for meme volatility!")
	} else if hour >= 2 && hour <= 6 {
		log.Printf("😴 Low activity - meme coins less volatile")
	} else {
		log.Printf("📈 Standard hours - moderate activity")
	}

	log.Println("===============================")
}

func printFinalEnhancedStats(totalScans, successfulScans, errorCount int, startTime time.Time) {
	uptime := time.Since(startTime)

	log.Println("======================================")
	log.Println("📋 Final Enhanced Statistics")
	log.Println("======================================")
	log.Printf("⏰ Total uptime: %v", uptime.Round(time.Second))
	log.Printf("🔍 Total scans: %d", totalScans)
	log.Printf("✅ Successful: %d", successfulScans)
	log.Printf("❌ Failed: %d", errorCount)

	if totalScans > 0 {
		successRate := float64(successfulScans) / float64(totalScans) * 100
		log.Printf("📊 Success rate: %.1f%%", successRate)
		log.Printf("⚡ Average scan time: %.1fs", uptime.Seconds()/float64(totalScans))
	}

	log.Println("======================================")
}
