package ow

// mathematics

import (
	"math/rand"
	"time"
)

////////////////////////////////////////////////////////////////
// CONSTANTS || TYPES || VARIABLES
////////////////////////////////////////////////////////////////

// used with the semantics of NaN.
const MININT = -int64(^uint64(0) >> 1)
const MAXINT = int64(^uint64(0) >> 1)
const ZERO = int64(0)
const ONE = int64(1)
const TWO = int64(2)

// nanoseconds to seconds
const GIGA = 1000000000

////////////////////////////////////////////////////////////////
// ARITHMETIC
////////////////////////////////////////////////////////////////

// random number generator
var Rng = rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

// is this integer even?
func Even(i int64) bool {
	return i%2 == 0
}

// is this integer odd?
func Odd(i int64) bool {
	return !Even(i)
}

// -1 ** i
func Alternate(i int64) int64 {
	if Even(i) {
		return 1
	} else {
		return -1
	}
}

func Max(i, j int64) int64 {
	if i > j {
		return i
	} else {
		return j
	}
}

func Min(i, j int64) int64 {
	if i < j {
		return i
	} else {
		return j
	}
}

// absolute value of an integer;
//
// golang only offers this functionality for floats
func Abs(i int64) int64 {
	if i > 0 {
		return i
	} else {
		return -i
	}
}

// x ** n
//
// v. TAOCP 4.6.3
func Pow(x, n int64) int64 {
	pow := ONE
	for n > ZERO {
		if n&ONE != ZERO {
			pow *= x
		}
		n >>= ONE
		x *= x
	}
	return pow
}
