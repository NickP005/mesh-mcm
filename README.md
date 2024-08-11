# mcminterface v0.1.0

Work In Progress!

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
    "QuerySize": 5
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

### ExpandIPs
Expands the IPs in the settings file.  
```go
func ExpandIPs() ()
```

### BenchmarkNodes
Benchmarks all the nodes in the settings file. Useful at startup to determine the best nodes to query in later connections.  
```go
func BenchmarkNodes(n int)
```
`n` specifies how many concurrent pings to send.  

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


## Notes
- The code is still in development and is not yet ready for production use.
- Every query asks for QuerySize nodes that are picked by PickNodes. That function picks randomly the nodes, but nodes that have lower ping time are more likely to be picked!