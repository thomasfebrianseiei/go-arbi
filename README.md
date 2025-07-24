# BSC Triangular Arbitrage Bot

This bot scans for triangular arbitrage opportunities between PancakeSwap and BiSwap on the Binance Smart Chain (BSC). It looks for routes with at least 0.5% profit (after accounting for gas and fees) and can execute trades automatically using flash loans.

## Features

- Supports multiple trading pairs
- Automatic switching between RPC endpoints
- Bidirectional arbitrage (PancakeSwap -> BiSwap -> PancakeSwap and BiSwap -> PancakeSwap -> BiSwap)
- Flash loan integration for gas-efficient arbitrage
- Minimum profit threshold to ensure profitability after gas costs
- Slippage protection
- Graceful error handling and shutdown

## Prerequisites

- Go 1.15 or later
- A BSC wallet with some BNB for gas
- (Optional) A deployed Flash Arbitrage contract for flash loan-based arbitrage

## Installation

1. Clone this repository:
   git clone https://github.com/yourusername/bsc-arbitrage-bot.git
   cd bsc-arbitrage-bot

2. Install dependencies:

3. Create a `.env` file in the project root with the following content:
