package keeper

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/contractregistry/types"
	pb "github.com/aequitas/aura/proto/aura/contractregistry/v1beta1"
)

// VerificationStatus represents the verification state of a contract
type VerificationStatus string

const (
	VerificationStatusUnverified VerificationStatus = "unverified"
	VerificationStatusPending    VerificationStatus = "pending"
	VerificationStatusVerified   VerificationStatus = "verified"
	VerificationStatusFailed     VerificationStatus = "failed"
	VerificationStatusExpired    VerificationStatus = "expired"
	VerificationStatusRevoked    VerificationStatus = "revoked"
)

// VerificationLevel represents the depth of verification
type VerificationLevel int

const (
	VerificationLevelBasic     VerificationLevel = 1 // Source code matches bytecode
	VerificationLevelStandard  VerificationLevel = 2 // Basic + metadata verification
	VerificationLevelFull      VerificationLevel = 3 // Standard + audit verification
	VerificationLevelCertified VerificationLevel = 4 // Full + third-party certification
)

// VerificationResult contains the complete verification outcome
type VerificationResult struct {
	ContractAddress   string
	Status            VerificationStatus
	Level             VerificationLevel
	Timestamp         time.Time
	ExpirationTime    time.Time
	SourceCodeHash    string
	BytecodeHash      string
	CompilerVersion   string
	OptimizationUsed  bool
	OptimizationRuns  uint64
	ConstructorArgs   []byte
	Libraries         map[string]string
	VerifiedBy        string
	VerificationProof []byte
	Issues            []VerificationIssue
	Warnings          []string
	SecurityFlags     []string
	AuditReferences   []string
	CertificationID   string
}

// VerificationIssue represents a problem found during verification
type VerificationIssue struct {
	Severity    string // "critical", "high", "medium", "low", "info"
	Code        string
	Description string
	Location    string
	Suggestion  string
}

// VerificationRequest represents a request to verify a contract
type VerificationRequest struct {
	ContractAddress  string
	SourceCode       string
	CompilerVersion  string
	OptimizationUsed bool
	OptimizationRuns uint64
	ConstructorArgs  []byte
	Libraries        map[string]string
	RequestedLevel   VerificationLevel
	Requester        string
}

// BytecodeAnalysis contains analysis of contract bytecode
type BytecodeAnalysis struct {
	Size                int
	HasConstructor      bool
	HasFallback         bool
	HasReceive          bool
	ExternalCalls       int
	StorageSlots        int
	EventSignatures     []string
	FunctionSelectors   []string
	ImmutableReferences map[string][]int
	Metadata            []byte
	MetadataHash        string
}

// SourceCodeAnalysis contains source code analysis results
type SourceCodeAnalysis struct {
	SecurityFlags []string
	Issues        []VerificationIssue
	Warnings      []string
}

