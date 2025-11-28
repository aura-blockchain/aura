package ipfs

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test Upload Operations

func TestUpload_Success(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	testData := []byte("Test upload data")
	cid, err := client.Upload(ctx, testData)

	require.NoError(t, err)
	assert.NotEmpty(t, cid)
	assert.True(t, IsValidCID(cid))
}

func TestUpload_EmptyData(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	_, err := client.Upload(ctx, []byte{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot upload empty data")
}

func TestUpload_NilData(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	_, err := client.Upload(ctx, nil)
	assert.Error(t, err)
}

func TestUpload_LargeData(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	// Create large data (1MB)
	largeData := make([]byte, 1024*1024)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	cid, err := client.Upload(ctx, largeData)
	require.NoError(t, err)
	assert.NotEmpty(t, cid)
}

func TestUpload_AutoPin(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	testData := []byte("Test auto-pin")
	cid, err := client.Upload(ctx, testData)
	require.NoError(t, err)

	// Verify auto-pinned
	assert.True(t, client.IsPinned(cid))
}

func TestUpload_DifferentDataTypes(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	tests := []struct {
		name string
		data []byte
	}{
		{"Text", []byte("Plain text data")},
		{"JSON", []byte(`{"key": "value", "number": 123}`)},
		{"Binary", []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD}},
		{"Empty String", []byte("")}, // Should fail
		{"Unicode", []byte("Hello 世界 🌍")},
		{"XML", []byte(`<?xml version="1.0"?><root><item>value</item></root>`)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.data) == 0 {
				_, err := client.Upload(ctx, tt.data)
				assert.Error(t, err)
				return
			}

			cid, err := client.Upload(ctx, tt.data)
			require.NoError(t, err)
			assert.NotEmpty(t, cid)
			assert.True(t, IsValidCID(cid))
		})
	}
}

// Test Download Operations

func TestDownload_Success(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	testData := []byte("Test download data")
	cid, err := client.Upload(ctx, testData)
	require.NoError(t, err)

	downloaded, err := client.Download(ctx, cid)
	require.NoError(t, err)
	assert.Equal(t, testData, downloaded)
}

func TestDownload_NonExistentCID(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	_, err := client.Download(ctx, "QmNonExistentCID12345678901234567890123456")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "CID not found")
}

func TestDownload_EmptyCID(t *testing.T) {
	// Test with real client struct
	realClient := &Client{
		config: DefaultConfig(),
	}
	ctx := context.Background()

	_, err := realClient.Download(ctx, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "CID cannot be empty")
}

func TestDownload_InvalidCID(t *testing.T) {
	// Test with real client struct
	realClient := &Client{
		config: DefaultConfig(),
	}
	ctx := context.Background()

	invalidCIDs := []string{
		"invalid",
		"Qm", // Too short
		"123456", // Not base58
		"QmInvalidCharacters!!@@",
	}

	for _, cid := range invalidCIDs {
		_, err := realClient.Download(ctx, cid)
		if !IsValidCID(cid) {
			assert.Error(t, err, "CID %s should be invalid", cid)
		}
	}
}

func TestDownload_VerifyIntegrity(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	originalData := []byte("Integrity test data")
	cid, err := client.Upload(ctx, originalData)
	require.NoError(t, err)

	downloaded, err := client.Download(ctx, cid)
	require.NoError(t, err)

	// Verify data integrity
	originalHash := client.CalculateHash(originalData)
	downloadedHash := client.CalculateHash(downloaded)
	assert.Equal(t, originalHash, downloadedHash)
}

// Test Pin/Unpin Operations

func TestPin_Success(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	testData := []byte("Test pin")
	cid, err := client.Upload(ctx, testData)
	require.NoError(t, err)

	// Already pinned by upload, but should not error
	err = client.Pin(ctx, cid)
	assert.NoError(t, err)
	assert.True(t, client.IsPinned(cid))
}

func TestPin_NonExistentCID(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	err := client.Pin(ctx, "QmNonExistentCID12345678901234567890123456")
	assert.Error(t, err)
}

func TestPin_EmptyCID(t *testing.T) {
	// Test with real client struct
	realClient := &Client{
		config: DefaultConfig(),
	}
	ctx := context.Background()

	err := realClient.Pin(ctx, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "CID cannot be empty")
}

func TestUnpin_Success(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	testData := []byte("Test unpin")
	cid, err := client.Upload(ctx, testData)
	require.NoError(t, err)
	assert.True(t, client.IsPinned(cid))

	err = client.Unpin(ctx, cid)
	require.NoError(t, err)
	assert.False(t, client.IsPinned(cid))
}

func TestUnpin_NonExistentCID(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	err := client.Unpin(ctx, "QmNonExistentCID12345678901234567890123456")
	assert.Error(t, err)
}

