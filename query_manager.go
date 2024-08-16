package go_mcminterface

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sync"
	"time"
)

var Settings SettingsType

// location of settings
var Settings_file = "settings.json"

var settingsFS embed.FS

// Global settings
type SettingsType struct {
	StartIPs           []string
	IPs                []string
	Nodes              []RemoteNode
	IPExpandDepth      int
	ForceQueryStartIPs bool // Forces to query only start ips bypassing PickNodes
	QuerySize          int  // Number of nodes to query, quorum is 50% + 1
	QueryTimeout       int  // Timeout in seconds
	MaxQueryAttempts   int  // Maximum number of attempts to query a block
}

type RemoteNode struct {
	IP       string
	LastSeen time.Time
	Ping     uint32
}

// Load default settings with embed in settings.json
func LoadDefaultSettings() {
	// Load default settings from embed
	data, err := settingsFS.ReadFile("settings.json")
	if err != nil {
		fmt.Println("LoadDefaultSettings(): Error reading settings.json:", err)
		return
	}

	var settings SettingsType
	err = json.Unmarshal(data, &settings)
	if err != nil {
		fmt.Println("LoadDefaultSettings(): Error decoding settings.json:", err)
		return
	}

	Settings = settings
}

// load settings from path. If the file does not exist, load default ones
func LoadSettings(paths ...string) SettingsType {
	path := ""
	if len(paths) > 0 {
		path = paths[0]
	} else {
		// set to user config dir
		dir, err := os.UserConfigDir()
		if err != nil {
			fmt.Println("Error getting user config dir:", err)
			return Settings
		}
		path = dir + "/mcminterface/settings.json"
	}

	if path == "" {
		path = Settings_file
	} else {
		Settings_file = path
	}

	// Load settings from path
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Error reading settings.json:", err)
		LoadDefaultSettings()
		return Settings
	}

	var settings SettingsType
	err = json.Unmarshal(data, &settings)
	if err != nil {
		fmt.Println("Error decoding settings.json:", err)
		LoadDefaultSettings()
		return Settings
	}

	Settings = settings
	return Settings
}

// save settings Settings_file
func SaveSettings(settings SettingsType) {
	file, err := os.Create(Settings_file)
	if err != nil {
		fmt.Println("Error creating settings.json")
	}
	defer file.Close()
	// format with indentation
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	err = encoder.Encode(settings)

	if err != nil {
		fmt.Println("Error encoding settings.json")
	}
}

// Expand known IPs
func ExpandIPs() {
	// Add start IPs to the settings IPs
	Settings.IPs = append(Settings.IPs, Settings.StartIPs...)
	queriedIPs := make(map[string]bool)

	for i := 0; i < Settings.IPExpandDepth; i++ {
		ips := make([]string, 0)
		ch := make(chan string)

		for _, ip := range Settings.IPs {
			if queriedIPs[ip] {
				continue // Skip already queried IPs
			}
			queriedIPs[ip] = true

			go func(ip string) {
				sd := ConnectToNode(ip)
				if sd.block_num == 0 {
					fmt.Println("Connection failed")
					ch <- ""
					return
				}
				new_ips, err := sd.GetIPList()
				if err != nil {
					fmt.Println("Error:", err)
					ch <- ""
					return
				}
				// Add new IPs to the list if they are not already present
				for _, new_ip := range new_ips {
					found := false
					for _, ip := range ips {
						if new_ip == ip {
							found = true
							break
						}
					}
					if !found {
						ips = append(ips, new_ip)
					}
				}
				ch <- ip
			}(ip)
		}

		timeout := time.After(time.Duration(Settings.QueryTimeout) * time.Second) // Set timeout of 5 seconds
		for range Settings.IPs {
			select {
			case ip := <-ch:
				if ip != "" {
					ips = append(ips, ip)
				}
			case <-timeout:
				fmt.Println("Timeout")
				return
			}
		}

		Settings.IPs = ips
	}
}