// VerifyContract performs comprehensive contract verification
func (k Keeper) VerifyContract(ctx sdk.Context, req VerificationRequest) (*VerificationResult, error) {
	// Get contract info
	info, found := k.GetContractInfo(ctx, req.ContractAddress)
	if !found {
		return nil, types.ErrContractNotFound
	}

	result := &VerificationResult{
		ContractAddress: req.ContractAddress,
		Status:          VerificationStatusPending,
		Level:           req.RequestedLevel,
		Timestamp:       ctx.BlockTime(),
		Libraries:       req.Libraries,
		VerifiedBy:      req.Requester,
		Issues:          make([]VerificationIssue, 0),
		Warnings:        make([]string, 0),
		SecurityFlags:   make([]string, 0),
		AuditReferences: make([]string, 0),
	}

	// Step 1: Compute source code hash
	sourceHash := sha256.Sum256([]byte(req.SourceCode))
	result.SourceCodeHash = hex.EncodeToString(sourceHash[:])

	// Step 2: Verify bytecode exists and compute hash
	if info.CodeId == 0 {
		result.Status = VerificationStatusFailed
		result.Issues = append(result.Issues, VerificationIssue{
			Severity:    "critical",
			Code:        "NO_BYTECODE",
			Description: "Contract has no associated bytecode",
			Suggestion:  "Ensure contract is properly deployed with bytecode",
		})
		return result, nil
	}

	// Step 3: Verify compiler version compatibility
	if err := k.verifyCompilerVersion(req.CompilerVersion); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Compiler version warning: %v", err))
	}
	result.CompilerVersion = req.CompilerVersion
	result.OptimizationUsed = req.OptimizationUsed
	result.OptimizationRuns = req.OptimizationRuns

	// Step 4: Perform source code analysis
	sourceAnalysis := k.analyzeSourceCode(req.SourceCode)
	result.SecurityFlags = append(result.SecurityFlags, sourceAnalysis.SecurityFlags...)
	result.Issues = append(result.Issues, sourceAnalysis.Issues...)
	result.Warnings = append(result.Warnings, sourceAnalysis.Warnings...)

	// Step 5: Level-specific verification
	switch req.RequestedLevel {
	case VerificationLevelBasic:
		if err := k.performBasicVerification(ctx, &info, req, result); err != nil {
			result.Status = VerificationStatusFailed
			return result, nil
		}
	case VerificationLevelStandard:
		if err := k.performStandardVerification(ctx, &info, req, result); err != nil {
			result.Status = VerificationStatusFailed
			return result, nil
		}
	case VerificationLevelFull:
		if err := k.performFullVerification(ctx, &info, req, result); err != nil {
			result.Status = VerificationStatusFailed
			return result, nil
		}
	case VerificationLevelCertified:
		if err := k.performCertifiedVerification(ctx, &info, req, result); err != nil {
			result.Status = VerificationStatusFailed
			return result, nil
		}
	}

	// Step 6: Check for critical issues
	hasCriticalIssues := false
	for _, issue := range result.Issues {
		if issue.Severity == "critical" {
			hasCriticalIssues = true
			break
		}
	}

	if hasCriticalIssues {
		result.Status = VerificationStatusFailed
	} else {
		result.Status = VerificationStatusVerified
		// Set expiration (1 year for verified contracts)
		result.ExpirationTime = ctx.BlockTime().Add(365 * 24 * time.Hour)
	}

	// Step 7: Store verification result
	k.storeVerificationResult(ctx, result)

	// Emit verification event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"contract_verified",
			sdk.NewAttribute("contract_address", req.ContractAddress),
			sdk.NewAttribute("status", string(result.Status)),
			sdk.NewAttribute("level", fmt.Sprintf("%d", result.Level)),
			sdk.NewAttribute("verified_by", req.Requester),
		),
	)

	return result, nil
}

