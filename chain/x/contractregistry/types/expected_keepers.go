// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	pb "github.com/aequitas/aura/proto/aura/contractregistry/v1beta1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ComplianceKeeper defines the expected interface for the compliance keeper
type ComplianceKeeper interface {
	GetKYCLevel(ctx sdk.Context, address string) (uint32, error)
	ScreenForSanctions(ctx sdk.Context, address string) (bool, error)
}

// VCKeeper defines the expected interface for the VC keeper
type VCKeeper interface {
	HasVC(ctx interface{}, address string, vcType string) bool
}

// ConfidenceScoreKeeper defines the expected interface for the confidence score keeper
type ConfidenceScoreKeeper interface {
	GetUserScore(ctx sdk.Context, address string) (uint64, bool)
}

// Type aliases for proto-generated interfaces
type (
	// MsgServer is the server API for Msg service
	MsgServer = pb.MsgServer

	// QueryServer is the server API for Query service
	QueryServer = pb.QueryServer
)
