// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	inclusionroutinespb "github.com/aequitas/aura/proto/aura/inclusionroutines/v1beta1"
)

// IRDefinitionToProto converts a types.IRDefinition to proto IRDefinition
// Since the types are structurally identical (just different packages),
// we create a new proto instance with the same field values
func IRDefinitionToProto(ir IRDefinition) *inclusionroutinespb.IRDefinition {
	return &inclusionroutinespb.IRDefinition{
		Id:               ir.Id,
		Name:             ir.Name,
		Arena:            inclusionroutinespb.Arena(ir.Arena),
		Description:      ir.Description,
		Score:            ir.Score,
		PoiReward:        ir.PoiReward,
		LocaleTags:       ir.LocaleTags,
		PrivacyTier:      inclusionroutinespb.PrivacyTier(ir.PrivacyTier),
		Version:          ir.Version,
		MetadataHash:     ir.MetadataHash,
		Status:           inclusionroutinespb.IRStatus(ir.Status),
		ActivationHeight: ir.ActivationHeight,
		SunsetHeight:     ir.SunsetHeight,
	}
}

// IRDefinitionsSliceToProto converts a slice of types.IRDefinition to proto IRDefinition pointers
func IRDefinitionsSliceToProto(irs []IRDefinition) []*inclusionroutinespb.IRDefinition {
	result := make([]*inclusionroutinespb.IRDefinition, len(irs))
	for i, ir := range irs {
		result[i] = IRDefinitionToProto(ir)
	}
	return result
}

// IRRateLimitToProto converts a types.IRRateLimit to proto IRRateLimit
func IRRateLimitToProto(limit IRRateLimit) *inclusionroutinespb.IRRateLimit {
	return &inclusionroutinespb.IRRateLimit{
		IrId:             limit.IrId,
		PerWalletPerHour: limit.PerWalletPerHour,
		PerWalletPerDay:  limit.PerWalletPerDay,
		PerBlockGlobal:   limit.PerBlockGlobal,
	}
}

// IRPrerequisiteToProto converts a types.IRPrerequisite to proto IRPrerequisite
func IRPrerequisiteToProto(prereq IRPrerequisite) *inclusionroutinespb.IRPrerequisite {
	return &inclusionroutinespb.IRPrerequisite{
		IrId:          prereq.IrId,
		RequiredIrIds: prereq.RequiredIrIds,
	}
}

// IRGraphNodeToProto converts a types.IRGraphNode to proto IRGraphNode
func IRGraphNodeToProto(node IRGraphNode) *inclusionroutinespb.IRGraphNode {
	return &inclusionroutinespb.IRGraphNode{
		IrId:       node.IrId,
		DependsOn:  node.DependsOn,
		RequiredBy: node.RequiredBy,
	}
}

// IRGraphNodesSliceToProto converts a slice of types.IRGraphNode to proto IRGraphNode pointers
func IRGraphNodesSliceToProto(nodes []IRGraphNode) []*inclusionroutinespb.IRGraphNode {
	result := make([]*inclusionroutinespb.IRGraphNode, len(nodes))
	for i, node := range nodes {
		result[i] = IRGraphNodeToProto(node)
	}
	return result
}

// ParamsToProto converts types.Params to proto Params
func ParamsToProto(params Params) *inclusionroutinespb.Params {
	return &inclusionroutinespb.Params{
		MaxIrPerLocale:       params.MaxIrPerLocale,
		DefaultRateLimitHour: params.DefaultRateLimitHour,
		SuspensionFee:        params.SuspensionFee,
		MinGovernanceDeposit: params.MinGovernanceDeposit,
	}
}
