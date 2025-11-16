package types

import (
	inclusionroutinespb "github.com/aequitas/aura/proto/aura/inclusionroutines/v1beta1"
)

// IRDefinition represents an internal IR definition type
type IRDefinition struct {
	ID               string
	Name             string
	Arena            inclusionroutinespb.Arena
	Description      string
	Score            int64
	POIReward        int64
	LocaleTags       []string
	PrivacyTier      inclusionroutinespb.PrivacyTier
	Version          string
	MetadataHash     string
	Status           inclusionroutinespb.IRStatus
	ActivationHeight int64
	SunsetHeight     int64
}

// IRPrerequisite represents prerequisite relationships
type IRPrerequisite struct {
	IRID          string
	RequiredIRIDs []string
}

// IRRateLimit represents rate limiting configuration
type IRRateLimit struct {
	IRID             string
	PerWalletPerHour int32
	PerWalletPerDay  int32
	PerBlockGlobal   int32
}

// IRGraphNode represents a node in the prerequisite graph
type IRGraphNode struct {
	IRID       string
	DependsOn  []string
	RequiredBy []string
}

// IRDefinitionFromProto converts proto IRDefinition to internal type
func IRDefinitionFromProto(pb *inclusionroutinespb.IRDefinition) IRDefinition {
	if pb == nil {
		return IRDefinition{}
	}
	return IRDefinition{
		ID:               pb.Id,
		Name:             pb.Name,
		Arena:            pb.Arena,
		Description:      pb.Description,
		Score:            pb.Score,
		POIReward:        pb.PoiReward,
		LocaleTags:       pb.LocaleTags,
		PrivacyTier:      pb.PrivacyTier,
		Version:          pb.Version,
		MetadataHash:     pb.MetadataHash,
		Status:           pb.Status,
		ActivationHeight: pb.ActivationHeight,
		SunsetHeight:     pb.SunsetHeight,
	}
}

// IRDefinitionToProto converts internal IRDefinition to proto type
func IRDefinitionToProto(ir IRDefinition) *inclusionroutinespb.IRDefinition {
	return &inclusionroutinespb.IRDefinition{
		Id:               ir.ID,
		Name:             ir.Name,
		Arena:            ir.Arena,
		Description:      ir.Description,
		Score:            ir.Score,
		PoiReward:        ir.POIReward,
		LocaleTags:       ir.LocaleTags,
		PrivacyTier:      ir.PrivacyTier,
		Version:          ir.Version,
		MetadataHash:     ir.MetadataHash,
		Status:           ir.Status,
		ActivationHeight: ir.ActivationHeight,
		SunsetHeight:     ir.SunsetHeight,
	}
}

// IRPrerequisiteFromProto converts proto IRPrerequisite to internal type
func IRPrerequisiteFromProto(pb *inclusionroutinespb.IRPrerequisite) IRPrerequisite {
	if pb == nil {
		return IRPrerequisite{}
	}
	return IRPrerequisite{
		IRID:          pb.IrId,
		RequiredIRIDs: pb.RequiredIrIds,
	}
}

// IRPrerequisiteToProto converts internal IRPrerequisite to proto type
func IRPrerequisiteToProto(prereq IRPrerequisite) *inclusionroutinespb.IRPrerequisite {
	return &inclusionroutinespb.IRPrerequisite{
		IrId:          prereq.IRID,
		RequiredIrIds: prereq.RequiredIRIDs,
	}
}

// IRRateLimitFromProto converts proto IRRateLimit to internal type
func IRRateLimitFromProto(pb *inclusionroutinespb.IRRateLimit) IRRateLimit {
	if pb == nil {
		return IRRateLimit{}
	}
	return IRRateLimit{
		IRID:             pb.IrId,
		PerWalletPerHour: pb.PerWalletPerHour,
		PerWalletPerDay:  pb.PerWalletPerDay,
		PerBlockGlobal:   pb.PerBlockGlobal,
	}
}

// IRRateLimitToProto converts internal IRRateLimit to proto type
func IRRateLimitToProto(limit IRRateLimit) *inclusionroutinespb.IRRateLimit {
	return &inclusionroutinespb.IRRateLimit{
		IrId:             limit.IRID,
		PerWalletPerHour: limit.PerWalletPerHour,
		PerWalletPerDay:  limit.PerWalletPerDay,
		PerBlockGlobal:   limit.PerBlockGlobal,
	}
}

// IRGraphNodeToProto converts internal IRGraphNode to proto type
func IRGraphNodeToProto(node IRGraphNode) *inclusionroutinespb.IRGraphNode {
	return &inclusionroutinespb.IRGraphNode{
		IrId:       node.IRID,
		DependsOn:  node.DependsOn,
		RequiredBy: node.RequiredBy,
	}
}

// IRGraphNodesSliceToProto converts a slice of internal IRGraphNode to proto slice
func IRGraphNodesSliceToProto(nodes []IRGraphNode) []*inclusionroutinespb.IRGraphNode {
	result := make([]*inclusionroutinespb.IRGraphNode, 0, len(nodes))
	for _, node := range nodes {
		result = append(result, IRGraphNodeToProto(node))
	}
	return result
}

// IRDefinitionsSliceToProto converts a slice of internal IRDefinition to proto slice
func IRDefinitionsSliceToProto(irs []IRDefinition) []*inclusionroutinespb.IRDefinition {
	result := make([]*inclusionroutinespb.IRDefinition, 0, len(irs))
	for _, ir := range irs {
		result = append(result, IRDefinitionToProto(ir))
	}
	return result
}
