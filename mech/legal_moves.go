package mech

import "sankofa/ow"

// all moves that may executed from a given position are legal;
// records for each legal move the capture size and target rank
type LegalMoves struct {
	// starting position
	Rank int64

	// sorted keys=moves
	// * killer-move heuristic
	// * consistent String output
	Moves []int8

	// move ⇢ next rank
	Next map[int8]int64

	// move ⇢ score
	Score map[int8]int8
}

// empty LegalMoves structure
func NewLegalMoves() *LegalMoves {
	r := new(LegalMoves)

	r.Rank = ow.MININT64
	r.Moves = make([]int8, 0, MOVE_CAP)
	r.Next = make(map[int8]int64, MOVE_CAP)
	r.Score = make(map[int8]int8, MOVE_CAP)

	return r
}

func (in *LegalMoves) Clone() *LegalMoves {
	out := NewLegalMoves()
	out.Rank = in.Rank

	out.Moves = make([]int8, len(in.Moves))
	copy(out.Moves, in.Moves)

	out.Next = make(map[int8]int64, len(in.Next))
	for k, v := range in.Next {
		out.Next[k] = v
	}

	out.Score = make(map[int8]int8, len(in.Score))
	for k, v := range in.Score {
		out.Score[k] = v
	}

	return out
}

func (legalMoves *LegalMoves) String() string {
	r := ow.Thousands(legalMoves.Rank) + " ⇢ "

	for _, key := range legalMoves.Moves {
		r += MoveToString(key) + "=" + ow.Thousands(legalMoves.Next[key]) + "/" + ow.Thousands(legalMoves.Score[key]) + " "
	}

	return r
}

// get all valid moves and their target positions.
// an empty set means "no valid move", game finished.
// if all moves would lead to the starvation of the opponent, then they are allowed.
func (position *Position) LegalMoves() *LegalMoves {
	legalMoves := NewLegalMoves()
	// including those that would lead to starvation
	allMoves := NewLegalMoves()

	rank := position.Rank()
	legalMoves.Rank = rank
	allMoves.Rank = rank

	// find valid moves that feed the opponent
	for move := SOUTHLEFT; move <= SOUTHRIGHT; move++ {
		if position.Board[move] > 0 {
			target := position.Move(move)
			rank := target.Rank()
			// all moves, including non-feeding
			// only allowable when there are NO feeding moves
			allMoves.Next[move] = rank
			allMoves.Moves = append(allMoves.Moves, move)

			if position.Stones() == target.Stones() {
				allMoves.Score[move] = 0
			} else {
				allMoves.Score[move] = position.Scores[1] - position.Scores[0] + target.Scores[1] - target.Scores[0]
			}
			// only moves that do not leave the opponent starved
			if target.Starved() {
				ow.Log("does not feed:", move)
			} else {
				legalMoves.Next[move] = rank
				legalMoves.Moves = append(legalMoves.Moves, move)

				if position.Stones() == target.Stones() {
					legalMoves.Score[move] = 0
				} else {
					legalMoves.Score[move] = position.Scores[1] - position.Scores[0] + target.Scores[1] - target.Scores[0]
				}
			}
		}
	}

	if len(legalMoves.Next) == 0 {
		ow.Log(position, "no feeding moves exist:", allMoves)
		return allMoves
	} else {
		ow.Log(position, "⇢", legalMoves)
		return legalMoves
	}
}
