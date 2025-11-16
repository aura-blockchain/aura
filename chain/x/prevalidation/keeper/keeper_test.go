package keeper

import (
	"testing"
	"time"

	"github.com/aequitas/aura/chain/x/prevalidation/params"
	"github.com/aequitas/aura/chain/x/prevalidation/types"
)

// MockConfidenceScoreKeeper implements ConfidenceScoreKeeper for testing
type MockConfidenceScoreKeeper struct {
	scores map[string]uint64
}

func NewMockConfidenceScoreKeeper() *MockConfidenceScoreKeeper {
	return &MockConfidenceScoreKeeper{
		scores: make(map[string]uint64),
	}
}

func (m *MockConfidenceScoreKeeper) GetUserScore(walletAddr string) (uint64, bool) {
	score, ok := m.scores[walletAddr]
	return score, ok
}

func (m *MockConfidenceScoreKeeper) SetUserScore(walletAddr string, score uint64) {
	m.scores[walletAddr] = score
}

func setupTestKeeper() *Keeper {
	paramsStore := params.NewStore(types.DefaultParams())
	keeper := NewKeeper(paramsStore)
	keeper.SetCurrentHeight(100)
	keeper.SetCurrentTime(time.Now().Unix())
	return keeper
}

func TestNewKeeper(t *testing.T) {
	keeper := setupTestKeeper()

	if keeper == nil {
		t.Fatal("keeper should not be nil")
	}

	if keeper.paramsStore == nil {
		t.Error("params store should not be nil")
	}

	if len(keeper.encryptionKeys) == 0 {
		t.Error("encryption keys should be initialized")
	}
}

func TestCreatePreValidatedTransaction(t *testing.T) {
	keeper := setupTestKeeper()
	mockCS := NewMockConfidenceScoreKeeper()
	mockCS.SetUserScore("aura1test", 500)
	keeper.SetConfidenceScoreKeeper(mockCS)

	txData := []byte("test transaction data")
	signer := "aura1test"
	txType := types.TxTypeIRCompletion

	tx, err := keeper.CreatePreValidatedTransaction(
		txType,
		"template-1",
		txData,
		signer,
		50000,
		map[string]string{"key": "value"},
	)

	if err != nil {
		t.Fatalf("failed to create pre-validated transaction: %v", err)
	}

	if tx.Id == "" {
		t.Error("transaction ID should not be empty")
	}

	if tx.TxType != txType {
		t.Errorf("expected tx type %v, got %v", txType, tx.TxType)
	}

	if tx.Signer != signer {
		t.Errorf("expected signer %s, got %s", signer, tx.Signer)
	}

	if tx.Status != types.ValidationStatusPending {
		t.Errorf("expected status PENDING, got %v", tx.Status)
	}
}

func TestGetPreValidatedTransaction(t *testing.T) {
	keeper := setupTestKeeper()
	mockCS := NewMockConfidenceScoreKeeper()
	mockCS.SetUserScore("aura1test", 500)
	keeper.SetConfidenceScoreKeeper(mockCS)

	// Create a transaction
	txData := []byte("test data")
	tx, err := keeper.CreatePreValidatedTransaction(
		types.TxTypeIRCompletion,
		"template-1",
		txData,
		"aura1test",
		50000,
		nil,
	)

	if err != nil {
		t.Fatalf("failed to create transaction: %v", err)
	}

	// Retrieve it
	retrieved, ok := keeper.GetPreValidatedTransaction(tx.Id)
	if !ok {
		t.Fatal("should find the transaction")
	}

	if retrieved.Id != tx.Id {
		t.Errorf("expected ID %s, got %s", tx.Id, retrieved.Id)
	}
}

func TestEncryptDecrypt(t *testing.T) {
	keeper := setupTestKeeper()

	originalData := []byte("sensitive transaction data")

	// Encrypt
	encrypted, err := keeper.encryptTransactionData(originalData)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	if len(encrypted) == 0 {
		t.Error("encrypted data should not be empty")
	}

	// Decrypt
	decrypted, err := keeper.decryptTransactionData(encrypted, keeper.currentEncryptionKeyID)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	if string(decrypted) != string(originalData) {
		t.Errorf("expected %s, got %s", originalData, decrypted)
	}
}

func TestRegisterTemplate(t *testing.T) {
	keeper := setupTestKeeper()

	template := &types.ValidationTemplate{
		Id:              "test-template",
		TxType:          types.TxTypeIRCompletion,
		Name:            "Test Template",
		Description:     "A test template",
		ValidationRules: "{}",
		ParameterSchema: "{}",
		GasFormula:      "50000",
		PriorityWeight:  100,
		Active:          true,
	}

	err := keeper.RegisterTemplate(template)
	if err != nil {
		t.Fatalf("failed to register template: %v", err)
	}

	// Retrieve it
	retrieved, ok := keeper.GetTemplate(template.Id)
	if !ok {
		t.Fatal("should find the template")
	}

	if retrieved.Name != template.Name {
		t.Errorf("expected name %s, got %s", template.Name, retrieved.Name)
	}
}

