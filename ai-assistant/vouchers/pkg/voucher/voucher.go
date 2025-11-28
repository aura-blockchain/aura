package voucher

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/scrypt"
)

const (
	defaultDirName = ".aura/assistant-vouchers"
	keyFileName    = "sponsor_key.json"
	issuedFile     = "issued.json"
	redeemedFile   = "redeemed.json"
)

// Voucher describes the preimage signed by sponsors when issuing credit.
type Voucher struct {
	ID        string `json:"id"`
	Amount    string `json:"amount"`
	Denom     string `json:"denom"`
	Sponsor   string `json:"sponsor"`
	ExpiresAt int64  `json:"expires_at"`
	Nonce     string `json:"nonce"`
	Notes     string `json:"notes,omitempty"`
}

// SignedVoucher is the transferable payload (base64 encoded when shared).
type SignedVoucher struct {
	Voucher   Voucher `json:"voucher"`
	Signature string  `json:"signature"` // base64
	PublicKey string  `json:"public_key"`
	IssuedAt  int64   `json:"issued_at"`
}

// RedeemedRecord tracks local redemption events for auditing.
type RedeemedRecord struct {
	VoucherID string `json:"voucher_id"`
	Assistant string `json:"assistant"`
	Timestamp int64  `json:"timestamp"`
	TxHash    string `json:"tx_hash,omitempty"`
	Notes     string `json:"notes,omitempty"`
}

// IssueOptions configures voucher creation.
type IssueOptions struct {
	Amount      string
	Denom       string
	Sponsor     string
	ExpiresAt   time.Time
	Notes       string
	Passphrase  string
	KeyPath     string
	DataDir     string
	Nonce       string
	OutputHuman bool
}

// RedeemOptions configures voucher redemption.
type RedeemOptions struct {
	EncodedVoucher string
	Assistant      string
	Passphrase     string
	ExpectPubKey   string
	KeyPath        string
	DataDir        string
	TxHash         string
	Notes          string
}

// Manager encapsulates filesystem paths.
type Manager struct {
	KeyPath string
	DataDir string
	Profile string
}

// NewManager builds a manager rooted in the user's config directory.
func NewManager(keyPath, dataDir, profile string) (*Manager, error) {
	resolvedKey, resolvedDir, resolvedProfile, err := resolvePaths(keyPath, dataDir, profile)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(resolvedDir, 0o700); err != nil {
		return nil, fmt.Errorf("ensure data dir: %w", err)
	}
	return &Manager{KeyPath: resolvedKey, DataDir: resolvedDir, Profile: resolvedProfile}, nil
}

func defaultDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, defaultDirName), nil
}

func resolvePaths(keyPath, dataDir, profile string) (string, string, string, error) {
	if profile == "" {
		profile = "default"
	}
	if keyPath != "" && dataDir != "" {
		return keyPath, dataDir, profile, nil
	}
	base, err := defaultDir()
	if err != nil {
		return "", "", "", err
	}
	profileDir := filepath.Join(base, profile)
	if dataDir == "" {
		dataDir = profileDir
	}
	if keyPath == "" {
		keyPath = filepath.Join(profileDir, keyFileName)
	}
	return keyPath, dataDir, profile, nil
}

// Issue creates, signs, and records a voucher.
func (m *Manager) Issue(opts IssueOptions) (SignedVoucher, error) {
	priv, pub, err := m.loadPrivateKey(opts.Passphrase)
	if err != nil {
		return SignedVoucher{}, err
	}
	nonce := opts.Nonce
	if nonce == "" {
		nonce = uuid.NewString()
	}
	if opts.Amount == "" {
		return SignedVoucher{}, fmt.Errorf("amount required")
	}
	if opts.Denom == "" {
		opts.Denom = "uaura"
	}
	if opts.Sponsor == "" {
		return SignedVoucher{}, fmt.Errorf("sponsor address required")
	}
	expiry := opts.ExpiresAt
	if expiry.IsZero() {
		expiry = time.Now().Add(30 * 24 * time.Hour)
	}

	v := Voucher{
		ID:        uuid.NewString(),
		Amount:    opts.Amount,
		Denom:     strings.ToLower(opts.Denom),
		Sponsor:   opts.Sponsor,
		ExpiresAt: expiry.Unix(),
		Nonce:     nonce,
		Notes:     opts.Notes,
	}
	payload, err := canonicalJSON(v)
	if err != nil {
		return SignedVoucher{}, err
	}
	sig := ed25519.Sign(priv, payload)
	signed := SignedVoucher{
		Voucher:   v,
		Signature: base64.StdEncoding.EncodeToString(sig),
		PublicKey: base64.StdEncoding.EncodeToString(pub),
		IssuedAt:  time.Now().Unix(),
	}
	if err := m.recordIssued(signed); err != nil {
		return SignedVoucher{}, err
	}
	return signed, nil
}

