package types

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ExecutionContext tracks contract execution state for reentrancy protection
type ExecutionContext struct {
	CallStack     []string          // Stack of contract addresses being executed
	CallDepth     uint32            // Current call depth
	MaxCallDepth  uint32            // Maximum allowed call depth
	GasConsumed   map[string]uint64 // Gas consumed per contract in this execution
	ExecutionTime time.Time         // When execution started
}

// NewExecutionContext creates a new execution context
func NewExecutionContext(maxCallDepth uint32) *ExecutionContext {
	return &ExecutionContext{
		CallStack:     make([]string, 0, maxCallDepth),
		CallDepth:     0,
		MaxCallDepth:  maxCallDepth,
		GasConsumed:   make(map[string]uint64),
		ExecutionTime: time.Now(),
	}
}

// PushContract adds a contract to the call stack
func (ec *ExecutionContext) PushContract(contractAddr string) error {
	// Check if already in call stack (reentrancy)
	for _, addr := range ec.CallStack {
		if addr == contractAddr {
			return ErrReentrancyDetected.Wrapf("contract %s already in call stack", contractAddr)
		}
	}

	// Check max depth
	if ec.CallDepth >= ec.MaxCallDepth {
		return ErrReentrancyDetected.Wrapf("max call depth %d exceeded", ec.MaxCallDepth)
	}

	ec.CallStack = append(ec.CallStack, contractAddr)
	ec.CallDepth++
	return nil
}

// PopContract removes a contract from the call stack
func (ec *ExecutionContext) PopContract(contractAddr string) error {
	if len(ec.CallStack) == 0 {
		return ErrSecurityViolation.Wrap("call stack underflow")
	}

	// Verify we're popping the right contract
	lastAddr := ec.CallStack[len(ec.CallStack)-1]
	if lastAddr != contractAddr {
		return ErrSecurityViolation.Wrapf("stack corruption: expected %s, got %s", lastAddr, contractAddr)
	}

	ec.CallStack = ec.CallStack[:len(ec.CallStack)-1]
	ec.CallDepth--
	return nil
}

// IsContractInStack checks if a contract is currently executing
func (ec *ExecutionContext) IsContractInStack(contractAddr string) bool {
	for _, addr := range ec.CallStack {
		if addr == contractAddr {
			return true
		}
	}
	return false
}

// RecordGasConsumption records gas consumed by a contract
func (ec *ExecutionContext) RecordGasConsumption(contractAddr string, gas uint64) {
	ec.GasConsumed[contractAddr] += gas
}

// GetGasConsumed returns total gas consumed by a contract
func (ec *ExecutionContext) GetGasConsumed(contractAddr string) uint64 {
	return ec.GasConsumed[contractAddr]
}

// CodeHashInfo stores information about contract code hashes
type CodeHashInfo struct {
	CodeID         uint64    `json:"code_id"`
	CodeHash       string    `json:"code_hash"`       // SHA256 hash of bytecode
	Creator        string    `json:"creator"`         // Who uploaded this code
	UploadTime     time.Time `json:"upload_time"`     // When code was uploaded
	IsTrusted      bool      `json:"is_trusted"`      // Whether code is on trusted list
	IsVerified     bool      `json:"is_verified"`     // Whether code has verified signature
	VerifierSig    []byte    `json:"verifier_sig"`    // Optional signature from verifier
	ReproducibleID string    `json:"reproducible_id"` // Optional reproducible build ID
}

// ComputeCodeHash calculates SHA256 hash of contract bytecode
func ComputeCodeHash(code []byte) string {
	hash := sha256.Sum256(code)
	return hex.EncodeToString(hash[:])
}

// WASMValidationResult holds results of bytecode validation
type WASMValidationResult struct {
	Valid              bool     `json:"valid"`
	HasValidMagicNumber bool     `json:"has_valid_magic_number"`
	HasValidFormat      bool     `json:"has_valid_format"`
	MaliciousPatterns   []string `json:"malicious_patterns"`
	ErrorMessage        string   `json:"error_message"`
}

