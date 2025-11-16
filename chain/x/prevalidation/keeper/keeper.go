package keeper

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/aequitas/aura/chain/x/prevalidation/params"
	"github.com/aequitas/aura/chain/x/prevalidation/types"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ConfidenceScoreKeeper interface for confidence score checks
type ConfidenceScoreKeeper interface {
	GetUserScore(walletAddr string) (uint64, bool)
}

// Keeper manages the state of the prevalidation module
type Keeper struct {
	mu                     sync.RWMutex
	preValidatedTxs        map[string]*types.PreValidatedTransaction // tx_id -> PreValidatedTransaction
	userPreValidatedTxs    map[string][]string                       // signer -> []tx_id
	templatesByType        map[types.TransactionType][]*types.ValidationTemplate
	templatesById          map[string]*types.ValidationTemplate
	metrics                *types.PreValidationMetrics
	paramsStore            *params.Store
	csKeeper               ConfidenceScoreKeeper
	currentHeight          uint64
	currentTime            int64
	encryptionKeys         map[string][]byte // key_id -> encryption key
	currentEncryptionKeyID string
	lastSchedulerRun       time.Time
	lastAutoScaleRun       time.Time
	typeAmounts            map[types.TransactionType]uint64 // Current amounts per type
	lastScalingAdjustment  map[types.TransactionType]time.Time

	// Cache management
	cacheOrder       []string             // For LRU/FIFO strategies
	cacheAccessCount map[string]uint64    // For LFU strategy
	cacheAccessTime  map[string]time.Time // For LRU strategy
}

// NewKeeper creates a new Keeper instance
func NewKeeper(store *params.Store) *Keeper {
	if store == nil {
		store = params.NewStore(types.DefaultParams())
	}

	k := &Keeper{
		preValidatedTxs:     make(map[string]*types.PreValidatedTransaction),
		userPreValidatedTxs: make(map[string][]string),
		templatesByType:     make(map[types.TransactionType][]*types.ValidationTemplate),
		templatesById:       make(map[string]*types.ValidationTemplate),
		metrics: &types.PreValidationMetrics{
			MetricsByType: make(map[string]*types.TypeMetrics),
			Last24Hours:   []*types.HourlyMetrics{},
			ControlGroup:  &types.ControlGroupMetrics{},
		},
		paramsStore:           store,
		encryptionKeys:        make(map[string][]byte),
		typeAmounts:           make(map[types.TransactionType]uint64),
		lastScalingAdjustment: make(map[types.TransactionType]time.Time),
		cacheOrder:            []string{},
		cacheAccessCount:      make(map[string]uint64),
		cacheAccessTime:       make(map[string]time.Time),
		currentTime:           time.Now().Unix(),
	}

	// Initialize encryption key
	k.initializeEncryptionKey()

	// Initialize type amounts from config
	k.initializeTypeAmounts()

	return k
}

// SetConfidenceScoreKeeper sets the confidence score keeper
func (k *Keeper) SetConfidenceScoreKeeper(keeper ConfidenceScoreKeeper) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.csKeeper = keeper
}

// SetCurrentHeight sets the current block height
func (k *Keeper) SetCurrentHeight(height uint64) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.currentHeight = height
}

// SetCurrentTime sets the current time
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
		return types.ErrInvalidParameters
	}
	return k.paramsStore.SetParams(params)
}

// initializeEncryptionKey creates the initial encryption key
func (k *Keeper) initializeEncryptionKey() {
	key := make([]byte, 32) // 256-bit key for AES-256
	if _, err := rand.Read(key); err != nil {
		// Fallback to deterministic key (not secure for production)
		h := sha256.New()
		h.Write([]byte("prevalidation-encryption-key"))
		key = h.Sum(nil)
	}

	keyID := k.generateKeyID()
	k.encryptionKeys[keyID] = key
	k.currentEncryptionKeyID = keyID
}

// generateKeyID generates a unique key ID
func (k *Keeper) generateKeyID() string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// initializeTypeAmounts initializes the amounts per transaction type
func (k *Keeper) initializeTypeAmounts() {
	params := k.GetParams()
	if params.AutoScalingConfig == nil {
		return
	}

	for txTypeStr, amount := range params.AutoScalingConfig.InitialAmounts {
		// Convert string back to TransactionType
		// This is a simplified approach; in production, you'd want proper conversion
		txType := k.parseTransactionType(txTypeStr)
		k.typeAmounts[txType] = amount
	}
}

