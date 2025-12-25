// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package ipfs

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"time"
)

// NOTE: These examples demonstrate IPFS client usage.
// For complete Data Registry integration examples including keeper usage,
// see the README.md file and the keeper package tests.

// Example1_BasicUploadDownload demonstrates basic IPFS upload and download
func Example1_BasicUploadDownload() {
	ctx := context.Background()

	fmt.Println("=== Example 1: Basic Upload and Download ===")

	// Create mock client (no IPFS daemon needed)
	client := NewMockClient()

	// Sample image data (in real usage, this would be actual image bytes)
	imageData := []byte("Sample JPEG image content - in production this would be actual image bytes")
	fmt.Printf("Content size: %s\n", FormatSize(int64(len(imageData))))

	// Upload to IPFS
	cid, err := client.Upload(ctx, imageData)
	if err != nil {
		fmt.Printf("Error uploading: %v\n", err)
		return
	}

	fmt.Printf("\nImage uploaded successfully!\n")
	fmt.Printf("IPFS CID: %s\n", cid)
	fmt.Printf("Gateway URL: %s\n", BuildIPFSGatewayURL(cid, ""))

	// Calculate content hash
	hash := client.CalculateHash(imageData)
	fmt.Printf("Content Hash (SHA256): %s\n", hex.EncodeToString(hash))

	// Download content
	fmt.Println("\nDownloading content from IPFS...")
	downloaded, err := client.Download(ctx, cid)
	if err != nil {
		fmt.Printf("Download failed: %v\n", err)
		return
	}

	fmt.Printf("Downloaded %d bytes\n", len(downloaded))

	// Verify hash
	if client.VerifyHash(downloaded, hash) {
		fmt.Println("Hash verification: PASSED")
	} else {
		fmt.Println("Hash verification: FAILED")
	}
}

// Example2_RealIPFSConnection demonstrates connecting to a real IPFS node
func Example2_RealIPFSConnection() {
	ctx := context.Background()

	fmt.Println("\n=== Example 2: Connect to Real IPFS Node ===")

	// Create real IPFS client config
	config := &Config{
		APIEndpoint: "http://localhost:5001",
		Timeout:     30 * time.Second,
		AutoPin:     true,
		MaxRetries:  3,
		RetryDelay:  1 * time.Second,
	}

	// Create client
	client, err := NewClient(config)
	if err != nil {
		fmt.Printf("Failed to create IPFS client: %v\n", err)
		fmt.Println("\nNote: This requires IPFS daemon running on localhost:5001")
		fmt.Println("Start IPFS with: ipfs daemon")
		fmt.Println("Or use IPFS Desktop: https://docs.ipfs.tech/install/ipfs-desktop/")
		return
	}

	// Check connection
	if !client.IsConnected(ctx) {
		fmt.Println("IPFS node is not reachable")
		fmt.Println("Make sure IPFS daemon is running")
		return
	}

	// Get node ID
	nodeID, err := client.GetNodeID(ctx)
	if err != nil {
		fmt.Printf("Failed to get node ID: %v\n", err)
		return
	}

	fmt.Printf("Connected to IPFS node successfully!\n")
	fmt.Printf("Node ID: %s\n", nodeID)
	fmt.Println("\nYou can now upload and download content to/from this IPFS node")
}

// Example3_ContentTypeDetection demonstrates content type detection
func Example3_ContentTypeDetection() {
	fmt.Println("\n=== Example 3: Content Type Detection ===")

	testCases := []struct {
		name     string
		data     []byte
		filename string
	}{
		{
			name:     "JPEG Image",
			data:     []byte{0xFF, 0xD8, 0xFF, 0xE0},
			filename: "photo.jpg",
		},
		{
			name:     "PNG Image",
			data:     []byte{0x89, 0x50, 0x4E, 0x47},
			filename: "screenshot.png",
		},
		{
			name:     "PDF Document",
			data:     []byte("%PDF-1.4"),
			filename: "scorecard.pdf",
		},
		{
			name:     "JSON Data",
			data:     []byte(`{"score": 72}`),
			filename: "data.json",
		},
		{
			name:     "Text File",
			data:     []byte("Par 4, 380 yards"),
			filename: "notes.txt",
		},
	}

	for _, tc := range testCases {
		contentType := DetectContentType(tc.data, tc.filename)
		ext := ExtractFileExtension(tc.filename, contentType)

		fmt.Printf("\n%s:\n", tc.name)
		fmt.Printf("  Filename: %s\n", tc.filename)
		fmt.Printf("  Content Type: %s\n", contentType)
		fmt.Printf("  Extension: %s\n", ext)
		fmt.Printf("  Is Image: %v\n", IsImageType(contentType))
		fmt.Printf("  Is Document: %v\n", IsDocumentType(contentType))
	}
}

