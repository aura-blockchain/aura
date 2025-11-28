package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/aequitas/aura/chain/x/inclusionroutines/types"
)

// GetPrerequisites retrieves the prerequisites for an IR from KV store
func (k *Keeper) GetPrerequisites(ctx sdk.Context, irID string) (types.IRPrerequisite, bool) {
	return k.GetPrerequisite(ctx, irID)
}

// SetPrerequisites sets the prerequisites for an IR in KV store
func (k *Keeper) SetPrerequisites(ctx sdk.Context, irID string, requiredIRIDs []string) error {
	// Check if IR exists
	if _, exists := k.GetIR(ctx, irID); !exists {
		return types.ErrIRNotFound
	}

	// Validate all required IRs exist
	for _, reqID := range requiredIRIDs {
		if reqID == irID {
			return types.ErrSelfPrerequisite
		}
		if _, exists := k.GetIR(ctx, reqID); !exists {
			return fmt.Errorf("%w: %s", types.ErrPrerequisiteNotFound, reqID)
		}
	}

	// Create new prerequisite relationship
	newPrereq := types.IRPrerequisite{
		IrId:          irID,
		RequiredIrIds: requiredIRIDs,
	}

	// Check for circular dependencies
	if err := k.detectCircularDependency(ctx, irID, requiredIRIDs); err != nil {
		return err
	}

	return k.SetPrerequisite(ctx, newPrereq)
}

