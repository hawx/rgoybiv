// Package rgoybiv is a port of the python package
// https://github.com/givp/RoyGBiv
package rgoybiv

import (
	"github.com/hawx/img/utils"
	"github.com/hawx/quantise"
	colorful "github.com/lucasb-eyer/go-colorful"
	"image"
	"image/color"
)

const (
	N_QUANTIZED           = 100
	MIN_DISTANCE          = 10.0
	MIN_PROMINENCE        = 0.01
	MIN_SATURATION        = 0.05
	MAX_COLORS            = 5
	BACKGROUND_PROMINENCE = 0.5
)

type Options struct {
	// start with an adaptive palette of this size
	NQuantized int

	// min distance to consider two colors different
	MinDistance float64

	// ignore if less than this proportion of image
	MinProminence float64

	// ignore if not saturated enough
	MinSaturation float64

	// keep only this many colors
	MaxColors int

	// level of Prominence indicating a bg color
	BackgroundProminence float32
}

type Palette struct {
	Colors     ColorProminences
	Background color.Color
}

func cf(c color.Color) colorful.Color {
	r, g, b, _ := utils.NormalisedRGBAf(c)
	return colorful.Color{r, g, b}
}

func sq(v float64) float64 {
	return v * v
}

// GetAverage finds the average colour of an image.
func GetAverage(img image.Image) color.Color {
	bounds := img.Bounds()

	var rt, gt, bt, at uint64 = 0, 0, 0, 0
	var t uint64 = uint64(bounds.Dx()) * uint64(bounds.Dy())

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()

			rt += uint64(uint8(r))
			gt += uint64(uint8(g))
			bt += uint64(uint8(b))
			at += uint64(uint8(a))
		}
	}

	return color.RGBA{
		uint8(rt / t), uint8(gt / t), uint8(bt / t), uint8(at / t),
	}
}

// GetPalette finds a palette of the dominant colours in an image. Various
// parameters, such as the number of dominant colours to return, are given by
// the Options.
func GetPalette(img image.Image, opts *Options) Palette {
	minDistance := MIN_DISTANCE
	minSaturation := MIN_SATURATION
	minProminence := MIN_PROMINENCE
	maxColors := MAX_COLORS
	nQuantized := N_QUANTIZED

	if opts != nil {
		if opts.MinDistance != 0 {
			minDistance = opts.MinDistance
		}
		if opts.MaxColors != 0 {
			maxColors = opts.MaxColors
		}
		if opts.MinProminence != 0 {
			minProminence = opts.MinProminence
		}
		if opts.MinSaturation != 0 {
			minSaturation = opts.MinSaturation
		}
		if opts.NQuantized != 0 {
			nQuantized = opts.NQuantized
		}
	}

	q := quantise.OctreeQuantiser{
		Depth:    6,
		Size:     nQuantized,
		Strategy: quantise.LEAST,
	}

	img = quantise.Quantise(img, q)

	sortedCols := countColors(img)

	colors, toCanonical := aggregate(sortedCols, minDistance)

	colors, bgColor := detectBackground(img, colors, toCanonical)

	// keep any color which meets the minimum saturation
	satColors := ColorProminences{}
	for _, c := range colors {
		if meetsMinSaturation(c.Value, minSaturation) {
			satColors = append(satColors, c)
		}
	}

	if bgColor != nil && !meetsMinSaturation(bgColor, minSaturation) {
		bgColor = nil
	} else {
		if len(satColors) > 0 {
			colors = satColors
		} else {
			// keep at least one color
			colors = colors[:1]
		}
	}

	// keep any color within 10% of the majority color
	finalColors := []ColorProminence{}
	for _, c := range colors {
		if c.Prominence >= colors[0].Prominence*minProminence {
			finalColors = append(finalColors, c)
		}
	}

	return Palette{finalColors[:maxColors], bgColor}
}

func detectBackground(img image.Image, colors []ColorProminence, toCanonical map[color.Color]color.Color) ([]ColorProminence, color.Color) {
	// more then half the image means background
	if colors[0].Prominence >= BACKGROUND_PROMINENCE {
		return colors[1:], colors[0].Value
	}

	h, w := img.Bounds().Dy(), img.Bounds().Dx()

	points := []image.Point{
		image.Pt(0, 0), image.Pt(0, h/2), image.Pt(0, h-1), image.Pt(w/2, h-1),
		image.Pt(w-1, h-1), image.Pt(w-1, h/2), image.Pt(w-1, 0), image.Pt(w/2, 0),
	}

	edgeDist := map[color.Color]int{}
	for _, p := range points {
		c := img.At(p.X, p.Y)
		if _, ok := edgeDist[c]; ok {
			edgeDist[c] += 1
		} else {
			edgeDist[c] = 1
		}
	}

	var majorityCol color.Color
	majorityCount := 0
	for c, n := range edgeDist {
		if n > majorityCount {
			majorityCol = c
			majorityCount = n
		}
	}

	var bgColor color.Color
	foundColors := colors

	if majorityCount >= 3 {
		// we have a background color
		canonicalBg := toCanonical[majorityCol]
		for _, c := range colors {
			if c.Value == canonicalBg {
				bgColor = canonicalBg
			} else {
				foundColors = append(foundColors, c)
			}
		}
	}

	return foundColors, bgColor
}

func meetsMinSaturation(c color.Color, threshold float64) bool {
	_, s, _ := cf(c).Hsv()

	return s > threshold
}
