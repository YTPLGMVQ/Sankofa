// a minimax evaluation of Oware positions
//
// # FEATURES
//
// Algorithm:
//   - parallel aspiration on discrete quartiles
//   - iterative deepener
//   - negamax
//   - transposition table (previous iteration) used for move ordering
//   - fail-soft α—β pruning
//   - simple heuristic score
//
// # DESIGN, TACTICS AND HACKS
//
// We use:
//   - asynchronous batch processing uses channels
//   - synchronous method calls use mutexes
//   - negamax returns a score value ± the heuristic, subsequently stored in the transposition table
//   - the score stored in *Game steps excludes the heuristics.
//   - negamax returns the *Game steps corresponding to the found score
//   - parallel aspiration has a bit transaction block that uses the non-locking inner methods
//   - WARNING negamax may return an empty or a truncated continuation when it returns a score from *TT
//   - a DB read operation is about one order of magnitude slower than the trivial heuristic we employ.
//     DB scores are only used for the leaves, since we are interested in the game continuation discovered by α—β.
//     a good compromise is to limit the database to the lower end-game levels e.g., up to 24.
//
// # CALL STACK
//
// Explore ⇢ Aspiration ⇢ Worker ⇢ NegaMax
package minimax

import (
	"fmt"
	"sankofa/db"
	"sankofa/mech"
	"sankofa/ow"
)

// complete search tree; disables cutoff
var Complete bool

