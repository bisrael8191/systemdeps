package systemdeps

import (
	"fmt"
)

// Used for finding cycles in DFS.
// Similar to the DFS Topological Sort Algorithm (https://en.wikipedia.org/wiki/Topological_sorting#Depth-first_search).
//
// Not meant to be used directly by users.
const (
	UNSEEN = iota
	EXPLORING
	DONE
)

// Systemd Unit directed graph structure.
// Unit is the vertex, requires become the edges in the graph.
type Unit struct {
	value    string
	requires []*Unit
	state    int
}

func (u Unit) String() string {
	return fmt.Sprintf(u.value)
}

// System graph representation.
type SystemGraph struct {
	HasCycle   bool
	CycleStart string
	CycleEnd   string
	units      map[string]*Unit
}

func newSystemGraph() *SystemGraph {
	return &SystemGraph{
		units: make(map[string]*Unit),
	}
}

// Create a new unit.
func (g *SystemGraph) createUnit(processName string) *Unit {
	if _, exists := g.units[processName]; !exists {
		g.units[processName] = &Unit{
			value: processName,
			state: UNSEEN,
		}
	}

	return g.units[processName]
}

// Create all process units and link them with their dependencies.
func (g *SystemGraph) createGraph(processes *Processes) {
	for _, p := range processes.Processes {
		unit := g.createUnit(p.Name)

		if len(p.Dependencies) > 0 {
			for _, depName := range p.Dependencies {
				if depUnit, exists := g.units[depName]; exists {
					unit.requires = append(unit.requires, depUnit)
				} else {
					// If not already in the unit map, must be a base dependency (I.E. docker, network, etc),
					// create a fake unit for it
					unit.requires = append(unit.requires, g.createUnit(depName))
				}
			}
		}
	}
}

// Helper function to print the structure.
func (g *SystemGraph) PrintGraph() {
	for _, unit := range g.units {
		fmt.Printf("Unit: %s\n    Requires: %v\n", unit, unit.requires)
	}
}

// Recursive DFS function.
func (g *SystemGraph) dfs(rootUnit *Unit) {
	// Base state, don't revisit completed nodes
	if rootUnit.state == DONE {
		return
	}

	// Set the current node to exploring
	rootUnit.state = EXPLORING

	// Loop through all requirements,
	// if it hasn't been seen yet run dfs recursively,
	// else if you see another exploring node on the stack, you've found a cycle
	for _, r := range rootUnit.requires {
		if r.state == UNSEEN {
			g.dfs(r)
		} else if r.state == EXPLORING {
			g.HasCycle = true
			g.CycleStart = rootUnit.value
			g.CycleEnd = r.value
		}
	}

	// Mark the node as fully done when the stack unwinds
	rootUnit.state = DONE
}

// Check if an array of processes has any dependency cycles.
func HasCycle(processes *Processes) (bool, *SystemGraph) {
	g := newSystemGraph()

	g.createGraph(processes)
	// g.printGraph()

	// Run DFS using every node as a starting point,
	// this guarantees that all nodes will be visited
	// even if the dependencies create multiple trees.
	// This is faster than it seems because the later loops
	// will just hit the base condition immediately.
	for _, unit := range g.units {
		g.dfs(unit)
	}

	return g.HasCycle, g
}