// parseTransactionType converts string to TransactionType
func (k *Keeper) parseTransactionType(s string) types.TransactionType {
	switch s {
	case types.TxTypeIRCompletion.String():
		return types.TxTypeIRCompletion
	case types.TxTypeDexSwap.String():
		return types.TxTypeDexSwap
	case types.TxTypeLPDeposit.String():
		return types.TxTypeLPDeposit
	case types.TxTypeLPWithdrawal.String():
		return types.TxTypeLPWithdrawal
	case types.TxTypeVCMint.String():
		return types.TxTypeVCMint
	case types.TxTypeBridgeTransfer.String():
		return types.TxTypeBridgeTransfer
	case types.TxTypeConfidenceScoreUpdate.String():
		return types.TxTypeConfidenceScoreUpdate
	case types.TxTypeIdentityChange.String():
		return types.TxTypeIdentityChange
	default:
		return types.TxTypeUnspecified
	}
}

// ============================
// ENCRYPTION/DECRYPTION
// ============================

// encryptTransactionData encrypts transaction data
func (k *Keeper) encryptTransactionData(data []byte) ([]byte, error) {
	key, ok := k.encryptionKeys[k.currentEncryptionKeyID]
	if !ok {
		return nil, types.ErrEncryptionFailed
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, types.ErrEncryptionFailed
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, types.ErrEncryptionFailed
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, types.ErrEncryptionFailed
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// decryptTransactionData decrypts transaction data
func (k *Keeper) decryptTransactionData(encryptedData []byte, keyID string) ([]byte, error) {
	key, ok := k.encryptionKeys[keyID]
	if !ok {
		return nil, types.ErrDecryptionFailed
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, types.ErrDecryptionFailed
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, types.ErrDecryptionFailed
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, types.ErrDecryptionFailed
	}

	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, types.ErrDecryptionFailed
	}

	return plaintext, nil
}

// ============================
// TEMPLATE MANAGEMENT
// ============================

// RegisterTemplate registers a new validation template
func (k *Keeper) RegisterTemplate(template *types.ValidationTemplate) error {
	if template.Id == "" {
		return types.ErrInvalidTemplate
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	template.CreatedAt = timestamppb.New(time.Unix(k.currentTime, 0))
	template.UpdatedAt = timestamppb.New(time.Unix(k.currentTime, 0))

	if template.Stats == nil {
		template.Stats = &types.TemplateStats{}
	}

	k.templatesById[template.Id] = template
	k.templatesByType[template.TxType] = append(k.templatesByType[template.TxType], template)

	return nil
}

// GetTemplate retrieves a template by ID
func (k *Keeper) GetTemplate(templateID string) (*types.ValidationTemplate, bool) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	template, ok := k.templatesById[templateID]
	return template, ok
}

// GetTemplatesByType retrieves all templates for a transaction type
func (k *Keeper) GetTemplatesByType(txType types.TransactionType) []*types.ValidationTemplate {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.templatesByType[txType]
}

// UpdateTemplateStats updates statistics for a template
func (k *Keeper) UpdateTemplateStats(templateID string, executed bool, expired bool) {
	k.mu.Lock()
	defer k.mu.Unlock()

	template, ok := k.templatesById[templateID]
	if !ok {
		return
	}

	if template.Stats == nil {
		template.Stats = &types.TemplateStats{}
	}

	if executed {
		template.Stats.TotalExecuted++
	}
	if expired {
		template.Stats.TotalExpired++
	}

	template.Stats.LastUsed = timestamppb.New(time.Unix(k.currentTime, 0))

	// Update cache hit rate
	total := template.Stats.TotalExecuted + template.Stats.TotalExpired
	if total > 0 {
		template.Stats.CacheHitRate = float64(template.Stats.TotalExecuted) / float64(total)
	}
}

// ============================
// PRE-VALIDATED TRANSACTION MANAGEMENT
// ============================

// CreatePreValidatedTransaction creates a new pre-validated transaction
func (k *Keeper) CreatePreValidatedTransaction(
	txType types.TransactionType,
	templateID string,
	txData []byte,
	signer string,
	estimatedGas uint64,
	context map[string]string,
) (*types.PreValidatedTransaction, error) {
	params := k.GetParams()

	// Check if enabled
	if !params.Enabled {
		return nil, types.ErrSchedulerDisabled
	}

	// Check confidence score
	if k.csKeeper != nil {
		score, exists := k.csKeeper.GetUserScore(signer)
		if !exists || score < params.MinConfidenceScore {
			return nil, types.ErrInsufficientConfidence
		}
	}

	// Encrypt transaction data
	encryptedData, err := k.encryptTransactionData(txData)
	if err != nil {
		return nil, err
	}

	// Generate transaction hash
	h := sha256.New()
	h.Write(txData)
	txHash := h.Sum(nil)

	// Calculate expiry
	expiresAt := time.Unix(k.currentTime, 0).Add(time.Duration(params.ExpiryHours) * time.Hour)

	// Generate unique ID
	txID := k.generateTxID(signer, txType, k.currentTime)

	preValidatedTx := &types.PreValidatedTransaction{
		Id:              txID,
		TxType:          txType,
		TemplateId:      templateID,
		EncryptedData:   encryptedData,
		EncryptionKeyId: k.currentEncryptionKeyID,
		ValidationMetadata: &types.ValidationMetadata{
			ValidatorNode:        "default-validator",
			ValidationDurationMs: 0, // Will be set during validation
			EngineVersion:        "1.0.0",
		},
		TxHash:          txHash,
		Signer:          signer,
		Nonce:           uint64(time.Now().UnixNano()),
		Status:          types.ValidationStatusPending,
		ExpiresAt:       timestamppb.New(expiresAt),
		ValidatedAt:     timestamppb.New(time.Unix(k.currentTime, 0)),
		ValidatedHeight: k.currentHeight,
		EstimatedGas:    estimatedGas,
		Context:         context,
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	// Check cache size
	if uint64(len(k.preValidatedTxs)) >= params.MaxCacheSize {
		// Evict based on strategy
		if err := k.evictTransaction(); err != nil {
			return nil, err
		}
	}

	// Store transaction
	k.preValidatedTxs[txID] = preValidatedTx
	k.userPreValidatedTxs[signer] = append(k.userPreValidatedTxs[signer], txID)

	// Update cache tracking
	k.cacheOrder = append(k.cacheOrder, txID)
	k.cacheAccessTime[txID] = time.Unix(k.currentTime, 0)
	k.cacheAccessCount[txID] = 0

	// Update metrics
	k.metrics.TotalPreValidations++

	return preValidatedTx, nil
}

// generateTxID generates a unique transaction ID
func (k *Keeper) generateTxID(signer string, txType types.TransactionType, timestamp int64) string {
	h := sha256.New()
	h.Write([]byte(signer))
	h.Write([]byte(fmt.Sprintf("%d", txType)))
	h.Write([]byte(fmt.Sprintf("%d", timestamp)))
	h.Write([]byte(fmt.Sprintf("%d", k.currentHeight)))
	return "pvtx:" + hex.EncodeToString(h.Sum(nil))[:32]
}

// GetPreValidatedTransaction retrieves a pre-validated transaction
func (k *Keeper) GetPreValidatedTransaction(txID string) (*types.PreValidatedTransaction, bool) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	tx, ok := k.preValidatedTxs[txID]
	if !ok {
		return nil, false
	}

	// Update access tracking for cache strategies
	k.mu.RUnlock()
	k.mu.Lock()
	k.cacheAccessCount[txID]++
	k.cacheAccessTime[txID] = time.Unix(k.currentTime, 0)
	k.mu.Unlock()
	k.mu.RLock()

	return tx, true
}

// FindPreValidatedTransaction finds a matching pre-validated transaction
func (k *Keeper) FindPreValidatedTransaction(
	txType types.TransactionType,
	signer string,
	context map[string]string,
) (*types.PreValidatedTransaction, bool) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	userTxs, ok := k.userPreValidatedTxs[signer]
	if !ok {
		return nil, false
	}

	currentTime := time.Unix(k.currentTime, 0)

	// Find best matching transaction
	for _, txID := range userTxs {
		tx, ok := k.preValidatedTxs[txID]
		if !ok || tx.TxType != txType {
			continue
		}

		// Check if valid and not expired
		if !tx.CanExecute(currentTime) {
			continue
		}

		// Check context matching (simplified - in production, use more sophisticated matching)
		if k.contextMatches(tx.Context, context) {
			return tx, true
		}
	}

	return nil, false
}

