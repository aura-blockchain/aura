# IPFS Integration Quick Start Guide

Get up and running with IPFS integration in 5 minutes.

## Option 1: Testing Without IPFS (Fastest)

Perfect for development and testing without installing IPFS.

```go
package main

import (
    "context"
    "fmt"
    "github.com/aequitas/aura/chain/x/dataregistry/ipfs"
)

func main() {
    // Create mock client (no IPFS needed!)
    client := ipfs.NewMockClient()
    ctx := context.Background()

    // Upload content
    content := []byte("Hello, IPFS!")
    cid, err := client.Upload(ctx, content)
    if err != nil {
        panic(err)
    }
    fmt.Println("Uploaded! CID:", cid)

    // Download content
    downloaded, err := client.Download(ctx, cid)
    if err != nil {
        panic(err)
    }
    fmt.Println("Downloaded:", string(downloaded))

    // Verify hash
    hash := client.CalculateHash(content)
    if client.VerifyHash(downloaded, hash) {
        fmt.Println("Content verified!")
    }
}
```

## Option 2: Install IPFS Desktop (Easiest for Real IPFS)

1. **Download IPFS Desktop**
   - Visit: https://docs.ipfs.tech/install/ipfs-desktop/
   - Download for your OS (Windows/Mac/Linux)
   - Install and run

2. **IPFS will start automatically**
   - API endpoint: `http://127.0.0.1:5001`
   - Gateway: `http://127.0.0.1:8080`
   - You'll see the IPFS icon in your system tray

3. **Use with AURA**

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/aequitas/aura/chain/x/dataregistry/ipfs"
)

func main() {
    // Create real IPFS client
    config := &ipfs.Config{
        APIEndpoint: "http://localhost:5001",
        Timeout:     30 * time.Second,
        AutoPin:     true,
    }

    client, err := ipfs.NewClient(config)
    if err != nil {
        fmt.Println("Error:", err)
        fmt.Println("Make sure IPFS Desktop is running!")
        return
    }

    ctx := context.Background()

    // Check connection
    if !client.IsConnected(ctx) {
        fmt.Println("IPFS not reachable. Is IPFS Desktop running?")
        return
    }

    nodeID, _ := client.GetNodeID(ctx)
    fmt.Println("Connected to IPFS node:", nodeID)

    // Upload content
    content := []byte("Hello from AURA on real IPFS!")
    cid, err := client.Upload(ctx, content)
    if err != nil {
        panic(err)
    }

    fmt.Println("Uploaded to IPFS!")
    fmt.Println("CID:", cid)
    fmt.Println("View in browser:", "http://127.0.0.1:8080/ipfs/"+cid)
    fmt.Println("View on public gateway:", "https://ipfs.io/ipfs/"+cid)

    // Download and verify
    downloaded, err := client.Download(ctx, cid)
    if err != nil {
        panic(err)
    }

    hash := client.CalculateHash(content)
    if client.VerifyHash(downloaded, hash) {
        fmt.Println("Content verified! ✓")
    }
}
```

## Option 3: Install Kubo CLI (For Advanced Users)

### macOS
```bash
brew install ipfs
ipfs init
ipfs daemon
```

### Linux
```bash
wget https://dist.ipfs.tech/kubo/v0.22.0/kubo_v0.22.0_linux-amd64.tar.gz
tar -xvzf kubo_v0.22.0_linux-amd64.tar.gz
cd kubo
sudo bash install.sh
ipfs init
ipfs daemon
```

### Windows
```powershell
# Download from https://dist.ipfs.tech/kubo/v0.22.0/kubo_v0.22.0_windows-amd64.zip
# Extract and add to PATH
ipfs init
ipfs daemon
```

## Using with Data Registry Keeper

```go
package main

import (
    "context"
    "github.com/aequitas/aura/chain/x/dataregistry/keeper"
    "github.com/aequitas/aura/chain/x/dataregistry/types"
)

