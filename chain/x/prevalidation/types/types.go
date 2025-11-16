package types

import (
	"fmt"

	pb "github.com/aequitas/aura/proto/aura/prevalidation/v1beta1"
)

// Type aliases for proto types
type (
	TransactionType            = pb.TransactionType
	ValidationStatus           = pb.ValidationStatus
	CacheStrategy              = pb.CacheStrategy
	PreValidatedTransaction    = pb.PreValidatedTransaction
	ValidationTemplate         = pb.ValidationTemplate
	ValidationMetadata         = pb.ValidationMetadata
	PreValidationMetrics       = pb.PreValidationMetrics
	TypeMetrics                = pb.TypeMetrics
	HourlyMetrics              = pb.HourlyMetrics
	ControlGroupMetrics        = pb.ControlGroupMetrics
	SchedulerConfig            = pb.SchedulerConfig
	AutoScalingConfig          = pb.AutoScalingConfig
	Params                     = pb.Params
	GenesisState               = pb.GenesisState
	TemplateStats              = pb.TemplateStats
	EventPreValidationCreated  = pb.EventPreValidationCreated
	EventPreValidationExecuted = pb.EventPreValidationExecuted
	EventPreValidationExpired  = pb.EventPreValidationExpired
	EventCacheHit              = pb.EventCacheHit
	EventCacheMiss             = pb.EventCacheMiss
	EventSchedulerRun          = pb.EventSchedulerRun
	EventAutoScaling           = pb.EventAutoScaling
	EventMetricsUpdate         = pb.EventMetricsUpdate
)

// Constants for transaction types
const (
	TxTypeUnspecified           = pb.TransactionType_TX_TYPE_UNSPECIFIED
	TxTypeIRCompletion          = pb.TransactionType_TX_TYPE_IR_COMPLETION
	TxTypeDexSwap               = pb.TransactionType_TX_TYPE_DEX_SWAP
	TxTypeLPDeposit             = pb.TransactionType_TX_TYPE_LP_DEPOSIT
	TxTypeLPWithdrawal          = pb.TransactionType_TX_TYPE_LP_WITHDRAWAL
	TxTypeVCMint                = pb.TransactionType_TX_TYPE_VC_MINT
	TxTypeBridgeTransfer        = pb.TransactionType_TX_TYPE_BRIDGE_TRANSFER
	TxTypeConfidenceScoreUpdate = pb.TransactionType_TX_TYPE_CONFIDENCE_SCORE_UPDATE
	TxTypeIdentityChange        = pb.TransactionType_TX_TYPE_IDENTITY_CHANGE
)

// Constants for validation status
const (
	ValidationStatusUnspecified = pb.ValidationStatus_VALIDATION_STATUS_UNSPECIFIED
	ValidationStatusPending     = pb.ValidationStatus_VALIDATION_STATUS_PENDING
	ValidationStatusValidated   = pb.ValidationStatus_VALIDATION_STATUS_VALIDATED
	ValidationStatusExpired     = pb.ValidationStatus_VALIDATION_STATUS_EXPIRED
	ValidationStatusExecuted    = pb.ValidationStatus_VALIDATION_STATUS_EXECUTED
	ValidationStatusFailed      = pb.ValidationStatus_VALIDATION_STATUS_FAILED
)

// Constants for cache strategy
const (
	CacheStrategyUnspecified = pb.CacheStrategy_CACHE_STRATEGY_UNSPECIFIED
	CacheStrategyLRU         = pb.CacheStrategy_CACHE_STRATEGY_LRU
	CacheStrategyLFU         = pb.CacheStrategy_CACHE_STRATEGY_LFU
	CacheStrategyFIFO        = pb.CacheStrategy_CACHE_STRATEGY_FIFO
	CacheStrategyAdaptive    = pb.CacheStrategy_CACHE_STRATEGY_ADAPTIVE
)

// Error definitions
var (
	ErrPreValidationNotFound    = fmt.Errorf("pre-validated transaction not found")
	ErrTemplateNotFound         = fmt.Errorf("validation template not found")
	ErrPreValidationExpired     = fmt.Errorf("pre-validation has expired")
	ErrPreValidationAlreadyUsed = fmt.Errorf("pre-validation already executed")
	ErrInvalidTransactionType   = fmt.Errorf("invalid transaction type")
	ErrInvalidTemplate          = fmt.Errorf("invalid validation template")
	ErrValidationFailed         = fmt.Errorf("pre-validation failed")
	ErrCacheFull                = fmt.Errorf("pre-validation cache is full")
	ErrInsufficientConfidence   = fmt.Errorf("insufficient confidence score for pre-validation")
	ErrEncryptionFailed         = fmt.Errorf("encryption failed")
	ErrDecryptionFailed         = fmt.Errorf("decryption failed")
	ErrSchedulerDisabled        = fmt.Errorf("scheduler is disabled")
	ErrNotOffPeakHours          = fmt.Errorf("not in off-peak hours")
	ErrMaxAttemptsExceeded      = fmt.Errorf("max validation attempts exceeded")
	ErrInvalidParameters        = fmt.Errorf("invalid parameters")
)

