# Sankofa 

**Sankofa**  is an application for the analysis of Oware games.
Sankofa is the bird that looks back into the past in order to understand the future.
Our goal is to enable players to recognize and to learn from their mistakes and to try alternative strategies.

The application is a Web server and the user interface is presented as a Web page at 'http://localhost:10000'.
An execution trace is displayed in the shell. Functionalitiy:
* the user plays both sides in order to explore a game's outcome
* this application is not an automated opponent
* all legal moves are playable and their evaluations are shown
* hover the cursor over GUI elements for hints
* access and display any previous position or any position in the proposed continuation
* build any legal position by adding or removing stones and start from there

# Build and Run

* you need Golang to build this application
* to build and install into ~/go/bin, run:
	- 'go mod init sankofa'
	- 'go mod tidy'
	- 'go install ./...'
* optional: run '~/go/bin/retrograde' to build a small end-game database
* run '~/go/bin/sankofa'
* open 'http://localhost:10000' in a Web browser with CSS and SVG capabilities
* '-h' parameter for help on the command line

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

# End-Game Database

**Retrograde** is a helper application that builds a database with the retrograde analysis of the game.
It starts with the end-games and incrementally works towards the starting position.
The retrograde analysis takes some shortcuts for the sake of performance:
* uses the simplified Awari rules which leads to slightly inaccurate scores
* does not filter out the unreachable positions but processes them like the rest
* stops once it has reached all possible positions for a level. Revisiting some could improve the evaluation.
* takes a huge amount of time to process the positions with many stones: useable only for end-games

# License

MIT
