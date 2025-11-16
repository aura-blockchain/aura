package keeper

import (
	"fmt"

	"github.com/aequitas/aura/chain/x/inclusionroutines/types"
)

// GetPrerequisites retrieves the prerequisites for an IR
func (k *Keeper) GetPrerequisites(irID string) (types.IRPrerequisite, bool) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	prereq, ok := k.prerequisites[irID]
	return prereq, ok
}

// SetPrerequisites sets the prerequisites for an IR
func (k *Keeper) SetPrerequisites(irID string, requiredIRIDs []string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	// Check if IR exists
	if _, exists := k.irs[irID]; !exists {
		return types.ErrIRNotFound
	}

	// Validate all required IRs exist
	for _, reqID := range requiredIRIDs {
		if reqID == irID {
			return types.ErrSelfPrerequisite
		}
		if _, exists := k.irs[reqID]; !exists {
			return fmt.Errorf("%w: %s", types.ErrPrerequisiteNotFound, reqID)
		}
	}

	// Create new prerequisite relationship
	newPrereq := types.IRPrerequisite{
		IRID:          irID,
		RequiredIRIDs: requiredIRIDs,
	}

	// Check for circular dependencies
	if err := k.detectCircularDependency(irID, requiredIRIDs); err != nil {
		return err
	}

	k.prerequisites[irID] = newPrereq
	return nil
}

// GetIRGraph returns the prerequisite dependency graph for an IR
// If irID is empty, returns the full graph
func (k *Keeper) GetIRGraph(irID string) []types.IRGraphNode {
	k.mu.RLock()
	defer k.mu.RUnlock()

	// Build reverse lookup (which IRs require each IR)
	requiredBy := make(map[string][]string)
	for id, prereq := range k.prerequisites {
		for _, reqID := range prereq.RequiredIRIDs {
			requiredBy[reqID] = append(requiredBy[reqID], id)
		}
	}

	nodes := make([]types.IRGraphNode, 0)

	if irID != "" {
		// Return subgraph for specific IR
		node := k.buildGraphNode(irID, requiredBy)
		nodes = append(nodes, node)

		// Add all dependencies recursively
		visited := make(map[string]bool)
		k.addDependencies(irID, &nodes, requiredBy, visited)
	} else {
		// Return full graph
		for id := range k.irs {
			node := k.buildGraphNode(id, requiredBy)
			nodes = append(nodes, node)
		}
	}

	return nodes
}

// ValidatePrerequisites checks if a wallet has completed all prerequisites for an IR
func (k *Keeper) ValidatePrerequisites(irID string, completedIRs []string) error {
	k.mu.RLock()
	defer k.mu.RUnlock()

	// Check if IR exists
	if _, exists := k.irs[irID]; !exists {
		return types.ErrIRNotFound
	}

	// Get prerequisites for this IR
	prereq, hasPrereqs := k.prerequisites[irID]
	if !hasPrereqs || len(prereq.RequiredIRIDs) == 0 {
		return nil // No prerequisites required
	}

	// Build set of completed IRs for fast lookup
	completed := make(map[string]bool)
	for _, id := range completedIRs {
		completed[id] = true
	}

	// Check all prerequisites are met
	missing := make([]string, 0)
	for _, reqID := range prereq.RequiredIRIDs {
		if !completed[reqID] {
			missing = append(missing, reqID)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("%w: missing %v", types.ErrPrerequisiteNotMet, missing)
	}

	return nil
}

// detectCircularDependency checks for cycles in the prerequisite graph
func (k *Keeper) detectCircularDependency(irID string, newRequirements []string) error {
	// Build temporary graph with new prerequisite
	graph := make(map[string][]string)
	for id, prereq := range k.prerequisites {
		graph[id] = prereq.RequiredIRIDs
	}
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
func (k *Keeper) buildGraphNode(irID string, requiredBy map[string][]string) types.IRGraphNode {
	node := types.IRGraphNode{
		IRID:       irID,
		DependsOn:  []string{},
		RequiredBy: []string{},
	}

	// Add prerequisites (what this IR depends on)
	if prereq, exists := k.prerequisites[irID]; exists {
		node.DependsOn = prereq.RequiredIRIDs
	}

	// Add what requires this IR
	if reqs, exists := requiredBy[irID]; exists {
		node.RequiredBy = reqs
	}

	return node
}

// addDependencies recursively adds dependencies to the graph
func (k *Keeper) addDependencies(irID string, nodes *[]types.IRGraphNode, requiredBy map[string][]string, visited map[string]bool) {
	if visited[irID] {
		return
	}
	visited[irID] = true

	prereq, exists := k.prerequisites[irID]
	if !exists {
		return
	}

	for _, depID := range prereq.RequiredIRIDs {
		if !visited[depID] {
			node := k.buildGraphNode(depID, requiredBy)
			*nodes = append(*nodes, node)
			k.addDependencies(depID, nodes, requiredBy, visited)
		}
	}
}
