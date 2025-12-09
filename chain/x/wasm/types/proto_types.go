// Package types provides WASM module type definitions.
// This file re-exports generated protobuf types from the proto package
// for use throughout the WASM module implementation.
package types

import (
	wasmpb "github.com/aequitas/aura/proto/aura/wasm/v1beta1"
)

// Re-export generated query types as type aliases
type (
	// Query request/response types
	QueryParamsRequest               = wasmpb.QueryParamsRequest
	QueryParamsResponse              = wasmpb.QueryParamsResponse
	QueryCodeRequest                 = wasmpb.QueryCodeRequest
	QueryCodeResponse                = wasmpb.QueryCodeResponse
	QueryCodesRequest                = wasmpb.QueryCodesRequest
	QueryCodesResponse               = wasmpb.QueryCodesResponse
	QueryContractInfoRequest         = wasmpb.QueryContractInfoRequest
	QueryContractInfoResponse        = wasmpb.QueryContractInfoResponse
	QueryContractHistoryRequest      = wasmpb.QueryContractHistoryRequest
	QueryContractHistoryResponse     = wasmpb.QueryContractHistoryResponse
	QueryAllContractStateRequest     = wasmpb.QueryAllContractStateRequest
	QueryAllContractStateResponse    = wasmpb.QueryAllContractStateResponse
	QueryRawContractStateRequest     = wasmpb.QueryRawContractStateRequest
	QueryRawContractStateResponse    = wasmpb.QueryRawContractStateResponse
	QuerySmartContractStateRequest   = wasmpb.QuerySmartContractStateRequest
	QuerySmartContractStateResponse  = wasmpb.QuerySmartContractStateResponse
	QuerySecurityStatsRequest        = wasmpb.QuerySecurityStatsRequest
	QuerySecurityStatsResponse       = wasmpb.QuerySecurityStatsResponse
	QueryAuthorizedUploadersRequest  = wasmpb.QueryAuthorizedUploadersRequest
	QueryAuthorizedUploadersResponse = wasmpb.QueryAuthorizedUploadersResponse
	QueryPausedContractsRequest      = wasmpb.QueryPausedContractsRequest
	QueryPausedContractsResponse     = wasmpb.QueryPausedContractsResponse
	QueryIsAuthorizedUploaderRequest  = wasmpb.QueryIsAuthorizedUploaderRequest
	QueryIsAuthorizedUploaderResponse = wasmpb.QueryIsAuthorizedUploaderResponse
	QueryIsContractPausedRequest      = wasmpb.QueryIsContractPausedRequest
	QueryIsContractPausedResponse     = wasmpb.QueryIsContractPausedResponse
	QueryContractAdminRequest         = wasmpb.QueryContractAdminRequest
	QueryContractAdminResponse        = wasmpb.QueryContractAdminResponse

	// Message types
	MsgStoreCode                    = wasmpb.MsgStoreCode
	MsgStoreCodeResponse            = wasmpb.MsgStoreCodeResponse
	MsgInstantiateContract          = wasmpb.MsgInstantiateContract
	MsgInstantiateContractResponse  = wasmpb.MsgInstantiateContractResponse
	MsgExecuteContract              = wasmpb.MsgExecuteContract
	MsgExecuteContractResponse      = wasmpb.MsgExecuteContractResponse
	MsgMigrateContract              = wasmpb.MsgMigrateContract
	MsgMigrateContractResponse      = wasmpb.MsgMigrateContractResponse
	MsgUpdateAdmin                  = wasmpb.MsgUpdateAdmin
	MsgUpdateAdminResponse          = wasmpb.MsgUpdateAdminResponse
	MsgClearAdmin                   = wasmpb.MsgClearAdmin
	MsgClearAdminResponse           = wasmpb.MsgClearAdminResponse
	MsgAuthorizeUploader            = wasmpb.MsgAuthorizeUploader
	MsgAuthorizeUploaderResponse    = wasmpb.MsgAuthorizeUploaderResponse
	MsgRevokeUploader               = wasmpb.MsgRevokeUploader
	MsgRevokeUploaderResponse       = wasmpb.MsgRevokeUploaderResponse
	MsgPauseContract                = wasmpb.MsgPauseContract
	MsgPauseContractResponse        = wasmpb.MsgPauseContractResponse
	MsgUnpauseContract              = wasmpb.MsgUnpauseContract
	MsgUnpauseContractResponse      = wasmpb.MsgUnpauseContractResponse
	MsgUpdateParams                 = wasmpb.MsgUpdateParams
	MsgUpdateParamsResponse         = wasmpb.MsgUpdateParamsResponse

	// Core types
	Params                           = wasmpb.Params
	CodeInfo                         = wasmpb.CodeInfo
	ContractInfo                     = wasmpb.ContractInfo
	ContractHistoryEntry             = wasmpb.ContractHistoryEntry
	Model                            = wasmpb.Model
	SecurityStats                    = wasmpb.SecurityStats
	AccessConfig                     = wasmpb.AccessConfig
	AbsoluteTxPosition               = wasmpb.AbsoluteTxPosition

	// Genesis types
	GenesisState                     = wasmpb.GenesisState
	Code                             = wasmpb.Code
	Contract                         = wasmpb.Contract
	Sequence                         = wasmpb.Sequence

	// Enums
	AccessType                       = wasmpb.AccessType
	ContractCodeHistoryOperationType = wasmpb.ContractCodeHistoryOperationType

	// Custom types
	RawContractMessage = wasmpb.RawContractMessage

	// Service interfaces from generated gRPC
	QueryServer                      = wasmpb.QueryServer
	MsgServer                        = wasmpb.MsgServer
	UnimplementedQueryServer         = wasmpb.UnimplementedQueryServer
	UnimplementedMsgServer           = wasmpb.UnimplementedMsgServer
)

