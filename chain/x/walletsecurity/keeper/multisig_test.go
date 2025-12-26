// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"
	"time"

	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/suite"

	wsproto "github.com/aequitas/aura/proto/aura/walletsecurity/v1beta1"
)

type MultiSigTestSuite struct {
	KeeperTestSuite
}

func TestMultiSigTestSuite(t *testing.T) {
	suite.Run(t, new(MultiSigTestSuite))
}

func (suite *MultiSigTestSuite) TestCreateMultiSigWallet() {
	tests := []struct {
		name            string
		creator         string
		signers         []string
		threshold       int32
		signerWeights   map[string]int32
		weightThreshold int32
		timeLock        *gogotypes.Duration
		wantErr         bool
		errContains     string
	}{
		{
			name:      "valid basic multisig",
			creator:   "aura1creator",
			signers:   []string{"aura1signer1", "aura1signer2", "aura1signer3"},
			threshold: 2,
			wantErr:   false,
		},
		{
			name:      "empty signers fails",
			creator:   "aura1creator",
			signers:   []string{},
			threshold: 1,
			wantErr:   true,
		},
		{
			name:        "empty creator fails",
			creator:     "",
			signers:     []string{"aura1signer1"},
			threshold:   1,
			wantErr:     true,
			errContains: "invalid",
		},
		{
			name:      "threshold too high fails",
			creator:   "aura1creator",
			signers:   []string{"aura1signer1", "aura1signer2"},
			threshold: 5,
			wantErr:   true,
		},
		{
			name:      "zero threshold fails",
			creator:   "aura1creator",
			signers:   []string{"aura1signer1", "aura1signer2"},
			threshold: 0,
			wantErr:   true,
		},
		{
			name:      "negative threshold fails",
			creator:   "aura1creator",
			signers:   []string{"aura1signer1", "aura1signer2"},
			threshold: -1,
			wantErr:   true,
		},
		{
			name:    "weighted multisig with valid weights",
			creator: "aura1creator",
			signers: []string{"aura1signer1", "aura1signer2", "aura1signer3"},
			signerWeights: map[string]int32{
				"aura1signer1": 3,
				"aura1signer2": 2,
				"aura1signer3": 1,
			},
			weightThreshold: 4,
			threshold:       2,
			wantErr:         false,
		},
		{
			name:    "weighted multisig with zero weight fails",
			creator: "aura1creator",
			signers: []string{"aura1signer1", "aura1signer2"},
			signerWeights: map[string]int32{
				"aura1signer1": 0,
				"aura1signer2": 2,
			},
			weightThreshold: 2,
			threshold:       1,
			wantErr:         true,
		},
		{
			name:    "weighted multisig with threshold > total weight fails",
			creator: "aura1creator",
			signers: []string{"aura1signer1", "aura1signer2"},
			signerWeights: map[string]int32{
				"aura1signer1": 1,
				"aura1signer2": 2,
			},
			weightThreshold: 10,
			threshold:       1,
			wantErr:         true,
		},
		{
			name:      "time locked multisig",
			creator:   "aura1creator",
			signers:   []string{"aura1signer1", "aura1signer2"},
			threshold: 2,
			timeLock:  &gogotypes.Duration{Seconds: 86400}, // 24 hours
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()
			k := suite.GetKeeper()
			ctx := suite.GetContext()

			wallet, err := k.CreateMultiSigWallet(
				ctx,
				tt.creator,
				tt.signers,
				tt.threshold,
				tt.signerWeights,
				tt.weightThreshold,
				tt.timeLock,
			)

			if tt.wantErr {
				suite.Require().Error(err)
				if tt.errContains != "" {
					suite.Require().Contains(err.Error(), tt.errContains)
				}
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(wallet)
				suite.Require().NotEmpty(wallet.WalletId)
				suite.Require().Equal(tt.creator, wallet.Creator)
				suite.Require().Equal(tt.signers, wallet.Signers)
				suite.Require().Equal(tt.threshold, wallet.Threshold)
				suite.Require().Equal(int32(len(tt.signers)), wallet.TotalSigners)

				if tt.timeLock != nil {
					suite.Require().True(wallet.TimeLocked)
				}
			}
		})
	}
}

