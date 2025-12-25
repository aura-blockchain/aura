// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package compliance

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"

	"github.com/aequitas/aura/chain/x/compliance/keeper"
	"github.com/aequitas/aura/chain/x/compliance/types"
)

var _ porttypes.IBCModule = IBCModule{}

// IBCModule implements the IBC module interface for the compliance module.
//
// IBC DISABLED FOR TESTNET:
// Cross-chain compliance features are not enabled in the testnet phase.
// All IBC callbacks return clear error messages indicating that IBC functionality
// will be available in v2.0 mainnet release.
//
// Current compliance features (testnet):
//   - Local KYC verification
//   - Sanctions screening
//   - GDPR data management
//   - AML risk scoring
//   - Tax reporting
//
// Planned v2.0 IBC features:
//   - Cross-chain KYC verification
//   - Interchain sanctions list synchronization
//   - Cross-chain compliance attestations
//   - Global AML risk aggregation
//   - Multi-jurisdiction tax reporting coordination
//
// Security considerations:
//   - All handlers return explicit errors instead of silent failures
//   - No panics or undefined behavior
//   - Clear messaging to users about feature availability
//   - Local compliance features remain fully functional
type IBCModule struct {
	keeper *keeper.Keeper
}

// NewIBCModule creates a new IBCModule instance.
func NewIBCModule(k *keeper.Keeper) IBCModule {
	return IBCModule{
		keeper: k,
	}
}

// OnChanOpenInit implements the IBCModule interface.
// IBC is not enabled for compliance module - returns clear error.
func (im IBCModule) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version string,
) (string, error) {
	return "", types.ErrIBCNotEnabled
}

// OnChanOpenTry implements the IBCModule interface.
// IBC is not enabled for compliance module - returns clear error.
func (im IBCModule) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (version string, err error) {
	return "", types.ErrIBCNotEnabled
}

// OnChanOpenAck implements the IBCModule interface.
// IBC is not enabled for compliance module - returns clear error.
func (im IBCModule) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	counterpartyChannelID string,
	counterpartyVersion string,
) error {
	return types.ErrIBCNotEnabled
}

// OnChanOpenConfirm implements the IBCModule interface.
// IBC is not enabled for compliance module - returns clear error.
func (im IBCModule) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return types.ErrIBCNotEnabled
}

// OnChanCloseInit implements the IBCModule interface.
// IBC is not enabled for compliance module - returns clear error.
func (im IBCModule) OnChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return types.ErrIBCNotEnabled
}

// OnChanCloseConfirm implements the IBCModule interface.
// IBC is not enabled for compliance module - returns clear error.
func (im IBCModule) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return types.ErrIBCNotEnabled
}

// OnRecvPacket implements the IBCModule interface.
// IBC is not enabled for compliance module - returns error acknowledgement.
//
// Security: This prevents silent packet acceptance while clearly communicating
// that the feature is disabled. The error acknowledgement will be relayed back
// to the sender, allowing them to know that cross-chain compliance is not available.
//
// Compliance data processing is restricted to local chain operations to ensure
// GDPR compliance and data sovereignty requirements are met.
func (im IBCModule) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	return channeltypes.NewErrorAcknowledgement(types.ErrIBCNotEnabled)
}

// OnAcknowledgementPacket implements the IBCModule interface.
// IBC is not enabled for compliance module - returns clear error.
func (im IBCModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	return types.ErrIBCNotEnabled
}

// OnTimeoutPacket implements the IBCModule interface.
// IBC is not enabled for compliance module - returns clear error.
func (im IBCModule) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	return types.ErrIBCNotEnabled
}

// NegotiateAppVersion implements the IBCModule interface.
// IBC is not enabled for compliance module - returns clear error.
//
// This method is called during the channel handshake to negotiate the application
// version. Since IBC is disabled, we return an error immediately.
func (im IBCModule) NegotiateAppVersion(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionID string,
	portID string,
	counterparty channeltypes.Counterparty,
	proposedVersion string,
) (string, error) {
	return "", types.ErrIBCNotEnabled
}

// GetAppVersion implements the IBCModule interface.
// IBC is not enabled for compliance module - returns empty version and false.
//
// Note: When IBC is enabled in v2.0, this should return the compliance module's
// IBC protocol version (e.g., "compliance-1").
func (im IBCModule) GetAppVersion(ctx sdk.Context, portID, channelID string) (string, bool) {
	// Return empty version and false to indicate IBC is not supported
	return "", false
}

// SendPacket is a helper function for when IBC is eventually enabled in v2.0.
// Currently returns an error since IBC is disabled.
//
// Future implementation will handle:
//   - Cross-chain KYC verification requests
//   - Sanctions list synchronization
//   - Compliance attestation propagation
//   - Multi-jurisdiction reporting
func (im IBCModule) SendPacket(
	ctx sdk.Context,
	sourcePort string,
	sourceChannel string,
	timeoutHeight uint64,
	timeoutTimestamp uint64,
	data []byte,
) error {
	return fmt.Errorf("SendPacket not available: %w", types.ErrIBCNotEnabled)
}
