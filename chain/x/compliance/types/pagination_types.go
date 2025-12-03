package types

import (
	"github.com/cosmos/cosmos-sdk/types/query"
)

// Paginated query request/response types
// These are placeholders until protobuf generation is working
// In production, these would be generated from proto/aura/compliance/v1beta1/compliance.proto

// QueryAllKYCRecordsRequest is the request type for the Query/AllKYCRecords RPC method
type QueryAllKYCRecordsRequest struct {
	Pagination *query.PageRequest `protobuf:"bytes,1,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

// QueryAllKYCRecordsResponse is the response type for the Query/AllKYCRecords RPC method
type QueryAllKYCRecordsResponse struct {
	Records    []*KYCRecord        `protobuf:"bytes,1,rep,name=records,proto3" json:"records,omitempty"`
	Pagination *query.PageResponse `protobuf:"bytes,2,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

// QueryAllAMLProfilesRequest is the request type for the Query/AllAMLProfiles RPC method
type QueryAllAMLProfilesRequest struct {
	Pagination *query.PageRequest `protobuf:"bytes,1,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

// QueryAllAMLProfilesResponse is the response type for the Query/AllAMLProfiles RPC method
type QueryAllAMLProfilesResponse struct {
	Profiles   []*AMLProfile       `protobuf:"bytes,1,rep,name=profiles,proto3" json:"profiles,omitempty"`
	Pagination *query.PageResponse `protobuf:"bytes,2,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

// QueryAllSanctionsResultsRequest is the request type for the Query/AllSanctionsResults RPC method
type QueryAllSanctionsResultsRequest struct {
	Pagination *query.PageRequest `protobuf:"bytes,1,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

// QueryAllSanctionsResultsResponse is the response type for the Query/AllSanctionsResults RPC method
type QueryAllSanctionsResultsResponse struct {
	Results    []*SanctionsScreeningResult `protobuf:"bytes,1,rep,name=results,proto3" json:"results,omitempty"`
	Pagination *query.PageResponse         `protobuf:"bytes,2,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

// QueryAllTransactionAlertsRequest is the request type for the Query/AllTransactionAlerts RPC method
type QueryAllTransactionAlertsRequest struct {
	Pagination *query.PageRequest `protobuf:"bytes,1,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

// QueryAllTransactionAlertsResponse is the response type for the Query/AllTransactionAlerts RPC method
type QueryAllTransactionAlertsResponse struct {
	Alerts     []*TransactionAlertList `protobuf:"bytes,1,rep,name=alerts,proto3" json:"alerts,omitempty"`
	Pagination *query.PageResponse     `protobuf:"bytes,2,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

// QueryAllGDPRConsentsRequest is the request type for the Query/AllGDPRConsents RPC method
type QueryAllGDPRConsentsRequest struct {
	Pagination *query.PageRequest `protobuf:"bytes,1,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

// QueryAllGDPRConsentsResponse is the response type for the Query/AllGDPRConsents RPC method
type QueryAllGDPRConsentsResponse struct {
	Consents   []*GDPRConsentList  `protobuf:"bytes,1,rep,name=consents,proto3" json:"consents,omitempty"`
	Pagination *query.PageResponse `protobuf:"bytes,2,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

// QueryAllTaxReportsRequest is the request type for the Query/AllTaxReports RPC method
type QueryAllTaxReportsRequest struct {
	Pagination *query.PageRequest `protobuf:"bytes,1,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

// QueryAllTaxReportsResponse is the response type for the Query/AllTaxReports RPC method
type QueryAllTaxReportsResponse struct {
	Reports    []*TaxReportList    `protobuf:"bytes,1,rep,name=reports,proto3" json:"reports,omitempty"`
	Pagination *query.PageResponse `protobuf:"bytes,2,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

// QueryAllGDPRRequestsRequest is the request type for the Query/AllGDPRRequests RPC method
type QueryAllGDPRRequestsRequest struct {
	Pagination *query.PageRequest `protobuf:"bytes,1,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

// QueryAllGDPRRequestsResponse is the response type for the Query/AllGDPRRequests RPC method
type QueryAllGDPRRequestsResponse struct {
	Requests   []*GDPRDataRequest  `protobuf:"bytes,1,rep,name=requests,proto3" json:"requests,omitempty"`
	Pagination *query.PageResponse `protobuf:"bytes,2,opt,name=pagination,proto3" json:"pagination,omitempty"`
}
