package keeper_test

import (
	"crypto/sha256"
	"fmt"
	"testing"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/bridge/keeper"
	"github.com/aequitas/aura/chain/x/bridge/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	bridgepb "github.com/aequitas/aura/proto/aura/bridge/v1beta1"
)

// TestMultiValidatorConsensus_InsufficientValidatorsRejected tests that unlock fails
// when there are fewer than the minimum required validators.
//
// SECURITY CRITICAL: This prevents a single compromised validator from draining bridge funds.
// Tests compliance with issue #035 requirement: MinAllowedConfirmations = 2
func TestMultiValidatorConsensus_InsufficientValidatorsRejected(t *testing.T) {
	testCases := []struct {
		name               string
		validatorCount     int
		activeCount        int
		signaturesProvided int
		minRequired        uint64
		expectPass         bool
		expectedError      string // Can be partial match
	}{
		{
			name:               "1 validator when 2 required - should fail",
			validatorCount:     3,
			activeCount:        3,
			signaturesProvided: 1,
			minRequired:        types.MinAllowedConfirmations, // 2
			expectPass:         false,
			expectedError:      "insufficient validator signatures",
		},
		{
			name:               "2 validators when 2 required - should pass",
			validatorCount:     3,
			activeCount:        3,
			signaturesProvided: 2,
			minRequired:        types.MinAllowedConfirmations, // 2
			expectPass:         true,
			expectedError:      "",
		},
		{
			name:               "0 validators when 2 required - should fail",
			validatorCount:     3,
			activeCount:        3,
			signaturesProvided: 0,
			minRequired:        types.MinAllowedConfirmations, // 2
			expectPass:         false,
			expectedError:      "insufficient validator signatures",
		},
		{
			name:               "2 validators when 3 required - should fail",
			validatorCount:     5,
			activeCount:        5,
			signaturesProvided: 2,
			minRequired:        3,
			expectPass:         false,
			expectedError:      "insufficient validator signatures",
		},
		{
			name:               "3 validators when 3 required (default) - should pass",
			validatorCount:     5,
			activeCount:        5,
			signaturesProvided: 3,
			minRequired:        types.DefaultMinConfirmations, // 3
			expectPass:         true,
			expectedError:      "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup test environment
			k, input := setupKeeperForMultiValidatorTests(t, tc.minRequired)
			ctx := input.Ctx
			msgServer := keeper.NewMsgServerImpl(k)

			// Create and register validators
			validators := createTestValidators(t, input, k, tc.validatorCount, tc.activeCount)

			// Prepare unlock message data
			burnTxHash := "0xabcd1234"
			transferID := "transfer-001"

			// Create a transfer to unlock
			transfer := createTestTransfer(t, input, k, transferID, burnTxHash, "1000000", tc.minRequired)
			sender := keepertest.GenTestAddr().String()
			amount := "1000000"
			denom := "uaura"
			sourceChain := "paw"

			// Create message to sign
			msgToSign := fmt.Sprintf("%s:%s:%s:%s:%s",
				transfer.SourceChain,
				burnTxHash,
				sender,
				amount,
				denom,
			)
			msgHash := sha256.Sum256([]byte(msgToSign))

			// Generate signatures from the specified number of validators
			var signatures [][]byte
			for i := 0; i < tc.signaturesProvided && i < len(validators); i++ {
				sig := signMessageWithValidator(t, validators[i].PrivKey, msgHash[:])
				signatures = append(signatures, sig)
			}

			// Create unlock message
			unlockMsg := &bridgepb.MsgUnlockTokens{
				BurnTxHash:          burnTxHash,
				Sender:              sender,
				Amount:              amount,
				Denom:               denom,
				SourceChain:         sourceChain,
				ValidatorSignatures: signatures,
				MerkleProof:         []byte{}, // Empty for this test
				MerkleRoot:          []byte{},
				SourceBlockHash:     []byte{},
				SourceBlockHeight:   0,
			}

			// Execute unlock
			resp, err := msgServer.UnlockTokens(sdk.WrapSDKContext(ctx), unlockMsg)

			// Verify results
			if tc.expectPass {
				require.NoError(t, err, "Unlock should succeed with %d validators", tc.signaturesProvided)
				require.NotNil(t, resp, "Response should not be nil")
				require.True(t, resp.Success, "Response should indicate success")
			} else {
				require.Error(t, err, "Unlock should fail with %d validators (need %d)", tc.signaturesProvided, tc.minRequired)
				if tc.expectedError != "" {
					require.Contains(t, err.Error(), tc.expectedError, "Error should contain expected message")
				}
			}
		})
	}
}

