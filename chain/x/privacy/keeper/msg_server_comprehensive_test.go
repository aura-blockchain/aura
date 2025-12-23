package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	sdkmath "cosmossdk.io/math"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/privacy/keeper"
	"github.com/aequitas/aura/chain/x/privacy/types"
	privacypb "github.com/aequitas/aura/proto/aura/privacy/v1beta1"
)

type MsgServerComprehensiveTestSuite struct {
	suite.Suite

	keeper    *keeper.Keeper
	ctx       sdk.Context
	msgServer privacypb.MsgServer
}

func TestMsgServerComprehensiveTestSuite(t *testing.T) {
	suite.Run(t, new(MsgServerComprehensiveTestSuite))
}

func (suite *MsgServerComprehensiveTestSuite) SetupTest() {
	input := keepertest.CreateTestInput(suite.T())
	suite.keeper = keeper.NewKeeper(
		input.Cdc,
		input.StoreKey,
		nil, // authKeeper
		nil, // bankKeeper
	)
	suite.ctx = input.Ctx
	suite.msgServer = keeper.NewMsgServerImpl(suite.keeper)

	// Set default params with all features enabled
	params := types.DefaultParams()
	params.EnableZkProofs = true
	params.EnableMixing = true
	params.EnableNetworkPrivacy = true
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)
}

// SubmitPrivateTransaction Tests

func (suite *MsgServerComprehensiveTestSuite) TestSubmitPrivateTransaction_NilRequest() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	resp, err := suite.msgServer.SubmitPrivateTransaction(goCtx, nil)

	suite.Require().Error(err)
	suite.Require().Nil(resp)
	suite.Require().Equal(codes.InvalidArgument, status.Code(err))
	suite.Require().Contains(err.Error(), "empty request")
}

func (suite *MsgServerComprehensiveTestSuite) TestSubmitPrivateTransaction_EmptySender() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	msg := &privacypb.MsgSubmitPrivateTransaction{
		Sender: "",
		PrivateTransaction: &privacypb.PrivateTransaction{},
	}

	resp, err := suite.msgServer.SubmitPrivateTransaction(goCtx, msg)

	suite.Require().Error(err)
	suite.Require().Nil(resp)
	suite.Require().Equal(codes.InvalidArgument, status.Code(err))
	suite.Require().Contains(err.Error(), "sender cannot be empty")
}

func (suite *MsgServerComprehensiveTestSuite) TestSubmitPrivateTransaction_NilTransaction() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	msg := &privacypb.MsgSubmitPrivateTransaction{
		Sender:             keepertest.GenTestAddr().String(),
		PrivateTransaction: nil,
	}

	resp, err := suite.msgServer.SubmitPrivateTransaction(goCtx, msg)

	suite.Require().Error(err)
	suite.Require().Nil(resp)
	suite.Require().Equal(codes.InvalidArgument, status.Code(err))
	suite.Require().Contains(err.Error(), "private transaction cannot be nil")
}

func (suite *MsgServerComprehensiveTestSuite) TestSubmitPrivateTransaction_ZkProofNotEnabled() {
	// Disable ZK proofs
	params := suite.keeper.GetParams(suite.ctx)
	params.EnableZkProofs = false
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)

	goCtx := sdk.WrapSDKContext(suite.ctx)

	msg := &privacypb.MsgSubmitPrivateTransaction{
		Sender: keepertest.GenTestAddr().String(),
		PrivateTransaction: &privacypb.PrivateTransaction{
			ZkProof: &privacypb.ZKProof{
				ProofType: "groth16",
				ProofData: []byte("proof_data"),
			},
		},
	}

	resp, err := suite.msgServer.SubmitPrivateTransaction(goCtx, msg)

	suite.Require().Error(err)
	suite.Require().Nil(resp)
	suite.Require().Equal(codes.FailedPrecondition, status.Code(err))
	suite.Require().Contains(err.Error(), "zk proofs not enabled")
}

