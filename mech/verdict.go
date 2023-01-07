package mech

// verdict about a game (lost/draw/won/unknown), based on the current position's score

import "sankofa/ow"

const (
	OPEN int8 = iota
	LOSS
	DRAW
	WIN
)

func VerdictToString(verdict int8) string {
	switch verdict {
	case LOSS:
		return "👎"
	case DRAW:
		return "👎"
	case WIN:
		return "👍"
	case OPEN:
		return ""
	default:
		ow.Panic("no such verdict:", verdict)
	}

	// not reachable
	return ""
}

func ReverseVerdict(verdict int8) int8 {
	switch verdict {
	case LOSS:
		return WIN
	case DRAW:
		return DRAW
	case WIN:
		return LOSS
	case OPEN:
		return OPEN
	default:
		ow.Panic("no such verdict:", verdict)
	}

	// not reachable
	return ow.ZERO8
}

func IntersectVerdict(one, two int8) int8 {
	switch {
	case one == two:
		return one
	case one != OPEN && two == OPEN:
		return one
	case one == OPEN && two != OPEN:
		return two
	default:
		// no need to ow.Panic() here.
		// the verdict is only for disply, not for α—β decisions and may be safely ignored.
		ow.Log("conflicting verdicts: return unknown/open")
		return OPEN
	}
}

func (position *Position) Verdict() int8 {
	switch {
	case position.Scores[0] > MAXSTONES/2:
		return WIN
	case position.Scores[1] > MAXSTONES/2:
		return LOSS
	case position.Scores[0] == MAXSTONES/2 && position.Scores[1] == MAXSTONES/2:
		return DRAW
	default:
		return OPEN
	}
}