// analyzeSourceCode performs static analysis on source code
func (k Keeper) analyzeSourceCode(sourceCode string) SourceCodeAnalysis {
	analysis := SourceCodeAnalysis{
		SecurityFlags: make([]string, 0),
		Issues:        make([]VerificationIssue, 0),
		Warnings:      make([]string, 0),
	}

	// Check for dangerous patterns
	dangerousPatterns := map[string]struct {
		pattern  string
		severity string
		code     string
		desc     string
		suggest  string
	}{
		"delegatecall": {
			pattern:  `delegatecall`,
			severity: "high",
			code:     "DELEGATECALL_DETECTED",
			desc:     "Contract uses delegatecall which can be dangerous if not properly secured",
			suggest:  "Ensure delegatecall targets are trusted and inputs are validated",
		},
		"selfdestruct": {
			pattern:  `selfdestruct|suicide`,
			severity: "critical",
			code:     "SELFDESTRUCT_DETECTED",
			desc:     "Contract contains selfdestruct which can permanently destroy the contract",
			suggest:  "Consider removing selfdestruct or adding strict access controls",
		},
		"tx_origin": {
			pattern:  `tx\.origin`,
			severity: "high",
			code:     "TX_ORIGIN_USAGE",
			desc:     "Using tx.origin for authorization is vulnerable to phishing attacks",
			suggest:  "Use msg.sender instead of tx.origin for authorization",
		},
		"unchecked_send": {
			pattern:  `\.send\s*\(.*\)\s*;`,
			severity: "medium",
			code:     "UNCHECKED_SEND",
			desc:     "Send return value may not be checked, potentially losing funds",
			suggest:  "Always check the return value of send() or use transfer()",
		},
		"timestamp_dependence": {
			pattern:  `block\.timestamp|now`,
			severity: "low",
			code:     "TIMESTAMP_DEPENDENCE",
			desc:     "Contract relies on block.timestamp which can be manipulated by miners",
			suggest:  "Avoid using timestamps for critical logic or use block numbers",
		},
		"assembly": {
			pattern:  `assembly\s*\{`,
			severity: "info",
			code:     "INLINE_ASSEMBLY",
			desc:     "Contract uses inline assembly which requires extra scrutiny",
			suggest:  "Ensure assembly code is thoroughly reviewed and tested",
		},
		"external_call_in_loop": {
			pattern:  `for\s*\([^)]*\)\s*\{[^}]*\.(call|send|transfer)`,
			severity: "high",
			code:     "EXTERNAL_CALL_IN_LOOP",
			desc:     "External calls in loops can lead to DoS vulnerabilities",
			suggest:  "Use pull-over-push pattern for payments",
		},
	}

	for name, check := range dangerousPatterns {
		matched, _ := regexp.MatchString(check.pattern, sourceCode)
		if matched {
			analysis.SecurityFlags = append(analysis.SecurityFlags, name)
			if check.severity == "critical" || check.severity == "high" {
				analysis.Issues = append(analysis.Issues, VerificationIssue{
					Severity:    check.severity,
					Code:        check.code,
					Description: check.desc,
					Suggestion:  check.suggest,
				})
			} else {
				analysis.Warnings = append(analysis.Warnings, check.desc)
			}
		}
	}

	// Check for reentrancy vulnerabilities
	if k.detectReentrancyRisk(sourceCode) {
		analysis.SecurityFlags = append(analysis.SecurityFlags, "potential_reentrancy")
		analysis.Issues = append(analysis.Issues, VerificationIssue{
			Severity:    "critical",
			Code:        "REENTRANCY_RISK",
			Description: "Contract may be vulnerable to reentrancy attacks",
			Suggestion:  "Use checks-effects-interactions pattern and/or ReentrancyGuard",
		})
	}

	// Check for integer overflow patterns (for older Solidity versions)
	if k.detectOverflowRisk(sourceCode) {
		analysis.SecurityFlags = append(analysis.SecurityFlags, "potential_overflow")
		analysis.Warnings = append(analysis.Warnings,
			"Contract may be vulnerable to integer overflow/underflow if using Solidity < 0.8.0")
	}

	// Check for access control patterns
	if !k.hasAccessControl(sourceCode) {
		analysis.Warnings = append(analysis.Warnings,
			"Contract may lack proper access control mechanisms")
	}

	return analysis
}

// detectReentrancyRisk checks for potential reentrancy vulnerabilities
func (k Keeper) detectReentrancyRisk(sourceCode string) bool {
	// Pattern: external call followed by state change
	patterns := []string{
		`\.call\{.*\}\([^)]*\)[\s\S]*?=`,   // call followed by assignment
		`\.transfer\([^)]*\)[\s\S]*?=`,     // transfer followed by assignment
		`\.send\([^)]*\)[\s\S]*?=`,         // send followed by assignment
	}

	for _, pattern := range patterns {
		matched, _ := regexp.MatchString(pattern, sourceCode)
		if matched {
			return true
		}
	}

	return false
}

// detectOverflowRisk checks for potential overflow vulnerabilities
func (k Keeper) detectOverflowRisk(sourceCode string) bool {
	// Check for arithmetic without SafeMath in older versions
	hasArithmetic, _ := regexp.MatchString(`[+\-*/]\s*=`, sourceCode)
	hasSafeMath, _ := regexp.MatchString(`SafeMath|using\s+.*for\s+uint`, sourceCode)
	hasSolidity08, _ := regexp.MatchString(`pragma solidity\s*[\^>=]*0\.[89]`, sourceCode)

	return hasArithmetic && !hasSafeMath && !hasSolidity08
}

