# Sankofa 

**Sankofa**  is an application for the analysis of Oware games.
Sankofa is the bird that looks back into the past in order to understand the future.
Our goal is to enable players to recognize and to learn from their mistakes and to try alternative strategies.

# Build

Run *make* in order to build and install the application.
You will need *Golang* and *make* installed.

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
* simple score heuristic

# License

MIT
