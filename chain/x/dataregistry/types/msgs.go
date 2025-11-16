package types

import (
	"fmt"
)

// Message type URLs
const (
	TypeMsgStoreDataItem  = "store_data_item"
	TypeMsgUpdateDataItem = "update_data_item"
	TypeMsgDeleteDataItem = "delete_data_item"
	TypeMsgVerifyDataItem = "verify_data_item"
	TypeMsgRevokeDataItem = "revoke_data_item"
)

// MsgStoreDataItem
type MsgStoreDataItem struct {
	Creator         string
	DataType        DataItemType
	Title           string
	Description     string
	ContentHash     []byte
	StorageLocation string
	IsEncrypted     bool
	GeoLocation     *GeoLocation
	Metadata        map[string]string
	AccessPolicy    *AccessPolicy
	Tags            []string
}

func (msg *MsgStoreDataItem) ValidateBasic() error {
	if msg.Creator == "" {
		return fmt.Errorf("creator cannot be empty")
	}
	if len(msg.ContentHash) == 0 {
		return fmt.Errorf("content hash cannot be empty")
	}
	if msg.StorageLocation == "" {
		return fmt.Errorf("storage location cannot be empty")
	}
	return nil
}

// MsgUpdateDataItem
type MsgUpdateDataItem struct {
	Creator      string
	DataID       string
	Title        string
	Description  string
	Metadata     map[string]string
	AccessPolicy *AccessPolicy
	Tags         []string
}

func (msg *MsgUpdateDataItem) ValidateBasic() error {
	if msg.Creator == "" {
		return fmt.Errorf("creator cannot be empty")
	}
	if msg.DataID == "" {
		return fmt.Errorf("data ID cannot be empty")
	}
	return nil
}

// MsgDeleteDataItem
type MsgDeleteDataItem struct {
	Creator string
	DataID  string
}

func (msg *MsgDeleteDataItem) ValidateBasic() error {
	if msg.Creator == "" {
		return fmt.Errorf("creator cannot be empty")
	}
	if msg.DataID == "" {
		return fmt.Errorf("data ID cannot be empty")
	}
	return nil
}

// MsgVerifyDataItem
type MsgVerifyDataItem struct {
	Verifier           string
	DataID             string
	Level              VerificationLevel
	ConfidenceScore    uint64
	Notes              string
	VerificationMethod string
	Proof              []byte
}

func (msg *MsgVerifyDataItem) ValidateBasic() error {
	if msg.Verifier == "" {
		return fmt.Errorf("verifier cannot be empty")
	}
	if msg.DataID == "" {
		return fmt.Errorf("data ID cannot be empty")
	}
	if msg.ConfidenceScore > 100 {
		return fmt.Errorf("confidence score must be between 0 and 100")
	}
	return nil
}

// MsgRevokeDataItem
type MsgRevokeDataItem struct {
	Authority string
	DataID    string
	Reason    string
}

func (msg *MsgRevokeDataItem) ValidateBasic() error {
	if msg.Authority == "" {
		return fmt.Errorf("authority cannot be empty")
	}
	if msg.DataID == "" {
		return fmt.Errorf("data ID cannot be empty")
	}
	return nil
}

// Query request/response types
type QueryGetDataItemRequest struct {
	DataId    string
	Requester string
}

type QueryGetDataItemResponse struct {
	DataItem DataItem
}

type QueryListDataItemsRequest struct {
	OwnerAddress string
	TypeFilter   DataItemType
	StatusFilter DataItemStatus
}

type QueryListDataItemsResponse struct {
	DataItems []DataItem
}

type QuerySearchDataItemsRequest struct {
	Query       string
	Tags        []string
	TypeFilter  DataItemType
	GeoLocation *GeoLocation
	RadiusKm    float64
	Requester   string
}

type QuerySearchDataItemsResponse struct {
	DataItems []DataItem
}

type QueryGetStatsRequest struct{}

type QueryGetStatsResponse struct {
	Stats RegistryStats
}

type QueryGetParamsRequest struct{}

type QueryGetParamsResponse struct {
	Params Params
}

// QueryClient interface
type QueryClient interface {
	GetDataItem(req *QueryGetDataItemRequest) (*QueryGetDataItemResponse, error)
	ListDataItems(req *QueryListDataItemsRequest) (*QueryListDataItemsResponse, error)
	SearchDataItems(req *QuerySearchDataItemsRequest) (*QuerySearchDataItemsResponse, error)
	GetStats(req *QueryGetStatsRequest) (*QueryGetStatsResponse, error)
	GetParams(req *QueryGetParamsRequest) (*QueryGetParamsResponse, error)
}

// NewQueryClient stub for CLI
func NewQueryClient(clientCtx interface{}) QueryClient {
	return nil // This would be implemented properly in production
}