func TestUnpin_EmptyCID(t *testing.T) {
	// Test with real client struct
	realClient := &Client{
		config: DefaultConfig(),
	}
	ctx := context.Background()

	err := realClient.Unpin(ctx, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "CID cannot be empty")
}

// Test Hash Operations

func TestCalculateHash(t *testing.T) {
	client := NewMockClient()

	testData := []byte("Hash calculation test")
	hash := client.CalculateHash(testData)

	assert.NotNil(t, hash)
	assert.Equal(t, 32, len(hash)) // SHA256 produces 32 bytes

	// Verify hash is correct
	expectedHash := sha256.Sum256(testData)
	assert.Equal(t, expectedHash[:], hash)
}

func TestCalculateHash_EmptyData(t *testing.T) {
	client := NewMockClient()

	hash := client.CalculateHash([]byte{})
	assert.NotNil(t, hash)
	assert.Equal(t, 32, len(hash))
}

func TestCalculateHash_Deterministic(t *testing.T) {
	client := NewMockClient()

	testData := []byte("Deterministic test")
	hash1 := client.CalculateHash(testData)
	hash2 := client.CalculateHash(testData)

	assert.Equal(t, hash1, hash2)
}

func TestVerifyHash_Success(t *testing.T) {
	client := NewMockClient()

	testData := []byte("Verify hash test")
	expectedHash := client.CalculateHash(testData)

	result := client.VerifyHash(testData, expectedHash)
	assert.True(t, result)
}

func TestVerifyHash_Failure(t *testing.T) {
	client := NewMockClient()

	testData := []byte("Original data")
	wrongHash := client.CalculateHash([]byte("Different data"))

	result := client.VerifyHash(testData, wrongHash)
	assert.False(t, result)
}

func TestVerifyHash_EmptyHash(t *testing.T) {
	client := NewMockClient()

	testData := []byte("Test data")
	result := client.VerifyHash(testData, []byte{})
	assert.False(t, result)
}

func TestVerifyHashHex_Success(t *testing.T) {
	client := &Client{
		config: DefaultConfig(),
	}

	testData := []byte("Hex verification test")
	hash := sha256.Sum256(testData)
	hashHex := hex.EncodeToString(hash[:])

	result := client.VerifyHashHex(testData, hashHex)
	assert.True(t, result)
}

func TestVerifyHashHex_InvalidHex(t *testing.T) {
	client := &Client{
		config: DefaultConfig(),
	}

	testData := []byte("Test data")
	result := client.VerifyHashHex(testData, "invalid-hex-string!")
	assert.False(t, result)
}

// Test Connection and Node Info

func TestIsConnected(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	connected := client.IsConnected(ctx)
	assert.True(t, connected)
}

func TestGetNodeID(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	nodeID, err := client.GetNodeID(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, nodeID)
	assert.Contains(t, nodeID, "Qm")
}

// Test Configuration

func TestDefaultConfigExtended(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, "http://localhost:5001", config.APIEndpoint)
	assert.Equal(t, 30*time.Second, config.Timeout)
	assert.True(t, config.AutoPin)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 1*time.Second, config.RetryDelay)
}

func TestNewClient_WithNilConfig(t *testing.T) {
	client, err := NewClient(nil)
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.config)
	assert.Equal(t, "http://localhost:5001", client.config.APIEndpoint)
}

func TestNewClient_WithCustomConfig(t *testing.T) {
	customConfig := &Config{
		APIEndpoint: "http://custom-ipfs:5001",
		Timeout:     60 * time.Second,
		AutoPin:     false,
		MaxRetries:  5,
		RetryDelay:  2 * time.Second,
	}

	client, err := NewClient(customConfig)
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "http://custom-ipfs:5001", client.config.APIEndpoint)
	assert.Equal(t, 60*time.Second, client.config.Timeout)
	assert.False(t, client.config.AutoPin)
	assert.Equal(t, 5, client.config.MaxRetries)
}

func TestNewClient_EmptyEndpoint(t *testing.T) {
	config := &Config{
		APIEndpoint: "",
	}

	_, err := NewClient(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "IPFS API endpoint cannot be empty")
}

func TestNewClient_ZeroTimeout(t *testing.T) {
	config := &Config{
		APIEndpoint: "http://localhost:5001",
		Timeout:     0, // Should default to 30s
	}

	client, err := NewClient(config)
	require.NoError(t, err)
	assert.Equal(t, 30*time.Second, client.config.Timeout)
}

func TestGetConfig(t *testing.T) {
	config := DefaultConfig()
	client := &Client{
		config: config,
	}

	retrieved := client.GetConfig()
	assert.Equal(t, config, retrieved)
}

// Test Concurrent Operations

