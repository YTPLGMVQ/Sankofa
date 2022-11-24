package minimax

import (
	"fmt"
	"sankofa/ow"
	"time"
)

// go routine that kills the iterative deepener when the timer expires
func (tt *TT) Terminator(limit float64) {
	ow.Log("terminate after:", limit)
	time.Sleep(time.Duration(limit * ow.GIGA))

	// transaction
	tt.AbortDeepener()
	fmt.Println("watchdog")
}

// iterative depender with duration limit; result in transposition table
func (tt *TT) Explore(goroutines int64, limit float64) *TT {
	// to restore initial conditions
	game := tt.Game()

	// the number of stones on the board limits the interval and is used in computing the SD
	level := ow.Level(tt.Game().Current().Rank())

	// -t Trace disables parallelism
	if Trace {
		goroutines = ow.ONE
		ow.Log("trace: no parallism:", goroutines, "goroutines")
	}

	// make copies, since Alpha and Beta will be used in further iterations
	a := ow.Max(Alpha, -level)
	b := ow.Min(Beta, level)

	// split in so many intervals as many processor cores are available
	intervals := Quartiles(a, b, level, goroutines)

	// iterative deepening
	go tt.Terminator(limit)
	for depth := ow.TWO; !tt.DeepenerAborted(); depth += 1 {
		// need transaction here, not to lose .abort!!!
		tt = tt.Restart().setGame(game).Aspiration(intervals, depth)
		ow.Log(tt.Game(), "depth:", depth)

		if tt.Base() > 0 {
			fmt.Println("base:", tt.Base(), "above bottom: break")
			ow.Log("base:", tt.Base(), "depth:", depth)
			break
		}

		if tt.DeepenerAborted() {
			ow.Log("watchdog: break and discard unfinished iteration")
			tt = tt.old
			break
		}
		fmt.Println(tt)
	}
	ow.Log("transposition:", tt)

	return tt
}
