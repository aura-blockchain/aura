// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package privacy

// This file implements OFF-CHAIN memo encryption utilities for privacy-preserving messages.
//
// IMPORTANT: All encryption functions are OFF-CHAIN utilities that use crypto/rand
// for nonces and ephemeral keys. NEVER call from consensus code.
//
// Encrypted memos allow attaching private messages to transactions:
// - Sender encrypts memo off-chain using recipient's public key (ECIES-style)
// - Encrypted memo is submitted with transaction as opaque bytes
// - Only recipient with private key can decrypt
// - Blockchain stores encrypted bytes without ever decrypting or re-encrypting
//
// Supported encryption algorithms:
// - AES-256-GCM: Fast, widely supported, 128-bit security
// - ChaCha20-Poly1305: Fast, constant-time, 256-bit security
// - XChaCha20-Poly1305: Extended nonce for better collision resistance
//
// ON-CHAIN VS OFF-CHAIN SEPARATION:
// - OFF-CHAIN: Encrypt(), encryptAESGCM(), encryptChaCha20Poly1305(), etc.
//   These functions encrypt memos using cryptographic randomness (nonces, ephemeral keys).
//   Used by wallet software before submitting transactions.
//
// - ON-CHAIN: Message handlers store encrypted memos as-is (opaque bytes).
//   No encryption, decryption, or cryptographic operations during consensus.
//   The blockchain is a transport layer for encrypted data.
//
// Encryption happens entirely off-chain. The blockchain only stores and retrieves
// encrypted bytes without performing any cryptographic operations.

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/chacha20poly1305"
)

const (
	// Encryption algorithms
	AlgorithmAES256GCM         = "AES-256-GCM"
	AlgorithmChaCha20Poly1305  = "ChaCha20-Poly1305"
	AlgorithmXChaCha20Poly1305 = "XChaCha20-Poly1305"
)

// MemoEncryptor handles encryption of transaction memos
type MemoEncryptor struct {
	algorithm string
	curve     elliptic.Curve
}

// NewMemoEncryptor creates a new memo encryptor
func NewMemoEncryptor(algorithm string) *MemoEncryptor {
	return &MemoEncryptor{
		algorithm: algorithm,
		curve:     elliptic.P256(),
	}
}

// EncryptedMemo represents an encrypted memo
type EncryptedMemo struct {
	Algorithm      string
	Nonce          []byte
	Ciphertext     []byte
	EphemeralPubKey []byte // For ECIES-style encryption
}

// Encrypt encrypts a memo for a recipient
func (me *MemoEncryptor) Encrypt(memo []byte, recipientPubKey []byte) (*EncryptedMemo, error) {
	if len(memo) == 0 {
		return nil, errors.New("memo cannot be empty")
	}
	if len(recipientPubKey) == 0 {
		return nil, errors.New("recipient public key cannot be empty")
	}

	switch me.algorithm {
	case AlgorithmAES256GCM:
		return me.encryptAESGCM(memo, recipientPubKey)
	case AlgorithmChaCha20Poly1305:
		return me.encryptChaCha20Poly1305(memo, recipientPubKey)
	case AlgorithmXChaCha20Poly1305:
		return me.encryptXChaCha20Poly1305(memo, recipientPubKey)
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", me.algorithm)
	}
}

// Decrypt decrypts an encrypted memo
func (me *MemoEncryptor) Decrypt(encryptedMemo *EncryptedMemo, privateKey []byte) ([]byte, error) {
	if encryptedMemo == nil {
		return nil, errors.New("encrypted memo cannot be nil")
	}
	if len(privateKey) == 0 {
		return nil, errors.New("private key cannot be empty")
	}

	switch encryptedMemo.Algorithm {
	case AlgorithmAES256GCM:
		return me.decryptAESGCM(encryptedMemo, privateKey)
	case AlgorithmChaCha20Poly1305:
		return me.decryptChaCha20Poly1305(encryptedMemo, privateKey)
	case AlgorithmXChaCha20Poly1305:
		return me.decryptXChaCha20Poly1305(encryptedMemo, privateKey)
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", encryptedMemo.Algorithm)
	}
}

// encryptAESGCM encrypts using AES-256-GCM
func (me *MemoEncryptor) encryptAESGCM(memo []byte, recipientPubKey []byte) (*EncryptedMemo, error) {
	// Derive shared secret using ECDH
	sharedSecret, ephemeralPubKey, err := me.deriveSharedSecret(recipientPubKey)
	if err != nil {
		return nil, err
	}

	// Derive encryption key from shared secret
	key := me.deriveKey(sharedSecret, 32)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt
	ciphertext := gcm.Seal(nil, nonce, memo, nil)

	return &EncryptedMemo{
		Algorithm:      AlgorithmAES256GCM,
		Nonce:          nonce,
		Ciphertext:     ciphertext,
		EphemeralPubKey: ephemeralPubKey,
	}, nil
}