// Benchmark all IPs in the time they take to ConnectToNode
func BenchmarkNodes(n int) {
	ch := make(chan RemoteNode)

	for i := 0; i < len(Settings.IPs); i += n {
		end := i + n
		if end > len(Settings.IPs) {
			end = len(Settings.IPs)
		}
		ips := Settings.IPs[i:end]

		for _, ip := range ips {
			go func(ip string) {
				start := time.Now()
				sd := ConnectToNode(ip)
				ping := time.Since(start)
				if sd.block_num == 0 {
					fmt.Println("Connection failed")
					ping = 10 * time.Second
				}
				// ping in milliseconds
				ch <- RemoteNode{IP: ip, Ping: uint32(ping / time.Millisecond)}
			}(ip)
		}
	}

	timeout := time.After(time.Duration(Settings.QueryTimeout) * time.Second)

	for i := 0; i < len(Settings.IPs); i += n {
		end := i + n
		if end > len(Settings.IPs) {
			end = len(Settings.IPs)
		}
		ips := Settings.IPs[i:end]

		for range ips {
			select {
			case node := <-ch:
				found := false
				for i, n := range Settings.Nodes {
					if n.IP == node.IP {
						Settings.Nodes[i].Ping = (n.Ping*2 + node.Ping) / 3
						Settings.Nodes[i].LastSeen = time.Now()
						found = true
						break
					}
				}
				if !found {
					Settings.Nodes = append(Settings.Nodes, node)
				}
			case <-timeout:
				fmt.Println("Timeout")
				return
			}
		}
	}

	close(ch)
}

// Pick n random nodes from Settings.Nodes
// the probability of picking a node is e**(-ping)
func PickNodes(n int) []RemoteNode {
	// if forcequerystartips is set, return the nodes with ip startip
	if Settings.ForceQueryStartIPs {
		nodes := make([]RemoteNode, 0)
		for _, node := range Settings.Nodes {
			if node.IP == Settings.StartIPs[0] {
				nodes = append(nodes, node)
			}
		}
		return nodes
	}

	if n >= len(Settings.Nodes) {
		return Settings.Nodes
	}

	nodes := make([]RemoteNode, 0)
	for i := 0; i < n; i++ {
		// calculate the sum of e**(-ping) for all nodes
		sum := 0.0
		for _, node := range Settings.Nodes {
			sum += math.Exp(-1 / float64(node.Ping/2))
		}
		// pick a random number between 0 and sum
		r := sum * rand.Float64()
		// find the node that corresponds to the random number
		for _, node := range Settings.Nodes {
			r -= math.Exp(-1 / float64(node.Ping/2))
			if r <= 0 {
				// if it is already in the list, decrease i and continue
				found := false
				for _, n := range nodes {
					if n.IP == node.IP {
						i--
						found = true
						break
					}
				}
				if !found {
					nodes = append(nodes, node)
				}
				break
			}
		}
	}
	return nodes
}

