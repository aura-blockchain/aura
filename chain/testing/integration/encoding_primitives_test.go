// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package integration_test

import (
	"encoding/json"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// EncodingPrimitivesTestSuite verifies that encoding/decoding of key data structures
// works correctly. This includes:
// - Block marshaling/unmarshaling
// - Transaction marshaling/unmarshaling
// - Message encoding/decoding
// - Proto serialization
// - JSON encoding/decoding
//
// These tests catch serialization bugs that could cause consensus failures.
type EncodingPrimitivesTestSuite struct {
	suite.Suite
	cdc      codec.Codec
	txConfig client.TxConfig
}

func (s *EncodingPrimitivesTestSuite) SetupSuite() {
	// Initialize encoding configuration using testutil from cosmos
	// Register bank module to enable MsgSend encoding/decoding
	encCfg := testutil.MakeTestEncodingConfig(bank.AppModuleBasic{})
	s.cdc = encCfg.Codec
	s.txConfig = encCfg.TxConfig
}

func TestEncodingPrimitivesTestSuite(t *testing.T) {
	suite.Run(t, new(EncodingPrimitivesTestSuite))
}

// TestTransactionEncoding tests transaction marshaling and unmarshaling
func (s *EncodingPrimitivesTestSuite) TestTransactionEncoding() {
	// Create a sample transaction
	msg := banktypes.NewMsgSend(
		sdk.AccAddress("sender"),
		sdk.AccAddress("recipient"),
		sdk.NewCoins(sdk.NewInt64Coin("uaura", 1000)),
	)

	// Create transaction builder
	txBuilder := s.txConfig.NewTxBuilder()
	err := txBuilder.SetMsgs(msg)
	require.NoError(s.T(), err, "SetMsgs should not error")

	txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewInt64Coin("uaura", 100)))
	txBuilder.SetGasLimit(200000)
	txBuilder.SetMemo("test transaction encoding")

	// Get transaction bytes
	signedTx := txBuilder.GetTx()
	txBytes, err := s.txConfig.TxEncoder()(signedTx)
	require.NoError(s.T(), err, "TxEncoder should not error")
	require.NotEmpty(s.T(), txBytes, "encoded transaction should not be empty")

	// Decode transaction
	decodedTx, err := s.txConfig.TxDecoder()(txBytes)
	require.NoError(s.T(), err, "TxDecoder should not error")
	require.NotNil(s.T(), decodedTx, "decoded transaction should not be nil")

	// Verify decoded transaction matches original
	decodedMsgs := decodedTx.GetMsgs()
	require.Equal(s.T(), 1, len(decodedMsgs), "should have one message")

	decodedMsg, ok := decodedMsgs[0].(*banktypes.MsgSend)
	require.True(s.T(), ok, "decoded message should be MsgSend")
	require.Equal(s.T(), msg.FromAddress, decodedMsg.FromAddress, "from address should match")
	require.Equal(s.T(), msg.ToAddress, decodedMsg.ToAddress, "to address should match")
	require.Equal(s.T(), msg.Amount, decodedMsg.Amount, "amount should match")
}

// TestTransactionJSONEncoding tests transaction JSON marshaling
func (s *EncodingPrimitivesTestSuite) TestTransactionJSONEncoding() {
	msg := banktypes.NewMsgSend(
		sdk.AccAddress("sender123456"),
		sdk.AccAddress("recipient890"),
		sdk.NewCoins(sdk.NewInt64Coin("uaura", 5000)),
	)

	txBuilder := s.txConfig.NewTxBuilder()
	err := txBuilder.SetMsgs(msg)
	require.NoError(s.T(), err)

	signedTx := txBuilder.GetTx()

	// Encode to JSON
	txJSON, err := s.txConfig.TxJSONEncoder()(signedTx)
	require.NoError(s.T(), err, "TxJSONEncoder should not error")
	require.NotEmpty(s.T(), txJSON, "JSON should not be empty")

	// Verify it's valid JSON
	var jsonMap map[string]interface{}
	err = json.Unmarshal(txJSON, &jsonMap)
	require.NoError(s.T(), err, "should be valid JSON")

	// Decode from JSON
	decodedTx, err := s.txConfig.TxJSONDecoder()(txJSON)
	require.NoError(s.T(), err, "TxJSONDecoder should not error")
	require.NotNil(s.T(), decodedTx, "decoded transaction should not be nil")
}

