package mech

import "sankofa/ow"

// execute moves on Oware positions

////////////////////////////////////////////////////////////////
// DATA TYPES
////////////////////////////////////////////////////////////////

// initial slice capacity
const MOVE_CAP = 6

////////////////////////////////////////////////////////////////
// CONVERSIONS
////////////////////////////////////////////////////////////////

// move names using the ABCDEFabcdef notation
const (
	A int64 = iota
	B
	C
	D
	E
	F
	a
	b
	c
	d
	e
	f
)

// letter to move object
func StringToMove(external string) int64 {
	var move int64
	switch external {
	case "A":
		move = A
	case "B":
		move = B
	case "C":
		move = C
	case "D":
		move = D
	case "E":
		move = E
	case "F":
		move = F
	case "a":
		move = a
	case "b":
		move = b
	case "c":
		move = c
	case "d":
		move = d
	case "e":
		move = e
	case "f":
		move = f
	default:
		ow.Panic("no such move:", external)
	}
	return move
}

// corresponding move on the other side of the table
func Swap(move int64) (x int64) {
	switch move {
	case A:
		x = a
	case B:
		x = b
	case C:
		x = c
	case D:
		x = d
	case E:
		x = e
	case F:
		x = f
	case a:
		x = A
	case b:
		x = B
	case c:
		x = C
	case d:
		x = D
	case e:
		x = E
	case f:
		x = F
	default:
		ow.Panic("no such move")
	}
	return x
}

func MoveToString(move int64) (s string) {
	switch move {
	case F:
		s = "F"
	case E:
		s = "E"
	case D:
		s = "D"
	case C:
		s = "C"
	case B:
		s = "B"
	case A:
		s = "A"
	case f:
		s = "f"
	case e:
		s = "e"
	case d:
		s = "d"
	case c:
		s = "c"
	case b:
		s = "b"
	case a:
		s = "a"
	default:
		ow.Panic("no such move")
	}
	return s
}

////////////////////////////////////////////////////////////////
// OPERATIONS
////////////////////////////////////////////////////////////////

// execute a move on a position; return new position
func (in *Position) Move(move int64) *Position {
	// plausibility: are there any stones to move?
	stones := in.Board[move]
	if stones == 0 {
		ow.Panic("cannot move an empty house:", in, move)
	}

	// initialize
	out := in.Clone()

	// saw
	var cursor int64
	out.Board[move] = 0
	for i := int64(move); stones > 0; i++ {
		switch i % 12 {
		case int64(move):
			continue
		default:
			out.Board[i%12] += 1
			stones -= 1
			cursor = i % 12
		}
	}

	// collection is forbidden in the case of a grand slam
	checkpoint := out.Clone()

	// collect
	stones = in.Board[move]
	for i := cursor; i > 5; i-- {
		if out.Board[i] == 2 || out.Board[i] == 3 {
			out.Scores[0] += out.Board[i]
			out.Board[i] = 0
		} else {
			break
		}
	}

	// grand slam captures nothing
	// except: when a grand slam is the only possible move
	// resume at chackpoint
	if !checkpoint.Reverse().IsStarved() && out.Reverse().IsStarved() {
		ow.Log("a grand slam captures nothing")
		out = checkpoint
	}

	return out.Reverse()
}

// execute a move from the current (cursor) position; return a new game
func (in *Game) Move(move int64) *Game {
	out := NewGame()

	out.Positions = make([]*Position, int(in.Cursor)+1)
	if copy(out.Positions, in.Positions[:in.Cursor+1]) != int(in.Cursor)+1 {
		ow.Panic("incomplete copy: positions")
	}

	out.Moves = make([]int64, int(in.Cursor))
	if copy(out.Moves, in.Moves[:in.Cursor]) != int(in.Cursor) {
		ow.Panic("incomplete copy: moves")
	}

	out.Cursor = in.Cursor
	out.Positions = append(out.Positions, out.Last().Move(move))
	out.Moves = append(out.Moves, move)
	out.Cursor = in.Cursor + 1

	// split stones, if cycle or no legal moves left
	if out.Cycle() || out.Last().IsStarved() {
		// split stones
		position := out.Last()
		position.Scores[0] += position.SouthStones()
		position.Scores[1] += position.NorthStones()
	}

	ow.Log(in, in.Last().Board, "+", MoveToString(move), "⇢ ", out)

	return out
}

// moves in hand with one house and an empty interval at right; helper method
func countMoves(stones, interval int64) (moves, rest int64) {
	if stones > interval {
		rest = stones - interval
		stones = interval
	}

	if stones > 0 {
		moves = stones*(stones+1)/2 + stones*(interval-1-stones)
	}

	ow.Log("stones:", stones, "inteval:", interval, "rest:", rest, "moves:", moves)
	return
}

// maximum number of consecutive moves without reaching to the opponent's side
func (position *Position) MovesInHand() int64 {
	mih := ow.ZERO
	intervals := make([]int64, 0, MOVE_CAP)
	intervals = append(intervals, ow.ONE)

	// SOUTHRIGHT cannot possibly be moved without affecting the opponent's board
	for i := SOUTHRIGHT - 1; i >= SOUTHLEFT; i-- {
		// new interval at each obstacle
		if position.Board[i] > SOUTHRIGHT-i {
			intervals = append(intervals, ow.ONE)
			continue
		}

		// count moves, but only if some stones are left (rest)
		rest := position.Board[i]
		if rest > 0 {
			// initial sawing
			mih += 1

			// count moves in each interval
			for j := len(intervals) - 1; rest > 0 && j >= 0; j-- {
				ow.Log("j:", j, "rest:", rest)
				var moves int64
				moves, rest = countMoves(rest, intervals[j])
				mih += moves
			}
		}

		// increment current interval
		intervals[len(intervals)-1] += 1
	}

	ow.Log(position.Board, "intervals:", intervals, "MIH:", mih)
	return mih
}
