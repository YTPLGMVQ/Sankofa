# Sankofa 

**Sankofa**  is an application for the analysis of Oware games.
Sankofa is the bird that looks back into the past in order to understand the future.
Our goal is to enable players to recognize and to learn from their mistakes and to try alternative strategies.

The application is a Web server and the user interface is presented as a Web page at 'http://localhost:10000'.
An execution trace is displayed in the shell. Functionalitiy:
* the user plays both sides in order to explore a game's outcome
* all legal moves are playable and their evaluations are shown
* this application is not an automated opponent
* hover the cursor over GUI elements for hints
* access and display any previous position or any position in the proposed continuation
* build any legal position by adding or removing stones and start from there

# Build and Run

* you need Golang to build this application
* initialize and build: 'go mod init sankofa && go mod tidy && go install ./...'
* optional: run '~/go/bin/retrograde' to build a small end-game database
* run: '~/go/bin/sankofa -h'
* open 'http://localhost:10000' in a Web browser with CSS and SVG capabilities

# Oware

The following variant of the Oware rules apply:
* Grand slams are allowed, but do not capture anything.
* The first player to capture 25 seeds instantly wins the game.
* Starved positions where no feeding (forced move) is possible end the game.
* A repeated position (the first cycle) ends the game.
* When the game ends, each player takes the seeds on her side of the board.

# Algorithm

**Sankofa** provides a MiniMax evaluation of Oware positions featuring:
* iterative deepener
* transposition table (thread cooperation; killer moves)
* parallel aspiration on discrete quartiles
* negamax
* fail-soft α—β pruning
* database with retrograde analysis (α—β leaves)
* simple score heuristic (α—β leaves not in the database)

**Retrograde** builds an end-game database:
* successively processes "levels" with a given number of stones, from zero upwards
* uses the simplified Awari rules which lead to slightly inaccurate Oware position evaluations
* does not filter out the unreachable positions but processes them like the rest
* iterates through a level until no new positions may be evaluated.
  Revisiting some positions could improve the level's evaluation.
* takes a huge amount of time to process the positions with many stones: useable only for end-games
* we recommend to evalute strongly connected components and the end-game up to level 12.

# License

MIT
