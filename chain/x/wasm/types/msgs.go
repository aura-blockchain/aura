package types

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Message types

// MsgStoreCode uploads contract code
type MsgStoreCode struct {
	Sender       string `json:"sender"`
	WASMByteCode []byte `json:"wasm_byte_code"`
}

// MsgStoreCodeResponse is the response for MsgStoreCode
type MsgStoreCodeResponse struct {
	CodeID uint64 `json:"code_id"`
}

// MsgInstantiateContract instantiates a contract
type MsgInstantiateContract struct {
	Sender string          `json:"sender"`
	Admin  string          `json:"admin,omitempty"`
	CodeID uint64          `json:"code_id"`
	Label  string          `json:"label"`
	Msg    json.RawMessage `json:"msg"`
	Funds  sdk.Coins       `json:"funds"`
}

// MsgInstantiateContractResponse is the response for MsgInstantiateContract
type MsgInstantiateContractResponse struct {
	Address string `json:"address"`
	Data    []byte `json:"data"`
}

// MsgExecuteContract executes a contract
type MsgExecuteContract struct {
	Sender   string          `json:"sender"`
	Contract string          `json:"contract"`
	Msg      json.RawMessage `json:"msg"`
	Funds    sdk.Coins       `json:"funds"`
}

// MsgExecuteContractResponse is the response for MsgExecuteContract
type MsgExecuteContractResponse struct {
	Data []byte `json:"data"`
}

// MsgMigrateContract migrates a contract to a new code version
type MsgMigrateContract struct {
	Sender   string          `json:"sender"`
	Contract string          `json:"contract"`
	CodeID   uint64          `json:"code_id"`
	Msg      json.RawMessage `json:"msg"`
}

// MsgMigrateContractResponse is the response for MsgMigrateContract
type MsgMigrateContractResponse struct {
	Data []byte `json:"data"`
}

// MsgUpdateAdmin updates contract admin
type MsgUpdateAdmin struct {
	Sender   string `json:"sender"`
	Contract string `json:"contract"`
	NewAdmin string `json:"new_admin"`
}

// MsgUpdateAdminResponse is the response for MsgUpdateAdmin
type MsgUpdateAdminResponse struct{}

// MsgClearAdmin clears contract admin
type MsgClearAdmin struct {
	Sender   string `json:"sender"`
	Contract string `json:"contract"`
}

// MsgClearAdminResponse is the response for MsgClearAdmin
type MsgClearAdminResponse struct{}

// MsgAuthorizeUploader authorizes a contract uploader
type MsgAuthorizeUploader struct {
	Authority string `json:"authority"`
	Uploader  string `json:"uploader"`
}

// MsgAuthorizeUploaderResponse is the response for MsgAuthorizeUploader
type MsgAuthorizeUploaderResponse struct{}

// MsgRevokeUploader revokes a contract uploader
type MsgRevokeUploader struct {
	Authority string `json:"authority"`
	Uploader  string `json:"uploader"`
}

// MsgRevokeUploaderResponse is the response for MsgRevokeUploader
type MsgRevokeUploaderResponse struct{}

// MsgPauseContract pauses a contract
type MsgPauseContract struct {
	Authority string `json:"authority"`
	Contract  string `json:"contract"`
}

// MsgPauseContractResponse is the response for MsgPauseContract
type MsgPauseContractResponse struct{}

// MsgUnpauseContract unpauses a contract
type MsgUnpauseContract struct {
	Authority string `json:"authority"`
	Contract  string `json:"contract"`
}

// MsgUnpauseContractResponse is the response for MsgUnpauseContract
type MsgUnpauseContractResponse struct{}

// MsgUpdateParams updates module parameters
type MsgUpdateParams struct {
	Authority string `json:"authority"`
	Params    Params `json:"params"`
}

// MsgUpdateParamsResponse is the response for MsgUpdateParams
type MsgUpdateParamsResponse struct{}

// ValidateBasic implementations