// AccessType constants
const (
	AccessTypeUnspecified   = wasmpb.AccessType_ACCESS_TYPE_UNSPECIFIED
	AccessTypeNobody        = wasmpb.AccessType_ACCESS_TYPE_NOBODY
	AccessTypeOnlyAddress   = wasmpb.AccessType_ACCESS_TYPE_ONLY_ADDRESS
	AccessTypeEverybody     = wasmpb.AccessType_ACCESS_TYPE_EVERYBODY
	AccessTypeAnyOfAddresses = wasmpb.AccessType_ACCESS_TYPE_ANY_OF_ADDRESSES
)

// ContractCodeHistoryOperationType constants
const (
	ContractCodeHistoryOperationTypeUnspecified = wasmpb.ContractCodeHistoryOperationType_CONTRACT_CODE_HISTORY_OPERATION_TYPE_UNSPECIFIED
	ContractCodeHistoryOperationTypeInit        = wasmpb.ContractCodeHistoryOperationType_CONTRACT_CODE_HISTORY_OPERATION_TYPE_INIT
	ContractCodeHistoryOperationTypeMigrate     = wasmpb.ContractCodeHistoryOperationType_CONTRACT_CODE_HISTORY_OPERATION_TYPE_MIGRATE
	ContractCodeHistoryOperationTypeGenesis     = wasmpb.ContractCodeHistoryOperationType_CONTRACT_CODE_HISTORY_OPERATION_TYPE_GENESIS
)

// Re-export gRPC registration functions
var (
	RegisterQueryServer = wasmpb.RegisterQueryServer
	RegisterMsgServer   = wasmpb.RegisterMsgServer
)

// Re-export gRPC client constructor
var NewQueryClient = wasmpb.NewQueryClient

// Re-export service descriptors for gRPC
var (
	Query_ServiceDesc = wasmpb.Query_ServiceDesc
	Msg_ServiceDesc   = wasmpb.Msg_serviceDesc
)
