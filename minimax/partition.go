package minimax

// partition the range of possible scores (α—β) into ranges holding a similar number of positions.
// works under the asumption that the "split" heuristic is correct for the relevant majority of scores.
// produces at most #partitions ranges, yet none shorter than 3 steps (including the ends α and β).
// compared to an exhaustive search of best quartiles split (minimum SD) and found to be good enough and much faster.
// local maxima produce false results, but these are very close to the optimum solution.

import (
	"fmt"
	"math"
	"sankofa/ow"
)

////////////////////////////////////////////////////////////////
// EMPIRICAL CUMULATIVE HISTOGRAM: [LEVEL, SCORE] → Σ(POSITIONS)
////////////////////////////////////////////////////////////////

// offset for scores ∈{-48, ..., 0, ..., 48}.
// required because we use the score as an array index, which MUST be >=0
const OFFSET = 48

// number of positions with a given split-score for each <=level and each offseted score;
// hist[level][offset-score] = #positions
var hist [49][97]int64

// initialize histogram with cumulative numbers of positions for each score.
func init() {
	// logging is always turned on here, because the command line flags were not processed yet.
	// we start at 1, and look back at level 0 with no relevant positions.
	for level := ow.ONE; level < 49; level++ {
		for stones := ow.ZERO; stones <= level; stones++ {
			// no scores for the unreachable level 47
			if level == 47 {
				continue
			}
			score := stones - (level - stones)
			s := ow.Repetitions(6, ow.Abs(stones))
			n := ow.Repetitions(6, ow.Abs(level-stones))
			hist[level][score+OFFSET] = s * n
		}
		for score := -48; score < 49; score++ {
			hist[level][score+OFFSET] += hist[level-1][score+OFFSET]
		}
	}
}

////////////////////////////////////////////////////////////////
// CLASS Intervals
////////////////////////////////////////////////////////////////

// minimum interval length
const INT_INC = 2

// number of reserved partition intervals
const INT_CAP = 8

type Intervals [][2]int64

func NewIntervals(α, β, partitions int64) (r Intervals) {
	r = make(Intervals, 0, INT_CAP)
	last := α
	for i := ow.ONE; i < partitions; i++ {
		if β-last < INT_INC {
			ow.Log("no place left at:", i)
			break
		}
		r = append(r, [2]int64{last, last + INT_INC})
		last += INT_INC
	}
	if last < β || α == β {
		r = append(r, [2]int64{last, β})
	}

	return
}

func (intervals Intervals) String() (r string) {
	for _, i := range intervals {
		r += ow.Thousands(i[0], i[1])
	}

	return
}

func (one Intervals) EQ(two Intervals) bool {
	if len(one) != len(two) {
		return false
	}
	for i := range one {
		if one[i][0] != two[i][0] {
			return false
		}
		if one[i][1] != two[i][1] {
			return false
		}
	}
	return true
}

////////////////////////////////////////////////////////////////
// NEAR-EQUAL-WEIGHT PARTITION OF THE SCORES RANGE
////////////////////////////////////////////////////////////////

// standard deviation of the empirical quartiles distribution found in "intervals"
func (intervals Intervals) SD(level int64) (SD float64) {

	Σ0 := float64(len(intervals))
	var Σ1, Σ2 float64
	for i := 0; i < len(intervals); i++ {
		// increment Σx and Σx²
		x := ow.ZERO
		for j := intervals[i][0]; j <= intervals[i][1]; j++ {
			x += hist[level][j+OFFSET]
		}
		Σ1 += float64(x)
		Σ2 += float64(x) * float64(x)
	}
	SD = math.Sqrt((Σ2 - Σ1*Σ1/float64(Σ0)) / float64(Σ0))

	return
}

// shift a stake one position to the left; return success value
func (intervals Intervals) SHL(pos int64) bool {
	if pos < ow.ONE || pos > int64(len(intervals)-1) {
		ow.Panic("position:", pos, "intervals:", intervals, "OUT OF RANGE")
	}

	if intervals[pos-1][1]-intervals[pos-1][0] > INT_INC {
		intervals[pos-1][1] -= 1
		intervals[pos][0] = intervals[pos-1][1]
		ow.Log("shl@", pos, intervals)
		return true
	} else {
		ow.Log("NO shl@", pos, intervals)
		return false
	}
}

// shift a stake one position to the right; return success value
func (intervals Intervals) SHR(pos int64) bool {
	if pos < ow.ONE || pos > int64(len(intervals)-1) {
		ow.Panic("position:", pos, "intervals:", intervals, "OUT OF RANGE")
	}

	if intervals[pos][1]-intervals[pos][0] > INT_INC {
		intervals[pos][0] += 1
		intervals[pos-1][1] = intervals[pos][0]
		ow.Log("shr@", pos, intervals)
		return true
	} else {
		ow.Log("NO shr@", pos, intervals)
		return false
	}
}

// n-wise partitioning a list (or however it's called) appears to be an NP-complete problem;
// gradient descent searches for a local minimum; this is good enough for our purposes and pretty fast
func (intervals Intervals) GradientDescent(level int64) {
	ow.Log(intervals)

	// gradient descent in the partition shift-space
	// stop at local minimum SD e.g., when no immediate improvement through shits is possible
	SD := intervals.SD(level)
	found := true
	for found {
		found = false
		for stake := ow.ONE; stake < int64(len(intervals)); stake++ {
			// one direction, then the opposite
			if intervals.SHR(stake) {
				a := intervals.SD(level) // compute once; it is costly
				if a < SD {
					found = true
					SD = a
					ow.Log(intervals)
				} else {
					// undo
					intervals.SHL(stake)
				}
			}
			if intervals.SHL(stake) {
				a := intervals.SD(level) // compute once; it is costly
				if a < SD {
					found = true
					SD = a
					ow.Log(intervals)
				} else {
					// undo
					intervals.SHR(stake)
				}
			}
		}
	}
}

// split the α—β interval into =partitions equal quartiles using a gradient descent algorithm.
func Quartiles(α, β, level, partitions int64) (r Intervals) {
	ow.Log("Quartiles: α:", α, "β:", β, "level:", level, "partitions:", partitions)

	r = NewIntervals(ow.Max(α, -level), ow.Min(β, level), partitions)
	r.GradientDescent(level)

	fmt.Println(ow.Thousands(α, β), "level=", level, "partitions=", partitions, r)
	return
}