func TestCacheEviction(t *testing.T) {
	keeper := setupTestKeeper()
	mockCS := NewMockConfidenceScoreKeeper()
	mockCS.SetUserScore("aura1test", 500)
	keeper.SetConfidenceScoreKeeper(mockCS)

	// Set a small cache size
	params := keeper.GetParams()
	params.MaxCacheSize = 5
	params.CacheStrategy = types.CacheStrategyFIFO
	keeper.SetParams(params)

	// Create more transactions than cache size
	for i := 0; i < 10; i++ {
		_, err := keeper.CreatePreValidatedTransaction(
			types.TxTypeIRCompletion,
			"template-1",
			[]byte("data"),
			"aura1test",
			50000,
			nil,
		)
		if err != nil {
			t.Fatalf("failed to create transaction %d: %v", i, err)
		}
	}

	// Cache should not exceed max size
	if uint64(len(keeper.preValidatedTxs)) > params.MaxCacheSize {
		t.Errorf("cache size %d exceeds max %d", len(keeper.preValidatedTxs), params.MaxCacheSize)
	}
}

func TestCleanupExpiredTransactions(t *testing.T) {
	keeper := setupTestKeeper()
	mockCS := NewMockConfidenceScoreKeeper()
	mockCS.SetUserScore("aura1test", 500)
	keeper.SetConfidenceScoreKeeper(mockCS)

	// Set short expiry time
	params := keeper.GetParams()
	params.ExpiryHours = 1
	keeper.SetParams(params)

	// Create a transaction
	tx, err := keeper.CreatePreValidatedTransaction(
		types.TxTypeIRCompletion,
		"template-1",
		[]byte("data"),
		"aura1test",
		50000,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create transaction: %v", err)
	}

	// Advance time past expiry
	keeper.SetCurrentTime(time.Now().Add(2 * time.Hour).Unix())

	// Cleanup
	expiredCount := keeper.CleanupExpiredTransactions()

	if expiredCount == 0 {
		t.Error("should have cleaned up at least one expired transaction")
	}

	// Transaction should be gone
	_, ok := keeper.GetPreValidatedTransaction(tx.Id)
	if ok {
		t.Error("expired transaction should be removed")
	}
}

func TestMetricsRecording(t *testing.T) {
	keeper := setupTestKeeper()

	// Record cache hit
	keeper.RecordCacheHit(types.TxTypeIRCompletion, 100)

	metrics := keeper.GetMetrics()
	if metrics.TotalCacheHits != 1 {
		t.Errorf("expected 1 cache hit, got %d", metrics.TotalCacheHits)
	}

	// Record cache miss
	keeper.RecordCacheMiss(types.TxTypeIRCompletion)

	metrics = keeper.GetMetrics()
	if metrics.TotalCacheMisses != 1 {
		t.Errorf("expected 1 cache miss, got %d", metrics.TotalCacheMisses)
	}

	// Check hit rate
	if metrics.OverallCacheHitRate != 0.5 {
		t.Errorf("expected hit rate 0.5, got %f", metrics.OverallCacheHitRate)
	}
}

func TestConfidenceScoreCheck(t *testing.T) {
	keeper := setupTestKeeper()
	mockCS := NewMockConfidenceScoreKeeper()
	keeper.SetConfidenceScoreKeeper(mockCS)

	// User with low score
	mockCS.SetUserScore("aura1low", 50)

	// Should fail for low confidence score
	_, err := keeper.CreatePreValidatedTransaction(
		types.TxTypeIRCompletion,
		"template-1",
		[]byte("data"),
		"aura1low",
		50000,
		nil,
	)

	if err != types.ErrInsufficientConfidence {
		t.Errorf("expected ErrInsufficientConfidence, got %v", err)
	}

	// User with high score
	mockCS.SetUserScore("aura1high", 500)

	// Should succeed
	_, err = keeper.CreatePreValidatedTransaction(
		types.TxTypeIRCompletion,
		"template-1",
		[]byte("data"),
		"aura1high",
		50000,
		nil,
	)

	if err != nil {
		t.Errorf("should succeed for high confidence score: %v", err)
	}
}

func TestExecutePreValidatedTransaction(t *testing.T) {
	keeper := setupTestKeeper()
	mockCS := NewMockConfidenceScoreKeeper()
	mockCS.SetUserScore("aura1test", 500)
	keeper.SetConfidenceScoreKeeper(mockCS)

	// Create a transaction
	originalData := []byte("test transaction data")
	tx, err := keeper.CreatePreValidatedTransaction(
		types.TxTypeIRCompletion,
		"template-1",
		originalData,
		"aura1test",
		50000,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create transaction: %v", err)
	}

	// Mark as validated
	tx.Status = types.ValidationStatusValidated
	keeper.preValidatedTxs[tx.Id] = tx

	// Execute it
	decryptedData, err := keeper.ExecutePreValidatedTransaction(tx.Id)
	if err != nil {
		t.Fatalf("failed to execute transaction: %v", err)
	}

	if string(decryptedData) != string(originalData) {
		t.Errorf("expected data %s, got %s", originalData, decryptedData)
	}

	// Check status updated
	executed, _ := keeper.GetPreValidatedTransaction(tx.Id)
	if executed.Status != types.ValidationStatusExecuted {
		t.Errorf("expected status EXECUTED, got %v", executed.Status)
	}
}
