// services/client.go - Enhanced version with automatic RPC switching
package services

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"arbitrage-bot/config"
)

// EthClient wraps ethereum client with enhanced RPC management
type EthClient struct {
	Client     *ethclient.Client
	Address    common.Address
	PrivateKey *ecdsa.PrivateKey
	Auth       *bind.TransactOpts

	// RPC management
	currentRPC   string
	rpcEndpoints []string
	rpcIndex     int
	failedRPCs   map[string]time.Time
	mu           sync.RWMutex

	// Connection health
	lastHealthCheck time.Time
	isHealthy       bool
}

// NewEthClient creates a new Ethereum client with RPC failover
func NewEthClient(cfg *config.Config) (*EthClient, error) {
	// Collect all RPC endpoints from config
	rpcEndpoints := collectRPCEndpoints(cfg)
	if len(rpcEndpoints) == 0 {
		return nil, fmt.Errorf("no RPC endpoints configured")
	}

	log.Printf("🌐 Found %d RPC endpoints for failover", len(rpcEndpoints))

	// Parse private key
	privateKey, err := crypto.HexToECDSA(cfg.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %v", err)
	}

	// Get address from private key
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("error casting public key to ECDSA")
	}
	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Create EthClient instance
	ethClient := &EthClient{
		Address:      address,
		PrivateKey:   privateKey,
		rpcEndpoints: rpcEndpoints,
		rpcIndex:     0,
		failedRPCs:   make(map[string]time.Time),
	}

	// Try to connect to first working RPC
	err = ethClient.connectToWorkingRPC()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to any RPC endpoint: %v", err)
	}

	// Setup transaction auth
	err = ethClient.setupAuth()
	if err != nil {
		return nil, fmt.Errorf("failed to setup transaction auth: %v", err)
	}

	log.Printf("✅ Connected to BSC via: %s", getShortRPCName(ethClient.currentRPC))
	return ethClient, nil
}

// collectRPCEndpoints extracts all RPC URLs from config
func collectRPCEndpoints(cfg *config.Config) []string {
	var endpoints []string

	// Check different possible field names in your config
	configRPCs := []string{
		cfg.BSCRPCURL,  // Primary field
		cfg.BSCRPCURL1, // Backup fields
		cfg.BSCRPCURL2,
		cfg.BSCRPCURL3,
		cfg.BSCRPCURL4,
		cfg.BSCRPCURL5,
		cfg.BSCRPCURL6,
		cfg.BSCRPCURL7,
		cfg.BSCRPCURL8,
	}

	// Add configured RPCs
	for _, rpc := range configRPCs {
		if rpc != "" && !contains(endpoints, rpc) {
			endpoints = append(endpoints, rpc)
		}
	}

	// Add fallback public RPCs if none configured
	if len(endpoints) == 0 {
		log.Println("⚠️ No RPC configured in config, using fallback public endpoints")
		endpoints = []string{
			"https://bsc-dataseed1.defibit.io/",
			"https://bsc-dataseed1.ninicoin.io/",
			"https://bsc-dataseed2.defibit.io/",
			"https://rpc.ankr.com/bsc",
			"https://bsc.rpc.blxrbdn.com/",
			"https://bsc-dataseed3.defibit.io/",
			"https://bsc-dataseed4.defibit.io/",
			"https://bscrpc.com",
		}
	}

	return endpoints
}

// connectToWorkingRPC tries to connect to the first working RPC
func (e *EthClient) connectToWorkingRPC() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Clean up expired failed RPCs (retry after 5 minutes)
	for rpc, failTime := range e.failedRPCs {
		if time.Since(failTime) > 5*time.Minute {
			delete(e.failedRPCs, rpc)
			log.Printf("🔄 RPC %s eligible for retry", getShortRPCName(rpc))
		}
	}

	var lastErr error
	attemptsCount := 0

	// Try each RPC endpoint
	for i := 0; i < len(e.rpcEndpoints); i++ {
		currentIndex := (e.rpcIndex + i) % len(e.rpcEndpoints)
		rpcURL := e.rpcEndpoints[currentIndex]

		// Skip recently failed RPCs
		if _, failed := e.failedRPCs[rpcURL]; failed {
			continue
		}

		attemptsCount++
		log.Printf("🔗 Attempting connection to %s...", getShortRPCName(rpcURL))

		client, err := ethclient.Dial(rpcURL)
		if err != nil {
			log.Printf("❌ Failed to connect to %s: %v", getShortRPCName(rpcURL), err)
			e.failedRPCs[rpcURL] = time.Now()
			lastErr = err
			continue
		}

		// Test the connection with a simple call
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		_, err = client.NetworkID(ctx)
		cancel()

		if err != nil {
			log.Printf("❌ RPC %s failed health check: %v", getShortRPCName(rpcURL), err)
			client.Close()
			e.failedRPCs[rpcURL] = time.Now()
			lastErr = err
			continue
		}

		// Success! Update client
		if e.Client != nil {
			e.Client.Close()
		}

		e.Client = client
		e.currentRPC = rpcURL
		e.rpcIndex = currentIndex
		e.isHealthy = true
		e.lastHealthCheck = time.Now()

		log.Printf("✅ Successfully connected to %s", getShortRPCName(rpcURL))
		return nil
	}

	if attemptsCount == 0 {
		return fmt.Errorf("all RPC endpoints are marked as failed, will retry in 5 minutes")
	}

	return fmt.Errorf("failed to connect to any RPC endpoint after %d attempts. Last error: %v", attemptsCount, lastErr)
}