// Redeem verifies and records a voucher redemption.
func (m *Manager) Redeem(opts RedeemOptions) (SignedVoucher, error) {
	signed, err := DecodeVoucher(opts.EncodedVoucher)
	if err != nil {
		return SignedVoucher{}, err
	}
	if opts.ExpectPubKey != "" && !strings.EqualFold(opts.ExpectPubKey, signed.PublicKey) {
		return SignedVoucher{}, fmt.Errorf("voucher signed by unexpected key")
	}
	pubBytes, err := base64.StdEncoding.DecodeString(signed.PublicKey)
	if err != nil {
		return SignedVoucher{}, fmt.Errorf("invalid public key encoding: %w", err)
	}
	if time.Now().Unix() > signed.Voucher.ExpiresAt {
		return SignedVoucher{}, fmt.Errorf("voucher %s expired", signed.Voucher.ID)
	}
	payload, err := canonicalJSON(signed.Voucher)
	if err != nil {
		return SignedVoucher{}, err
	}
	sig, err := base64.StdEncoding.DecodeString(signed.Signature)
	if err != nil {
		return SignedVoucher{}, fmt.Errorf("invalid signature encoding: %w", err)
	}
	if !ed25519.Verify(ed25519.PublicKey(pubBytes), payload, sig) {
		return SignedVoucher{}, errors.New("voucher signature invalid")
	}

	record := RedeemedRecord{
		VoucherID: signed.Voucher.ID,
		Assistant: opts.Assistant,
		Timestamp: time.Now().Unix(),
		TxHash:    opts.TxHash,
		Notes:     opts.Notes,
	}
	if err := m.recordRedeemed(record); err != nil {
		return SignedVoucher{}, err
	}
	return signed, nil
}

// EncodeVoucher serializes and base64-encodes a signed voucher.
func EncodeVoucher(s SignedVoucher) (string, error) {
	buf, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf), nil
}

// DecodeVoucher decodes a base64 or JSON voucher payload.
func DecodeVoucher(raw string) (SignedVoucher, error) {
	raw = strings.TrimSpace(raw)
	var data []byte
	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		data = []byte(raw)
	} else {
		data = decoded
	}
	var signed SignedVoucher
	if err := json.Unmarshal(data, &signed); err != nil {
		return SignedVoucher{}, fmt.Errorf("decode voucher payload: %w", err)
	}
	return signed, nil
}

// GenerateKey stores a new sponsor keypair.
func (m *Manager) GenerateKey(passphrase string, overwrite bool) (ed25519.PublicKey, error) {
	if _, err := os.Stat(m.KeyPath); err == nil && !overwrite {
		return nil, fmt.Errorf("key already exists at %s", m.KeyPath)
	}
	if err := os.MkdirAll(filepath.Dir(m.KeyPath), 0o700); err != nil {
		return nil, err
	}
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	if err := saveKey(m.KeyPath, priv, passphrase); err != nil {
		return nil, err
	}
	return pub, nil
}

func (m *Manager) loadPrivateKey(passphrase string) (ed25519.PrivateKey, ed25519.PublicKey, error) {
	content, err := os.ReadFile(m.KeyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read key: %w", err)
	}
	var payload keyFile
	if err := json.Unmarshal(content, &payload); err != nil {
		return nil, nil, fmt.Errorf("parse key file: %w", err)
	}
	var privBytes []byte
	if payload.CipherText != "" {
		if passphrase == "" {
			return nil, nil, errors.New("passphrase required for encrypted key")
		}
		privBytes, err = decryptPayload(payload, passphrase)
		if err != nil {
			return nil, nil, err
		}
	} else {
		privBytes, err = base64.StdEncoding.DecodeString(payload.PrivateKey)
		if err != nil {
			return nil, nil, fmt.Errorf("decode private key: %w", err)
		}
	}
	if len(privBytes) != ed25519.SeedSize && len(privBytes) != ed25519.PrivateKeySize {
		return nil, nil, errors.New("invalid ed25519 key length")
	}
	var priv ed25519.PrivateKey
	if len(privBytes) == ed25519.SeedSize {
		priv = ed25519.NewKeyFromSeed(privBytes)
	} else {
		priv = ed25519.PrivateKey(privBytes)
	}
	pubBytes, err := base64.StdEncoding.DecodeString(payload.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("decode public key: %w", err)
	}
	return priv, ed25519.PublicKey(pubBytes), nil
}

