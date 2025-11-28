package types

import "fmt"

// ProtoMessage implementations for query request/response types

// QueryParamsRequest
func (m *QueryParamsRequest) Reset()         { *m = QueryParamsRequest{} }
func (m *QueryParamsRequest) String() string { return "QueryParamsRequest{}" }
func (m *QueryParamsRequest) ProtoMessage()  {}

// QueryParamsResponse
func (m *QueryParamsResponse) Reset() { *m = QueryParamsResponse{} }
func (m *QueryParamsResponse) String() string {
	return fmt.Sprintf("QueryParamsResponse{Params: %+v}", m.Params)
}
func (m *QueryParamsResponse) ProtoMessage() {}

// QueryCodeRequest
func (m *QueryCodeRequest) Reset() { *m = QueryCodeRequest{} }
func (m *QueryCodeRequest) String() string {
	return fmt.Sprintf("QueryCodeRequest{CodeId: %d}", m.CodeId)
}
func (m *QueryCodeRequest) ProtoMessage() {}

// QueryCodeResponse
func (m *QueryCodeResponse) Reset() { *m = QueryCodeResponse{} }
func (m *QueryCodeResponse) String() string {
	return fmt.Sprintf("QueryCodeResponse{CodeID: %d}", m.CodeID)
}
func (m *QueryCodeResponse) ProtoMessage() {}

// QueryCodesRequest
func (m *QueryCodesRequest) Reset()         { *m = QueryCodesRequest{} }
func (m *QueryCodesRequest) String() string { return "QueryCodesRequest{}" }
func (m *QueryCodesRequest) ProtoMessage()  {}

// QueryCodesResponse
func (m *QueryCodesResponse) Reset() { *m = QueryCodesResponse{} }
func (m *QueryCodesResponse) String() string {
	return fmt.Sprintf("QueryCodesResponse{Count: %d}", len(m.CodeInfos))
}
func (m *QueryCodesResponse) ProtoMessage() {}

// QueryContractInfoRequest
func (m *QueryContractInfoRequest) Reset() { *m = QueryContractInfoRequest{} }
func (m *QueryContractInfoRequest) String() string {
	return fmt.Sprintf("QueryContractInfoRequest{Address: %s}", m.Address)
}
func (m *QueryContractInfoRequest) ProtoMessage() {}

// QueryContractInfoResponse
func (m *QueryContractInfoResponse) Reset() { *m = QueryContractInfoResponse{} }
func (m *QueryContractInfoResponse) String() string {
	return fmt.Sprintf("QueryContractInfoResponse{Address: %s}", m.Address)
}
func (m *QueryContractInfoResponse) ProtoMessage() {}

// QueryContractHistoryRequest
func (m *QueryContractHistoryRequest) Reset() { *m = QueryContractHistoryRequest{} }
func (m *QueryContractHistoryRequest) String() string {
	return fmt.Sprintf("QueryContractHistoryRequest{Address: %s}", m.Address)
}
func (m *QueryContractHistoryRequest) ProtoMessage() {}

// QueryContractHistoryResponse
func (m *QueryContractHistoryResponse) Reset() { *m = QueryContractHistoryResponse{} }
func (m *QueryContractHistoryResponse) String() string {
	return fmt.Sprintf("QueryContractHistoryResponse{Entries: %d}", len(m.Entries))
}
func (m *QueryContractHistoryResponse) ProtoMessage() {}

// QueryAllContractStateRequest
func (m *QueryAllContractStateRequest) Reset() { *m = QueryAllContractStateRequest{} }
func (m *QueryAllContractStateRequest) String() string {
	return fmt.Sprintf("QueryAllContractStateRequest{Address: %s}", m.Address)
}
func (m *QueryAllContractStateRequest) ProtoMessage() {}

// QueryAllContractStateResponse
func (m *QueryAllContractStateResponse) Reset() { *m = QueryAllContractStateResponse{} }
func (m *QueryAllContractStateResponse) String() string {
	return fmt.Sprintf("QueryAllContractStateResponse{Models: %d}", len(m.Models))
}
func (m *QueryAllContractStateResponse) ProtoMessage() {}

// QueryRawContractStateRequest
func (m *QueryRawContractStateRequest) Reset() { *m = QueryRawContractStateRequest{} }
func (m *QueryRawContractStateRequest) String() string {
	return fmt.Sprintf("QueryRawContractStateRequest{Address: %s}", m.Address)
}
func (m *QueryRawContractStateRequest) ProtoMessage() {}

// QueryRawContractStateResponse
func (m *QueryRawContractStateResponse) Reset() { *m = QueryRawContractStateResponse{} }
func (m *QueryRawContractStateResponse) String() string {
	return fmt.Sprintf("QueryRawContractStateResponse{Data: %d bytes}", len(m.Data))
}
func (m *QueryRawContractStateResponse) ProtoMessage() {}

