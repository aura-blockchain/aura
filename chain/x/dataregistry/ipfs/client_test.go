package ipfs

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockClient_Upload(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	testData := []byte("Hello, IPFS!")

	// Test successful upload
	cid, err := client.Upload(ctx, testData)
	require.NoError(t, err)
	assert.NotEmpty(t, cid)
	assert.True(t, IsValidCID(cid), "CID should be valid")

	// Test empty data
	_, err = client.Upload(ctx, []byte{})
	assert.Error(t, err)
}

func TestMockClient_Download(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	testData := []byte("Hello, IPFS!")

	// Upload first
	cid, err := client.Upload(ctx, testData)
	require.NoError(t, err)

	// Test successful download
	downloaded, err := client.Download(ctx, cid)
	require.NoError(t, err)
	assert.Equal(t, testData, downloaded)

	// Test download non-existent CID
	_, err = client.Download(ctx, "QmInvalidCID")
	assert.Error(t, err)
}

func TestMockClient_Pin(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	testData := []byte("Test pinning")

	// Upload and pin
	cid, err := client.Upload(ctx, testData)
	require.NoError(t, err)

	// Verify pinned
	assert.True(t, client.IsPinned(cid))

	// Pin again (should not error)
	err = client.Pin(ctx, cid)
	assert.NoError(t, err)

	// Pin non-existent CID
	err = client.Pin(ctx, "QmInvalidCID")
	assert.Error(t, err)
}

func TestMockClient_Unpin(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	testData := []byte("Test unpinning")

	// Upload (auto-pins)
	cid, err := client.Upload(ctx, testData)
	require.NoError(t, err)
	assert.True(t, client.IsPinned(cid))

	// Unpin
	err = client.Unpin(ctx, cid)
	require.NoError(t, err)
	assert.False(t, client.IsPinned(cid))

	// Unpin non-existent CID
	err = client.Unpin(ctx, "QmInvalidCID")
	assert.Error(t, err)
}

func TestMockClient_CalculateHash(t *testing.T) {
	client := NewMockClient()

	testData := []byte("Hash this!")
	hash := client.CalculateHash(testData)

	// Verify hash length
	assert.Equal(t, 32, len(hash), "SHA256 hash should be 32 bytes")

	// Verify hash is correct
	expectedHash := sha256.Sum256(testData)
	assert.Equal(t, expectedHash[:], hash)
}

func TestMockClient_VerifyHash(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	testData := []byte("Verify me!")

	// Upload
	cid, err := client.Upload(ctx, testData)
	require.NoError(t, err)

	// Download
	downloaded, err := client.Download(ctx, cid)
	require.NoError(t, err)

	// Calculate hash of original
	expectedHash := client.CalculateHash(testData)

	// Verify hash matches
	assert.True(t, client.VerifyHash(downloaded, expectedHash))

	// Verify hash doesn't match wrong data
	wrongHash := client.CalculateHash([]byte("wrong data"))
	assert.False(t, client.VerifyHash(downloaded, wrongHash))
}

func TestMockClient_IsConnected(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	// Mock client is always connected
	assert.True(t, client.IsConnected(ctx))
}

func TestMockClient_GetNodeID(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	nodeID, err := client.GetNodeID(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, nodeID)
}

func TestClient_VerifyHashHex(t *testing.T) {
	client := NewMockClient()

	testData := []byte("Test hex verification")
	hash := client.CalculateHash(testData)
	hashHex := hex.EncodeToString(hash)

	// Test with Client wrapper
	realClient := &Client{
		config: DefaultConfig(),
	}

	// Verify with correct hex hash
	assert.True(t, realClient.VerifyHashHex(testData, hashHex))

	// Verify with wrong hex hash
	wrongHash := hex.EncodeToString(client.CalculateHash([]byte("wrong")))
	assert.False(t, realClient.VerifyHashHex(testData, wrongHash))

	// Verify with invalid hex
	assert.False(t, realClient.VerifyHashHex(testData, "invalid-hex"))
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, "http://localhost:5001", config.APIEndpoint)
	assert.True(t, config.AutoPin)
	assert.Equal(t, 3, config.MaxRetries)
	assert.NotZero(t, config.Timeout)
	assert.NotZero(t, config.RetryDelay)
}

func TestNewClient_WithConfig(t *testing.T) {
	// Test with nil config (should use defaults)
	client, err := NewClient(nil)
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.config)

	// Test with custom config
	config := &Config{
		APIEndpoint: "http://custom:5001",
		Timeout:     60000000000, // 60 seconds in nanoseconds
		AutoPin:     false,
		MaxRetries:  5,
	}
	client, err = NewClient(config)
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "http://custom:5001", client.config.APIEndpoint)
	assert.False(t, client.config.AutoPin)

	// Test with empty endpoint (should error)
	config = &Config{
		APIEndpoint: "",
	}
	_, err = NewClient(config)
	assert.Error(t, err)
}

func TestUploadDownloadIntegration(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	// Test multiple uploads and downloads
	testCases := []struct {
		name string
		data []byte
	}{
		{"Small text", []byte("Small text")},
		{"Large text", []byte("This is a much larger text that should still work fine with IPFS")},
		{"Binary data", []byte{0x00, 0x01, 0x02, 0xFF, 0xFE}},
		{"JSON", []byte(`{"key": "value", "number": 123}`)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Upload
			cid, err := client.Upload(ctx, tc.data)
			require.NoError(t, err)
			assert.NotEmpty(t, cid)

			// Download
			downloaded, err := client.Download(ctx, cid)
			require.NoError(t, err)

			// Verify data matches
			assert.Equal(t, tc.data, downloaded)

			// Verify hash
			hash := client.CalculateHash(tc.data)
			assert.True(t, client.VerifyHash(downloaded, hash))
		})
	}
}

func TestConcurrentOperations(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	// Test concurrent uploads with proper synchronization
	numGoroutines := 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer func() { done <- true }()

			data := []byte(fmt.Sprintf("Concurrent upload %d", index))
			cid, err := client.Upload(ctx, data)
			if err != nil {
				t.Errorf("Upload failed: %v", err)
				return
			}

			// Download and verify
			downloaded, err := client.Download(ctx, cid)
			if err != nil {
				t.Errorf("Download failed: %v", err)
				return
			}

			if !bytes.Equal(data, downloaded) {
				t.Errorf("Data mismatch: expected %v, got %v", data, downloaded)
			}
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}
