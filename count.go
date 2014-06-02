package rgoybiv

import (
	"image"
	"image/color"
	"sort"
)

type colorCount struct {
	Key   color.Color
	Value uint32
}

type colorCounts []colorCount

func (p colorCounts) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p colorCounts) Len() int {
	return len(p)
}

func (p colorCounts) Less(i, j int) bool {
	// Reverse sort
	return p[i].Value > p[j].Value
}

func countColors(img image.Image) colorCounts {
	bounds := img.Bounds()
	dist := map[color.Color]uint32{}
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := color.RGBAModel.Convert(img.At(x, y))

			if _, ok := dist[c]; ok {
				dist[c] += 1
			} else {
				dist[c] = 1
			}
		}
	}

	p := make(colorCounts, len(dist))
	i := 0
	for k, v := range dist {
		p[i] = colorCount{k, v}
		i++
	}
	sort.Sort(p)
	return p
}
