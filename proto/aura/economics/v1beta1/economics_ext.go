package v1beta1

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

// UnpackInterfaces implements the UnpackInterfacesMessage interface for Params.
// Since Params doesn't contain any google.protobuf.Any fields, this is a no-op.
func (m *Params) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return nil
}

// UnpackInterfaces implements the UnpackInterfacesMessage interface for GenesisState.
// Since GenesisState doesn't contain any google.protobuf.Any fields, this is a no-op.
func (m *GenesisState) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return nil
}

// UnpackInterfaces implements the UnpackInterfacesMessage interface for VestingSchedule.
// Since VestingSchedule doesn't contain any google.protobuf.Any fields, this is a no-op.
func (m *VestingSchedule) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return nil
}

// UnpackInterfaces implements the UnpackInterfacesMessage interface for VoteLock.
// Since VoteLock doesn't contain any google.protobuf.Any fields, this is a no-op.
func (m *VoteLock) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return nil
}

// UnpackInterfaces implements the UnpackInterfacesMessage interface for Proposal.
// Since Proposal doesn't contain any google.protobuf.Any fields, this is a no-op.
func (m *Proposal) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return nil
}

// UnpackInterfaces implements the UnpackInterfacesMessage interface for Vote.
// Since Vote doesn't contain any google.protobuf.Any fields, this is a no-op.
func (m *Vote) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return nil
}

// UnpackInterfaces implements the UnpackInterfacesMessage interface for Deposit.
// Since Deposit doesn't contain any google.protobuf.Any fields, this is a no-op.
func (m *Deposit) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return nil
}

// UnpackInterfaces implements the UnpackInterfacesMessage interface for VoteDelegation.
// Since VoteDelegation doesn't contain any google.protobuf.Any fields, this is a no-op.
func (m *VoteDelegation) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return nil
}

// UnpackInterfaces implements the UnpackInterfacesMessage interface for PendingTreasuryTx.
// Since PendingTreasuryTx doesn't contain any google.protobuf.Any fields, this is a no-op.
func (m *PendingTreasuryTx) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return nil
}
