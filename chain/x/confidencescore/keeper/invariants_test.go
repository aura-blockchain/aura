package keeper

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/confidencescore/types"
)

type InvariantsTestSuite struct {
	KeeperTestSuite
}

func TestInvariantsTestSuite(t *testing.T) {
	suite.Run(t, new(InvariantsTestSuite))
}

func (suite *InvariantsTestSuite) TestAllInvariants() {
	// Test: All invariants on empty keeper
	inv := AllInvariants(suite.Keeper)
	msg, broken := inv()
	suite.False(broken, "all invariants should pass on empty keeper")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestRegisterInvariants() {
	// Register invariants - should not panic
	suite.NotPanics(func() {
		RegisterInvariants(suite.Keeper)
	})
}

func (suite *InvariantsTestSuite) TestParamsInvariant() {
	// Test: valid params pass
	inv := ParamsInvariant(suite.Keeper)
	msg, broken := inv()
	suite.False(broken, "valid params should pass")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestUserRecordConsistencyInvariant() {
	// Test: valid user record passes
	walletAddr := "cosmos1test"

	suite.Keeper.RegisterUser(walletAddr)
	suite.Keeper.RecordIRCompletion(walletAddr, "ir-1", "proof-hash-1")

	inv := UserRecordConsistencyInvariant(suite.Keeper)
	msg, broken := inv()
	suite.False(broken, "valid user record should pass")
	suite.Empty(msg)

	// Test: completion count mismatch (manual manipulation)
	suite.SetupTest()
	suite.Keeper.RegisterUser(walletAddr)
	// Manually set incorrect completion count
	suite.Keeper.mu.Lock()
	if record, ok := suite.Keeper.userRecords[walletAddr]; ok {
		record.IRCompletions = 10 // Incorrect count
		suite.Keeper.userRecords[walletAddr] = record
	}
	suite.Keeper.mu.Unlock()

	msg, broken = inv()
	suite.True(broken, "completion count mismatch should break invariant")
	suite.Contains(msg, "completion count mismatch")

	// Test: score exceeds maximum
	suite.SetupTest()
	suite.Keeper.RegisterUser(walletAddr)
	suite.Keeper.mu.Lock()
	if record, ok := suite.Keeper.userRecords[walletAddr]; ok {
		record.ConfidenceScore = 20000 // Exceeds max
		suite.Keeper.userRecords[walletAddr] = record
	}
	suite.Keeper.mu.Unlock()

	msg, broken = inv()
	suite.True(broken, "score exceeding maximum should break invariant")
	suite.Contains(msg, "score exceeding maximum")

	// Test: future last update time
	suite.SetupTest()
	suite.Keeper.RegisterUser(walletAddr)
	suite.Keeper.mu.Lock()
	if record, ok := suite.Keeper.userRecords[walletAddr]; ok {
		record.LastUpdateTime = suite.Keeper.currentTime + 7200 // 2 hours in future
		suite.Keeper.userRecords[walletAddr] = record
	}
	suite.Keeper.mu.Unlock()

	msg, broken = inv()
	suite.True(broken, "future last update time should break invariant")
	suite.Contains(msg, "future last update time")
}

func (suite *InvariantsTestSuite) TestCompletionValidityInvariant() {
	// Test: valid completions pass
	walletAddr := "cosmos1test"

	suite.Keeper.RegisterUser(walletAddr)
	suite.Keeper.RecordIRCompletion(walletAddr, "ir-1", "proof-hash-1")

	inv := CompletionValidityInvariant(suite.Keeper)
	msg, broken := inv()
	suite.False(broken, "valid completions should pass")
	suite.Empty(msg)

	// Test: empty IR ID
	suite.SetupTest()
	suite.Keeper.RegisterUser(walletAddr)
	suite.Keeper.mu.Lock()
	suite.Keeper.completions[walletAddr] = map[string]types.IRCompletion{
		"": { // Empty IR ID
			CompletionTime: suite.Keeper.currentTime,
			ProofHash:      "proof-hash",
		},
	}
	suite.Keeper.mu.Unlock()

	msg, broken = inv()
	suite.True(broken, "empty IR ID should break invariant")
	suite.Contains(msg, "empty IR ID")

	// Test: zero completion timestamp
	suite.SetupTest()
	suite.Keeper.RegisterUser(walletAddr)
	suite.Keeper.mu.Lock()
	suite.Keeper.completions[walletAddr] = map[string]types.IRCompletion{
		"ir-1": {
			CompletionTime: 0, // Zero timestamp
			ProofHash:      "proof-hash",
		},
	}
	suite.Keeper.mu.Unlock()

	msg, broken = inv()
	suite.True(broken, "zero completion timestamp should break invariant")
	suite.Contains(msg, "zero timestamp")

	// Test: empty proof hash
	suite.SetupTest()
	suite.Keeper.RegisterUser(walletAddr)
	suite.Keeper.mu.Lock()
	suite.Keeper.completions[walletAddr] = map[string]types.IRCompletion{
		"ir-1": {
			CompletionTime: suite.Keeper.currentTime,
			ProofHash:      "", // Empty proof hash
		},
	}
	suite.Keeper.mu.Unlock()

	msg, broken = inv()
	suite.True(broken, "empty proof hash should break invariant")
	suite.Contains(msg, "empty proof hash")

	// Test: future completion time
	suite.SetupTest()
	suite.Keeper.RegisterUser(walletAddr)
	suite.Keeper.mu.Lock()
	suite.Keeper.completions[walletAddr] = map[string]types.IRCompletion{
		"ir-1": {
			CompletionTime: suite.Keeper.currentTime + 7200, // 2 hours in future
			ProofHash:      "proof-hash",
		},
	}
	suite.Keeper.mu.Unlock()

	msg, broken = inv()
	suite.True(broken, "future completion time should break invariant")
	suite.Contains(msg, "future completion")
}

func (suite *InvariantsTestSuite) TestScoreRangeInvariant() {
	// Test: valid scores pass
	walletAddr := "cosmos1test"

	suite.Keeper.RegisterUser(walletAddr)

	inv := ScoreRangeInvariant(suite.Keeper)
	msg, broken := inv()
	suite.False(broken, "valid scores should pass")
	suite.Empty(msg)

	// Test: score exceeds maximum
	suite.SetupTest()
	suite.Keeper.RegisterUser(walletAddr)
	suite.Keeper.mu.Lock()
	if record, ok := suite.Keeper.userRecords[walletAddr]; ok {
		record.ConfidenceScore = 20000 // Exceeds max of 10000
		suite.Keeper.userRecords[walletAddr] = record
	}
	suite.Keeper.mu.Unlock()

	msg, broken = inv()
	suite.True(broken, "score exceeding maximum should break invariant")
	suite.Contains(msg, "exceeds maximum")

	// Test: score history exceeds maximum
	suite.SetupTest()
	suite.Keeper.RegisterUser(walletAddr)
	suite.Keeper.mu.Lock()
	suite.Keeper.scoreHistory[walletAddr] = []types.ScoreChange{
		{
			OldScore:  100,
			NewScore:  20000, // Exceeds max
			Reason:    "test",
			Timestamp: suite.Keeper.currentTime,
		},
	}
	suite.Keeper.mu.Unlock()

	msg, broken = inv()
	suite.True(broken, "score history exceeding maximum should break invariant")
	suite.Contains(msg, "exceeds maximum")
}

func (suite *InvariantsTestSuite) TestSlashRecordIntegrityInvariant() {
	// Test: valid slash records pass
	walletAddr := "cosmos1test"

	suite.Keeper.RegisterUser(walletAddr)
	suite.Keeper.SlashScore(walletAddr, 100, "violation")

	inv := SlashRecordIntegrityInvariant(suite.Keeper)
	msg, broken := inv()
	suite.False(broken, "valid slash records should pass")
	suite.Empty(msg)

	// Test: slash records for non-existent user
	suite.SetupTest()
	suite.Keeper.mu.Lock()
	suite.Keeper.slashRecords["nonexistent"] = []types.SlashRecord{
		{
			SlashAmount: 100,
			Reason:      "test",
			SlashTime:   suite.Keeper.currentTime,
		},
	}
	suite.Keeper.mu.Unlock()

	msg, broken = inv()
	suite.True(broken, "slash records for non-existent user should break invariant")
	suite.Contains(msg, "non-existent user")

	// Test: non-positive slash amount
	suite.SetupTest()
	suite.Keeper.RegisterUser(walletAddr)
	suite.Keeper.mu.Lock()
	suite.Keeper.slashRecords[walletAddr] = []types.SlashRecord{
		{
			SlashAmount: 0, // Non-positive
			Reason:      "test",
			SlashTime:   suite.Keeper.currentTime,
		},
	}
	suite.Keeper.mu.Unlock()

	msg, broken = inv()
	suite.True(broken, "non-positive slash amount should break invariant")
	suite.Contains(msg, "non-positive amount")

	// Test: empty reason
	suite.SetupTest()
	suite.Keeper.RegisterUser(walletAddr)
	suite.Keeper.mu.Lock()
	suite.Keeper.slashRecords[walletAddr] = []types.SlashRecord{
		{
			SlashAmount: 100,
			Reason:      "", // Empty reason
			SlashTime:   suite.Keeper.currentTime,
		},
	}
	suite.Keeper.mu.Unlock()

	msg, broken = inv()
	suite.True(broken, "empty reason should break invariant")
	suite.Contains(msg, "empty reason")

	// Test: zero timestamp
	suite.SetupTest()
	suite.Keeper.RegisterUser(walletAddr)
	suite.Keeper.mu.Lock()
	suite.Keeper.slashRecords[walletAddr] = []types.SlashRecord{
		{
			SlashAmount: 100,
			Reason:      "test",
			SlashTime:   0, // Zero timestamp
		},
	}
	suite.Keeper.mu.Unlock()

	msg, broken = inv()
	suite.True(broken, "zero timestamp should break invariant")
	suite.Contains(msg, "zero timestamp")
}

func (suite *InvariantsTestSuite) TestProofHashUniquenessInvariant() {
	// Test: valid proof hashes pass
	walletAddr := "cosmos1test"

	suite.Keeper.RegisterUser(walletAddr)
	suite.Keeper.RecordIRCompletion(walletAddr, "ir-1", "proof-hash-1")

	inv := ProofHashUniquenessInvariant(suite.Keeper)
	msg, broken := inv()
	suite.False(broken, "valid proof hashes should pass")
	suite.Empty(msg)

	// Test: proof hash count mismatch
	suite.SetupTest()
	suite.Keeper.RegisterUser(walletAddr)
	suite.Keeper.RecordIRCompletion(walletAddr, "ir-1", "proof-hash-1")
	suite.Keeper.mu.Lock()
	// Add extra proof hash
	suite.Keeper.proofHashes[walletAddr]["extra-hash"] = true
	suite.Keeper.mu.Unlock()

	msg, broken = inv()
	suite.True(broken, "proof hash count mismatch should break invariant")
	suite.Contains(msg, "doesn't match completions")

	// Test: completion proof hash not in proof hash map
	suite.SetupTest()
	suite.Keeper.RegisterUser(walletAddr)
	suite.Keeper.mu.Lock()
	suite.Keeper.completions[walletAddr] = map[string]types.IRCompletion{
		"ir-1": {
			CompletionTime: suite.Keeper.currentTime,
			ProofHash:      "missing-hash",
		},
	}
	suite.Keeper.proofHashes[walletAddr] = map[string]bool{}
	suite.Keeper.mu.Unlock()

	msg, broken = inv()
	suite.True(broken, "missing proof hash should break invariant")
	suite.Contains(msg, "not in proof hash map")
}
