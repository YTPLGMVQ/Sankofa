package minimax

import (
	"sankofa/mech"
	"sankofa/ow"
	"sort"
)

// no-lock: position for rank
func (tt *TT) _position(rank int64) *mech.Position {
	var r *mech.Position

	// get it if available
	r, ok := tt.positions[rank]

	if !ok {
		// or create it if not
		r = mech.Unrank(rank)
		tt.positions[rank] = r
	}

	return r
}

// position for given rank; lazy memoization
func (tt *TT) Position(rank int64) *mech.Position {
	tt.mutex.Lock()
	defer tt.mutex.Unlock()

	return tt._position(rank)
}

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
		p := tt._position(rank)
		r = p.LegalMoves()
		tt.legalMoves[rank] = r
	}

	return r
}

// lazy memeoization
func (tt *TT) MovesInHand(rank int64) int8 {
	var r int8

	tt.mutex.Lock()
	defer tt.mutex.Unlock()

	// get it if available
	r, ok := tt.movesInHand[rank]
	tt.cntMovesInHand += 1

	if !ok {
		// or create it if not
		r = tt._position(rank).MovesInHand()
		tt.movesInHand[rank] = r
	}

	return r
}

// sorted list of best/killer moves for a given rank
func (tt *TT) KillerMoves(rank int64) []int8 {
	legalMoves := tt.LegalMoves(rank)

	// re-sort, since the contextual information might have changed
	// WARNING without numeric sorting, the order is random and the results with 1 thread not predictible
	sort.Slice(legalMoves.Moves, func(a, b int) bool {
		moveA := legalMoves.Moves[a]
		moveB := legalMoves.Moves[b]

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

	return legalMoves.Moves
}
