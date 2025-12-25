// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package ipfs

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidCID(t *testing.T) {
	tests := []struct {
		name  string
		cid   string
		valid bool
	}{
		// Valid CIDv0
		{"Valid CIDv0", "QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG", true},
		{"Valid CIDv0 2", "QmPZ9gcCEpqKTo6aq61g2nXGUhM4iCL3ewB6LDXZCtioEB", true},

		// Valid CIDv1
		{"Valid CIDv1 bafy", "bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi", true},
		{"Valid CIDv1 bafk", "bafkreifjjcie6lypi6ny7amxnfftagclbuxndqonfipmb64f2km2devei4", true},

		// Invalid CIDs
		{"Empty string", "", false},
		{"Too short", "Qm123", false},
		{"Invalid prefix", "Xm" + "YwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG", false},
		{"Random string", "not-a-cid-at-all", false},
		{"Only letters", "abcdefghijklmnop", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidCID(tt.cid)
			assert.Equal(t, tt.valid, result, "CID: %s", tt.cid)
		})
	}
}

func TestDetectContentType(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		filename string
		expected string
		contains string // Alternative: check if result contains this
	}{
		{
			name:     "Empty data",
			data:     []byte{},
			filename: "",
			expected: "application/octet-stream",
		},
		{
			name:     "JPEG image",
			data:     []byte{0xFF, 0xD8, 0xFF},
			filename: "test.jpg",
			contains: "image/jpeg",
		},
		{
			name:     "PNG image",
			data:     []byte{0x89, 0x50, 0x4E, 0x47},
			filename: "test.png",
			contains: "image/png",
		},
		{
			name:     "PDF from extension",
			data:     []byte("%PDF-1.4"),
			filename: "document.pdf",
			expected: "application/pdf",
		},
		{
			name:     "JSON from extension",
			data:     []byte(`{"key": "value"}`),
			filename: "data.json",
			contains: "application/json",
		},
		{
			name:     "Plain text",
			data:     []byte("Hello, World!"),
			filename: "test.txt",
			contains: "text/plain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectContentType(tt.data, tt.filename)
			if tt.expected != "" {
				assert.Equal(t, tt.expected, result)
			} else if tt.contains != "" {
				assert.Contains(t, result, tt.contains)
			}
		})
	}
}

func TestContentTypeChecks(t *testing.T) {
	assert.True(t, IsImageType("image/jpeg"))
	assert.True(t, IsImageType("image/png"))
	assert.False(t, IsImageType("video/mp4"))

	assert.True(t, IsVideoType("video/mp4"))
	assert.True(t, IsVideoType("video/webm"))
	assert.False(t, IsVideoType("image/jpeg"))

	assert.True(t, IsAudioType("audio/mpeg"))
	assert.True(t, IsAudioType("audio/wav"))
	assert.False(t, IsAudioType("video/mp4"))

	assert.True(t, IsDocumentType("application/pdf"))
	assert.True(t, IsDocumentType("text/plain"))
	assert.True(t, IsDocumentType("application/vnd.openxmlformats-officedocument.wordprocessingml.document"))
	assert.False(t, IsDocumentType("image/jpeg"))
}

func TestCalculateSHA256(t *testing.T) {
	testData := []byte("Test data for hashing")

	// Test hex output
	hashHex := CalculateSHA256(testData)
	assert.Equal(t, 64, len(hashHex), "SHA256 hex string should be 64 characters")
	assert.Regexp(t, "^[a-f0-9]{64}$", hashHex, "Should be valid hex")

	// Test bytes output
	hashBytes := CalculateSHA256Bytes(testData)
	assert.Equal(t, 32, len(hashBytes), "SHA256 should be 32 bytes")

	// Verify they match
	assert.Equal(t, hashHex, hex.EncodeToString(hashBytes))

	// Test consistency
	hash1 := CalculateSHA256(testData)
	hash2 := CalculateSHA256(testData)
	assert.Equal(t, hash1, hash2, "Hash should be consistent")

	// Test different data produces different hash
	hash3 := CalculateSHA256([]byte("Different data"))
	assert.NotEqual(t, hash1, hash3, "Different data should produce different hash")
}

func TestVerifyContentHash(t *testing.T) {
	testData := []byte("Verify this content")
	correctHash := CalculateSHA256Bytes(testData)
	wrongHash := CalculateSHA256Bytes([]byte("Wrong content"))

	// Test with bytes
	assert.True(t, VerifyContentHash(testData, correctHash))
	assert.False(t, VerifyContentHash(testData, wrongHash))

	// Test with hex
	correctHashHex := CalculateSHA256(testData)
	wrongHashHex := CalculateSHA256([]byte("Wrong content"))

	assert.True(t, VerifyContentHashHex(testData, correctHashHex))
	assert.False(t, VerifyContentHashHex(testData, wrongHashHex))

	// Test with invalid hex
	assert.False(t, VerifyContentHashHex(testData, "invalid-hex-string"))
}

