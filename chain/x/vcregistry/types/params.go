package types

import (
	"fmt"

	vcregistrypb "github.com/aequitas/aura/proto/aura/vcregistry/v1beta1"
)

// Params defines the parameters for the vcregistry module
type Params struct {
	// Minting limits
	MaxVcsPerUser  uint64 `json:"max_vcs_per_user"`
	MaxMintPerDay  uint64 `json:"max_mint_per_day"`
	MaxMintPerHour uint64 `json:"max_mint_per_hour"`

	// Default settings
	DefaultVcExpiryDays uint64 `json:"default_vc_expiry_days"`

	// Revocation
	RevocationMerkleUpdateFrequency uint64 `json:"revocation_merkle_update_frequency"`

	// DID
	DidPrefix  string `json:"did_prefix"`
	DidNetwork string `json:"did_network"`

	// Fees
	MintFee               string `json:"mint_fee"`
	RevokeFee             string `json:"revoke_fee"`
	PolicyCreationDeposit string `json:"policy_creation_deposit"`

	// Rate limiting
	RateLimitingEnabled bool `json:"rate_limiting_enabled"`
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return Params{
		MaxVcsPerUser:                   50,
		MaxMintPerDay:                   5,
		MaxMintPerHour:                  2,
		DefaultVcExpiryDays:             365,
		RevocationMerkleUpdateFrequency: 100,
		DidPrefix:                       "did:aura",
		DidNetwork:                      "mainnet",
		MintFee:                         "1000000uaura",
		RevokeFee:                       "0uaura",
		PolicyCreationDeposit:           "10000000uaura",
		RateLimitingEnabled:             true,
	}
}

// DefaultParamsProto returns a default set of parameters in proto format
func DefaultParamsProto() *vcregistrypb.Params {
	defaults := DefaultParams()
	return ParamsToProto(defaults)
}

// ParamsFromProto converts proto Params to internal Params type
func ParamsFromProto(pb *vcregistrypb.Params) Params {
	if pb == nil {
		return Params{}
	}

	return Params{
		MaxVcsPerUser:                   pb.MaxVcsPerUser,
		MaxMintPerDay:                   pb.MaxMintPerDay,
		MaxMintPerHour:                  pb.MaxMintPerHour,
		DefaultVcExpiryDays:             pb.DefaultVcExpiryDays,
		RevocationMerkleUpdateFrequency: pb.RevocationMerkleUpdateFrequency,
		DidPrefix:                       pb.DidPrefix,
		DidNetwork:                      pb.DidNetwork,
		MintFee:                         pb.MintFee,
		RevokeFee:                       pb.RevokeFee,
		PolicyCreationDeposit:           pb.PolicyCreationDeposit,
		RateLimitingEnabled:             pb.RateLimitingEnabled,
	}
}

// ParamsToProto converts internal Params to proto Params type
func ParamsToProto(p Params) *vcregistrypb.Params {
	return &vcregistrypb.Params{
		MaxVcsPerUser:                   p.MaxVcsPerUser,
		MaxMintPerDay:                   p.MaxMintPerDay,
		MaxMintPerHour:                  p.MaxMintPerHour,
		DefaultVcExpiryDays:             p.DefaultVcExpiryDays,
		RevocationMerkleUpdateFrequency: p.RevocationMerkleUpdateFrequency,
		DidPrefix:                       p.DidPrefix,
		DidNetwork:                      p.DidNetwork,
		MintFee:                         p.MintFee,
		RevokeFee:                       p.RevokeFee,
		PolicyCreationDeposit:           p.PolicyCreationDeposit,
		RateLimitingEnabled:             p.RateLimitingEnabled,
	}
}

// Validate performs validation on the Params
func (p Params) Validate() error {
	if p.MaxVcsPerUser == 0 {
		return fmt.Errorf("max vcs per user must be positive")
	}

	if p.MaxMintPerDay == 0 {
		return fmt.Errorf("max mint per day must be positive")
	}

	if p.MaxMintPerHour == 0 {
		return fmt.Errorf("max mint per hour must be positive")
	}

	if p.MaxMintPerHour > p.MaxMintPerDay {
		return fmt.Errorf("max mint per hour cannot exceed max mint per day")
	}

	if p.DefaultVcExpiryDays == 0 {
		return fmt.Errorf("default vc expiry days must be positive")
	}

	if p.RevocationMerkleUpdateFrequency == 0 {
		return fmt.Errorf("revocation merkle update frequency must be positive")
	}

	if p.DidPrefix == "" {
		return fmt.Errorf("did prefix cannot be empty")
	}

	if p.DidNetwork == "" {
		return fmt.Errorf("did network cannot be empty")
	}

	if p.DidNetwork != "mainnet" && p.DidNetwork != "testnet" {
		return fmt.Errorf("did network must be either 'mainnet' or 'testnet', got %s", p.DidNetwork)
	}

	if p.MintFee == "" {
		return fmt.Errorf("mint fee cannot be empty")
	}

	if p.RevokeFee == "" {
		return fmt.Errorf("revoke fee cannot be empty")
	}

	if p.PolicyCreationDeposit == "" {
		return fmt.Errorf("policy creation deposit cannot be empty")
	}

	return nil
}
