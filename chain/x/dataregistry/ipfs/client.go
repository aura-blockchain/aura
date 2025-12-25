// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package ipfs

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"sync"
	"time"

	shell "github.com/ipfs/go-ipfs-api"
)

// IPFSClient provides methods for interacting with IPFS
type IPFSClient interface {
	// Upload uploads data to IPFS and returns the CID
	Upload(ctx context.Context, data []byte) (string, error)

	// Download downloads data from IPFS by CID
	Download(ctx context.Context, cid string) ([]byte, error)

	// Pin pins content to ensure it persists
	Pin(ctx context.Context, cid string) error

	// Unpin unpins content
	Unpin(ctx context.Context, cid string) error

	// CalculateHash calculates SHA256 hash of data
	CalculateHash(data []byte) []byte

	// VerifyHash verifies that downloaded data matches expected hash
	VerifyHash(data []byte, expectedHash []byte) bool

	// IsConnected checks if connection to IPFS node is active
	IsConnected(ctx context.Context) bool

	// GetNodeID returns the IPFS node ID
	GetNodeID(ctx context.Context) (string, error)
}

// Client implements IPFSClient interface
type Client struct {
	shell   *shell.Shell
	config  *Config
	timeout time.Duration
}

// Config holds IPFS client configuration
type Config struct {
	// APIEndpoint is the IPFS API endpoint (default: http://localhost:5001)
	APIEndpoint string

	// Timeout for IPFS operations
	Timeout time.Duration

	// AutoPin automatically pins uploaded content
	AutoPin bool

	// MaxRetries for failed operations
	MaxRetries int

	// RetryDelay between retries
	RetryDelay time.Duration
}

// DefaultConfig returns default IPFS configuration
func DefaultConfig() *Config {
	return &Config{
		APIEndpoint: "http://localhost:5001",
		Timeout:     30 * time.Second,
		AutoPin:     true,
		MaxRetries:  3,
		RetryDelay:  1 * time.Second,
	}
}

// NewClient creates a new IPFS client
func NewClient(config *Config) (*Client, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Validate configuration
	if config.APIEndpoint == "" {
		return nil, fmt.Errorf("IPFS API endpoint cannot be empty")
	}

	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	// Create shell connection to IPFS node
	sh := shell.NewShell(config.APIEndpoint)
	if sh == nil {
		return nil, fmt.Errorf("failed to create IPFS shell")
	}

	// Set timeout
	sh.SetTimeout(config.Timeout)

	client := &Client{
		shell:   sh,
		config:  config,
		timeout: config.Timeout,
	}

	return client, nil
}

// Upload uploads data to IPFS and returns the CID
func (c *Client) Upload(ctx context.Context, data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("cannot upload empty data")
	}

	var cid string
	var err error

	// Retry logic
	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(c.config.RetryDelay):
			}
		}

		// Upload to IPFS
		cid, err = c.shell.Add(bytes.NewReader(data))
		if err == nil {
			break
		}
	}

	if err != nil {
		return "", fmt.Errorf("failed to upload to IPFS after %d attempts: %w", c.config.MaxRetries+1, err)
	}

	// Auto-pin if configured
	if c.config.AutoPin {
		if err := c.Pin(ctx, cid); err != nil {
			return "", fmt.Errorf("uploaded to IPFS (CID: %s) but failed to pin: %w", cid, err)
		}
	}

	return cid, nil
}

// Download downloads data from IPFS by CID
func (c *Client) Download(ctx context.Context, cid string) ([]byte, error) {
	if cid == "" {
		return nil, fmt.Errorf("CID cannot be empty")
	}

	// Validate CID format
	if !IsValidCID(cid) {
		return nil, fmt.Errorf("invalid CID format: %s", cid)
	}

	var data []byte
	var err error

	// Retry logic
	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(c.config.RetryDelay):
			}
		}

		// Download from IPFS
		reader, err := c.shell.Cat(cid)
		if err != nil {
			continue
		}

		data, err = io.ReadAll(reader)
		reader.Close()

		if err == nil {
			break
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to download from IPFS after %d attempts: %w", c.config.MaxRetries+1, err)
	}

	return data, nil
}

