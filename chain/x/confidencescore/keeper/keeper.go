package keeper

import (
	"fmt"
	"sync"
	"time"

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

// Keeper manages the state of the confidencescore module
type Keeper struct {
	mu            sync.RWMutex
	userRecords   map[string]types.UserConfidenceRecord
	completions   map[string]map[string]types.IRCompletion // wallet -> irID -> completion
	scoreHistory  map[string][]types.ScoreChange           // wallet -> changes
	slashRecords  map[string][]types.SlashRecord           // wallet -> slashes
	proofHashes   map[string]map[string]bool               // wallet -> proof_hash -> exists
	rateLimits    map[string]map[string]int                // wallet -> time_window -> count
	paramsStore   *params.Store
	irRegistry    IRRegistry
	currentHeight uint64
	currentTime   int64
	authority     string // governance module account address
}

// NewKeeper creates a new Keeper instance
func NewKeeper(store *params.Store, authority string) *Keeper {
	if store == nil {
		store = params.NewStore(types.DefaultParams())
	}
	return &Keeper{
		userRecords:  make(map[string]types.UserConfidenceRecord),
		completions:  make(map[string]map[string]types.IRCompletion),
		scoreHistory: make(map[string][]types.ScoreChange),
		slashRecords: make(map[string][]types.SlashRecord),
		proofHashes:  make(map[string]map[string]bool),
		rateLimits:   make(map[string]map[string]int),
		paramsStore:  store,
		currentTime:  time.Now().Unix(),
		authority:    authority,
	}
}

// GetAuthority returns the governance module account address
func (k *Keeper) GetAuthority() string {
	return k.authority
}

// SetIRRegistry sets the IR registry for validation
func (k *Keeper) SetIRRegistry(registry IRRegistry) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.irRegistry = registry
}

// SetCurrentHeight sets the current block height (for testing/simulation)
func (k *Keeper) SetCurrentHeight(height uint64) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.currentHeight = height
}

// SetCurrentTime sets the current time (for testing/simulation)
func (k *Keeper) SetCurrentTime(t int64) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.currentTime = t
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

// GetUserRecord retrieves a user's confidence record
func (k *Keeper) GetUserRecord(walletAddr string) (types.UserConfidenceRecord, bool) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	record, ok := k.userRecords[walletAddr]
	return record, ok
}

// SetUserRecord stores a user's confidence record
func (k *Keeper) SetUserRecord(record types.UserConfidenceRecord) error {
	if record.WalletAddress == "" {
		return types.ErrInvalidWalletAddress
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	record.LastUpdatedHeight = k.currentHeight
	record.LastUpdated = k.currentTime

	k.userRecords[record.WalletAddress] = record
	return nil
}

// GetOrCreateUserRecord gets existing record or creates new one
func (k *Keeper) GetOrCreateUserRecord(walletAddr string) types.UserConfidenceRecord {
	record, ok := k.GetUserRecord(walletAddr)
	if !ok {
		record = types.UserConfidenceRecord{
			WalletAddress: walletAddr,
			TotalScore:    0,
			CompletedIRs:  []types.IRCompletion{},
			HasAnchor:     false,
			ArenaScores:   make(map[string]types.ArenaScore),
			Status:        types.VerificationStatusUnverified,
		}
	}
	return record
}

// GetIRCompletion retrieves a specific IR completion for a user
func (k *Keeper) GetIRCompletion(walletAddr, irID string) (types.IRCompletion, bool) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	if userCompletions, ok := k.completions[walletAddr]; ok {
		completion, exists := userCompletions[irID]
		return completion, exists
	}
	return types.IRCompletion{}, false
}

