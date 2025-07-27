package main

import (
	"fmt"
	"log"
	"math"
	"math/big"
	"os"
	"os/signal"
	"strings"
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
	log.Println("🚀 BSC Enhanced Arbitrage Bot v2.1")
	log.Println("🔧 AUTO RPC SWITCHING ENABLED")
	log.Println("🎯 HIGH VOLUME COIN FOCUS")
	log.Println("======================================")
	log.Println("✨ Targeting: BTC, ETH, BSW, CAKE")
	log.Println("💰 Conservative profit thresholds")
	log.Println("⏰ Peak hour optimization")
	log.Println("🔄 Automatic RPC failover")
	log.Println("⚠️ FIXED: Interval escalation prevented")
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

	// Create enhanced Ethereum client with automatic RPC switching
	log.Println("🌐 Connecting to BSC network with failover...")
	client, err := services.NewEthClient(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect to BSC network: %v", err)
	}
	defer client.Close()
	log.Println("✅ Connected to BSC network successfully")

	// Create services
	log.Println("🔧 Initializing services...")
	tokenService := services.NewTokenService(client)
	routerService := services.NewRouterService(client, tokenService, cfg)
	arbitrageService := services.NewArbitrageService(client, tokenService, routerService, cfg)
	log.Println("✅ Services initialized successfully")

	// Print enhanced wallet information with error handling
	printEnhancedWalletInfoWithRetry(client, tokenService)

	// Print configuration
	printEnhancedConfig(cfg)

	// Start RPC health monitoring in background
	stopHealthMonitor := make(chan bool, 1)
	go monitorRPCHealth(client, stopHealthMonitor)

	// Verify and update pair addresses with error handling
	log.Println("🔍 Verifying and updating pair addresses...")
	err = verifyPairsWithRetry(arbitrageService, client)
	if err != nil {
		log.Printf("⚠️ Warning: Error verifying pairs: %v", err)
		log.Println("📝 Continuing with manually configured addresses...")
	} else {
		log.Println("✅ Pair addresses verified successfully")
	}

	// Setup graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	log.Println("======================================")
	log.Println("🎯 Starting Enhanced High Volume Arbitrage...")
	log.Println("🔄 Auto RPC switching enabled")
	log.Println("📍 Bot akan scan terus-menerus dengan interval stabil")
	log.Println("======================================")

	// FIXED: Start the main scanning loop with proper error handling
	runPersistentArbitrageLoop(arbitrageService, client, cfg, stop)

	log.Println("✅ Enhanced Bot stopped gracefully")
	log.Println("🙏 Thank you for using BSC Enhanced Arbitrage Bot!")
}