func (suite *MsgServerComprehensiveTestSuite) TestSubmitPrivateTransaction_Success() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	sender := keepertest.GenTestAddr().String()
	msg := &privacypb.MsgSubmitPrivateTransaction{
		Sender: sender,
		PrivateTransaction: &privacypb.PrivateTransaction{
			ZkProof: &privacypb.ZKProof{
				ProofType: "groth16",
				ProofData: []byte("proof_data"),
			},
		},
	}

	resp, err := suite.msgServer.SubmitPrivateTransaction(goCtx, msg)

	suite.Require().NoError(err)
	suite.Require().NotNil(resp)
	suite.Require().True(resp.Success)
	suite.Require().NotEmpty(resp.TxHash)

	// Verify event was emitted
	events := suite.ctx.EventManager().Events()
	suite.Require().NotEmpty(events)

	found := false
	for _, event := range events {
		if event.Type == types.EventTypePrivateTransaction {
			found = true
			break
		}
	}
	suite.Require().True(found, "Event not emitted")
}

// CreateMixingPool Tests

func (suite *MsgServerComprehensiveTestSuite) TestCreateMixingPool_NilRequest() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	resp, err := suite.msgServer.CreateMixingPool(goCtx, nil)

	suite.Require().Error(err)
	suite.Require().Nil(resp)
	suite.Require().Equal(codes.InvalidArgument, status.Code(err))
	suite.Require().Contains(err.Error(), "empty request")
}

func (suite *MsgServerComprehensiveTestSuite) TestCreateMixingPool_EmptyCreator() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	msg := &privacypb.MsgCreateMixingPool{
		Creator:         "",
		MinParticipants: 2,
		MaxParticipants: 10,
	}

	resp, err := suite.msgServer.CreateMixingPool(goCtx, msg)

	suite.Require().Error(err)
	suite.Require().Nil(resp)
	suite.Require().Equal(codes.InvalidArgument, status.Code(err))
	suite.Require().Contains(err.Error(), "creator cannot be empty")
}

func (suite *MsgServerComprehensiveTestSuite) TestCreateMixingPool_MixingNotEnabled() {
	// Disable mixing
	params := suite.keeper.GetParams(suite.ctx)
	params.EnableMixing = false
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)

	goCtx := sdk.WrapSDKContext(suite.ctx)

	msg := &privacypb.MsgCreateMixingPool{
		Creator:         keepertest.GenTestAddr().String(),
		MinParticipants: 2,
		MaxParticipants: 10,
	}

	resp, err := suite.msgServer.CreateMixingPool(goCtx, msg)

	suite.Require().Error(err)
	suite.Require().Nil(resp)
	suite.Require().Equal(codes.FailedPrecondition, status.Code(err))
	suite.Require().Contains(err.Error(), "mixing not enabled")
}

func (suite *MsgServerComprehensiveTestSuite) TestCreateMixingPool_MinParticipantsTooLow() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	msg := &privacypb.MsgCreateMixingPool{
		Creator:         keepertest.GenTestAddr().String(),
		MinParticipants: 1, // Must be at least 2
		MaxParticipants: 10,
	}

	resp, err := suite.msgServer.CreateMixingPool(goCtx, msg)

	suite.Require().Error(err)
	suite.Require().Nil(resp)
	suite.Require().Equal(codes.InvalidArgument, status.Code(err))
	suite.Require().Contains(err.Error(), "minimum participants must be at least 2")
}

func (suite *MsgServerComprehensiveTestSuite) TestCreateMixingPool_MaxLessThanMin() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	msg := &privacypb.MsgCreateMixingPool{
		Creator:         keepertest.GenTestAddr().String(),
		MinParticipants: 10,
		MaxParticipants: 5, // Less than min
	}

	resp, err := suite.msgServer.CreateMixingPool(goCtx, msg)

	suite.Require().Error(err)
	suite.Require().Nil(resp)
	suite.Require().Equal(codes.InvalidArgument, status.Code(err))
	suite.Require().Contains(err.Error(), "max participants must be >= min participants")
}