// SwitchRPC switches to the next available RPC endpoint
func (e *EthClient) SwitchRPC() error {
	log.Printf("🔄 Switching RPC from %s due to connection issues...", getShortRPCName(e.currentRPC))

	// Mark current RPC as failed
	e.mu.Lock()
	e.failedRPCs[e.currentRPC] = time.Now()
	e.isHealthy = false
	e.mu.Unlock()

	// Try to connect to next working RPC
	err := e.connectToWorkingRPC()
	if err != nil {
		return fmt.Errorf("RPC switch failed: %v", err)
	}

	// Update auth for new connection
	err = e.setupAuth()
	if err != nil {
		return fmt.Errorf("failed to setup auth after RPC switch: %v", err)
	}

	log.Printf("✅ Successfully switched to %s", getShortRPCName(e.currentRPC))
	return nil
}

// setupAuth creates transaction auth for the current connection
func (e *EthClient) setupAuth() error {
	chainID := big.NewInt(56) // BSC chain ID
	auth, err := bind.NewKeyedTransactorWithChainID(e.PrivateKey, chainID)
	if err != nil {
		return err
	}

	auth.GasLimit = uint64(600000)         // Default gas limit
	auth.GasPrice = big.NewInt(5000000000) // 5 Gwei

	e.Auth = auth
	return nil
}

// HealthCheck checks if current RPC is still working
func (e *EthClient) HealthCheck() bool {
	e.mu.RLock()
	lastCheck := e.lastHealthCheck
	isHealthy := e.isHealthy
	e.mu.RUnlock()

	// Only check every 30 seconds
	if time.Since(lastCheck) < 30*time.Second && isHealthy {
		return true
	}

	// Perform health check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := e.Client.NetworkID(ctx)

	e.mu.Lock()
	e.lastHealthCheck = time.Now()
	e.isHealthy = (err == nil)
	e.mu.Unlock()

	if err != nil {
		log.Printf("⚠️ Health check failed for %s: %v", getShortRPCName(e.currentRPC), err)
		return false
	}

	return true
}

// AutoSwitchOnError automatically switches RPC if an error indicates connection issues
func (e *EthClient) AutoSwitchOnError(err error) bool {
	if err == nil {
		return false
	}

	errorStr := strings.ToLower(err.Error())

	// Check for connection-related errors
	connectionErrors := []string{
		"connection refused",
		"connection reset",
		"timeout",
		"dial tcp",
		"i/o timeout",
		"network is unreachable",
		"no such host",
		"connection timed out",
		"context deadline exceeded",
		"eof",
		"broken pipe",
	}

	for _, connErr := range connectionErrors {
		if strings.Contains(errorStr, connErr) {
			log.Printf("🔄 Detected connection error, attempting RPC switch: %v", err)

			switchErr := e.SwitchRPC()
			if switchErr != nil {
				log.Printf("❌ Auto RPC switch failed: %v", switchErr)
				return false
			}

			log.Printf("✅ Auto RPC switch successful")
			return true
		}
	}

	return false
}

// GetCurrentRPCInfo returns information about current RPC
func (e *EthClient) GetCurrentRPCInfo() (string, bool, int, int) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.currentRPC, e.isHealthy, e.rpcIndex + 1, len(e.rpcEndpoints)
}

// GetFailedRPCCount returns the number of currently failed RPCs
func (e *EthClient) GetFailedRPCCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return len(e.failedRPCs)
}

// LogConnectionStatus logs current connection status
func (e *EthClient) LogConnectionStatus() {
	currentRPC, isHealthy, rpcIndex, totalRPCs := e.GetCurrentRPCInfo()
	failedCount := e.GetFailedRPCCount()

	status := "🟢 Healthy"
	if !isHealthy {
		status = "🔴 Unhealthy"
	}

	log.Printf("🌐 RPC Status: %s | Current: %s (%d/%d) | Failed: %d",
		status, getShortRPCName(currentRPC), rpcIndex, totalRPCs, failedCount)
}