// contextMatches checks if contexts match (simplified implementation)
func (k *Keeper) contextMatches(templateContext, requestContext map[string]string) bool {
	// For now, require exact match of all keys
	// In production, implement more sophisticated matching logic
	if len(templateContext) != len(requestContext) {
		return false
	}

	for key, val := range templateContext {
		if requestContext[key] != val {
			return false
		}
	}

	return true
}

// ExecutePreValidatedTransaction executes a pre-validated transaction
func (k *Keeper) ExecutePreValidatedTransaction(txID string) ([]byte, error) {
	tx, ok := k.GetPreValidatedTransaction(txID)
	if !ok {
		return nil, types.ErrPreValidationNotFound
	}

	currentTime := time.Unix(k.currentTime, 0)

	// Check if can execute
	if !tx.CanExecute(currentTime) {
		if tx.IsExpired(currentTime) {
			return nil, types.ErrPreValidationExpired
		}
		return nil, types.ErrPreValidationAlreadyUsed
	}

	// Decrypt transaction data
	txData, err := k.decryptTransactionData(tx.EncryptedData, tx.EncryptionKeyId)
	if err != nil {
		return nil, err
	}

	// Mark as executed
	k.mu.Lock()
	tx.MarkExecuted(k.currentHeight, currentTime)
	k.mu.Unlock()

	// Update metrics
	k.metrics.TotalExecuted++
	k.UpdateTemplateStats(tx.TemplateId, true, false)

	return txData, nil
}

