package actions

import (
	"fmt"
)

// DependencyGraph manages action dependencies
type DependencyGraph struct {
	nodes     map[string]bool
	edges     map[string][]string // adjacency list: from -> []to
	inDegree  map[string]int      // number of incoming edges
}

// NewDependencyGraph creates a new dependency graph
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		nodes:    make(map[string]bool),
		edges:    make(map[string][]string),
		inDegree: make(map[string]int),
	}
}

// AddNode adds a node to the graph
func (g *DependencyGraph) AddNode(name string) {
	if _, exists := g.nodes[name]; !exists {
		g.nodes[name] = true
		g.inDegree[name] = 0
		g.edges[name] = []string{}
	}
}

// AddEdge adds a directed edge from -> to (from must complete before to)
func (g *DependencyGraph) AddEdge(from, to string) {
	// Ensure both nodes exist
	g.AddNode(from)
	g.AddNode(to)

	// Add edge
	g.edges[from] = append(g.edges[from], to)
	g.inDegree[to]++
}

// HasCycle detects if there's a cycle in the graph
func (g *DependencyGraph) HasCycle() bool {
	// Use DFS with color marking for cycle detection
	// White (0): unvisited, Gray (1): visiting, Black (2): visited
	colors := make(map[string]int)
	for node := range g.nodes {
		colors[node] = 0 // white
	}

	var dfs func(node string) bool
	dfs = func(node string) bool {
		colors[node] = 1 // gray

		for _, neighbor := range g.edges[node] {
			if colors[neighbor] == 1 { // gray - back edge found
				return true
			}
			if colors[neighbor] == 0 { // white - unvisited
				if dfs(neighbor) {
					return true
				}
			}
		}

		colors[node] = 2 // black
		return false
	}

	// Check each component
	for node := range g.nodes {
		if colors[node] == 0 {
			if dfs(node) {
				return true
			}
		}
	}

	return false
}

// TopologicalSort returns nodes in topological order
func (g *DependencyGraph) TopologicalSort() ([]string, error) {
	if g.HasCycle() {
		return nil, ErrCyclicDependency(g.findCycle())
	}

	// Create a copy of inDegree to work with
	inDegree := make(map[string]int)
	for k, v := range g.inDegree {
		inDegree[k] = v
	}

	// Find all nodes with no incoming edges
	queue := []string{}
	for node, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, node)
		}
	}

	result := []string{}

	// Process nodes with no dependencies first
	for len(queue) > 0 {
		// Dequeue
		node := queue[0]
		queue = queue[1:]
		result = append(result, node)

		// For each neighbor, decrease in-degree
		for _, neighbor := range g.edges[node] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// Check if all nodes were processed
	if len(result) != len(g.nodes) {
		return nil, fmt.Errorf("unable to create topological sort")
	}

	return result, nil
}

// ValidateSequence checks if a sequence of actions respects dependencies
func (g *DependencyGraph) ValidateSequence(actions []string) error {
	// Create a subgraph with only the specified actions
	subgraph := NewDependencyGraph()
	actionSet := make(map[string]bool)

	for _, action := range actions {
		actionSet[action] = true
		subgraph.AddNode(action)
	}

	// Add edges only between actions in our set
	for _, action := range actions {
		if edges, exists := g.edges[action]; exists {
			for _, to := range edges {
				if actionSet[to] {
					subgraph.AddEdge(action, to)
				}
			}
		}
	}

	// Check for cycles in the subgraph
	if subgraph.HasCycle() {
		return ErrCyclicDependency(subgraph.findCycle())
	}

	return nil
}

// GetDependencies returns all dependencies for a given action
func (g *DependencyGraph) GetDependencies(action string) []string {
	var deps []string
	visited := make(map[string]bool)

	var dfs func(node string)
	dfs = func(node string) {
		// Look for nodes that must complete before this one
		for from, edges := range g.edges {
			for _, to := range edges {
				if to == node && !visited[from] {
					visited[from] = true
					deps = append(deps, from)
					dfs(from) // Recursively get transitive dependencies
				}
			}
		}
	}

	dfs(action)
	return deps
}

// GetDependents returns all actions that depend on the given action
func (g *DependencyGraph) GetDependents(action string) []string {
	var deps []string
	visited := make(map[string]bool)

	var dfs func(node string)
	dfs = func(node string) {
		if edges, exists := g.edges[node]; exists {
			for _, to := range edges {
				if !visited[to] {
					visited[to] = true
					deps = append(deps, to)
					dfs(to) // Recursively get transitive dependents
				}
			}
		}
	}

	dfs(action)
	return deps
}

// findCycle returns the nodes involved in a cycle (if one exists)
func (g *DependencyGraph) findCycle() []string {
	colors := make(map[string]int)
	parent := make(map[string]string)
	var cycle []string

	for node := range g.nodes {
		colors[node] = 0 // white
	}

	var dfs func(node string) bool
	dfs = func(node string) bool {
		colors[node] = 1 // gray

		for _, neighbor := range g.edges[node] {
			parent[neighbor] = node
			if colors[neighbor] == 1 { // gray - back edge found
				// Reconstruct cycle
				cycle = []string{neighbor}
				current := node
				for current != neighbor && current != "" {
					cycle = append(cycle, current)
					current = parent[current]
				}
				return true
			}
			if colors[neighbor] == 0 { // white - unvisited
				if dfs(neighbor) {
					return true
				}
			}
		}

		colors[node] = 2 // black
		return false
	}

	// Check each component
	for node := range g.nodes {
		if colors[node] == 0 {
			if dfs(node) {
				break
			}
		}
	}

	return cycle
}

// GetParallelGroups returns groups of actions that can be executed in parallel
func (g *DependencyGraph) GetParallelGroups(actions []string) ([][]string, error) {
	// First get the topological order
	order, err := g.TopologicalSort()
	if err != nil {
		return nil, err
	}

	// Filter to only include requested actions
	actionSet := make(map[string]bool)
	for _, a := range actions {
		actionSet[a] = true
	}

	// Group actions by their level (distance from root)
	levels := make(map[string]int)
	maxLevel := 0

	// Calculate levels
	for _, action := range order {
		if !actionSet[action] {
			continue
		}

		level := 0
		// Find max level of dependencies
		for from, edges := range g.edges {
			for _, to := range edges {
				if to == action && actionSet[from] {
					if l, exists := levels[from]; exists && l >= level {
						level = l + 1
					}
				}
			}
		}
		levels[action] = level
		if level > maxLevel {
			maxLevel = level
		}
	}

	// Group by level
	groups := make([][]string, maxLevel+1)
	for action, level := range levels {
		groups[level] = append(groups[level], action)
	}

	// Remove empty groups
	result := [][]string{}
	for _, group := range groups {
		if len(group) > 0 {
			result = append(result, group)
		}
	}

	return result, nil
}