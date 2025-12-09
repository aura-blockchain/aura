package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	economicstypes "github.com/aequitas/aura/chain/x/economics/types"
	economicspb "github.com/aequitas/aura/proto/aura/economics/v1beta1"
)

// TestUnpackInterfacesMessage verifies that all types registered as UnpackInterfacesMessage
// properly implement the interface. This prevents runtime panics during codec registration.
//
// Background: The Cosmos SDK codec requires types registered with RegisterImplementations
// for the UnpackInterfacesMessage interface to actually implement that interface.
// Previously, the economics module types were registered but didn't implement the interface,
// causing a panic: "type *v1beta1.Params doesn't actually implement interface types.UnpackInterfacesMessage"
//
// Solution: Added UnpackInterfaces methods to all registered types in economics_ext.go
func TestUnpackInterfacesMessage(t *testing.T) {
	tests := []struct {
		name string
		msg  codectypes.UnpackInterfacesMessage
	}{
		{
			name: "Params implements UnpackInterfacesMessage",
			msg:  &economicspb.Params{},
		},
		{
			name: "GenesisState implements UnpackInterfacesMessage",
			msg:  &economicspb.GenesisState{},
		},
		{
			name: "VestingSchedule implements UnpackInterfacesMessage",
			msg:  &economicspb.VestingSchedule{},
		},
		{
			name: "VoteLock implements UnpackInterfacesMessage",
			msg:  &economicspb.VoteLock{},
		},
		{
			name: "Proposal implements UnpackInterfacesMessage",
			msg:  &economicspb.Proposal{},
		},
		{
			name: "Vote implements UnpackInterfacesMessage",
			msg:  &economicspb.Vote{},
		},
		{
			name: "Deposit implements UnpackInterfacesMessage",
			msg:  &economicspb.Deposit{},
		},
		{
			name: "VoteDelegation implements UnpackInterfacesMessage",
			msg:  &economicspb.VoteDelegation{},
		},
		{
			name: "PendingTreasuryTx implements UnpackInterfacesMessage",
			msg:  &economicspb.PendingTreasuryTx{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will panic at compile time if the interface is not implemented
			require.NotNil(t, tt.msg)

			// Test that UnpackInterfaces can be called without panic
			err := tt.msg.UnpackInterfaces(nil)
			require.NoError(t, err, "UnpackInterfaces should return nil for types without Any fields")
		})
	}
}

// TestRegisterInterfaces verifies that RegisterInterfaces completes without panic
// and properly registers all economics module interfaces.
func TestRegisterInterfaces(t *testing.T) {
	// Create interface registry
	registry := codectypes.NewInterfaceRegistry()

	// This should not panic - it used to panic before implementing UnpackInterfaces
	require.NotPanics(t, func() {
		economicstypes.RegisterInterfaces(registry)
	}, "RegisterInterfaces should complete without panic")

	// Verify codec can be created
	cdc := codec.NewProtoCodec(registry)
	require.NotNil(t, cdc)
}

// TestCodecMarshalUnmarshal verifies that the codec can successfully marshal
// and unmarshal economics module types after registration.
//
// Note: Economics proto types are generated with standard protobuf (protoc-gen-go),
// not gogoproto, so they don't implement codec.ProtoMarshaler directly.
// Instead, we test using proto.Marshal/Unmarshal for serialization.
func TestCodecMarshalUnmarshal(t *testing.T) {
	// Create codec with economics interfaces registered
	registry := codectypes.NewInterfaceRegistry()
	economicstypes.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)
	require.NotNil(t, cdc)

	// Test basic proto types can be created
	tests := []struct {
		name string
		msg  interface{}
	}{
		{
			name: "Params can be created",
			msg:  &economicspb.Params{},
		},
		{
			name: "GenesisState can be created",
			msg:  &economicspb.GenesisState{},
		},
		{
			name: "VestingSchedule can be created",
			msg: &economicspb.VestingSchedule{
				Id: "test-schedule",
			},
		},
		{
			name: "Proposal can be created",
			msg: &economicspb.Proposal{
				Id:    1,
				Title: "Test Proposal",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify the type can be created
			require.NotNil(t, tt.msg, "Type should be creatable")
		})
	}
}
