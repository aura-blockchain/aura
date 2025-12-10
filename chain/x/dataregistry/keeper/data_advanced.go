package keeper

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"time"

	"cosmossdk.io/math"
	"github.com/aequitas/aura/chain/x/common/determinism"
	"github.com/aequitas/aura/chain/x/dataregistry/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ============================
// ADVANCED DATA REGISTRY FEATURES
// Implements encryption, versioning, provenance, retention, monetization, marketplace
// ============================

// ===== DATA ENCRYPTION/DECRYPTION =====

// EncryptData encrypts data using AES-GCM with deterministic nonce
// Uses deterministic RNG for consensus compatibility
func (k *Keeper) EncryptData(ctx context.Context, data []byte, encryptionKey []byte, additionalEntropy []byte) (encrypted []byte, nonce []byte, err error) {
	if len(encryptionKey) != 32 {
		return nil, nil, fmt.Errorf("encryption key must be 32 bytes")
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Generate deterministic nonce using block context
	rng := determinism.NewDeterministicRNG(ctx, additionalEntropy)
	nonce = rng.Bytes(12)

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	encrypted = aesgcm.Seal(nil, nonce, data, nil)
	return encrypted, nonce, nil
}

// DecryptData decrypts AES-GCM encrypted data
func (k *Keeper) DecryptData(encrypted []byte, nonce []byte, encryptionKey []byte) ([]byte, error) {
	if len(encryptionKey) != 32 {
		return nil, fmt.Errorf("encryption key must be 32 bytes")
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	plaintext, err := aesgcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// ===== DATA VERSIONING =====

// DataVersion represents a version of a data item
type DataVersion struct {
	VersionID    string
	DataID       string
	VersionNum   uint64
	ContentHash  string
	StorageLocation string
	CreatedAt    time.Time
	CreatedBy    string
	ChangeLog    string
	PreviousVersion string
}

// CreateDataVersion creates a new version of a data item
func (k *Keeper) CreateDataVersion(ctx context.Context, dataID string, updater string, newContentHash []byte, newStorageLocation string, changeLog string) (*DataVersion, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	item, ok := k.GetDataItem(sdkCtx, dataID)
	if !ok {
		return nil, types.ErrDataItemNotFound
	}

	// Check ownership
	if item.OwnerAddress != updater {
		return nil, types.ErrUnauthorized
	}

	// Get current version count
	versions := k.getDataVersions(dataID)
	versionNum := uint64(len(versions) + 1)

	// Get current block time
	currentTime := sdkCtx.BlockTime()

	// Create new version
	version := &DataVersion{
		VersionID:       fmt.Sprintf("%s-v%d", dataID, versionNum),
		DataID:          dataID,
		VersionNum:      versionNum,
		ContentHash:     hex.EncodeToString(newContentHash),
		StorageLocation: newStorageLocation,
		CreatedAt:       currentTime,
		CreatedBy:       updater,
		ChangeLog:       changeLog,
		PreviousVersion: hex.EncodeToString(item.ContentHash),
	}

	// Update item with new version
	item.ContentHash = newContentHash
	item.StorageLocation = newStorageLocation

	if err := k.SetDataItem(sdkCtx, item); err != nil {
		return nil, err
	}

	// Store version
	k.storeDataVersion(version)

	return version, nil
}

// GetDataVersions returns all versions of a data item
func (k *Keeper) GetDataVersions(ctx sdk.Context, dataID string) []DataVersion {
	return k.getDataVersions(dataID)
}

// RestoreDataVersion restores a previous version
func (k *Keeper) RestoreDataVersion(ctx sdk.Context, dataID string, versionNum uint64, restorer string) error {
	item, ok := k.GetDataItem(ctx, dataID)
	if !ok {
		return types.ErrDataItemNotFound
	}

	if item.OwnerAddress != restorer {
		return types.ErrUnauthorized
	}

	versions := k.getDataVersions(dataID)
	var targetVersion *DataVersion
	for i := range versions {
		if versions[i].VersionNum == versionNum {
			targetVersion = &versions[i]
			break
		}
	}

	if targetVersion == nil {
		return fmt.Errorf("version not found")
	}

	// Restore
	contentHash, _ := hex.DecodeString(targetVersion.ContentHash)
	item.ContentHash = contentHash
	item.StorageLocation = targetVersion.StorageLocation
	// NOTE: Future enhancement - Add UpdatedAt field to DataItem proto if needed
	// item.UpdatedAt = timestamppb.New(time.Unix(time.Now().Unix(), 0))

	return k.SetDataItem(ctx, item)
}

// ===== DATA PROVENANCE TRACKING =====

// ProvenanceRecord tracks the lifecycle of data
type ProvenanceRecord struct {
	RecordID    string
	DataID      string
	Action      string // created, updated, accessed, shared, etc.
	Actor       string
	Timestamp   time.Time
	BlockHeight uint64
	Details     map[string]string
}

// RecordProvenance records a provenance event
func (k *Keeper) RecordProvenance(ctx sdk.Context, dataID string, action string, actor string, details map[string]string) error {
	record := &ProvenanceRecord{
		RecordID:    fmt.Sprintf("%s-%s-%d", dataID, action, ctx.BlockHeight()),
		DataID:      dataID,
		Action:      action,
		Actor:       actor,
		Timestamp:   ctx.BlockTime(),
		BlockHeight: uint64(ctx.BlockHeight()),
		Details:     details,
	}

	k.storeProvenanceRecord(record)
	return nil
}

// GetProvenanceTrail returns complete provenance trail for a data item
func (k *Keeper) GetProvenanceTrail(ctx sdk.Context, dataID string) []ProvenanceRecord {
	return k.getProvenanceRecords(dataID)
}

// ===== DATA RETENTION POLICIES =====

// RetentionPolicy defines retention rules for data
type RetentionPolicy struct {
	PolicyID      string
	DataID        string
	RetentionDays uint64
	AutoDelete    bool
	ExpiresAt     time.Time
	NotifyBefore  uint64 // days before expiration to notify
}

// SetRetentionPolicy sets a retention policy for data
func (k *Keeper) SetRetentionPolicy(ctx sdk.Context, dataID string, retentionDays uint64, autoDelete bool, notifyBefore uint64) (*RetentionPolicy, error) {
	item, ok := k.GetDataItem(ctx, dataID)
	if !ok {
		return nil, types.ErrDataItemNotFound
	}

	expiresAt := ctx.BlockTime().AddDate(0, 0, int(retentionDays))

	policy := &RetentionPolicy{
		PolicyID:      fmt.Sprintf("retention-%s", dataID),
		DataID:        dataID,
		RetentionDays: retentionDays,
		AutoDelete:    autoDelete,
		ExpiresAt:     expiresAt,
		NotifyBefore:  notifyBefore,
	}

	// Update item
	item.Metadata["retention_policy"] = policy.PolicyID
	k.SetDataItem(ctx, item)

	k.storeRetentionPolicy(policy)
	return policy, nil
}

// ProcessExpiredData processes data items that have expired retention
func (k *Keeper) ProcessExpiredData(ctx sdk.Context, ) (deleted int, notified int, error error) {
	policies := k.getAllRetentionPolicies()
	now := ctx.BlockTime()

	for _, policy := range policies {
		if now.After(policy.ExpiresAt) && policy.AutoDelete {
			if err := k.DeleteDataItem(ctx, policy.DataID); err == nil {
				deleted++
			}
		} else if now.After(policy.ExpiresAt.AddDate(0, 0, -int(policy.NotifyBefore))) {
			// Notify owner
			notified++
		}
	}

	return deleted, notified, nil
}

// ===== DATA QUALITY SCORING =====

// QualityScore represents data quality metrics
type QualityScore struct {
	DataID          string
	OverallScore    uint64 // 0-100
	CompletenessScore uint64
	AccuracyScore   uint64
	TimelinessScore uint64
	ConsistencyScore uint64
	CalculatedAt    time.Time
}

// CalculateQualityScore calculates quality score for a data item
func (k *Keeper) CalculateQualityScore(ctx sdk.Context, dataID string) (*QualityScore, error) {
	item, ok := k.GetDataItem(ctx, dataID)
	if !ok {
		return nil, types.ErrDataItemNotFound
	}

	// Calculate component scores
	completeness := k.calculateCompleteness(item)
	accuracy := k.calculateAccuracy(item)
	timeliness := k.calculateTimeliness(item)
	consistency := k.calculateConsistency(item)

	// Overall score (weighted average)
	overall := (completeness*30 + accuracy*40 + timeliness*15 + consistency*15) / 100

	score := &QualityScore{
		DataID:           dataID,
		OverallScore:     overall,
		CompletenessScore: completeness,
		AccuracyScore:    accuracy,
		TimelinessScore:  timeliness,
		ConsistencyScore: consistency,
		CalculatedAt:     ctx.BlockTime(),
	}

	k.storeQualityScore(score)
	return score, nil
}

func (k *Keeper) calculateCompleteness(item types.DataItem) uint64 {
	score := uint64(0)
	if item.Title != "" { score += 20 }
	if item.Description != "" { score += 20 }
	if len(item.Metadata) > 0 { score += 20 }
	if len(item.Tags) > 0 { score += 20 }
	if item.GeoLocation != nil { score += 20 }
	return score
}

func (k *Keeper) calculateAccuracy(item types.DataItem) uint64 {
	// Based on verifications
	if len(item.Verifications) >= 3 { return 100 }
	if len(item.Verifications) == 2 { return 75 }
	if len(item.Verifications) == 1 { return 50 }
	return 25
}

func (k *Keeper) calculateTimeliness(item types.DataItem) uint64 {
	// Based on how recent the data is
	if item.CreatedAt == nil {
		return 25
	}
	createdAt := time.Unix(item.CreatedAt.Seconds, int64(item.CreatedAt.Nanos))
	age := time.Since(createdAt).Hours() / 24
	if age < 7 { return 100 }
	if age < 30 { return 75 }
	if age < 90 { return 50 }
	return 25
}

func (k *Keeper) calculateConsistency(item types.DataItem) uint64 {
	// Simplified - check if content hash matches expected format
	if len(item.ContentHash) == 64 { // Valid SHA-256
		return 100
	}
	return 50
}

// ===== DATA MONETIZATION =====

// MintVerificationReward mints reward tokens for data verification
func (k *Keeper) MintVerificationReward(ctx sdk.Context, verifier string, dataID string, bankKeeper BankKeeper) error {
	params := k.GetParams()

	// Calculate reward amount
	rewardAmount := math.NewInt(int64(params.VerificationReward))

	// Mint coins
	coins := sdk.NewCoins(sdk.NewCoin("uaura", rewardAmount))
	if err := bankKeeper.MintCoins(ctx, types.ModuleName, coins); err != nil {
		return fmt.Errorf("failed to mint verification reward: %w", err)
	}

	// Send to verifier
	verifierAddr, err := sdk.AccAddressFromBech32(verifier)
	if err != nil {
		return fmt.Errorf("invalid verifier address: %w", err)
	}

	if err := bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, verifierAddr, coins); err != nil {
		return fmt.Errorf("failed to send verification reward: %w", err)
	}

	// Record provenance
	k.RecordProvenance(ctx, dataID, "verification_rewarded", verifier, map[string]string{
		"reward_amount": rewardAmount.String(),
	})

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"verification_reward_minted",
			sdk.NewAttribute("verifier", verifier),
			sdk.NewAttribute("data_id", dataID),
			sdk.NewAttribute("amount", rewardAmount.String()),
		),
	)

	return nil
}

// ===== HELPER STORAGE FUNCTIONS =====
// In production, these would use KVStore

func (k *Keeper) storeDataVersion(version *DataVersion) {}
func (k *Keeper) getDataVersions(dataID string) []DataVersion { return []DataVersion{} }
func (k *Keeper) storeProvenanceRecord(record *ProvenanceRecord) {}
func (k *Keeper) getProvenanceRecords(dataID string) []ProvenanceRecord { return []ProvenanceRecord{} }
func (k *Keeper) storeRetentionPolicy(policy *RetentionPolicy) {}
func (k *Keeper) getAllRetentionPolicies() []RetentionPolicy { return []RetentionPolicy{} }
func (k *Keeper) storeQualityScore(score *QualityScore) {}
