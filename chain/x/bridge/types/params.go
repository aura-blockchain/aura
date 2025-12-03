package types

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Default values for security parameters
var (
	DefaultTimelockDuration  = 24 * time.Hour     // 24 hour timelock for withdrawals
	DefaultFraudProofWindow  = 7 * 24 * time.Hour // 7 day window for fraud proofs
)

// Security constants for validator confirmation requirements
const (
	DefaultMinConfirmations = uint64(3) // Default: require 3 validators for security
	MinAllowedConfirmations = uint64(2) // Absolute minimum: never allow less than 2 validators
)

// Parameter store keys
var (
	KeyBridgeEnabled                = []byte("BridgeEnabled")
	KeyMinConfirmations             = []byte("MinConfirmations")
	KeyBridgeFeeBasisPoints         = []byte("BridgeFeeBasisPoints")
	KeyCoreMaxTransferAmount        = []byte("MaxTransferAmount")
	KeyValidatorThresholdPercentage = []byte("ValidatorThresholdPercentage")
	KeySupplyCaps                   = []byte("SupplyCaps")
	KeyDailyMintLimit               = []byte("DailyMintLimit")
	KeyHourlyMintLimit              = []byte("HourlyMintLimit")
	KeyPaused                       = []byte("Paused")
	KeyPausedChains                 = []byte("PausedChains")
	KeyAutoPauseEnabled             = []byte("AutoPauseEnabled")
	KeyAutoPauseThreshold           = []byte("AutoPauseThreshold")
	KeyEmergencyPauseAddresses      = []byte("EmergencyPauseAddresses")
	KeyFraudProofWindow             = []byte("FraudProofWindow")
	KeySlashFraudSignature          = []byte("SlashFraudSignature")
	KeySlashDoubleSigning           = []byte("SlashDoubleSigning")
	KeySlashOffline                 = []byte("SlashOffline")
	KeyMinSigningWindow             = []byte("MinSigningWindow")
	KeyMinSignedPerWindow           = []byte("MinSignedPerWindow")
)

// Params defines the parameters persisted in the Cosmos SDK param store.
type Params struct {
	BridgeEnabled                bool              `json:"bridge_enabled"`
	MinConfirmations             uint64            `json:"min_confirmations"`
	BridgeFeeBasisPoints         uint64            `json:"bridge_fee_basis_points"`
	MaxTransferAmount            string            `json:"max_transfer_amount"`
	ValidatorThresholdPercentage uint64            `json:"validator_threshold_percentage"`
	SupplyCaps                   map[string]string `json:"supply_caps"`       // Per-token supply caps (denom -> amount)
	DailyMintLimit               string            `json:"daily_mint_limit"`  // Global daily mint limit
	HourlyMintLimit              string            `json:"hourly_mint_limit"` // Global hourly mint limit

	// Circuit breaker / emergency pause parameters
	Paused                  bool     `json:"paused"`                      // Global pause flag
	PausedChains            []string `json:"paused_chains"`               // Per-chain pause list
	AutoPauseEnabled        bool     `json:"auto_pause_enabled"`          // Enable automatic pause on anomaly
	AutoPauseThreshold      string   `json:"auto_pause_threshold"`        // Max hourly mint amount before auto-pause
	EmergencyPauseAddresses []string `json:"emergency_pause_addresses"`   // Authorized addresses for emergency pause

	// Fraud proof window - time period for challenging transfers before finalization
	// SECURITY CRITICAL: Provides time window for fraud proof submission before tokens are released
	FraudProofWindow int64 `json:"fraud_proof_window"` // Duration in seconds (e.g., 3600 = 1 hour)

	// Validator slashing parameters
	// SECURITY CRITICAL: Economic deterrent against malicious validator behavior
	SlashFraudSignature string `json:"slash_fraud_signature"` // Slash fraction for signing fraudulent transfers (e.g., "0.50" = 50%)
	SlashDoubleSigning  string `json:"slash_double_signing"`  // Slash fraction for double-signing (e.g., "1.00" = 100%)
	SlashOffline        string `json:"slash_offline"`         // Slash fraction for being offline (e.g., "0.01" = 1%)

	// Liveness tracking parameters for offline slashing
	MinSigningWindow   int64  `json:"min_signing_window"`    // Number of blocks to track liveness (e.g., 10000)
	MinSignedPerWindow string `json:"min_signed_per_window"` // Minimum percentage of signatures required (e.g., "0.50" = 50%)
}

