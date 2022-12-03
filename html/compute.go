package html

// compute a position's parameters which are then needed by the web front-end

import (
	"fmt"
	"sankofa/mech"
	"sankofa/minimax"
	"sankofa/ow"
)

////////////////////////////////////////////////////////////////
// GLOBALS
////////////////////////////////////////////////////////////////

// duration limit for the iterative deepener
var DurationLimit = float64(1)

// degree of parallelism
var Goroutines = 5

////////////////////////////////////////////////////////////////
// TYPES
////////////////////////////////////////////////////////////////

type Position struct {
	position *mech.Position
	// α—β: minimax score
	αβ string
	// δν: advantage in degrees of freedom
	// changed: Δ the number of stones in the house
	rank        int64
	δν, changed int8
	// movable: can we move this house?
	// scored: has a valid score?
	// attack: this moves captures stones
	// check: this field is attacked
	movable, scored, attack, check bool
}

type Game struct {
	game *mech.Game
	// previous is from the S perspective
	south, north, previous *Position
	// position *after* index-house is moved; or nil if move not possible
	moves [12]*Position
	// transposition table
	tt *minimax.TT
}

////////////////////////////////////////////////////////////////
// WORK
////////////////////////////////////////////////////////////////

func (position *Position) String() string {
	var r string
	if position.movable {
		r += "position: " + position.position.Board.String() +
			", rank: " + ow.Thousands(position.rank) +
			", α—β: " + position.αβ +
			", Δν: " + ow.Thousands(position.δν) + ", "
	}
	r += "movable: " + ow.YesNo(position.movable) +
		", changed: " + ow.Thousands(position.changed) +
		", attack: " + ow.YesNo(position.attack) +
		", check: " + ow.YesNo(position.check)
	return r
}

func Analysis(trail string) *Game {
	////////////////////////////////////////////////////////////////
	// PREPARE DATA STRUCTURES
	////////////////////////////////////////////////////////////////

	game := new(Game)

	game.game = mech.StringToGame(trail)

	fmt.Println("rank:", game.game.Current().Rank())
	fmt.Println(game.game.String(), "⇢request")

	for i := mech.SOUTHLEFT; i <= mech.NORTHRIGHT; i++ {
		game.moves[i] = new(Position)
	}

	game.south = new(Position)
	game.previous = new(Position)
	game.north = new(Position)

	// triggers String() to output the position/board
	game.south.movable = true
	game.previous.movable = true
	game.north.movable = true

	// board orientation
	if ow.Even(game.game.Cursor) {
		// S to move
		game.south.position = game.game.Current()
		game.previous.position = game.game.BeforeCurrent().Reverse()
		game.north.position = game.game.Current().Reverse()
	} else {
		// N to move
		game.south.position = game.game.Current().Reverse()
		game.previous.position = game.game.BeforeCurrent()
		game.north.position = game.game.Current()
	}

	game.south.rank = game.south.position.Rank()
	game.previous.rank = game.previous.position.Rank()
	game.north.rank = game.north.position.Rank()

	////////////////////////////////////////////////////////////////
	// SERIOUS WORK
	////////////////////////////////////////////////////////////////

	game.tt = minimax.NewTT(game.game).Explore(Goroutines, DurationLimit)
	game.game = game.tt.Game()

	////////////////////////////////////////////////////////////////
	// PROCESS RESULTS
	////////////////////////////////////////////////////////////////

	// board orientation
	if ow.Even(game.game.Cursor) {
		// S to move
		game.south.scored = game.tt.Known(game.south.rank)
		game.south.αβ = game.tt.Interval(game.south.rank).String()
	} else {
		// N to move
		game.north.scored = game.tt.Known(game.north.rank)
		game.north.αβ = game.tt.Interval(game.north.rank).String()
	}
	ow.Log(game.tt)

	game.south.δν = game.south.position.MovesInHand() - game.north.position.MovesInHand()
	game.north.δν = -game.south.δν

	southMoves := game.south.position.LegalMoves().Next
	northMoves := game.north.position.LegalMoves().Next

	// ⇢ compute all possible moves
	for i := mech.SOUTHLEFT; i <= mech.NORTHRIGHT; i++ {
		var rank int64
		var ok bool

		if i < mech.NORTHLEFT {
			rank, ok = southMoves[i]
		} else {
			rank, ok = northMoves[i-mech.NORTHLEFT]
		}

		if ok {
			game.moves[i].movable = true
			game.moves[i].rank = rank
			game.moves[i].position = mech.Unrank(rank)
			// ⇢ αβ
			game.moves[i].scored = game.tt.Known(rank)
			game.moves[i].αβ = game.tt.Interval(rank).Reverse().String()

			// ⇢ δν
			game.moves[i].δν = game.moves[i].position.MovesInHand() - game.moves[i].position.Reverse().MovesInHand()

			// ⇢ check / both sides
			// i = attacking move
			// j = captured position
			if i < mech.NORTHLEFT {
				// S to move, affects N side
				for j := mech.NORTHLEFT; j <= mech.NORTHRIGHT; j++ {
					if game.south.position.Board[j] > 0 && game.moves[i].position.Reverse().Board[j] == 0 {
						game.moves[i].attack = true
						game.moves[j].check = true
					}
				}

			} else {
				// N to move, affects S side
				for j := mech.SOUTHLEFT; j <= mech.SOUTHRIGHT; j++ {
					if game.south.position.Board[j] > 0 && game.moves[i].position.Board[j] == 0 {
						game.moves[i].attack = true
						game.moves[j].check = true
					}
				}
			}
		}

		// ⇢ changed
		game.moves[i].changed = ow.Max(0, game.south.position.Board[i]-game.previous.position.Board[i])
	}

	ow.Log(game)

	fmt.Println(game.tt)

	return game
}
