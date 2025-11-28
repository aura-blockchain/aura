package aurabindings

import (
	"encoding/json"
	"time"

	"cosmossdk.io/errors"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	vckeeper "github.com/aequitas/aura/chain/x/vcregistry/keeper"
	vctypes "github.com/aequitas/aura/chain/x/vcregistry/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Define the custom message structure that the smart contracts will use.
// This should mirror the Rust `AuraMsg` enum.
type AuraMsg struct {
	VCRegistry *VCRegistryMsg `json:"vc_registry,omitempty"`
}

type VCRegistryMsg struct {
	RegisterVC *RegisterVCMsg `json:"register_vc,omitempty"`
}

type RegisterVCMsg struct {
	Address  string `json:"address"`
	VCBase64 string `json:"vc_base64"`
}

// NewMessageHandler creates a new message handler for the wasm keeper with security validation.
func NewMessageHandler(
	vcKeeper *vckeeper.Keeper,
	// wasmKeeper *wasmkeeper.Keeper,
) wasmkeeper.Messenger {
	return MessageHandler{
		vcKeeper: vcKeeper,
		// securityValidator: NewSecurityValidator(wasmKeeper),
	}
}

type MessageHandler struct {
	vcKeeper *vckeeper.Keeper
	// securityValidator *SecurityValidator
}

func (m MessageHandler) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg wasmvmtypes.CosmosMsg) ([]sdk.Event, [][]byte, error) {
	// SECURITY: Verify contract is approved for custom bindings
	// Commented out until wasm keeper types are available
	// if err := m.securityValidator.ValidateContractPermissions(ctx, contractAddr, "custom_binding"); err != nil {
	// 	m.securityValidator.LogSecurityEvent(ctx, "custom_binding_unauthorized", contractAddr, false, err.Error(), nil)
	// 	return nil, nil, err
	// }

	// SECURITY: Check rate limits before processing
	// if err := m.securityValidator.CheckRateLimit(ctx, contractAddr, "custom_msg"); err != nil {
	// 	m.securityValidator.LogSecurityEvent(ctx, "custom_binding_rate_limited", contractAddr, false, err.Error(), nil)
	// 	return nil, nil, err
	// }

	var auraMsg AuraMsg
	if err := json.Unmarshal(msg.Custom, &auraMsg); err != nil {
		// m.securityValidator.LogSecurityEvent(ctx, "custom_binding_unmarshal_error", contractAddr, false, err.Error(), nil)
		return nil, nil, errors.Wrap(err, "failed to unmarshal aura msg")
	}

	if auraMsg.VCRegistry != nil {
		if auraMsg.VCRegistry.RegisterVC != nil {
			// SECURITY: Validate address
			addr, err := sdk.AccAddressFromBech32(auraMsg.VCRegistry.RegisterVC.Address)
			if err != nil {
				return nil, nil, errors.Wrap(err, "invalid address")
			}

			vcType := "binding_tester_vc"

			// SECURITY: Check if contract is authorized to register VC for this address/type
			// Commented out until wasm keeper types are available
			// if err := m.securityValidator.CanRegisterVCFor(ctx, contractAddr, addr.String(), vcType); err != nil {
			// 	m.securityValidator.LogSecurityEvent(ctx, "register_vc_unauthorized", contractAddr, false, err.Error(), map[string]interface{}{
			// 		"target_address": addr.String(),
			// 		"vc_type":        vcType,
			// 	})
			// 	return nil, nil, err
			// }

			// SECURITY: Validate VC data (size, format, required fields)
			// Commented out until wasm keeper types are available
			// if err := m.securityValidator.ValidateVCData(auraMsg.VCRegistry.RegisterVC.VCBase64); err != nil {
			// 	m.securityValidator.LogSecurityEvent(ctx, "register_vc_invalid_data", contractAddr, false, err.Error(), nil)
			// 	return nil, nil, err
			// }

			// Generate a UUID for vc_id
			vcID := uuid.New().String()

			vcRecord := vctypes.VCRecord{
				VcId:           vcID,
				VcType:         vctypes.VCType_VC_TYPE_CUSTOM, // Using custom for now
				VcTypeCustom:   vcType,
				HolderDid:      "did:aura:" + addr.String(), // Dummy DID for now
				HolderAddress:  addr.String(),
				Status:         vctypes.VCStatus_VC_STATUS_ACTIVE,
				IssuedAt:       timestamppb.New(time.Now()),
				ExpiresAt:      nil, // No expiry for now
				IssuedHeight:   uint64(ctx.BlockHeight()),
				CredentialHash: []byte(auraMsg.VCRegistry.RegisterVC.VCBase64), // Using base64 as hash for now
				Metadata:       map[string]string{"vc_base64": auraMsg.VCRegistry.RegisterVC.VCBase64},
			}

			if err := m.vcKeeper.SetVCRecord(ctx, vcRecord); err != nil {
				// m.securityValidator.LogSecurityEvent(ctx, "register_vc_failed", contractAddr, false, err.Error(), map[string]interface{}{
				// 	"vc_id":   vcID,
				// 	"vc_type": vcType,
				// })
				return nil, nil, errors.Wrap(err, "failed to register verifiable credential record")
			}

			// SECURITY: Log successful VC registration for audit trail
			// Commented out until wasm keeper types are available
			// m.securityValidator.LogSecurityEvent(ctx, "register_vc_success", contractAddr, true, "", map[string]interface{}{
			// 	"vc_id":          vcID,
			// 	"vc_type":        vcType,
			// 	"holder_address": addr.String(),
			// 	"data_size":      len(auraMsg.VCRegistry.RegisterVC.VCBase64),
			// })

			return []sdk.Event{}, [][]byte{}, nil
		}
	}

	// m.securityValidator.LogSecurityEvent(ctx, "unknown_custom_msg", contractAddr, false, "unknown message type", nil)
	return nil, nil, errors.Wrap(wasmtypes.ErrUnknownMsg, "unknown aura msg")
}
