// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	identitypb "github.com/aequitas/aura/proto/aura/identity/v1beta1"
	"google.golang.org/grpc"
)

// RegisterMsgServer registers the Msg gRPC service
func RegisterMsgServer(s grpc.ServiceRegistrar, srv identitypb.MsgServer) {
	identitypb.RegisterMsgServer(s, srv)
}

// RegisterQueryServer registers the Query gRPC service
func RegisterQueryServer(s grpc.ServiceRegistrar, srv identitypb.QueryServer) {
	identitypb.RegisterQueryServer(s, srv)
}
