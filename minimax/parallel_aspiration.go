package minimax

import (
	"fmt"
	"sankofa/mech"
	"sankofa/ow"
	"strconv"
	"time"
)

// infinite interval
var Alpha = ow.MININT8
var Beta = ow.MAXINT8

// if all the first-level successors of the root node have scores.
// WARNING: must be called within a tt.mutex.Lock() transaction.
func (tt *TT) Finished() bool {
	// tt.Game().Current() must have a known score
	rank := tt.game.Current().Rank()
	_, ok := tt.tt[rank]
	if !ok || !tt.tt[rank].IsFinal() {
		return false
	}

	// tt.Game().Current().LegalMoves() must have known scores
	r := true // search for refutation
	for k, v := range tt.game.Current().LegalMoves().Next {
		// the rank is finished: it has a final score, not an interval
		// cannot use tt.Known() since this method locks
		_, ok := tt.tt[v]
		if !ok || !tt.tt[v].IsFinal() {
			ow.Log("rank:", tt.game.Current(), "move:", mech.MoveToString(k), "next:", v, "not finished")
			r = false
			break
		}
	}

	return r
}

// goroutine that wraps a negamax call
//
// DESIGN we implement batch processing with channels and method call APIs with mutexes
func (tt *TT) Worker(α, β int8, depth int) {
	timeStamp := time.Now().UTC().UnixNano()
	ow.Log("α:", α, ", β:", β, ", depth:", depth)

	score, game := tt.NegaMax(tt.Game(), α, β, depth)
	game.Cursor = tt.Game().Cursor

	duration := (float64(time.Now().UTC().UnixNano()) - float64(timeStamp)) / ow.GIGA64F
	ow.Log(game, ", [", α, "..", β, "]",
		", duration:", strconv.FormatFloat(duration, 'f', 2, 64))

	fmt.Println(ow.Thousands(α, β), "⇢", tt.Interval(tt.Game().Current().Rank()), game)

	// set game if the current go routine found the leading score.
	// WARNING: the operation is atomic: use lock-less method variants
	{
		tt.mutex.Lock()
		// save the game if a score value (not interval) was found, that matches the negamax return value
		if !tt.found {
			rank := game.Current().Rank()
			if tt._known(rank) && tt._interval(rank).IsFinal() && tt._interval(rank).Score() == score {
				fmt.Println("continuation ⇢", game, "score±heuristic:", score)
				tt._setGame(game)
				tt.found = true
			}
		}
		// abort the iteration if all 1st-level scores are known
		if !tt._iterationAborted() && tt.Finished() {
			tt._abortIteration()
			ow.Log("rank:", tt.game.Current().Rank(), ": finished, abort iteration")
			fmt.Println("finished ⇢ cancel the other goroutines")
		}
		tt.mutex.Unlock()
	}

	// signal finished event to the WaitGroup
	tt.Dec()
}

// Baudet's parallel aspiration search, a divide and conquer algorithm
func (tt *TT) Aspiration(intervals Intervals, depth int) *TT {
	ow.Log(tt.Game(), intervals, "depth:", depth)

	// start one worker per interval
	for _, interval := range intervals {
		tt.Inc()
		go tt.Worker(interval[0], interval[1], depth)
	}

	// finish goroutines
	ow.Log("waiting for the parallel aspiration to end")
	tt.Wait()
	fmt.Println("all go routines done")

	// done
	return tt
}