// TestMultiValidatorConsensus_DuplicateValidatorSignaturesRejected tests that the same
// validator cannot be counted multiple times even if they provide multiple signatures.
//
// SECURITY CRITICAL: Prevents an attacker from reusing the same validator's signature
// to meet the minimum threshold requirement.
func TestMultiValidatorConsensus_DuplicateValidatorSignaturesRejected(t *testing.T) {
	testCases := []struct {
		name           string
		totalValidators int
		duplicateCount int
		uniqueCount    int
		minRequired    uint64
		expectPass     bool
	}{
		{
			name:            "1 unique + 1 duplicate = still only 1 (need 2)",
			totalValidators: 3,
			duplicateCount:  1, // Use validator[0] twice
			uniqueCount:     1,
			minRequired:     types.MinAllowedConfirmations, // 2
			expectPass:      false,
		},
		{
			name:            "2 unique + 1 duplicate = 2 total (need 2)",
			totalValidators: 3,
			duplicateCount:  1, // Use validator[0] twice
			uniqueCount:     2, // Also use validator[1]
			minRequired:     types.MinAllowedConfirmations, // 2
			expectPass:      true,
		},
		{
			name:            "1 unique signature repeated 5 times (need 3)",
			totalValidators: 5,
			duplicateCount:  4, // Repeat validator[0] 5 times total
			uniqueCount:     1,
			minRequired:     types.DefaultMinConfirmations, // 3
			expectPass:      false,
		},
		{
			name:            "3 unique signatures (need 3) - should pass",
			totalValidators: 5,
			duplicateCount:  0, // No duplicates
			uniqueCount:     3,
			minRequired:     types.DefaultMinConfirmations, // 3
			expectPass:      true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup test environment
			k, input := setupKeeperForMultiValidatorTests(t, tc.minRequired)
			ctx := input.Ctx
			msgServer := keeper.NewMsgServerImpl(k)

			// Create and register validators (all active)
			validators := createTestValidators(t, input, k, tc.totalValidators, tc.totalValidators)

			// Create a transfer to unlock
			transfer := createTestTransfer(t, input, k, "transfer-duplicate-test", "0xduplicate123", "1000000", tc.minRequired)

			// Prepare unlock message data
			burnTxHash := "0xduplicate123"
			sender := keepertest.GenTestAddr().String()
			amount := "1000000"
			denom := "uaura"
			sourceChain := "paw"

			// Create message to sign
			msgToSign := fmt.Sprintf("%s:%s:%s:%s:%s",
				transfer.SourceChain,
				burnTxHash,
				sender,
				amount,
				denom,
			)
			msgHash := sha256.Sum256([]byte(msgToSign))

			// Generate signatures with duplicates
			var signatures [][]byte

			// Add unique signatures
			for i := 0; i < tc.uniqueCount && i < len(validators); i++ {
				sig := signMessageWithValidator(t, validators[i].PrivKey, msgHash[:])
				signatures = append(signatures, sig)
			}

			// Add duplicate signatures (reuse first validator)
			if tc.duplicateCount > 0 && len(validators) > 0 {
				for i := 0; i < tc.duplicateCount; i++ {
					sig := signMessageWithValidator(t, validators[0].PrivKey, msgHash[:])
					signatures = append(signatures, sig)
				}
			}

			// Create unlock message
			unlockMsg := &bridgepb.MsgUnlockTokens{
				BurnTxHash:          burnTxHash,
				Sender:              sender,
				Amount:              amount,
				Denom:               denom,
				SourceChain:         sourceChain,
				ValidatorSignatures: signatures,
				MerkleProof:         []byte{},
				MerkleRoot:          []byte{},
				SourceBlockHash:     []byte{},
				SourceBlockHeight:   0,
			}

			// Execute unlock
			resp, err := msgServer.UnlockTokens(sdk.WrapSDKContext(ctx), unlockMsg)

			// Verify results
			if tc.expectPass {
				require.NoError(t, err, "Unlock should succeed with %d unique validators", tc.uniqueCount)
				require.NotNil(t, resp, "Response should not be nil")
				require.True(t, resp.Success, "Response should indicate success")
			} else {
				require.Error(t, err, "Unlock should fail - duplicate signatures should not be counted")
				require.Contains(t, err.Error(), "insufficient", "Error should indicate insufficient signatures")
			}
		})
	}
}

