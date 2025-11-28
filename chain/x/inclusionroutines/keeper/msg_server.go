package keeper

import (
	inclusionroutinespb "github.com/aequitas/aura/proto/aura/inclusionroutines/v1beta1"
)

// MsgServer implements the inclusionroutines Msg service
type MsgServer struct {
	inclusionroutinespb.UnimplementedMsgServer
	keeper *Keeper
}

// NewMsgServer returns a new MsgServer
func NewMsgServer(k *Keeper) inclusionroutinespb.MsgServer {
	return &MsgServer{keeper: k}
}

var _ inclusionroutinespb.MsgServer = &MsgServer{}
