// build the HTML page shown by the server
//
// # DESIGN, TACTICS AND HACKS
//
// REST design, as follows:
//   - requests are encoded in the URL.
//     there are no cookies, no state database etc.
//   - replies are in a single HTML with embedded CSS and SVG
package html

import (
	"fmt"
	"sankofa/mech"
	"sankofa/ow"
	"strings"
	"time"
)

////////////////////////////////////////////////////////////////
// VIRTUAL OWARE BOARD
////////////////////////////////////////////////////////////////

// game history table width
// (NOT width of the board!)
const TABLEWIDTH = 5 * 2

// UTF-8 runes that describe the house editing operations
const INC = "+"
const INC2 = "⊕" // increment by several stones
const DEC = "-"
const DEC2 = "⊖" // decrement by several stones

// build Web page with GUI for game position and move history
// WARNING reason for spaghetti: linear story, not much can be reused
func Display(rest string) string {
	ow.Log("request:", rest)
	var html string

	fmt.Println("................................................................................")
	game := mech.StringToGame(rest)
	Analysis := Analysis(rest)

	html += `<!doctype html>
<html>
<head>
<meta charset="utf-8">
`
	html += CSS()
	html += "<title>Oware</title>\n"
	html += "</head>\n"
	html += "<body>\n"

	////////////////////////////////////////////////////////////////
	// α—β = scores
	// ν = degrees of freedom = moves-in-hand
	// Δν = ν advantage when compared to the other side's ν
	//
	// α—β|Δν = position evaluation
	////////////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////////////
	// NORTH
	////////////////////////////////////////////////////////////////

	html += "<table>\n"

	html += "<tr>\n"
	if ow.Odd(Analysis.game.Cursor) {
		// N to move
		html += "<th title=\"Δν=moves-in-hand advantage; α—β=predicted score\">α—β|Δν&nbsp;&nbsp;</th>\n"
		html += "<td title=\"♟︎ α—β|Δν\">" + Analysis.north.αβ + "|" + ow.Thousands(Analysis.north.δν) + "</td>\n"

		// ... moves
		for i := mech.NORTHRIGHT; i >= mech.NORTHLEFT; i-- {
			if Analysis.moves[i].movable {
				// highlight recommended move
				if ow.Odd(Analysis.game.Cursor) && Analysis.moves[i].attack {
					html += "<th title=\"♟︎ α—β|Δν after moving " + mech.MoveToString(i) + "\">"
				} else {
					html += "<td title=\"♟︎ α—β|Δν after moving " + mech.MoveToString(i) + "\">"
				}
				html += Analysis.moves[i].αβ
				html += "|" + ow.Thousands(-Analysis.moves[i].δν)
				// highlight attacks
				if ow.Odd(Analysis.game.Cursor) && Analysis.moves[i].attack {
					html += "</th>"
				} else {
					html += "</td>"
				}
			} else {
				// not a valid move
				html += "<td></td>\n"
			}
		}
	} else {
		html += "<td/><td>⁕</td>"
	}
	html += "</tr>\n"

	////////////////////////////////////////////////////////////////
	// stone counters
	////////////////////////////////////////////////////////////////
	html += "<tr>\n"
	html += "<th>Stones</th>\n"

	html += "<td title=\"♟︎ stone counter\">"
	html += ow.Thousands(Analysis.south.position.NorthStones())
	html += "</td>\n"
	// ... moves
	for i := mech.NORTHRIGHT; i >= mech.NORTHLEFT; i-- {
		html += "<td title=\"" + mech.MoveToString(i) + " stone counter\">"
		html += "<a title=\"remove four stones\" href =\"/" + ow.Thousands(Analysis.south.position.Edit(i, -4).Rank()) + "\">" + DEC2 + "</a> "
		html += "<a title=\"remove one stone\" href =\"/" + ow.Thousands(Analysis.south.position.Edit(i, -1).Rank()) + "\">" + DEC + "</a> "
		html += ow.Thousands(Analysis.south.position.Board[i]) + " "
		html += "<a title=\"add one stone\" href =\"/" + ow.Thousands(Analysis.south.position.Edit(i, 1).Rank()) + "\">" + INC + "</a> "
		html += "<a title=\"add four stones\" href =\"/" + ow.Thousands(Analysis.south.position.Edit(i, 4).Rank()) + "\">" + INC2 + "</a>"
		html += "</td>\n"
	}
	html += "</tr>\n"

	////////////////////////////////////////////////////////////////
	// house names
	////////////////////////////////////////////////////////////////
	html += `<tr>
<td></td>
<td>♟︎</td>
<th>f</th>
<th>e</th>
<th>d</th>
<th>c</th>
<th>b</th>
<th>a</th>
</tr>
`

	////////////////////////////////////////////////////////////////
	// graphics
	////////////////////////////////////////////////////////////////
	html += "<tr>\n"
	html += "<td/>\n"
	// score
	html += "<td title=\"♟︎ score\">" + ow.Thousands(Analysis.south.position.Scores[1]) + "</td>\n"
	// images
	for i := mech.NORTHRIGHT; i >= mech.NORTHLEFT; i-- {
		// move on reflected board
		j := i - mech.NORTHLEFT
		if ow.Odd(Analysis.game.Cursor) && Analysis.moves[i].movable &&
			!(Analysis.game.Cursor == len(Analysis.game.Positions)-1 && Analysis.game.GameOver()) {
			html += "<td title=\"play " + mech.MoveToString(i) + "\">\n<a href=\""
			html += Analysis.game.Move(j).String() + "\">\n"
			html += SVG(Analysis.south.position.Board[i], Analysis.moves[i].changed, Analysis.moves[i].check, false) + "</a>\n</td>\n"
		} else {
			html += "<td>\n" + SVG(Analysis.south.position.Board[i], Analysis.moves[i].changed, Analysis.moves[i].check, false) + "</td>\n"
		}
	}
	html += "</tr>\n"

	////////////////////////////////////////////////////////////////
	// SOUTH
	////////////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////////////
	// graphics
	////////////////////////////////////////////////////////////////
	html += "<tr>\n"
	html += "<td/>\n"
	// score
	html += "<td title=\"♙ score\">" + ow.Thousands(Analysis.south.position.Scores[0]) + "</td>\n"
	// images
	for i := mech.SOUTHLEFT; i <= mech.SOUTHRIGHT; i++ {
		if ow.Even(Analysis.game.Cursor) && Analysis.moves[i].movable &&
			!(Analysis.game.Cursor == len(Analysis.game.Positions)-1 && Analysis.game.GameOver()) {
			html += "<td title=\"play " + mech.MoveToString(i) + "\">\n<a href=\""
			html += Analysis.game.Move(i).String() + "\">\n"
			html += SVG(Analysis.south.position.Board[i], Analysis.moves[i].changed, Analysis.moves[i].check, false) + "</a>\n</td>\n"
		} else {
			html += "<td>\n" + SVG(Analysis.south.position.Board[i], Analysis.moves[i].changed, Analysis.moves[i].check, false) + "</td>\n"
		}
	}

	html += "</tr>\n"

	////////////////////////////////////////////////////////////////
	// house names
	////////////////////////////////////////////////////////////////
	html += `<tr>
<td></td>
<td>♙</td>
<th>A</th>
<th>B</th>
<th>C</th>
<th>D</th>
<th>E</th>
<th>F</th>
</tr>
`

	////////////////////////////////////////////////////////////////
	// stone counters
	////////////////////////////////////////////////////////////////
	html += "<tr>\n"
	html += "<th>Stones</th>\n"
	html += "<td title=\"♙ stone counter\">"
	html += ow.Thousands(Analysis.south.position.SouthStones())
	html += "</td>\n"
	for i := mech.SOUTHLEFT; i <= mech.SOUTHRIGHT; i++ {
		html += "<td title=\"" + mech.MoveToString(i) + " stone counter\">"
		html += "<a title=\"remove four stones\" href =\"/" + ow.Thousands(Analysis.south.position.Edit(i, -4).Rank()) + "\">" + DEC2 + "</a> "
		html += "<a title=\"remove one stone\" href =\"/" + ow.Thousands(Analysis.south.position.Edit(i, -1).Rank()) + "\">" + DEC + "</a> "
		html += ow.Thousands(Analysis.south.position.Board[i]) + " "
		html += "<a title=\"add one stone\" href =\"/" + ow.Thousands(Analysis.south.position.Edit(i, 1).Rank()) + "\">" + INC + "</a> "
		html += "<a title=\"add four stones\" href =\"/" + ow.Thousands(Analysis.south.position.Edit(i, 4).Rank()) + "\">" + INC2 + "</a>"
		html += "</td>\n"
	}
	html += "</tr>\n"

	html += "<tr>\n"
	if ow.Even(Analysis.game.Cursor) {
		// S to move
		html += "<th title=\"Δν=moves-in-hand advantage; α—β=predicted score\">α—β|Δν&nbsp;&nbsp;</th>\n"
		html += "<td title=\"♙ α—β|Δν\">" + Analysis.south.αβ + "|" + ow.Thousands(Analysis.south.δν) + "</td>\n"

		// ... moves
		for i := mech.SOUTHLEFT; i <= mech.SOUTHRIGHT; i++ {
			if Analysis.moves[i].movable {
				// highlight recommended move
				if ow.Even(Analysis.game.Cursor) && Analysis.moves[i].attack {
					html += "<th title=\"♙ α—β|Δν after moving " + mech.MoveToString(i) + "\">"
				} else {
					html += "<td title=\"♙ α—β|Δν after moving " + mech.MoveToString(i) + "\">"
				}
				html += Analysis.moves[i].αβ
				html += "|" + ow.Thousands(-Analysis.moves[i].δν)

				// highlight attacks
				if ow.Even(Analysis.game.Cursor) && Analysis.moves[i].attack {
					html += "</th>"
				} else {
					html += "</td>"
				}
			} else {
				// not a valid move
				html += "<td></td>\n"
			}
		}
	} else {
		html += "<td/><td>⁕</td>"
	}
	html += "</tr>\n"
	html += "</table>\n"

	////////////////////////////////////////////////////////////////
	// KEY DATA
	////////////////////////////////////////////////////////////////

	// navigation to initial/rev/final position
	html += "<table>\n"
	html += "<tr>\n"
	html += "<td title=\"game start\"><a href =\"/" + ow.Thousands(mech.INIRANK) + "\">⇐</a></td>\n"
	html += "<td title=\"reverse the board\"><a href =\"/" + ow.Thousands(Analysis.north.position.Rank()) + "\"> ↺ </a></td>\n"
	html += "<td title=\"empty board\"><a href =\"/" + ow.Thousands(mech.MINRANK) + "\">⇒</a></td>\n"
	html += "</tr>\n"
	html += "</table>\n"

	// key information
	html += "<table>\n"
	html += "<tr><td>Rank: " + ow.Thousands(Analysis.game.Current().Rank()) + ".</td></tr>\n"
	html += "<tr><td>" + ow.Thousands(Analysis.south.position.Stones()) + " stones on the board.</td></tr>\n"
	html += "<tr><td>α—β search depth: " + ow.Thousands(Analysis.tt.Depth()-Analysis.tt.Base()) + ".</td></tr>\n"
	html += "</table>\n"

	////////////////////////////////////////////////////////////////
	// GAME HISTORY
	////////////////////////////////////////////////////////////////
	html += "<p>\n"
	html += "<table id=\"moves\">\n"
	html += "<tr>\n"

	// display α—β continuation only if the cursor is at the last position.
	var continuation bool
	clone := game.Clone()
	if game.Cursor == len(game.Positions)-1 {
		clone = Analysis.game.Clone()
		continuation = true
	}

	for i, position := range clone.Positions {
		// skip initial position: it has no leading move
		if i == 0 {
			continue
		}
		// odd moves are south moves: print number
		if ow.Odd(i) {
			// close row, except for first time
			if i > 1 && i%TABLEWIDTH == 1 {
				html += "</tr>\n<tr>\n"
			}
			html += "<th id=\"left\">"
			html += ow.Thousands((i + 1) / 2)
			html += ".</th>\n"
		}

		// precompute move
		move := strings.ToLower(mech.MoveToString(clone.Moves[i-1]))
		if ow.Odd(i) {
			move = strings.ToUpper(move)
		}
		// show captures
		delta := position.Scores[1] - clone.Positions[i-1].Scores[0]
		if delta > 0 {
			move += "+" + ow.Thousands(delta)
		}
		if clone.GameOver() || clone.Cycle() || clone.Last().IsStarved() {
			if i == len(clone.Positions)-1 {
				move += " Ω"
			}
		}

		// ... leading position
		html += "<td title=\"play " + ow.Thousands((i+1)/2) + ". " + move + "\" id=\"left\">\n"

		clone.Cursor = i
		html += "<a href =\"" + clone.String() + "\""

		if i == game.Cursor {
			html += " class=\"cur\""
		}
		if i > game.Cursor && continuation {
			html += " class=\"cont\""
		}

		html += ">"
		if ow.Odd(i) {
			html += move
		} else {
			html += move
		}
		html += "</a>\n"
		html += "</td>\n"
	}
	// finish the last row
	for i := int64(len(clone.Positions)); i%TABLEWIDTH != 1; i++ {
		// move number
		if ow.Odd(i) {
			html += "<th/>"
		}
		html += "<td></td>\n"
	}
	html += "</tr>"
	html += "</table>"

	////////////////////////////////////////////////////////////////
	// DESCRIPTION IN WORDS
	////////////////////////////////////////////////////////////////

	html += "<table>\n"
	// game status

	if Analysis.game.Cycle() {
		html += "<tr><th>A repeated position ends the game.</th></tr>\n"
	}
	if Analysis.game.Last().IsStarved() {
		if ow.Odd(Analysis.game.Cursor) {
			html += "<tr><th>♟︎ to move but starved.</th></tr>\n"
		} else {
			html += "<tr><th>♙ to move but starved.</th></tr>\n"
		}
	}
	if Analysis.game.Last().IsWin() {
		ow.Log("WIN")
		if ow.Even(Analysis.game.Cursor) {
			// South to move
			html += "<tr><th>♙ wins.</th></tr>\n"
		} else {
			// Noth to move
			html += "<tr><th>♟︎ wins.</th></tr>\n"
		}
	}
	if Analysis.game.Last().IsLoss() {
		ow.Log("LOSS")
		if ow.Even(Analysis.game.Cursor) {
			// South to move
			html += "<tr><th>♟︎ wins.</th></tr>\n"
		} else {
			// Noth to move
			html += "<tr><th>♙ wins.</th></tr>\n"
		}
	}
	if Analysis.game.Last().IsDraw() {
		ow.Log("DRAW")
		html += "<tr><th>Draw game.</th></tr>\n"
	}
	html += "</table>\n"

	////////////////////////////////////////////////////////////////
	// COPYRIGHT
	////////////////////////////////////////////////////////////////
	html += `<p>
<small>
Copyright ©2019-2022 Carlo Monte.
</small>
</p>
</body>
</html>
`
	fmt.Printf("%.2f seconds\n", float64(time.Now().UTC().UnixNano()-Analysis.tt.Begin())/ow.GIGA64F)
	return html
}