// hasAccessControl checks if contract has access control patterns
func (k Keeper) hasAccessControl(sourceCode string) bool {
	accessPatterns := []string{
		`onlyOwner`,
		`onlyAdmin`,
		`require\s*\(\s*msg\.sender\s*==`,
		`AccessControl`,
		`Ownable`,
		`modifier\s+only`,
	}

	for _, pattern := range accessPatterns {
		matched, _ := regexp.MatchString(pattern, sourceCode)
		if matched {
			return true
		}
	}

	return false
}

// verifyCompilerVersion validates the compiler version
func (k Keeper) verifyCompilerVersion(version string) error {
	// List of approved compiler versions
	approvedVersions := []string{
		"0.8.0", "0.8.1", "0.8.2", "0.8.3", "0.8.4", "0.8.5", "0.8.6", "0.8.7",
		"0.8.8", "0.8.9", "0.8.10", "0.8.11", "0.8.12", "0.8.13", "0.8.14",
		"0.8.15", "0.8.16", "0.8.17", "0.8.18", "0.8.19", "0.8.20", "0.8.21",
		"0.8.22", "0.8.23", "0.8.24", "0.8.25",
	}

	// Clean version string
	cleanVersion := strings.TrimPrefix(version, "v")
	cleanVersion = strings.Split(cleanVersion, "+")[0] // Remove commit hash

	for _, approved := range approvedVersions {
		if cleanVersion == approved {
			return nil
		}
	}

	// Check for deprecated versions
	deprecatedVersions := []string{
		"0.4", "0.5", "0.6", "0.7",
	}
	for _, deprecated := range deprecatedVersions {
		if strings.HasPrefix(cleanVersion, deprecated) {
			return fmt.Errorf("compiler version %s is deprecated and may have security vulnerabilities", version)
		}
	}

	return fmt.Errorf("compiler version %s is not in the approved list", version)
}

// performBasicVerification performs level 1 verification
func (k Keeper) performBasicVerification(ctx sdk.Context, info *pb.ContractInfo, req VerificationRequest, result *VerificationResult) error {
	// Check source code URL availability
	if info.Metadata.SourceCodeUrl != "" {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("Source code URL: %s", info.Metadata.SourceCodeUrl))
	}

	return nil
}

// performStandardVerification performs level 2 verification
func (k Keeper) performStandardVerification(ctx sdk.Context, info *pb.ContractInfo, req VerificationRequest, result *VerificationResult) error {
	// First perform basic verification
	if err := k.performBasicVerification(ctx, info, req, result); err != nil {
		return err
	}

	// Verify metadata completeness
	if info.Metadata.Description == "" {
		result.Warnings = append(result.Warnings, "Contract lacks description")
	}
	if info.Metadata.Version == "" {
		result.Warnings = append(result.Warnings, "Contract version not specified")
	}

	// Check constructor arguments if provided
	if len(req.ConstructorArgs) > 0 {
		result.ConstructorArgs = req.ConstructorArgs
	}

	return nil
}

// performFullVerification performs level 3 verification
func (k Keeper) performFullVerification(ctx sdk.Context, info *pb.ContractInfo, req VerificationRequest, result *VerificationResult) error {
	// First perform standard verification
	if err := k.performStandardVerification(ctx, info, req, result); err != nil {
		return err
	}

	// Verify audit status via compliance requirements
	if !info.Compliance.RequireAudit {
		result.Issues = append(result.Issues, VerificationIssue{
			Severity:    "medium",
			Code:        "NO_AUDIT_REQUIREMENT",
			Description: "Contract does not require audit",
			Suggestion:  "Consider getting the contract audited by a reputable firm",
		})
	} else if info.Compliance.AuditReportUri != "" {
		result.AuditReferences = append(result.AuditReferences, info.Compliance.AuditReportUri)

		// Check audit date if available
		if info.Compliance.LastAuditDate != nil && !info.Compliance.LastAuditDate.IsZero() {
			daysSinceAudit := ctx.BlockTime().Sub(*info.Compliance.LastAuditDate).Hours() / 24
			if daysSinceAudit > 365 {
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("Audit is over 1 year old (%.0f days)", daysSinceAudit))
			}
		}
	}

	// Verify security policy
	if !info.SecurityPolicy.AllowPause {
		result.Warnings = append(result.Warnings,
			"Contract pause functionality is disabled")
	}
	if info.SecurityPolicy.RateLimitPerUser == 0 {
		result.Warnings = append(result.Warnings,
			"No rate limiting configured for this contract")
	}

	return nil
}

