package keeper

import (
	"fmt"
	"time"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/confidencescore/params"
	"github.com/aequitas/aura/chain/x/confidencescore/types"
)

// IRRegistry defines the interface for interacting with the inclusionroutines module
type IRRegistry interface {
	GetIRPrerequisites(irID string) ([]string, error)
	IsIRActive(irID string) bool
	GetIRScore(irID string) (uint64, error)
	GetIRArena(irID string) (string, error)
}

// Keeper manages the state of the confidencescore module using persistent KV store.
// All state is stored deterministically in the KV store to ensure consensus safety.
type Keeper struct {
	storeService store.KVStoreService
	cdc          codec.BinaryCodec
	paramsStore  *params.Store
	irRegistry   IRRegistry
	authority    string // governance module account address
	logger       log.Logger
}

// NewKeeper creates a new Keeper instance with persistent KV store.
// All state is persisted to the KV store - no in-memory maps are used.
func NewKeeper(
	storeService store.KVStoreService,
	cdc codec.BinaryCodec,
	paramsStore *params.Store,
	authority string,
	logger log.Logger,
) *Keeper {
	if paramsStore == nil {
		paramsStore = params.NewStore(types.DefaultParams())
	}

	return &Keeper{
		storeService: storeService,
		cdc:          cdc,
		paramsStore:  paramsStore,
		authority:    authority,
		logger:       logger,
	}
}

// GetAuthority returns the governance module account address
func (k *Keeper) GetAuthority() string {
	return k.authority
}

// SetIRRegistry sets the IR registry for validation
func (k *Keeper) SetIRRegistry(registry IRRegistry) {
	k.irRegistry = registry
}

// GetParams returns the current module parameters
func (k *Keeper) GetParams() types.Params {
	if k.paramsStore != nil {
		return k.paramsStore.GetParams()
	}
	return types.DefaultParams()
}

// SetParams sets new module parameters
func (k *Keeper) SetParams(params types.Params) error {
	if k.paramsStore == nil {
		return types.ErrUnauthorized
	}
	return k.paramsStore.SetParams(params)
}

// ============================
// USER RECORD MANAGEMENT
// ============================

// GetUserRecord retrieves a user's confidence record from KV store
func (k *Keeper) GetUserRecord(ctx sdk.Context, walletAddr string) (types.UserConfidenceRecord, bool) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get([]byte(types.UserRecordStoreKey(walletAddr)))
	if err != nil || bz == nil {
		return types.UserConfidenceRecord{}, false
	}

	var record types.UserConfidenceRecord
	if err := k.cdc.Unmarshal(bz, &record); err != nil {
		return types.UserConfidenceRecord{}, false
	}

	return record, true
}

// SetUserRecord stores a user's confidence record to KV store
func (k *Keeper) SetUserRecord(ctx sdk.Context, record types.UserConfidenceRecord) error {
	if record.WalletAddress == "" {
		return types.ErrInvalidWalletAddress
	}

	// Update timestamps using block context
	record.LastUpdatedHeight = uint64(ctx.BlockHeight())
	record.LastUpdated = timestampFromTime(ctx.BlockTime())

	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(&record)
	if err != nil {
		return fmt.Errorf("failed to marshal user record: %w", err)
	}

	if err := store.Set([]byte(types.UserRecordStoreKey(record.WalletAddress)), bz); err != nil {
		return fmt.Errorf("failed to store user record: %w", err)
	}

	return nil
}

// GetOrCreateUserRecord gets existing record or creates new one
func (k *Keeper) GetOrCreateUserRecord(ctx sdk.Context, walletAddr string) types.UserConfidenceRecord {
	record, ok := k.GetUserRecord(ctx, walletAddr)
	if !ok {
		record = types.UserConfidenceRecord{
			WalletAddress: walletAddr,
			TotalScore:    0,
			CompletedIrs:  []*types.IRCompletion{},
			HasAnchor:     false,
			ArenaScores:   make(map[string]*types.ArenaScore),
			Status:        types.VerificationStatusUnverified,
		}
	}
	return record
}

