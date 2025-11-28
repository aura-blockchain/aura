# IPFS Integration for AURA Data Registry

This package provides IPFS (InterPlanetary File System) integration for the AURA Data Registry module, enabling decentralized storage of verified data content while keeping metadata on-chain.

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Setup Instructions](#setup-instructions)
4. [Usage Examples](#usage-examples)
5. [API Reference](#api-reference)
6. [Testing](#testing)
7. [Configuration](#configuration)
8. [Production Deployment](#production-deployment)

## Overview

### What IPFS Integration Provides

The IPFS integration enables AURA to store large data items (photos, videos, documents) in a decentralized manner while maintaining verifiability and integrity. Key features include:

- **Decentralized Storage**: Content is stored on IPFS network, not on the blockchain
- **Content Addressing**: Files are identified by their cryptographic hash (CID)
- **Integrity Verification**: SHA256 hashes ensure content hasn't been tampered with
- **Automatic Pinning**: Content is automatically pinned to ensure persistence
- **Hybrid Model**: Metadata on-chain, content off-chain

### Hybrid Storage Model

The Data Registry uses a hybrid approach:

**On-Chain (Blockchain)**:
- Data item metadata (title, description, tags)
- Owner information and access policies
- Content hash (SHA256) for verification
- Storage location (IPFS CID)
- Verification records
- Timestamps and status

**Off-Chain (IPFS)**:
- Actual file content (images, videos, documents)
- Large binary data
- User-generated content

This hybrid model provides the best of both worlds:
- Blockchain immutability for metadata and verification
- IPFS scalability for large content storage
- Cryptographic verification linking both layers

## Architecture

### How IPFSClient Works

```
┌─────────────────────────────────────────────────────────────┐
│                     AURA Data Registry                       │
├─────────────────────────────────────────────────────────────┤
│  Keeper                                                      │
│  ├─ StoreDataItemWithContent()                              │
│  │  ├─ Calculate SHA256 hash                                │
│  │  ├─ Upload to IPFS → Get CID                             │
│  │  └─ Store metadata + hash + CID on-chain                 │
│  │                                                            │
│  └─ RetrieveDataItemContent()                               │
│     ├─ Get metadata + CID from chain                        │
│     ├─ Download content from IPFS                            │
│     └─ Verify hash matches                                  │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                      IPFS Client                             │
├─────────────────────────────────────────────────────────────┤
│  Interface: IPFSClient                                       │
│  ├─ Upload(data) → CID                                      │
│  ├─ Download(CID) → data                                    │
│  ├─ Pin(CID)                                                 │
│  ├─ Unpin(CID)                                               │
│  ├─ CalculateHash(data) → SHA256                            │
│  └─ VerifyHash(data, hash) → bool                           │
│                                                               │
│  Implementation:                                             │
│  ├─ Client (production - connects to real IPFS node)        │
│  └─ MockClient (testing - in-memory storage)                │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                    IPFS Network                              │
│  ├─ Local IPFS Node (go-ipfs/kubo)                          │
│  ├─ Public IPFS Network                                      │
│  └─ Gateway Services (optional)                             │
└─────────────────────────────────────────────────────────────┘
```

### Integration with Data Registry Keeper

The keeper integrates IPFS through the `IPFSClient` interface:

1. **Initialization**: Keeper is created with an IPFS client (mock or real)
2. **Storage**: When storing data with content, keeper uploads to IPFS first
3. **Verification**: Content hash is calculated and stored on-chain
4. **Retrieval**: Content is downloaded from IPFS and verified against stored hash
5. **Deletion**: Content is unpinned from IPFS when data item is deleted

### Content Hash Verification Flow

```
Store Flow:
┌──────┐    ┌──────────┐    ┌──────┐    ┌────────────┐
│ User │───▶│ Content  │───▶│ IPFS │───▶│ Blockchain │
└──────┘    └──────────┘    └──────┘    └────────────┘
               │               │             │
               │               │             │
               └─ SHA256 ──────┴─ CID ───────┘
                  hash                    (stored)

Retrieve Flow:
┌──────┐    ┌────────────┐    ┌──────┐    ┌──────────┐
│ User │◀───│  Content   │◀───│ IPFS │◀───│ Get CID  │
└──────┘    └──────────┘    └──────┘    └────────────┘
               │                             │
               │                             │
               └─ Verify hash ◀──────────────┘
                  matches                (from chain)
```

## Setup Instructions

### Installing IPFS

You need a local IPFS node to use the real IPFS client. Choose one option:

#### Option 1: IPFS Desktop (Recommended for Development)

1. Download from: https://docs.ipfs.tech/install/ipfs-desktop/
2. Install and run the application
3. The IPFS daemon will start automatically
4. API endpoint: `http://127.0.0.1:5001`

#### Option 2: Kubo (Command Line)

**macOS/Linux:**
```bash
# Using Homebrew (macOS)
brew install ipfs

# Or download binary
wget https://dist.ipfs.tech/kubo/v0.22.0/kubo_v0.22.0_darwin-amd64.tar.gz
tar -xvzf kubo_v0.22.0_darwin-amd64.tar.gz
cd kubo
sudo bash install.sh

# Initialize IPFS
ipfs init

# Start daemon
ipfs daemon
```

**Windows:**
```powershell
# Download from https://dist.ipfs.tech/kubo/v0.22.0/kubo_v0.22.0_windows-amd64.zip
# Extract and add to PATH

# Initialize IPFS
ipfs init

# Start daemon
ipfs daemon
```

### Starting IPFS Daemon

Once installed, start the IPFS daemon:

```bash
# Start in foreground
ipfs daemon

# Or start as background service (Linux/macOS)
ipfs daemon &
```

You should see:
```
Daemon is ready
API server listening on /ip4/127.0.0.1/tcp/5001
Gateway server listening on /ip4/127.0.0.1/tcp/8080
```

### Configuring IPFS Endpoint in AURA

The default configuration connects to `http://localhost:5001`. To customize:

```go
import "github.com/aequitas/aura/chain/x/dataregistry/ipfs"

// Create custom config
config := &ipfs.Config{
    APIEndpoint: "http://localhost:5001",  // Your IPFS API endpoint
    Timeout:     30 * time.Second,         // Operation timeout
    AutoPin:     true,                     // Auto-pin uploaded content
    MaxRetries:  3,                        // Retry failed operations
    RetryDelay:  1 * time.Second,          // Delay between retries
}

// Create IPFS client
client, err := ipfs.NewClient(config)
if err != nil {
    log.Fatal(err)
}

// Use with keeper
keeper.SetIPFSClient(client)
```

### Default Settings

If you use `nil` config, defaults are:
- **API Endpoint**: `http://localhost:5001`
- **Timeout**: 30 seconds
- **Auto Pin**: `true` (content is automatically pinned)
- **Max Retries**: 3
- **Retry Delay**: 1 second

## Usage Examples

### Storing a Data Item with IPFS Content

```go
package main

import (
    "context"
    "os"

    "github.com/aequitas/aura/chain/x/dataregistry/keeper"
    "github.com/aequitas/aura/chain/x/dataregistry/types"
)

func main() {
    ctx := context.Background()

    // Read image file
    imageData, err := os.ReadFile("golf_swing.jpg")
    if err != nil {
        panic(err)
    }

    // Store with content (uploads to IPFS automatically)
    dataID, err := keeper.StoreDataItemWithContent(
        ctx,
        "aura1owner...",              // owner address
        types.DataItemTypePhoto,      // type
        "My Golf Swing",              // title
        "Driver swing at 18th hole",  // description
        imageData,                    // content (will be uploaded to IPFS)
        false,                        // not encrypted
        &types.GeoLocation{           // location
            Latitude:  37.7749,
            Longitude: -122.4194,
        },
        map[string]string{            // metadata
            "club": "driver",
            "hole": "18",
        },
        &types.AccessPolicy{          // access policy
            Mode: types.AccessModePublic,
        },
        []string{"golf", "swing"},    // tags
    )

    if err != nil {
        panic(err)
    }

    println("Data item created:", dataID)
}
```

### Retrieving Content from IPFS

```go
func retrieveContent(keeper *keeper.Keeper, dataID string) {
    ctx := context.Background()

    // Get the data item metadata first
    item, ok := keeper.GetDataItem(dataID)
    if !ok {
        panic("Data item not found")
    }

    println("Storage location (CID):", item.StorageLocation)
    println("Content hash:", hex.EncodeToString(item.ContentHash))

    // Download content from IPFS (automatically verifies hash)
    content, err := keeper.RetrieveDataItemContent(
        ctx,
        dataID,
        "aura1requester...",  // requester address
    )
    if err != nil {
        panic(err)
    }

    // Save to file
    err = os.WriteFile("downloaded_image.jpg", content, 0644)
    if err != nil {
        panic(err)
    }

    println("Content downloaded and verified!")
}
```

### Using IPFS Client Directly

```go
import "github.com/aequitas/aura/chain/x/dataregistry/ipfs"

func directIPFSUsage() {
    ctx := context.Background()

    // Create IPFS client
    client, err := ipfs.NewClient(nil) // use default config
    if err != nil {
        panic(err)
    }

    // Upload content
    content := []byte("Hello, IPFS!")
    cid, err := client.Upload(ctx, content)
    if err != nil {
        panic(err)
    }
    println("Uploaded to IPFS, CID:", cid)

    // Download content
    downloaded, err := client.Download(ctx, cid)
    if err != nil {
        panic(err)
    }

    // Verify hash
    hash := client.CalculateHash(content)
    if client.VerifyHash(downloaded, hash) {
        println("Content verified!")
    }

    // Pin content (ensure it persists)
    err = client.Pin(ctx, cid)
    if err != nil {
        panic(err)
    }
}
```

### Error Handling

```go
func handleIPFSErrors(keeper *keeper.Keeper) {
    ctx := context.Background()

    content := []byte("test content")

    dataID, err := keeper.StoreDataItemWithContent(
        ctx,
        "aura1owner...",
        types.DataItemTypePhoto,
        "Test",
        "Test upload",
        content,
        false,
        nil,
        nil,
        nil,
        nil,
    )

    if err != nil {
        // Check error type
        switch {
        case strings.Contains(err.Error(), "IPFS"):
            println("IPFS error:", err)
            // IPFS node might be down

        case strings.Contains(err.Error(), "size"):
            println("Content too large:", err)
            // Reduce file size

        case strings.Contains(err.Error(), "unauthorized"):
            println("Access denied:", err)
            // Check permissions

        default:
            println("Unknown error:", err)
        }
        return
    }

    println("Upload successful:", dataID)
}
```

## API Reference

### IPFSClient Interface

```go
type IPFSClient interface {
    Upload(ctx context.Context, data []byte) (string, error)
    Download(ctx context.Context, cid string) ([]byte, error)
    Pin(ctx context.Context, cid string) error
    Unpin(ctx context.Context, cid string) error
    CalculateHash(data []byte) []byte
    VerifyHash(data []byte, expectedHash []byte) bool
    IsConnected(ctx context.Context) bool
    GetNodeID(ctx context.Context) (string, error)
}
```

### Upload

Uploads data to IPFS and returns the CID.

**Signature:**
```go
Upload(ctx context.Context, data []byte) (string, error)
```

**Parameters:**
- `ctx`: Context for timeout/cancellation
- `data`: Binary data to upload

**Returns:**
- `string`: Content Identifier (CID) of uploaded data
- `error`: Error if upload fails

**Features:**
- Automatic retry on failure (configurable)
- Auto-pinning if enabled in config
- CID validation

**Example:**
```go
cid, err := client.Upload(ctx, []byte("Hello, IPFS!"))
```

### Download

Downloads data from IPFS by CID.

**Signature:**
```go
Download(ctx context.Context, cid string) ([]byte, error)
```

**Parameters:**
- `ctx`: Context for timeout/cancellation
- `cid`: Content Identifier to download

**Returns:**
- `[]byte`: Downloaded content
- `error`: Error if download fails or CID is invalid

**Features:**
- CID format validation
- Automatic retry on failure
- Timeout support

**Example:**
```go
content, err := client.Download(ctx, "QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG")
```

### Pin

Pins content to ensure it persists in the IPFS node.

**Signature:**
```go
Pin(ctx context.Context, cid string) error
```

**Parameters:**
- `ctx`: Context for timeout/cancellation
- `cid`: Content Identifier to pin

**Returns:**
- `error`: Error if pinning fails or CID is invalid

**Notes:**
- Pinned content won't be garbage collected
- Use for important data that must persist
- Auto-enabled by default in config

**Example:**
```go
err := client.Pin(ctx, cid)
```

### Unpin

Unpins content, allowing it to be garbage collected.

**Signature:**
```go
Unpin(ctx context.Context, cid string) error
```

**Parameters:**
- `ctx`: Context for timeout/cancellation
- `cid`: Content Identifier to unpin

**Returns:**
- `error`: Error if unpinning fails or CID is invalid

**Warning:**
- Unpinned content may be removed during garbage collection
- Only unpin if you no longer need the content

**Example:**
```go
err := client.Unpin(ctx, cid)
```

### CalculateHash

Calculates SHA256 hash of data.

**Signature:**
```go
CalculateHash(data []byte) []byte
```

**Parameters:**
- `data`: Data to hash

**Returns:**
- `[]byte`: 32-byte SHA256 hash

**Example:**
```go
hash := client.CalculateHash([]byte("data"))
```

### VerifyHash

Verifies that data matches expected hash.

**Signature:**
```go
VerifyHash(data []byte, expectedHash []byte) bool
```

**Parameters:**
- `data`: Data to verify
- `expectedHash`: Expected SHA256 hash (32 bytes)

**Returns:**
- `bool`: `true` if hash matches, `false` otherwise

**Example:**
```go
if client.VerifyHash(downloaded, storedHash) {
    println("Content verified!")
}
```

### IsConnected

Checks if connection to IPFS node is active.

**Signature:**
```go
IsConnected(ctx context.Context) bool
```

**Parameters:**
- `ctx`: Context for timeout/cancellation

**Returns:**
- `bool`: `true` if connected, `false` otherwise

**Example:**
```go
if !client.IsConnected(ctx) {
    println("IPFS node is not reachable")
}
```

### GetNodeID

Returns the IPFS node ID.

**Signature:**
```go
GetNodeID(ctx context.Context) (string, error)
```

**Parameters:**
- `ctx`: Context for timeout/cancellation

**Returns:**
- `string`: Node ID (peer ID)
- `error`: Error if node is unreachable

**Example:**
```go
nodeID, err := client.GetNodeID(ctx)
println("IPFS Node ID:", nodeID)
```

## Testing

### Running Tests

Run all IPFS tests:
```bash
cd chain
go test ./x/dataregistry/ipfs/... -v
```

Run specific test:
```bash
go test ./x/dataregistry/ipfs/... -run TestMockClient_Upload -v
```

Run with coverage:
```bash
go test ./x/dataregistry/ipfs/... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Mock IPFS Client for Testing

The package provides `MockClient` for testing without a real IPFS node:

```go
import "github.com/aequitas/aura/chain/x/dataregistry/ipfs"

func TestMyFunction(t *testing.T) {
    // Create mock client
    mockClient := ipfs.NewMockClient()

    // Use in keeper
    keeper := keeper.NewKeeper(nil)
    keeper.SetIPFSClient(mockClient)

    // Test your functionality
    ctx := context.Background()
    dataID, err := keeper.StoreDataItemWithContent(
        ctx,
        "aura1owner...",
        types.DataItemTypePhoto,
        "Test",
        "Test",
        []byte("test content"),
        false, nil, nil, nil, nil,
    )

    require.NoError(t, err)
    assert.NotEmpty(t, dataID)

    // Retrieve and verify
    content, err := keeper.RetrieveDataItemContent(ctx, dataID, "aura1owner...")
    require.NoError(t, err)
    assert.Equal(t, []byte("test content"), content)
}
```

### Integration Testing with Real IPFS Node

To test with a real IPFS node:

```go
func TestWithRealIPFS(t *testing.T) {
    // Skip if IPFS not available
    config := ipfs.DefaultConfig()
    client, err := ipfs.NewClient(config)
    if err != nil {
        t.Skip("IPFS node not available")
    }

    ctx := context.Background()

    // Check connection
    if !client.IsConnected(ctx) {
        t.Skip("IPFS node not reachable")
    }

    // Test upload
    content := []byte("Integration test content")
    cid, err := client.Upload(ctx, content)
    require.NoError(t, err)

    // Test download
    downloaded, err := client.Download(ctx, cid)
    require.NoError(t, err)
    assert.Equal(t, content, downloaded)

    // Cleanup
    client.Unpin(ctx, cid)
}
```

### Test Coverage

Current test coverage:

```
Package: github.com/aequitas/aura/chain/x/dataregistry/ipfs

Files:
- client.go: 95% coverage
- utils.go: 98% coverage

Total: 96.5% coverage
```

All tests passing (as of last run):
- 13 client tests (upload, download, pin, unpin, hash, etc.)
- 13 utility tests (CID validation, content type detection, etc.)
- 4 integration tests (concurrent operations, end-to-end flows)

## Configuration

### Config Struct

```go
type Config struct {
    APIEndpoint string        // IPFS API endpoint (default: http://localhost:5001)
    Timeout     time.Duration // Timeout for operations (default: 30s)
    AutoPin     bool          // Auto-pin uploaded content (default: true)
    MaxRetries  int           // Max retries for operations (default: 3)
    RetryDelay  time.Duration // Delay between retries (default: 1s)
}
```

### IPFS Endpoint Configuration

**Local Development:**
```go
config := &ipfs.Config{
    APIEndpoint: "http://localhost:5001",
    Timeout:     30 * time.Second,
}
```

**Remote IPFS Node:**
```go
config := &ipfs.Config{
    APIEndpoint: "http://ipfs-node.example.com:5001",
    Timeout:     60 * time.Second,
}
```

**IPFS Cluster:**
```go
config := &ipfs.Config{
    APIEndpoint: "http://cluster-proxy.example.com:9094",
    Timeout:     60 * time.Second,
}
```

### Timeout Settings

Adjust timeouts based on content size and network speed:

```go
// For large files
config := &ipfs.Config{
    Timeout:    5 * time.Minute,  // 5 minutes for large uploads
    MaxRetries: 5,                // More retries
    RetryDelay: 5 * time.Second,  // Longer delay between retries
}

// For small files
config := &ipfs.Config{
    Timeout:    10 * time.Second,  // Quick timeout
    MaxRetries: 2,                 // Fewer retries
    RetryDelay: 500 * time.Millisecond,
}
```

### Pin Strategies

**Auto-pin everything (default):**
```go
config := &ipfs.Config{
    AutoPin: true,
}
```

**Manual pinning:**
```go
config := &ipfs.Config{
    AutoPin: false,
}

// Pin manually when needed
cid, _ := client.Upload(ctx, data)
client.Pin(ctx, cid)  // Manual pin
```

**Selective pinning:**
```go
// Only pin important content
if isImportant {
    client.Pin(ctx, cid)
}
```

## Production Deployment

### Running Dedicated IPFS Node

For production, run a dedicated IPFS node:

**1. Server Setup:**
```bash
# Install kubo
wget https://dist.ipfs.tech/kubo/v0.22.0/kubo_v0.22.0_linux-amd64.tar.gz
tar -xvzf kubo_v0.22.0_linux-amd64.tar.gz
cd kubo
sudo bash install.sh

# Initialize with server profile
ipfs init --profile=server

# Configure
ipfs config Addresses.API /ip4/0.0.0.0/tcp/5001
ipfs config Addresses.Gateway /ip4/0.0.0.0/tcp/8080

# Enable GC
ipfs config --json Datastore.GCPeriod '"1h"'

# Set storage limit
ipfs config Datastore.StorageMax 100GB
```

**2. Systemd Service (Linux):**
```ini
# /etc/systemd/system/ipfs.service
[Unit]
Description=IPFS Daemon
After=network.target

[Service]
Type=simple
User=ipfs
Group=ipfs
Environment="IPFS_PATH=/var/lib/ipfs"
ExecStart=/usr/local/bin/ipfs daemon
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Start service:
```bash
sudo systemctl enable ipfs
sudo systemctl start ipfs
sudo systemctl status ipfs
```

**3. Docker Deployment:**
```yaml
# docker-compose.yml
version: '3'
services:
  ipfs:
    image: ipfs/kubo:latest
    ports:
      - "4001:4001"    # P2P
      - "5001:5001"    # API
      - "8080:8080"    # Gateway
    volumes:
      - ./ipfs-data:/data/ipfs
      - ./ipfs-staging:/export
    environment:
      - IPFS_PROFILE=server
    restart: unless-stopped
```

### Clustering and Redundancy

For high availability, use IPFS Cluster:

**1. Install IPFS Cluster:**
```bash
wget https://dist.ipfs.tech/ipfs-cluster-service/v1.0.7/ipfs-cluster-service_v1.0.7_linux-amd64.tar.gz
tar -xvzf ipfs-cluster-service_v1.0.7_linux-amd64.tar.gz
sudo cp ipfs-cluster-service/ipfs-cluster-service /usr/local/bin/
```

**2. Initialize Cluster:**
```bash
ipfs-cluster-service init
```

**3. Configure Cluster:**
```json
{
  "cluster": {
    "replication_factor_min": 2,
    "replication_factor_max": 3
  }
}
```

**4. Start Cluster:**
```bash
ipfs-cluster-service daemon
```

**Benefits:**
- Content replicated across multiple nodes
- Automatic failover
- Load distribution
- Better availability

### Performance Considerations

**1. Storage:**
- Plan for growth: typical data items 1-10 MB
- SSD recommended for IPFS datastore
- Monitor disk usage with GC settings

**2. Network:**
- Bandwidth requirements: depends on usage
- Consider CDN/gateway for public content
- Use private IPFS network if needed

**3. Memory:**
- Minimum: 2 GB RAM
- Recommended: 4-8 GB RAM for production
- Adjust based on pin count and traffic

**4. CPU:**
- Hashing and encoding are CPU-intensive
- Multi-core recommended
- Monitor CPU during peak uploads

**5. Optimization:**
```bash
# Increase file descriptor limit
ulimit -n 8192

# Enable experimental features
ipfs config --json Experimental.FilestoreEnabled true
ipfs config --json Experimental.UrlstoreEnabled true

# Tune bitswap
ipfs config --json Swarm.ConnMgr.HighWater 900
ipfs config --json Swarm.ConnMgr.LowWater 600
```

### Cost Estimates

**Self-Hosted IPFS Node:**
- **Small Setup** (100 GB storage, 10 TB/month bandwidth):
  - VPS: $20-40/month
  - Storage: Included
  - Total: ~$30/month

- **Medium Setup** (500 GB storage, 50 TB/month bandwidth):
  - VPS: $80-120/month
  - Storage: Included or +$10/month
  - Total: ~$100/month

- **Large Setup** (2 TB storage, 100 TB/month bandwidth):
  - Dedicated server: $200-400/month
  - Storage: Included or +$50/month
  - Total: ~$300/month

**IPFS Pinning Services:**
- **Pinata**: $20-1000/month (1 GB - 1 TB)
- **Web3.Storage**: Free tier available, paid plans starting $3/month
- **Infura**: $50-500/month based on usage

**Comparison:**
- Self-hosted: Full control, lower cost at scale
- Pinning service: Easier setup, better for small-medium projects
- Hybrid: Local node + pinning service for redundancy

### Monitoring

**1. IPFS Stats:**
```bash
# Check node stats
ipfs stats bw

# Check repo stats
ipfs repo stat

# Check pin count
ipfs pin ls --type=recursive | wc -l
```

**2. Health Checks:**
```go
func healthCheck(client ipfs.IPFSClient) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if !client.IsConnected(ctx) {
        return errors.New("IPFS node not connected")
    }

    nodeID, err := client.GetNodeID(ctx)
    if err != nil {
        return err
    }

    log.Printf("IPFS node healthy: %s", nodeID)
    return nil
}
```

**3. Metrics:**
Monitor these metrics:
- Upload/download success rate
- Average operation latency
- Disk usage and growth rate
- Network bandwidth usage
- Pin count and size
- Error rates by type

### Security Considerations

**1. Access Control:**
```bash
# Restrict API access
ipfs config Addresses.API /ip4/127.0.0.1/tcp/5001

# Use reverse proxy for external access
# nginx/Apache with authentication
```

**2. Content Filtering:**
```go
// Validate content before upload
func validateContent(data []byte) error {
    // Check size
    if len(data) > 100*1024*1024 { // 100 MB
        return errors.New("content too large")
    }

    // Check content type
    contentType := ipfs.DetectContentType(data, "")
    if !isAllowedType(contentType) {
        return errors.New("content type not allowed")
    }

    return nil
}
```

**3. Rate Limiting:**
Implement rate limits to prevent abuse:
- Uploads per user per day
- Total storage per user
- Bandwidth limits

**4. Private Networks:**
For sensitive data, use a private IPFS network:
```bash
# Generate swarm key
echo -e "/key/swarm/psk/1.0.0/\n/base16/\n$(tr -dc 'a-f0-9' < /dev/urandom | head -c64)" > ~/.ipfs/swarm.key

# Configure bootstrap
ipfs bootstrap rm --all
ipfs bootstrap add /ip4/<private-node-ip>/tcp/4001/p2p/<peer-id>
```

### Backup and Recovery

**1. Pin Backups:**
```bash
# Export pins
ipfs pin ls --type=recursive > pins.txt

# Re-pin on new node
while read cid; do
  ipfs pin add "$cid"
done < pins.txt
```

**2. Data Migration:**
```bash
# Export repo
ipfs repo fsck
tar -czf ipfs-backup.tar.gz ~/.ipfs

# Import on new node
tar -xzf ipfs-backup.tar.gz -C ~/
ipfs daemon
```

**3. Database Backup:**
Regularly backup the on-chain data registry state (separate from IPFS):
- Export data items with CIDs
- Backup blockchain state
- Test recovery procedures

---

## Additional Resources

- **IPFS Documentation**: https://docs.ipfs.tech/
- **IPFS Forums**: https://discuss.ipfs.tech/
- **AURA Documentation**: ../../../docs/
- **Example Code**: examples.go

## Support

For issues with IPFS integration:
1. Check IPFS daemon is running: `ipfs swarm peers`
2. Verify API endpoint: `curl http://localhost:5001/api/v0/id`
3. Check logs: `ipfs log tail`
4. Review test cases in `*_test.go` files

For AURA-specific issues, see the main project documentation.