// MaliciousPattern defines patterns to detect in WASM bytecode
type MaliciousPattern struct {
	Name        string
	Description string
	Pattern     []byte
}

// GetMaliciousPatterns returns list of patterns to scan for
func GetMaliciousPatterns() []MaliciousPattern {
	return []MaliciousPattern{
		{
			Name:        "env.exit",
			Description: "Attempts to call env.exit which can halt execution",
			Pattern:     []byte("env.exit"),
		},
		{
			Name:        "env.abort",
			Description: "Attempts to call env.abort which can terminate execution",
			Pattern:     []byte("env.abort"),
		},
		{
			Name:        "wasi_snapshot",
			Description: "WASI system calls that should not be in smart contracts",
			Pattern:     []byte("wasi_snapshot"),
		},
		{
			Name:        "proc_exit",
			Description: "Process exit calls that can terminate execution",
			Pattern:     []byte("proc_exit"),
		},
		{
			Name:        "fd_write",
			Description: "File descriptor write operations (WASI)",
			Pattern:     []byte("fd_write"),
		},
		{
			Name:        "fd_read",
			Description: "File descriptor read operations (WASI)",
			Pattern:     []byte("fd_read"),
		},
	}
}

// RateLimitTracker tracks rate limits for contract operations
type RateLimitTracker struct {
	ContractAddr   string            `json:"contract_addr"`
	LastReset      time.Time         `json:"last_reset"`
	ExecutionCount uint64            `json:"execution_count"`
	QueryCount     uint64            `json:"query_count"`
	CustomMsgCount uint64            `json:"custom_msg_count"`
	WindowDuration time.Duration     `json:"window_duration"`
	Limits         RateLimitSettings `json:"limits"`
}

// RateLimitSettings defines rate limit thresholds
type RateLimitSettings struct {
	MaxExecutionsPerWindow uint64 `json:"max_executions_per_window"`
	MaxQueriesPerWindow    uint64 `json:"max_queries_per_window"`
	MaxCustomMsgsPerWindow uint64 `json:"max_custom_msgs_per_window"`
}

// DefaultRateLimits returns default rate limit settings
func DefaultRateLimits() RateLimitSettings {
	return RateLimitSettings{
		MaxExecutionsPerWindow: 100,
		MaxQueriesPerWindow:    1000, // Higher for queries
		MaxCustomMsgsPerWindow: 50,
	}
}

// NewRateLimitTracker creates a new rate limit tracker
func NewRateLimitTracker(contractAddr string, windowDuration time.Duration) *RateLimitTracker {
	return &RateLimitTracker{
		ContractAddr:   contractAddr,
		LastReset:      time.Now(),
		ExecutionCount: 0,
		QueryCount:     0,
		CustomMsgCount: 0,
		WindowDuration: windowDuration,
		Limits:         DefaultRateLimits(),
	}
}

// CheckAndIncrementExecution checks if execution is allowed and increments counter
func (rlt *RateLimitTracker) CheckAndIncrementExecution() error {
	rlt.resetIfNeeded()
	if rlt.ExecutionCount >= rlt.Limits.MaxExecutionsPerWindow {
		return ErrSecurityViolation.Wrapf("execution rate limit exceeded: %d/%d",
			rlt.ExecutionCount, rlt.Limits.MaxExecutionsPerWindow)
	}
	rlt.ExecutionCount++
	return nil
}

// CheckAndIncrementQuery checks if query is allowed and increments counter
func (rlt *RateLimitTracker) CheckAndIncrementQuery() error {
	rlt.resetIfNeeded()
	if rlt.QueryCount >= rlt.Limits.MaxQueriesPerWindow {
		return ErrSecurityViolation.Wrapf("query rate limit exceeded: %d/%d",
			rlt.QueryCount, rlt.Limits.MaxQueriesPerWindow)
	}
	rlt.QueryCount++
	return nil
}

// CheckAndIncrementCustomMsg checks if custom message is allowed and increments counter
func (rlt *RateLimitTracker) CheckAndIncrementCustomMsg() error {
	rlt.resetIfNeeded()
	if rlt.CustomMsgCount >= rlt.Limits.MaxCustomMsgsPerWindow {
		return ErrSecurityViolation.Wrapf("custom message rate limit exceeded: %d/%d",
			rlt.CustomMsgCount, rlt.Limits.MaxCustomMsgsPerWindow)
	}
	rlt.CustomMsgCount++
	return nil
}