// ============================
// IR COMPLETION MANAGEMENT
// ============================

// GetIRCompletion retrieves a specific IR completion for a user from KV store
func (k *Keeper) GetIRCompletion(ctx sdk.Context, walletAddr, irID string) (types.IRCompletion, bool) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get([]byte(types.IRCompletionStoreKey(walletAddr, irID)))
	if err != nil || bz == nil {
		return types.IRCompletion{}, false
	}

	var completion types.IRCompletion
	if err := k.cdc.Unmarshal(bz, &completion); err != nil {
		return types.IRCompletion{}, false
	}

	return completion, true
}

// SetIRCompletion stores an IR completion to KV store
func (k *Keeper) SetIRCompletion(ctx sdk.Context, walletAddr string, completion types.IRCompletion) error {
	if walletAddr == "" {
		return types.ErrInvalidWalletAddress
	}
	if completion.IrId == "" {
		return types.ErrInvalidIRID
	}

	// Update timestamps using block context
	completion.CompletedHeight = uint64(ctx.BlockHeight())
	completion.CompletedAt = timestampFromTime(ctx.BlockTime())

	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(&completion)
	if err != nil {
		return fmt.Errorf("failed to marshal IR completion: %w", err)
	}

	if err := store.Set([]byte(types.IRCompletionStoreKey(walletAddr, completion.IrId)), bz); err != nil {
		return fmt.Errorf("failed to store IR completion: %w", err)
	}

	return nil
}

// ============================
// SCORE HISTORY MANAGEMENT
// ============================

// GetScoreHistory retrieves score change history for a user from KV store
func (k *Keeper) GetScoreHistory(ctx sdk.Context, walletAddr string, fromHeight, toHeight uint64, limit int) []types.ScoreChange {
	store := k.storeService.OpenKVStore(ctx)

	// Use iterator to scan score history
	prefix := []byte(types.ScoreHistoryStoreKeyPrefix + walletAddr + "/")
	iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return []types.ScoreChange{}
	}
	defer iterator.Close()

	history := []types.ScoreChange{}
	for ; iterator.Valid(); iterator.Next() {
		var change types.ScoreChange
		if err := k.cdc.Unmarshal(iterator.Value(), &change); err != nil {
			continue
		}

		// Filter by height range if specified
		if (fromHeight == 0 || change.BlockHeight >= fromHeight) &&
			(toHeight == 0 || change.BlockHeight <= toHeight) {
			history = append(history, change)
		}

		if limit > 0 && len(history) >= limit {
			break
		}
	}

	return history
}

// AddScoreChange adds a score change to history in KV store
func (k *Keeper) AddScoreChange(ctx sdk.Context, change types.ScoreChange) error {
	// Update timestamps using block context
	change.BlockHeight = uint64(ctx.BlockHeight())
	change.Timestamp = timestampFromTime(ctx.BlockTime())

	store := k.storeService.OpenKVStore(ctx)

	// Use unique key combining wallet, height, and tx hash
	key := fmt.Sprintf("%s%s/%d/%s",
		types.ScoreHistoryStoreKeyPrefix,
		// TODO: WalletAddress field not in proto - needs to be tracked separately
		"unknown",
		change.BlockHeight,
		change.TxHash)

	bz, err := k.cdc.Marshal(&change)
	if err != nil {
		return fmt.Errorf("failed to marshal score change: %w", err)
	}

	if err := store.Set([]byte(key), bz); err != nil {
		return fmt.Errorf("failed to store score change: %w", err)
	}

	return nil
}

// ============================
// SLASH RECORD MANAGEMENT
// ============================

