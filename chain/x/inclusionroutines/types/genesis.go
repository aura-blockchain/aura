package types

import (
	"encoding/json"
	"fmt"
	"os"

	inclusionroutinespb "github.com/aequitas/aura/proto/aura/inclusionroutines/v1beta1"
	"github.com/cosmos/cosmos-sdk/codec"
)

// IRDefinitionFromJSON represents an IR definition from the JSON file
type IRDefinitionFromJSON struct {
	ID             string   `json:"id"`
	Arena          string   `json:"arena"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Score          int64    `json:"score"`
	POIReward      int64    `json:"poi_reward,omitempty"`
	IsPrerequisite bool     `json:"is_prerequisite,omitempty"`
	PrivacyTier    string   `json:"privacy_tier"`
	LocaleTags     []string `json:"locale_tags"`
}

// IRDataFile represents the structure of ir_definitions.json
type IRDataFile struct {
	Version           string                 `json:"version"`
	TotalCount        int                    `json:"total_count"`
	LastUpdated       string                 `json:"last_updated"`
	SourceDocument    string                 `json:"source_document"`
	Description       string                 `json:"description"`
	InclusionRoutines []IRDefinitionFromJSON `json:"inclusion_routines"`
}

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *inclusionroutinespb.GenesisState {
	return &inclusionroutinespb.GenesisState{
		Params:        DefaultParamsProto(),
		Irs:           []*inclusionroutinespb.IRDefinition{},
		Prerequisites: []*inclusionroutinespb.IRPrerequisite{},
		RateLimits:    []*inclusionroutinespb.IRRateLimit{},
	}
}

// LoadGenesisFromFile loads the genesis state from the IR definitions JSON file
func LoadGenesisFromFile(filePath string) (*inclusionroutinespb.GenesisState, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read ir definitions file: %w", err)
	}

	var irData IRDataFile
	if err := json.Unmarshal(data, &irData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal ir definitions: %w", err)
	}

	genesis := DefaultGenesisState()

	// Convert JSON IR definitions to proto format
	for _, ir := range irData.InclusionRoutines {
		protoIR := &inclusionroutinespb.IRDefinition{
			Id:               ir.ID,
			Name:             ir.Name,
			Arena:            arenaFromString(ir.Arena),
			Description:      ir.Description,
			Score:            ir.Score,
			PoiReward:        ir.POIReward,
			LocaleTags:       ir.LocaleTags,
			PrivacyTier:      privacyTierFromString(ir.PrivacyTier),
			Version:          "1.0",
			MetadataHash:     "",
			Status:           inclusionroutinespb.IRStatus_IR_STATUS_ACTIVE,
			ActivationHeight: 0,
			SunsetHeight:     0,
		}
		genesis.Irs = append(genesis.Irs, protoIR)
	}

	return genesis, nil
}

// ValidateGenesisState validates the genesis state
func ValidateGenesisState(state *inclusionroutinespb.GenesisState) error {
	if state == nil {
		return fmt.Errorf("genesis state cannot be nil")
	}

	// Validate params
	if state.Params != nil {
		params := ParamsFromProto(state.Params)
		if err := params.Validate(); err != nil {
			return fmt.Errorf("params validation failed: %w", err)
		}
	}

	// Validate IR definitions
	irIDs := make(map[string]struct{}, len(state.Irs))
	for _, ir := range state.Irs {
		if ir.Id == "" {
			return fmt.Errorf("ir definition missing id")
		}
		if ir.Name == "" {
			return fmt.Errorf("ir %s missing name", ir.Id)
		}
		if ir.Description == "" {
			return fmt.Errorf("ir %s missing description", ir.Id)
		}
		if ir.Score < 0 {
			return fmt.Errorf("ir %s has negative score: %d", ir.Id, ir.Score)
		}
		if ir.PoiReward < 0 {
			return fmt.Errorf("ir %s has negative poi reward: %d", ir.Id, ir.PoiReward)
		}
		if len(ir.LocaleTags) == 0 {
			return fmt.Errorf("ir %s missing locale tags", ir.Id)
		}
		if ir.SunsetHeight > 0 && ir.ActivationHeight > 0 && ir.SunsetHeight <= ir.ActivationHeight {
			return fmt.Errorf("ir %s sunset height must be after activation height", ir.Id)
		}

		if _, exists := irIDs[ir.Id]; exists {
			return fmt.Errorf("duplicate ir id: %s", ir.Id)
		}
		irIDs[ir.Id] = struct{}{}
	}

	// Validate prerequisites
	for _, prereq := range state.Prerequisites {
		if prereq.IrId == "" {
			return fmt.Errorf("prerequisite missing ir_id")
		}
		if _, exists := irIDs[prereq.IrId]; !exists {
			return fmt.Errorf("prerequisite references non-existent ir: %s", prereq.IrId)
		}
		for _, reqID := range prereq.RequiredIrIds {
			if reqID == prereq.IrId {
				return fmt.Errorf("ir %s cannot be its own prerequisite", prereq.IrId)
			}
			if _, exists := irIDs[reqID]; !exists {
				return fmt.Errorf("prerequisite %s references non-existent required ir: %s", prereq.IrId, reqID)
			}
		}
	}

	// Validate rate limits
	for _, rateLimit := range state.RateLimits {
		if rateLimit.IrId == "" {
			return fmt.Errorf("rate limit missing ir_id")
		}
		if _, exists := irIDs[rateLimit.IrId]; !exists {
			return fmt.Errorf("rate limit references non-existent ir: %s", rateLimit.IrId)
		}
		if rateLimit.PerWalletPerHour < 0 {
			return fmt.Errorf("rate limit for %s has negative per_wallet_per_hour", rateLimit.IrId)
		}
		if rateLimit.PerWalletPerDay < 0 {
			return fmt.Errorf("rate limit for %s has negative per_wallet_per_day", rateLimit.IrId)
		}
		if rateLimit.PerBlockGlobal < 0 {
			return fmt.Errorf("rate limit for %s has negative per_block_global", rateLimit.IrId)
		}
	}

	// Check for circular dependencies in prerequisites
	if err := validateNoCycles(state.Prerequisites); err != nil {
		return fmt.Errorf("prerequisite validation failed: %w", err)
	}

	return nil
}

// DefaultGenesis returns the default genesis as raw JSON
func DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(DefaultGenesisState())
}

// arenaFromString converts a string arena to proto enum
func arenaFromString(s string) inclusionroutinespb.Arena {
	switch s {
	case "ANCHOR":
		return inclusionroutinespb.Arena_ARENA_ANCHOR
	case "Biometric":
		return inclusionroutinespb.Arena_ARENA_BIOMETRIC
	case "Possession":
		return inclusionroutinespb.Arena_ARENA_POSSESSION
	case "Knowledge":
		return inclusionroutinespb.Arena_ARENA_KNOWLEDGE
	case "Social":
		return inclusionroutinespb.Arena_ARENA_SOCIAL
	case "Geolocation":
		return inclusionroutinespb.Arena_ARENA_GEOLOCATION
	case "High-Assurance":
		return inclusionroutinespb.Arena_ARENA_HIGH_ASSURANCE
	case "Persistence":
		return inclusionroutinespb.Arena_ARENA_PERSISTENCE
	case "Specialized":
		return inclusionroutinespb.Arena_ARENA_SPECIALIZED
	default:
		return inclusionroutinespb.Arena_ARENA_UNSPECIFIED
	}
}

// privacyTierFromString converts a string privacy tier to proto enum
func privacyTierFromString(s string) inclusionroutinespb.PrivacyTier {
	switch s {
	case "low":
		return inclusionroutinespb.PrivacyTier_PRIVACY_TIER_LOW
	case "medium":
		return inclusionroutinespb.PrivacyTier_PRIVACY_TIER_MEDIUM
	case "high":
		return inclusionroutinespb.PrivacyTier_PRIVACY_TIER_HIGH
	default:
		return inclusionroutinespb.PrivacyTier_PRIVACY_TIER_UNSPECIFIED
	}
}

// validateNoCycles checks for circular dependencies in the prerequisite graph
func validateNoCycles(prereqs []*inclusionroutinespb.IRPrerequisite) error {
	graph := make(map[string][]string)
	for _, p := range prereqs {
		graph[p.IrId] = p.RequiredIrIds
	}

	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var hasCycle func(string) bool
	hasCycle = func(node string) bool {
		visited[node] = true
		recStack[node] = true

		for _, neighbor := range graph[node] {
			if !visited[neighbor] {
				if hasCycle(neighbor) {
					return true
				}
			} else if recStack[neighbor] {
				return true
			}
		}

		recStack[node] = false
		return false
	}

	for node := range graph {
		if !visited[node] {
			if hasCycle(node) {
				return ErrCircularDependency
			}
		}
	}

	return nil
}
