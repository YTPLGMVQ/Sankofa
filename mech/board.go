package mech

// representation of an Oware board

import (
	"regexp"
	"sankofa/ow"
	"strconv"
)

////////////////////////////////////////////////////////////////
// CONSTANTS
////////////////////////////////////////////////////////////////

// maximum number of stones on the board
const MAXSTONES = int8(48)

// lowest possible rank; empty board
const MINRANK = int64(0)

// rank of initial position
const INIRANK = int64(1224204106872)

// largest possible rank; we start counting at 0
const MAXRANK = int64(1399358844974)

// board corners
const SOUTHLEFT = int8(0)
const SOUTHRIGHT = int8(5)
const NORTHLEFT = int8(6)
const NORTHRIGHT = int8(11)

////////////////////////////////////////////////////////////////
// DATA TYPES
////////////////////////////////////////////////////////////////

// array with a number stones in each house
type Board [12]int8

////////////////////////////////////////////////////////////////
// CONVERSIONS
////////////////////////////////////////////////////////////////

// import from text, e.g.: 4.4.4.4.4.4-4.4.4.4.4.4
func StringToBoard(external string) Board {
	var board Board
	rex := regexp.MustCompile("[^0-9]+")
	for i, stones := range rex.Split(external, -1) {
		s, err := strconv.ParseInt(stones, 10, 8) // ParseInt always returns int64
		if err == nil {
			board[i] = int8(s) // ParseInt always returns an int64
		} else {
			ow.Panic("cannot parse number of stones in house: ", i)
		}
	}
	return board
}

func (houses Board) String() string {
	var r string
	for i := range houses {
		if i > 0 {
			if i == 6 {
				r += "-"
			} else {
				r += "."
			}
		}
		r += ow.Thousands(houses[i])
	}
	return r
}