// Pin pins content to ensure it persists
func (c *Client) Pin(ctx context.Context, cid string) error {
	if cid == "" {
		return fmt.Errorf("CID cannot be empty")
	}

	// Validate CID format
	if !IsValidCID(cid) {
		return fmt.Errorf("invalid CID format: %s", cid)
	}

	err := c.shell.Pin(cid)
	if err != nil {
		return fmt.Errorf("failed to pin CID %s: %w", cid, err)
	}

	return nil
}

// Unpin unpins content
func (c *Client) Unpin(ctx context.Context, cid string) error {
	if cid == "" {
		return fmt.Errorf("CID cannot be empty")
	}

	// Validate CID format
	if !IsValidCID(cid) {
		return fmt.Errorf("invalid CID format: %s", cid)
	}

	err := c.shell.Unpin(cid)
	if err != nil {
		return fmt.Errorf("failed to unpin CID %s: %w", cid, err)
	}

	return nil
}

// CalculateHash calculates SHA256 hash of data
func (c *Client) CalculateHash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// IsConnected checks if connection to IPFS node is active
func (c *Client) IsConnected(ctx context.Context) bool {
	_, err := c.shell.ID()
	return err == nil
}

// GetNodeID returns the IPFS node ID
func (c *Client) GetNodeID(ctx context.Context) (string, error) {
	info, err := c.shell.ID()
	if err != nil {
		return "", fmt.Errorf("failed to get node ID: %w", err)
	}
	return info.ID, nil
}

// VerifyHash verifies that downloaded data matches expected hash
func (c *Client) VerifyHash(data []byte, expectedHash []byte) bool {
	actualHash := c.CalculateHash(data)
	return bytes.Equal(actualHash, expectedHash)
}

// VerifyHashHex verifies data against hex-encoded hash
func (c *Client) VerifyHashHex(data []byte, expectedHashHex string) bool {
	expectedHash, err := hex.DecodeString(expectedHashHex)
	if err != nil {
		return false
	}
	return c.VerifyHash(data, expectedHash)
}

// GetConfig returns the client configuration
func (c *Client) GetConfig() *Config {
	return c.config
}

// MockClient is a mock implementation for testing
type MockClient struct {
	mu      sync.RWMutex
	storage map[string][]byte // CID -> data mapping
	pinned  map[string]bool   // CID -> pinned status
}

// NewMockClient creates a new mock IPFS client for testing
func NewMockClient() *MockClient {
	return &MockClient{
		storage: make(map[string][]byte),
		pinned:  make(map[string]bool),
	}
}

// Upload implements mock upload
func (m *MockClient) Upload(ctx context.Context, data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("cannot upload empty data")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate mock CID from hash (valid CIDv0 format with base58 characters)
	hash := sha256.Sum256(data)
	hexStr := hex.EncodeToString(hash[:22])
	cid := "Qm" + hexStr // 46 character string (Qm + 44 hex chars)

	m.storage[cid] = data
	m.pinned[cid] = true // Auto-pin by default

	return cid, nil
}

// Download implements mock download
func (m *MockClient) Download(ctx context.Context, cid string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, ok := m.storage[cid]
	if !ok {
		return nil, fmt.Errorf("CID not found: %s", cid)
	}
	return data, nil
}

// Pin implements mock pin
func (m *MockClient) Pin(ctx context.Context, cid string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.storage[cid]; !ok {
		return fmt.Errorf("CID not found: %s", cid)
	}
	m.pinned[cid] = true
	return nil
}

// Unpin implements mock unpin
func (m *MockClient) Unpin(ctx context.Context, cid string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.storage[cid]; !ok {
		return fmt.Errorf("CID not found: %s", cid)
	}
	m.pinned[cid] = false
	return nil
}

// CalculateHash implements mock hash calculation
func (m *MockClient) CalculateHash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// VerifyHash implements mock hash verification
func (m *MockClient) VerifyHash(data []byte, expectedHash []byte) bool {
	actualHash := m.CalculateHash(data)
	if len(actualHash) != len(expectedHash) {
		return false
	}
	for i := range actualHash {
		if actualHash[i] != expectedHash[i] {
			return false
		}
	}
	return true
}

// IsConnected implements mock connection check
func (m *MockClient) IsConnected(ctx context.Context) bool {
	return true
}

// GetNodeID implements mock node ID retrieval
func (m *MockClient) GetNodeID(ctx context.Context) (string, error) {
	return "QmMockNodeID", nil
}

// IsPinned checks if a CID is pinned (for testing)
func (m *MockClient) IsPinned(cid string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.pinned[cid]
}