// GetSlashRecords retrieves all slash records for a user from KV store
func (k *Keeper) GetSlashRecords(ctx sdk.Context, walletAddr string) []types.SlashRecord {
	store := k.storeService.OpenKVStore(ctx)

	// Use iterator to scan slash records
	prefix := []byte(types.SlashRecordStoreKeyPrefix + walletAddr + "/")
	iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return []types.SlashRecord{}
	}
	defer iterator.Close()

	records := []types.SlashRecord{}
	for ; iterator.Valid(); iterator.Next() {
		var record types.SlashRecord
		if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
			continue
		}
		records = append(records, record)
	}

	return records
}

// AddSlashRecord adds a slash record to KV store
func (k *Keeper) AddSlashRecord(ctx sdk.Context, record types.SlashRecord) error {
	if record.WalletAddress == "" {
		return types.ErrInvalidWalletAddress
	}

	// Update timestamps using block context
	record.SlashHeight = uint64(ctx.BlockHeight())
	record.SlashTime = timestampFromTime(ctx.BlockTime())

	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(&record)
	if err != nil {
		return fmt.Errorf("failed to marshal slash record: %w", err)
	}

	key := types.SlashRecordStoreKey(record.WalletAddress, record.SlashTxHash)
	if err := store.Set([]byte(key), bz); err != nil {
		return fmt.Errorf("failed to store slash record: %w", err)
	}

	return nil
}

// GetSlashRecord retrieves a specific slash record from KV store
func (k *Keeper) GetSlashRecord(ctx sdk.Context, walletAddr, slashTxHash string) (types.SlashRecord, bool) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get([]byte(types.SlashRecordStoreKey(walletAddr, slashTxHash)))
	if err != nil || bz == nil {
		return types.SlashRecord{}, false
	}

	var record types.SlashRecord
	if err := k.cdc.Unmarshal(bz, &record); err != nil {
		return types.SlashRecord{}, false
	}

	return record, true
}

// UpdateSlashRecord updates an existing slash record in KV store
func (k *Keeper) UpdateSlashRecord(ctx sdk.Context, walletAddr string, record types.SlashRecord) error {
	_, ok := k.GetSlashRecord(ctx, walletAddr, record.SlashTxHash)
	if !ok {
		return types.ErrSlashNotFound
	}

	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(&record)
	if err != nil {
		return fmt.Errorf("failed to marshal slash record: %w", err)
	}

	key := types.SlashRecordStoreKey(walletAddr, record.SlashTxHash)
	if err := store.Set([]byte(key), bz); err != nil {
		return fmt.Errorf("failed to update slash record: %w", err)
	}

	return nil
}

// ============================
// PROOF HASH MANAGEMENT (Replay Prevention)
// ============================

// HasProofHash checks if a proof hash exists in KV store (for replay prevention)
func (k *Keeper) HasProofHash(ctx sdk.Context, walletAddr string, proofHash []byte) bool {
	store := k.storeService.OpenKVStore(ctx)
	key := types.ProofHashStoreKey(walletAddr, proofHash)
	exists, err := store.Has([]byte(key))
	if err != nil {
		return false
	}
	return exists
}

// SetProofHash stores a proof hash in KV store (for replay prevention)
func (k *Keeper) SetProofHash(ctx sdk.Context, walletAddr string, proofHash []byte) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.ProofHashStoreKey(walletAddr, proofHash)

	// Store a simple marker (empty byte array)
	if err := store.Set([]byte(key), []byte{0x01}); err != nil {
		return fmt.Errorf("failed to store proof hash: %w", err)
	}

	return nil
}

// ============================
// RATE LIMIT MANAGEMENT
// ============================

// GetRateLimit retrieves rate limit counter from KV store
func (k *Keeper) GetRateLimit(ctx sdk.Context, walletAddr, action, timeWindow string) int32 {
	store := k.storeService.OpenKVStore(ctx)
	key := types.RateLimitStoreKey(walletAddr, fmt.Sprintf("%s:%s", action, timeWindow))

	bz, err := store.Get([]byte(key))
	if err != nil || bz == nil {
		return 0
	}

	// Simple counter stored as int32
	if len(bz) != 4 {
		return 0
	}

	return int32(bz[0]) | int32(bz[1])<<8 | int32(bz[2])<<16 | int32(bz[3])<<24
}