// SetIRCompletion stores an IR completion
func (k *Keeper) SetIRCompletion(walletAddr string, completion types.IRCompletion) error {
	if walletAddr == "" {
		return types.ErrInvalidWalletAddress
	}
	if completion.IRID == "" {
		return types.ErrInvalidIRID
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	if k.completions[walletAddr] == nil {
		k.completions[walletAddr] = make(map[string]types.IRCompletion)
	}

	completion.CompletedHeight = k.currentHeight
	completion.CompletedAt = k.currentTime

	k.completions[walletAddr][completion.IRID] = completion
	return nil
}

// GetScoreHistory retrieves score change history for a user
func (k *Keeper) GetScoreHistory(walletAddr string, fromHeight, toHeight uint64, limit int) []types.ScoreChange {
	k.mu.RLock()
	defer k.mu.RUnlock()

	history, ok := k.scoreHistory[walletAddr]
	if !ok {
		return []types.ScoreChange{}
	}

	// Filter by height range if specified
	filtered := []types.ScoreChange{}
	for _, change := range history {
		if (fromHeight == 0 || change.BlockHeight >= fromHeight) &&
			(toHeight == 0 || change.BlockHeight <= toHeight) {
			filtered = append(filtered, change)
		}

		if limit > 0 && len(filtered) >= limit {
			break
		}
	}

	return filtered
}

// AddScoreChange adds a score change to history
func (k *Keeper) AddScoreChange(change types.ScoreChange) {
	k.mu.Lock()
	defer k.mu.Unlock()

	change.BlockHeight = k.currentHeight
	change.Timestamp = k.currentTime

	k.scoreHistory[change.TxHash] = append(k.scoreHistory[change.TxHash], change)
}

// GetSlashRecords retrieves slash records for a user
func (k *Keeper) GetSlashRecords(walletAddr string) []types.SlashRecord {
	k.mu.RLock()
	defer k.mu.RUnlock()

	records, ok := k.slashRecords[walletAddr]
	if !ok {
		return []types.SlashRecord{}
	}
	return records
}

// AddSlashRecord adds a slash record
func (k *Keeper) AddSlashRecord(record types.SlashRecord) error {
	if record.WalletAddress == "" {
		return types.ErrInvalidWalletAddress
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	record.SlashHeight = k.currentHeight
	record.SlashTime = k.currentTime

	k.slashRecords[record.WalletAddress] = append(k.slashRecords[record.WalletAddress], record)
	return nil
}

// GetSlashRecord retrieves a specific slash record
func (k *Keeper) GetSlashRecord(walletAddr, slashTxHash string) (types.SlashRecord, bool) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	records, ok := k.slashRecords[walletAddr]
	if !ok {
		return types.SlashRecord{}, false
	}

	for _, record := range records {
		if record.SlashTxHash == slashTxHash {
			return record, true
		}
	}

	return types.SlashRecord{}, false
}

// UpdateSlashRecord updates an existing slash record
func (k *Keeper) UpdateSlashRecord(walletAddr string, record types.SlashRecord) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	records, ok := k.slashRecords[walletAddr]
	if !ok {
		return types.ErrSlashNotFound
	}

	for i, r := range records {
		if r.SlashTxHash == record.SlashTxHash {
			k.slashRecords[walletAddr][i] = record
			return nil
		}
	}

	return types.ErrSlashNotFound
}

// InitGenesis initializes the keeper from genesis state
func (k *Keeper) InitGenesis(genesis types.GenesisState) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	// Set params
	if err := k.paramsStore.SetParams(genesis.Params); err != nil {
		return fmt.Errorf("failed to set params: %w", err)
	}

	// Load user records
	for _, record := range genesis.UserRecords {
		k.userRecords[record.WalletAddress] = record

		// Index completions
		if k.completions[record.WalletAddress] == nil {
			k.completions[record.WalletAddress] = make(map[string]types.IRCompletion)
		}
		for _, completion := range record.CompletedIRs {
			k.completions[record.WalletAddress][completion.IRID] = completion
		}
	}

	// Load slash records
	for _, slashRecord := range genesis.SlashRecords {
		k.slashRecords[slashRecord.WalletAddress] = append(
			k.slashRecords[slashRecord.WalletAddress],
			slashRecord,
		)
	}

	return nil
}

// ExportGenesis exports the current state for genesis
func (k *Keeper) ExportGenesis() types.GenesisState {
	k.mu.RLock()
	defer k.mu.RUnlock()

	userRecords := make([]types.UserConfidenceRecord, 0, len(k.userRecords))
	for _, record := range k.userRecords {
		userRecords = append(userRecords, record)
	}

	slashRecords := []types.SlashRecord{}
	for _, records := range k.slashRecords {
		slashRecords = append(slashRecords, records...)
	}

	return types.GenesisState{
		Params:       k.GetParams(),
		UserRecords:  userRecords,
		SlashRecords: slashRecords,
	}
}

// IsVerified checks if a user has verified status (CS >= threshold)
func (k *Keeper) IsVerified(walletAddr string) bool {
	record, ok := k.GetUserRecord(walletAddr)
	if !ok {
		return false
	}

	params := k.GetParams()
	return record.TotalScore >= params.VerificationThreshold &&
		record.Status == types.VerificationStatusVerified
}

// HasCompletedIR checks if a user has completed a specific IR
func (k *Keeper) HasCompletedIR(walletAddr, irID string) bool {
	_, ok := k.GetIRCompletion(walletAddr, irID)
	return ok
}

// GetArenaScore retrieves the score for a specific arena
func (k *Keeper) GetArenaScore(walletAddr, arena string) (uint64, error) {
	record, ok := k.GetUserRecord(walletAddr)
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
func (k *Keeper) GetUserScore(walletAddr string) (uint64, bool) {
	record, ok := k.GetUserRecord(walletAddr)
	if !ok {
		return 0, false
	}
	return record.TotalScore, true
}

// GetAnchorInfo retrieves anchor completion information for a user
func (k *Keeper) GetAnchorInfo(walletAddr string) (interface{}, bool) {
	record, ok := k.GetUserRecord(walletAddr)
	if !ok {
		return nil, false
	}

	if record.AnchorInfo == nil {
		return nil, false
	}

	return record.AnchorInfo, true
}

// ListVerifiedUsers returns a list of verified users
func (k *Keeper) ListVerifiedUsers(minScore uint64, limit int) ([]string, []uint64) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	params := k.GetParams()
	if minScore == 0 {
		minScore = params.VerificationThreshold
	}

	wallets := []string{}
	scores := []uint64{}

	for addr, record := range k.userRecords {
		if record.TotalScore >= minScore && record.Status == types.VerificationStatusVerified {
			wallets = append(wallets, addr)
			scores = append(scores, record.TotalScore)

			if limit > 0 && len(wallets) >= limit {
				break
			}
		}
	}

	return wallets, scores
}
