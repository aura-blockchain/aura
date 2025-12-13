package testutil

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RandomString generates a random string of specified length
func RandomString(length int) string {
	bytes := make([]byte, length)
	fillRandom(bytes)
	return hex.EncodeToString(bytes)[:length]
}

// RandomBytes generates random bytes of specified length
func RandomBytes(length int) []byte {
	bytes := make([]byte, length)
	fillRandom(bytes)
	return bytes
}

// RandomInt generates a random integer between min and max
func RandomInt(min, max int64) int64 {
	bytes := make([]byte, 8)
	fillRandom(bytes)
	val := int64(0)
	for _, b := range bytes {
		val = val*256 + int64(b)
	}
	if val < 0 {
		val = -val
	}
	return min + (val % (max - min + 1))
}

// RandomAmount generates a random coin amount
func RandomAmount(denom string) sdk.Coin {
	amount := RandomInt(1000, 1000000)
	return sdk.NewCoin(denom, math.NewInt(amount))
}

// RandomCoins generates random coins
func RandomCoins() sdk.Coins {
	return sdk.NewCoins(
		RandomAmount("aura"),
		RandomAmount("stake"),
	)
}

// RandomAddress generates a random SDK address
func RandomAddress() sdk.AccAddress {
	return sdk.AccAddress(RandomBytes(20))
}

// RandomValAddress generates a random validator address
func RandomValAddress() sdk.ValAddress {
	return sdk.ValAddress(RandomBytes(20))
}

// RandomDID generates a random DID
func RandomDID() string {
	return fmt.Sprintf("did:aura:%s", RandomString(16))
}

// RandomIPFSHash generates a random IPFS hash
func RandomIPFSHash() string {
	// Generate a valid-looking IPFS CIDv0 hash
	return "Qm" + RandomString(44)
}

// RandomProof generates random proof bytes
func RandomProof() []byte {
	return RandomBytes(256)
}

func fillRandom(b []byte) {
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("crypto/rand failed: %v", err))
	}
}

// RandomTimestamp generates a random timestamp
func RandomTimestamp() time.Time {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	offset := RandomInt(0, 365*24*3600)
	return base.Add(time.Duration(offset) * time.Second)
}

// RandomBool generates a random boolean
func RandomBool() bool {
	return RandomInt(0, 1) == 1
}

// RandomDecimal generates a random decimal between 0 and 1
func RandomDecimal() math.LegacyDec {
	val := RandomInt(0, 1000000)
	return math.LegacyNewDec(val).QuoInt64(1000000)
}

// TestDataGenerator provides methods to generate test data
type TestDataGenerator struct {
	seed int64
}

// NewTestDataGenerator creates a new test data generator
func NewTestDataGenerator(seed int64) *TestDataGenerator {
	return &TestDataGenerator{seed: seed}
}

// GenerateVCMetadata generates test VC metadata
func (g *TestDataGenerator) GenerateVCMetadata() string {
	return fmt.Sprintf(`{
		"@context": ["https://www.w3.org/2018/credentials/v1"],
		"id": "%s",
		"type": ["VerifiableCredential"],
		"issuer": "%s",
		"issuanceDate": "%s",
		"credentialSubject": {
			"id": "%s",
			"name": "Test Subject"
		}
	}`, RandomDID(), RandomDID(), time.Now().Format(time.RFC3339), RandomDID())
}

// GenerateBridgeTransfer generates test bridge transfer data
func (g *TestDataGenerator) GenerateBridgeTransfer() map[string]interface{} {
	return map[string]interface{}{
		"source_chain":      "ethereum",
		"destination_chain": "aura",
		"token":             "ETH",
		"amount":            RandomInt(1, 1000),
		"sender":            RandomAddress().String(),
		"recipient":         RandomAddress().String(),
		"nonce":             RandomInt(1, 1000000),
	}
}

// GenerateInclusionRoutine generates test inclusion routine data
func (g *TestDataGenerator) GenerateInclusionRoutine() map[string]interface{} {
	return map[string]interface{}{
		"id":          RandomString(32),
		"name":        "Test IR",
		"description": "Test Inclusion Routine",
		"creator":     RandomAddress().String(),
		"fee":         RandomAmount("aura"),
		"active":      true,
	}
}

// GenerateDEXOrder generates test DEX order data
func (g *TestDataGenerator) GenerateDEXOrder() map[string]interface{} {
	return map[string]interface{}{
		"order_id":       RandomString(32),
		"trader":         RandomAddress().String(),
		"token_in":       "aura",
		"token_out":      "atom",
		"amount_in":      RandomInt(100, 10000),
		"min_amount_out": RandomInt(50, 5000),
		"timestamp":      time.Now().Unix(),
	}
}

// GenerateComplianceRecord generates test compliance record
func (g *TestDataGenerator) GenerateComplianceRecord() map[string]interface{} {
	return map[string]interface{}{
		"subject":    RandomAddress().String(),
		"status":     "verified",
		"verifier":   RandomAddress().String(),
		"timestamp":  time.Now().Unix(),
		"expiration": time.Now().Add(365 * 24 * time.Hour).Unix(),
	}
}
