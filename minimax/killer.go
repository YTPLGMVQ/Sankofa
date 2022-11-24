package minimax

import (
	"sankofa/mech"
	"sankofa/ow"
	"sort"
)

// μολων λαβε; lazy memeoization
func (tt *TT) LegalMoves(rank int64) *mech.LegalMoves {
	var r *mech.LegalMoves

	tt.mutex.Lock()
	defer tt.mutex.Unlock()

	// get it if available
	r, ok := tt.legalMoves[rank]
	tt.cntLegalMoves += 1

	if !ok {
		// or create it if not
		r = mech.Unrank(rank).LegalMoves()
		tt.legalMoves[rank] = r
	}

	return r
}

// lazy memeoization
func (tt *TT) MovesInHand(rank int64) int64 {
	var r int64

	tt.mutex.Lock()
	defer tt.mutex.Unlock()

	// get it if available
	r, ok := tt.movesInHand[rank]
	tt.cntMovesInHand += 1

	if !ok {
		// or create it if not
		r = mech.Unrank(rank).MovesInHand()
		tt.movesInHand[rank] = r
	}

	return r
}

// sorted list of best/killer moves for a given rank
func (tt *TT) KillerMoves(rank int64) []int64 {
	legalMoves := tt.LegalMoves(rank)

	// build sorted array
	moves := make([]int64, 0)
	for move := range legalMoves.Next {
		moves = append(moves, move)
	}

	// re-sort, since the contextual information might have changed
	// WARNING without numeric sorting, the order is random and the results with 1 thread not predictible
	sort.Slice(moves, func(a, b int) bool {
		moveA := moves[a]
		moveB := moves[b]

		rankA := legalMoves.Next[moveA]
		rankB := legalMoves.Next[moveB]

		var intervalA, intervalB *Interval
		if tt.old != nil {
			intervalA = tt.old.Interval(rankA)
			intervalB = tt.old.Interval(rankB)
			ow.Log(intervalA, ">?", intervalB)
		}

		// sort criteria: TT, captures, feeding, front-back
		switch {
		case intervalA != nil && intervalB != nil && intervalA.Disjoint(intervalB):
			ow.Log("sort: score:", legalMoves, "|", mech.MoveToString(moveA), mech.MoveToString(moveB), "|", intervalA, intervalB)
			ow.Log(moveA, " ⇢ ", intervalA)
			ow.Log(moveB, " ⇢ ", intervalB)
			return intervalA.Gt(intervalB)
		case legalMoves.Score[moveA] != legalMoves.Score[moveB]:
			ow.Log("sort: captures:", legalMoves, "|", mech.MoveToString(moveA), mech.MoveToString(moveB))
			return legalMoves.Score[moveA] > legalMoves.Score[moveB]
		case tt.MovesInHand(rankA) != tt.MovesInHand(rankB):
			// preserve MIH; LT since we have the opponent's perspective!
			ow.Log("sort: Δν:", legalMoves, "|", mech.MoveToString(moveA), mech.MoveToString(moveB), "|", tt.MovesInHand(rankA), tt.MovesInHand(rankB))
			return tt.MovesInHand(rankA) < tt.MovesInHand(rankB)
		default:
			// base line: numeric sort
			ow.Log("sort: numerical:", legalMoves, "|", mech.MoveToString(moveA), mech.MoveToString(moveB))
			return moveA > moveB
		}
	})

	ow.Log("sorted legal moves:", legalMoves)

	return moves
}