// performCertifiedVerification performs level 4 verification
func (k Keeper) performCertifiedVerification(ctx sdk.Context, info *pb.ContractInfo, req VerificationRequest, result *VerificationResult) error {
	// First perform full verification
	if err := k.performFullVerification(ctx, info, req, result); err != nil {
		return err
	}

	// Require audit for certification
	if info.Compliance.AuditReportUri == "" {
		result.Issues = append(result.Issues, VerificationIssue{
			Severity:    "high",
			Code:        "NO_AUDIT_FOR_CERTIFICATION",
			Description: "Contract requires an audit report for certification",
			Suggestion:  "Obtain a security audit from an approved auditor",
		})
		return fmt.Errorf("no audit for certification")
	}

	// Check audit is recent (within 1 year)
	if info.Compliance.LastAuditDate != nil && !info.Compliance.LastAuditDate.IsZero() {
		daysSinceAudit := ctx.BlockTime().Sub(*info.Compliance.LastAuditDate).Hours() / 24
		if daysSinceAudit > 365 {
			result.Issues = append(result.Issues, VerificationIssue{
				Severity:    "high",
				Code:        "STALE_AUDIT",
				Description: "Audit is more than 1 year old",
				Suggestion:  "Obtain an updated security audit",
			})
			return fmt.Errorf("audit too old for certification")
		}
	}

	// Generate certification ID
	certData := fmt.Sprintf("%s-%s-%d",
		req.ContractAddress,
		result.SourceCodeHash,
		ctx.BlockTime().Unix())
	certHash := sha256.Sum256([]byte(certData))
	result.CertificationID = hex.EncodeToString(certHash[:16])

	return nil
}

// storeVerificationResult stores the verification result using custom encoding
func (k Keeper) storeVerificationResult(ctx sdk.Context, result *VerificationResult) {
	store := ctx.KVStore(k.storeKey)
	key := types.VerificationResultKey(result.ContractAddress)

	// Encode verification result manually since we don't have a proto message
	// Format: status(1) + level(1) + timestamp(8) + expiration(8) + sourceHashLen(2) + sourceHash + verifiedByLen(2) + verifiedBy + certIdLen(2) + certId
	var buf bytes.Buffer

	// Status (1 byte)
	statusByte := byte(0)
	switch result.Status {
	case VerificationStatusVerified:
		statusByte = 1
	case VerificationStatusFailed:
		statusByte = 2
	case VerificationStatusExpired:
		statusByte = 3
	case VerificationStatusRevoked:
		statusByte = 4
	case VerificationStatusPending:
		statusByte = 5
	}
	buf.WriteByte(statusByte)

	// Level (1 byte)
	buf.WriteByte(byte(result.Level))

	// Timestamp (8 bytes)
	timestampBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampBytes, uint64(result.Timestamp.Unix()))
	buf.Write(timestampBytes)

	// Expiration (8 bytes)
	expirationBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(expirationBytes, uint64(result.ExpirationTime.Unix()))
	buf.Write(expirationBytes)

	// Source code hash (length-prefixed)
	sourceHashBytes := []byte(result.SourceCodeHash)
	lenBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(lenBytes, uint16(len(sourceHashBytes)))
	buf.Write(lenBytes)
	buf.Write(sourceHashBytes)

	// Bytecode hash (length-prefixed)
	bytecodeHashBytes := []byte(result.BytecodeHash)
	binary.BigEndian.PutUint16(lenBytes, uint16(len(bytecodeHashBytes)))
	buf.Write(lenBytes)
	buf.Write(bytecodeHashBytes)

	// Compiler version (length-prefixed)
	compilerBytes := []byte(result.CompilerVersion)
	binary.BigEndian.PutUint16(lenBytes, uint16(len(compilerBytes)))
	buf.Write(lenBytes)
	buf.Write(compilerBytes)

	// Verified by (length-prefixed)
	verifiedByBytes := []byte(result.VerifiedBy)
	binary.BigEndian.PutUint16(lenBytes, uint16(len(verifiedByBytes)))
	buf.Write(lenBytes)
	buf.Write(verifiedByBytes)

	// Certification ID (length-prefixed)
	certIdBytes := []byte(result.CertificationID)
	binary.BigEndian.PutUint16(lenBytes, uint16(len(certIdBytes)))
	buf.Write(lenBytes)
	buf.Write(certIdBytes)

	store.Set(key, buf.Bytes())
}