// ValidatePrerequisites checks if a wallet has completed all prerequisites for an IR
func (k *Keeper) ValidatePrerequisites(ctx sdk.Context, irID string, completedIRs []string) error {
	// Check if IR exists
	if _, exists := k.GetIR(ctx, irID); !exists {
		return types.ErrIRNotFound
	}

	// Get prerequisites for this IR
	prereq, hasPrereqs := k.GetPrerequisite(ctx, irID)
	if !hasPrereqs || len(prereq.RequiredIrIds) == 0 {
		return nil // No prerequisites required
	}

	// Build set of completed IRs for fast lookup
	completed := make(map[string]bool)
	for _, id := range completedIRs {
		completed[id] = true
	}

	// Check all prerequisites are met
	missing := make([]string, 0)
	for _, requiredIRID := range prereq.RequiredIrIds {
		if !completed[requiredIRID] {
			missing = append(missing, requiredIRID)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("%w: missing %v", types.ErrPrerequisiteNotMet, missing)
	}

	return nil
}

// GetIRGraph returns the prerequisite dependency graph for an IR
// If irID is empty, returns the full graph
func (k *Keeper) GetIRGraph(ctx sdk.Context, irID string) []types.IRGraphNode {
	// Build reverse lookup (which IRs require each IR)
	requiredBy := make(map[string][]string)
	allIRs := k.GetAllIRs(ctx)

	for _, ir := range allIRs {
		if prereq, exists := k.GetPrerequisite(ctx, ir.Id); exists {
			for _, reqID := range prereq.RequiredIrIds {
				requiredBy[reqID] = append(requiredBy[reqID], ir.Id)
			}
		}
	}

	nodes := make([]types.IRGraphNode, 0)

	if irID != "" {
		// Return subgraph for specific IR
		node := k.buildGraphNode(ctx, irID, requiredBy)
		nodes = append(nodes, node)

		// Add all dependencies recursively
		visited := make(map[string]bool)
		k.addDependencies(ctx, irID, &nodes, requiredBy, visited)
	} else {
		// Return full graph
		for _, ir := range allIRs {
			node := k.buildGraphNode(ctx, ir.Id, requiredBy)
			nodes = append(nodes, node)
		}
	}

	return nodes
}

// GetPrerequisiteChain returns full prerequisite chain for an IR
func (k *Keeper) GetPrerequisiteChain(ctx sdk.Context, irID string) ([]types.IRGraphNode, error) {
	chain := []types.IRGraphNode{}
	visited := make(map[string]bool)
	k.buildPrerequisiteChain(ctx, irID, 0, &chain, visited)
	return chain, nil
}

// ValidatePrerequisiteChain validates entire prerequisite dependency chain
func (k *Keeper) ValidatePrerequisiteChain(ctx sdk.Context, irID string, completedIRs map[string]bool) (bool, []string, error) {
	prereq, exists := k.GetPrerequisite(ctx, irID)
	if !exists {
		// No prerequisites
		return true, []string{}, nil
	}

	missingPrereqs := []string{}

	// Check direct prerequisites
	for _, requiredIRID := range prereq.RequiredIrIds {
		if !completedIRs[requiredIRID] {
			missingPrereqs = append(missingPrereqs, requiredIRID)

			// Recursively check sub-prerequisites
			subValid, subMissing, _ := k.ValidatePrerequisiteChain(ctx, requiredIRID, completedIRs)
			if !subValid {
				missingPrereqs = append(missingPrereqs, subMissing...)
			}
		}
	}

	valid := len(missingPrereqs) == 0
	return valid, missingPrereqs, nil
}

// detectCircularDependency checks for cycles in the prerequisite graph
func (k *Keeper) detectCircularDependency(ctx sdk.Context, irID string, newRequirements []string) error {
	// Build temporary graph with new prerequisite
	graph := make(map[string][]string)

	// Load existing prerequisites
	allIRs := k.GetAllIRs(ctx)
	for _, ir := range allIRs {
		if prereq, exists := k.GetPrerequisite(ctx, ir.Id); exists {
			graph[ir.Id] = prereq.RequiredIrIds
		}
	}

	// Add new requirement
	graph[irID] = newRequirements

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

	if hasCycle(irID) {
		return types.ErrCircularDependency
	}

	return nil
}

// buildGraphNode creates a graph node for an IR
func (k *Keeper) buildGraphNode(ctx sdk.Context, irID string, requiredBy map[string][]string) types.IRGraphNode {
	node := types.IRGraphNode{
		IrId:       irID,
		DependsOn:  []string{},
		RequiredBy: []string{},
	}

	// Add prerequisites (what this IR depends on)
	if prereq, exists := k.GetPrerequisite(ctx, irID); exists {
		node.DependsOn = prereq.RequiredIrIds
	}

	// Add what requires this IR
	if reqs, exists := requiredBy[irID]; exists {
		node.RequiredBy = reqs
	}

	return node
}

// addDependencies recursively adds dependencies to the graph
func (k *Keeper) addDependencies(ctx sdk.Context, irID string, nodes *[]types.IRGraphNode, requiredBy map[string][]string, visited map[string]bool) {
	if visited[irID] {
		return
	}
	visited[irID] = true

	prereq, exists := k.GetPrerequisite(ctx, irID)
	if !exists {
		return
	}

	for _, depID := range prereq.RequiredIrIds {
		if !visited[depID] {
			node := k.buildGraphNode(ctx, depID, requiredBy)
			*nodes = append(*nodes, node)
			k.addDependencies(ctx, depID, nodes, requiredBy, visited)
		}
	}
}

// buildPrerequisiteChain recursively builds prerequisite chain
func (k *Keeper) buildPrerequisiteChain(ctx sdk.Context, irID string, depth int, chain *[]types.IRGraphNode, visited map[string]bool) {
	if visited[irID] || depth > 10 {
		return
	}

	visited[irID] = true

	prereq, exists := k.GetPrerequisite(ctx, irID)
	if !exists {
		return
	}

	for _, requiredID := range prereq.RequiredIrIds {
		node := types.IRGraphNode{
			IrId:       requiredID,
			DependsOn:  []string{},
			RequiredBy: []string{irID},
		}
		*chain = append(*chain, node)

		// Recursively add sub-prerequisites
		k.buildPrerequisiteChain(ctx, requiredID, depth+1, chain, visited)
	}
}

// RemovePrerequisites removes all prerequisites for an IR
func (k *Keeper) RemovePrerequisites(ctx sdk.Context, irID string) error {
	// Check if IR exists
	if _, exists := k.GetIR(ctx, irID); !exists {
		return types.ErrIRNotFound
	}

	// Delete prerequisites from store
	store := k.storeService.OpenKVStore(ctx)
	return store.Delete([]byte(types.PrerequisiteStoreKey(irID)))
}

// GetPrerequisiteDepth returns the maximum depth of prerequisites for an IR
func (k *Keeper) GetPrerequisiteDepth(ctx sdk.Context, irID string) int {
	visited := make(map[string]bool)
	return k.calculatePrerequisiteDepth(ctx, irID, visited)
}

// calculatePrerequisiteDepth recursively calculates prerequisite depth
func (k *Keeper) calculatePrerequisiteDepth(ctx sdk.Context, irID string, visited map[string]bool) int {
	if visited[irID] {
		return 0
	}
	visited[irID] = true

	prereq, exists := k.GetPrerequisite(ctx, irID)
	if !exists || len(prereq.RequiredIrIds) == 0 {
		return 0
	}

	maxDepth := 0
	for _, requiredID := range prereq.RequiredIrIds {
		depth := k.calculatePrerequisiteDepth(ctx, requiredID, visited)
		if depth > maxDepth {
			maxDepth = depth
		}
	}

	return maxDepth + 1
}