func (m MsgStoreCode) ValidateBasic() error {
	if m.Sender == "" {
		return ErrUnauthorized.Wrap("sender cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return ErrUnauthorized.Wrapf("invalid sender address: %s", err)
	}
	if len(m.WASMByteCode) == 0 {
		return ErrInvalidContractCode.Wrap("wasm byte code cannot be empty")
	}
	return nil
}

func (m MsgInstantiateContract) ValidateBasic() error {
	if m.Sender == "" {
		return ErrUnauthorized.Wrap("sender cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return ErrUnauthorized.Wrapf("invalid sender address: %s", err)
	}
	if m.Admin != "" {
		if _, err := sdk.AccAddressFromBech32(m.Admin); err != nil {
			return ErrInvalidAdmin.Wrapf("invalid admin address: %s", err)
		}
	}
	if m.CodeID == 0 {
		return ErrInvalidContractCode.Wrap("code id cannot be zero")
	}
	if m.Label == "" {
		return ErrUnauthorized.Wrap("label cannot be empty")
	}
	if len(m.Msg) == 0 {
		return ErrUnauthorized.Wrap("init message cannot be empty")
	}
	if !m.Funds.IsValid() {
		return ErrUnauthorized.Wrap("invalid funds")
	}
	return nil
}

func (m MsgExecuteContract) ValidateBasic() error {
	if m.Sender == "" {
		return ErrUnauthorized.Wrap("sender cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return ErrUnauthorized.Wrapf("invalid sender address: %s", err)
	}
	if m.Contract == "" {
		return ErrInvalidContractAddress.Wrap("contract cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(m.Contract); err != nil {
		return ErrInvalidContractAddress.Wrapf("invalid contract address: %s", err)
	}
	if len(m.Msg) == 0 {
		return ErrUnauthorized.Wrap("execute message cannot be empty")
	}
	if !m.Funds.IsValid() {
		return ErrUnauthorized.Wrap("invalid funds")
	}
	return nil
}

func (m MsgMigrateContract) ValidateBasic() error {
	if m.Sender == "" {
		return ErrUnauthorized.Wrap("sender cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return ErrUnauthorized.Wrapf("invalid sender address: %s", err)
	}
	if m.Contract == "" {
		return ErrInvalidContractAddress.Wrap("contract cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(m.Contract); err != nil {
		return ErrInvalidContractAddress.Wrapf("invalid contract address: %s", err)
	}
	if m.CodeID == 0 {
		return ErrInvalidContractCode.Wrap("code id cannot be zero")
	}
	if len(m.Msg) == 0 {
		return ErrUnauthorized.Wrap("migrate message cannot be empty")
	}
	return nil
}

func (m MsgUpdateAdmin) ValidateBasic() error {
	if m.Sender == "" {
		return ErrUnauthorized.Wrap("sender cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return ErrUnauthorized.Wrapf("invalid sender address: %s", err)
	}
	if m.Contract == "" {
		return ErrInvalidContractAddress.Wrap("contract cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(m.Contract); err != nil {
		return ErrInvalidContractAddress.Wrapf("invalid contract address: %s", err)
	}
	if m.NewAdmin == "" {
		return ErrInvalidAdmin.Wrap("new admin cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(m.NewAdmin); err != nil {
		return ErrInvalidAdmin.Wrapf("invalid new admin address: %s", err)
	}
	return nil
}

func (m MsgClearAdmin) ValidateBasic() error {
	if m.Sender == "" {
		return ErrUnauthorized.Wrap("sender cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return ErrUnauthorized.Wrapf("invalid sender address: %s", err)
	}
	if m.Contract == "" {
		return ErrInvalidContractAddress.Wrap("contract cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(m.Contract); err != nil {
		return ErrInvalidContractAddress.Wrapf("invalid contract address: %s", err)
	}
	return nil
}

func (m MsgAuthorizeUploader) ValidateBasic() error {
	if m.Authority == "" {
		return ErrUnauthorized.Wrap("authority cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return ErrUnauthorized.Wrapf("invalid authority address: %s", err)
	}
	if m.Uploader == "" {
		return ErrUnauthorized.Wrap("uploader cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(m.Uploader); err != nil {
		return ErrUnauthorized.Wrapf("invalid uploader address: %s", err)
	}
	return nil
}

func (m MsgRevokeUploader) ValidateBasic() error {
	if m.Authority == "" {
		return ErrUnauthorized.Wrap("authority cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return ErrUnauthorized.Wrapf("invalid authority address: %s", err)
	}
	if m.Uploader == "" {
		return ErrUnauthorized.Wrap("uploader cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(m.Uploader); err != nil {
		return ErrUnauthorized.Wrapf("invalid uploader address: %s", err)
	}
	return nil
}

func (m MsgPauseContract) ValidateBasic() error {
	if m.Authority == "" {
		return ErrUnauthorized.Wrap("authority cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return ErrUnauthorized.Wrapf("invalid authority address: %s", err)
	}
	if m.Contract == "" {
		return ErrInvalidContractAddress.Wrap("contract cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(m.Contract); err != nil {
		return ErrInvalidContractAddress.Wrapf("invalid contract address: %s", err)
	}
	return nil
}

func (m MsgUnpauseContract) ValidateBasic() error {
	if m.Authority == "" {
		return ErrUnauthorized.Wrap("authority cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return ErrUnauthorized.Wrapf("invalid authority address: %s", err)
	}
	if m.Contract == "" {
		return ErrInvalidContractAddress.Wrap("contract cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(m.Contract); err != nil {
		return ErrInvalidContractAddress.Wrapf("invalid contract address: %s", err)
	}
	return nil
}

func (m MsgUpdateParams) ValidateBasic() error {
	if m.Authority == "" {
		return ErrUnauthorized.Wrap("authority cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return ErrUnauthorized.Wrapf("invalid authority address: %s", err)
	}
	return m.Params.Validate()
}

// Implement proto.Message interface for all message types

func (m *MsgStoreCode) Reset()         { *m = MsgStoreCode{} }
func (m *MsgStoreCode) String() string { return fmt.Sprintf("MsgStoreCode{Sender: %s}", m.Sender) }
func (m *MsgStoreCode) ProtoMessage()  {}

func (m *MsgInstantiateContract) Reset() { *m = MsgInstantiateContract{} }
func (m *MsgInstantiateContract) String() string {
	return fmt.Sprintf("MsgInstantiateContract{Sender: %s, CodeID: %d}", m.Sender, m.CodeID)
}
func (m *MsgInstantiateContract) ProtoMessage() {}

func (m *MsgExecuteContract) Reset() { *m = MsgExecuteContract{} }
func (m *MsgExecuteContract) String() string {
	return fmt.Sprintf("MsgExecuteContract{Sender: %s, Contract: %s}", m.Sender, m.Contract)
}
func (m *MsgExecuteContract) ProtoMessage() {}

func (m *MsgMigrateContract) Reset() { *m = MsgMigrateContract{} }
func (m *MsgMigrateContract) String() string {
	return fmt.Sprintf("MsgMigrateContract{Sender: %s, Contract: %s, CodeID: %d}", m.Sender, m.Contract, m.CodeID)
}
func (m *MsgMigrateContract) ProtoMessage() {}

func (m *MsgUpdateAdmin) Reset()         { *m = MsgUpdateAdmin{} }
func (m *MsgUpdateAdmin) String() string { return fmt.Sprintf("MsgUpdateAdmin{Sender: %s}", m.Sender) }
func (m *MsgUpdateAdmin) ProtoMessage()  {}

func (m *MsgClearAdmin) Reset()         { *m = MsgClearAdmin{} }
func (m *MsgClearAdmin) String() string { return fmt.Sprintf("MsgClearAdmin{Sender: %s}", m.Sender) }
func (m *MsgClearAdmin) ProtoMessage()  {}

func (m *MsgAuthorizeUploader) Reset() { *m = MsgAuthorizeUploader{} }
func (m *MsgAuthorizeUploader) String() string {
	return fmt.Sprintf("MsgAuthorizeUploader{Authority: %s, Uploader: %s}", m.Authority, m.Uploader)
}
func (m *MsgAuthorizeUploader) ProtoMessage() {}

func (m *MsgRevokeUploader) Reset() { *m = MsgRevokeUploader{} }
func (m *MsgRevokeUploader) String() string {
	return fmt.Sprintf("MsgRevokeUploader{Authority: %s, Uploader: %s}", m.Authority, m.Uploader)
}
func (m *MsgRevokeUploader) ProtoMessage() {}

func (m *MsgPauseContract) Reset() { *m = MsgPauseContract{} }
func (m *MsgPauseContract) String() string {
	return fmt.Sprintf("MsgPauseContract{Authority: %s, Contract: %s}", m.Authority, m.Contract)
}
func (m *MsgPauseContract) ProtoMessage() {}

func (m *MsgUnpauseContract) Reset() { *m = MsgUnpauseContract{} }
func (m *MsgUnpauseContract) String() string {
	return fmt.Sprintf("MsgUnpauseContract{Authority: %s, Contract: %s}", m.Authority, m.Contract)
}
func (m *MsgUnpauseContract) ProtoMessage() {}

func (m *MsgUpdateParams) Reset()         { *m = MsgUpdateParams{} }
func (m *MsgUpdateParams) String() string { return fmt.Sprintf("MsgUpdateParams{Authority: %s}", m.Authority) }
func (m *MsgUpdateParams) ProtoMessage()  {}
