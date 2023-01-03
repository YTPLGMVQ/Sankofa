package minimax

// transposition table for oware scores, stored as *Interval structs plus
// additional fuctionality needed in the same scope:
//   * goroutine synchronization
//   * memoization of *mechanics.Game and *mechanics.LegalMoves
//   * *TT.LegalMoves() and not *TT.setScore() initializes the *Interval stored in *TT
//   * counters

import (
	"fmt"
	"sankofa/mech"
	"sankofa/ow"
	"sort"
	"strconv"
	"sync"
	"time"
)

////////////////////////////////////////////////////////////////
// DATA
////////////////////////////////////////////////////////////////

// initial capacity of the transposition table
const TT_CAP = 4096

// dump transposition table when done
var Dump bool

// thread-safe, global transposition table for negamax with additional functionality
//
// DESIGN we implement method call APIs with mutexes and batch processing with channels
type TT struct {
	// the transposition table: *TT.tt[rank] ⇢ [α, β], i.e., a score range
	tt map[int64]*Interval
	// from previous iteration of the deepener
	old *TT

	// memoization of CPU-intensive evaluations
	positions   map[int64]*mech.Position
	legalMoves  map[int64]*mech.LegalMoves
	movesInHand map[int64]int8

	// timers
	globalTimeStamp    int64
	iterationTimeStamp int64

	// input game followed by α—β continuation
	game  *mech.Game
	found bool

	// synchronization
	mutex     sync.RWMutex    // data access
	waitGroup *sync.WaitGroup // everybddy is done (parallel aspiration)
	cntWg     int             // wait group counter

	// timed out
	cancelIteration chan struct{} // cancellation signal for an iteration
	cancelDeepener  chan struct{} // cancellation signal for the iterative deepener

	// counters
	base           int // distance from bottom; ideally it should be 0
	depth          int // depth of the current deepener iteration
	visited        int // visied nodes: total
	cumVisited     int // cumulative visited nodes (all iterations)
	cntTt          int // retrievals from the transposition table
	cntLegalMoves  int // retrievals from the legalMoves table
	cntMovesInHand int // retrievals from the movesInHand table
	cutOff         int // number of cutoffs
	over           int // game over (won, starved or cycle)
	database       int // number of bottom-level nodes using scores from the database
	heuristic      int // number of bottom-level nodes evaluated using the heuristic
	killed         int // number of interrupted goroutines
}

////////////////////////////////////////////////////////////////
// TRANSPOSITION TABLE
////////////////////////////////////////////////////////////////

// the numer of stones on the board sets the α—β bandwidth
func NewTT(game *mech.Game) *TT {
	tt := new(TT)
	tt.tt = make(map[int64]*Interval, TT_CAP)
	tt.positions = make(map[int64]*mech.Position, TT_CAP)
	tt.legalMoves = make(map[int64]*mech.LegalMoves, TT_CAP)
	tt.movesInHand = make(map[int64]int8, TT_CAP)

	tt.globalTimeStamp = time.Now().UTC().UnixNano()
	tt.iterationTimeStamp = time.Now().UTC().UnixNano()

	tt.game = game
	tt.waitGroup = new(sync.WaitGroup)

	tt.cancelIteration = make(chan struct{})
	tt.cancelDeepener = make(chan struct{})

	tt.depth = ow.MININT // monotonously incrementing
	tt.base = ow.MAXINT  // monotonously decrementing

	return tt
}

func (tt *TT) String() string {
	tt.mutex.RLock()
	defer tt.mutex.RUnlock()

	r := tt.game.String() +
		" | visited: " + ow.Thousands(tt.visited) +
		", Σ: " + ow.Thousands(tt.cumVisited) +
		", database: " + ow.Thousands(tt.database) +
		", heuristic: " + ow.Thousands(tt.heuristic) +
		", game-over: " + ow.Thousands(tt.over) +
		" | depth: " + ow.Thousands(tt.depth-tt.base) +
		", killed: " + ow.Thousands(tt.killed) +
		" | TT: size: " + ow.Thousands(len(tt.tt)) +
		", #rd: " + ow.Thousands(tt.cntTt) +
		" | LEGAL: size: " + ow.Thousands(len(tt.legalMoves)) +
		", #rd: " + ow.Thousands(tt.cntLegalMoves) +
		" | Δν: size: " + ow.Thousands(len(tt.movesInHand)) +
		", #rd: " + ow.Thousands(tt.cntMovesInHand) +
		" | " + strconv.FormatFloat((float64(time.Now().UTC().UnixNano())-float64(tt.iterationTimeStamp))/ow.GIGA64F, 'f', 2, 64) + " sec."

	return r
}

