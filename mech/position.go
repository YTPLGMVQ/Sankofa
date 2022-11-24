package mech

// Oware position = board + score

import (
	"fmt"
	"sankofa/ow"
)

////////////////////////////////////////////////////////////////
// DATA TYPES
////////////////////////////////////////////////////////////////

// position = board + score
type Position struct {
	Board  Board    // number of stones in each house
	Scores [2]int64 // scores
}

////////////////////////////////////////////////////////////////
// CONVERSIONS
////////////////////////////////////////////////////////////////

// from human-readable format, e.g.: 4.4.4.4.4.4-4.4.4.4.4.4
func StringToPosition(external string) *Position {
	position := new(Position)
	position.Board = StringToBoard(external)
	return position
}

func (position *Position) String() string {
	return fmt.Sprintf(
		"rank: %v, board: %v, score: %v:%v",
		position.Rank(),
		position.Board.String(),
		position.Scores[0],
		position.Scores[1],
	)
}

////////////////////////////////////////////////////////////////
// RANKING
////////////////////////////////////////////////////////////////

// bijection; assigns each position a distinct rank in the contiguous interval [0, MAX]
func (position *Position) Rank() int64 {
	var combinadics [13]int64
	combinadics[0] = -1
	for i := SOUTHLEFT; i <= NORTHRIGHT; i++ {
		combinadics[i+1] = combinadics[i] + position.Board[i] + 1
	}

	rank := int64(-1)
	for i, c := range combinadics {
		rank = rank + ow.Binomial(int64(c), int64(i))
	}

	if position.Stones() > MAXSTONES {
		ow.Panic("out of range: ", rank)
	}

	return rank
}

// position for a given rank
func Unrank(rank int64) *Position {
	if rank < MINRANK || MAXRANK > 1399358844974 {
		ow.Panic("out of range: ", rank)
	}
	var combinadics [13]int64
	position := new(Position)
	rest := rank
	for d := int64(12); d > 0; d-- {
		var i int64
		for i = ow.ZERO; ow.Binomial(i+1, d) <= rest; {
			i = i + 1
		}
		combinadics[d] = i
		rest = rest - ow.Binomial(i, d)
	}
	combinadics[0] = -1
	for j := 0; j < 12; j++ {
		position.Board[j] = combinadics[j+1] - combinadics[j] - 1
	}

	return position
}

////////////////////////////////////////////////////////////////
// OPERATIONS
////////////////////////////////////////////////////////////////

// clone position; board & everything
func (in *Position) Clone() *Position {
	out := new(Position)
	out.Scores = in.Scores
	for i := range out.Board {
		out.Board[i] = in.Board[i]
	}
	return out
}

// reverse board and the score; the opponent moves
func (in *Position) Reverse() *Position {
	out := in.Clone()
	out.Scores[0], out.Scores[1] = in.Scores[1], in.Scores[0]
	for i := range out.Board {
		if i < 6 {
			out.Board[i] = in.Board[i+6]
		} else {
			out.Board[i] = in.Board[i-6]
		}
	}
	return out
}

// change the number of stones in a house by a given count
func (in *Position) Edit(i, count int64) *Position {
	out := in.Clone()

	// avoid overflows
	count = ow.Min(count, MAXSTONES-in.Stones())
	out.Board[i] = ow.Max(out.Board[i]+count, 0)

	return out
}

////////////////////////////////////////////////////////////////
// SCORING / EQ
////////////////////////////////////////////////////////////////

// same board (disregards the score)?
func (this *Position) EQ(other *Position) bool {
	return this.Board == other.Board
}

// number of stones on the northern side of the board
func (position *Position) NorthStones() int64 {
	var stones int64
	for i := NORTHLEFT; i <= NORTHRIGHT; i++ {
		stones += position.Board[i]
	}
	return stones
}

// number of stones on the southern side of the board
func (position *Position) SouthStones() int64 {
	var stones int64
	for i := SOUTHLEFT; i <= SOUTHRIGHT; i++ {
		stones += position.Board[i]
	}
	return stones
}

// total number of stones on the board
func (position *Position) Stones() int64 {
	return position.SouthStones() + position.NorthStones()
}

// score when each player takes the stones on her side
func (position *Position) Split() int64 {
	return position.SouthStones() - position.NorthStones()
}

// score at this position;
// Game.Move() etc. is responsibile for the book keeping
func (position *Position) Score() int64 {
	return position.Scores[0] - position.Scores[1]
}

// is South left without stones (starved)?;
// starving the opponent is allowed only when there are no feeding moves left
func (position *Position) IsStarved() bool {
	for i := SOUTHLEFT; i <= SOUTHRIGHT; i++ {
		if position.Board[i] > 0 {
			return false
		}
	}
	return true
}

// is it a win?
func (position *Position) IsWin() bool {
	if position.Scores[0] > MAXSTONES/2 {
		ow.Log("win:", position.Board)
		return true
	}
	return false
}

// is it a loss?
func (position *Position) IsLoss() bool {
	if position.Scores[1] > MAXSTONES/2 {
		ow.Log("loss:", position.Board)
		return true
	}
	return false
}

// is it a draw?
func (position *Position) IsDraw() bool {
	if position.Scores[0] == MAXSTONES/2 && position.Scores[1] == MAXSTONES/2 {
		ow.Log("draw:", position)
		return true
	}
	return false
}

// game over?
func (position *Position) GameOver() bool {
	if position.IsWin() || position.IsLoss() || position.IsDraw() || position.IsStarved() {
		ow.Log("game over", position.Board)
		return true
	}
	return false
}
