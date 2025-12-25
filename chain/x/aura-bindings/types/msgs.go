// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"context"
)

// MsgServer is the server API for Msg service
type MsgServer interface {
	// EmptyMethod is a placeholder
	EmptyMethod(context.Context, *EmptyRequest) (*EmptyResponse, error)
}

// EmptyRequest is a placeholder request type
type EmptyRequest struct{}

// EmptyResponse is a placeholder response type
type EmptyResponse struct{}

// RegisterMsgServer registers the msg server
func RegisterMsgServer(s interface{}, impl MsgServer) {
	// Stub - would be implemented by gRPC server registration
}
