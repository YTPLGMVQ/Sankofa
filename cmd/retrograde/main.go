package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime/pprof"
	"sankofa/db"
	"sankofa/ow"
	"sankofa/scc"
	"strconv"
	"syscall"
	"time"
)

// profile
const PROFILE = "profile.out"

// cancellation signal
var cancel chan struct{}

func main() {
	// prepare for shutdown
	defer func() {
		shutDown()
	}()

	// channel for cancel signal
	cancel = make(chan struct{})

	// catch signals
	var channel = make(chan os.Signal, 1)
	signal.Notify(channel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	go func() {
		signal := <-channel
		fmt.Println("received signal:", signal)
		// signal process termination
		close(cancel)
	}()

	// proper usage message
	flag.Usage = func() {
		//		fmt.Fprintf(os.Stderr, "%s\nRetrograde analysis with the Awari rules.\nIgnore oscillating scores.\n", os.Args[0])
		fmt.Fprintln(os.Stdout, `RETROGRADE analysis with the Awari rule set.
* When a position is repeated (cycle), Oware gives each player the seeds on her side.
* Awari gives each player half of the seeds on the board (down to ½ a stone); repeated positions have an effective Awari score of 0.
* Database scores are used in SANKOFA for the leaves of the α—β search tree. The Oware-Awari inaccuracy is thus tolerated.
* The database is built incrementally, layer for layer, starting with the empty board.
* Strongly connected components in lower layers discovered using Tarjan's algorithm (single threaded, in-memory).
* SCC members belong to cycles. Their scores are initialized accordingly.
* A layer is incrementally processed, until there are no NEW nodes to score.
* Additional iterations may further improve the score accuracy, but are avoided for performance reasons.
* A partial database can be used by SANKOFA.
* The complete databse (1.1TB) requires a very long processing time.
* Unreachable nodes are scored as well. No significant performance improvement is expected by avoiding them.
Copyright ©2019-2023 Carlo Monte.
................................................................................`)
		fmt.Fprintf(os.Stdout, "%s: start a Web server that shows an interactive Oware board\n", os.Args[0])
		flag.PrintDefaults()
	}

	// flags
	f := int(-1)       // from level
	t := int(12)       // to level: lowest useable level
	s := int(12)       // maximum level for SCC initialization
	var profiling bool // enable profiling
	//
	flag.StringVar(&db.FileName, "d", db.FileName, "database file")
	flag.IntVar(&goroutines, "g", 8, "number of parallel Go-routines")
	flag.BoolVar(&profiling, "p", false, "enable CPU profiling")
	flag.IntVar(&f, "f", f, "from level; overrides the saved checkpoint when >0")
	flag.IntVar(&s, "s", s, "maximum level where to initialize strongly connected component member's scores")
	flag.IntVar(&t, "t", t, "to level")
	flag.BoolVar(&ow.Verbose, "v", false, "be chatty")
	flag.Parse()

	// open/create DB file
	db.Open()

	// start level
	checkpoint := db.GetState()
	fmt.Println("checkpoint:", checkpoint)
	fromLevel := checkpoint
	if f >= 0 {
		fromLevel = int8(f)
	}
	toLevel := int8(t)
	fmt.Println("levels: from:", fromLevel, "to:", toLevel)

	var proFile os.File
	// start profiling
	if profiling {
		ow.Log("profile/cpu")
		proFile, err := os.Create(PROFILE)
		ow.Check(err)
		ow.Log("fd:", proFile.Fd())
		err = pprof.StartCPUProfile(proFile)
		ow.Check(err)
	}

levels:
	// retrograde analysis
	for l := ow.Max(fromLevel, 0); l <= ow.Max(toLevel, 0); l++ {
		if l == 47 {
			continue
		}
		db.SetState(l)
		levelTimeStamp := time.Now().UTC().UnixNano()

		var fromRank, toRank int64
		if l > 0 {
			fromRank = ow.LevelUpperLimits[l-1] + 1
			toRank = ow.LevelUpperLimits[l]
		}
		fmt.Println("................................................................................")
		fmt.Println(l, "stones:", fromRank, "⇢", toRank, "=", toRank-fromRank, "ranks")

		var it int

		if l <= int8(s) {
			scc := scc.Tarjan(l)
			for _, rank := range scc {
				ow.Log("scc: rank:", rank)
				db.SetScore(rank, 0)
			}
			fmt.Println(len(scc), "strongly connected component member node's scores initialized")
		}

		for {
			it++
			iterationTimeStamp := time.Now().UTC().UnixNano()

			feed := startWorkers(Visit)
			for r := toRank; r >= fromRank; r-- {
				ow.Log("rank:", r)
				feed <- r

				// break if requested
				select {
				case <-cancel:
					ow.Log("canceled")
					break levels
				default:
					ow.Log("keep on")
				}
			}

			stopWorkers(feed)

			fmt.Println("iteration:", it, ":", counter(), "scored,",
				strconv.FormatFloat(ow.GIGA64F*float64(toRank-fromRank)/float64(time.Now().UTC().UnixNano()-iterationTimeStamp), 'f', 0, 64), "ranks/second")

			if counter() == 0 {
				ow.Log("break")
				break
			}
		}
		fmt.Println("average", strconv.FormatFloat(ow.GIGA64F*float64(toRank-fromRank)/float64(time.Now().UTC().UnixNano()-levelTimeStamp), 'f', 0, 64), "ranks/second")

	}

	// goodbye
	fmt.Println("DONE")

	// profile
	if profiling {
		ow.Log("stop profiling")
		pprof.StopCPUProfile()
		proFile.Close()
	}

	// shut down deferred to here
}

func shutDown() {
	ow.Log("shut down")
	// database
	db.Close()
	ow.Log("END")
}
