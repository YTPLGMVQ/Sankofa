# Sankofa 

**Sankofa**  is an application for the analysis of Oware games.
Sankofa is the bird that looks back into the past in order to understand the future.
Our goal is to enable players to recognize and to learn from their mistakes and to try alternative strategies.

**Retrograde** is a helper application that builds a database with the retrograde analysis of the game.
The retrograde analysis takes some shortcuts for the sake of performance:
* uses the simplified Awari rules which leads to slightly inaccurate scores
* does not filter for the unreachable positions but processes them like the rest
* stops once it has reached all possible positions for a level. Revisiting some could improve the evaluation.
* takes a huge amount of time to process the positions with many stones: useable only for end-games

# Build and Run

* you need Golang to build this application
* run 'go install ./...' to build and install into ~/go/bin
* optional: run '~/go/bin/retrograde' to build a small end-game database
* run '~/go/bin/sankofa'
* open 'http://localhost:10000' in a Web browser with CSS and SVG capabilities

# Oware

The following variant of the Oware rules apply:
* Grand slams are allowed, but do not capture anything.
* The first player to capture 25 seeds instantly wins the game.
* Starved positions where no feeding (forced move) is possible end the game.
* A repeated position (the first cycle) ends the game.
* When the game ends, each player takes the seeds on her side of the board.

# Algorithm

Sankofa provides a MiniMax evaluation of Oware positions featuring:
* iterative deepener
* transposition table (thread cooperation; killer moves)
* parallel aspiration on discrete quartiles
* negamax
* fail-soft α—β pruning
* database with retrograde analysis (α—β leaves)
* simple score heuristic (α—β leaves not in the database)

# License

MIT
