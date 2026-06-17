package topology

import (
	"github.com/nora-language/nora/pkg/semantic"
)

// SolveLiveness runs backward data-flow analysis on all nodes in a Control-Flow Graph
// to compute the LiveIn and LiveOut sets for all variables.
func SolveLiveness(allNodes []*CFGNode, exitNode *CFGNode) {
	// Include the exit node in the iteration list
	nodes := make([]*CFGNode, len(allNodes))
	copy(nodes, allNodes)
	if exitNode != nil {
		nodes = append(nodes, exitNode)
	}

	// Classic fixed-point iteration
	changed := true
	for changed {
		changed = false

		// Iterate backwards through the nodes list (often faster convergence for backward analysis)
		for i := len(nodes) - 1; i >= 0; i-- {
			node := nodes[i]

			// 1. Compute LiveOut[n] = Union of LiveIn[s] for all s in Succs[n]
			newLiveOut := make(map[*semantic.Symbol]bool)
			for _, succ := range node.Succs {
				for sym, live := range succ.LiveIn {
					if live {
						newLiveOut[sym] = true
					}
				}
			}

			// Check if LiveOut changed
			if !equalSymbolMaps(node.LiveOut, newLiveOut) {
				node.LiveOut = newLiveOut
				changed = true
			}

			// 2. Compute LiveIn[n] = Gen[n] U (LiveOut[n] \ Kill[n])
			newLiveIn := make(map[*semantic.Symbol]bool)
			// Start with Gen[n]
			for sym, gen := range node.Gen {
				if gen {
					newLiveIn[sym] = true
				}
			}
			// Add (LiveOut[n] \ Kill[n])
			for sym, live := range node.LiveOut {
				if live && !node.Kill[sym] {
					newLiveIn[sym] = true
				}
			}

			// Check if LiveIn changed
			if !equalSymbolMaps(node.LiveIn, newLiveIn) {
				node.LiveIn = newLiveIn
				changed = true
			}
		}
	}
}

// equalSymbolMaps checks if two map[*semantic.Symbol]bool represent the same set of active symbols
func equalSymbolMaps(m1, m2 map[*semantic.Symbol]bool) bool {
	// Count true values
	t1 := countTrueSymbols(m1)
	t2 := countTrueSymbols(m2)
	if t1 != t2 {
		return false
	}

	for sym, live := range m1 {
		if live && !m2[sym] {
			return false
		}
	}
	return true
}

func countTrueSymbols(m map[*semantic.Symbol]bool) int {
	count := 0
	for _, live := range m {
		if live {
			count++
		}
	}
	return count
}