func (suite *MsgServerComprehensiveTestSuite) TestCreateMixingPool_Success() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	creator := keepertest.GenTestAddr().String()
	msg := &privacypb.MsgCreateMixingPool{
		Creator:         creator,
		MinParticipants: 2,
		MaxParticipants: 10,
		Denomination:    sdkmath.NewInt(1000000),
		MixingRounds:    3,
	}

	resp, err := suite.msgServer.CreateMixingPool(goCtx, msg)

	suite.Require().NoError(err)
	suite.Require().NotNil(resp)
	suite.Require().NotEmpty(resp.PoolId)

	// Verify pool was stored
	pool, err := suite.keeper.GetMixingPool(suite.ctx, resp.PoolId)
	suite.Require().NoError(err)
	suite.Require().NotNil(pool)
	suite.Require().Equal(msg.MinParticipants, pool.MinParticipants)
	suite.Require().Equal(msg.MaxParticipants, pool.MaxParticipants)
	suite.Require().Equal("pending", pool.Status)

	// Verify event was emitted
	events := suite.ctx.EventManager().Events()
	found := false
	for _, event := range events {
		if event.Type == types.EventTypeMixingPool {
			found = true
			break
		}
	}
	suite.Require().True(found, "Event not emitted")
}

// JoinMixingPool Tests

func (suite *MsgServerComprehensiveTestSuite) TestJoinMixingPool_NilRequest() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	resp, err := suite.msgServer.JoinMixingPool(goCtx, nil)

	suite.Require().Error(err)
	suite.Require().Nil(resp)
	suite.Require().Equal(codes.InvalidArgument, status.Code(err))
	suite.Require().Contains(err.Error(), "empty request")
}

func (suite *MsgServerComprehensiveTestSuite) TestJoinMixingPool_EmptyParticipant() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	msg := &privacypb.MsgJoinMixingPool{
		Participant: "",
		PoolId:      "pool_123",
	}

	resp, err := suite.msgServer.JoinMixingPool(goCtx, msg)

	suite.Require().Error(err)
	suite.Require().Nil(resp)
	suite.Require().Equal(codes.InvalidArgument, status.Code(err))
	suite.Require().Contains(err.Error(), "participant cannot be empty")
}

func (suite *MsgServerComprehensiveTestSuite) TestJoinMixingPool_EmptyPoolId() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	msg := &privacypb.MsgJoinMixingPool{
		Participant: keepertest.GenTestAddr().String(),
		PoolId:      "",
	}

	resp, err := suite.msgServer.JoinMixingPool(goCtx, msg)

	suite.Require().Error(err)
	suite.Require().Nil(resp)
	suite.Require().Equal(codes.InvalidArgument, status.Code(err))
	suite.Require().Contains(err.Error(), "pool id cannot be empty")
}

func (suite *MsgServerComprehensiveTestSuite) TestJoinMixingPool_PoolNotFound() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	msg := &privacypb.MsgJoinMixingPool{
		Participant: keepertest.GenTestAddr().String(),
		PoolId:      "nonexistent_pool",
	}

	resp, err := suite.msgServer.JoinMixingPool(goCtx, msg)

	suite.Require().Error(err)
	suite.Require().Nil(resp)
	suite.Require().Equal(codes.NotFound, status.Code(err))
	suite.Require().Contains(err.Error(), "mixing pool not found")
}

func (suite *MsgServerComprehensiveTestSuite) TestJoinMixingPool_Success() {
	// First create a pool
	goCtx := sdk.WrapSDKContext(suite.ctx)

	createMsg := &privacypb.MsgCreateMixingPool{
		Creator:         keepertest.GenTestAddr().String(),
		MinParticipants: 2,
		MaxParticipants: 10,
		Denomination:    sdkmath.NewInt(1000000),
		MixingRounds:    3,
	}

	createResp, err := suite.msgServer.CreateMixingPool(goCtx, createMsg)
	suite.Require().NoError(err)

	// Join the pool
	participant := keepertest.GenTestAddr().String()
	joinMsg := &privacypb.MsgJoinMixingPool{
		Participant: participant,
		PoolId:      createResp.PoolId,
	}

	joinResp, err := suite.msgServer.JoinMixingPool(goCtx, joinMsg)

	suite.Require().NoError(err)
	suite.Require().NotNil(joinResp)
	suite.Require().True(joinResp.Success)
	suite.Require().Equal(uint32(0), joinResp.ParticipantIndex)

	// Verify pool was updated
	pool, err := suite.keeper.GetMixingPool(suite.ctx, createResp.PoolId)
	suite.Require().NoError(err)
	suite.Require().Len(pool.Participants, 1)
}