func (suite *MultiSigTestSuite) TestCreatePendingMultiSigTransaction() {
	suite.SetupTest()
	k := suite.GetKeeper()
	ctx := suite.GetContext()

	// First create a wallet
	wallet, err := k.CreateMultiSigWallet(
		ctx,
		"aura1creator",
		[]string{"aura1signer1", "aura1signer2", "aura1signer3"},
		2,
		nil,
		0,
		nil,
	)
	suite.Require().NoError(err)

	tests := []struct {
		name        string
		walletID    string
		txData      []byte
		txType      string
		description string
		expiration  time.Duration
		wantErr     bool
	}{
		{
			name:        "valid pending transaction",
			walletID:    wallet.WalletId,
			txData:      []byte("transfer 100 tokens to alice"),
			txType:      "transfer",
			description: "Send tokens to Alice",
			expiration:  24 * time.Hour,
			wantErr:     false,
		},
		{
			name:        "non-existent wallet fails",
			walletID:    "invalid_wallet_id",
			txData:      []byte("transfer"),
			txType:      "transfer",
			description: "Test",
			expiration:  time.Hour,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			pendingTx, err := k.CreatePendingMultiSigTransaction(
				ctx,
				tt.walletID,
				tt.txData,
				tt.txType,
				tt.description,
				tt.expiration,
			)

			if tt.wantErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(pendingTx)
				suite.Require().NotEmpty(pendingTx.TxId)
				suite.Require().Equal(tt.walletID, pendingTx.WalletId)
				suite.Require().Equal(tt.txData, pendingTx.TxData)
				suite.Require().Equal(tt.txType, pendingTx.TxType)
				suite.Require().Equal(tt.description, pendingTx.Description)
				suite.Require().Len(pendingTx.SignedBy, 0)
			}
		})
	}
}

func (suite *MultiSigTestSuite) TestSignMultiSigTransaction() {
	suite.SetupTest()
	k := suite.GetKeeper()
	ctx := suite.GetContext()

	// Create wallet
	signers := []string{"aura1signer1", "aura1signer2", "aura1signer3"}
	wallet, err := k.CreateMultiSigWallet(ctx, "aura1creator", signers, 2, nil, 0, nil)
	suite.Require().NoError(err)

	// Create pending transaction
	pendingTx, err := k.CreatePendingMultiSigTransaction(
		ctx, wallet.WalletId, []byte("test tx"), "transfer", "Test", 24*time.Hour,
	)
	suite.Require().NoError(err)

	// Valid signature (minimum 64 bytes)
	validSig := make([]byte, 64)
	for i := range validSig {
		validSig[i] = byte(i)
	}

	tests := []struct {
		name            string
		txID            string
		signer          string
		signature       []byte
		wantErr         bool
		wantReady       bool
		errContains     string
	}{
		{
			name:      "first valid signature",
			txID:      pendingTx.TxId,
			signer:    "aura1signer1",
			signature: validSig,
			wantErr:   false,
			wantReady: false, // Need 2 signatures
		},
		{
			name:      "second valid signature reaches threshold",
			txID:      pendingTx.TxId,
			signer:    "aura1signer2",
			signature: validSig,
			wantErr:   false,
			wantReady: true, // Now have 2 signatures
		},
		{
			name:        "unauthorized signer fails",
			txID:        pendingTx.TxId,
			signer:      "aura1unauthorized",
			signature:   validSig,
			wantErr:     true,
			errContains: "invalid signer",
		},
		{
			name:        "invalid tx id fails",
			txID:        "invalid_tx_id",
			signer:      "aura1signer3",
			signature:   validSig,
			wantErr:     true,
		},
		{
			name:        "short signature fails",
			txID:        pendingTx.TxId,
			signer:      "aura1signer3",
			signature:   []byte("short"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			ready, err := k.SignMultiSigTransaction(ctx, tt.txID, tt.signer, tt.signature)

			if tt.wantErr {
				suite.Require().Error(err)
				if tt.errContains != "" {
					suite.Require().Contains(err.Error(), tt.errContains)
				}
			} else {
				suite.Require().NoError(err)
				suite.Require().Equal(tt.wantReady, ready)
			}
		})
	}
}

