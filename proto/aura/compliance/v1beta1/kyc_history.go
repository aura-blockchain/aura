package v1beta1

import (
	"google.golang.org/protobuf/types/known/timestamppb"
)

// KYCHistoryEntry represents a historical snapshot of a KYC record
// Used for audit trail and version tracking when records are updated
// This is a temporary type definition until protobuf regeneration
type KYCHistoryEntry struct {
	Address      string                 `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	Version      uint64                 `protobuf:"varint,2,opt,name=version,proto3" json:"version,omitempty"`
	Snapshot     *KYCRecord             `protobuf:"bytes,3,opt,name=snapshot,proto3" json:"snapshot,omitempty"`
	UpdatedAt    *timestamppb.Timestamp `protobuf:"bytes,4,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
	UpdatedBy    string                 `protobuf:"bytes,5,opt,name=updated_by,json=updatedBy,proto3" json:"updated_by,omitempty"`
	UpdateReason string                 `protobuf:"bytes,6,opt,name=update_reason,json=updateReason,proto3" json:"update_reason,omitempty"`
}

// KYCHistoryList stores history entries keyed by address
type KYCHistoryList struct {
	Entries []*KYCHistoryEntry `protobuf:"bytes,1,rep,name=entries,proto3" json:"entries,omitempty"`
}

// Reset implements proto.Message
func (m *KYCHistoryEntry) Reset() { *m = KYCHistoryEntry{} }

// String implements proto.Message
func (m *KYCHistoryEntry) String() string { return "KYCHistoryEntry" }

// ProtoMessage implements proto.Message
func (*KYCHistoryEntry) ProtoMessage() {}

// Reset implements proto.Message
func (m *KYCHistoryList) Reset() { *m = KYCHistoryList{} }

// String implements proto.Message
func (m *KYCHistoryList) String() string { return "KYCHistoryList" }

// ProtoMessage implements proto.Message
func (*KYCHistoryList) ProtoMessage() {}