// no-lock: set initial/saved game
func (tt *TT) _setGame(game *mech.Game) *TT {
	tt.game = game.Clone()
	return tt
}

// set initial/saved game
func (tt *TT) setGame(game *mech.Game) *TT {
	tt.mutex.Lock()
	defer tt.mutex.Unlock()

	return tt._setGame(game)
}

// return the best game found
func (tt *TT) Game() *mech.Game {
	tt.mutex.RLock()
	defer tt.mutex.RUnlock()

	r := tt.game.Clone()
	return r
}

// return the global time stamp
func (tt *TT) Begin() int64 {
	tt.mutex.Lock()
	defer tt.mutex.Unlock()

	r := tt.globalTimeStamp
	return r
}

// add to a given partial score more partial information from a parallel aspiration thread
func (tt *TT) setScore(rank int64, α, β, score int8) *TT {
	ow.Log("rank:", rank, ", α:", α, ", score:", score, ", β:", β)
	if β < α {
		ow.Panic("rank:", rank, "α=", α, " > β=", β)
	}

	// captures reduce the level and narrow the [α, β] interval
	level := ow.Level(rank)
	// even limits have limits
	if score < -level || score > level {
		ow.Panic("score:", score, "out of level:", level)
	}

	tt.mutex.Lock()
	defer tt.mutex.Unlock()

	// get it if available
	old, ok := tt.tt[rank]
	// if not, create one
	if !ok {
		// or create it if not
		old = NewInterval(rank, ow.MININT8, ow.MAXINT8)
	} else if old.IsFinal() {
		return tt
	}

	var new *Interval
	switch {
	case score <= α:
		// interpretation of fail-soft negamax
		ow.Log("score <= α")
		new = NewInterval(rank, -level, score)
	case score >= β:
		// interpretation of fail-soft negamax
		ow.Log("score >= β")
		new = NewInterval(rank, score, level)
	default:
		ow.Log("α < score < β")
		new = NewInterval(rank, score, score)
	}

	if old.Disjoint(new) {
		ow.Log("rank:", rank, "disjoint")
		tt.tt[rank] = new
	} else {
		tt.tt[rank] = old.Intersect(new)
	}

	ow.Log("rank:", rank, ": old:", old, "⋂ new:", new, "⇢", tt.tt[rank])
	return tt
}

// no-lock: do we have a record for the given score?
func (tt *TT) _known(rank int64) bool {
	_, ok := tt.tt[rank]
	return ok
}

// do we have a record for the given score?
func (tt *TT) Known(rank int64) bool {
	tt.mutex.Lock()
	defer tt.mutex.Unlock()

	return tt._known(rank)
}

// no-lock: find score interval in transposition table
func (tt *TT) _interval(rank int64) *Interval {
	interval, ok := tt.tt[rank]
	if ok {
		tt.cntTt += 1
		return interval
	} else {
		return nil
	}
}

// find score interval in transposition table
func (tt *TT) Interval(rank int64) *Interval {
	tt.mutex.Lock()
	defer tt.mutex.Unlock()

	return tt._interval(rank)
}

// dump the contents of TT to STDOUT for tracing purposes
func (tt *TT) Dump() *TT {
	fmt.Println(tt)

	tt.mutex.RLock()
	defer tt.mutex.RUnlock()

	// extract keys
	// must do some anaerobics (int⇢int⇢int64) because of sortSlice() accepting only ints
	ranks := make([]int, 0)
	for rank := range tt.tt {
		ranks = append(ranks, int(rank))
	}

	// sort keys
	sort.Slice(ranks, func(a, b int) bool {
		return ranks[a] < ranks[b]
	})

	// output intervals
	for _, r := range ranks {
		fmt.Println(r, "⇢", tt.tt[int64(r)])
	}

	return tt
}

////////////////////////////////////////////////////////////////
// SYNCHRONIZATION
////////////////////////////////////////////////////////////////

// increment waiting group and ancillary counter
func (tt *TT) Inc() *TT {
	tt.waitGroup.Add(1)

	tt.mutex.Lock()
	tt.cntWg += 1
	tt.mutex.Unlock()

	return tt
}

