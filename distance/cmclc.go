// Package distance implements the CMC(l:c) algorithm for finding the distance
// between two colours
package distance

import (
	"math"
	"image/color"
	colorful "github.com/lucasb-eyer/go-colorful"
)


func cf(c color.Color) colorful.Color {
	d := color.NRGBAModel.Convert(c).(color.NRGBA)
	r := d.R; g := d.G; b := d.B

	rn := float64(uint8(r))
	gn := float64(uint8(g))
	bn := float64(uint8(b))

	return colorful.Color{rn, gn, bn}
}

func sq(v float64) float64 {
	return v * v
}


// Implements the "Delta E (CMC)" colour distance algorithm. Uses two
// parameters l and c, typically expressed as CMC(l:c). This, then, implements
// CMC(2:1).
//
// See http://www.brucelindbloom.com/index.html?Eqn_DeltaE_CMC.html
func Distance(a, b color.Color) float64 {
	const (
		l float64 = 2
		c float64 = 1
	)

	l1, a1, b1 := cf(a).Lab()
	l2, a2, b2 := cf(b).Lab()

	deltaL := l1 - l2
	deltaA := a1 - a2
	deltaB := b1 - b2

	c1 := math.Sqrt(sq(a1) + sq(b1))
	c2 := math.Sqrt(sq(a2) + sq(b2))
	deltaC := c1 - c2

	h1 := math.Atan2(b1, a1)
	deltaH := math.Sqrt(sq(deltaA) + sq(deltaB) + sq(deltaC))

	sl := 0.511
	if l1 >= 16 {
		sl = (0.040975 * l1) / (1 + 0.01765 * l1)
	}

	sc := (0.0638 * c1) / (1 + 0.0131 * c1) + 0.638

	var t float64
	if 164 <= h1 && h1 <= 345 {
		t = 0.56 + math.Abs(0.2 * math.Cos(h1 + 168))
	} else {
		t = 0.36 + math.Abs(0.4 * math.Cos(h1 + 35))
	}

	f := math.Sqrt(math.Pow(c1, 4) / (math.Pow(c1, 4) + 1900))

	sh := sc * (f * t + 1 - f)

	deltaE := math.Sqrt(
		sq(deltaL / (l * sl)) +
		sq(deltaC / (c * sc)) +
		sq(deltaH / sh))

	return deltaE
}