// IncrementRateLimit increments a rate limit counter in KV store
func (k *Keeper) IncrementRateLimit(ctx sdk.Context, walletAddr, action, timeWindow string) error {
	current := k.GetRateLimit(ctx, walletAddr, action, timeWindow)
	newCount := current + 1

	store := k.storeService.OpenKVStore(ctx)
	key := types.RateLimitStoreKey(walletAddr, fmt.Sprintf("%s:%s", action, timeWindow))

	// Encode as 4-byte little-endian
	bz := []byte{
		byte(newCount),
		byte(newCount >> 8),
		byte(newCount >> 16),
		byte(newCount >> 24),
	}

	if err := store.Set([]byte(key), bz); err != nil {
		return fmt.Errorf("failed to increment rate limit: %w", err)
	}

	return nil
}

// ============================
// GENESIS MANAGEMENT
// ============================
// Note: InitGenesis and ExportGenesis are implemented in genesis.go

// ============================
// QUERY HELPERS
// ============================

// IsVerified checks if a user has verified status (CS >= threshold)
func (k *Keeper) IsVerified(ctx sdk.Context, walletAddr string) bool {
	record, ok := k.GetUserRecord(ctx, walletAddr)
	if !ok {
		return false
	}

	params := k.GetParams()
	return record.TotalScore >= params.VerificationThreshold &&
		record.Status == types.VerificationStatusVerified
}

// HasCompletedIR checks if a user has completed a specific IR
func (k *Keeper) HasCompletedIR(ctx sdk.Context, walletAddr, irID string) bool {
	_, ok := k.GetIRCompletion(ctx, walletAddr, irID)
	return ok
}

// GetArenaScore retrieves the score for a specific arena
func (k *Keeper) GetArenaScore(ctx sdk.Context, walletAddr, arena string) (uint64, error) {
	record, ok := k.GetUserRecord(ctx, walletAddr)
	if !ok {
		return 0, types.ErrUserRecordNotFound
	}

	arenaScore, ok := record.ArenaScores[arena]
	if !ok {
		return 0, nil
	}

	return arenaScore.TotalScore, nil
}

// GetUserScore retrieves the total confidence score for a user
func (k *Keeper) GetUserScore(ctx sdk.Context, walletAddr string) (uint64, bool) {
	record, ok := k.GetUserRecord(ctx, walletAddr)
	if !ok {
		return 0, false
	}
	return record.TotalScore, true
}

// GetAnchorInfo retrieves anchor completion information for a user
func (k *Keeper) GetAnchorInfo(ctx sdk.Context, walletAddr string) (interface{}, bool) {
	record, ok := k.GetUserRecord(ctx, walletAddr)
	if !ok {
		return nil, false
	}

	if record.AnchorInfo == nil {
		return nil, false
	}

	return record.AnchorInfo, true
}

// ListVerifiedUsers returns a list of verified users
func (k *Keeper) ListVerifiedUsers(ctx sdk.Context, minScore uint64, limit int) ([]string, []uint64) {
	store := k.storeService.OpenKVStore(ctx)
	params := k.GetParams()

	if minScore == 0 {
		minScore = params.VerificationThreshold
	}

	wallets := []string{}
	scores := []uint64{}

	// Iterate over all user records
	prefix := []byte(types.UserRecordStoreKeyPrefix)
	iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return wallets, scores
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var record types.UserConfidenceRecord
		if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
			continue
		}

		if record.TotalScore >= minScore && record.Status == types.VerificationStatusVerified {
			wallets = append(wallets, record.WalletAddress)
			scores = append(scores, record.TotalScore)

			if limit > 0 && len(wallets) >= limit {
				break
			}
		}
	}

	return wallets, scores
}

// timestampFromTime converts time.Time to protobuf timestamp
func timestampFromTime(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}