func (suite *MultiSigTestSuite) TestSignMultiSigTransactionDuplicateSigner() {
	suite.SetupTest()
	k := suite.GetKeeper()
	ctx := suite.GetContext()

	wallet, _ := k.CreateMultiSigWallet(ctx, "aura1creator", []string{"aura1signer1", "aura1signer2"}, 2, nil, 0, nil)
	pendingTx, _ := k.CreatePendingMultiSigTransaction(ctx, wallet.WalletId, []byte("tx"), "transfer", "Test", time.Hour)

	validSig := make([]byte, 64)

	// First signature succeeds
	_, err := k.SignMultiSigTransaction(ctx, pendingTx.TxId, "aura1signer1", validSig)
	suite.Require().NoError(err)

	// Same signer again fails
	_, err = k.SignMultiSigTransaction(ctx, pendingTx.TxId, "aura1signer1", validSig)
	suite.Require().Error(err)
}

func (suite *MultiSigTestSuite) TestExecuteMultiSigTransaction() {
	suite.SetupTest()
	k := suite.GetKeeper()
	ctx := suite.GetContext()

	// Create wallet with threshold 2
	wallet, _ := k.CreateMultiSigWallet(ctx, "aura1creator", []string{"aura1signer1", "aura1signer2"}, 2, nil, 0, nil)
	pendingTx, _ := k.CreatePendingMultiSigTransaction(ctx, wallet.WalletId, []byte("tx"), "transfer", "Test", time.Hour)

	validSig := make([]byte, 64)

	// Sign with both signers
	k.SignMultiSigTransaction(ctx, pendingTx.TxId, "aura1signer1", validSig)
	k.SignMultiSigTransaction(ctx, pendingTx.TxId, "aura1signer2", validSig)

	// Execute should succeed
	err := k.ExecuteMultiSigTransaction(ctx, pendingTx.TxId)
	suite.Require().NoError(err)

	// Transaction should be deleted
	_, err = k.GetPendingMultiSigTx(ctx, pendingTx.TxId)
	suite.Require().Error(err)
}

func (suite *MultiSigTestSuite) TestExecuteMultiSigTransactionInsufficientSignatures() {
	suite.SetupTest()
	k := suite.GetKeeper()
	ctx := suite.GetContext()

	wallet, _ := k.CreateMultiSigWallet(ctx, "aura1creator", []string{"aura1signer1", "aura1signer2"}, 2, nil, 0, nil)
	pendingTx, _ := k.CreatePendingMultiSigTransaction(ctx, wallet.WalletId, []byte("tx"), "transfer", "Test", time.Hour)

	// Only sign with one signer (need 2)
	validSig := make([]byte, 64)
	k.SignMultiSigTransaction(ctx, pendingTx.TxId, "aura1signer1", validSig)

	// Execute should fail
	err := k.ExecuteMultiSigTransaction(ctx, pendingTx.TxId)
	suite.Require().Error(err)
}

