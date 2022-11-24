// methods for tracing the recursion's progresss

package minimax

import (
	"fmt"
	"sankofa/mech"
	"sankofa/ow"
)

// trace single-threaded negamax?
var Trace bool

func scoreString(len, score int64) string {
	switch {
	case score == ow.MININT || score == ow.MAXINT:
		return ow.Thousands(ow.MININT)
	case ow.Even(len):
		return ow.Thousands(score)
	default:
		return ow.Thousands(-score)
	}
}

// normalize α—β interval w.r.t. odd/even ply depth
func αβString(len, α, β int64) string {
	switch {
	case ow.Even(len):
		return ow.Thousands(α, β)
	default:
		return ow.Thousands(-β, -α)
	}
}

// show how a score is built from the parts left and right of the cursor
func joinScore(game *mech.Game, score int64) string {

	// cumulated score up to the cursor
	join := scoreString(game.Cursor-1, game.BeforeCurrent().Score())

	// negamax score after the cursor
	if (ow.Even(game.Cursor) && score >= 0) || (ow.Odd(game.Cursor) && score <= 0) {
		join += "+"
	}
	join += scoreString(game.Cursor, score)

	// game score: on the last position
	join += "⇢"
	join += scoreString(int64(len(game.Moves)), game.Last().Score())

	return join
}

// uniform trace messages
func trace(what string, game *mech.Game, score, α, β int64, legalMoves *mech.LegalMoves) {
	if Trace {

		// off by 1 beacause of the initial position
		var board, side string
		if ow.Even(game.Cursor) {
			board = game.Current().Board.String()
			side = "♙"
		} else {
			board = game.Current().Reverse().Board.String()
			side = "♟︎"
		}

		// output
		fmt.Println(fmt.Sprintf("%v %v %v %v, %v, %v, %v",
			side,
			what,
			game,
			joinScore(game, score),
			αβString(game.Cursor, α, β),
			board,
			legalMoves,
		))
	}
}
