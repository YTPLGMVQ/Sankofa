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
const MININT8 = -int8(^uint8(0) >> 1)
const MAXINT8 = int8(^uint8(0) >> 1)

const MININT64 = -int64(^uint64(0) >> 1)
const MAXINT64 = int64(^uint64(0) >> 1)

const MININT = -int(^uint(0) >> 1)
const MAXINT = int(^uint(0) >> 1)

const MINUSONE64 = int64(-1)
const ZERO64 = int64(0)
const ONE64 = int64(1)
const TWO64 = int64(2)

const ZERO8 = int8(0)
const ONE8 = int8(1)
const TWO8 = int8(2)

// nanoseconds to seconds
const GIGA = 1000000000
const GIGA64F = float64(GIGA)

////////////////////////////////////////////////////////////////
// ARITHMETIC
////////////////////////////////////////////////////////////////

// random number generator
var Rng = rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

// is this integer even?
func Even[N int8 | int64 | int](i N) bool {
	return i%2 == 0
}

// is this integer odd?
func Odd[N int8 | int64 | int](i N) bool {
	return !Even(i)
}

// -1 ** i
func Alternate[N int8 | int64 | int](i N) N {
	if Even(i) {
		return 1
	} else {
		return -1
	}
}

func Max[N int8 | int64 | int | float64](i, j N) N {
	if i > j {
		return i
	} else {
		return j
	}
}

func Min[N int8 | int64 | int | float64](i, j N) N {
	if i < j {
		return i
	} else {
		return j
	}
}

// absolute value of an integer;
//
// golang only offers this functionality for floats
func Abs[N int8 | int64 | int](i N) N {
	if i > 0 {
		return i
	} else {
		return -i
	}
}

// x ** n
//
// v. TAOCP 4.6.3
func Pow[N int8 | int64 | int](x, n N) N {
	pow := N(1)
	for n > 0 {
		if n&1 != 0 {
			pow *= x
		}
		n >>= 1
		x *= x
	}
	return pow
}