// DefaultParams returns default module parameters
func DefaultParams() *Params {
	return &Params{
		Enabled: true,
		SchedulerConfig: &SchedulerConfig{
			OffPeakHours:       []uint32{2, 3, 4, 5, 6}, // 2am-6am
			Timezone:           "UTC",
			Enabled:            true,
			RunIntervalMinutes: 30,
			MaxPerRun:          1000,
			AllowPeakHours:     false,
		},
		AutoScalingConfig: &AutoScalingConfig{
			Enabled: true,
			InitialAmounts: map[string]uint64{
				TxTypeIRCompletion.String():          100, // Highest frequency
				TxTypeDexSwap.String():               50,
				TxTypeLPDeposit.String():             30,
				TxTypeLPWithdrawal.String():          30,
				TxTypeVCMint.String():                20,
				TxTypeBridgeTransfer.String():        15,
				TxTypeConfidenceScoreUpdate.String(): 25,
				TxTypeIdentityChange.String():        10,
			},
			TargetCacheHitRate: 0.80, // 80% target hit rate
			MinCacheHitRate:    0.50, // Scale down below 50%
			MaxAmounts: map[string]uint64{
				TxTypeIRCompletion.String():          1000,
				TxTypeDexSwap.String():               500,
				TxTypeLPDeposit.String():             300,
				TxTypeLPWithdrawal.String():          300,
				TxTypeVCMint.String():                200,
				TxTypeBridgeTransfer.String():        150,
				TxTypeConfidenceScoreUpdate.String(): 250,
				TxTypeIdentityChange.String():        100,
			},
			ScaleUpFactor:         1.5,
			ScaleDownFactor:       0.75,
			CooldownMinutes:       60,
			EvaluationPeriodHours: 24,
		},
		CacheStrategy:              CacheStrategyAdaptive,
		MaxCacheSize:               10000,
		ExpiryHours:                72, // 3 days
		EncryptionAlgorithm:        "AES-256-GCM",
		ControlGroupPercentage:     5.0, // 5% control group
		MinConfidenceScore:         100,
		EnergyCostPerValidationKwh: 0.0001, // 0.1 Wh
		EnergyCostPerExecutionKwh:  0.001,  // 1 Wh
		MetricsEnabled:             true,
		DetailedLogging:            false,
		MaxValidationAttempts:      3,
		RetryDelaySeconds:          30,
	}
}

// ValidateParams validates the parameters
func ValidateParams(p *Params) error {
	if p.SchedulerConfig == nil {
		return fmt.Errorf("scheduler config is required")
	}
	if p.AutoScalingConfig == nil {
		return fmt.Errorf("auto-scaling config is required")
	}
	if p.MaxCacheSize == 0 {
		return fmt.Errorf("max cache size must be greater than 0")
	}
	if p.ExpiryHours == 0 {
		return fmt.Errorf("expiry hours must be greater than 0")
	}
	if p.ControlGroupPercentage < 0 || p.ControlGroupPercentage > 100 {
		return fmt.Errorf("control group percentage must be between 0 and 100")
	}
	if p.EncryptionAlgorithm == "" {
		return fmt.Errorf("encryption algorithm is required")
	}
	return nil
}

// Note: Methods on proto type aliases have been moved to helpers.go
// to avoid "cannot define new methods on non-local type" compilation errors

// TransactionTypeName returns a human-readable name for a transaction type
func TransactionTypeName(txType TransactionType) string {
	switch txType {
	case TxTypeIRCompletion:
		return "IR Completion"
	case TxTypeDexSwap:
		return "DEX Swap"
	case TxTypeLPDeposit:
		return "LP Deposit"
	case TxTypeLPWithdrawal:
		return "LP Withdrawal"
	case TxTypeVCMint:
		return "VC Mint"
	case TxTypeBridgeTransfer:
		return "Bridge Transfer"
	case TxTypeConfidenceScoreUpdate:
		return "Confidence Score Update"
	case TxTypeIdentityChange:
		return "Identity Change"
	default:
		return "Unknown"
	}
}

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *GenesisState {
	params := DefaultParams()
	return &GenesisState{
		Params:                   params,
		PreValidatedTransactions: []*PreValidatedTransaction{},
		Templates:                []*ValidationTemplate{},
		Metrics: &PreValidationMetrics{
			MetricsByType: make(map[string]*TypeMetrics),
			CurrentHour:   &HourlyMetrics{},
			Last_24Hours:  []*HourlyMetrics{},
			ControlGroup:  &ControlGroupMetrics{},
		},
	}
}