// decryptAESGCM decrypts using AES-256-GCM
func (me *MemoEncryptor) decryptAESGCM(encryptedMemo *EncryptedMemo, privateKey []byte) ([]byte, error) {
	// Recover shared secret
	sharedSecret, err := me.recoverSharedSecret(encryptedMemo.EphemeralPubKey, privateKey)
	if err != nil {
		return nil, err
	}

	// Derive encryption key
	key := me.deriveKey(sharedSecret, 32)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decrypt
	plaintext, err := gcm.Open(nil, encryptedMemo.Nonce, encryptedMemo.Ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// encryptChaCha20Poly1305 encrypts using ChaCha20-Poly1305
func (me *MemoEncryptor) encryptChaCha20Poly1305(memo []byte, recipientPubKey []byte) (*EncryptedMemo, error) {
	// Derive shared secret
	sharedSecret, ephemeralPubKey, err := me.deriveSharedSecret(recipientPubKey)
	if err != nil {
		return nil, err
	}

	// Derive encryption key
	key := me.deriveKey(sharedSecret, chacha20poly1305.KeySize)

	// Create ChaCha20-Poly1305 AEAD
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create ChaCha20-Poly1305: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt
	ciphertext := aead.Seal(nil, nonce, memo, nil)

	return &EncryptedMemo{
		Algorithm:      AlgorithmChaCha20Poly1305,
		Nonce:          nonce,
		Ciphertext:     ciphertext,
		EphemeralPubKey: ephemeralPubKey,
	}, nil
}

// decryptChaCha20Poly1305 decrypts using ChaCha20-Poly1305
func (me *MemoEncryptor) decryptChaCha20Poly1305(encryptedMemo *EncryptedMemo, privateKey []byte) ([]byte, error) {
	// Recover shared secret
	sharedSecret, err := me.recoverSharedSecret(encryptedMemo.EphemeralPubKey, privateKey)
	if err != nil {
		return nil, err
	}

	// Derive encryption key
	key := me.deriveKey(sharedSecret, chacha20poly1305.KeySize)

	// Create ChaCha20-Poly1305 AEAD
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create ChaCha20-Poly1305: %w", err)
	}

	// Decrypt
	plaintext, err := aead.Open(nil, encryptedMemo.Nonce, encryptedMemo.Ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// encryptXChaCha20Poly1305 encrypts using XChaCha20-Poly1305
func (me *MemoEncryptor) encryptXChaCha20Poly1305(memo []byte, recipientPubKey []byte) (*EncryptedMemo, error) {
	// Derive shared secret
	sharedSecret, ephemeralPubKey, err := me.deriveSharedSecret(recipientPubKey)
	if err != nil {
		return nil, err
	}

	// Derive encryption key
	key := me.deriveKey(sharedSecret, chacha20poly1305.KeySize)

	// Create XChaCha20-Poly1305 AEAD
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create XChaCha20-Poly1305: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt
	ciphertext := aead.Seal(nil, nonce, memo, nil)

	return &EncryptedMemo{
		Algorithm:      AlgorithmXChaCha20Poly1305,
		Nonce:          nonce,
		Ciphertext:     ciphertext,
		EphemeralPubKey: ephemeralPubKey,
	}, nil
}

// decryptXChaCha20Poly1305 decrypts using XChaCha20-Poly1305
func (me *MemoEncryptor) decryptXChaCha20Poly1305(encryptedMemo *EncryptedMemo, privateKey []byte) ([]byte, error) {
	// Recover shared secret
	sharedSecret, err := me.recoverSharedSecret(encryptedMemo.EphemeralPubKey, privateKey)
	if err != nil {
		return nil, err
	}

	// Derive encryption key
	key := me.deriveKey(sharedSecret, chacha20poly1305.KeySize)

	// Create XChaCha20-Poly1305 AEAD
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create XChaCha20-Poly1305: %w", err)
	}

	// Decrypt
	plaintext, err := aead.Open(nil, encryptedMemo.Nonce, encryptedMemo.Ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// deriveSharedSecret derives a shared secret using ECDH
func (me *MemoEncryptor) deriveSharedSecret(recipientPubKey []byte) ([]byte, []byte, error) {
	// Generate ephemeral key pair
	ephemeralPriv, ephemeralX, ephemeralY, err := elliptic.GenerateKey(me.curve, rand.Reader) //nolint:staticcheck // legacy curve use retained for backward compatibility
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate ephemeral key: %w", err)
	}

	ephemeralPubKey := elliptic.MarshalCompressed(me.curve, ephemeralX, ephemeralY)

	// Unmarshal recipient's public key
	recipientX, recipientY := elliptic.UnmarshalCompressed(me.curve, recipientPubKey)
	if recipientX == nil {
		return nil, nil, errors.New("invalid recipient public key")
	}

	// Compute shared secret: ephemeralPriv * recipientPubKey
	sharedX, sharedY := me.curve.ScalarMult(recipientX, recipientY, ephemeralPriv)
	sharedSecret := elliptic.MarshalCompressed(me.curve, sharedX, sharedY)

	return sharedSecret, ephemeralPubKey, nil
}

// recoverSharedSecret recovers the shared secret using the ephemeral public key and private key
func (me *MemoEncryptor) recoverSharedSecret(ephemeralPubKey []byte, privateKey []byte) ([]byte, error) {
	// Unmarshal ephemeral public key
	ephemeralX, ephemeralY := elliptic.UnmarshalCompressed(me.curve, ephemeralPubKey)
	if ephemeralX == nil {
		return nil, errors.New("invalid ephemeral public key")
	}

	// Compute shared secret: privateKey * ephemeralPubKey
	sharedX, sharedY := me.curve.ScalarMult(ephemeralX, ephemeralY, privateKey)
	sharedSecret := elliptic.MarshalCompressed(me.curve, sharedX, sharedY)

	return sharedSecret, nil
}

// deriveKey derives an encryption key from a shared secret
func (me *MemoEncryptor) deriveKey(sharedSecret []byte, keySize int) []byte {
	hasher := sha256.New()
	hasher.Write(sharedSecret)
	hasher.Write([]byte("memo_encryption"))
	key := hasher.Sum(nil)
	return key[:keySize]
}
