// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"strings"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aequitas/aura/chain/x/vcregistry/types"
)

// FuzzMapVCAuthorizationError ensures ErrUnauthorized surfaces as PermissionDenied regardless of message shape.
func FuzzMapVCAuthorizationError(f *testing.F) {
	f.Add("signer mismatch")
	f.Add("invalid authority")

	f.Fuzz(func(t *testing.T, msg string) {
		err := mapVCAuthorizationError(types.ErrUnauthorized.Wrap(msg))
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("expected status error, got %v", err)
		}
		if st.Code() != codes.PermissionDenied {
			t.Fatalf("expected PermissionDenied, got %s (message=%q)", st.Code(), strings.ToLower(msg))
		}
	})
}
