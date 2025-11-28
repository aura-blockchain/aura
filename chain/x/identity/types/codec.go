package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"

	identitypb "github.com/aequitas/aura/proto/aura/identity/v1beta1"
)

// RegisterLegacyAminoCodec registers the necessary x/identity interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// Role management
	legacy.RegisterAminoMsg(cdc, &identitypb.MsgCreateRole{}, "identity/CreateRole")
	legacy.RegisterAminoMsg(cdc, &identitypb.MsgAssignRole{}, "identity/AssignRole")
	legacy.RegisterAminoMsg(cdc, &identitypb.MsgRevokeRole{}, "identity/RevokeRole")

	// Identity change management
	legacy.RegisterAminoMsg(cdc, &identitypb.MsgRequestIdentityChange{}, "identity/RequestIdentityChange")
	legacy.RegisterAminoMsg(cdc, &identitypb.MsgSubmitAssistantProof{}, "identity/SubmitAssistantProof")
	legacy.RegisterAminoMsg(cdc, &identitypb.MsgApplyIdentityChange{}, "identity/ApplyIdentityChange")
	legacy.RegisterAminoMsg(cdc, &identitypb.MsgRejectIdentityChange{}, "identity/RejectIdentityChange")
	legacy.RegisterAminoMsg(cdc, &identitypb.MsgSuspendIdentityChanges{}, "identity/SuspendIdentityChanges")

	// Multisig wallet management
	legacy.RegisterAminoMsg(cdc, &identitypb.MsgCreateMultisigWallet{}, "identity/CreateMultisigWallet")
	legacy.RegisterAminoMsg(cdc, &identitypb.MsgCreateMultisigProposal{}, "identity/CreateMultisigProposal")
	legacy.RegisterAminoMsg(cdc, &identitypb.MsgSignMultisigProposal{}, "identity/SignMultisigProposal")
	legacy.RegisterAminoMsg(cdc, &identitypb.MsgExecuteMultisigProposal{}, "identity/ExecuteMultisigProposal")

	// Time-locked actions
	legacy.RegisterAminoMsg(cdc, &identitypb.MsgProposeTimeLockedAction{}, "identity/ProposeTimeLockedAction")
	legacy.RegisterAminoMsg(cdc, &identitypb.MsgExecuteTimeLockedAction{}, "identity/ExecuteTimeLockedAction")
	legacy.RegisterAminoMsg(cdc, &identitypb.MsgCancelTimeLockedAction{}, "identity/CancelTimeLockedAction")

	// Emergency admin
	legacy.RegisterAminoMsg(cdc, &identitypb.MsgActivateEmergencyAdmin{}, "identity/ActivateEmergencyAdmin")
	legacy.RegisterAminoMsg(cdc, &identitypb.MsgDeactivateEmergencyAdmin{}, "identity/DeactivateEmergencyAdmin")

	// Validator key rotation
	legacy.RegisterAminoMsg(cdc, &identitypb.MsgRotateValidatorKey{}, "identity/RotateValidatorKey")

	// Session management
	legacy.RegisterAminoMsg(cdc, &identitypb.MsgCreateSession{}, "identity/CreateSession")
	legacy.RegisterAminoMsg(cdc, &identitypb.MsgEndSession{}, "identity/EndSession")

	// Params
	legacy.RegisterAminoMsg(cdc, &identitypb.MsgUpdateParams{}, "identity/UpdateParams")
}

// RegisterInterfaces registers the x/identity interfaces types with the interface registry
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		// Role management
		&identitypb.MsgCreateRole{},
		&identitypb.MsgAssignRole{},
		&identitypb.MsgRevokeRole{},

		// Identity change management
		&identitypb.MsgRequestIdentityChange{},
		&identitypb.MsgSubmitAssistantProof{},
		&identitypb.MsgApplyIdentityChange{},
		&identitypb.MsgRejectIdentityChange{},
		&identitypb.MsgSuspendIdentityChanges{},

		// Multisig wallet management
		&identitypb.MsgCreateMultisigWallet{},
		&identitypb.MsgCreateMultisigProposal{},
		&identitypb.MsgSignMultisigProposal{},
		&identitypb.MsgExecuteMultisigProposal{},

		// Time-locked actions
		&identitypb.MsgProposeTimeLockedAction{},
		&identitypb.MsgExecuteTimeLockedAction{},
		&identitypb.MsgCancelTimeLockedAction{},

		// Emergency admin
		&identitypb.MsgActivateEmergencyAdmin{},
		&identitypb.MsgDeactivateEmergencyAdmin{},

		// Validator key rotation
		&identitypb.MsgRotateValidatorKey{},

		// Session management
		&identitypb.MsgCreateSession{},
		&identitypb.MsgEndSession{},

		// Params
		&identitypb.MsgUpdateParams{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &identitypb.Msg_ServiceDesc)
}