// Example4_PinningContent demonstrates pinning and unpinning
func Example4_PinningContent() {
	ctx := context.Background()

	fmt.Println("\n=== Example 4: Pinning and Unpinning Content ===")

	client := NewMockClient()

	// Upload content (automatically pinned by default)
	content := []byte("Important content that should persist")
	cid, err := client.Upload(ctx, content)
	if err != nil {
		fmt.Printf("Upload failed: %v\n", err)
		return
	}

	fmt.Printf("Content uploaded and pinned: %s\n", cid)

	// Check if pinned (mock client specific method)
	if client.IsPinned(cid) {
		fmt.Println("Content is pinned (will not be garbage collected)")
	}

	// Unpin content
	err = client.Unpin(ctx, cid)
	if err != nil {
		fmt.Printf("Unpin failed: %v\n", err)
		return
	}

	fmt.Println("\nContent unpinned")
	if !client.IsPinned(cid) {
		fmt.Println("Content is now unpinned (may be garbage collected)")
	}

	// Re-pin content
	err = client.Pin(ctx, cid)
	if err != nil {
		fmt.Printf("Pin failed: %v\n", err)
		return
	}

	fmt.Println("\nContent re-pinned")
	if client.IsPinned(cid) {
		fmt.Println("Content is pinned again")
	}
}

// Example5_ErrorHandling demonstrates error handling
func Example5_ErrorHandling() {
	ctx := context.Background()

	fmt.Println("\n=== Example 5: Error Handling ===")

	client := NewMockClient()

	// Error 1: Empty content
	fmt.Println("\n1. Uploading empty content:")
	_, err := client.Upload(ctx, []byte{})
	if err != nil {
		fmt.Printf("Expected error: %v\n", err)
	}

	// Error 2: Invalid CID
	fmt.Println("\n2. Downloading invalid CID:")
	_, err = client.Download(ctx, "invalid-cid")
	if err != nil {
		fmt.Printf("Expected error: %v\n", err)
	}

	// Error 3: Non-existent CID
	fmt.Println("\n3. Downloading non-existent CID:")
	_, err = client.Download(ctx, "QmNonExistentCID123456789012345678901234567890")
	if err != nil {
		fmt.Printf("Expected error: %v\n", err)
	}

	// Error 4: Hash mismatch
	fmt.Println("\n4. Detecting content tampering:")
	originalContent := []byte("Original content")
	tamperedContent := []byte("Tampered content")
	originalHash := client.CalculateHash(originalContent)

	if !client.VerifyHash(tamperedContent, originalHash) {
		fmt.Println("Hash mismatch detected! Content may have been tampered with.")
	}
}

// Example6_HashCalculation demonstrates hash calculation and verification
func Example6_HashCalculation() {
	fmt.Println("\n=== Example 6: Hash Calculation and Verification ===")

	client := NewMockClient()

	content := []byte("This is the content we want to hash and verify")

	// Calculate hash (bytes)
	hashBytes := client.CalculateHash(content)
	fmt.Printf("Hash (bytes): %d bytes\n", len(hashBytes))

	// Calculate hash (hex string)
	hashHex := CalculateSHA256(content)
	fmt.Printf("Hash (hex): %s\n", hashHex)

	// Verify the two are equal
	fmt.Printf("Hex matches bytes: %v\n", hashHex == hex.EncodeToString(hashBytes))

	// Verify content
	fmt.Println("\nVerifying correct content:")
	if client.VerifyHash(content, hashBytes) {
		fmt.Println("Verification PASSED")
	}

	// Verify incorrect content
	fmt.Println("\nVerifying incorrect content:")
	wrongContent := []byte("Different content")
	if !client.VerifyHash(wrongContent, hashBytes) {
		fmt.Println("Verification FAILED (as expected)")
	}

	// Verify with hex hash
	fmt.Println("\nVerifying with hex hash:")
	realClient := &Client{config: DefaultConfig()}
	if realClient.VerifyHashHex(content, hashHex) {
		fmt.Println("Hex verification PASSED")
	}
}