func TestFormatCID(t *testing.T) {
	tests := []struct {
		name     string
		cid      string
		expected string
	}{
		{"Short CID", "QmShort", "QmShort"},
		{"Long CID", "QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG", "QmYwAPJzv5...79ojWnPbdG"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatCID(tt.cid)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{100, "100 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
		{1099511627776, "1.0 TB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := FormatSize(tt.bytes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateDataSize(t *testing.T) {
	// Valid sizes
	assert.NoError(t, ValidateDataSize(100, 1000))
	assert.NoError(t, ValidateDataSize(1000, 1000))
	assert.NoError(t, ValidateDataSize(100, 0)) // No limit

	// Invalid sizes
	assert.Error(t, ValidateDataSize(0, 1000))    // Zero size
	assert.Error(t, ValidateDataSize(-1, 1000))   // Negative size
	assert.Error(t, ValidateDataSize(1001, 1000)) // Exceeds limit
}

func TestExtractFileExtension(t *testing.T) {
	tests := []struct {
		filename    string
		contentType string
		expected    string
		contains    bool // if true, check if result contains expected, not exact match
	}{
		{"test.jpg", "", ".jpg", false},
		{"document.pdf", "", ".pdf", false},
		{"", "image/jpeg", ".j", true}, // Can be .jpeg, .jpe, or .jfif depending on system
		{"", "application/pdf", ".pdf", false},
		{"file.txt", "text/plain", ".txt", false},
		{"noext", "", "", false},
		{"", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename+"|"+tt.contentType, func(t *testing.T) {
			result := ExtractFileExtension(tt.filename, tt.contentType)
			if tt.contains {
				assert.Contains(t, result, tt.expected)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input       string
		contains    string
		notContains string
	}{
		{"normal.txt", "normal.txt", ""},
		{"../../etc/passwd", "passwd", ".."},
		{"file$name.txt", "file_name.txt", "$"},
		{"file|name.txt", "file_name.txt", "|"},
		{"file;name.txt", "file_name.txt", ";"},
		{"/path/to/file.txt", "file.txt", "/path"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := SanitizeFilename(tt.input)
			if tt.contains != "" {
				assert.Contains(t, result, tt.contains)
			}
			if tt.notContains != "" {
				assert.NotContains(t, result, tt.notContains)
			}
		})
	}

	// Test length limiting
	longName := string(make([]byte, 300))
	result := SanitizeFilename(longName)
	assert.LessOrEqual(t, len(result), 255)
}

func TestBuildIPFSGatewayURL(t *testing.T) {
	cid := "QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG"

	tests := []struct {
		gateway  string
		expected string
	}{
		{"", "https://ipfs.io/ipfs/" + cid},
		{"https://ipfs.io", "https://ipfs.io/ipfs/" + cid},
		{"https://ipfs.io/", "https://ipfs.io/ipfs/" + cid},
		{"https://gateway.pinata.cloud", "https://gateway.pinata.cloud/ipfs/" + cid},
	}

	for _, tt := range tests {
		t.Run(tt.gateway, func(t *testing.T) {
			result := BuildIPFSGatewayURL(cid, tt.gateway)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContentTypeToDataItemType(t *testing.T) {
	tests := []struct {
		contentType string
		expected    string
	}{
		{"image/jpeg", "PHOTO"},
		{"image/png", "PHOTO"},
		{"video/mp4", "VIDEO"},
		{"audio/mpeg", "AUDIO"},
		{"application/pdf", "DOCUMENT_PDF"},
		{"text/plain", "CUSTOM"},
		{"application/octet-stream", "CUSTOM"},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			result := ContentTypeToDataItemType(tt.contentType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIPFSError(t *testing.T) {
	baseErr := assert.AnError
	ipfsErr := NewIPFSError("upload", "QmTest", baseErr)

	assert.Error(t, ipfsErr)
	assert.Contains(t, ipfsErr.Error(), "upload")
	assert.Contains(t, ipfsErr.Error(), "QmTest")

	// Test without CID
	ipfsErr2 := NewIPFSError("connect", "", baseErr)
	assert.Error(t, ipfsErr2)
	assert.Contains(t, ipfsErr2.Error(), "connect")
}