// TestBlockEncoding tests CometBFT block marshaling
func (s *EncodingPrimitivesTestSuite) TestBlockEncoding() {
	// Create a sample block
	block := cmtproto.Block{
		Header: cmtproto.Header{
			ChainID: "aura-test-1",
			Height:  100,
			Time:    time.Now(),
		},
		Data: cmtproto.Data{
			Txs: [][]byte{
				[]byte("tx1"),
				[]byte("tx2"),
			},
		},
	}

	// Marshal block to bytes
	blockBytes, err := block.Marshal()
	require.NoError(s.T(), err, "block Marshal should not error")
	require.NotEmpty(s.T(), blockBytes, "marshaled block should not be empty")

	// Unmarshal block
	var decodedBlock cmtproto.Block
	err = decodedBlock.Unmarshal(blockBytes)
	require.NoError(s.T(), err, "block Unmarshal should not error")

	// Verify decoded block matches original
	require.Equal(s.T(), block.Header.ChainID, decodedBlock.Header.ChainID, "chain ID should match")
	require.Equal(s.T(), block.Header.Height, decodedBlock.Header.Height, "height should match")
	require.Equal(s.T(), len(block.Data.Txs), len(decodedBlock.Data.Txs), "tx count should match")
}

// TestMessageEncoding tests individual message encoding
func (s *EncodingPrimitivesTestSuite) TestMessageEncoding() {
	msg := banktypes.NewMsgSend(
		sdk.AccAddress("sender"),
		sdk.AccAddress("receiver"),
		sdk.NewCoins(sdk.NewInt64Coin("uaura", 2000)),
	)

	// Marshal to binary
	msgBytes, err := s.cdc.Marshal(msg)
	require.NoError(s.T(), err, "message Marshal should not error")
	require.NotEmpty(s.T(), msgBytes, "marshaled message should not be empty")

	// Unmarshal from binary
	var decodedMsg banktypes.MsgSend
	err = s.cdc.Unmarshal(msgBytes, &decodedMsg)
	require.NoError(s.T(), err, "message Unmarshal should not error")

	// Verify decoded message matches original
	require.Equal(s.T(), msg.FromAddress, decodedMsg.FromAddress, "from address should match")
	require.Equal(s.T(), msg.ToAddress, decodedMsg.ToAddress, "to address should match")
	require.Equal(s.T(), msg.Amount, decodedMsg.Amount, "amount should match")
}

// TestMessageJSONEncoding tests message JSON encoding
func (s *EncodingPrimitivesTestSuite) TestMessageJSONEncoding() {
	msg := banktypes.NewMsgSend(
		sdk.AccAddress("sender"),
		sdk.AccAddress("receiver"),
		sdk.NewCoins(sdk.NewInt64Coin("uaura", 3000)),
	)

	// Marshal to JSON
	msgJSON, err := s.cdc.MarshalJSON(msg)
	require.NoError(s.T(), err, "MarshalJSON should not error")
	require.NotEmpty(s.T(), msgJSON, "JSON should not be empty")

	// Verify valid JSON
	var jsonMap map[string]interface{}
	err = json.Unmarshal(msgJSON, &jsonMap)
	require.NoError(s.T(), err, "should be valid JSON")

	// Unmarshal from JSON
	var decodedMsg banktypes.MsgSend
	err = s.cdc.UnmarshalJSON(msgJSON, &decodedMsg)
	require.NoError(s.T(), err, "UnmarshalJSON should not error")
	require.Equal(s.T(), msg.FromAddress, decodedMsg.FromAddress)
	require.Equal(s.T(), msg.ToAddress, decodedMsg.ToAddress)
}

