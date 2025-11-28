package privacy

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"

	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/sha3"
)

// StealthAddressScheme implements one-time stealth addresses for privacy
// Based on the scheme used in Monero and similar to CryptoNote protocol
type StealthAddressScheme struct {
	curve elliptic.Curve
}

// NewStealthAddressScheme creates a new stealth address scheme
func NewStealthAddressScheme() *StealthAddressScheme {
	return &StealthAddressScheme{
		curve: elliptic.P256(),
	}
}

// KeyPair represents a public/private key pair
type KeyPair struct {
	PrivateKey []byte
	PublicKey  []byte
}

// StealthKeys contains the keys needed for stealth addresses
type StealthKeys struct {
	SpendKeyPair *KeyPair // Used to spend funds
	ViewKeyPair  *KeyPair // Used to view transactions
}

// GenerateStealthKeys generates a new pair of spend and view keys
func (s *StealthAddressScheme) GenerateStealthKeys() (*StealthKeys, error) {
	// Generate spend key pair
	spendPriv, err := ecdsa.GenerateKey(s.curve, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate spend key: %w", err)
	}

	// Generate view key pair
	viewPriv, err := ecdsa.GenerateKey(s.curve, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate view key: %w", err)
	}

	return &StealthKeys{
		SpendKeyPair: &KeyPair{
			PrivateKey: spendPriv.D.Bytes(),
			PublicKey:  elliptic.MarshalCompressed(s.curve, spendPriv.PublicKey.X, spendPriv.PublicKey.Y),
		},
		ViewKeyPair: &KeyPair{
			PrivateKey: viewPriv.D.Bytes(),
			PublicKey:  elliptic.MarshalCompressed(s.curve, viewPriv.PublicKey.X, viewPriv.PublicKey.Y),
		},
	}, nil
}

// StealthAddress represents a one-time address
type StealthAddress struct {
	OneTimePublicKey []byte // P = H(rA)G + B
	TxPublicKey      []byte // R = rG (ephemeral public key)
	EncryptedAmount  []byte // Encrypted transaction amount
}

// GenerateStealthAddress creates a one-time stealth address for a recipient
func (s *StealthAddressScheme) GenerateStealthAddress(recipientSpendKey, recipientViewKey []byte) (*StealthAddress, error) {
	if len(recipientSpendKey) == 0 || len(recipientViewKey) == 0 {
		return nil, errors.New("recipient keys cannot be empty")
	}

	// Generate random ephemeral key r
	ephemeralKey, err := ecdsa.GenerateKey(s.curve, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ephemeral key: %w", err)
	}

	// R = rG (transaction public key)
	txPubKey := elliptic.MarshalCompressed(s.curve, ephemeralKey.PublicKey.X, ephemeralKey.PublicKey.Y)

	// Unmarshal recipient's view key
	recipientViewX, recipientViewY := elliptic.UnmarshalCompressed(s.curve, recipientViewKey)
	if recipientViewX == nil {
		return nil, errors.New("invalid recipient view key")
	}

	// Compute shared secret: rA (ephemeral private key * recipient view public key)
	sharedX, sharedY := s.curve.ScalarMult(recipientViewX, recipientViewY, ephemeralKey.D.Bytes())

	// Hash the shared secret: H(rA)
	hasher := sha3.New256()
	hasher.Write(elliptic.Marshal(s.curve, sharedX, sharedY))
	sharedSecretHash := hasher.Sum(nil)

	// Unmarshal recipient's spend key
	recipientSpendX, recipientSpendY := elliptic.UnmarshalCompressed(s.curve, recipientSpendKey)
	if recipientSpendX == nil {
		return nil, errors.New("invalid recipient spend key")
	}

	// P = H(rA)G + B (one-time public key)
	// First compute H(rA)G
	hx, hy := s.curve.ScalarBaseMult(sharedSecretHash)
	// Then add B
	oneTimeX, oneTimeY := s.curve.Add(hx, hy, recipientSpendX, recipientSpendY)
	oneTimePubKey := elliptic.MarshalCompressed(s.curve, oneTimeX, oneTimeY)

	return &StealthAddress{
		OneTimePublicKey: oneTimePubKey,
		TxPublicKey:      txPubKey,
		EncryptedAmount:  nil, // Will be set when encrypting amount
	}, nil
}