// WithRetry executes a function with automatic retry and RPC switching
func (e *EthClient) WithRetry(operation string, fn func() error) error {
	maxRetries := 3
	baseDelay := 2 * time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Check RPC health before operation
		if !e.HealthCheck() {
			log.Printf("⚠️ RPC unhealthy before %s, attempting switch...", operation)
			if err := e.SwitchRPC(); err != nil {
				log.Printf("❌ RPC switch failed: %v", err)
			}
		}

		// Execute the operation
		err := fn()
		if err == nil {
			// Success
			if attempt > 0 {
				log.Printf("✅ %s succeeded after %d retries", operation, attempt)
			}
			return nil
		}

		// Log the error
		log.Printf("❌ %s attempt %d/%d failed: %v", operation, attempt+1, maxRetries, err)

		// Check if this is a connection error that warrants RPC switching
		if e.AutoSwitchOnError(err) {
			log.Printf("🔄 RPC switched due to connection error in %s", operation)
			// Don't count RPC switch attempts against retry limit
			continue
		}

		// If this is the last attempt, return the error
		if attempt == maxRetries-1 {
			return fmt.Errorf("%s failed after %d attempts: %v", operation, maxRetries, err)
		}

		// Wait before retrying (exponential backoff)
		delay := time.Duration(attempt+1) * baseDelay
		log.Printf("⏳ Retrying %s in %v...", operation, delay)
		time.Sleep(delay)
	}

	return fmt.Errorf("%s failed after all retries", operation)
}

// GetTokenBalanceWithRetry gets token balance with automatic retry
func (e *EthClient) GetTokenBalanceWithRetry(tokenAddr, walletAddr common.Address) (*big.Int, error) {
	var balance *big.Int

	err := e.WithRetry("GetTokenBalance", func() error {
		var err error
		balance, err = e.getTokenBalanceOnce(tokenAddr, walletAddr)
		return err
	})

	return balance, err
}

// GetNativeBalanceWithRetry gets native BNB balance with automatic retry
func (e *EthClient) GetNativeBalanceWithRetry(walletAddr common.Address) (*big.Int, error) {
	var balance *big.Int

	err := e.WithRetry("GetNativeBalance", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var err error
		balance, err = e.Client.BalanceAt(ctx, walletAddr, nil)
		return err
	})

	return balance, err
}

// Internal helper functions
func (e *EthClient) getTokenBalanceOnce(tokenAddr, walletAddr common.Address) (*big.Int, error) {
	// ERC20 balanceOf function signature
	balanceOfSig := "70a08231" // balanceOf(address)
	paddedAddress := common.LeftPadBytes(walletAddr.Bytes(), 32)
	callData := common.Hex2Bytes(balanceOfSig + common.Bytes2Hex(paddedAddress))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create proper CallMsg struct
	callMsg := ethereum.CallMsg{
		To:   &tokenAddr,
		Data: callData,
	}

	result, err := e.Client.CallContract(ctx, callMsg, nil)
	if err != nil {
		return nil, err
	}

	balance := new(big.Int).SetBytes(result)
	return balance, nil
}

// Close closes the client connection
func (e *EthClient) Close() {
	if e.Client != nil {
		e.Client.Close()
	}
}

// Helper functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func getShortRPCName(rpcURL string) string {
	// Extract domain from URL for shorter logging
	if strings.Contains(rpcURL, "binance.org") {
		return "Binance"
	} else if strings.Contains(rpcURL, "defibit.io") {
		return "DefiBit"
	} else if strings.Contains(rpcURL, "ninicoin.io") {
		return "NiniCoin"
	} else if strings.Contains(rpcURL, "ankr.com") {
		return "Ankr"
	} else if strings.Contains(rpcURL, "blxrbdn.com") {
		return "BloxRoute"
	} else if strings.Contains(rpcURL, "nodereal.io") {
		return "NodeReal"
	} else if strings.Contains(rpcURL, "core.chainstack.com") {
		return "Chainstack"
	} else if strings.Contains(rpcURL, "getblock.io") {
		return "GetBlock"
	} else if strings.Contains(rpcURL, "infura.io") {
		return "Infura"
	} else if strings.Contains(rpcURL, "bscrpc.com") {
		return "BSC-RPC"
	}
	return "Custom"
}

// IsConnectionError checks if an error is connection-related (exported for use in other packages)
func IsConnectionError(err error) bool {
	if err == nil {
		return false
	}

	errorStr := strings.ToLower(err.Error())
	connectionErrors := []string{
		"connection refused",
		"connection reset",
		"timeout",
		"dial tcp",
		"i/o timeout",
		"network is unreachable",
		"no such host",
		"connection timed out",
		"context deadline exceeded",
		"eof",
		"broken pipe",
	}

	for _, connErr := range connectionErrors {
		if strings.Contains(errorStr, connErr) {
			return true
		}
	}

	return false
}
