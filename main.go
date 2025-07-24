// main.go
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
	// Setup log format with more detailed information
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("======================================")
	log.Println("Starting BSC Triangular Arbitrage Bot")
	log.Println("======================================")

	// Load and validate configuration
	cfg := config.LoadConfig()
	if err := cfg.ValidateConfig(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Initialize contract ABIs
	log.Println("Initializing contract ABIs...")
	err := contracts.Initialize()
	if err != nil {
		log.Fatalf("Failed to initialize contract ABIs: %v", err)
	}
	log.Println("✓ Contract ABIs initialized successfully")

	// Create Ethereum client with retry logic
	log.Println("Connecting to BSC network...")
	var client *services.EthClient
	var retries = 3
	for i := 0; i < retries; i++ {
		client, err = services.NewEthClient(cfg)
		if err != nil {
			log.Printf("Failed to create Ethereum client (attempt %d/%d): %v", i+1, retries, err)
			if i < retries-1 {
				time.Sleep(5 * time.Second)
				continue
			}
			log.Fatalf("Failed to connect after %d attempts", retries)
		}
		break
	}
	defer client.Close()
	log.Println("✓ Connected to BSC network successfully")

	// Create services
	log.Println("Initializing services...")
	tokenService := services.NewTokenService(client)
	routerService := services.NewRouterService(client, tokenService, cfg)
	arbitrageService := services.NewArbitrageService(client, tokenService, routerService, cfg)
	log.Println("✓ Services initialized successfully")

	// Print wallet information
	printWalletInfo(client, tokenService)

	// Verify and update pair addresses
	log.Println("Verifying and updating pair addresses...")
	err = arbitrageService.VerifyAndUpdatePairs()
	if err != nil {
		log.Printf("Warning: Error verifying pairs: %v", err)
		log.Println("Continuing with manually configured addresses...")
	} else {
		log.Println("✓ Pair addresses verified successfully")
	}

	// Setup graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start arbitrage loop in a goroutine
	done := make(chan bool, 1)

	log.Println("======================================")
	log.Println("Starting arbitrage monitoring...")
	log.Printf("Min profit threshold: %.2f%%", cfg.MinProfit*100)
	log.Printf("Scan interval: %d seconds", cfg.CooldownPeriod)
	log.Println("Press Ctrl+C to stop")
	log.Println("======================================")

	go runArbitrageLoop(arbitrageService, cfg, done, stop)

	// Wait for termination signal
	<-stop
	log.Println("\n======================================")
	log.Println("Shutdown signal received...")
	log.Println("======================================")

	// Signal the arbitrage loop to stop
	done <- true

	// Wait a bit for graceful shutdown
	time.Sleep(2 * time.Second)

	log.Println("✓ Bot stopped gracefully")
	log.Println("Thank you for using BSC Arbitrage Bot!")
}

// printWalletInfo prints comprehensive wallet information
func printWalletInfo(client *services.EthClient, tokenService *services.TokenService) {
	log.Println("======================================")
	log.Println("Wallet Information")
	log.Println("======================================")
	log.Printf("Address: %s", client.Address.Hex())

	// Get native BNB balance
	bnbBalance, err := client.Client.BalanceAt(context.Background(), client.Address, nil)
	if err != nil {
		log.Printf("Error getting BNB balance: %v", err)
	} else {
		bnbFloat := new(big.Float).SetInt(bnbBalance)
		bnbFloat.Quo(bnbFloat, big.NewFloat(1e18))
		bnbBalanceFloat, _ := bnbFloat.Float64()
		log.Printf("Native BNB Balance: %.6f BNB", bnbBalanceFloat)
	}

	// Get WBNB balance
	wbnbAddr := common.HexToAddress(config.WBNB)
	wbnbBalance, err := tokenService.GetTokenBalance(wbnbAddr, client.Address)
	if err != nil {
		log.Printf("Error getting WBNB balance: %v", err)
	} else {
		decimals, err := tokenService.GetTokenDecimals(wbnbAddr)
		if err != nil {
			log.Printf("Error getting WBNB decimals: %v", err)
		} else {
			readableBalance := tokenService.ConvertToReadable(wbnbBalance, decimals)
			log.Printf("WBNB Balance: %.6f WBNB", readableBalance)
		}
	}

	// Get USDT balance for reference
	usdtAddr := common.HexToAddress(config.USDT)
	usdtBalance, err := tokenService.GetTokenBalance(usdtAddr, client.Address)
	if err != nil {
		log.Printf("USDT Balance: Unable to fetch (%v)", err)
	} else {
		decimals, err := tokenService.GetTokenDecimals(usdtAddr)
		if err == nil {
			readableBalance := tokenService.ConvertToReadable(usdtBalance, decimals)
			log.Printf("USDT Balance: %.2f USDT", readableBalance)
		}
	}

	log.Println("======================================")
}

