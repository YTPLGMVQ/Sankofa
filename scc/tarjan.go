package scc

import (
	"math/big"
	"sankofa/mech"
	"sankofa/ow"
)

// index: the traversal order
var index []int64

// lowLink: lowest index reachable in the subtree of given node
var lowLink []int64

// stack of nodes
var stack *Stack

// membership bit maps
var stackMember, indexMember *big.Int

// range
var high, low int64

// current index
var count int64

func Tarjan(level int8) (scc []int64) {
	scc = make([]int64, 0)

	// initialize globals
	// range
	ow.Log("level:", level)
	low = ow.ZERO64
	if level > 0 {
		low = ow.LevelUpperLimits[level-1] + 1
	}
	high = ow.LevelUpperLimits[level]
	ow.Log("from:", low, "to:", high)

	index = make([]int64, high-low+1)
	lowLink = make([]int64, high-low+1)
	stack = NewStack()
	stackMember = big.NewInt(0)
	indexMember = big.NewInt(0)

	for rank := low; rank <= high; rank++ {
		// WARNING Tarjan only works on 64-bit architectures: indices would overflow an int32
		if indexMember.Bit(int(rank-low)) == 0 {
			ow.Log("⇢enter with:", rank)
			scc = append(scc, strongConnect(rank)...)
		} else {
			ow.Log("⇠skip/seen:", rank)
		}
	}

	ow.Log("remaining stack size:", stack.Size())
	for pop, ok := stack.Pop(); ok; pop, ok = stack.Pop() {
		ow.Log("POP: leftover:", pop, "index:", index[pop-low], "lowLink:", lowLink[pop-low], "stack size:", stack.Size())
	}

	return
}

// Find strongly connected components in a directed graph.
// Oware positions are the nodes, legal moves at the same level are the vertices.
// The procedure is called for each node (as a potential starting point).
// Already seen nodes are then skipped.
// The minor split score is saved for all SCC members that are not already set.
func strongConnect(rank int64) []int64 {
	scc := make([]int64, 0)

	// initialize book keeping
	index[rank-low] = count
	// WARNING Tarjan only works on 64-bit architectures: indices would overflow an int32
	indexMember.SetBit(indexMember, int(rank-low), 1)
	lowLink[rank-low] = count
	count = count + 1

	// place on stack
	stack.Push(rank)
	stackMember.SetBit(stackMember, int(rank-low), 1)
	ow.Log("PUSH:", rank, "index:", index[rank-low], "stack size:", stack.Size())

	// successors
	legalMoves := mech.Unrank(rank).LegalMoves()
	for move := range legalMoves.Next {
		// only same-level
		if legalMoves.Score[move] != 0 {
			continue
		}

		// (rank, next) is an edge of the directed graph.
		next := legalMoves.Next[move]
		// WARNING Tarjan only works on 64-bit architectures: indices would overflow an int32
		if indexMember.Bit(int(next-low)) == 0 {
			ow.Log("recurse on:", next)
			strongConnect(next)
			lowLink[rank-low] = ow.Min(lowLink[rank-low], lowLink[next-low])
		} else if stackMember.Bit(int(next-low)) == 1 {
			// WARNING Tarjan only works on 64-bit architectures: indices would overflow an int32
			lowLink[rank-low] = ow.Min(lowLink[rank-low], index[next-low])
		}
	}

	// dump strongly connected component if this is a root node
	if lowLink[rank-low] == index[rank-low] {
		ow.Log("root:", rank, "index:", index[rank-low], "lowLink:", lowLink[rank-low])
		// count elements strongly connected component
		var size int64
		for pop, ok := stack.Pop(); ok; pop, ok = stack.Pop() {
			size = size + 1
			// WARNING Tarjan only works on 64-bit architectures: indices would overflow an int32
			stackMember.SetBit(stackMember, int(rank-low), 0)
			ow.Log("POP:", pop, "index:", index[pop-low], "stack size:", stack.Size())
			// until reached root
			if pop == rank {
				// one-man SCCs don't count
				if size == 1 {
					ow.Log("free:", rank, "index:", index[rank-low], "lowLink:", lowLink[rank-low])
				} else {
					ow.Log("SCC root: ", rank, "index:", index[rank-low], "lowLink:", lowLink[rank-low])
					// output
					scc = append(scc, pop)
				}
				break
			} else {
				ow.Log("SCC element: ", pop, "index:", index[pop-low], "lowLink:", lowLink[pop-low])
				// output
				scc = append(scc, pop)
			}
		}
	}

	return scc
}