// resetIfNeeded resets counters if window has expired
func (rlt *RateLimitTracker) resetIfNeeded() {
	if time.Since(rlt.LastReset) >= rlt.WindowDuration {
		rlt.ExecutionCount = 0
		rlt.QueryCount = 0
		rlt.CustomMsgCount = 0
		rlt.LastReset = time.Now()
	}
}

// SecurityAuditEvent represents a security-related event for logging
type SecurityAuditEvent struct {
	EventType      string                 `json:"event_type"`
	ContractAddr   string                 `json:"contract_addr"`
	Sender         string                 `json:"sender"`
	BlockHeight    int64                  `json:"block_height"`
	Timestamp      time.Time              `json:"timestamp"`
	Success        bool                   `json:"success"`
	ErrorMessage   string                 `json:"error_message,omitempty"`
	AdditionalData map[string]interface{} `json:"additional_data,omitempty"`
}

// NewSecurityAuditEvent creates a new security audit event
func NewSecurityAuditEvent(
	eventType string,
	contractAddr string,
	sender string,
	ctx sdk.Context,
	success bool,
	errorMsg string,
) SecurityAuditEvent {
	return SecurityAuditEvent{
		EventType:      eventType,
		ContractAddr:   contractAddr,
		Sender:         sender,
		BlockHeight:    ctx.BlockHeight(),
		Timestamp:      ctx.BlockTime(),
		Success:        success,
		ErrorMessage:   errorMsg,
		AdditionalData: make(map[string]interface{}),
	}
}

// AddData adds additional data to the audit event
func (sae *SecurityAuditEvent) AddData(key string, value interface{}) {
	if sae.AdditionalData == nil {
		sae.AdditionalData = make(map[string]interface{})
	}
	sae.AdditionalData[key] = value
}

// ContractPermissions defines permissions for custom bindings
type ContractPermissions struct {
	ContractAddr          string   `json:"contract_addr"`
	CanUseCustomBindings  bool     `json:"can_use_custom_bindings"`
	CanRegisterVC         bool     `json:"can_register_vc"`
	CanQueryVC            bool     `json:"can_query_vc"`
	AllowedVCTypes        []string `json:"allowed_vc_types"`
	MaxVCRegistrations    uint64   `json:"max_vc_registrations"`
	MaxVCQueries          uint64   `json:"max_vc_queries"`
	IsApprovedForBindings bool     `json:"is_approved_for_bindings"`
}

// DefaultContractPermissions returns default (restricted) permissions
func DefaultContractPermissions(contractAddr string) ContractPermissions {
	return ContractPermissions{
		ContractAddr:          contractAddr,
		CanUseCustomBindings:  false,
		CanRegisterVC:         false,
		CanQueryVC:            false,
		AllowedVCTypes:        []string{},
		MaxVCRegistrations:    0,
		MaxVCQueries:          0,
		IsApprovedForBindings: false,
	}
}

// CanRegisterVCFor checks if contract can register VC for a specific type
func (cp *ContractPermissions) CanRegisterVCFor(vcType string) bool {
	if !cp.CanRegisterVC {
		return false
	}

	// If no specific types allowed, deny all
	if len(cp.AllowedVCTypes) == 0 {
		return false
	}

	// Check if this specific type is allowed
	for _, allowedType := range cp.AllowedVCTypes {
		if allowedType == vcType || allowedType == "*" {
			return true
		}
	}

	return false
}

// CanQueryVCsFor checks if contract can query VCs of a specific type
func (cp *ContractPermissions) CanQueryVCsFor(vcType string) bool {
	if !cp.CanQueryVC {
		return false
	}

	// If no specific types allowed, deny all
	if len(cp.AllowedVCTypes) == 0 {
		return false
	}

	// Check if this specific type is allowed
	for _, allowedType := range cp.AllowedVCTypes {
		if allowedType == vcType || allowedType == "*" {
			return true
		}
	}

	return false
}

