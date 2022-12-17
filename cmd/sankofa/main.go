// Oware web REST server
package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"sankofa/db"
	"sankofa/html"
	"sankofa/ow"
)

func main() {
	ow.Verbose = true

	// proper usage message
	flag.Usage = func() {
		flag.CommandLine.SetOutput(os.Stdout)
		fmt.Fprintln(os.Stdout, `SANKOFA is the bird that looks back into the past in order to understand the future.
This application is for the analysis of Oware games.
Our goal is to enable players to recognize and to learn from their mistakes and to try alternative strategies.
The following variant of the Oware rules apply:
* Grand slams are allowed, but do not capture anything.
* The first player to capture 25 seeds instantly wins the game.
* Starved positions where no feeding (forced move) is possible end the game.
* A repeated position (the first cycle) ends the game.
* When the game ends, each player takes the seeds on her side of the board.
SANKOFA provides a MiniMax evaluation of Oware positions featuring:
* iterative deepener
* transposition table (thread cooperation; killer moves)
* parallel aspiration on discrete quartiles
* negamax
* fail-soft α—β pruning
* simple score heuristic
CAVEATS
* MiniMax adds a heuristic value for the deepest position; the game continuation does not.
* MiniMax may end early with a saved score from other threads, leading to a truncated game continuation.
Copyright ©2019-2022 Carlo Monte.
................................................................................`)
		fmt.Fprintf(os.Stdout, "%s: start a Web server that shows an interactive Oware board\n", os.Args[0])
		flag.PrintDefaults()
	}

	// command line
	var ipPort string
	flag.StringVar(&db.FileName, "d", db.FileName, "database file")
	flag.IntVar(&html.Goroutines, "g", 5, "number of parallel Go-routines")
	flag.StringVar(&ipPort, "i", "localhost:10000", "listen on IP:Port")
	flag.Float64Var(&html.DurationLimit, "t", 1, "response time  in seconds")
	flag.BoolVar(&ow.Verbose, "v", false, "verbose")
	flag.Parse()

	// informative output
	fmt.Println("................................................................................")
	fmt.Println("run with -h for HELP")

	// open/create DB file
	db.Open()

	// start web server
	http.HandleFunc("/", html.PlayHandler)
	ow.Log("starting web server on:", ipPort)
	fmt.Println("point your browser to:", "http://"+ipPort)

	ow.Panic(http.ListenAndServe(ipPort, nil))
}