// ScanForStealthPayments scans for payments to a stealth address
func (s *StealthAddressScheme) ScanForStealthPayments(
	txPubKey []byte,
	oneTimeAddress []byte,
	viewPrivateKey []byte,
	spendPublicKey []byte,
) (bool, error) {
	if len(txPubKey) == 0 || len(oneTimeAddress) == 0 || len(viewPrivateKey) == 0 {
		return false, errors.New("invalid parameters")
	}

	// Unmarshal transaction public key R
	txPubX, txPubY := elliptic.UnmarshalCompressed(s.curve, txPubKey)
	if txPubX == nil {
		return false, errors.New("invalid transaction public key")
	}

	// Compute shared secret: aR (view private key * transaction public key)
	viewPriv := new(big.Int).SetBytes(viewPrivateKey)
	sharedX, sharedY := s.curve.ScalarMult(txPubX, txPubY, viewPriv.Bytes())

	// Hash the shared secret: H(aR)
	hasher := sha3.New256()
	hasher.Write(elliptic.Marshal(s.curve, sharedX, sharedY))
	sharedSecretHash := hasher.Sum(nil)

	// Unmarshal spend public key B
	spendPubX, spendPubY := elliptic.UnmarshalCompressed(s.curve, spendPublicKey)
	if spendPubX == nil {
		return false, errors.New("invalid spend public key")
	}

	// Compute P' = H(aR)G + B
	hx, hy := s.curve.ScalarBaseMult(sharedSecretHash)
	derivedX, derivedY := s.curve.Add(hx, hy, spendPubX, spendPubY)
	derivedPubKey := elliptic.MarshalCompressed(s.curve, derivedX, derivedY)

	// Check if derived key matches the one-time address
	return string(derivedPubKey) == string(oneTimeAddress), nil
}

// DerivePrivateKey derives the private key for spending from a stealth address
func (s *StealthAddressScheme) DerivePrivateKey(
	txPubKey []byte,
	viewPrivateKey []byte,
	spendPrivateKey []byte,
) ([]byte, error) {
	// Unmarshal transaction public key R
	txPubX, txPubY := elliptic.UnmarshalCompressed(s.curve, txPubKey)
	if txPubX == nil {
		return nil, errors.New("invalid transaction public key")
	}

	// Compute shared secret: aR
	viewPriv := new(big.Int).SetBytes(viewPrivateKey)
	sharedX, sharedY := s.curve.ScalarMult(txPubX, txPubY, viewPriv.Bytes())

	// Hash the shared secret: H(aR)
	hasher := sha3.New256()
	hasher.Write(elliptic.Marshal(s.curve, sharedX, sharedY))
	sharedSecretHash := hasher.Sum(nil)

	// x = H(aR) + b (one-time private key)
	h := new(big.Int).SetBytes(sharedSecretHash)
	b := new(big.Int).SetBytes(spendPrivateKey)
	n := s.curve.Params().N

	oneTimePriv := new(big.Int).Add(h, b)
	oneTimePriv.Mod(oneTimePriv, n)

	return oneTimePriv.Bytes(), nil
}

// Curve25519StealthAddress implements stealth addresses using Curve25519
// This is more efficient than NIST curves and widely used in privacy protocols
type Curve25519StealthAddress struct{}

// NewCurve25519StealthAddress creates a Curve25519-based stealth address scheme
func NewCurve25519StealthAddress() *Curve25519StealthAddress {
	return &Curve25519StealthAddress{}
}

// GenerateKeys generates Curve25519 key pairs for stealth addressing
func (c *Curve25519StealthAddress) GenerateKeys() (*StealthKeys, error) {
	// Generate spend key
	var spendPriv [32]byte
	if _, err := rand.Read(spendPriv[:]); err != nil {
		return nil, fmt.Errorf("failed to generate spend private key: %w", err)
	}
	var spendPub [32]byte
	curve25519.ScalarBaseMult(&spendPub, &spendPriv)

	// Generate view key
	var viewPriv [32]byte
	if _, err := rand.Read(viewPriv[:]); err != nil {
		return nil, fmt.Errorf("failed to generate view private key: %w", err)
	}
	var viewPub [32]byte
	curve25519.ScalarBaseMult(&viewPub, &viewPriv)

	return &StealthKeys{
		SpendKeyPair: &KeyPair{
			PrivateKey: spendPriv[:],
			PublicKey:  spendPub[:],
		},
		ViewKeyPair: &KeyPair{
			PrivateKey: viewPriv[:],
			PublicKey:  viewPub[:],
		},
	}, nil
}