// fail-soft NegaMax with α—β pruning and killer-move heuristic;
// modifies the input *TT and returns the game continuation;
// score and verdict are used internally.
func (tt *TT) NegaMax(game *mech.Game, α, β int8, depth int) (score, verdict int8, continuation *mech.Game) {
	// sanity check
	if β < α {
		ow.Panic("α=", α, "> β=", β, game)
	}

	// visit this node
	tt.incVisited()

	// update base and depth
	tt.setDepth(depth)
	tt.setBase(depth)
	ow.Log(game, "depth:", depth, "TT: base:", tt.Base(), "depth:", tt.Depth())

	// pre-compute some useful values
	position := game.Current()
	rank := position.Rank()
	legalMoves := tt.LegalMoves(rank) // for tracing
	verdict = game.Current().Verdict()

	ow.Log(game, "⇢visit: game:", game, "@", game.Cursor, "position:", position, "legal moves:", legalMoves, "α:", α, "β:", β, "depth:", depth)

	trace(">>", game, game.Last().Score()+game.BeforeLast().Score(), α, β, legalMoves)

	////////////////////////////////////////////////////////////////
	// BASE CONDITIONS
	////////////////////////////////////////////////////////////////

	switch {
	case game.Current().FinalScore():
		// the final score has already been achieved
		score := ow.ZERO8

		ow.Log(game, "⇠over:", game, "|", game.Current().Board, "score:", score, "verdict:", verdict)
		tt.incOver()
		trace("<< over", game, score, α, β, legalMoves)

		// save score to the transposition table
		tt.save(rank, α, β, score, verdict)

		return score, verdict, game
	case position.Starved():
		// starved (terminal)
		// cannot use game.Last().Score(), as this is not set if there is no preceding move
		score := game.Current().SaveSplit()
		verdict = game.Current().Verdict() // score changed by operation above

		ow.Log(game, "⇠starved:", game, "|", game.Current().Board, "score:", score, "verdict:", verdict)
		tt.incOver()
		trace("<< starved", game, score, α, β, legalMoves)

		// save score to the transposition table
		tt.save(rank, α, β, score, verdict)

		return score, mech.LOSS, game
	case game.Cycle():
		// cycle (terminal)
		score := game.Capture()
		ow.Log(game, "⇠cycle:", game, "|", game.Current().Board, "score:", score, "verdict:", verdict)
		tt.incOver()
		trace("<< cycle", game, score, α, β, legalMoves)

		// save score to the transposition table
		tt.save(rank, α, β, score, verdict)

		return score, verdict, game
	case tt.Known(rank) && tt.Interval(rank).Scored():
		i := tt.Interval(rank)
		score, verdict := i.Score(), i.Verdict()
		// counter incremented by *TT.Interval() call
		ow.Log(game, "⇠TT:", game, "|", game.Current().Board, "score:", score, "verdict:", verdict)
		return score, verdict, game
	case depth == 0:
		// reached recursion depth limit
		// search for a score in the database
		score, ini := db.GetScore(rank)
		if ini {
			ow.Log(game, "⇠bottom+database:", game, "|", game.Current().Board, "score:", score, "verdict:", verdict)
			tt.incDatabase()
		} else {
			// evaluate score using a heuristic
			score = game.Heuristic()
			ow.Log(game, "⇠bottom+heuristic:", game, "|", game.Current().Board, "score:", score, "verdict:", verdict)
			tt.incHeuristic()
		}
		trace("<< bottom", game, score, α, β, legalMoves)
		return score, verdict, game
	case tt.IterationAborted() || tt.DeepenerAborted():
		// stop processing
		score := game.Heuristic()
		ow.Log(game, "⇠cancelled:", game, "|", game.Current().Board, "score:", score, "verdict:", verdict)
		trace("<< done", game, score, α, β, legalMoves)
		return score, verdict, game
	}

	////////////////////////////////////////////////////////////////
	// RECURSE
	////////////////////////////////////////////////////////////////

	// "game" imutable
	bestGame := game.Clone()
	bestScore := ow.MININT8

	// side
	side := "♟︎"
	if ow.Even(game.Cursor) {
		side = "♙"
	}

	killerMoves := tt.KillerMoves(rank)
	for _, move := range killerMoves {
		ow.Log(game, "rank:", rank, "killer move:", mech.MoveToString(move), "⇢ successor:", legalMoves.Next[move], ", captures:", legalMoves.Score[move])
		ow.Log(game, "α:", α, ", best score:", bestScore, ", β:", β)

		var s, v int8
		var g *mech.Game

		// trim α—β to plausible score range for this level
		mv := game.Move(move)
		lv := ow.Level(mv.Current().Rank())
		a := ow.Max(-lv, ow.Min(lv, α))
		b := ow.Max(-lv, ow.Min(lv, β))

		if Complete || tt.Depth()-depth <= 1 {
			// ignore cuts when traversing the entire tree
			s, v, g = tt.NegaMax(mv, -b, -a, depth-1)
		} else {
			// according to Marsland: -ow.Max(α, bestScore)
			s, v, g = tt.NegaMax(mv, -b, -ow.Max(a, ow.Min(b, bestScore)), depth-1)
		}

		t := legalMoves.Score[move] - s // best score candidate
		if t > bestScore {
			if Trace {
				fmt.Println(fmt.Sprintf("%v || best %v %v⇢%v ∈? %v",
					side,
					bestGame,
					scoreString(g.Cursor, bestScore),
					scoreString(g.Cursor, t),
					αβString(g.Cursor, α, β),
				))
			}

			bestScore = t
			verdict = mech.ReverseVerdict(v)
			bestGame = g
			bestGame.Cursor = game.Cursor // restore cursor position
			ow.Log(game, "best score:", bestScore)

			// cut off only when not displaying complete tree and not at the first level
			if (verdict == mech.WIN || bestScore >= β) && !Complete {
				if tt.Depth()-depth > 1 {
					if Trace {
						// output
						if ow.Even(game.Cursor) {
							fmt.Println(side + " || cut " + bestGame.String() + " " + αβString(game.Cursor, α, β) + " <= " + scoreString(game.Cursor, bestScore))
						} else {
							fmt.Println(side + " || cut " + bestGame.String() + " " + scoreString(game.Cursor, bestScore) + " <= " + αβString(game.Cursor, α, β))
						}
					}
					tt.incCutOff()

					// "break" realizes the cut off
					break
				} else {
					if Trace {
						fmt.Println(side + " || no cut at the first level")
					}
				}
			}
		}
	}

	ow.Log(game, "⇠α—β: game:", game, "|", game.Current().Board, "score:", bestScore)

	// cannot happen, since a move is guaranteed to exist (starved check above)
	if bestScore == ow.MININT8 {
		ow.Panic("recursion on a final position")
	}

	// save score to the transposition table
	tt.save(rank, α, β, bestScore, verdict)

	trace("<< nmax", bestGame, bestScore, α, β, legalMoves)
	return bestScore, verdict, bestGame
}