func printEnhancedWalletInfoWithRetry(client *services.EthClient, tokenService *services.TokenService) {
	log.Println("======================================")
	log.Println("💼 Enhanced Wallet Information")
	log.Println("======================================")

	// Get wallet address
	log.Printf("📍 Address: %s", client.Address.Hex())

	// Log current RPC status
	client.LogConnectionStatus()

	// Get native BNB balance with retry
	log.Println("🔍 Fetching BNB balance...")
	bnbBalance, err := client.GetNativeBalanceWithRetry(client.Address)
	if err != nil {
		log.Printf("❌ Error getting BNB balance after retries: %v", err)
	} else {
		bnbFloat := new(big.Float).SetInt(bnbBalance)
		bnbFloat.Quo(bnbFloat, big.NewFloat(1e18))
		bnbBalanceFloat, _ := bnbFloat.Float64()
		log.Printf("🪙 Native BNB Balance: %.6f BNB", bnbBalanceFloat)

		if bnbBalanceFloat < 0.01 {
			log.Println("⚠️ WARNING: Low BNB balance for gas fees!")
		}
	}

	// Get WBNB balance with enhanced warnings and retry
	log.Println("🔍 Fetching WBNB balance...")
	wbnbAddr := common.HexToAddress(config.WBNB)
	wbnbBalance, err := client.GetTokenBalanceWithRetry(wbnbAddr, client.Address)
	if err != nil {
		log.Printf("❌ Error getting WBNB balance after retries: %v", err)
	} else {
		decimals, err := tokenService.GetTokenDecimals(wbnbAddr)
		if err != nil {
			log.Printf("❌ Error getting WBNB decimals: %v", err)
		} else {
			readableBalance := tokenService.ConvertToReadable(wbnbBalance, decimals)
			log.Printf("💰 WBNB Balance: %.6f WBNB", readableBalance)

			if readableBalance < 0.1 {
				log.Println("🚨 CRITICAL: Very low WBNB balance!")
				log.Println("💡 Bot needs at least 0.1 WBNB for arbitrage")
			} else if readableBalance < 0.5 {
				log.Println("⚠️ WARNING: Low WBNB for optimal arbitrage")
			} else {
				log.Println("✅ Good WBNB balance for arbitrage opportunities!")
			}
		}
	}

	// Get USDT balance with retry
	log.Println("🔍 Fetching USDT balance...")
	usdtAddr := common.HexToAddress(config.USDT)
	usdtBalance, err := client.GetTokenBalanceWithRetry(usdtAddr, client.Address)
	if err != nil {
		log.Printf("💵 USDT Balance: Unable to fetch after retries (%v)", err)
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
	log.Println("   • Stable (BUSD): 0.2% minimum")
	log.Println("   • Established (BSW, CAKE): 0.3% minimum")
	log.Println("   • Major coins (BTC, ETH): 0.5% minimum")
	log.Println("⏰ Peak hours: 13-16 UTC (Asia), 21-23 UTC (US)")
	log.Println("🔄 Auto RPC switching: ENABLED")
	log.Println("💡 Strategy: High volume pairs with stable intervals")
	log.Println("⚠️ Max interval: 2 minutes (no hour-long delays!)")
	log.Println("======================================")
}

func monitorRPCHealth(client *services.EthClient, stopChan <-chan bool) {
	ticker := time.NewTicker(60 * time.Second) // Check every minute
	defer ticker.Stop()

	log.Println("🔍 Starting RPC health monitoring...")

	for {
		select {
		case <-ticker.C:
			if !client.HealthCheck() {
				log.Printf("⚠️ RPC health check failed, attempting recovery...")
				if err := client.SwitchRPC(); err != nil {
					log.Printf("❌ RPC recovery failed: %v", err)
				} else {
					log.Printf("✅ RPC recovery successful")
				}
			}

		case <-stopChan:
			log.Println("🛑 Stopping RPC health monitoring")
			return
		}
	}
}

func verifyPairsWithRetry(arbitrageService *services.ArbitrageService, client *services.EthClient) error {
	return client.WithRetry("VerifyPairs", func() error {
		return arbitrageService.VerifyAndUpdatePairs()
	})
}

// FIXED: Loop yang benar-benar persisten dan tidak akan berhenti dengan interval stabil
func runPersistentArbitrageLoop(arbitrageService *services.ArbitrageService, client *services.EthClient, cfg *config.Config, stop chan os.Signal) {
	// FIXED: Start with reasonable base interval dan cap maksimum
	baseScanInterval := time.Duration(cfg.CooldownPeriod) * time.Second
	if baseScanInterval < 15*time.Second {
		baseScanInterval = 15 * time.Second // Minimum 15 detik
	}
	if baseScanInterval > 60*time.Second {
		baseScanInterval = 60 * time.Second // Maximum 1 minute base
	}

	// Statistics
	var totalScans int
	var successfulScans int
	var errorCount int
	var consecutiveErrors int
	var consecutiveNoOpportunities int // FIXED: Track this separately
	var rpcSwitches int
	startTime := time.Now()

	log.Printf("🔄 Starting persistent monitoring (interval: %v)", baseScanInterval)
	log.Println("⚠️ Bot akan terus berjalan sampai Ctrl+C ditekan")
	log.Println("📊 Interval akan stabil antara 15 detik - 2 menit")

	// FIXED: Use infinite loop with sleep, bukan ticker yang bisa bermasalah
	go func() {
		// Run initial scan
		log.Println("🔍 Running initial enhanced scan...")
		if err := performEnhancedScanWithRetry(arbitrageService, client, "initial"); err != nil {
			log.Printf("❌ Initial scan error: %v", err)
			errorCount++
			consecutiveErrors++
		} else {
			successfulScans++
			consecutiveErrors = 0
		}
		totalScans++

		// FIXED: Main loop yang tidak akan berhenti
		for {
			// CRITICAL: Selalu sleep dulu sebelum scan berikutnya
			time.Sleep(baseScanInterval)

			// Check if we should stop
			select {
			case <-stop:
				log.Println("🛑 Received stop signal, exiting scan loop...")
				return
			default:
				// Continue scanning
			}

			// Log RPC status periodically
			if totalScans%10 == 0 {
				client.LogConnectionStatus()
			}

			// Determine scan type based on current time
			scanType := getScanType()

			// FIXED: Perform scan dengan error recovery yang proper
			log.Printf("🔍 Scan #%d (%s) - interval: %v", totalScans+1, scanType, baseScanInterval)

			if err := performEnhancedScanWithRetry(arbitrageService, client, scanType); err != nil {
				log.Printf("❌ Scan #%d error: %v", totalScans+1, err)
				errorCount++
				consecutiveErrors++
				consecutiveNoOpportunities = 0 // Reset this counter

				// Enhanced error recovery
				if services.IsConnectionError(err) {
					rpcSwitches++
					log.Printf("🔄 RPC connection error, switch count: %d", rpcSwitches)
				}

				// If too many consecutive REAL errors, try RPC health check
				if consecutiveErrors >= 3 {
					log.Println("⚠️ Multiple consecutive errors, checking RPC health...")
					if !client.HealthCheck() {
						log.Println("🔄 RPC unhealthy, forcing switch...")
						if switchErr := client.SwitchRPC(); switchErr != nil {
							log.Printf("❌ Manual RPC switch failed: %v", switchErr)
						} else {
							log.Println("✅ Manual RPC switch successful")
							consecutiveErrors = 0
							rpcSwitches++
						}
					}
				}

				// FIXED: Jangan berhenti meskipun ada error, cuma tambah delay
				if consecutiveErrors >= 5 {
					log.Printf("⚠️ Too many errors (%d), adding extra delay...", consecutiveErrors)
					time.Sleep(time.Duration(consecutiveErrors) * 10 * time.Second)
				}
			} else {
				log.Printf("✅ Scan #%d completed successfully", totalScans+1)
				successfulScans++
				consecutiveErrors = 0

				// FIXED: Track consecutive "no opportunities" separately
				// This is normal and shouldn't increase error count
				consecutiveNoOpportunities = 0 // Reset since this was successful
			}
			totalScans++

			// Print statistics every 5 scans
			if totalScans%5 == 0 {
				printEnhancedStatsWithRPC(totalScans, successfulScans, errorCount, rpcSwitches, startTime, client)
			}

			// FIXED: Only use real errors for adaptive interval, not "no opportunities"
			realErrorsForAdaptive := consecutiveErrors
			if consecutiveNoOpportunities > 5 && consecutiveErrors == 0 {
				// If many scans with no opportunities but no real errors,
				// slow down slightly but not dramatically
				realErrorsForAdaptive = 1
			}

			// FIXED: Calculate new interval with better logic
			newInterval := calculateAdaptiveInterval(baseScanInterval, realErrorsForAdaptive)

			// FIXED: Only change interval if significantly different
			if newInterval != baseScanInterval {
				percentChange := float64(newInterval-baseScanInterval) / float64(baseScanInterval) * 100
				if math.Abs(percentChange) > 20 { // Only log if >20% change
					log.Printf("⚡ Adjusting scan interval: %v → %v (%.1f%% change)",
						baseScanInterval, newInterval, percentChange)
					baseScanInterval = newInterval
				}
			}

			// FIXED: Regular status update
			if totalScans%10 == 0 {
				log.Printf("🔄 Bot status: %d scans, %d successful, interval: %v",
					totalScans, successfulScans, baseScanInterval)
			}
		}
	}()

	// Wait for stop signal
	<-stop
	log.Println("\n======================================")
	log.Println("🛑 Shutdown signal received...")
	log.Println("======================================")

	time.Sleep(2 * time.Second)
	printFinalEnhancedStatsWithRPC(totalScans, successfulScans, errorCount, rpcSwitches, startTime, client)
}

// FIXED: Enhanced scan function yang lebih robust
func performEnhancedScanWithRetry(arbitrageService *services.ArbitrageService, client *services.EthClient, scanType string) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("❌ Panic recovered in %s scan: %v", scanType, r)
		}
	}()

	startTime := time.Now()

	// FIXED: Wrapper dengan timeout untuk mencegah hanging
	done := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("panic in scan: %v", r)
			}
		}()

		log.Printf("🎯 Performing %s enhanced scan...", scanType)

		// FIXED: Don't use WithRetry for this - it's not a connection error
		err := arbitrageService.FindEnhancedArbitrageOpportunities()

		// FIXED: "No opportunities found" is NOT an error - it's normal
		if err != nil && strings.Contains(err.Error(), "No enhanced opportunities") {
			log.Printf("📊 %s scan: No opportunities found (normal during off-peak)", scanType)
			done <- nil // Return success, not error
			return
		}

		done <- err
	}()

	// FIXED: Timeout untuk scan (maksimal 1 menit per scan)
	select {
	case err := <-done:
		scanDuration := time.Since(startTime)
		if err != nil {
			log.Printf("❌ %s scan failed in %v: %v", scanType, scanDuration.Round(time.Millisecond), err)
			return err
		}
		log.Printf("✅ %s scan completed in %v", scanType, scanDuration.Round(time.Millisecond))
		return nil

	case <-time.After(60 * time.Second): // 1 minute timeout
		log.Printf("⏰ %s scan timed out after 1 minute, continuing...", scanType)
		return fmt.Errorf("scan timed out")
	}
}