func (suite *MsgServerComprehensiveTestSuite) TestJoinMixingPool_AlreadyParticipating() {
	// Create a pool
	goCtx := sdk.WrapSDKContext(suite.ctx)

	createMsg := &privacypb.MsgCreateMixingPool{
		Creator:         keepertest.GenTestAddr().String(),
		MinParticipants: 2,
		MaxParticipants: 10,
		Denomination:    sdkmath.NewInt(1000000),
		MixingRounds:    3,
	}

	createResp, err := suite.msgServer.CreateMixingPool(goCtx, createMsg)
	suite.Require().NoError(err)

	// Join the pool
	participant := keepertest.GenTestAddr().String()
	joinMsg := &privacypb.MsgJoinMixingPool{
		Participant: participant,
		PoolId:      createResp.PoolId,
	}

	_, err = suite.msgServer.JoinMixingPool(goCtx, joinMsg)
	suite.Require().NoError(err)

	// Try to join again
	joinResp, err := suite.msgServer.JoinMixingPool(goCtx, joinMsg)

	suite.Require().Error(err)
	suite.Require().Nil(joinResp)
	suite.Require().Equal(codes.AlreadyExists, status.Code(err))
	suite.Require().Contains(err.Error(), "already participating in pool")
}

func (suite *MsgServerComprehensiveTestSuite) TestJoinMixingPool_PoolFull() {
	// Create a pool with max 2 participants
	goCtx := sdk.WrapSDKContext(suite.ctx)

	createMsg := &privacypb.MsgCreateMixingPool{
		Creator:         keepertest.GenTestAddr().String(),
		MinParticipants: 2,
		MaxParticipants: 2,
		Denomination:    sdkmath.NewInt(1000000),
		MixingRounds:    3,
	}

	createResp, err := suite.msgServer.CreateMixingPool(goCtx, createMsg)
	suite.Require().NoError(err)

	// Fill the pool with 2 participants
	joinMsg1 := &privacypb.MsgJoinMixingPool{
		Participant: keepertest.GenTestAddr().String(),
		PoolId:      createResp.PoolId,
	}

	_, err = suite.msgServer.JoinMixingPool(goCtx, joinMsg1)
	suite.Require().NoError(err)

	joinMsg2 := &privacypb.MsgJoinMixingPool{
		Participant: keepertest.GenTestAddr().String(),
		PoolId:      createResp.PoolId,
	}

	_, err = suite.msgServer.JoinMixingPool(goCtx, joinMsg2)
	suite.Require().NoError(err)

	// Try to join full pool (3rd participant)
	joinMsg3 := &privacypb.MsgJoinMixingPool{
		Participant: keepertest.GenTestAddr().String(),
		PoolId:      createResp.PoolId,
	}

	joinResp, err := suite.msgServer.JoinMixingPool(goCtx, joinMsg3)

	suite.Require().Error(err)
	suite.Require().Nil(joinResp)
	suite.Require().Equal(codes.FailedPrecondition, status.Code(err))
	suite.Require().Contains(err.Error(), "mixing pool is full")
}

func (suite *MsgServerComprehensiveTestSuite) TestJoinMixingPool_StatusChangesToReady() {
	// Create a pool with min 2 participants
	goCtx := sdk.WrapSDKContext(suite.ctx)

	createMsg := &privacypb.MsgCreateMixingPool{
		Creator:         keepertest.GenTestAddr().String(),
		MinParticipants: 2,
		MaxParticipants: 10,
		Denomination:    sdkmath.NewInt(1000000),
		MixingRounds:    3,
	}

	createResp, err := suite.msgServer.CreateMixingPool(goCtx, createMsg)
	suite.Require().NoError(err)

	// Join with first participant
	joinMsg1 := &privacypb.MsgJoinMixingPool{
		Participant: keepertest.GenTestAddr().String(),
		PoolId:      createResp.PoolId,
	}

	_, err = suite.msgServer.JoinMixingPool(goCtx, joinMsg1)
	suite.Require().NoError(err)

	// Pool should still be pending
	pool, err := suite.keeper.GetMixingPool(suite.ctx, createResp.PoolId)
	suite.Require().NoError(err)
	suite.Require().Equal("pending", pool.Status)

	// Join with second participant
	joinMsg2 := &privacypb.MsgJoinMixingPool{
		Participant: keepertest.GenTestAddr().String(),
		PoolId:      createResp.PoolId,
	}

	_, err = suite.msgServer.JoinMixingPool(goCtx, joinMsg2)
	suite.Require().NoError(err)

	// Pool should now be ready
	pool, err = suite.keeper.GetMixingPool(suite.ctx, createResp.PoolId)
	suite.Require().NoError(err)
	suite.Require().Equal("ready", pool.Status)
}

