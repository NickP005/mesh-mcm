# mcminterface v1.0.5

Work In Progress! Beta for Mochimo 3.0!

This is a Golang library for interfacing with the MCM Network through the native socket/tcp protocol.  
Written by [NickP005](https://github.com/NickP005)  

## Usage

In query basics there is the main function you can test with. Compilation as usual with
```
go build
```
and then run the binary.  

Or alternatively you can run the code with
```
go run .
```

There is a file, `settings.json`, that you can edit to change the startup settings. Below is an example of the file:
```json
{
    "StartIPs": [
        "0.0.0.0"
    ],
    "IPs": [
        "0.0.0.0"
    ],
    "Nodes": [
        {
            "IP": "0.0.0.0",
            "LastSeen": "2024-08-05T12:54:28.9042045+02:00",
            "Ping": 792
        },
    ],
    "IPExpandDepth": 2,
    "ForceQueryStartIPs": false,
    "QuerySize": 5,
    "QueryTimeout": 5,
    "QueryRetries": 3
}
```

## Examples
### Interface startup
```go
package main

import (
    "fmt"
    "github.com/NickP005/go_mcminterface"
)

func main() {
    go_mcminterface.LoadSettings("settings.json") // Load the settings from settings.json
    go_mcminterface.ExpandIPs()
    go_mcminterface.BenchmarkNodes(5) // Benchmark with 5 concurrent pings
    fmt.Println("Settings loaded and nodes benchmarked.")
}
```

### Resolving a tag
```go
func resolveTag() {
    tag := "0f8213c50de73ee326009d6a1475d1dba1105777"

    addr, err := mcminterface.QueryTagResolveHex(tag)
    if err != nil {
        fmt.Println("Error resolving tag:", err)
        return
    }
    // Print the bytes of the address
    fmt.Println("WOTS+ address bound to the tag:", addr.Address[:])

    // Print the amount of nanoMCM in the address
    fmt.Println("Balance:", addr.GetAmount())
}
```

## Suggested Functions
Below there are the functions that are meant to be official: they query multiple nodes and return the most common result that is agreed by more than 50% of the nodes called.  
Functions such as tag resolve haven't been implemented in query_manager.go yet, but are present in queries.go.  

### LoadSettings
Loads the settings from the `settings.json` file. For the code to properly work, ensure to call it before doing any QueryX function.   
```go
func LoadSettings() (SettingsType)
```

### SaveSettings
Saves the settings to the `settings.json` file.  
```go
func SaveSettings(settings Settings)
```

### BenchmarkNodes
Benchmarks all the nodes in the settings file. Useful at startup to determine the best nodes to query in later connections.  
```go
func BenchmarkNodes(n int)
```
`n` specifies how many concurrent pings to send.  

### ExpandIPs
Expands the IPs in the settings file.  
Doing a benchmark after this function is recommended.  
```go
func ExpandIPs() ()
```

### QueryBalance
Queries the balance of the specified full WOTS+ address given as hex.  
```go
func QueryBalance(wots_address string) (uint64, error) 
```

### QueryTagResolveHex
Queries the tag resolve of the specified tag given as hex.  
```go
func QueryTagResolveHex(tag string) (string, error)
```

### QueryBlockFromNumber
Queries the block from the specified block number.  
If the block number is 0, it will return the latest block.  
```go
func QueryBlockFromNumber(block_number uint64) (Block, error)
```

### QueryLatestBlockNumber
Queries the latest block number.  
```go
func QueryLatestBlockNumber() (uint64, error)
```

### GetTrailers
Queries the trailers of the specified block.  
```go
func GetTrailers(start_block uint32, count uint32) ([]BTRAILER, error)
```


## Notes
- The code is still in development and is not yet ready for production use.
- Every query asks for QuerySize nodes that are picked by PickNodes. That function picks randomly the nodes, but nodes that have lower ping time are more likely to be picked!

## Contact
For any questions or suggestions, feel free to contact me on Discord in my [Discord Development Server](https://discord.gg/rasRT6wQwx)  

[![Discord](https://img.shields.io/badge/Discord-7289DA?style=for-the-badge&logo=discord&logoColor=white)](https://discord.gg/rasRT6wQwx)