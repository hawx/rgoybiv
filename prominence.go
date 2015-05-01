package rgoybiv

import (
	"image/color"
	"math"
	"sort"

	"hawx.me/code/rgoybiv/distance"
)

type ColorProminence struct {
	Value      color.Color
	Prominence float64
}

type ColorProminences []ColorProminence

func (a ColorProminences) Len() int {
	return len(a)
}

func (a ColorProminences) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ColorProminences) Less(i, j int) bool {
	// reverse sort
	return a[i].Prominence > a[j].Prominence
}

func aggregate(cs colorCounts, minDistance float64) (ColorProminences, map[color.Color]color.Color) {
	var nPixels uint32 = 0

	toCanonical := map[color.Color]color.Color{
		color.Black: color.Black,
		color.White: color.White,
	}

	aggregated := map[color.Color]uint32{
		color.Black: 0,
		color.White: 0,
	}

	for _, col := range cs {
		c := col.Key
		n := col.Value

		nPixels += n

		if _, ok := aggregated[c]; ok {
			aggregated[c] += n
		} else {
			var nearest color.Color
			d := math.MaxFloat64

			for k, _ := range aggregated {
				ds := distance.Distance(c, k)
				if ds < d {
					d = ds
					nearest = k
				}
			}

			if d < minDistance {
				// nearby match
				aggregated[nearest] += n
				toCanonical[c] = nearest
			} else {
				// no nearby match
				aggregated[c] = n
				toCanonical[c] = c
			}
		}
	}

	colors := ColorProminences{}
	for c, n := range aggregated {
		colors = append(colors, ColorProminence{c, float64(n) / float64(nPixels)})
	}

	sort.Sort(colors)

	return colors, toCanonical
}