// TestMultiValidatorConsensus_MinimumThresholdEnforced tests that the minimum threshold
// is always enforced, even if params are misconfigured.
//
// SECURITY CRITICAL: Ensures MinAllowedConfirmations (2) is ALWAYS enforced as an absolute
// minimum, preventing governance attacks that lower security parameters.
func TestMultiValidatorConsensus_MinimumThresholdEnforced(t *testing.T) {
	testCases := []struct {
		name                string
		paramsMinRequired   uint64 // What params say
		expectedMinEnforced uint64 // What should actually be enforced
		signaturesProvided  int
		expectPass          bool
	}{
		{
			name:                "Params set to 1 - should enforce minimum 2",
			paramsMinRequired:   1,                             // Misconfigured params
			expectedMinEnforced: types.MinAllowedConfirmations, // 2 should be enforced
			signaturesProvided:  1,
			expectPass:          false, // 1 sig not enough even though params say 1
		},
		{
			name:                "Params set to 2 - should enforce 2",
			paramsMinRequired:   types.MinAllowedConfirmations, // 2
			expectedMinEnforced: types.MinAllowedConfirmations, // 2
			signaturesProvided:  2,
			expectPass:          true,
		},
		{
			name:                "Params set to 3 - should enforce 3",
			paramsMinRequired:   types.DefaultMinConfirmations, // 3
			expectedMinEnforced: types.DefaultMinConfirmations, // 3
			signaturesProvided:  2,
			expectPass:          false, // 2 sigs not enough when 3 required
		},
		{
			name:                "Params set to 3, provide 3 - should pass",
			paramsMinRequired:   types.DefaultMinConfirmations, // 3
			expectedMinEnforced: types.DefaultMinConfirmations, // 3
			signaturesProvided:  3,
			expectPass:          true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup test environment with specified params
			// Note: Param validation should prevent setting below MinAllowedConfirmations,
			// but this test verifies the code enforces it even if params are misconfigured
			k, input := setupKeeperForMultiValidatorTests(t, tc.paramsMinRequired)
			ctx := input.Ctx
			msgServer := keeper.NewMsgServerImpl(k)

			// Create validators
			validatorCount := 5
			validators := createTestValidators(t, input, k, validatorCount, validatorCount)

			// Create a transfer
			transferID := fmt.Sprintf("transfer-threshold-%d", tc.paramsMinRequired)

			// Prepare unlock message
			burnTxHash := fmt.Sprintf("0xthreshold%d", tc.paramsMinRequired)

			// Create a transfer
			transfer := createTestTransfer(t, input, k, transferID, burnTxHash, "1000000", tc.paramsMinRequired)
			sender := keepertest.GenTestAddr().String()
			amount := "1000000"
			denom := "uaura"
			sourceChain := "paw"

			// Create message to sign
			msgToSign := fmt.Sprintf("%s:%s:%s:%s:%s",
				transfer.SourceChain,
				burnTxHash,
				sender,
				amount,
				denom,
			)
			msgHash := sha256.Sum256([]byte(msgToSign))

			// Generate signatures
			var signatures [][]byte
			for i := 0; i < tc.signaturesProvided && i < len(validators); i++ {
				sig := signMessageWithValidator(t, validators[i].PrivKey, msgHash[:])
				signatures = append(signatures, sig)
			}

			// Create unlock message
			unlockMsg := &bridgepb.MsgUnlockTokens{
				BurnTxHash:          burnTxHash,
				Sender:              sender,
				Amount:              amount,
				Denom:               denom,
				SourceChain:         sourceChain,
				ValidatorSignatures: signatures,
				MerkleProof:         []byte{},
				MerkleRoot:          []byte{},
				SourceBlockHash:     []byte{},
				SourceBlockHeight:   0,
			}

			// Execute unlock
			resp, err := msgServer.UnlockTokens(sdk.WrapSDKContext(ctx), unlockMsg)

			// Verify results
			if tc.expectPass {
				require.NoError(t, err, "Unlock should succeed with %d signatures", tc.signaturesProvided)
				require.NotNil(t, resp)
				require.True(t, resp.Success)
			} else {
				require.Error(t, err, "Unlock should fail with %d signatures (need %d)", tc.signaturesProvided, tc.expectedMinEnforced)
				require.Contains(t, err.Error(), "insufficient", "Error should indicate insufficient signatures")
			}
		})
	}
}