// runArbitrageLoop continuously scans for and executes arbitrage opportunities
func runArbitrageLoop(arbitrageService *services.ArbitrageService, cfg *config.Config, done chan bool, stop chan os.Signal) {
	ticker := time.NewTicker(time.Duration(cfg.CooldownPeriod) * time.Second)
	defer ticker.Stop()

	// Statistics
	var totalScans int
	var successfulScans int
	var errorCount int
	var consecutiveErrors int
	startTime := time.Now()

	// Run initial scan
	log.Println("Running initial arbitrage scan...")
	if err := performArbitrageScan(arbitrageService); err != nil {
		log.Printf("Initial scan error: %v", err)
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
			// Perform arbitrage scan
			if err := performArbitrageScan(arbitrageService); err != nil {
				log.Printf("Error in arbitrage scan: %v", err)
				errorCount++
				consecutiveErrors++

				// If too many consecutive errors, try switching RPC
				if consecutiveErrors >= 3 {
					log.Println("Multiple consecutive errors, attempting to switch RPC...")
					if switchErr := arbitrageService.Client.SwitchRPC(); switchErr != nil {
						log.Printf("Failed to switch RPC: %v", switchErr)
					} else {
						log.Println("✓ Successfully switched to backup RPC")
						consecutiveErrors = 0
					}
				}
			} else {
				successfulScans++
				consecutiveErrors = 0
			}
			totalScans++

			// Print statistics every 10 scans
			if totalScans%10 == 0 {
				printStats(totalScans, successfulScans, errorCount, startTime)
			}

		case <-done:
			log.Println("Stopping arbitrage loop...")
			printFinalStats(totalScans, successfulScans, errorCount, startTime)
			return
		}
	}
}

// performArbitrageScan performs a single arbitrage scan with error handling
func performArbitrageScan(arbitrageService *services.ArbitrageService) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic recovered in arbitrage scan: %v", r)
		}
	}()

	return arbitrageService.FindArbitrageOpportunities()
}

// printStats prints current statistics
func printStats(totalScans, successfulScans, errorCount int, startTime time.Time) {
	uptime := time.Since(startTime)
	successRate := float64(successfulScans) / float64(totalScans) * 100

	log.Println("--- Statistics ---")
	log.Printf("Uptime: %v", uptime.Round(time.Second))
	log.Printf("Total scans: %d", totalScans)
	log.Printf("Successful scans: %d (%.1f%%)", successfulScans, successRate)
	log.Printf("Errors: %d", errorCount)
	log.Printf("Average scan interval: %.1fs", uptime.Seconds()/float64(totalScans))
	log.Println("-----------------")
}

// printFinalStats prints final statistics before shutdown
func printFinalStats(totalScans, successfulScans, errorCount int, startTime time.Time) {
	uptime := time.Since(startTime)

	log.Println("======================================")
	log.Println("Final Statistics")
	log.Println("======================================")
	log.Printf("Total uptime: %v", uptime.Round(time.Second))
	log.Printf("Total scans performed: %d", totalScans)
	log.Printf("Successful scans: %d", successfulScans)
	log.Printf("Failed scans: %d", errorCount)
	if totalScans > 0 {
		successRate := float64(successfulScans) / float64(totalScans) * 100
		log.Printf("Success rate: %.1f%%", successRate)
		log.Printf("Average scan time: %.1fs", uptime.Seconds()/float64(totalScans))
	}
	log.Println("======================================")
}