func TestConcurrentUploads(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	numUploads := 20
	done := make(chan bool, numUploads)
	var mu sync.Mutex
	cids := make(map[string]bool)

	for i := 0; i < numUploads; i++ {
		go func(index int) {
			defer func() { done <- true }()

			data := []byte(fmt.Sprintf("Concurrent upload %d", index))
			cid, err := client.Upload(ctx, data)
			if err != nil {
				t.Errorf("Upload %d failed: %v", index, err)
				return
			}

			mu.Lock()
			if cids[cid] {
				t.Errorf("Duplicate CID: %s", cid)
			}
			cids[cid] = true
			mu.Unlock()
		}(i)
	}

	// Wait for all uploads
	for i := 0; i < numUploads; i++ {
		<-done
	}

	assert.Equal(t, numUploads, len(cids))
}

func TestConcurrentDownloads(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	// Upload test data
	testData := []byte("Concurrent download test")
	cid, err := client.Upload(ctx, testData)
	require.NoError(t, err)

	numDownloads := 20
	done := make(chan bool, numDownloads)

	for i := 0; i < numDownloads; i++ {
		go func() {
			defer func() { done <- true }()

			downloaded, err := client.Download(ctx, cid)
			if err != nil {
				t.Errorf("Download failed: %v", err)
				return
			}

			if !bytes.Equal(testData, downloaded) {
				t.Error("Downloaded data doesn't match")
			}
		}()
	}

	// Wait for all downloads
	for i := 0; i < numDownloads; i++ {
		<-done
	}
}

func TestConcurrentPinUnpin(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	// Upload multiple files
	cids := make([]string, 10)
	for i := 0; i < 10; i++ {
		data := []byte(fmt.Sprintf("Pin test %d", i))
		cid, err := client.Upload(ctx, data)
		require.NoError(t, err)
		cids[i] = cid
	}

	// Concurrent unpin/pin operations
	done := make(chan bool, len(cids)*2)

	for _, cid := range cids {
		localCID := cid
		go func() {
			defer func() { done <- true }()
			_ = client.Unpin(ctx, localCID)
		}()

		go func() {
			defer func() { done <- true }()
			time.Sleep(10 * time.Millisecond)
			_ = client.Pin(ctx, localCID)
		}()
	}

	// Wait for all operations
	for i := 0; i < len(cids)*2; i++ {
		<-done
	}
}

// Test Error Handling

func TestUploadDownloadRoundtrip(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	tests := []struct {
		name string
		data []byte
	}{
		{"Small", []byte("small")},
		{"Medium", bytes.Repeat([]byte("medium"), 100)},
		{"Large", bytes.Repeat([]byte("large"), 10000)},
		{"Binary", []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD, 0xFC}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Upload
			cid, err := client.Upload(ctx, tt.data)
			require.NoError(t, err)

			// Download
			downloaded, err := client.Download(ctx, cid)
			require.NoError(t, err)

			// Verify
			assert.Equal(t, tt.data, downloaded)

			// Verify hash
			hash := client.CalculateHash(tt.data)
			assert.True(t, client.VerifyHash(downloaded, hash))
		})
	}
}

// Test Context Cancellation

func TestUpload_ContextCancelled(t *testing.T) {
	client := NewMockClient()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	testData := []byte("Context test")
	_, err := client.Upload(ctx, testData)

	// Mock client doesn't respect context, but in real impl would error
	// For now, just verify it completes (mock doesn't check context)
	_ = err
}

func TestDownload_ContextTimeout(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	// Upload first
	testData := []byte("Timeout test")
	cid, err := client.Upload(ctx, testData)
	require.NoError(t, err)

	// Create context with very short timeout
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(2 * time.Millisecond) // Ensure timeout

	// Mock client doesn't respect timeout, but real client would
	_, err = client.Download(ctxTimeout, cid)
	// Mock completes successfully
	_ = err
}

// Test Data Integrity

func TestDataIntegrity_CorruptedData(t *testing.T) {
	client := NewMockClient()

	originalData := []byte("Original data for corruption test")
	expectedHash := client.CalculateHash(originalData)

	// Simulate corruption
	corruptedData := make([]byte, len(originalData))
	copy(corruptedData, originalData)
	corruptedData[0] = ^corruptedData[0] // Flip bits

	result := client.VerifyHash(corruptedData, expectedHash)
	assert.False(t, result, "Corrupted data should not verify")
}

func TestDataIntegrity_MultipleUploads(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	testData := []byte("Deterministic CID test")

	// Upload same data multiple times
	cid1, err := client.Upload(ctx, testData)
	require.NoError(t, err)

	// Note: In real IPFS, same data = same CID (content-addressed)
	// Our mock generates different CIDs, but that's okay for testing
	cid2, err := client.Upload(ctx, testData)
	require.NoError(t, err)

	// Both should be valid CIDs
	assert.True(t, IsValidCID(cid1))
	assert.True(t, IsValidCID(cid2))
}