// evictTransaction evicts a transaction based on cache strategy
func (k *Keeper) evictTransaction() error {
	params := k.GetParams()

	var txIDToEvict string

	switch params.CacheStrategy {
	case types.CacheStrategyFIFO:
		if len(k.cacheOrder) > 0 {
			txIDToEvict = k.cacheOrder[0]
			k.cacheOrder = k.cacheOrder[1:]
		}

	case types.CacheStrategyLRU:
		// Find least recently accessed
		var oldestTime time.Time
		for txID, accessTime := range k.cacheAccessTime {
			if txIDToEvict == "" || accessTime.Before(oldestTime) {
				txIDToEvict = txID
				oldestTime = accessTime
			}
		}

	case types.CacheStrategyLFU:
		// Find least frequently used
		var lowestCount uint64 = ^uint64(0)
		for txID, count := range k.cacheAccessCount {
			if count < lowestCount {
				txIDToEvict = txID
				lowestCount = count
			}
		}

	case types.CacheStrategyAdaptive:
		// Hybrid: Consider both frequency and recency
		var lowestScore float64 = -1
		currentTime := time.Unix(k.currentTime, 0)
		for txID := range k.preValidatedTxs {
			accessCount := float64(k.cacheAccessCount[txID])
			timeSinceAccess := currentTime.Sub(k.cacheAccessTime[txID]).Hours()
			score := accessCount / (1 + timeSinceAccess) // Higher is better
			if lowestScore < 0 || score < lowestScore {
				txIDToEvict = txID
				lowestScore = score
			}
		}

	default:
		// Default to FIFO
		if len(k.cacheOrder) > 0 {
			txIDToEvict = k.cacheOrder[0]
			k.cacheOrder = k.cacheOrder[1:]
		}
	}

	if txIDToEvict == "" {
		return types.ErrCacheFull
	}

	// Remove transaction
	tx := k.preValidatedTxs[txIDToEvict]
	delete(k.preValidatedTxs, txIDToEvict)
	delete(k.cacheAccessCount, txIDToEvict)
	delete(k.cacheAccessTime, txIDToEvict)

	// Remove from user index
	if tx != nil {
		userTxs := k.userPreValidatedTxs[tx.Signer]
		newUserTxs := []string{}
		for _, id := range userTxs {
			if id != txIDToEvict {
				newUserTxs = append(newUserTxs, id)
			}
		}
		k.userPreValidatedTxs[tx.Signer] = newUserTxs
	}

	return nil
}

// CleanupExpiredTransactions removes expired pre-validated transactions
func (k *Keeper) CleanupExpiredTransactions() uint64 {
	k.mu.Lock()
	defer k.mu.Unlock()

	currentTime := time.Unix(k.currentTime, 0)
	expiredCount := uint64(0)

	for txID, tx := range k.preValidatedTxs {
		if tx.IsExpired(currentTime) && tx.Status != types.ValidationStatusExecuted {
			tx.MarkExpired()
			k.UpdateTemplateStats(tx.TemplateId, false, true)

			// Remove from cache
			delete(k.preValidatedTxs, txID)
			delete(k.cacheAccessCount, txID)
			delete(k.cacheAccessTime, txID)

			// Remove from user index
			userTxs := k.userPreValidatedTxs[tx.Signer]
			newUserTxs := []string{}
			for _, id := range userTxs {
				if id != txID {
					newUserTxs = append(newUserTxs, id)
				}
			}
			k.userPreValidatedTxs[tx.Signer] = newUserTxs

			expiredCount++
			k.metrics.TotalExpired++
		}
	}

	return expiredCount
}

// GetMetrics returns current metrics
func (k *Keeper) GetMetrics() *types.PreValidationMetrics {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.metrics
}

// RecordCacheHit records a cache hit
func (k *Keeper) RecordCacheHit(txType types.TransactionType, timeSavedMs uint64) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.metrics.RecordCacheHit(txType, timeSavedMs)
}

// RecordCacheMiss records a cache miss
func (k *Keeper) RecordCacheMiss(txType types.TransactionType) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.metrics.RecordCacheMiss(txType)
}

// GetTypeAmount returns the current amount for a transaction type
func (k *Keeper) GetTypeAmount(txType types.TransactionType) uint64 {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.typeAmounts[txType]
}

// SetTypeAmount sets the amount for a transaction type
func (k *Keeper) SetTypeAmount(txType types.TransactionType, amount uint64) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.typeAmounts[txType] = amount
}
