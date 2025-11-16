package dataregistry

import (
	"context"
	"time"

	"github.com/aequitas/aura/chain/x/dataregistry/keeper"
	"github.com/aequitas/aura/chain/x/dataregistry/types"
	dataregistrypb "github.com/aequitas/aura/proto/aura/dataregistry/v1beta1"
)

// MsgServer defines the message server interface
type MsgServer interface {
	StoreDataItem(ctx context.Context, msg *MsgStoreDataItem) (*MsgStoreDataItemResponse, error)
	UpdateDataItem(ctx context.Context, msg *MsgUpdateDataItem) (*MsgUpdateDataItemResponse, error)
	DeleteDataItem(ctx context.Context, msg *MsgDeleteDataItem) (*MsgDeleteDataItemResponse, error)
	VerifyDataItem(ctx context.Context, msg *MsgVerifyDataItem) (*MsgVerifyDataItemResponse, error)
	RevokeDataItem(ctx context.Context, msg *MsgRevokeDataItem) (*MsgRevokeDataItemResponse, error)
}

// msgServer implements MsgServer
type msgServer struct {
	dataregistrypb.UnimplementedMsgServer
	keeper *keeper.Keeper
}

// NewMsgServer creates a new MsgServer
func NewMsgServer(k *keeper.Keeper) MsgServer {
	return &msgServer{keeper: k}
}

// Message types
type MsgStoreDataItem struct {
	Creator         string
	DataType        types.DataItemType
	Title           string
	Description     string
	ContentHash     []byte
	StorageLocation string
	IsEncrypted     bool
	GeoLocation     *types.GeoLocation
	Metadata        map[string]string
	AccessPolicy    *types.AccessPolicy
	Tags            []string
}

type MsgStoreDataItemResponse struct {
	DataID    string
	CreatedAt time.Time
}

type MsgUpdateDataItem struct {
	Creator      string
	DataID       string
	Title        string
	Description  string
	Metadata     map[string]string
	AccessPolicy *types.AccessPolicy
	Tags         []string
}

type MsgUpdateDataItemResponse struct {
	UpdatedAt time.Time
}

type MsgDeleteDataItem struct {
	Creator string
	DataID  string
}

type MsgDeleteDataItemResponse struct {
	DeletedAt time.Time
}

type MsgVerifyDataItem struct {
	Verifier           string
	DataID             string
	Level              types.VerificationLevel
	ConfidenceScore    uint64
	Notes              string
	VerificationMethod string
	Proof              []byte
}

type MsgVerifyDataItemResponse struct {
	VerifiedAt         time.Time
	VerificationReward uint64
}

type MsgRevokeDataItem struct {
	Authority string
	DataID    string
	Reason    string
}

type MsgRevokeDataItemResponse struct {
	RevokedAt time.Time
}

// StoreDataItem stores a new data item
func (s *msgServer) StoreDataItem(ctx context.Context, msg *MsgStoreDataItem) (*MsgStoreDataItemResponse, error) {
	dataID, err := s.keeper.StoreDataItem(
		msg.Creator,
		msg.DataType,
		msg.Title,
		msg.Description,
		msg.ContentHash,
		msg.StorageLocation,
		msg.IsEncrypted,
		msg.GeoLocation,
		msg.Metadata,
		msg.AccessPolicy,
		msg.Tags,
	)
	if err != nil {
		return nil, err
	}

	return &MsgStoreDataItemResponse{
		DataID:    dataID,
		CreatedAt: time.Now(),
	}, nil
}

// UpdateDataItem updates an existing data item
func (s *msgServer) UpdateDataItem(ctx context.Context, msg *MsgUpdateDataItem) (*MsgUpdateDataItemResponse, error) {
	err := s.keeper.UpdateDataItem(
		msg.DataID,
		msg.Creator,
		msg.Title,
		msg.Description,
		msg.Metadata,
		msg.AccessPolicy,
		msg.Tags,
	)
	if err != nil {
		return nil, err
	}

	return &MsgUpdateDataItemResponse{
		UpdatedAt: time.Now(),
	}, nil
}

// DeleteDataItem deletes a data item
func (s *msgServer) DeleteDataItem(ctx context.Context, msg *MsgDeleteDataItem) (*MsgDeleteDataItemResponse, error) {
	// Get item to verify ownership
	item, ok := s.keeper.GetDataItem(msg.DataID)
	if !ok {
		return nil, types.ErrDataItemNotFound
	}

	if item.OwnerAddress != msg.Creator {
		return nil, types.ErrUnauthorized
	}

	err := s.keeper.DeleteDataItem(msg.DataID)
	if err != nil {
		return nil, err
	}

	return &MsgDeleteDataItemResponse{
		DeletedAt: time.Now(),
	}, nil
}

// VerifyDataItem adds verification to a data item
func (s *msgServer) VerifyDataItem(ctx context.Context, msg *MsgVerifyDataItem) (*MsgVerifyDataItemResponse, error) {
	err := s.keeper.VerifyDataItem(
		msg.DataID,
		msg.Verifier,
		msg.Level,
		msg.ConfidenceScore,
		msg.Notes,
		msg.VerificationMethod,
		msg.Proof,
	)
	if err != nil {
		return nil, err
	}

	params := s.keeper.GetParams()

	return &MsgVerifyDataItemResponse{
		VerifiedAt:         time.Now(),
		VerificationReward: params.VerificationReward,
	}, nil
}

// RevokeDataItem revokes a data item
func (s *msgServer) RevokeDataItem(ctx context.Context, msg *MsgRevokeDataItem) (*MsgRevokeDataItemResponse, error) {
	err := s.keeper.RevokeDataItem(
		msg.DataID,
		msg.Authority,
		msg.Reason,
	)
	if err != nil {
		return nil, err
	}

	return &MsgRevokeDataItemResponse{
		RevokedAt: time.Now(),
	}, nil
}
