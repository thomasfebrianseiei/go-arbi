package utils

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

// FormatBigInt formats a big.Int to a readable string with decimals
func FormatBigInt(value *big.Int, decimals uint8) string {
	if value == nil {
		return "0"
	}
	
	// Convert to big.Float
	floatValue := new(big.Float).SetInt(value)
	
	// Calculate divisor (10^decimals)
	divisor := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil))
	
	// Divide by 10^decimals
	floatValue.Quo(floatValue, divisor)
	
	// Convert to string with proper precision
	return fmt.Sprintf("%."+fmt.Sprintf("%d", decimals)+"f", floatValue)
}

// AddressToChecksum converts an address to checksum format
func AddressToChecksum(address string) string {
	if !strings.HasPrefix(address, "0x") {
		address = "0x" + address
	}
	return common.HexToAddress(address).Hex()
}

// CalculatePercentage calculates percentage difference between two values
func CalculatePercentage(a, b *big.Int) float64 {
	if b.Cmp(big.NewInt(0)) == 0 {
		return 0
	}
	
	// Convert to big.Float for division
	aFloat := new(big.Float).SetInt(a)
	bFloat := new(big.Float).SetInt(b)
	
	// Calculate percentage
	percent := new(big.Float).Quo(aFloat, bFloat)
	percent = percent.Sub(percent, big.NewFloat(1))
	percent = percent.Mul(percent, big.NewFloat(100))
	
	// Convert to float64
	result, _ := percent.Float64()
	return result
}

// StringSliceContains checks if a string slice contains a string
func StringSliceContains(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

// WaitForConfirmation prompts the user to confirm an action
func WaitForConfirmation(prompt string) bool {
	var response string
	fmt.Print(prompt + " (y/n): ")
	fmt.Scanln(&response)
	return strings.ToLower(response) == "y" || strings.ToLower(response) == "yes"
}