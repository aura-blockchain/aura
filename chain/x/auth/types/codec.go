package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"

	authpb "github.com/aequitas/aura/proto/aura/auth/v1beta1"
)

// RegisterLegacyAminoCodec registers the necessary x/auth interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// Register all message types for Amino JSON serialization
	cdc.RegisterConcrete(&authpb.MsgCreateRole{}, "aura/auth/MsgCreateRole", nil)
	cdc.RegisterConcrete(&authpb.MsgAssignRole{}, "aura/auth/MsgAssignRole", nil)
	cdc.RegisterConcrete(&authpb.MsgRevokeRole{}, "aura/auth/MsgRevokeRole", nil)
	cdc.RegisterConcrete(&authpb.MsgCreateMultisigWallet{}, "aura/auth/MsgCreateMultisigWallet", nil)
	cdc.RegisterConcrete(&authpb.MsgCreateMultisigProposal{}, "aura/auth/MsgCreateMultisigProposal", nil)
	cdc.RegisterConcrete(&authpb.MsgSignMultisigProposal{}, "aura/auth/MsgSignMultisigProposal", nil)
	cdc.RegisterConcrete(&authpb.MsgExecuteMultisigProposal{}, "aura/auth/MsgExecuteMultisigProposal", nil)
	cdc.RegisterConcrete(&authpb.MsgProposeTimeLockedAction{}, "aura/auth/MsgProposeTimeLockedAction", nil)
	cdc.RegisterConcrete(&authpb.MsgExecuteTimeLockedAction{}, "aura/auth/MsgExecuteTimeLockedAction", nil)
	cdc.RegisterConcrete(&authpb.MsgCancelTimeLockedAction{}, "aura/auth/MsgCancelTimeLockedAction", nil)
	cdc.RegisterConcrete(&authpb.MsgActivateEmergencyAdmin{}, "aura/auth/MsgActivateEmergencyAdmin", nil)
	cdc.RegisterConcrete(&authpb.MsgDeactivateEmergencyAdmin{}, "aura/auth/MsgDeactivateEmergencyAdmin", nil)
	cdc.RegisterConcrete(&authpb.MsgInitiateValidatorKeyRotation{}, "aura/auth/MsgInitiateValidatorKeyRotation", nil)
	cdc.RegisterConcrete(&authpb.MsgCompleteValidatorKeyRotation{}, "aura/auth/MsgCompleteValidatorKeyRotation", nil)
	cdc.RegisterConcrete(&authpb.MsgCreateSession{}, "aura/auth/MsgCreateSession", nil)
	cdc.RegisterConcrete(&authpb.MsgRevokeSession{}, "aura/auth/MsgRevokeSession", nil)
}

// RegisterInterfaces registers the x/auth interfaces types with the interface registry
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(registry, &authpb.Msg_ServiceDesc)

	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&authpb.MsgCreateRole{},
		&authpb.MsgAssignRole{},
		&authpb.MsgRevokeRole{},
		&authpb.MsgCreateMultisigWallet{},
		&authpb.MsgCreateMultisigProposal{},
		&authpb.MsgSignMultisigProposal{},
		&authpb.MsgExecuteMultisigProposal{},
		&authpb.MsgProposeTimeLockedAction{},
		&authpb.MsgExecuteTimeLockedAction{},
		&authpb.MsgCancelTimeLockedAction{},
		&authpb.MsgActivateEmergencyAdmin{},
		&authpb.MsgDeactivateEmergencyAdmin{},
		&authpb.MsgInitiateValidatorKeyRotation{},
		&authpb.MsgCompleteValidatorKeyRotation{},
		&authpb.MsgCreateSession{},
		&authpb.MsgRevokeSession{},
	)
}