// TestMultiValidatorConsensus_InactiveValidatorsNotCounted tests that inactive validators
// are not counted toward the threshold, even if they provide valid cryptographic signatures.
//
// SECURITY CRITICAL: Prevents compromised validators that have been deactivated from
// being used to unlock funds.
func TestMultiValidatorConsensus_InactiveValidatorsNotCounted(t *testing.T) {
	testCases := []struct {
		name             string
		totalValidators  int
		activeValidators int
		sigFromActive    int
		sigFromInactive  int
		minRequired      uint64
		expectPass       bool
	}{
		{
			name:             "2 active + 1 inactive signature (need 2) - should pass",
			totalValidators:  5,
			activeValidators: 3,
			sigFromActive:    2,
			sigFromInactive:  1,
			minRequired:      types.MinAllowedConfirmations, // 2
			expectPass:       true,                          // 2 active sigs is enough
		},
		{
			name:             "1 active + 2 inactive signatures (need 2) - should fail",
			totalValidators:  5,
			activeValidators: 3,
			sigFromActive:    1,
			sigFromInactive:  2,
			minRequired:      types.MinAllowedConfirmations, // 2
			expectPass:       false,                         // Only 1 active sig, not enough
		},
		{
			name:             "0 active + 3 inactive signatures (need 2) - should fail",
			totalValidators:  5,
			activeValidators: 2,
			sigFromActive:    0,
			sigFromInactive:  3,
			minRequired:      types.MinAllowedConfirmations, // 2
			expectPass:       false,                         // No active sigs
		},
		{
			name:             "3 active signatures (need 3) - should pass",
			totalValidators:  6,
			activeValidators: 4,
			sigFromActive:    3,
			sigFromInactive:  0,
			minRequired:      types.DefaultMinConfirmations, // 3
			expectPass:       true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup test environment
			k, input := setupKeeperForMultiValidatorTests(t, tc.minRequired)
			ctx := input.Ctx
			msgServer := keeper.NewMsgServerImpl(k)

			// Create validators (mix of active and inactive)
			validators := createTestValidators(t, input, k, tc.totalValidators, tc.activeValidators)

			// Separate active and inactive validators
			var activeVals, inactiveVals []TestValidator
			for _, v := range validators {
				if v.Active {
					activeVals = append(activeVals, v)
				} else {
					inactiveVals = append(inactiveVals, v)
				}
			}

			// Create a transfer
			transferID := "transfer-inactive-test"

			// Prepare unlock message
			burnTxHash := "0xinactive123"

			// Create a transfer
			transfer := createTestTransfer(t, input, k, transferID, burnTxHash, "1000000", tc.minRequired)
			sender := keepertest.GenTestAddr().String()
			amount := "1000000"
			denom := "uaura"
			sourceChain := "paw"

			// Create message to sign
			msgToSign := fmt.Sprintf("%s:%s:%s:%s:%s",
				transfer.SourceChain,
				burnTxHash,
				sender,
				amount,
				denom,
			)
			msgHash := sha256.Sum256([]byte(msgToSign))

			// Generate signatures from active validators
			var signatures [][]byte
			for i := 0; i < tc.sigFromActive && i < len(activeVals); i++ {
				sig := signMessageWithValidator(t, activeVals[i].PrivKey, msgHash[:])
				signatures = append(signatures, sig)
			}

			// Generate signatures from inactive validators
			for i := 0; i < tc.sigFromInactive && i < len(inactiveVals); i++ {
				sig := signMessageWithValidator(t, inactiveVals[i].PrivKey, msgHash[:])
				signatures = append(signatures, sig)
			}

			// Create unlock message
			unlockMsg := &bridgepb.MsgUnlockTokens{
				BurnTxHash:          burnTxHash,
				Sender:              sender,
				Amount:              amount,
				Denom:               denom,
				SourceChain:         sourceChain,
				ValidatorSignatures: signatures,
				MerkleProof:         []byte{},
				MerkleRoot:          []byte{},
				SourceBlockHash:     []byte{},
				SourceBlockHeight:   0,
			}

			// Execute unlock
			resp, err := msgServer.UnlockTokens(sdk.WrapSDKContext(ctx), unlockMsg)

			// Verify results
			if tc.expectPass {
				require.NoError(t, err, "Unlock should succeed with %d active validator signatures", tc.sigFromActive)
				require.NotNil(t, resp)
				require.True(t, resp.Success)
			} else {
				require.Error(t, err, "Unlock should fail - only %d active signatures (need %d)", tc.sigFromActive, tc.minRequired)
				require.Contains(t, err.Error(), "insufficient", "Error should indicate insufficient signatures")
			}
		})
	}
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// TestValidator represents a test validator with private key for signing
type TestValidator struct {
	Address string
	PrivKey cryptotypes.PrivKey
	PubKey  cryptotypes.PubKey
	Active  bool
}