// RevokeViewKey Tests

func (suite *MsgServerComprehensiveTestSuite) TestRevokeViewKey_NilRequest() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	resp, err := suite.msgServer.RevokeViewKey(goCtx, nil)

	suite.Require().Error(err)
	suite.Require().Nil(resp)
	suite.Require().Equal(codes.InvalidArgument, status.Code(err))
	suite.Require().Contains(err.Error(), "empty request")
}

func (suite *MsgServerComprehensiveTestSuite) TestRevokeViewKey_EmptyOwner() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	msg := &privacypb.MsgRevokeViewKey{
		Owner:         "",
		PublicViewKey: []byte("public_key"),
	}

	resp, err := suite.msgServer.RevokeViewKey(goCtx, msg)

	suite.Require().Error(err)
	suite.Require().Nil(resp)
	suite.Require().Equal(codes.InvalidArgument, status.Code(err))
	suite.Require().Contains(err.Error(), "owner cannot be empty")
}

func (suite *MsgServerComprehensiveTestSuite) TestRevokeViewKey_EmptyPublicKey() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	msg := &privacypb.MsgRevokeViewKey{
		Owner:         keepertest.GenTestAddr().String(),
		PublicViewKey: []byte{},
	}

	resp, err := suite.msgServer.RevokeViewKey(goCtx, msg)

	suite.Require().Error(err)
	suite.Require().Nil(resp)
	suite.Require().Equal(codes.InvalidArgument, status.Code(err))
	suite.Require().Contains(err.Error(), "public view key cannot be empty")
}

func (suite *MsgServerComprehensiveTestSuite) TestRevokeViewKey_Success() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	owner := keepertest.GenTestAddr().String()
	publicKey := make([]byte, 32)
	for i := range publicKey {
		publicKey[i] = byte(i)
	}

	msg := &privacypb.MsgRevokeViewKey{
		Owner:         owner,
		PublicViewKey: publicKey,
	}

	resp, err := suite.msgServer.RevokeViewKey(goCtx, msg)

	// Should succeed even if key doesn't exist
	suite.Require().NoError(err)
	suite.Require().NotNil(resp)
	suite.Require().True(resp.Success)

	// Verify event was emitted
	events := suite.ctx.EventManager().Events()
	found := false
	for _, event := range events {
		if event.Type == types.EventTypeViewKey {
			found = true
			break
		}
	}
	suite.Require().True(found, "Event not emitted")
}

// UpdateNetworkPrivacy Tests

func (suite *MsgServerComprehensiveTestSuite) TestUpdateNetworkPrivacy_NilRequest() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	resp, err := suite.msgServer.UpdateNetworkPrivacy(goCtx, nil)

	suite.Require().Error(err)
	suite.Require().Nil(resp)
	suite.Require().Equal(codes.InvalidArgument, status.Code(err))
	suite.Require().Contains(err.Error(), "empty request")
}

func (suite *MsgServerComprehensiveTestSuite) TestUpdateNetworkPrivacy_EmptySender() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	msg := &privacypb.MsgUpdateNetworkPrivacy{
		Sender:         "",
		NetworkPrivacy: &privacypb.NetworkPrivacy{},
	}

	resp, err := suite.msgServer.UpdateNetworkPrivacy(goCtx, msg)

	suite.Require().Error(err)
	suite.Require().Nil(resp)
	suite.Require().Equal(codes.InvalidArgument, status.Code(err))
	suite.Require().Contains(err.Error(), "sender cannot be empty")
}

