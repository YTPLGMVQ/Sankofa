package html

// build pit diagrams as SVG

import (
	"fmt"
	"math"
	"sankofa/ow"
)

////////////////////////////////////////////////////////////////
// INLINE SVG GRAPHICS
////////////////////////////////////////////////////////////////

// convention: float64 starts with capitals, integers with small
const (
	svgRadius   = float64(65)   // half of the width/breadth of the square image
	houseRadius = float64(64)   // radius of a house
	stoneRadius = float64(14)   // radius of a stone
	spacer      = float64(1.07) // put a small space between the stones
	randomLimit = 10            // place at random after so many stones
	randomRing  = 7             // number of stones on the random ring
)

// stone radius with margin
var R = stoneRadius * spacer

// vertical henycomb offset
var Y = math.Sqrt(float64(3)/float64(4)) * 2 * stoneRadius * spacer

// place stones in a honeycomb pattern
var combPosition = [][]float64{
	// hexagon: 2-7
	{R * -1, Y},
	{R * 2, 0},
	{R * -1, Y * -1},
	{R, Y * -1},
	{R, Y},
	{R * -2, 0},
	// center: 1
	{0, 0},
	// stacked among three others
	// R is the radius, Y is the comb offset
	{0, R / math.Cos(math.Pi/6)},
	{R, R * -1 * math.Tan(math.Pi/6)},
	{R * -1, R * -1 * math.Tan(math.Pi/6)},
}

// a ring of stones, randomly rotated
func ring(count, moved int8) string {
	ow.Log(count, moved)
	var r string

	countFloat := float64(count)
	RingRadius := stoneRadius / math.Sin(math.Pi/countFloat)

	if count == 1 {
		if moved > 0 {
			r += fmt.Sprintf("<circle cx=\"%f\" cy=\"%f\" r=\"%f\" id=\"moved\"/>\n", svgRadius, svgRadius, stoneRadius)
		} else {
			r += fmt.Sprintf("<circle cx=\"%f\" cy=\"%f\" r=\"%f\" id=\"stone\"/>\n", svgRadius, svgRadius, stoneRadius)
		}
	} else {
		// random rotation of sone placement
		Rnd := ow.Rng.Float64() * 2 * math.Pi
		for i := ow.ZERO8; i < count; i++ {
			I := float64(i)
			φ := math.Pi * (2*I/countFloat + Rnd)
			X := svgRadius + RingRadius*math.Cos(φ)*spacer
			Y := svgRadius + RingRadius*math.Sin(φ)*spacer
			if i >= count-moved {
				r += fmt.Sprintf("<circle cx=\"%f\" cy=\"%f\" r=\"%f\" id=\"moved\"/>\n", X, Y, stoneRadius)
			} else {
				r += fmt.Sprintf("<circle cx=\"%f\" cy=\"%f\" r=\"%f\" id=\"stone\"/>\n", X, Y, stoneRadius)
			}
		}
	}

	return r
}

// stones placed on specific positions on a honeybee comb
// the orientation is somewhat random, for a more natural look
func comb(count, moved int8) string {
	ow.Log(count, moved)
	var r string
	Rnd := ow.Rng.Float64() * float64(2) * math.Pi
	sin, cos := math.Sincos(Rnd)
	ow.Log(Rnd, sin, cos)
	for i := ow.ZERO8; i < count; i++ {
		x := svgRadius + combPosition[i][0]*cos - combPosition[i][1]*sin
		y := svgRadius + combPosition[i][0]*sin + combPosition[i][1]*cos
		if i >= count-moved {
			r += fmt.Sprintf("<circle cx=\"%f\" cy=\"%f\" r=\"%f\" id=\"moved\"/>\n", x, y, stoneRadius)
		} else {
			r += fmt.Sprintf("<circle cx=\"%f\" cy=\"%f\" r=\"%f\" id=\"stone\"/>\n", x, y, stoneRadius)
		}
	}
	return r
}

// random place stones in huge houses
func random(count, moved int8) string {
	ow.Log(count, moved)
	var r string

	r += comb(randomLimit, randomLimit-(count-moved))

	countFloat := float64(count)
	RingRadius := stoneRadius / math.Sin(math.Pi/randomRing)

	// random rotation of sone placement
	for i := ow.ZERO8; i < count-randomLimit; i++ {
		I := float64(i)
		R := ow.Rng.Float64() * RingRadius
		φ := math.Pi * (2*I/(countFloat-randomLimit) + ow.Rng.Float64()*2*math.Pi)
		X := svgRadius + R*math.Cos(φ)
		Y := svgRadius + R*math.Sin(φ)
		if i >= count-moved-randomLimit {
			r += fmt.Sprintf("<circle cx=\"%f\" cy=\"%f\" r=\"%f\" id=\"moved\"/>\n", X, Y, stoneRadius)
		} else {
			r += fmt.Sprintf("<circle cx=\"%f\" cy=\"%f\" r=\"%f\" id=\"stone\"/>\n", X, Y, stoneRadius)
		}
	}

	return r
}

// facade for various stone placement functions
func stones(count, moved int8) string {
	var r string

	switch {
	case count < 6:
		r += ring(count, moved)
	case count <= randomLimit:
		r += comb(count, moved)
	default:
		r += random(count, moved)
	}

	return r
}

// create a string with an SVG diagram of a house
func SVG(count, moved int8, check, embedCSS bool) string {
	var r string
	if embedCSS {
		r += fmt.Sprintf("<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"%f\" height=\"%f\">\n", 2*svgRadius, 2*svgRadius)
		r += fmt.Sprintf(CSS())
	} else {
		r += fmt.Sprintf("<svg width=\"%f\" height=\"%f\">\n", 2*svgRadius, 2*svgRadius)
	}
	if check {
		r += fmt.Sprintf("<circle cx=\"%f\" cy=\"%f\" r=\"%f\" id=\"check\"/>\n", svgRadius, svgRadius, houseRadius)
	} else {
		r += fmt.Sprintf("<circle cx=\"%f\" cy=\"%f\" r=\"%f\" id=\"house\"/>\n", svgRadius, svgRadius, houseRadius)
	}
	r += stones(count, moved)
	r += "</svg>\n"
	return r
}