// GasEstimate holds pre-flight gas estimation
type GasEstimate struct {
	StorageGas     uint64 `json:"storage_gas"`      // Gas for storage operations
	ComputationGas uint64 `json:"computation_gas"`  // Gas for computation
	CallGas        uint64 `json:"call_gas"`         // Gas for contract calls
	TotalEstimate  uint64 `json:"total_estimate"`   // Total estimated gas
	SafetyMargin   uint64 `json:"safety_margin"`    // Additional safety margin
}

// CalculateStorageGas calculates gas cost for storing code
func CalculateStorageGas(codeSize uint64, gasPerByte uint64) uint64 {
	return codeSize * gasPerByte
}

// CalculateTotalGas calculates total gas with safety margin
func (ge *GasEstimate) CalculateTotalGas() uint64 {
	baseGas := ge.StorageGas + ge.ComputationGas + ge.CallGas
	ge.TotalEstimate = baseGas
	ge.SafetyMargin = baseGas / 10 // 10% safety margin
	return ge.TotalEstimate + ge.SafetyMargin
}

// Constants for security limits
const (
	// MaxCallDepth is the maximum allowed call depth
	MaxCallDepth = 5

	// GasPerByte is the gas cost per byte of code storage
	GasPerByte = 100

	// WASMMagicNumber is the WASM binary magic number (0x00 0x61 0x73 0x6d)
	WASMMagicNumberByte0 = 0x00
	WASMMagicNumberByte1 = 0x61
	WASMMagicNumberByte2 = 0x73
	WASMMagicNumberByte3 = 0x6d

	// MinContractSize is the minimum valid WASM contract size
	MinContractSize = 8 // At least magic number + version

	// MaxVCDataSize is the maximum size for VC data in custom bindings
	MaxVCDataSize = 1024 * 100 // 100KB

	// RateLimitWindow is the default rate limit window duration
	RateLimitWindow = 1 * time.Minute
)

// Key prefixes for security-related storage
var (
	// CodeHashPrefix stores code hash information
	CodeHashPrefix = []byte{0x10}

	// ExecutionContextPrefix stores execution contexts (transient)
	ExecutionContextPrefix = []byte{0x11}

	// RateLimitPrefix stores rate limit trackers
	RateLimitPrefix = []byte{0x12}

	// ContractPermissionsPrefix stores contract permissions
	ContractPermissionsPrefix = []byte{0x13}

	// TrustedCodeHashPrefix stores trusted code hashes
	TrustedCodeHashPrefix = []byte{0x14}

	// SecurityAuditLogPrefix stores security audit logs
	SecurityAuditLogPrefix = []byte{0x15}
)

// GetCodeHashKey returns the storage key for code hash info
func GetCodeHashKey(codeID uint64) []byte {
	return append(CodeHashPrefix, sdk.Uint64ToBigEndian(codeID)...)
}

// GetExecutionContextKey returns the storage key for execution context
func GetExecutionContextKey(txHash string) []byte {
	return append(ExecutionContextPrefix, []byte(txHash)...)
}

// GetRateLimitKey returns the storage key for rate limit tracker
func GetRateLimitKey(contractAddr string) []byte {
	return append(RateLimitPrefix, []byte(contractAddr)...)
}

// GetContractPermissionsKey returns the storage key for contract permissions
func GetContractPermissionsKey(contractAddr string) []byte {
	return append(ContractPermissionsPrefix, []byte(contractAddr)...)
}

// GetTrustedCodeHashKey returns the storage key for trusted code hash
func GetTrustedCodeHashKey(codeHash string) []byte {
	return append(TrustedCodeHashPrefix, []byte(codeHash)...)
}

// GetSecurityAuditLogKey returns the storage key for security audit log
func GetSecurityAuditLogKey(blockHeight int64, index uint64) []byte {
	key := append(SecurityAuditLogPrefix, sdk.Uint64ToBigEndian(uint64(blockHeight))...)
	return append(key, sdk.Uint64ToBigEndian(index)...)
}