func (suite *MsgServerComprehensiveTestSuite) TestUpdateNetworkPrivacy_NilNetworkPrivacy() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	msg := &privacypb.MsgUpdateNetworkPrivacy{
		Sender:         keepertest.GenTestAddr().String(),
		NetworkPrivacy: nil,
	}

	resp, err := suite.msgServer.UpdateNetworkPrivacy(goCtx, msg)

	suite.Require().Error(err)
	suite.Require().Nil(resp)
	suite.Require().Equal(codes.InvalidArgument, status.Code(err))
	suite.Require().Contains(err.Error(), "network privacy cannot be nil")
}

func (suite *MsgServerComprehensiveTestSuite) TestUpdateNetworkPrivacy_NotEnabled() {
	// Disable network privacy
	params := suite.keeper.GetParams(suite.ctx)
	params.EnableNetworkPrivacy = false
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)

	goCtx := sdk.WrapSDKContext(suite.ctx)

	msg := &privacypb.MsgUpdateNetworkPrivacy{
		Sender:         keepertest.GenTestAddr().String(),
		NetworkPrivacy: &privacypb.NetworkPrivacy{},
	}

	resp, err := suite.msgServer.UpdateNetworkPrivacy(goCtx, msg)

	suite.Require().Error(err)
	suite.Require().Nil(resp)
	suite.Require().Equal(codes.FailedPrecondition, status.Code(err))
	suite.Require().Contains(err.Error(), "network privacy not enabled")
}

func (suite *MsgServerComprehensiveTestSuite) TestUpdateNetworkPrivacy_Success() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	sender := keepertest.GenTestAddr().String()
	msg := &privacypb.MsgUpdateNetworkPrivacy{
		Sender: sender,
		NetworkPrivacy: &privacypb.NetworkPrivacy{
			NetworkType: "TOR",
		},
	}

	resp, err := suite.msgServer.UpdateNetworkPrivacy(goCtx, msg)

	suite.Require().NoError(err)
	suite.Require().NotNil(resp)
	suite.Require().True(resp.Success)

	// Verify event was emitted
	events := suite.ctx.EventManager().Events()
	found := false
	for _, event := range events {
		if event.Type == types.EventTypeNetworkPrivacy {
			found = true
			break
		}
	}
	suite.Require().True(found, "Event not emitted")
}

// UpdateParams Tests

func (suite *MsgServerComprehensiveTestSuite) TestUpdateParams_NilRequest() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	resp, err := suite.msgServer.UpdateParams(goCtx, nil)

	suite.Require().Error(err)
	suite.Require().Nil(resp)
	suite.Require().Equal(codes.InvalidArgument, status.Code(err))
	suite.Require().Contains(err.Error(), "empty request")
}

func (suite *MsgServerComprehensiveTestSuite) TestUpdateParams_Success() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// Use valid authority address
	validAuthority := keepertest.GenTestAddr().String()
	msg := &privacypb.MsgUpdateParams{
		Authority: validAuthority,
		Params: privacypb.Params{
			EnableZkProofs:                 true,
			EnableStealthAddresses:         true,
			EnableRingSignatures:           true,
			EnableConfidentialTransactions: true,
			EnableNetworkPrivacy:           true,
			EnableMixing:                   true,
			MinRingSize:                    11,
			MaxRingSize:                    100,
			MinMixingParticipants:          3,
			MixingFee:                      sdkmath.NewInt(1000),
			ZkProofVerificationCost:        50000,
		},
	}

	resp, err := suite.msgServer.UpdateParams(goCtx, msg)

	suite.Require().NoError(err)
	suite.Require().NotNil(resp)

	// Verify params were updated
	params := suite.keeper.GetParams(suite.ctx)
	suite.Require().Equal(uint32(11), params.MinRingSize)
	suite.Require().Equal(uint32(3), params.MinMixingParticipants)

	// Verify event was emitted
	events := suite.ctx.EventManager().Events()
	found := false
	for _, event := range events {
		if event.Type == types.EventTypeUpdateParams {
			found = true
			break
		}
	}
	suite.Require().True(found, "Event not emitted")
}