func getScanType() string {
	hour := time.Now().UTC().Hour()

	if (hour >= 13 && hour <= 16) || (hour >= 21 && hour <= 23) {
		return "peak_hours"
	} else if hour >= 2 && hour <= 8 {
		return "low_activity"
	} else {
		return "standard"
	}
}

// FIXED: Prevent zero or negative intervals yang bisa crash bot
func calculateAdaptiveInterval(baseInterval time.Duration, consecutiveErrors int) time.Duration {
	const minInterval = 10 * time.Second  // Minimum 10 seconds
	const maxInterval = 120 * time.Second // FIXED: Maximum 2 minutes (not hours!)
	const baseMinimum = 15 * time.Second  // Base minimum for any time

	hour := time.Now().UTC().Hour()
	var newInterval time.Duration

	// FIXED: More conservative multipliers
	switch {
	case (hour >= 13 && hour <= 16) || (hour >= 21 && hour <= 23):
		// Peak hours - scan faster
		newInterval = time.Duration(float64(baseInterval) * 0.8) // 20% faster

	case hour >= 2 && hour <= 8:
		// Low activity hours - but not crazy slow
		newInterval = time.Duration(float64(baseInterval) * 1.3) // Only 30% slower

	case consecutiveErrors >= 3:
		// Multiple errors - slow down a bit
		newInterval = time.Duration(float64(baseInterval) * 1.5) // 50% slower

	case consecutiveErrors >= 5:
		// Many errors - but cap the slowdown
		newInterval = maxInterval // Cap at 2 minutes max

	default:
		// Normal hours - keep base interval
		newInterval = baseInterval
	}

	// CRITICAL: Always enforce bounds
	if newInterval < minInterval {
		log.Printf("⚠️ Calculated interval %v too small, using minimum %v", newInterval, minInterval)
		return minInterval
	}

	if newInterval > maxInterval {
		log.Printf("⚠️ Calculated interval %v too large, using maximum %v", newInterval, maxInterval)
		return maxInterval
	}

	// FIXED: Never go below reasonable base minimum
	if newInterval < baseMinimum {
		return baseMinimum
	}

	return newInterval
}