// Example7_CIDValidation demonstrates CID validation
func Example7_CIDValidation() {
	fmt.Println("\n=== Example 7: CID Validation ===")

	testCases := []struct {
		cid   string
		valid bool
		desc  string
	}{
		{
			cid:   "QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG",
			valid: true,
			desc:  "Valid CIDv0",
		},
		{
			cid:   "bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi",
			valid: true,
			desc:  "Valid CIDv1 (bafy)",
		},
		{
			cid:   "bafkreifjjcie6lypi6ny7amxnfftagclbuxndqonfipmb64f2km2devei4",
			valid: true,
			desc:  "Valid CIDv1 (bafk)",
		},
		{
			cid:   "not-a-valid-cid",
			valid: false,
			desc:  "Invalid CID",
		},
		{
			cid:   "",
			valid: false,
			desc:  "Empty CID",
		},
		{
			cid:   "Qm123",
			valid: false,
			desc:  "Too short CID",
		},
	}

	for _, tc := range testCases {
		result := IsValidCID(tc.cid)
		status := "INVALID"
		if result {
			status = "VALID"
		}

		fmt.Printf("\n%s:\n", tc.desc)
		fmt.Printf("  CID: %s\n", FormatCID(tc.cid))
		fmt.Printf("  Status: %s", status)

		if result == tc.valid {
			fmt.Println(" ✓")
		} else {
			fmt.Println(" ✗ UNEXPECTED")
		}
	}
}

// Example8_FileOperations demonstrates file operations
func Example8_FileOperations() {
	ctx := context.Background()

	fmt.Println("\n=== Example 8: File Operations ===")

	client := NewMockClient()

	// Create sample file content
	fileContent := []byte("This is a sample file that will be uploaded to IPFS")
	filename := "sample.txt"

	fmt.Printf("Original file: %s (%s)\n", filename, FormatSize(int64(len(fileContent))))

	// Detect content type
	contentType := DetectContentType(fileContent, filename)
	fmt.Printf("Content type: %s\n", contentType)

	// Sanitize filename
	safeFilename := SanitizeFilename(filename)
	fmt.Printf("Safe filename: %s\n", safeFilename)

	// Upload to IPFS
	cid, err := client.Upload(ctx, fileContent)
	if err != nil {
		fmt.Printf("Upload failed: %v\n", err)
		return
	}

	fmt.Printf("\nFile uploaded to IPFS: %s\n", cid)

	// Download from IPFS
	downloaded, err := client.Download(ctx, cid)
	if err != nil {
		fmt.Printf("Download failed: %v\n", err)
		return
	}

	// Save to file (demonstration - commented out to avoid creating files)
	fmt.Printf("\nDownloaded %s\n", FormatSize(int64(len(downloaded))))
	fmt.Printf("Would save to: downloaded_%s\n", safeFilename)

	// In real usage:
	// err = os.WriteFile("downloaded_"+safeFilename, downloaded, 0644)
}