type keyFile struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key,omitempty"`
	CipherText string `json:"cipher_text,omitempty"`
	Salt       string `json:"salt,omitempty"`
	Nonce      string `json:"nonce,omitempty"`
	KDF        string `json:"kdf,omitempty"`
	CreatedAt  string `json:"created_at"`
}

func saveKey(path string, priv ed25519.PrivateKey, passphrase string) error {
	payload := keyFile{
		PublicKey: base64.StdEncoding.EncodeToString(priv.Public().(ed25519.PublicKey)),
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	if passphrase == "" {
		payload.PrivateKey = base64.StdEncoding.EncodeToString(priv)
	} else {
		secret, err := encryptPayload(priv, passphrase)
		if err != nil {
			return err
		}
		payload.CipherText = secret.cipher
		payload.Salt = secret.salt
		payload.Nonce = secret.nonce
		payload.KDF = "scrypt"
	}
	buf, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, buf, 0o600)
}

type encryptedBlob struct {
	cipher string
	salt   string
	nonce  string
}

func encryptPayload(data []byte, passphrase string) (*encryptedBlob, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	key, err := scrypt.Key([]byte(passphrase), salt, 1<<15, 8, 1, 32)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	cipherText := gcm.Seal(nil, nonce, data, nil)
	return &encryptedBlob{
		cipher: base64.StdEncoding.EncodeToString(cipherText),
		salt:   base64.StdEncoding.EncodeToString(salt),
		nonce:  base64.StdEncoding.EncodeToString(nonce),
	}, nil
}

func decryptPayload(payload keyFile, passphrase string) ([]byte, error) {
	if payload.Salt == "" || payload.Nonce == "" {
		return nil, errors.New("encrypted key missing salt or nonce")
	}
	salt, err := base64.StdEncoding.DecodeString(payload.Salt)
	if err != nil {
		return nil, err
	}
	nonce, err := base64.StdEncoding.DecodeString(payload.Nonce)
	if err != nil {
		return nil, err
	}
	key, err := scrypt.Key([]byte(passphrase), salt, 1<<15, 8, 1, 32)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	cipherBytes, err := base64.StdEncoding.DecodeString(payload.CipherText)
	if err != nil {
		return nil, err
	}
	plain, err := gcm.Open(nil, nonce, cipherBytes, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt private key: %w", err)
	}
	return plain, nil
}

func canonicalJSON(v Voucher) ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimSpace(buf.Bytes()), nil
}

func (m *Manager) recordIssued(v SignedVoucher) error {
	path := filepath.Join(m.DataDir, issuedFile)
	return appendJSON(path, v)
}

func (m *Manager) recordRedeemed(r RedeemedRecord) error {
	path := filepath.Join(m.DataDir, redeemedFile)
	return appendJSON(path, r)
}

func appendJSON(path string, value interface{}) error {
	var existing []json.RawMessage
	data, err := os.ReadFile(path)
	if err == nil && len(data) > 0 {
		if err := json.Unmarshal(data, &existing); err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
	}
	entry, err := json.Marshal(value)
	if err != nil {
		return err
	}
	existing = append(existing, entry)
	buf, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, buf, 0o600)
}

// KeyPathDefault returns the resolved key path for a profile.
func KeyPathDefault(profile string) (string, error) {
	keyPath, _, _, err := resolvePaths("", "", profile)
	return keyPath, err
}

// DataDirDefault returns the resolved data directory for a profile.
func DataDirDefault(profile string) (string, error) {
	_, dataDir, _, err := resolvePaths("", "", profile)
	return dataDir, err
}