// DefaultParams returns default parameters used by the param store.
func DefaultParams() Params {
	return Params{
		BridgeEnabled:                true,
		MinConfirmations:             DefaultMinConfirmations, // SECURITY: Require 3 validators minimum
		BridgeFeeBasisPoints:         30,
		MaxTransferAmount:            "1000000000",
		ValidatorThresholdPercentage: 67,
		SupplyCaps:                   make(map[string]string), // Empty by default, set per token
		DailyMintLimit:               "10000000000",           // 10 billion per day default
		HourlyMintLimit:              "1000000000",            // 1 billion per hour default

		// Circuit breaker defaults
		Paused:                  false,
		PausedChains:            []string{},                               // No chains paused by default
		AutoPauseEnabled:        false,                                    // Disabled by default, enable after testing
		AutoPauseThreshold:      "5000000000",                             // 5 billion per hour triggers auto-pause
		EmergencyPauseAddresses: []string{},                               // Must be set by governance
		FraudProofWindow:        int64(DefaultFraudProofWindow.Seconds()), // 7 days in seconds (604800)

		// Validator slashing defaults
		// SECURITY: Severe punishments deter malicious behavior
		SlashFraudSignature: "0.50",  // 50% of stake slashed for signing fraudulent transfers
		SlashDoubleSigning:  "1.00",  // 100% of stake slashed (tombstoned) for double-signing
		SlashOffline:        "0.01",  // 1% of stake slashed for being offline
		MinSigningWindow:    10000,   // Track liveness over 10,000 blocks (~18 hours at 6s blocks)
		MinSignedPerWindow:  "0.50",  // Must sign at least 50% of blocks in window
	}
}

// ParamKeyTable returns the parameter key table
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyBridgeEnabled, &p.BridgeEnabled, validateBool),
		paramtypes.NewParamSetPair(KeyMinConfirmations, &p.MinConfirmations, validateMinConfirmations),
		paramtypes.NewParamSetPair(KeyBridgeFeeBasisPoints, &p.BridgeFeeBasisPoints, validateUint64Core),
		paramtypes.NewParamSetPair(KeyCoreMaxTransferAmount, &p.MaxTransferAmount, validateStringNotEmpty),
		paramtypes.NewParamSetPair(KeyValidatorThresholdPercentage, &p.ValidatorThresholdPercentage, validateUint64Core),
		paramtypes.NewParamSetPair(KeySupplyCaps, &p.SupplyCaps, validateSupplyCaps),
		paramtypes.NewParamSetPair(KeyDailyMintLimit, &p.DailyMintLimit, validateStringNotEmpty),
		paramtypes.NewParamSetPair(KeyHourlyMintLimit, &p.HourlyMintLimit, validateStringNotEmpty),
		paramtypes.NewParamSetPair(KeyPaused, &p.Paused, validateBool),
		paramtypes.NewParamSetPair(KeyPausedChains, &p.PausedChains, validateStringSlice),
		paramtypes.NewParamSetPair(KeyAutoPauseEnabled, &p.AutoPauseEnabled, validateBool),
		paramtypes.NewParamSetPair(KeyAutoPauseThreshold, &p.AutoPauseThreshold, validateStringNotEmpty),
		paramtypes.NewParamSetPair(KeyEmergencyPauseAddresses, &p.EmergencyPauseAddresses, validateStringSlice),
		paramtypes.NewParamSetPair(KeyFraudProofWindow, &p.FraudProofWindow, validateFraudProofWindow),
		paramtypes.NewParamSetPair(KeySlashFraudSignature, &p.SlashFraudSignature, validateSlashFraction),
		paramtypes.NewParamSetPair(KeySlashDoubleSigning, &p.SlashDoubleSigning, validateSlashFraction),
		paramtypes.NewParamSetPair(KeySlashOffline, &p.SlashOffline, validateSlashFraction),
		paramtypes.NewParamSetPair(KeyMinSigningWindow, &p.MinSigningWindow, validateSigningWindow),
		paramtypes.NewParamSetPair(KeyMinSignedPerWindow, &p.MinSignedPerWindow, validateSlashFraction),
	}
}

func validateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return ErrInvalidParam
	}
	return nil
}

func validateUint64Core(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return ErrInvalidParam
	}
	return nil
}

// validateMinConfirmations ensures MinConfirmations meets security requirements
func validateMinConfirmations(i interface{}) error {
	val, ok := i.(uint64)
	if !ok {
		return ErrInvalidParam
	}
	if val < MinAllowedConfirmations {
		return fmt.Errorf("MinConfirmations must be >= %d for security, got %d",
			MinAllowedConfirmations, val)
	}
	return nil
}

func validateStringNotEmpty(i interface{}) error {
	s, ok := i.(string)
	if !ok {
		return ErrInvalidParam
	}
	if s == "" {
		return fmt.Errorf("value cannot be empty")
	}
	return nil
}

func validateStringSlice(i interface{}) error {
	_, ok := i.([]string)
	if !ok {
		return ErrInvalidParam
	}
	return nil
}

func validateSupplyCaps(i interface{}) error {
	caps, ok := i.(map[string]string)
	if !ok {
		return ErrInvalidParam
	}
	// Validate each supply cap value is a valid integer
	for denom, cap := range caps {
		if denom == "" {
			return fmt.Errorf("supply cap denom cannot be empty")
		}
		if _, ok := sdkmath.NewIntFromString(cap); !ok {
			return fmt.Errorf("invalid supply cap for %s: %s must be a valid integer", denom, cap)
		}
	}
	return nil
}

// validateFraudProofWindow ensures the fraud proof window is reasonable
func validateFraudProofWindow(i interface{}) error {
	val, ok := i.(int64)
	if !ok {
		return ErrInvalidParam
	}
	// Minimum 1 hour (3600 seconds), maximum 30 days (2592000 seconds)
	// This prevents too-short windows (no time to challenge) and too-long windows (tokens locked forever)
	if val < 3600 {
		return fmt.Errorf("fraud proof window must be at least 1 hour (3600 seconds), got %d", val)
	}
	if val > 2592000 {
		return fmt.Errorf("fraud proof window cannot exceed 30 days (2592000 seconds), got %d", val)
	}
	return nil
}

// validateSlashFraction ensures slash fractions are valid decimals between 0 and 1
func validateSlashFraction(i interface{}) error {
	val, ok := i.(string)
	if !ok {
		return ErrInvalidParam
	}
	if val == "" {
		return fmt.Errorf("slash fraction cannot be empty")
	}
	dec, err := sdkmath.LegacyNewDecFromStr(val)
	if err != nil {
		return fmt.Errorf("invalid decimal format for slash fraction: %s", val)
	}
	// Must be between 0 and 1 (0% to 100%)
	if dec.IsNegative() {
		return fmt.Errorf("slash fraction cannot be negative, got %s", val)
	}
	if dec.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("slash fraction cannot exceed 1.0 (100%%), got %s", val)
	}
	return nil
}

// validateSigningWindow ensures the signing window is reasonable
func validateSigningWindow(i interface{}) error {
	val, ok := i.(int64)
	if !ok {
		return ErrInvalidParam
	}
	// Minimum 100 blocks, maximum 100,000 blocks
	// This prevents too-short windows (no meaningful liveness tracking) and
	// too-long windows (excessive state storage)
	if val < 100 {
		return fmt.Errorf("signing window must be at least 100 blocks, got %d", val)
	}
	if val > 100000 {
		return fmt.Errorf("signing window cannot exceed 100,000 blocks, got %d", val)
	}
	return nil
}

