package bridge

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"

	"github.com/aequitas/aura/chain/x/bridge/keeper"
	"github.com/aequitas/aura/chain/x/bridge/types"
)

var _ porttypes.IBCModule = IBCModule{}

// IBCModule implements the IBC module interface for the bridge module.
//
// IBC DISABLED FOR TESTNET:
// IBC-based cross-chain bridging is not enabled in the testnet phase.
// The bridge module currently uses attestation-based bridging with validator consensus.
// All IBC callbacks return clear error messages indicating that IBC functionality
// will be available in v2.0 mainnet release.
//
// Current bridging mechanism (testnet):
//   - Attestation-based with validator multi-sig
//   - Block header verification by oracle
//   - Timelock and fraud proof protection
//   - Circuit breaker for security
//
// Planned v2.0 IBC features:
//   - IBC light client verification
//   - IBC packet-based asset transfers
//   - IBC-native cross-chain messaging
//   - Connection to other Cosmos SDK chains
//
// Security considerations:
//   - All handlers return explicit errors instead of silent failures
//   - No panics or undefined behavior
//   - Clear messaging to users about feature availability
//   - Existing attestation-based bridging remains fully functional
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
// IBC is not enabled for bridge module - returns clear error.
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
// IBC is not enabled for bridge module - returns clear error.
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
// IBC is not enabled for bridge module - returns clear error.
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
// IBC is not enabled for bridge module - returns clear error.
func (im IBCModule) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return types.ErrIBCNotEnabled
}

// OnChanCloseInit implements the IBCModule interface.
// IBC is not enabled for bridge module - returns clear error.
func (im IBCModule) OnChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return types.ErrIBCNotEnabled
}

// OnChanCloseConfirm implements the IBCModule interface.
// IBC is not enabled for bridge module - returns clear error.
func (im IBCModule) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return types.ErrIBCNotEnabled
}

// OnRecvPacket implements the IBCModule interface.
// IBC is not enabled for bridge module - returns error acknowledgement.
//
// Security: This prevents silent packet acceptance while clearly communicating
// that the feature is disabled. The error acknowledgement will be relayed back
// to the sender, allowing them to know that IBC bridging is not available.
//
// Users should use the attestation-based bridging mechanism via MsgInitiateTransfer
// and MsgSubmitAttestation instead.
func (im IBCModule) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	return channeltypes.NewErrorAcknowledgement(types.ErrIBCNotEnabled)
}

// OnAcknowledgementPacket implements the IBCModule interface.
// IBC is not enabled for bridge module - returns clear error.
func (im IBCModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	return types.ErrIBCNotEnabled
}

// OnTimeoutPacket implements the IBCModule interface.
// IBC is not enabled for bridge module - returns clear error.
func (im IBCModule) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	return types.ErrIBCNotEnabled
}

// NegotiateAppVersion implements the IBCModule interface.
// IBC is not enabled for bridge module - returns clear error.
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
// IBC is not enabled for bridge module - returns empty version and false.
//
// Note: When IBC is enabled in v2.0, this should return the bridge module's
// IBC protocol version (e.g., "bridge-1" or use ICS-20 token transfer standard).
func (im IBCModule) GetAppVersion(ctx sdk.Context, portID, channelID string) (string, bool) {
	// Return empty version and false to indicate IBC is not supported
	return "", false
}

// SendPacket is a helper function for when IBC is eventually enabled in v2.0.
// Currently returns an error since IBC is disabled.
//
// Future implementation will handle:
//   - IBC token transfers (ICS-20)
//   - Cross-chain asset locks
//   - Bridge state synchronization
//   - Multi-hop routing
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