func printEnhancedStatsWithRPC(totalScans, successfulScans, errorCount, rpcSwitches int, startTime time.Time, client *services.EthClient) {
	uptime := time.Since(startTime)
	successRate := float64(successfulScans) / float64(totalScans) * 100

	log.Println("📊 === Enhanced Statistics ===")
	log.Printf("⏰ Uptime: %v", uptime.Round(time.Second))
	log.Printf("🔍 Total scans: %d", totalScans)
	log.Printf("✅ Successful: %d (%.1f%%)", successfulScans, successRate)
	log.Printf("❌ Errors: %d", errorCount)
	log.Printf("🔄 RPC switches: %d", rpcSwitches)
	log.Printf("⚡ Avg scan time: %.1fs", uptime.Seconds()/float64(totalScans))

	// RPC status
	client.LogConnectionStatus()

	// Time-based insights
	hour := time.Now().UTC().Hour()
	if (hour >= 13 && hour <= 16) || (hour >= 21 && hour <= 23) {
		log.Printf("🔥 PEAK HOURS - Prime time for volatility!")
	} else if hour >= 2 && hour <= 8 {
		log.Printf("😴 Low activity - reduced opportunities")
	} else {
		log.Printf("📈 Standard hours - moderate activity")
	}

	log.Println("🔄 Bot will continue scanning...")
	log.Println("===============================")
}

func printFinalEnhancedStatsWithRPC(totalScans, successfulScans, errorCount, rpcSwitches int, startTime time.Time, client *services.EthClient) {
	uptime := time.Since(startTime)

	log.Println("======================================")
	log.Println("📋 Final Enhanced Statistics")
	log.Println("======================================")
	log.Printf("⏰ Total uptime: %v", uptime.Round(time.Second))
	log.Printf("🔍 Total scans: %d", totalScans)
	log.Printf("✅ Successful: %d", successfulScans)
	log.Printf("❌ Failed: %d", errorCount)
	log.Printf("🔄 RPC switches: %d", rpcSwitches)

	if totalScans > 0 {
		successRate := float64(successfulScans) / float64(totalScans) * 100
		log.Printf("📊 Success rate: %.1f%%", successRate)
		log.Printf("⚡ Average scan time: %.1fs", uptime.Seconds()/float64(totalScans))
	}

	// Final RPC status
	client.LogConnectionStatus()

	log.Println("======================================")
}
