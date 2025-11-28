package keeper

import (
	economicsecuritypb "github.com/aequitas/aura/proto/aura/economicsecurity/v1beta1"
)

// MsgServer implements the economicsecurity Msg service
type MsgServer struct {
	economicsecuritypb.UnimplementedMsgServer
	keeper *Keeper
}

// NewMsgServer returns a new MsgServer
func NewMsgServer(k *Keeper) economicsecuritypb.MsgServer {
	return &MsgServer{keeper: k}
}

var _ economicsecuritypb.MsgServer = &MsgServer{}
