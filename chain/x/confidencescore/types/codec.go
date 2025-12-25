// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"

	cspb "github.com/aequitas/aura/proto/aura/confidencescore/v1beta1"
)

// RegisterInterfaces registers confidencescore module interfaces. Currently relies on proto-generated registrations.
func RegisterInterfaces(reg codectypes.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(reg, &cspb.Msg_serviceDesc)

	reg.RegisterImplementations(
		(*sdk.Msg)(nil),
		&cspb.MsgRecordIRCompletion{},
		&cspb.MsgRecalculateScore{},
		&cspb.MsgSlashScore{},
		&cspb.MsgAppealSlash{},
		&cspb.MsgResolveAppeal{},
	)
}