// Query the balance of an address given as hex
func QueryBalance(wots_address string) (uint64, error) {
	wots_addr := WotsAddressFromHex(wots_address)

	// connect to a random node
	nodes := PickNodes(Settings.QuerySize)
	balances := make([]WotsAddress, 0)

	// Ask for result on the same time
	ch := make(chan WotsAddress)

	for _, node := range nodes {
		go func(node RemoteNode) {
			sd := ConnectToNode(node.IP)
			if sd.block_num == 0 {
				fmt.Println("Connection failed")
				ch <- WotsAddress{}
				return
			}
			// get the balance of the wots_addr GetBalance
			balance, err := sd.GetBalance(wots_addr)
			if err != nil {
				fmt.Println("Error:", err)
				ch <- WotsAddress{}
				return
			}
			wots_addr.Amount = balance
			ch <- wots_addr
		}(node)
	}

	timeout := time.After(time.Duration(Settings.QueryTimeout) * time.Second)
	for range nodes {
		select {
		case balance := <-ch:
			if balance.Amount != 0 {
				balances = append(balances, balance)
			}
		case <-timeout:
			fmt.Println("Timeout")
			//return 0, fmt.Errorf("timeout")
		}
	}

	close(ch)

	// Calculate the most frequent balance
	counts := make(map[uint64]int)
	for _, balance := range balances {
		counts[balance.Amount]++
	}

	// See if there is a balance that reaches quorum

	max_balance := uint64(0)
	for balance, count := range counts {
		if count >= Settings.QuerySize/2+1 {
			max_balance = balance
			break
		}
	}

	// If no balance reaches quorum, return 0
	if max_balance == 0 {
		return 0, fmt.Errorf("no balance reaches quorum")
	}

	// Save the most frequent balance to the wots_address
	wots_addr.Amount = max_balance
	//fmt.Println("Wots address:", wots_addr)

	return max_balance, nil
}

// QueryBlockHash queries the block hash (HASHLEN) of a block number
// if block number is 0, it returns the hash of the last block.
func QueryBlockHash(block_num uint64) ([HASHLEN]byte, error) {
	// connect to a random node
	nodes := PickNodes(Settings.QuerySize)
	hashes := make([][HASHLEN]byte, 0)

	// Ask for result on the same time
	ch := make(chan [HASHLEN]byte)
	var wg sync.WaitGroup

	for _, node := range nodes {
		wg.Add(1)
		go func(node RemoteNode) {
			defer wg.Done()
			sd := ConnectToNode(node.IP)
			if sd.block_num == 0 {
				fmt.Println("Connection failed")
				ch <- [HASHLEN]byte{}
				return
			}
			// get the block hash
			hash, err := sd.GetBlockHash(block_num)
			if err != nil {
				fmt.Println("Error:", err)
				ch <- [HASHLEN]byte{}
				return
			}
			ch <- hash
		}(node)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	timeout := time.After(time.Duration(Settings.QueryTimeout) * time.Second)

	for range nodes {
		select {
		case hash := <-ch:
			if hash != [HASHLEN]byte{} {
				hashes = append(hashes, hash)
			}
		case <-timeout:
			fmt.Println("Timeout triggered")
			// stop the goroutines
			//return [HASHLEN]byte{}, fmt.Errorf("timeout")
		}
	}

	// Calculate the most frequent hash
	counts := make(map[[HASHLEN]byte]int)
	for _, hash := range hashes {
		counts[hash]++
	}

	// See if there is a hash that reaches quorum
	var max_hash [HASHLEN]byte
	for hash, count := range counts {
		if count >= Settings.QuerySize/2+1 {
			max_hash = hash
			break
		}
	}

	// If no hash reaches quorum, return 0
	if max_hash == [HASHLEN]byte{} {
		return [HASHLEN]byte{}, fmt.Errorf("no hash reaches quorum")
	}

	return max_hash, nil
}

// QueryBlockBytes
// 1. Gets the block hash 2. Gets the block bytes from a random node until the hash matches
func QueryBlockBytes(block_num uint64) ([]byte, error) {
	// get the block hash
	hash, err := QueryBlockHash(block_num)
	if err != nil {
		return nil, err
	}

	found := false
	attempts := 0
	var block []byte
	for !found {
		// connect to one random node
		nodes := PickNodes(1)
		node := nodes[0]
		sd := ConnectToNode(node.IP)
		if sd.block_num == 0 {
			fmt.Println("Connection failed")
			// try again with another node
			continue
		}
		// if block number is 0, get the latest block
		if block_num == 0 {
			block_num = sd.block_num
		}
		// get the block bytes
		block, err = sd.GetBlockBytes(block_num)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}
		// check if the sha256 matches the bytes[:-HASHLEN]
		sha256_hash := sha256.Sum256(block[:len(block)-HASHLEN])
		if sha256_hash == hash {
			found = true
		}
		attempts++
		if attempts > Settings.MaxQueryAttempts {
			return nil, fmt.Errorf("max query attempts reached")
		}
	}
	return block, nil
}