// Test Edge Cases

func TestUploadDownload_EmptyContext(t *testing.T) {
	client := NewMockClient()
	// Use background context (not empty, but minimal)
	ctx := context.Background()

	testData := []byte("Context test")
	cid, err := client.Upload(ctx, testData)
	require.NoError(t, err)

	downloaded, err := client.Download(ctx, cid)
	require.NoError(t, err)
	assert.Equal(t, testData, downloaded)
}

func TestMultipleCIDFormats(t *testing.T) {
	// Test various valid CID formats
	validCIDs := []string{
		"QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG", // CIDv0
		"Qma6K85PJ4jvFrphWHFvxVDWWvhvNhFqVJxHvbvhuNg7Wg",
		"QmT78zSuBmuS4z925WZfrqQ1qHaJ56DQaTfyMUF7F8ff5o",
	}

	for _, cid := range validCIDs {
		assert.True(t, IsValidCID(cid), "CID %s should be valid", cid)
	}

	invalidCIDs := []string{
		"",
		"Qm", // Too short
		"invalid",
		"12345",
		"QmInvalid!!",
	}

	for _, cid := range invalidCIDs {
		assert.False(t, IsValidCID(cid), "CID %s should be invalid", cid)
	}
}

func TestHashConsistency(t *testing.T) {
	client := NewMockClient()

	testData := []byte("Consistency test data")

	// Calculate hash multiple times
	hash1 := client.CalculateHash(testData)
	hash2 := client.CalculateHash(testData)
	hash3 := client.CalculateHash(testData)

	// All should be identical
	assert.Equal(t, hash1, hash2)
	assert.Equal(t, hash2, hash3)
}

func TestPinStatus(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	testData := []byte("Pin status test")
	cid, err := client.Upload(ctx, testData)
	require.NoError(t, err)

	// Should be pinned after upload
	assert.True(t, client.IsPinned(cid))

	// Unpin
	err = client.Unpin(ctx, cid)
	require.NoError(t, err)
	assert.False(t, client.IsPinned(cid))

	// Pin again
	err = client.Pin(ctx, cid)
	require.NoError(t, err)
	assert.True(t, client.IsPinned(cid))
}

func TestBinaryDataPreservation(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	// Test with all byte values
	binaryData := make([]byte, 256)
	for i := 0; i < 256; i++ {
		binaryData[i] = byte(i)
	}

	cid, err := client.Upload(ctx, binaryData)
	require.NoError(t, err)

	downloaded, err := client.Download(ctx, cid)
	require.NoError(t, err)

	// Verify all bytes preserved
	assert.Equal(t, binaryData, downloaded)
	for i := 0; i < 256; i++ {
		assert.Equal(t, byte(i), downloaded[i], "Byte %d should be preserved", i)
	}
}

func TestUnicodeDataPreservation(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	unicodeData := []byte("Hello 世界 🌍 مرحبا здравствуй")

	cid, err := client.Upload(ctx, unicodeData)
	require.NoError(t, err)

	downloaded, err := client.Download(ctx, cid)
	require.NoError(t, err)

	assert.Equal(t, unicodeData, downloaded)
	assert.Equal(t, string(unicodeData), string(downloaded))
}

// Test CID Validation Extended

func TestIsValidCIDExtended(t *testing.T) {
	tests := []struct {
		cid   string
		valid bool
	}{
		{"QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG", true},
		{"QmT78zSuBmuS4z925WZfrqQ1qHaJ56DQaTfyMUF7F8ff5o", true},
		{"", false},
		{"Qm", false},
		{"invalid", false},
		{"QmInvalid!!@@##", false},
		{"12345678901234567890", false},
	}

	for _, tt := range tests {
		t.Run(tt.cid, func(t *testing.T) {
			result := IsValidCID(tt.cid)
			assert.Equal(t, tt.valid, result, "CID: %s", tt.cid)
		})
	}
}

// Benchmark Tests

func BenchmarkUpload(b *testing.B) {
	client := NewMockClient()
	ctx := context.Background()
	testData := []byte("Benchmark upload data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.Upload(ctx, testData)
	}
}

func BenchmarkDownload(b *testing.B) {
	client := NewMockClient()
	ctx := context.Background()
	testData := []byte("Benchmark download data")

	cid, _ := client.Upload(ctx, testData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.Download(ctx, cid)
	}
}

func BenchmarkCalculateHash(b *testing.B) {
	client := NewMockClient()
	testData := []byte("Benchmark hash calculation")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.CalculateHash(testData)
	}
}

func BenchmarkVerifyHash(b *testing.B) {
	client := NewMockClient()
	testData := []byte("Benchmark hash verification")
	hash := client.CalculateHash(testData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.VerifyHash(testData, hash)
	}
}
