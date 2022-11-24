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
	for n := ZERO; n < 256; n++ {
		for k := ZERO; k <= n; k++ {
			bin[n][k] = _binomial(n, k)
		}
	}
}

// combinations of n taken by k without repetitions
func _binomial(n, k int64) int64 {
	// be lazy
	if 2*k > n {
		return _binomial(n, n-k)
	}

	// factorials(
	c := ONE
	for i := ONE; i <= k; i++ {
		c *= (n - i + 1)
		c /= i
	}
	return c
}

// combinations of n taken by k without repetitions
func Binomial(n, k int64) int64 {
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
func Repetitions(n, k int64) int64 {
	return Binomial(n+k-1, k)
}