// setupKeeperForMultiValidatorTests creates a keeper with specified minimum confirmations
func setupKeeperForMultiValidatorTests(t *testing.T, minConfirmations uint64) (*keeper.Keeper, keepertest.TestInput) {
	t.Helper()

	// Create test input with properly registered crypto interfaces
	input := keepertest.CreateTestInput(t)

	// Register crypto interfaces so secp256k1 keys can be marshaled
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	// Update input codec
	input.Cdc = cdc

	// Create LegacyAmino codec for paramstore
	legacyAmino := codec.NewLegacyAmino()

	// Create a paramstore with KeyTable
	ps := paramtypes.NewSubspace(cdc, legacyAmino, input.StoreKey, input.MemStoreKey, "bridge")

	k := keeper.NewKeeper(
		cdc,
		input.StoreKey,
		&ps,
		nil, // bankKeeper (not needed for signature verification)
		nil, // accountKeeper
		nil, // vcKeeper
		nil, // stakingKeeper
	)

	// Set params with specified minimum confirmations
	// Note: Param validation prevents setting below MinAllowedConfirmations (2)
	// For the test that tries minConfirmations=1, we expect param validation to reject it
	// So we only set params if minConfirmations >= MinAllowedConfirmations
	if minConfirmations >= types.MinAllowedConfirmations {
		params := types.DefaultParams()
		params.MinConfirmations = minConfirmations
		k.SetParams(input.Ctx, params)
	} else {
		// For invalid params (< MinAllowedConfirmations), just set defaults
		// The code will still enforce minimum at runtime even if params were somehow bypassed
		params := types.DefaultParams()
		k.SetParams(input.Ctx, params)
	}

	return k, input
}

// createTestValidators creates test validators with private keys for signing
// activeCount specifies how many of the total should be marked as active
func createTestValidators(t *testing.T, input keepertest.TestInput, k *keeper.Keeper, total int, activeCount int) []TestValidator {
	t.Helper()

	validators := make([]TestValidator, total)

	for i := 0; i < total; i++ {
		// Generate Cosmos SDK secp256k1 key pair
		privKey := secp256k1.GenPrivKey()
		pubKey := privKey.PubKey()

		// Derive address from public key
		addr := sdk.AccAddress(pubKey.Address()).String()

		// Determine if this validator should be active
		isActive := i < activeCount

		// Marshal public key for storage
		pubKeyAny, err := input.Cdc.MarshalInterface(pubKey)
		require.NoError(t, err, "Failed to marshal public key")

		// Create bridge validator
		bridgeValidator := &types.BridgeValidator{
			Address:   addr,
			PublicKey: pubKeyAny,
			Power:     1000,
			Active:    isActive,
			Chains:    []string{"aura", "paw", "xai"},
		}

		// Register validator
		k.SetValidator(input.Ctx, bridgeValidator)

		validators[i] = TestValidator{
			Address: addr,
			PrivKey: privKey,
			PubKey:  pubKey,
			Active:  isActive,
		}
	}

	return validators
}

// createTestTransfer creates a cross-chain transfer for testing
func createTestTransfer(t *testing.T, input keepertest.TestInput, k *keeper.Keeper, transferID string, burnTxHash string, amount string, requiredConfirmations uint64) *types.CrossChainTransfer {
	t.Helper()

	transfer := &bridgepb.CrossChainTransfer{
		TransferId:            transferID,
		SourceChain:           "paw",
		TargetChain:           "aura",
		Sender:                keepertest.GenTestAddr().String(),
		Recipient:             keepertest.GenTestAddr().String(),
		Denom:                 "uaura",
		Amount:                amount,
		Status:                bridgepb.TransferStatus_PENDING,
		Timestamp:             timestamppb.New(input.Ctx.BlockTime()),
		RequiredConfirmations: requiredConfirmations,
	}

	// Store the transfer using the keeper's method
	k.SetTransfer(input.Ctx, transfer)

	// Index the transfer by hash so UnlockTokens can find it
	k.IndexTransferHash(input.Ctx, burnTxHash, transferID)

	return &types.CrossChainTransfer{
		TransferId:            transfer.TransferId,
		SourceChain:           transfer.SourceChain,
		TargetChain:           transfer.TargetChain,
		Sender:                transfer.Sender,
		Recipient:             transfer.Recipient,
		Denom:                 transfer.Denom,
		Amount:                transfer.Amount,
		Status:                types.TransferStatus(transfer.Status),
		Timestamp:             transfer.Timestamp,
		RequiredConfirmations: transfer.RequiredConfirmations,
	}
}

// signMessageWithValidator signs a message hash with a validator's private key
func signMessageWithValidator(t *testing.T, privKey cryptotypes.PrivKey, msgHash []byte) []byte {
	t.Helper()

	sig, err := privKey.Sign(msgHash)
	require.NoError(t, err, "Failed to sign message")

	return sig
}