// decrement waiting group and ancillary counter
func (tt *TT) Dec() *TT {
	tt.waitGroup.Done()

	tt.mutex.Lock()
	tt.cntWg -= 1
	tt.mutex.Unlock()

	return tt
}

// wait for remaining goroutines to finish
func (tt *TT) Wait() *TT {
	tt.waitGroup.Wait()

	tt.mutex.Lock()
	tt.cntWg = 0
	tt.mutex.Unlock()

	return tt
}

// no-lock: request everyone to stop working
func (tt *TT) _abortIteration() *TT {
	select {
	case <-tt.cancelIteration:
		// BENIGN
		// this might happen, since the sequence IsDone()...Done() is not continuosly locked.
		ow.Log("too late")
	default:
		close(tt.cancelIteration)
		// still running workers will be killed
		tt.killed = tt.cntWg
	}

	return tt
}

// request everyone to stop working
func (tt *TT) AbortIteration() *TT {
	tt.mutex.Lock()
	defer tt.mutex.Unlock()

	return tt._abortIteration()
}

// no-lock: are we done here?
func (tt *TT) _iterationAborted() bool {
	select {
	case <-tt.cancelIteration:
		return true
	default:
		return false
	}
}

// are we done here?
func (tt *TT) IterationAborted() bool {
	tt.mutex.RLock()
	defer tt.mutex.RUnlock()

	return tt._iterationAborted()
}

func (tt *TT) AbortDeepener() *TT {
	tt.mutex.Lock()
	defer tt.mutex.Unlock()

	select {
	case <-tt.cancelDeepener:
		// BENIGN
		// this might happen, since the sequence IsDone()...Done() is not continuosly locked.
		ow.Log("too late")
	default:
		close(tt.cancelDeepener)
	}

	return tt
}

func (tt *TT) DeepenerAborted() bool {
	tt.mutex.RLock()
	defer tt.mutex.RUnlock()

	select {
	case <-tt.cancelDeepener:
		return true
	default:
		return false
	}
}

// restart synchronization for another deepener iteration
func (tt *TT) Restart() *TT {
	ow.Log("restart")

	r := NewTT(tt.game)

	tt.mutex.Lock()
	defer tt.mutex.Unlock()

	r.old = tt
	r.cancelDeepener = tt.cancelDeepener
	r.depth = tt.depth
	r.legalMoves = tt.legalMoves
	r.movesInHand = tt.movesInHand
	r.globalTimeStamp = tt.globalTimeStamp
	r.cumVisited = tt.cumVisited

	return r
}

////////////////////////////////////////////////////////////////
// COUNTERS
////////////////////////////////////////////////////////////////

// update depth
func (tt *TT) setDepth(depth int) *TT {
	tt.mutex.Lock()
	defer tt.mutex.Unlock()

	tt.depth = ow.Max(tt.depth, depth)
	return tt
}

// update base
func (tt *TT) setBase(depth int) *TT {
	tt.mutex.Lock()
	defer tt.mutex.Unlock()

	tt.base = ow.Min(tt.base, depth)
	return tt
}

// get current iteration depth
func (tt *TT) Depth() int {
	tt.mutex.RLock()
	defer tt.mutex.RUnlock()

	aux := tt.depth
	return aux
}

// get distance between current iteration depth and actually reached depth.
func (tt *TT) Base() int {
	tt.mutex.RLock()
	defer tt.mutex.RUnlock()

	aux := tt.base
	return aux
}

func (tt *TT) incVisited() *TT {
	tt.mutex.Lock()
	defer tt.mutex.Unlock()

	tt.visited++
	tt.cumVisited++
	return tt
}

func (tt *TT) incOver() *TT {
	tt.mutex.Lock()
	defer tt.mutex.Unlock()

	tt.over++
	return tt
}

func (tt *TT) incDatabase() *TT {
	tt.mutex.Lock()
	defer tt.mutex.Unlock()

	tt.database++
	return tt
}

func (tt *TT) incHeuristic() *TT {
	tt.mutex.Lock()
	defer tt.mutex.Unlock()

	tt.heuristic++
	return tt
}

func (tt *TT) incCutOff() *TT {
	tt.mutex.Lock()
	defer tt.mutex.Unlock()

	tt.cutOff++
	return tt
}
