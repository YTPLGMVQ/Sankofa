package ow

// combinatorics

// index for binomial coefficients
type nk struct {
	n, k int64
}

// saved binomial coefficients
var bin [256][256]int64

// WARNING init() methods are run in the file name lexicographic order
// WARNING MUST run before init() in level.go

// memoize binomial coefficients within a plausible range
func init() {
	for n := ZERO64; n < 256; n++ {
		for k := ZERO64; k <= n; k++ {
			bin[n][k] = _binomial(n, k)
		}
	}
}

// combinations of n taken by k without repetitions
func _binomial[N int8 | int64 | int](n, k N) int64 {
	// be lazy
	if 2*k > n {
		return _binomial(n, n-k)
	}

	// factorials(
	c := ONE64
	for i := ONE64; i <= int64(k); i++ {
		c *= (int64(n) - i + 1)
		c /= i
	}
	return c
}

// combinations of n taken by k without repetitions
func Binomial[N int8 | int64 | int](n, k N) int64 {
	// out of bounds
	if k == 0 || k == n {
		return 1
	}
	if n < 1 || k < 1 || k > n {
		return 0
	}
	return bin[n][k]
}

// combinations of n taken by k WITH repetitions
func Repetitions[N int8 | int64 | int](n, k N) int64 {
	return Binomial(n+k-1, k)
}
