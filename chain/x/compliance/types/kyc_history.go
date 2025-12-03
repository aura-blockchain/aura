package types

import (
	pb "github.com/aequitas/aura/proto/aura/compliance/v1beta1"
)

// KYCHistoryEntry represents a historical snapshot of a KYC record
// Used for audit trail and version tracking when records are updated
// This is a type alias to the protobuf-generated type
type KYCHistoryEntry = pb.KYCHistoryEntry

// KYCHistoryList stores history entries keyed by address
// This is a type alias to the protobuf-generated type
type KYCHistoryList = pb.KYCHistoryList

// QueryKYCHistoryRequest is the request type for the Query/KycHistory RPC method
type QueryKYCHistoryRequest struct {
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
}

// QueryKYCHistoryResponse is the response type for the Query/KycHistory RPC method
type QueryKYCHistoryResponse struct {
	History []*KYCHistoryEntry `protobuf:"bytes,1,rep,name=history,proto3" json:"history,omitempty"`
}

// Reset implements proto.Message
func (m *QueryKYCHistoryRequest) Reset() { *m = QueryKYCHistoryRequest{} }

// String implements proto.Message
func (m *QueryKYCHistoryRequest) String() string { return "QueryKYCHistoryRequest" }

// ProtoMessage implements proto.Message
func (*QueryKYCHistoryRequest) ProtoMessage() {}

// Reset implements proto.Message
func (m *QueryKYCHistoryResponse) Reset() { *m = QueryKYCHistoryResponse{} }

// String implements proto.Message
func (m *QueryKYCHistoryResponse) String() string { return "QueryKYCHistoryResponse" }

// ProtoMessage implements proto.Message
func (*QueryKYCHistoryResponse) ProtoMessage() {}