// QueryBlockFromNumber
func QueryBlockFromNumber(block_num uint64) (Block, error) {
	// get the block bytes
	block_bytes, err := QueryBlockBytes(block_num)
	if err != nil {
		return Block{}, err
	}
	// create the block from the bytes
	block := BlockFromBytes(block_bytes)
	return block, nil
}

// QueryTagResolve queries the tag resolve
func QueryTagResolve(tag []byte) (WotsAddress, error) {
	// connect to a random node
	nodes := PickNodes(Settings.QuerySize)
	addresses := make([]WotsAddress, 0)

	// Ask for result on the same time
	ch := make(chan WotsAddress)

	for _, node := range nodes {
		go func(node RemoteNode) {
			sd := ConnectToNode(node.IP)
			if sd.block_num == 0 {
				fmt.Println("Connection failed")
				ch <- WotsAddress{}
				return
			}
			// get the address from the tag
			addr, err := sd.ResolveTag(tag)
			if err != nil {
				fmt.Println("Error:", err)
				ch <- WotsAddress{}
				return
			}
			ch <- addr
		}(node)
	}

	timeout := time.After(time.Duration(Settings.QueryTimeout) * time.Second)

	for range nodes {
		select {
		case addr := <-ch:
			if addr.Amount != 0 {
				addresses = append(addresses, addr)
			}
		case <-timeout:
			fmt.Println("Timeout")
			//return WotsAddress{}, fmt.Errorf("timeout")
		}
	}

	close(ch)

	// Calculate the most frequent address
	counts := make(map[WotsAddress]int)
	for _, addr := range addresses {
		counts[addr]++
	}

	// See if there is an address that reaches quorum

	var max_addr WotsAddress
	for addr, count := range counts {
		if count >= Settings.QuerySize/2+1 {
			max_addr = addr
			break
		}
	}

	// If no address reaches quorum, return 0
	if max_addr.Amount == 0 {
		return WotsAddress{}, fmt.Errorf("no address reaches quorum")
	}

	return max_addr, nil
}

// QueryTagResolveHex
func QueryTagResolveHex(tag_hex string) (WotsAddress, error) {
	tag, err := hex.DecodeString(tag_hex)
	if err != nil {
		return WotsAddress{}, err
	}
	return QueryTagResolve(tag)
}

// QueryLatestBlockNumber
func QueryLatestBlockNumber() (uint64, error) {
	// connect to a random node
	nodes := PickNodes(Settings.QuerySize)
	block_numbers := make([]uint64, 0)

	// Ask for result on the same time
	ch := make(chan uint64)

	for _, node := range nodes {
		go func(node RemoteNode) {
			sd := ConnectToNode(node.IP)
			if sd.block_num == 0 {
				fmt.Println("Connection failed")
				ch <- 0
				return
			}
			// get the latest block number
			ch <- sd.block_num
		}(node)
	}

	timeout := time.After(time.Duration(Settings.QueryTimeout) * time.Second)

	for range nodes {
		select {
		case block_num := <-ch:
			if block_num != 0 {
				block_numbers = append(block_numbers, block_num)
			}
		case <-timeout:
			fmt.Println("Timeout")
			//return 0, fmt.Errorf("timeout")
		}
	}

	close(ch)

	// Calculate the most frequent block number
	counts := make(map[uint64]int)
	for _, block_num := range block_numbers {
		counts[block_num]++
	}

	// See if there is a block number that reaches quorum

	var max_block_num uint64
	for block_num, count := range counts {
		if count >= Settings.QuerySize/2+1 {
			max_block_num = block_num
			break
		}
	}

	// If no block number reaches quorum, return 0
	if max_block_num == 0 {
		return 0, fmt.Errorf("no block number reaches quorum")
	}

	return max_block_num, nil
}

