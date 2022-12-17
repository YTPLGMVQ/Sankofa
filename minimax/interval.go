package minimax

import (
	"sankofa/ow"
)

// a range (interval) containing a score
type Interval struct {
	rank      int64
	low, high int8
}

// for logging and HTML output
func (interval *Interval) String() string {
	switch {
	case interval == nil:
		return "∅"
	case interval.low == ow.MININT8 && interval.high == ow.MAXINT8:
		return "∀"
	case interval.low == ow.MININT8:
		return "≤" + ow.Thousands(interval.high)
	case interval.high == ow.MAXINT8:
		return "≥" + ow.Thousands(interval.low)
	case interval.low == interval.high:
		return ow.Thousands(interval.low)
	default:
		return ow.Thousands(interval.low, interval.high)
	}
}

func NewInterval(rank int64, low, high int8) *Interval {
	r := new(Interval)
	r.rank = rank
	level := ow.Level(rank)
	r.low = ow.Max(-level, low)
	r.high = ow.Min(level, high)

	if r.high < r.low {
		ow.Panic("high < low:", rank, ":", r)
	}

	ow.Log("rank:", rank, ": low:", low, ", high:", high, "⇢", r)

	return r
}

func (interval *Interval) Clone() *Interval {
	return NewInterval(interval.rank, interval.low, interval.high)
}

// reverse the interval; score range from the perspective of the opponent
func (interval *Interval) Reverse() *Interval {
	if interval == nil {
		return nil
	}
	return NewInterval(interval.rank, -interval.high, -interval.low)
}

// do we have a final score (a zero-width interval)?
func (interval *Interval) IsFinal() bool {
	return interval.low == interval.high
}

func (interval *Interval) Score() int8 {
	if interval.IsFinal() {
		return interval.low
	} else {
		ow.Panic("score request on non-final interval:", interval)
	}

	// pro forma
	return ow.MININT8
}

func (interval *Interval) Plus(x int8) *Interval {
	r := interval.Clone()
	r.low += x
	r.high += x

	return r
}

func (one *Interval) EQ(two *Interval) bool {
	if one.rank == two.rank && one.low == two.low && one.high == two.high {
		return true
	} else {
		return false
	}
}

// disjoint intervals (not intersectable)?
func (one *Interval) Disjoint(two *Interval) bool {
	if one.high < two.low || one.low > two.high {
		return true
	} else {
		return false
	}
}

// interval intersection
func (one *Interval) Intersect(two *Interval) *Interval {
	if one.rank != two.rank {
		ow.Panic("different ranks:", one.rank, ":", one, "!=", two.rank, ":", two)
	}

	if one.Disjoint(two) {
		ow.Log(one.rank, ":", one, "⋂", two, "⇢ ∅, return:", two)
		return two.Clone()
	} else {
		int := NewInterval(one.rank, ow.Max(one.low, two.low), ow.Min(one.high, two.high))
		ow.Log(one.rank, ":", one, "⋂", two, "⇢", int)

		// sanity check
		if int.high < int.low {
			ow.Panic("low < high:", int.rank, ":", int)
		}

		return int
	}
}

// rank comparison
func (one *Interval) RankLt(two *Interval) bool {
	if one.rank < two.rank {
		return true
	} else {
		return false
	}
}

// score interval comparison
func (one *Interval) Gt(two *Interval) bool {
	if !one.Disjoint(two) {
		ow.Panic("overlapping intervals are not compareable:", one.rank, ":", one, "<>", two.rank, ":", two)
	}

	if one.low > two.high {
		return true
	} else {
		return false
	}
}