// Example9_BatchOperations demonstrates batch upload/download
func Example9_BatchOperations() {
	ctx := context.Background()

	fmt.Println("\n=== Example 9: Batch Operations ===")

	client := NewMockClient()

	// Batch upload
	files := []struct {
		name    string
		content []byte
	}{
		{"file1.txt", []byte("Content of file 1")},
		{"file2.txt", []byte("Content of file 2")},
		{"file3.txt", []byte("Content of file 3")},
	}

	cids := make([]string, 0, len(files))

	fmt.Println("Uploading files...")
	for _, file := range files {
		cid, err := client.Upload(ctx, file.content)
		if err != nil {
			fmt.Printf("Failed to upload %s: %v\n", file.name, err)
			continue
		}

		cids = append(cids, cid)
		fmt.Printf("  %s → %s\n", file.name, FormatCID(cid))
	}

	fmt.Printf("\nUploaded %d files\n", len(cids))

	// Batch download
	fmt.Println("\nDownloading files...")
	for i, cid := range cids {
		content, err := client.Download(ctx, cid)
		if err != nil {
			fmt.Printf("Failed to download %s: %v\n", cid, err)
			continue
		}

		fmt.Printf("  File %d: %s (%d bytes)\n", i+1, FormatCID(cid), len(content))
	}
}

// Example10_ConfigurationOptions demonstrates different configuration options
func Example10_ConfigurationOptions() {
	fmt.Println("\n=== Example 10: Configuration Options ===")

	// Default configuration
	fmt.Println("\n1. Default Configuration:")
	defaultConfig := DefaultConfig()
	fmt.Printf("  Endpoint: %s\n", defaultConfig.APIEndpoint)
	fmt.Printf("  Timeout: %v\n", defaultConfig.Timeout)
	fmt.Printf("  Auto-pin: %v\n", defaultConfig.AutoPin)
	fmt.Printf("  Max retries: %d\n", defaultConfig.MaxRetries)
	fmt.Printf("  Retry delay: %v\n", defaultConfig.RetryDelay)

	// Custom configuration for large files
	fmt.Println("\n2. Configuration for Large Files:")
	largeFileConfig := &Config{
		APIEndpoint: "http://localhost:5001",
		Timeout:     5 * time.Minute,
		AutoPin:     true,
		MaxRetries:  5,
		RetryDelay:  5 * time.Second,
	}
	fmt.Printf("  Timeout: %v (longer for large files)\n", largeFileConfig.Timeout)
	fmt.Printf("  Max retries: %d\n", largeFileConfig.MaxRetries)
	fmt.Printf("  Retry delay: %v\n", largeFileConfig.RetryDelay)

	// Configuration for production
	fmt.Println("\n3. Production Configuration:")
	prodConfig := &Config{
		APIEndpoint: "http://ipfs-prod.internal:5001",
		Timeout:     60 * time.Second,
		AutoPin:     true,
		MaxRetries:  5,
		RetryDelay:  2 * time.Second,
	}
	fmt.Printf("  Endpoint: %s (dedicated node)\n", prodConfig.APIEndpoint)
	fmt.Printf("  Timeout: %v\n", prodConfig.Timeout)
	fmt.Printf("  Max retries: %d (higher for reliability)\n", prodConfig.MaxRetries)

	// Configuration for testing
	fmt.Println("\n4. Testing Configuration (Mock Client):")
	fmt.Println("  No configuration needed!")
	fmt.Println("  Use: client := ipfs.NewMockClient()")
	fmt.Println("  Perfect for unit tests and CI/CD")
}

// RunAllExamples runs all example functions
func RunAllExamples() {
	fmt.Println("================================================")
	fmt.Println("       AURA IPFS Integration - Examples")
	fmt.Println("================================================")

	Example1_BasicUploadDownload()
	Example2_RealIPFSConnection()
	Example3_ContentTypeDetection()
	Example4_PinningContent()
	Example5_ErrorHandling()
	Example6_HashCalculation()
	Example7_CIDValidation()
	Example8_FileOperations()
	Example9_BatchOperations()
	Example10_ConfigurationOptions()

	fmt.Println("\n================================================")
	fmt.Println("       All examples completed!")
	fmt.Println("================================================")
	fmt.Println("\nFor Data Registry integration examples, see:")
	fmt.Println("  - README.md (comprehensive documentation)")
	fmt.Println("  - keeper package tests")
	fmt.Println("  - msg_server integration")
}

// SaveExample saves content to a file (helper for examples)
func SaveExample(filename string, content []byte) error {
	return os.WriteFile(filename, content, 0644)
}

// LoadExample loads content from a file (helper for examples)
func LoadExample(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}
