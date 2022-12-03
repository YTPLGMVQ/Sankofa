package ow

// string conversions

import (
	"strconv"
)

////////////////////////////////////////////////////////////////
// type conversions to string and escape strings.
////////////////////////////////////////////////////////////////

// boolean to string
func YesNo(tell bool) string {
	if tell {
		return "yes"
	} else {
		return "no"
	}
}

// integer to string
func Thousands[N int8 | int64 | int](n ...N) string {
	var r string
	var plural bool

	for i, v := range n {
		m := int64(v)
		switch {
		case m == MININT64:
			r += "-∞"
		case m == MAXINT64:
			r += "+∞"
		default:
			r += strconv.FormatInt(m, 10)
		}
		if i < len(n)-1 {
			r += ", "
			plural = true
		}
	}

	if plural {
		r = "[" + r + "]"
	}

	return r
}