// QuerySmartContractStateRequest
func (m *QuerySmartContractStateRequest) Reset() { *m = QuerySmartContractStateRequest{} }
func (m *QuerySmartContractStateRequest) String() string {
	return fmt.Sprintf("QuerySmartContractStateRequest{Address: %s}", m.Address)
}
func (m *QuerySmartContractStateRequest) ProtoMessage() {}

// QuerySmartContractStateResponse
func (m *QuerySmartContractStateResponse) Reset() { *m = QuerySmartContractStateResponse{} }
func (m *QuerySmartContractStateResponse) String() string {
	return fmt.Sprintf("QuerySmartContractStateResponse{Data: %d bytes}", len(m.Data))
}
func (m *QuerySmartContractStateResponse) ProtoMessage() {}

// QuerySecurityStatsRequest
func (m *QuerySecurityStatsRequest) Reset()         { *m = QuerySecurityStatsRequest{} }
func (m *QuerySecurityStatsRequest) String() string { return "QuerySecurityStatsRequest{}" }
func (m *QuerySecurityStatsRequest) ProtoMessage()  {}

// QuerySecurityStatsResponse
func (m *QuerySecurityStatsResponse) Reset() { *m = QuerySecurityStatsResponse{} }
func (m *QuerySecurityStatsResponse) String() string {
	return fmt.Sprintf("QuerySecurityStatsResponse{Stats: %+v}", m.Stats)
}
func (m *QuerySecurityStatsResponse) ProtoMessage() {}

// QueryAuthorizedUploadersRequest
func (m *QueryAuthorizedUploadersRequest) Reset() { *m = QueryAuthorizedUploadersRequest{} }
func (m *QueryAuthorizedUploadersRequest) String() string {
	return "QueryAuthorizedUploadersRequest{}"
}
func (m *QueryAuthorizedUploadersRequest) ProtoMessage() {}

// QueryAuthorizedUploadersResponse
func (m *QueryAuthorizedUploadersResponse) Reset() { *m = QueryAuthorizedUploadersResponse{} }
func (m *QueryAuthorizedUploadersResponse) String() string {
	return fmt.Sprintf("QueryAuthorizedUploadersResponse{Uploaders: %d}", len(m.Uploaders))
}
func (m *QueryAuthorizedUploadersResponse) ProtoMessage() {}

// QueryPausedContractsRequest
func (m *QueryPausedContractsRequest) Reset()         { *m = QueryPausedContractsRequest{} }
func (m *QueryPausedContractsRequest) String() string { return "QueryPausedContractsRequest{}" }
func (m *QueryPausedContractsRequest) ProtoMessage()  {}

// QueryPausedContractsResponse
func (m *QueryPausedContractsResponse) Reset() { *m = QueryPausedContractsResponse{} }
func (m *QueryPausedContractsResponse) String() string {
	return fmt.Sprintf("QueryPausedContractsResponse{Contracts: %d}", len(m.Contracts))
}
func (m *QueryPausedContractsResponse) ProtoMessage() {}

// QueryIsAuthorizedUploaderRequest
func (m *QueryIsAuthorizedUploaderRequest) Reset() { *m = QueryIsAuthorizedUploaderRequest{} }
func (m *QueryIsAuthorizedUploaderRequest) String() string {
	return fmt.Sprintf("QueryIsAuthorizedUploaderRequest{Address: %s}", m.Address)
}
func (m *QueryIsAuthorizedUploaderRequest) ProtoMessage() {}

// QueryIsAuthorizedUploaderResponse
func (m *QueryIsAuthorizedUploaderResponse) Reset() { *m = QueryIsAuthorizedUploaderResponse{} }
func (m *QueryIsAuthorizedUploaderResponse) String() string {
	return fmt.Sprintf("QueryIsAuthorizedUploaderResponse{IsAuthorized: %v}", m.IsAuthorized)
}
func (m *QueryIsAuthorizedUploaderResponse) ProtoMessage() {}

// QueryIsContractPausedRequest
func (m *QueryIsContractPausedRequest) Reset() { *m = QueryIsContractPausedRequest{} }
func (m *QueryIsContractPausedRequest) String() string {
	return fmt.Sprintf("QueryIsContractPausedRequest{Address: %s}", m.Address)
}
func (m *QueryIsContractPausedRequest) ProtoMessage() {}

// QueryIsContractPausedResponse
func (m *QueryIsContractPausedResponse) Reset() { *m = QueryIsContractPausedResponse{} }
func (m *QueryIsContractPausedResponse) String() string {
	return fmt.Sprintf("QueryIsContractPausedResponse{IsPaused: %v}", m.IsPaused)
}
func (m *QueryIsContractPausedResponse) ProtoMessage() {}