// QueryBTrailers using GetTrailersBytes
func queryBTrailers(start_block uint32, count uint32) ([]BTRAILER, error) {
	nodes := PickNodes(Settings.QuerySize)
	trailers_bytes := make([][]byte, 0)

	// Ask for result on the same time
	ch := make(chan []byte)
	var wg sync.WaitGroup

	for _, node := range nodes {
		wg.Add(1)
		go func(node RemoteNode) {
			defer wg.Done()
			sd := ConnectToNode(node.IP)
			if sd.block_num == 0 {
				fmt.Println("Connection failed")
				ch <- nil
				return
			}
			// get the trailers bytes
			tf_bytes, err := sd.GetTrailersBytes(start_block, count)
			if err != nil {
				fmt.Println("Error:", err)
				ch <- nil
				return
			}
			ch <- tf_bytes
		}(node)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	timeout := time.After(time.Duration(Settings.QueryTimeout) * time.Second)

	for range nodes {
		select {
		case tf_bytes := <-ch:
			if tf_bytes != nil {
				trailers_bytes = append(trailers_bytes, tf_bytes)
			}
		case <-timeout:
			fmt.Println("Timeout")
			//return nil, fmt.Errorf("timeout")
		}
	}

	// Calculate the most frequent trailers
	counts := make(map[string]int)
	for _, tf_bytes := range trailers_bytes {
		counts[string(tf_bytes)]++
	}

	// See if there is a trailers that reaches quorum
	var max_tf_bytes []byte
	for tf_bytes, count := range counts {
		if count >= Settings.QuerySize/2+1 {
			max_tf_bytes = []byte(tf_bytes)
			break
		}
	}

	// If no trailers reaches quorum, return 0
	if max_tf_bytes == nil {
		return nil, fmt.Errorf("no trailers reaches quorum")
	}

	// Convert the bytes to BTRAILER
	trailers := make([]BTRAILER, 0)
	for i := 0; i < len(max_tf_bytes); i += 160 {
		trailer := bTrailerFromBytes(max_tf_bytes[i : i+160])
		trailers = append(trailers, trailer)
	}

	return trailers, nil
}

// QueryBTrailers queries the block trailers starting from `start_block` and fetches
// `count` trailers. It splits the request into chunks of 1000 trailers and processes them concurrently.
func QueryBTrailers(start_block uint32, count uint32) ([]BTRAILER, error) {
	var mu sync.Mutex
	var wg sync.WaitGroup
	errCh := make(chan error, 1)
	results := make(map[uint32][]BTRAILER)

	fullChunks := count / 1000
	remainder := count % 1000

	queryChunk := func(start uint32, count uint32, index uint32) {
		defer wg.Done()
		chunkTrailers, err := queryBTrailers(start, count)
		if err != nil {
			select {
			case errCh <- err:
			default:
			}
			return
		}
		mu.Lock()
		results[index] = chunkTrailers
		mu.Unlock()
	}

	// Launch goroutines for each full chunk
	for i := uint32(0); i < fullChunks; i++ {
		wg.Add(1)
		go queryChunk(start_block+i*1000, 1000, i)
	}

	// Launch a goroutine for the remainder
	if remainder > 0 {
		wg.Add(1)
		go queryChunk(start_block+fullChunks*1000, remainder, fullChunks)
	}

	// Wait for all goroutines to finish
	go func() {
		wg.Wait()
		close(errCh)
	}()

	// Check for errors
	for err := range errCh {
		if err != nil {
			return nil, fmt.Errorf("error querying trailers: %w", err)
		}
	}

	// Collect results in order
	var trailers []BTRAILER
	for i := uint32(0); i <= fullChunks; i++ {
		if chunk, exists := results[i]; exists {
			trailers = append(trailers, chunk...)
		}
	}

	return trailers, nil
}
