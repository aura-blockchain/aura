package types

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/types/query"
)

// Query request and response types

// QueryParamsRequest is the request for QueryParams
type QueryParamsRequest struct{}

// QueryParamsResponse is the response for QueryParams
type QueryParamsResponse struct {
	Params Params `json:"params"`
}

// QueryCodeRequest is the request for QueryCode
type QueryCodeRequest struct {
	CodeId uint64 `json:"code_id"`
}

// QueryCodeResponse is the response for QueryCode
type QueryCodeResponse struct {
	CodeID uint64 `json:"code_id"`
	Data   []byte `json:"data"`
}

// QueryCodesRequest is the request for QueryCodes
type QueryCodesRequest struct {
	Pagination *query.PageRequest `json:"pagination,omitempty"`
}

// QueryCodesResponse is the response for QueryCodes
type QueryCodesResponse struct {
	CodeInfos  []CodeInfo          `json:"code_infos"`
	Pagination *query.PageResponse `json:"pagination,omitempty"`
}

// QueryContractInfoRequest is the request for QueryContractInfo
type QueryContractInfoRequest struct {
	Address string `json:"address"`
}

// QueryContractInfoResponse is the response for QueryContractInfo
type QueryContractInfoResponse struct {
	Address  string `json:"address"`
	CodeID   uint64 `json:"code_id"`
	Creator  string `json:"creator"`
	Admin    string `json:"admin,omitempty"`
	Label    string `json:"label"`
	IsPaused bool   `json:"is_paused"`
}

// QueryContractHistoryRequest is the request for QueryContractHistory
type QueryContractHistoryRequest struct {
	Address    string              `json:"address"`
	Pagination *query.PageRequest  `json:"pagination,omitempty"`
}

// QueryContractHistoryResponse is the response for QueryContractHistory
type QueryContractHistoryResponse struct {
	Entries    []ContractHistoryEntry `json:"entries"`
	Pagination *query.PageResponse    `json:"pagination,omitempty"`
}

// QueryAllContractStateRequest is the request for QueryAllContractState
type QueryAllContractStateRequest struct {
	Address    string             `json:"address"`
	Pagination *query.PageRequest `json:"pagination,omitempty"`
}

// QueryAllContractStateResponse is the response for QueryAllContractState
type QueryAllContractStateResponse struct {
	Models     []Model             `json:"models"`
	Pagination *query.PageResponse `json:"pagination,omitempty"`
}

// QueryRawContractStateRequest is the request for QueryRawContractState
type QueryRawContractStateRequest struct {
	Address   string `json:"address"`
	QueryData []byte `json:"query_data"`
}

// QueryRawContractStateResponse is the response for QueryRawContractState
type QueryRawContractStateResponse struct {
	Data []byte `json:"data"`
}

// QuerySmartContractStateRequest is the request for QuerySmartContractState
type QuerySmartContractStateRequest struct {
	Address   string          `json:"address"`
	QueryData json.RawMessage `json:"query_data"`
}

// QuerySmartContractStateResponse is the response for QuerySmartContractState
type QuerySmartContractStateResponse struct {
	Data json.RawMessage `json:"data"`
}

// QuerySecurityStatsRequest is the request for QuerySecurityStats
type QuerySecurityStatsRequest struct{}

// QuerySecurityStatsResponse is the response for QuerySecurityStats
type QuerySecurityStatsResponse struct {
	Stats SecurityStats `json:"stats"`
}

// QueryAuthorizedUploadersRequest is the request for QueryAuthorizedUploaders
type QueryAuthorizedUploadersRequest struct {
	Pagination *query.PageRequest `json:"pagination,omitempty"`
}

// QueryAuthorizedUploadersResponse is the response for QueryAuthorizedUploaders
type QueryAuthorizedUploadersResponse struct {
	Uploaders  []string            `json:"uploaders"`
	Pagination *query.PageResponse `json:"pagination,omitempty"`
}

// QueryPausedContractsRequest is the request for QueryPausedContracts
type QueryPausedContractsRequest struct {
	Pagination *query.PageRequest `json:"pagination,omitempty"`
}

// QueryPausedContractsResponse is the response for QueryPausedContracts
type QueryPausedContractsResponse struct {
	Contracts  []string            `json:"contracts"`
	Pagination *query.PageResponse `json:"pagination,omitempty"`
}

// QueryIsAuthorizedUploaderRequest is the request for QueryIsAuthorizedUploader
type QueryIsAuthorizedUploaderRequest struct {
	Address string `json:"address"`
}

// QueryIsAuthorizedUploaderResponse is the response for QueryIsAuthorizedUploader
type QueryIsAuthorizedUploaderResponse struct {
	IsAuthorized bool `json:"is_authorized"`
}

// QueryIsContractPausedRequest is the request for QueryIsContractPaused
type QueryIsContractPausedRequest struct {
	Address string `json:"address"`
}

// QueryIsContractPausedResponse is the response for QueryIsContractPaused
type QueryIsContractPausedResponse struct {
	IsPaused bool `json:"is_paused"`
}

// Helper types

// CodeInfo contains code metadata
type CodeInfo struct {
	CodeID   uint64 `json:"code_id"`
	Creator  string `json:"creator"`
	CodeHash []byte `json:"code_hash"`
}

// ContractHistoryEntry represents a contract code version update
type ContractHistoryEntry struct {
	Operation string `json:"operation"`
	CodeID    uint64 `json:"code_id"`
	Updated   int64  `json:"updated"`
	Msg       []byte `json:"msg,omitempty"`
}

// Model represents a key-value pair in contract storage
type Model struct {
	Key   []byte `json:"key"`
	Value []byte `json:"value"`
}