// GetVerificationResult retrieves a stored verification result
func (k Keeper) GetVerificationResult(ctx sdk.Context, contractAddr string) (*VerificationResult, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.VerificationResultKey(contractAddr)

	bz := store.Get(key)
	if len(bz) < 18 { // Minimum: status(1) + level(1) + timestamp(8) + expiration(8)
		return nil, false
	}

	result := &VerificationResult{
		ContractAddress: contractAddr,
		Issues:          make([]VerificationIssue, 0),
		Warnings:        make([]string, 0),
		SecurityFlags:   make([]string, 0),
		AuditReferences: make([]string, 0),
		Libraries:       make(map[string]string),
	}

	// Parse status
	switch bz[0] {
	case 1:
		result.Status = VerificationStatusVerified
	case 2:
		result.Status = VerificationStatusFailed
	case 3:
		result.Status = VerificationStatusExpired
	case 4:
		result.Status = VerificationStatusRevoked
	case 5:
		result.Status = VerificationStatusPending
	default:
		result.Status = VerificationStatusUnverified
	}

	// Parse level
	result.Level = VerificationLevel(bz[1])

	// Parse timestamp
	result.Timestamp = time.Unix(int64(binary.BigEndian.Uint64(bz[2:10])), 0)

	// Parse expiration
	result.ExpirationTime = time.Unix(int64(binary.BigEndian.Uint64(bz[10:18])), 0)

	// Parse length-prefixed strings
	offset := 18

	// Source code hash
	if offset+2 <= len(bz) {
		strLen := int(binary.BigEndian.Uint16(bz[offset : offset+2]))
		offset += 2
		if offset+strLen <= len(bz) {
			result.SourceCodeHash = string(bz[offset : offset+strLen])
			offset += strLen
		}
	}

	// Bytecode hash
	if offset+2 <= len(bz) {
		strLen := int(binary.BigEndian.Uint16(bz[offset : offset+2]))
		offset += 2
		if offset+strLen <= len(bz) {
			result.BytecodeHash = string(bz[offset : offset+strLen])
			offset += strLen
		}
	}

	// Compiler version
	if offset+2 <= len(bz) {
		strLen := int(binary.BigEndian.Uint16(bz[offset : offset+2]))
		offset += 2
		if offset+strLen <= len(bz) {
			result.CompilerVersion = string(bz[offset : offset+strLen])
			offset += strLen
		}
	}

	// Verified by
	if offset+2 <= len(bz) {
		strLen := int(binary.BigEndian.Uint16(bz[offset : offset+2]))
		offset += 2
		if offset+strLen <= len(bz) {
			result.VerifiedBy = string(bz[offset : offset+strLen])
			offset += strLen
		}
	}

	// Certification ID
	if offset+2 <= len(bz) {
		strLen := int(binary.BigEndian.Uint16(bz[offset : offset+2]))
		offset += 2
		if offset+strLen <= len(bz) {
			result.CertificationID = string(bz[offset : offset+strLen])
		}
	}

	return result, true
}

// IsVerified checks if a contract is currently verified
func (k Keeper) IsVerified(ctx sdk.Context, contractAddr string) bool {
	result, found := k.GetVerificationResult(ctx, contractAddr)
	if !found {
		return false
	}

	// Check status and expiration
	if result.Status != VerificationStatusVerified {
		return false
	}

	if ctx.BlockTime().After(result.ExpirationTime) {
		return false
	}

	return true
}

// GetVerificationLevel returns the verification level of a contract
func (k Keeper) GetVerificationLevel(ctx sdk.Context, contractAddr string) VerificationLevel {
	result, found := k.GetVerificationResult(ctx, contractAddr)
	if !found || result.Status != VerificationStatusVerified {
		return 0
	}
	return result.Level
}