func (suite *MultiSigTestSuite) TestAddSignerToMultiSigWallet() {
	suite.SetupTest()
	k := suite.GetKeeper()
	ctx := suite.GetContext()

	wallet, _ := k.CreateMultiSigWallet(ctx, "aura1creator", []string{"aura1signer1", "aura1signer2"}, 2, nil, 0, nil)

	tests := []struct {
		name      string
		walletID  string
		newSigner string
		requester string
		wantErr   bool
	}{
		{
			name:      "authorized requester adds signer",
			walletID:  wallet.WalletId,
			newSigner: "aura1newsigner",
			requester: "aura1signer1",
			wantErr:   false,
		},
		{
			name:      "unauthorized requester fails",
			walletID:  wallet.WalletId,
			newSigner: "aura1newsigner2",
			requester: "aura1unauthorized",
			wantErr:   true,
		},
		{
			name:      "add existing signer fails",
			walletID:  wallet.WalletId,
			newSigner: "aura1signer2",
			requester: "aura1signer1",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := k.AddSignerToMultiSigWallet(ctx, tt.walletID, tt.newSigner, tt.requester)

			if tt.wantErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

func (suite *MultiSigTestSuite) TestRemoveSignerFromMultiSigWallet() {
	suite.SetupTest()
	k := suite.GetKeeper()
	ctx := suite.GetContext()

	wallet, _ := k.CreateMultiSigWallet(ctx, "aura1creator", []string{"aura1signer1", "aura1signer2", "aura1signer3"}, 2, nil, 0, nil)

	tests := []struct {
		name           string
		signerToRemove string
		requester      string
		wantErr        bool
	}{
		{
			name:           "valid removal",
			signerToRemove: "aura1signer3",
			requester:      "aura1signer1",
			wantErr:        false,
		},
		{
			name:           "unauthorized requester fails",
			signerToRemove: "aura1signer2",
			requester:      "aura1unauthorized",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := k.RemoveSignerFromMultiSigWallet(ctx, wallet.WalletId, tt.signerToRemove, tt.requester)

			if tt.wantErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

func (suite *MultiSigTestSuite) TestRemoveSignerBreaksThreshold() {
	suite.SetupTest()
	k := suite.GetKeeper()
	ctx := suite.GetContext()

	// Create wallet with 2 signers, threshold 2
	wallet, _ := k.CreateMultiSigWallet(ctx, "aura1creator", []string{"aura1signer1", "aura1signer2"}, 2, nil, 0, nil)

	// Removing a signer would make threshold impossible
	err := k.RemoveSignerFromMultiSigWallet(ctx, wallet.WalletId, "aura1signer2", "aura1signer1")
	suite.Require().Error(err)
}

func (suite *MultiSigTestSuite) TestWeightedMultiSigExecution() {
	suite.SetupTest()
	k := suite.GetKeeper()
	ctx := suite.GetContext()

	// Create weighted multisig: weights 3, 2, 1 with threshold 4
	weights := map[string]int32{
		"aura1signer1": 3,
		"aura1signer2": 2,
		"aura1signer3": 1,
	}
	wallet, _ := k.CreateMultiSigWallet(ctx, "aura1creator", []string{"aura1signer1", "aura1signer2", "aura1signer3"}, 2, weights, 4, nil)
	pendingTx, _ := k.CreatePendingMultiSigTransaction(ctx, wallet.WalletId, []byte("tx"), "transfer", "Test", time.Hour)

	validSig := make([]byte, 64)

	// Signer1 (weight 3) - not enough for threshold 4
	ready, err := k.SignMultiSigTransaction(ctx, pendingTx.TxId, "aura1signer1", validSig)
	suite.Require().NoError(err)
	suite.Require().False(ready)

	// Signer3 (weight 1) - now have 4, meets threshold
	ready, err = k.SignMultiSigTransaction(ctx, pendingTx.TxId, "aura1signer3", validSig)
	suite.Require().NoError(err)
	suite.Require().True(ready)
}

func (suite *MultiSigTestSuite) TestIsAuthorizedSigner() {
	k := suite.GetKeeper()

	signers := []string{"aura1signer1", "aura1signer2", "aura1signer3"}

	tests := []struct {
		address  string
		expected bool
	}{
		{"aura1signer1", true},
		{"aura1signer2", true},
		{"aura1signer3", true},
		{"aura1unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		result := k.isAuthorizedSigner(tt.address, signers)
		suite.Require().Equal(tt.expected, result, "address: %s", tt.address)
	}
}

func (suite *MultiSigTestSuite) TestHasAlreadySigned() {
	k := suite.GetKeeper()

	signedBy := []string{"aura1signer1", "aura1signer2"}

	tests := []struct {
		address  string
		expected bool
	}{
		{"aura1signer1", true},
		{"aura1signer2", true},
		{"aura1signer3", false},
		{"", false},
	}

	for _, tt := range tests {
		result := k.hasAlreadySigned(tt.address, signedBy)
		suite.Require().Equal(tt.expected, result, "address: %s", tt.address)
	}
}

func (suite *MultiSigTestSuite) TestIsReadyToExecute() {
	k := suite.GetKeeper()

	// Standard multisig
	standardWallet := &wsproto.MultiSigWallet{
		Threshold: 2,
	}
	standardTx := &wsproto.PendingMultiSigTransaction{
		SignedBy: []string{"signer1", "signer2"},
	}
	suite.Require().True(k.isReadyToExecute(standardTx, standardWallet))

	standardTx.SignedBy = []string{"signer1"}
	suite.Require().False(k.isReadyToExecute(standardTx, standardWallet))

	// Weighted multisig
	weightedWallet := &wsproto.MultiSigWallet{
		SignerWeights:   map[string]int32{"signer1": 3, "signer2": 2},
		WeightThreshold: 4,
	}
	weightedTx := &wsproto.PendingMultiSigTransaction{
		CurrentWeight: 5,
	}
	suite.Require().True(k.isReadyToExecute(weightedTx, weightedWallet))

	weightedTx.CurrentWeight = 3
	suite.Require().False(k.isReadyToExecute(weightedTx, weightedWallet))
}
