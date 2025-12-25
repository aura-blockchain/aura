// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	confidencescorepb "github.com/aequitas/aura/proto/aura/confidencescore/v1beta1"
)

// MsgServer implements the confidencescore Msg service
type MsgServer struct {
	confidencescorepb.UnimplementedMsgServer
	keeper *Keeper
}

// NewMsgServer returns a new MsgServer
func NewMsgServer(k *Keeper) confidencescorepb.MsgServer {
	return &MsgServer{keeper: k}
}

var _ confidencescorepb.MsgServer = &MsgServer{}