// TestResponseEncoding tests ABCI response encoding
func (s *EncodingPrimitivesTestSuite) TestResponseEncoding() {
	// Create sample ResponseDeliverTx (now ExecTxResult in newer versions)
	response := abci.ExecTxResult{
		Code:      0,
		Data:      []byte("success"),
		Log:       "transaction executed successfully",
		Info:      "test info",
		GasWanted: 200000,
		GasUsed:   150000,
	}

	// Marshal to binary
	respBytes, err := response.Marshal()
	require.NoError(s.T(), err, "response Marshal should not error")
	require.NotEmpty(s.T(), respBytes, "marshaled response should not be empty")

	// Unmarshal from binary
	var decodedResp abci.ExecTxResult
	err = decodedResp.Unmarshal(respBytes)
	require.NoError(s.T(), err, "response Unmarshal should not error")

	// Verify decoded response matches original
	require.Equal(s.T(), response.Code, decodedResp.Code)
	require.Equal(s.T(), response.Data, decodedResp.Data)
	require.Equal(s.T(), response.Log, decodedResp.Log)
	require.Equal(s.T(), response.GasWanted, decodedResp.GasWanted)
	require.Equal(s.T(), response.GasUsed, decodedResp.GasUsed)
}

// TestAnyEncoding tests google.protobuf.Any encoding (used for extensibility)
func (s *EncodingPrimitivesTestSuite) TestAnyEncoding() {
	msg := banktypes.NewMsgSend(
		sdk.AccAddress("sender"),
		sdk.AccAddress("receiver"),
		sdk.NewCoins(sdk.NewInt64Coin("uaura", 4000)),
	)

	// Pack into Any
	anyMsg, err := types.NewAnyWithValue(msg)
	require.NoError(s.T(), err, "NewAnyWithValue should not error")
	require.NotNil(s.T(), anyMsg, "Any should not be nil")

	// Marshal Any to binary
	anyBytes, err := s.cdc.Marshal(anyMsg)
	require.NoError(s.T(), err, "Any Marshal should not error")
	require.NotEmpty(s.T(), anyBytes, "marshaled Any should not be empty")

	// Unmarshal Any from binary
	var decodedAny types.Any
	err = s.cdc.Unmarshal(anyBytes, &decodedAny)
	require.NoError(s.T(), err, "Any Unmarshal should not error")

	// Unpack from Any using interface registry
	var unpackedMsg sdk.Msg
	err = s.cdc.InterfaceRegistry().UnpackAny(&decodedAny, &unpackedMsg)
	require.NoError(s.T(), err, "UnpackAny should not error")
	require.NotNil(s.T(), unpackedMsg, "unpacked message should not be nil")

	// Type assert to MsgSend
	decodedMsg, ok := unpackedMsg.(*banktypes.MsgSend)
	require.True(s.T(), ok, "unpacked message should be MsgSend")
	require.Equal(s.T(), msg.FromAddress, decodedMsg.FromAddress)
}

// TestEncodingConsistency verifies encoding is deterministic
func (s *EncodingPrimitivesTestSuite) TestEncodingConsistency() {
	msg := banktypes.NewMsgSend(
		sdk.AccAddress("consistent_sender"),
		sdk.AccAddress("consistent_receiver"),
		sdk.NewCoins(sdk.NewInt64Coin("uaura", 1234)),
	)

	// Encode multiple times
	bytes1, err1 := s.cdc.Marshal(msg)
	bytes2, err2 := s.cdc.Marshal(msg)
	bytes3, err3 := s.cdc.Marshal(msg)

	require.NoError(s.T(), err1)
	require.NoError(s.T(), err2)
	require.NoError(s.T(), err3)

	// All encodings should be identical (deterministic)
	require.Equal(s.T(), bytes1, bytes2, "encoding should be deterministic")
	require.Equal(s.T(), bytes2, bytes3, "encoding should be deterministic")
}

// TestLargeDataStructure tests encoding of larger structures
func (s *EncodingPrimitivesTestSuite) TestLargeDataStructure() {
	// Create a block with many transactions
	txs := make([][]byte, 100)
	for i := 0; i < 100; i++ {
		txs[i] = []byte{byte(i)}
	}

	block := cmtproto.Block{
		Header: cmtproto.Header{
			ChainID: "aura-test-1",
			Height:  1000,
		},
		Data: cmtproto.Data{
			Txs: txs,
		},
	}

	// Marshal large block
	blockBytes, err := block.Marshal()
	require.NoError(s.T(), err, "large block Marshal should not error")
	require.NotEmpty(s.T(), blockBytes, "marshaled block should not be empty")

	// Unmarshal large block
	var decodedBlock cmtproto.Block
	err = decodedBlock.Unmarshal(blockBytes)
	require.NoError(s.T(), err, "large block Unmarshal should not error")
	require.Equal(s.T(), 100, len(decodedBlock.Data.Txs), "tx count should match")
}