// RevokeVerification revokes a contract's verification
func (k Keeper) RevokeVerification(ctx sdk.Context, contractAddr string, revoker string, reason string) error {
	result, found := k.GetVerificationResult(ctx, contractAddr)
	if !found {
		return types.ErrContractNotFound
	}

	// Only governance or original verifier can revoke
	if revoker != k.authority && revoker != result.VerifiedBy {
		return types.ErrNotContractAdmin
	}

	result.Status = VerificationStatusRevoked
	k.storeVerificationResult(ctx, result)

	// Emit revocation event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"verification_revoked",
			sdk.NewAttribute("contract_address", contractAddr),
			sdk.NewAttribute("revoked_by", revoker),
			sdk.NewAttribute("reason", reason),
		),
	)

	return nil
}

// VerifyBytecodeMatch verifies that compiled bytecode matches deployed bytecode
func (k Keeper) VerifyBytecodeMatch(compiledBytecode, deployedBytecode []byte) (bool, error) {
	if len(compiledBytecode) == 0 || len(deployedBytecode) == 0 {
		return false, fmt.Errorf("bytecode cannot be empty")
	}

	// Remove metadata hash from bytecode (last 43 bytes for Solidity >= 0.6.0)
	compiledClean := k.removeMetadataFromBytecode(compiledBytecode)
	deployedClean := k.removeMetadataFromBytecode(deployedBytecode)

	return bytes.Equal(compiledClean, deployedClean), nil
}

// removeMetadataFromBytecode removes the Solidity metadata hash from bytecode
func (k Keeper) removeMetadataFromBytecode(bytecode []byte) []byte {
	if len(bytecode) < 2 {
		return bytecode
	}

	// Check for metadata pattern
	// The last two bytes indicate the length of the metadata
	metadataLength := int(bytecode[len(bytecode)-2])<<8 + int(bytecode[len(bytecode)-1])

	if metadataLength > 0 && metadataLength < len(bytecode)-2 {
		// Verify it looks like CBOR metadata (starts with 0xa2)
		metadataStart := len(bytecode) - 2 - metadataLength
		if metadataStart > 0 && bytecode[metadataStart] == 0xa2 {
			return bytecode[:metadataStart]
		}
	}

	return bytecode
}

// AnalyzeBytecode performs analysis on contract bytecode
func (k Keeper) AnalyzeBytecode(bytecode []byte) (*BytecodeAnalysis, error) {
	if len(bytecode) == 0 {
		return nil, fmt.Errorf("bytecode cannot be empty")
	}

	analysis := &BytecodeAnalysis{
		Size:                len(bytecode),
		EventSignatures:     make([]string, 0),
		FunctionSelectors:   make([]string, 0),
		ImmutableReferences: make(map[string][]int),
	}

	// Extract function selectors (first 4 bytes of keccak256 hash)
	// Look for PUSH4 opcode (0x63) followed by 4 bytes
	for i := 0; i < len(bytecode)-5; i++ {
		if bytecode[i] == 0x63 { // PUSH4
			selector := hex.EncodeToString(bytecode[i+1 : i+5])
			analysis.FunctionSelectors = append(analysis.FunctionSelectors, "0x"+selector)
		}
	}

	// Check for constructor (CODECOPY pattern)
	for i := 0; i < len(bytecode)-1; i++ {
		if bytecode[i] == 0x39 { // CODECOPY
			analysis.HasConstructor = true
			break
		}
	}

	// Check for fallback/receive (specific patterns)
	for i := 0; i < len(bytecode)-2; i++ {
		if bytecode[i] == 0x36 && bytecode[i+1] == 0x60 && bytecode[i+2] == 0x00 {
			// CALLDATASIZE PUSH1 0x00 pattern often indicates fallback
			analysis.HasFallback = true
		}
	}

	// Count external calls (CALL, DELEGATECALL, STATICCALL)
	for _, b := range bytecode {
		switch b {
		case 0xf1, 0xf2, 0xf4, 0xfa: // CALL, CALLCODE, DELEGATECALL, STATICCALL
			analysis.ExternalCalls++
		case 0x55: // SSTORE
			analysis.StorageSlots++
		}
	}

	// Extract metadata
	metadataClean := k.removeMetadataFromBytecode(bytecode)
	if len(metadataClean) < len(bytecode) {
		analysis.Metadata = bytecode[len(metadataClean):]
		metadataHash := sha256.Sum256(analysis.Metadata)
		analysis.MetadataHash = hex.EncodeToString(metadataHash[:])
	}

	return analysis, nil
}

