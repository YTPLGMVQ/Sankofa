// mechanics of Oware game play
//
// # DESIGN, TACTICS AND HACKS
//
// Types:
//   - ranks are int64
//   - stones, scores, moves and verdicts are signed bytes | int8
//   - array indices are int
//
// REASONS
//   - ranks would overflow int on 32-bit architectures such as Rpi0W2
//   - a database entry shall be small (1 byte) in order to keep the size acceptable (even so: 1.4TB)
//   - Golang requires array indices to be int
//   - even so, some type acrobatics is required, because the flag pacakge cannot produce int8 command line arguments.
//
// Further:
//   - destructive operations on *Position or *Game return a new object and do not touch the original.
//   - cloning *Game for any destructive operation is rather efficient:
//     only Position pointers and moves are allocated/copied.
package mech

// representation of an Oware game

import (
	"sankofa/ow"
	"strconv"
	"strings"
)

////////////////////////////////////////////////////////////////
// SEQUENCE = GAME = LIST OF MOVES
////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////
// DATA TYPES
////////////////////////////////////////////////////////////////

// initial slice capacity
const GAME_CAP = 24

// game = a list of positions, a list of moves and a cursor;
// one less moves than positions;
// the cursor is a position index
type Game struct {
	Positions []*Position
	Moves     []int8
	Cursor    int
}

////////////////////////////////////////////////////////////////
// CONVERSIONS
////////////////////////////////////////////////////////////////

// empty game structure
func NewGame() *Game {
	game := new(Game)
	game.Positions = make([]*Position, 0, GAME_CAP)
	game.Moves = make([]int8, 0, GAME_CAP)
	game.Cursor = 0
	return game
}

// import from text (REST); format:
//
// /RANK/MOVE/.../!CURRENT-MOVE/.../MOVE(SCORE-SCORE)/...
func StringToGame(rest string) *Game {
	ow.Log(rest)

	// build game
	game := NewGame()
	cur := ow.MININT
	first := true

	for i, elem := range strings.Split(rest, "/") {
		switch i {
		case 0:
			// server address || empty
			continue
		case 1:
			// initial position
			rank, err := strconv.ParseInt(elem, 10, 64)
			ow.Check(err)
			if rank < MINRANK || rank > MAXRANK {
				ow.Panic("rank out of range:", rank)
			}
			game.Positions = append(game.Positions, Unrank(rank))
		default:
			// moves
			if game.GameOver() {
				ow.Panic("do no parse past game-over")
			}

			if len(elem) == 0 {
				ow.Panic("empty move:", i)
			}

			var move int8

			if elem[0] == '!' {
				// if len(elem) == 2 {
				move = StringToMove(strings.ToUpper(string(elem[1])))
				ow.Log("cursor marker:", i, first)
				if first {
					// take cursor at first "*" occurrence
					cur = i - 1
					first = false
				}
			} else {
				move = StringToMove(strings.ToUpper(string(elem[0])))
			}
			game = game.Move(move)

		}
	}

	// sanity check
	if len(game.Positions) != len(game.Moves)+1 {
		ow.Panic(len(game.Positions), "positions for", len(game.Moves), "moves")
	}

	game.Cursor = cur
	if game.Cursor == ow.MININT {
		game.Cursor = len(game.Positions) - 1
	}
	ow.Log("cursor:", game.Cursor, "/", len(game.Positions)-1)

	// done
	ow.Log(game)
	return game
}

