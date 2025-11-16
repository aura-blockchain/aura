package privacy

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"math/big"
	"sync"
	"time"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/sha3"
)

// EncryptionAlgorithm defines the encryption algorithm type
type EncryptionAlgorithm string

const (
	AlgorithmAES256GCM         EncryptionAlgorithm = "AES-256-GCM"
	AlgorithmChaCha20Poly1305  EncryptionAlgorithm = "ChaCha20-Poly1305"
	AlgorithmXChaCha20Poly1305 EncryptionAlgorithm = "XChaCha20-Poly1305"
)

// EncryptedMemo represents an encrypted transaction memo
type EncryptedMemo struct {
	Ciphertext         []byte
	Nonce              []byte
	Algorithm          EncryptionAlgorithm
	RecipientPublicKey []byte
	EphemeralPublicKey []byte
	AuthenticationTag  []byte
}

// MemoEncryptor handles encryption of transaction memos
type MemoEncryptor struct {
	algorithm EncryptionAlgorithm
}

// NewMemoEncryptor creates a new memo encryptor
func NewMemoEncryptor(algorithm EncryptionAlgorithm) *MemoEncryptor {
	return &MemoEncryptor{
		algorithm: algorithm,
	}
}

// Encrypt encrypts a memo for a recipient
func (me *MemoEncryptor) Encrypt(
	memo []byte,
	recipientPublicKey []byte,
) (*EncryptedMemo, error) {
	if len(memo) == 0 {
		return nil, errors.New("memo cannot be empty")
	}

	if len(recipientPublicKey) == 0 {
		return nil, errors.New("recipient public key cannot be empty")
	}

	switch me.algorithm {
	case AlgorithmAES256GCM:
		return me.encryptAES256GCM(memo, recipientPublicKey)
	case AlgorithmChaCha20Poly1305:
		return me.encryptChaCha20Poly1305(memo, recipientPublicKey)
	case AlgorithmXChaCha20Poly1305:
		return me.encryptXChaCha20Poly1305(memo, recipientPublicKey)
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", me.algorithm)
	}
}

// Decrypt decrypts an encrypted memo
func (me *MemoEncryptor) Decrypt(
	encryptedMemo *EncryptedMemo,
	privateKey []byte,
) ([]byte, error) {
	if encryptedMemo == nil {
		return nil, errors.New("encrypted memo is nil")
	}

	if len(privateKey) == 0 {
		return nil, errors.New("private key cannot be empty")
	}

	switch encryptedMemo.Algorithm {
	case AlgorithmAES256GCM:
		return me.decryptAES256GCM(encryptedMemo, privateKey)
	case AlgorithmChaCha20Poly1305:
		return me.decryptChaCha20Poly1305(encryptedMemo, privateKey)
	case AlgorithmXChaCha20Poly1305:
		return me.decryptXChaCha20Poly1305(encryptedMemo, privateKey)
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", encryptedMemo.Algorithm)
	}
}

// encryptAES256GCM encrypts using AES-256-GCM
func (me *MemoEncryptor) encryptAES256GCM(
	memo []byte,
	recipientPublicKey []byte,
) (*EncryptedMemo, error) {
	// Generate ephemeral key pair
	curve := elliptic.P256()
	ephemeralPrivate, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ephemeral key: %w", err)
	}

	// Derive shared secret using ECDH
	sharedSecret := deriveSharedSecret(ephemeralPrivate, recipientPublicKey)

	// Derive encryption key
	encryptionKey := deriveEncryptionKey(sharedSecret, []byte("aes-256-gcm"))

	// Create AES cipher
	block, err := aes.NewCipher(encryptionKey[:32])
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

	// Extract authentication tag (last 16 bytes)
	authTag := ciphertext[len(ciphertext)-16:]
	ciphertext = ciphertext[:len(ciphertext)-16]

	ephemeralPublicKey := elliptic.MarshalCompressed(
		curve,
		ephemeralPrivate.PublicKey.X,
		ephemeralPrivate.PublicKey.Y,
	)

	return &EncryptedMemo{
		Ciphertext:         ciphertext,
		Nonce:              nonce,
		Algorithm:          AlgorithmAES256GCM,
		RecipientPublicKey: recipientPublicKey,
		EphemeralPublicKey: ephemeralPublicKey,
		AuthenticationTag:  authTag,
	}, nil
}

