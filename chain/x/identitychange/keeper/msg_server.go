package keeper

import (
	identitychangepb "github.com/aequitas/aura/proto/aura/identitychange/v1beta1"
)

// MsgServer implements the identitychange Msg service
type MsgServer struct {
	identitychangepb.UnimplementedMsgServer
	keeper *Keeper
}

// NewMsgServer returns a new MsgServer
func NewMsgServer(k *Keeper) identitychangepb.MsgServer {
	return &MsgServer{keeper: k}
}

var _ identitychangepb.MsgServer = &MsgServer{}
