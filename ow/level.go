package ow

// conversions between rank and level (#stones on the board)

// number of positions for a given level (=nr. of stones)
var LevelUpperLimits [50]int64

// initialize table
func init() {
	LevelUpperLimits[0] = ZERO64
	for level := 1; level < 50; level++ {
		LevelUpperLimits[level] = LevelUpperLimits[level-1] + Repetitions(12, level)
	}
}

// number of stones (level) for a given rank
func Level(rank int64) int8 {
	if rank < 0 {
		Panic("negative rank:", rank)
	}
	for level := ZERO8; level < 49; level++ {
		if rank <= LevelUpperLimits[level] {
			return level
		}
	}
	Panic("out of range:", rank)
	return 0
}
