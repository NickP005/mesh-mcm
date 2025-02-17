# mcminterface v1.0.16

A production-ready Golang library for interfacing with the Mochimo Network through the native socket/tcp protocol.
Written by [NickP005](https://github.com/NickP005)

## Installation

```bash
go get github.com/NickP005/go_mcminterface
```

## Core Functions

### Query Manager Functions

These are the primary functions you should use as they implement consensus mechanisms by querying multiple nodes:

#### LoadSettings
```go
func LoadSettings(path string) (SettingsType)
```
Loads configuration from a settings file. Must be called before any other operations.
- `path`: Path to settings.json file
- Returns: Current settings configuration

#### QueryBalance
```go
func QueryBalance(wots_address string) (uint64, error)
```
Queries the balance of a WOTS+ address across multiple nodes to reach consensus.
- `wots_address`: Hexadecimal WOTS+ address
- Returns: Balance in nanoMCM and error if any

#### QueryLatestBlockNumber
```go
func QueryLatestBlockNumber() (uint64, error)
```
Gets the latest block number with consensus from multiple nodes.
- Returns: Block number and error if any

#### QueryBlockFromNumber
```go
func QueryBlockFromNumber(block_number uint64) (Block, error)
```
Retrieves a specific block with consensus verification.
- `block_number`: Block number to query (0 for latest)
- Returns: Block structure and error if any

#### GetTrailers
```go
func GetTrailers(start_block uint32, count uint32) ([]BTRAILER, error)
```
Fetches block trailers with consensus verification.
- `start_block`: Starting block number
- `count`: Number of trailers to retrieve
- Returns: Array of block trailers and error if any

### Network Management Functions

#### BenchmarkNodes
```go
func BenchmarkNodes(n int)
```
Benchmarks nodes in parallel to determine optimal query targets.
- `n`: Number of concurrent benchmark operations

#### ExpandIPs
```go
func ExpandIPs()
```
Discovers new nodes by querying known nodes for their peers.
Run BenchmarkNodes after this to evaluate new nodes.

#### SaveSettings
```go
func SaveSettings(settings Settings)
```
Persists current network configuration to settings file.
- `settings`: Settings structure to save

## Configuration

Example `settings.json`:
```json
{
    "StartIPs": ["seed1.mochimo.org", "seed2.mochimo.org"],
    "IPs": [],
    "Nodes": [],
    "IPExpandDepth": 2,
    "ForceQueryStartIPs": false,
    "QuerySize": 5,
    "QueryTimeout": 5,
    "QueryRetries": 3
}
```

### Configuration Fields
- `StartIPs`: Initial seed nodes
- `IPs`: Known network nodes
- `Nodes`: Benchmarked nodes with performance data
- `IPExpandDepth`: How many levels deep to search for new nodes
- `ForceQueryStartIPs`: Always include seed nodes in queries
- `QuerySize`: Number of nodes to query for consensus
- `QueryTimeout`: Single query timeout in seconds
- `QueryRetries`: Number of retries per query

## Complete Example

```go
package main

import (
    "fmt"
    "github.com/NickP005/go_mcminterface"
)

func main() {
    // Initialize network
    go_mcminterface.LoadSettings("settings.json")
    go_mcminterface.ExpandIPs()
    go_mcminterface.BenchmarkNodes(5)

    // Query latest block
    blockNum, err := go_mcminterface.QueryLatestBlockNumber()
    if err != nil {
        panic(err)
    }
    fmt.Printf("Latest block: %d\n", blockNum)

    // Query balance
    address := "YOUR_WOTS_ADDRESS_HERE"
    balance, err := go_mcminterface.QueryBalance(address)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Balance: %d nanoMCM\n", balance)

    // Get block data
    block, err := go_mcminterface.QueryBlockFromNumber(blockNum)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Block hash: %x\n", block.Bhash)
}
```

## Implementation Details

- All query functions use a consensus mechanism requiring >50% agreement among nodes
- Node selection is weighted by ping performance
- Automatic retry and failover mechanisms are built-in
- Concurrent operations are handled safely

# Support & Community

Join our communities for support and discussions:

<div align="center">

[![NickP005 Development Server](https://img.shields.io/discord/709417966881472572?color=7289da&label=NickP005%20Development%20Server&logo=discord&logoColor=white)](https://discord.gg/Q5jM8HJhNT)   
[![Mochimo Official](https://img.shields.io/discord/460867662977695765?color=7289da&label=Mochimo%20Official&logo=discord&logoColor=white)](https://discord.gg/SvdXdr2j3Y)

</div>

- **NickP005 Development Server**: Technical support and development discussions
- **Mochimo Official**: General Mochimo blockchain discussions and community