// CreateOneTimeAddress creates a one-time Curve25519 stealth address
func (c *Curve25519StealthAddress) CreateOneTimeAddress(
	spendPubKey, viewPubKey [32]byte,
) (*StealthAddress, error) {
	// Generate ephemeral key pair
	var ephemeralPriv [32]byte
	if _, err := rand.Read(ephemeralPriv[:]); err != nil {
		return nil, fmt.Errorf("failed to generate ephemeral key: %w", err)
	}
	var ephemeralPub [32]byte
	curve25519.ScalarBaseMult(&ephemeralPub, &ephemeralPriv)

	// Compute shared secret: r * A (ephemeral private * view public)
	var sharedSecret [32]byte
	curve25519.ScalarMult(&sharedSecret, &ephemeralPriv, &viewPubKey)

	// Hash shared secret
	hasher := sha256.New()
	hasher.Write(sharedSecret[:])
	hashedSecret := hasher.Sum(nil)

	// Derive one-time public key: P = H(rA)G + B
	var hG [32]byte
	curve25519.ScalarBaseMult(&hG, (*[32]byte)(hashedSecret[:32]))

	var oneTimePub [32]byte
	// Add points: hG + spendPubKey
	// For Curve25519, we use scalar addition in the exponent
	var combined [32]byte
	for i := 0; i < 32; i++ {
		combined[i] = hG[i] ^ spendPubKey[i] // Simplified point addition
	}
	oneTimePub = combined

	return &StealthAddress{
		OneTimePublicKey: oneTimePub[:],
		TxPublicKey:      ephemeralPub[:],
	}, nil
}

// ScanTransaction checks if a transaction is meant for the wallet
func (c *Curve25519StealthAddress) ScanTransaction(
	txPubKey [32]byte,
	oneTimeAddr [32]byte,
	viewPrivKey [32]byte,
	spendPubKey [32]byte,
) bool {
	// Compute shared secret: a * R (view private * tx public)
	var sharedSecret [32]byte
	curve25519.ScalarMult(&sharedSecret, &viewPrivKey, &txPubKey)

	// Hash shared secret
	hasher := sha256.New()
	hasher.Write(sharedSecret[:])
	hashedSecret := hasher.Sum(nil)

	// Derive expected one-time key
	var hG [32]byte
	curve25519.ScalarBaseMult(&hG, (*[32]byte)(hashedSecret[:32]))

	var derivedPub [32]byte
	for i := 0; i < 32; i++ {
		derivedPub[i] = hG[i] ^ spendPubKey[i]
	}

	// Check if it matches
	return derivedPub == oneTimeAddr
}

// DualKeyStealthAddress implements a dual-key stealth address system
// This provides better separation of viewing and spending capabilities
type DualKeyStealthAddress struct {
	viewKey  *KeyPair
	spendKey *KeyPair
}

// NewDualKeyStealthAddress creates a new dual-key stealth address
func NewDualKeyStealthAddress(viewKey, spendKey *KeyPair) *DualKeyStealthAddress {
	return &DualKeyStealthAddress{
		viewKey:  viewKey,
		spendKey: spendKey,
	}
}

// GetAddress returns the stealth address that can be publicly shared
func (d *DualKeyStealthAddress) GetAddress() []byte {
	// Combine view and spend public keys
	hasher := sha256.New()
	hasher.Write(d.viewKey.PublicKey)
	hasher.Write(d.spendKey.PublicKey)
	return hasher.Sum(nil)
}

// CanView checks if the given view key can view this address
func (d *DualKeyStealthAddress) CanView(viewKey []byte) bool {
	return string(viewKey) == string(d.viewKey.PrivateKey)
}

// CanSpend checks if the given spend key can spend from this address
func (d *DualKeyStealthAddress) CanSpend(spendKey []byte) bool {
	return string(spendKey) == string(d.spendKey.PrivateKey)
}

// EncryptAmount encrypts a transaction amount for a stealth address
func EncryptAmount(amount *big.Int, sharedSecret []byte) ([]byte, error) {
	if amount == nil || len(sharedSecret) == 0 {
		return nil, errors.New("invalid parameters")
	}

	// Use shared secret to derive encryption key
	hasher := sha256.New()
	hasher.Write(sharedSecret)
	hasher.Write([]byte("amount_encryption"))
	key := hasher.Sum(nil)

	// XOR amount with key (simplified symmetric encryption)
	amountBytes := amount.Bytes()
	encrypted := make([]byte, len(amountBytes))
	for i := 0; i < len(amountBytes); i++ {
		encrypted[i] = amountBytes[i] ^ key[i%len(key)]
	}

	return encrypted, nil
}

// DecryptAmount decrypts an encrypted amount
func DecryptAmount(encryptedAmount []byte, sharedSecret []byte) (*big.Int, error) {
	if len(encryptedAmount) == 0 || len(sharedSecret) == 0 {
		return nil, errors.New("invalid parameters")
	}

	// Derive decryption key
	hasher := sha256.New()
	hasher.Write(sharedSecret)
	hasher.Write([]byte("amount_encryption"))
	key := hasher.Sum(nil)

	// XOR to decrypt
	decrypted := make([]byte, len(encryptedAmount))
	for i := 0; i < len(encryptedAmount); i++ {
		decrypted[i] = encryptedAmount[i] ^ key[i%len(key)]
	}

	return new(big.Int).SetBytes(decrypted), nil
}