// BatchVerifyContracts verifies multiple contracts in a batch
func (k Keeper) BatchVerifyContracts(ctx sdk.Context, requests []VerificationRequest) ([]*VerificationResult, error) {
	results := make([]*VerificationResult, 0, len(requests))

	for _, req := range requests {
		result, err := k.VerifyContract(ctx, req)
		if err != nil {
			// Continue with other verifications even if one fails
			result = &VerificationResult{
				ContractAddress: req.ContractAddress,
				Status:          VerificationStatusFailed,
				Issues: []VerificationIssue{{
					Severity:    "critical",
					Code:        "VERIFICATION_ERROR",
					Description: err.Error(),
				}},
			}
		}
		results = append(results, result)
	}

	return results, nil
}

// IterateVerificationResults iterates over all verification results
func (k Keeper) IterateVerificationResults(ctx sdk.Context, fn func(result *VerificationResult) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.VerificationResultKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		// Extract contract address from key
		key := iterator.Key()
		if len(key) <= len(types.VerificationResultKeyPrefix) {
			continue
		}
		contractAddr := string(key[len(types.VerificationResultKeyPrefix):])

		result, found := k.GetVerificationResult(ctx, contractAddr)
		if !found {
			continue
		}

		if fn(result) {
			break
		}
	}
}

// GetVerifiedContractsCount returns the count of verified contracts
func (k Keeper) GetVerifiedContractsCount(ctx sdk.Context) uint64 {
	var count uint64
	k.IterateVerificationResults(ctx, func(result *VerificationResult) bool {
		if result.Status == VerificationStatusVerified && ctx.BlockTime().Before(result.ExpirationTime) {
			count++
		}
		return false
	})
	return count
}

// GetExpiredVerifications returns contracts with expired verifications
func (k Keeper) GetExpiredVerifications(ctx sdk.Context) []*VerificationResult {
	expired := make([]*VerificationResult, 0)

	k.IterateVerificationResults(ctx, func(result *VerificationResult) bool {
		if result.Status == VerificationStatusVerified && ctx.BlockTime().After(result.ExpirationTime) {
			expired = append(expired, result)
		}
		return false
	})

	return expired
}

// RefreshExpiredVerifications marks expired verifications as expired
func (k Keeper) RefreshExpiredVerifications(ctx sdk.Context) int {
	expired := k.GetExpiredVerifications(ctx)

	for _, result := range expired {
		result.Status = VerificationStatusExpired
		k.storeVerificationResult(ctx, result)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"verification_expired",
				sdk.NewAttribute("contract_address", result.ContractAddress),
			),
		)
	}

	return len(expired)
}

// GetVerificationsByStatus returns all verifications with a specific status
func (k Keeper) GetVerificationsByStatus(ctx sdk.Context, status VerificationStatus) []*VerificationResult {
	results := make([]*VerificationResult, 0)

	k.IterateVerificationResults(ctx, func(result *VerificationResult) bool {
		if result.Status == status {
			results = append(results, result)
		}
		return false
	})

	return results
}

// GetVerificationsByLevel returns all verified contracts at or above a specific level
func (k Keeper) GetVerificationsByLevel(ctx sdk.Context, minLevel VerificationLevel) []*VerificationResult {
	results := make([]*VerificationResult, 0)

	k.IterateVerificationResults(ctx, func(result *VerificationResult) bool {
		if result.Status == VerificationStatusVerified && result.Level >= minLevel {
			results = append(results, result)
		}
		return false
	})

	return results
}

// GetCertifiedContracts returns all contracts with certified verification level
func (k Keeper) GetCertifiedContracts(ctx sdk.Context) []*VerificationResult {
	return k.GetVerificationsByLevel(ctx, VerificationLevelCertified)
}
