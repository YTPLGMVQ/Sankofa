package main

import (
	"sankofa/db"
	"sankofa/mech"
	"sankofa/ow"
)

////////////////////////////////////////////////////////////////
// WORKER BEE
////////////////////////////////////////////////////////////////

// Retrograde score computation.
// Saves a score if it is sure about it (no revisions).
func Visit(rank int64) {
	score, ini := db.GetScore(rank)
	ow.Log("rank:", rank, "score:", score)

	position := mech.Unrank(rank)
	if position.IsStarved() {
		split := position.Split()
		ow.Log(rank, "starved")
		if score != split {
			db.SetScore(rank, position.Split())
			incCounter()
		}
		return
	}

	legalMoves := position.LegalMoves()
	ow.Log(legalMoves)

	// search for the best attainable score.
	// start with the worst.
	max := ow.MININT8

	// find move with max. score
	var found bool
	for move, nxRank := range legalMoves.Next {
		nxScore, ini := db.GetScore(nxRank)
		if ini {
			found = true
			max = ow.Max(max, legalMoves.Score[move]-nxScore)
			ow.Log("successor: rank:", nxRank, "capture:", legalMoves.Score[move], "nxScore:", nxScore)
		} else {
			max = ow.Max(max, legalMoves.Score[move])
			ow.Log("successor: rank:", nxRank, "not initialized: zero")
		}
		ow.Log("successor: rank:", nxRank, "move:", mech.MoveToString(move), "captures:", legalMoves.Score[move], "- next score:", nxScore)
	}

	// count changed items
	if found && max != score {
		level := ow.Level(rank)
		if max < -level || max > level {
			ow.Panic("score:", max, "out of level:", level)
		}
		ow.Log("save: rank:", rank, "position:", position, "score ⇢", max)
		db.SetScore(rank, max)
	}

	// only count nodes that become initialized
	// ignore flip-flops
	if found && !ini {
		incCounter()
	}
}