// decryptAES256GCM decrypts AES-256-GCM encrypted memo
func (me *MemoEncryptor) decryptAES256GCM(
	encryptedMemo *EncryptedMemo,
	privateKey []byte,
) ([]byte, error) {
	// Derive shared secret
	curve := elliptic.P256()
	privKey := new(ecdsa.PrivateKey)
	privKey.D = new(big.Int).SetBytes(privateKey)
	privKey.PublicKey.Curve = curve
	privKey.PublicKey.X, privKey.PublicKey.Y = curve.ScalarBaseMult(privateKey)

	sharedSecret := deriveSharedSecret(privKey, encryptedMemo.EphemeralPublicKey)

	// Derive decryption key
	decryptionKey := deriveEncryptionKey(sharedSecret, []byte("aes-256-gcm"))

	// Create AES cipher
	block, err := aes.NewCipher(decryptionKey[:32])
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Reconstruct ciphertext with auth tag
	fullCiphertext := append(encryptedMemo.Ciphertext, encryptedMemo.AuthenticationTag...)

	// Decrypt
	plaintext, err := gcm.Open(nil, encryptedMemo.Nonce, fullCiphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// encryptChaCha20Poly1305 encrypts using ChaCha20-Poly1305
func (me *MemoEncryptor) encryptChaCha20Poly1305(
	memo []byte,
	recipientPublicKey []byte,
) (*EncryptedMemo, error) {
	// Generate ephemeral key pair
	curve := elliptic.P256()
	ephemeralPrivate, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ephemeral key: %w", err)
	}

	// Derive shared secret
	sharedSecret := deriveSharedSecret(ephemeralPrivate, recipientPublicKey)

	// Derive encryption key
	encryptionKey := deriveEncryptionKey(sharedSecret, []byte("chacha20-poly1305"))

	// Create ChaCha20-Poly1305 cipher
	aead, err := chacha20poly1305.New(encryptionKey[:32])
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt
	ciphertext := aead.Seal(nil, nonce, memo, nil)

	// Extract authentication tag
	authTag := ciphertext[len(ciphertext)-16:]
	ciphertext = ciphertext[:len(ciphertext)-16]

	ephemeralPublicKey := elliptic.MarshalCompressed(
		curve,
		ephemeralPrivate.PublicKey.X,
		ephemeralPrivate.PublicKey.Y,
	)

	return &EncryptedMemo{
		Ciphertext:         ciphertext,
		Nonce:              nonce,
		Algorithm:          AlgorithmChaCha20Poly1305,
		RecipientPublicKey: recipientPublicKey,
		EphemeralPublicKey: ephemeralPublicKey,
		AuthenticationTag:  authTag,
	}, nil
}

// decryptChaCha20Poly1305 decrypts ChaCha20-Poly1305 encrypted memo
func (me *MemoEncryptor) decryptChaCha20Poly1305(
	encryptedMemo *EncryptedMemo,
	privateKey []byte,
) ([]byte, error) {
	// Derive shared secret
	curve := elliptic.P256()
	privKey := new(ecdsa.PrivateKey)
	privKey.D = new(big.Int).SetBytes(privateKey)
	privKey.PublicKey.Curve = curve
	privKey.PublicKey.X, privKey.PublicKey.Y = curve.ScalarBaseMult(privateKey)

	sharedSecret := deriveSharedSecret(privKey, encryptedMemo.EphemeralPublicKey)

	// Derive decryption key
	decryptionKey := deriveEncryptionKey(sharedSecret, []byte("chacha20-poly1305"))

	// Create ChaCha20-Poly1305 cipher
	aead, err := chacha20poly1305.New(decryptionKey[:32])
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Reconstruct ciphertext with auth tag
	fullCiphertext := append(encryptedMemo.Ciphertext, encryptedMemo.AuthenticationTag...)

	// Decrypt
	plaintext, err := aead.Open(nil, encryptedMemo.Nonce, fullCiphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// encryptXChaCha20Poly1305 encrypts using XChaCha20-Poly1305
func (me *MemoEncryptor) encryptXChaCha20Poly1305(
	memo []byte,
	recipientPublicKey []byte,
) (*EncryptedMemo, error) {
	// Generate ephemeral key pair
	curve := elliptic.P256()
	ephemeralPrivate, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ephemeral key: %w", err)
	}

	// Derive shared secret
	sharedSecret := deriveSharedSecret(ephemeralPrivate, recipientPublicKey)

	// Derive encryption key
	encryptionKey := deriveEncryptionKey(sharedSecret, []byte("xchacha20-poly1305"))

	// Create XChaCha20-Poly1305 cipher
	aead, err := chacha20poly1305.NewX(encryptionKey[:32])
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Generate nonce (24 bytes for XChaCha20)
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt
	ciphertext := aead.Seal(nil, nonce, memo, nil)

	// Extract authentication tag
	authTag := ciphertext[len(ciphertext)-16:]
	ciphertext = ciphertext[:len(ciphertext)-16]

	ephemeralPublicKey := elliptic.MarshalCompressed(
		curve,
		ephemeralPrivate.PublicKey.X,
		ephemeralPrivate.PublicKey.Y,
	)

	return &EncryptedMemo{
		Ciphertext:         ciphertext,
		Nonce:              nonce,
		Algorithm:          AlgorithmXChaCha20Poly1305,
		RecipientPublicKey: recipientPublicKey,
		EphemeralPublicKey: ephemeralPublicKey,
		AuthenticationTag:  authTag,
	}, nil
}

// decryptXChaCha20Poly1305 decrypts XChaCha20-Poly1305 encrypted memo
func (me *MemoEncryptor) decryptXChaCha20Poly1305(
	encryptedMemo *EncryptedMemo,
	privateKey []byte,
) ([]byte, error) {
	// Derive shared secret
	curve := elliptic.P256()
	privKey := new(ecdsa.PrivateKey)
	privKey.D = new(big.Int).SetBytes(privateKey)
	privKey.PublicKey.Curve = curve
	privKey.PublicKey.X, privKey.PublicKey.Y = curve.ScalarBaseMult(privateKey)

	sharedSecret := deriveSharedSecret(privKey, encryptedMemo.EphemeralPublicKey)

	// Derive decryption key
	decryptionKey := deriveEncryptionKey(sharedSecret, []byte("xchacha20-poly1305"))

	// Create XChaCha20-Poly1305 cipher
	aead, err := chacha20poly1305.NewX(decryptionKey[:32])
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Reconstruct ciphertext with auth tag
	fullCiphertext := append(encryptedMemo.Ciphertext, encryptedMemo.AuthenticationTag...)

	// Decrypt
	plaintext, err := aead.Open(nil, encryptedMemo.Nonce, fullCiphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// deriveSharedSecret derives a shared secret using ECDH
func deriveSharedSecret(privateKey *ecdsa.PrivateKey, publicKeyBytes []byte) []byte {
	curve := privateKey.PublicKey.Curve
	x, y := elliptic.UnmarshalCompressed(curve, publicKeyBytes)
	if x == nil {
		// Fallback to basic derivation
		hasher := sha256.New()
		hasher.Write(privateKey.D.Bytes())
		hasher.Write(publicKeyBytes)
		return hasher.Sum(nil)
	}

	// Compute shared point
	sharedX, _ := curve.ScalarMult(x, y, privateKey.D.Bytes())

	// Hash the shared point
	hasher := sha256.New()
	hasher.Write(sharedX.Bytes())
	return hasher.Sum(nil)
}

// deriveEncryptionKey derives an encryption key from shared secret
func deriveEncryptionKey(sharedSecret []byte, info []byte) []byte {
	// Use HKDF-like derivation
	hasher := sha3.New256()
	hasher.Write(sharedSecret)
	hasher.Write(info)
	return hasher.Sum(nil)
}

// ViewKeyType defines the type of view key
type ViewKeyType string

const (
	ViewKeyTypeIncoming ViewKeyType = "INCOMING"
	ViewKeyTypeOutgoing ViewKeyType = "OUTGOING"
	ViewKeyTypeAudit    ViewKeyType = "AUDIT"
	ViewKeyTypeFull     ViewKeyType = "FULL"
)

// ViewKey represents a key for selective disclosure
type ViewKey struct {
	Type        ViewKeyType
	PublicKey   []byte
	PrivateKey  []byte
	Address     []byte
	Permissions []string
	ExpiresAt   *time.Time
	CreatedAt   time.Time
}

// ViewKeyManager manages view keys for selective disclosure
type ViewKeyManager struct {
	viewKeys map[string]*ViewKey
	mu       sync.RWMutex
}

// NewViewKeyManager creates a new view key manager
func NewViewKeyManager() *ViewKeyManager {
	return &ViewKeyManager{
		viewKeys: make(map[string]*ViewKey),
	}
}

// GenerateViewKey generates a new view key
func (vkm *ViewKeyManager) GenerateViewKey(
	keyType ViewKeyType,
	address []byte,
	permissions []string,
	expiresAt *time.Time,
) (*ViewKey, error) {
	// Generate key pair
	curve := elliptic.P256()
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	publicKey := elliptic.MarshalCompressed(
		curve,
		privateKey.PublicKey.X,
		privateKey.PublicKey.Y,
	)

	viewKey := &ViewKey{
		Type:        keyType,
		PublicKey:   publicKey,
		PrivateKey:  privateKey.D.Bytes(),
		Address:     address,
		Permissions: permissions,
		ExpiresAt:   expiresAt,
		CreatedAt:   time.Now(),
	}

	vkm.mu.Lock()
	vkm.viewKeys[string(publicKey)] = viewKey
	vkm.mu.Unlock()

	return viewKey, nil
}

// GetViewKey retrieves a view key
func (vkm *ViewKeyManager) GetViewKey(publicKey []byte) (*ViewKey, error) {
	vkm.mu.RLock()
	defer vkm.mu.RUnlock()

	viewKey, exists := vkm.viewKeys[string(publicKey)]
	if !exists {
		return nil, errors.New("view key not found")
	}

	// Check expiration
	if viewKey.ExpiresAt != nil && time.Now().After(*viewKey.ExpiresAt) {
		return nil, errors.New("view key expired")
	}

	return viewKey, nil
}

// RevokeViewKey revokes a view key
func (vkm *ViewKeyManager) RevokeViewKey(publicKey []byte) error {
	vkm.mu.Lock()
	defer vkm.mu.Unlock()

	if _, exists := vkm.viewKeys[string(publicKey)]; !exists {
		return errors.New("view key not found")
	}

	delete(vkm.viewKeys, string(publicKey))
	return nil
}

// VerifyPermission verifies if a view key has a specific permission
func (vkm *ViewKeyManager) VerifyPermission(publicKey []byte, permission string) (bool, error) {
	viewKey, err := vkm.GetViewKey(publicKey)
	if err != nil {
		return false, err
	}

	for _, p := range viewKey.Permissions {
		if p == permission {
			return true, nil
		}
	}

	return false, nil
}

// DecryptWithViewKey decrypts data using a view key
func (vkm *ViewKeyManager) DecryptWithViewKey(
	encryptedData []byte,
	viewKeyPublic []byte,
) ([]byte, error) {
	viewKey, err := vkm.GetViewKey(viewKeyPublic)
	if err != nil {
		return nil, err
	}

	// Use view key's private key to decrypt
	encryptor := NewMemoEncryptor(AlgorithmChaCha20Poly1305)

	// Reconstruct EncryptedMemo from data
	// This is simplified - in production would have proper serialization
	if len(encryptedData) < 48 {
		return nil, errors.New("encrypted data too short")
	}

	encryptedMemo := &EncryptedMemo{
		Ciphertext: encryptedData,
		Nonce:      encryptedData[:12],
		Algorithm:  AlgorithmChaCha20Poly1305,
	}

	return encryptor.Decrypt(encryptedMemo, viewKey.PrivateKey)
}

// GrantViewAccess creates a view key with specific permissions
func (vkm *ViewKeyManager) GrantViewAccess(
	address []byte,
	permissions []string,
	duration time.Duration,
) (*ViewKey, error) {
	expiresAt := time.Now().Add(duration)

	return vkm.GenerateViewKey(
		ViewKeyTypeAudit,
		address,
		permissions,
		&expiresAt,
	)
}

// ListActiveViewKeys returns all active view keys for an address
func (vkm *ViewKeyManager) ListActiveViewKeys(address []byte) []*ViewKey {
	vkm.mu.RLock()
	defer vkm.mu.RUnlock()

	var active []*ViewKey
	now := time.Now()

	for _, vk := range vkm.viewKeys {
		if string(vk.Address) == string(address) {
			if vk.ExpiresAt == nil || now.Before(*vk.ExpiresAt) {
				active = append(active, vk)
			}
		}
	}

	return active
}
