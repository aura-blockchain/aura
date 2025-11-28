package types

import (
	"context"
)

// MsgServer is the server API for Msg service
type MsgServer interface {
	StoreCode(context.Context, *MsgStoreCode) (*MsgStoreCodeResponse, error)
	InstantiateContract(context.Context, *MsgInstantiateContract) (*MsgInstantiateContractResponse, error)
	ExecuteContract(context.Context, *MsgExecuteContract) (*MsgExecuteContractResponse, error)
	MigrateContract(context.Context, *MsgMigrateContract) (*MsgMigrateContractResponse, error)
	UpdateAdmin(context.Context, *MsgUpdateAdmin) (*MsgUpdateAdminResponse, error)
	ClearAdmin(context.Context, *MsgClearAdmin) (*MsgClearAdminResponse, error)
	AuthorizeUploader(context.Context, *MsgAuthorizeUploader) (*MsgAuthorizeUploaderResponse, error)
	RevokeUploader(context.Context, *MsgRevokeUploader) (*MsgRevokeUploaderResponse, error)
	PauseContract(context.Context, *MsgPauseContract) (*MsgPauseContractResponse, error)
	UnpauseContract(context.Context, *MsgUnpauseContract) (*MsgUnpauseContractResponse, error)
	UpdateParams(context.Context, *MsgUpdateParams) (*MsgUpdateParamsResponse, error)
}

// QueryServer is the server API for Query service
type QueryServer interface {
	Params(context.Context, *QueryParamsRequest) (*QueryParamsResponse, error)
	Code(context.Context, *QueryCodeRequest) (*QueryCodeResponse, error)
	Codes(context.Context, *QueryCodesRequest) (*QueryCodesResponse, error)
	ContractInfo(context.Context, *QueryContractInfoRequest) (*QueryContractInfoResponse, error)
	ContractHistory(context.Context, *QueryContractHistoryRequest) (*QueryContractHistoryResponse, error)
	AllContractState(context.Context, *QueryAllContractStateRequest) (*QueryAllContractStateResponse, error)
	RawContractState(context.Context, *QueryRawContractStateRequest) (*QueryRawContractStateResponse, error)
	SmartContractState(context.Context, *QuerySmartContractStateRequest) (*QuerySmartContractStateResponse, error)
	SecurityStats(context.Context, *QuerySecurityStatsRequest) (*QuerySecurityStatsResponse, error)
	AuthorizedUploaders(context.Context, *QueryAuthorizedUploadersRequest) (*QueryAuthorizedUploadersResponse, error)
	PausedContracts(context.Context, *QueryPausedContractsRequest) (*QueryPausedContractsResponse, error)
	IsAuthorizedUploader(context.Context, *QueryIsAuthorizedUploaderRequest) (*QueryIsAuthorizedUploaderResponse, error)
	IsContractPaused(context.Context, *QueryIsContractPausedRequest) (*QueryIsContractPausedResponse, error)
}

// QueryClient is the client API for Query service (stub for now)
type QueryClient interface {
	Params(ctx context.Context, in *QueryParamsRequest) (*QueryParamsResponse, error)
	Code(ctx context.Context, in *QueryCodeRequest) (*QueryCodeResponse, error)
	Codes(ctx context.Context, in *QueryCodesRequest) (*QueryCodesResponse, error)
	ContractInfo(ctx context.Context, in *QueryContractInfoRequest) (*QueryContractInfoResponse, error)
	ContractHistory(ctx context.Context, in *QueryContractHistoryRequest) (*QueryContractHistoryResponse, error)
	AllContractState(ctx context.Context, in *QueryAllContractStateRequest) (*QueryAllContractStateResponse, error)
	RawContractState(ctx context.Context, in *QueryRawContractStateRequest) (*QueryRawContractStateResponse, error)
	SmartContractState(ctx context.Context, in *QuerySmartContractStateRequest) (*QuerySmartContractStateResponse, error)
	SecurityStats(ctx context.Context, in *QuerySecurityStatsRequest) (*QuerySecurityStatsResponse, error)
	AuthorizedUploaders(ctx context.Context, in *QueryAuthorizedUploadersRequest) (*QueryAuthorizedUploadersResponse, error)
	PausedContracts(ctx context.Context, in *QueryPausedContractsRequest) (*QueryPausedContractsResponse, error)
	IsAuthorizedUploader(ctx context.Context, in *QueryIsAuthorizedUploaderRequest) (*QueryIsAuthorizedUploaderResponse, error)
	IsContractPaused(ctx context.Context, in *QueryIsContractPausedRequest) (*QueryIsContractPausedResponse, error)
}

// NewQueryClient creates a stub query client
func NewQueryClient(cc interface{}) QueryClient {
	return &queryClient{}
}

type queryClient struct{}

func (c *queryClient) Params(ctx context.Context, in *QueryParamsRequest) (*QueryParamsResponse, error) {
	return nil, nil
}

func (c *queryClient) Code(ctx context.Context, in *QueryCodeRequest) (*QueryCodeResponse, error) {
	return nil, nil
}

func (c *queryClient) Codes(ctx context.Context, in *QueryCodesRequest) (*QueryCodesResponse, error) {
	return nil, nil
}

func (c *queryClient) ContractInfo(ctx context.Context, in *QueryContractInfoRequest) (*QueryContractInfoResponse, error) {
	return nil, nil
}

func (c *queryClient) ContractHistory(ctx context.Context, in *QueryContractHistoryRequest) (*QueryContractHistoryResponse, error) {
	return nil, nil
}

func (c *queryClient) AllContractState(ctx context.Context, in *QueryAllContractStateRequest) (*QueryAllContractStateResponse, error) {
	return nil, nil
}

func (c *queryClient) RawContractState(ctx context.Context, in *QueryRawContractStateRequest) (*QueryRawContractStateResponse, error) {
	return nil, nil
}

func (c *queryClient) SmartContractState(ctx context.Context, in *QuerySmartContractStateRequest) (*QuerySmartContractStateResponse, error) {
	return nil, nil
}

func (c *queryClient) SecurityStats(ctx context.Context, in *QuerySecurityStatsRequest) (*QuerySecurityStatsResponse, error) {
	return nil, nil
}

func (c *queryClient) AuthorizedUploaders(ctx context.Context, in *QueryAuthorizedUploadersRequest) (*QueryAuthorizedUploadersResponse, error) {
	return nil, nil
}

func (c *queryClient) PausedContracts(ctx context.Context, in *QueryPausedContractsRequest) (*QueryPausedContractsResponse, error) {
	return nil, nil
}

func (c *queryClient) IsAuthorizedUploader(ctx context.Context, in *QueryIsAuthorizedUploaderRequest) (*QueryIsAuthorizedUploaderResponse, error) {
	return nil, nil
}

func (c *queryClient) IsContractPaused(ctx context.Context, in *QueryIsContractPausedRequest) (*QueryIsContractPausedResponse, error) {
	return nil, nil
}