// REST format
func (game *Game) String() string {
	// sanity check
	if len(game.Positions) != len(game.Moves)+1 {
		ow.Panic(len(game.Positions), "positions for", len(game.Moves), "moves")
	}
	if game.Cursor > len(game.Positions) {
		ow.Panic("cursor:", game.Cursor, "> #positions:", len(game.Positions))
	}

	moves := "/" + ow.Thousands(game.First().Rank())
	for i, move := range game.Moves {
		moves += "/"
		// no cursor needed for the last move
		if game.Cursor == i+1 && game.Cursor < len(game.Positions)-1 {
			moves += "!"
		}
		if ow.Even(i) {
			moves += MoveToString(move)
		} else {
			moves += MoveToString(ReverseMove(move))
		}

		// show score: captures or last position
		if (i > 1 && game.Positions[i].Scores[0] != game.Positions[i+1].Scores[1]) || i == len(game.Moves)-1 {
			// off by 1 beacause of the initial position
			moves += "("
			if ow.Even(i) {
				moves += ow.Thousands(game.Positions[i+1].Scores[1]) + "-" + ow.Thousands(game.Positions[i+1].Scores[0])
			} else {
				moves += ow.Thousands(game.Positions[i+1].Scores[0]) + "-" + ow.Thousands(game.Positions[i+1].Scores[1])
			}
			moves += ")"
		}
	}
	if game.GameOver() {
		moves += "."
	}
	return moves
}

////////////////////////////////////////////////////////////////
// OPERATIONS
////////////////////////////////////////////////////////////////

func (in *Game) Clone() *Game {
	out := NewGame()

	out.Cursor = in.Cursor

	out.Positions = make([]*Position, len(in.Positions))
	if copy(out.Positions, in.Positions) != len(in.Positions) {
		ow.Panic("incomplete copy: positions")
	}

	out.Moves = make([]int8, len(in.Moves))
	if copy(out.Moves, in.Moves) != len(in.Moves) {
		ow.Panic("incomplete copy: moves")
	}

	return out
}

// is this the same game?
func (this *Game) EQ(other *Game) bool {
	if len(this.Positions) != len(other.Positions) {
		ow.Log("different #positions:", this, "!=", other)
		return false
	}
	if this.Cursor != other.Cursor {
		ow.Log("different cursor:", this.Cursor, "!=", other.Cursor)
		return false
	}
	for i := range this.Positions {
		if !this.Positions[i].EQ(other.Positions[i]) {
			ow.Log("different position at:", i, ":", this, "!=", other)
			return false
		}
	}
	return true
}

////////////////////////////////////////////////////////////////
// PLACES IN THE GAME
////////////////////////////////////////////////////////////////

// first position
func (game *Game) First() *Position {
	return game.Positions[0]
}

// current position, pointed at by the cursor
func (game *Game) Current() *Position {
	return game.Positions[game.Cursor]
}

// position before the cursor
func (game *Game) BeforeCurrent() *Position {
	// no position before the first: use Reverse() instead
	if len(game.Positions) < 2 || game.Cursor < 1 {
		return game.Current().Reverse()
	}
	return game.Positions[game.Cursor-1]
}

// the last position
func (game *Game) Last() *Position {
	return game.Positions[len(game.Positions)-1]
}

// forelast position
func (game *Game) BeforeLast() *Position {
	// no position before the first: use Reverse() instead
	if len(game.Positions) < 2 {
		return game.Current().Reverse()
	}
	return game.Positions[len(game.Positions)-2]
}

////////////////////////////////////////////////////////////////
// SCORING
////////////////////////////////////////////////////////////////

// heuristic score for the current position;
// does not change the recorded scores
func (game *Game) Heuristic() int8 {
	position := game.Positions[game.Cursor]
	level := ow.Level(position.Rank())
	if ow.Even(level) {
		// only an even number of stones can be split to n:n
		return 0
	} else {
		return -1
	}
}

// score of the current position
func (game *Game) Capture() int8 {
	// no position before the first: use Reverse() instead
	if len(game.Positions) < 2 || game.Cursor < 1 {
		return game.Current().Score()
	} else {
		return game.Current().Score() + game.BeforeCurrent().Score()
	}
}

// true if the last position happens more than once
func (game *Game) Cycle() bool {
	for i := ow.ZERO64; i < int64(len(game.Positions)-1); i++ {
		if game.Last().EQ(game.Positions[i]) {
			ow.Log("found cycle: @", ow.Thousands(len(game.Positions)-1), "== @", ow.Thousands(i))
			return true
		}
	}
	return false
}

// game over if the last position's score is final or there is a cycle
func (game *Game) GameOver() bool {
	// WARNING do not log game.String(): it calls GameOver() ??? stack overflow
	if game.Last().GameOver() || game.Cycle() {
		return true
	}
	return false
}