func main() {
    // Create keeper with mock IPFS (for testing)
    k := keeper.NewKeeper(nil)

    // Or create with real IPFS
    // ipfsConfig := ipfs.DefaultConfig()
    // k, err := keeper.NewKeeperWithIPFS(nil, ipfsConfig)

    ctx := context.Background()

    // Store data with IPFS content
    photoData := []byte("Golf swing photo data")
    dataID, err := k.StoreDataItemWithContent(
        ctx,
        "aura1owner",
        types.DataItemTypePhoto,
        "My Golf Swing",
        "Driver on 18th hole",
        photoData,
        false, // not encrypted
        nil,   // no geo location
        nil,   // no metadata
        &types.AccessPolicy{
            Mode: types.AccessModePublic,
        },
        []string{"golf", "swing"},
    )

    if err != nil {
        panic(err)
    }

    println("Data item created:", dataID)

    // Retrieve content
    content, err := k.RetrieveDataItemContent(ctx, dataID, "aura1owner")
    if err != nil {
        panic(err)
    }

    println("Content retrieved and verified! Size:", len(content))
}
```

## Running Tests

```bash
cd chain
go test ./x/dataregistry/ipfs/... -v
```

## Running Examples

```bash
cd chain/x/dataregistry/ipfs
go run examples.go
```

Or run individual examples:

```go
package main

import "github.com/aequitas/aura/chain/x/dataregistry/ipfs"

func main() {
    // Run all examples
    ipfs.RunAllExamples()

    // Or run specific examples
    // ipfs.Example4_DirectIPFSUsage()
    // ipfs.Example6_ContentTypeDetection()
    // ipfs.Example9_MockClientForTesting()
}
```

## Troubleshooting

### "Failed to create IPFS client"
- Make sure IPFS is running: `ipfs swarm peers` (should show connected peers)
- Check API is accessible: `curl http://localhost:5001/api/v0/id`
- Verify endpoint in config: should be `http://localhost:5001`

### "IPFS node is not reachable"
- Check if IPFS Desktop is running (system tray icon)
- Or check if daemon is running: `ps aux | grep ipfs`
- Start daemon: `ipfs daemon`

### "Cannot bind to 5001"
- Another IPFS instance is running
- Or port is in use by another service
- Change port in IPFS config: `ipfs config Addresses.API /ip4/127.0.0.1/tcp/5002`

### Tests failing
- If using mock client: should always work
- If using real IPFS: make sure daemon is running
- Check test output for specific errors

## Next Steps

1. **Read the full README**: `README.md` for comprehensive documentation
2. **Review examples**: `examples.go` for complete usage examples
3. **Check architecture**: `README.md` has detailed architecture diagrams
4. **Production deployment**: See README section on production deployment

## Quick Reference

### Create Mock Client
```go
client := ipfs.NewMockClient()
```

### Create Real Client
```go
config := ipfs.DefaultConfig()
client, err := ipfs.NewClient(config)
```

### Upload Content
```go
cid, err := client.Upload(ctx, []byte("content"))
```

### Download Content
```go
content, err := client.Download(ctx, cid)
```

### Verify Hash
```go
hash := client.CalculateHash(originalContent)
verified := client.VerifyHash(downloadedContent, hash)
```

### Check Connection
```go
connected := client.IsConnected(ctx)
```

## API Endpoints (Default)

- **IPFS API**: http://localhost:5001
- **IPFS Gateway**: http://localhost:8080
- **Public Gateway**: https://ipfs.io

## Useful Commands

```bash
# Check IPFS version
ipfs version

# View node ID
ipfs id

# List pinned content
ipfs pin ls

# Check repo stats
ipfs repo stat

# Check connected peers
ipfs swarm peers

# Add file to IPFS
ipfs add <file>

# Get file from IPFS
ipfs cat <cid>

# View in browser
open http://localhost:8080/ipfs/<cid>
```

## Resources

- Full Documentation: `README.md`
- Code Examples: `examples.go`
- Test Files: `*_test.go`
- IPFS Documentation: https://docs.ipfs.tech/
- IPFS Desktop: https://docs.ipfs.tech/install/ipfs-desktop/

---

Ready to use IPFS with AURA! Start with the mock client for testing, then move to real IPFS when ready.