// Validate performs comprehensive validation of bridge parameters
func (p Params) Validate() error {
	// CRITICAL SECURITY: Enforce minimum validator confirmations
	// This prevents a single compromised validator from draining the bridge
	if p.MinConfirmations < MinAllowedConfirmations {
		return fmt.Errorf("MinConfirmations must be >= %d for security (prevents single validator control), got %d",
			MinAllowedConfirmations, p.MinConfirmations)
	}

	// Validate fee is not excessive (max 10% = 1000 basis points)
	if p.BridgeFeeBasisPoints > 1000 {
		return fmt.Errorf("BridgeFeeBasisPoints cannot exceed 1000 (10%%), got %d",
			p.BridgeFeeBasisPoints)
	}

	// Validate threshold percentage is reasonable (must be > 50% for majority)
	if p.ValidatorThresholdPercentage < 51 || p.ValidatorThresholdPercentage > 100 {
		return fmt.Errorf("ValidatorThresholdPercentage must be between 51-100, got %d",
			p.ValidatorThresholdPercentage)
	}

	// Validate max transfer amount is a valid integer
	if p.MaxTransferAmount != "" {
		if _, ok := sdkmath.NewIntFromString(p.MaxTransferAmount); !ok {
			return fmt.Errorf("MaxTransferAmount must be a valid integer, got %s",
				p.MaxTransferAmount)
		}
	}

	// Validate auto-pause threshold is a valid integer
	if p.AutoPauseThreshold != "" {
		if _, ok := sdkmath.NewIntFromString(p.AutoPauseThreshold); !ok {
			return fmt.Errorf("AutoPauseThreshold must be a valid integer, got %s",
				p.AutoPauseThreshold)
		}
	}

	// Validate supply caps
	if err := validateSupplyCaps(p.SupplyCaps); err != nil {
		return err
	}

	// Validate daily mint limit
	if p.DailyMintLimit != "" {
		if _, ok := sdkmath.NewIntFromString(p.DailyMintLimit); !ok {
			return fmt.Errorf("DailyMintLimit must be a valid integer, got %s",
				p.DailyMintLimit)
		}
	}

	// Validate hourly mint limit
	if p.HourlyMintLimit != "" {
		if _, ok := sdkmath.NewIntFromString(p.HourlyMintLimit); !ok {
			return fmt.Errorf("HourlyMintLimit must be a valid integer, got %s",
				p.HourlyMintLimit)
		}
	}

	// Validate fraud proof window
	if err := validateFraudProofWindow(p.FraudProofWindow); err != nil {
		return err
	}

	// Validate slashing fractions
	if err := validateSlashFraction(p.SlashFraudSignature); err != nil {
		return fmt.Errorf("invalid SlashFraudSignature: %w", err)
	}
	if err := validateSlashFraction(p.SlashDoubleSigning); err != nil {
		return fmt.Errorf("invalid SlashDoubleSigning: %w", err)
	}
	if err := validateSlashFraction(p.SlashOffline); err != nil {
		return fmt.Errorf("invalid SlashOffline: %w", err)
	}
	if err := validateSlashFraction(p.MinSignedPerWindow); err != nil {
		return fmt.Errorf("invalid MinSignedPerWindow: %w", err)
	}

	// Validate signing window
	if err := validateSigningWindow(p.MinSigningWindow); err != nil {
		return err
	}

	return nil
}

// DefaultGenesis returns the default genesis state for the bridge module
func DefaultGenesis() *GenesisState {
	params := DefaultParams()
	return &GenesisState{
		Params: &BridgeParams{
			Enabled:                      params.BridgeEnabled,
			MinConfirmations:             params.MinConfirmations,
			BridgeFeeBasisPoints:         params.BridgeFeeBasisPoints,
			MaxTransferAmount:            params.MaxTransferAmount,
			ValidatorThresholdPercentage: params.ValidatorThresholdPercentage,
		},
	}
}
