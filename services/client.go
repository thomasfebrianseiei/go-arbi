package services

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"arbitrage-bot/config"
)

// EthClient manages the connection to the Ethereum network
type EthClient struct {
	Client       *ethclient.Client
	PrivateKey   *ecdsa.PrivateKey
	Address      common.Address
	RpcURLs      []string
	CurrentRpcIdx int
}

// NewEthClient creates a new Ethereum client
func NewEthClient(cfg *config.Config) (*EthClient, error) {
	// Parse private key
	privateKey, err := crypto.HexToECDSA(cfg.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %v", err)
	}
	
	// Get public address
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("error casting public key to ECDSA")
	}
	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	
	client := &EthClient{
		PrivateKey:    privateKey,
		Address:       address,
		RpcURLs:       cfg.RpcURLs,
		CurrentRpcIdx: 0,
	}
	
	// Connect to first RPC endpoint
	err = client.Connect(0)
	if err != nil {
		return nil, err
	}
	
	log.Printf("Connected to BSC network, address: %s\n", address.Hex())
	
	return client, nil
}

// Connect connects to a specific RPC endpoint
func (c *EthClient) Connect(index int) error {
	if index >= len(c.RpcURLs) {
		return fmt.Errorf("RPC index out of bounds")
	}
	
	log.Printf("Connecting to RPC endpoint: %s\n", c.RpcURLs[index])
	client, err := ethclient.Dial(c.RpcURLs[index])
	if err != nil {
		return err
	}
	
	// Close the old client connection if it exists
	if c.Client != nil {
		c.Client.Close()
	}
	
	c.Client = client
	c.CurrentRpcIdx = index
	
	return nil
}

// SwitchRPC switches to the next available RPC endpoint
func (c *EthClient) SwitchRPC() error {
	nextIdx := (c.CurrentRpcIdx + 1) % len(c.RpcURLs)
	err := c.Connect(nextIdx)
	if err != nil {
		return err
	}
	
	log.Printf("Switched to RPC endpoint %d: %s\n", nextIdx, c.RpcURLs[nextIdx])
	return nil
}

// GetCurrentBlock gets the current block number
func (c *EthClient) GetCurrentBlock() (uint64, error) {
	return c.Client.BlockNumber(context.Background())
}

// Close closes the client connection
func (c *EthClient) Close() {
	if c.Client != nil {
		c.Client.Close()
	}
}