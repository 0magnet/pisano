package border

import (
	"image"
	"image/color"
	"math"
)

// Image draws the border into a box of the given size in pixels.
//
// The HTML render is the one that ships, but it can only be looked at in a
// browser, and a browser brings its own page zoom, font substitution and
// device-pixel scaling to the picture — three ways for a screenshot to disagree
// with the thing that was composed. This draws the same grid straight onto a
// bitmap with nothing in between, so what comes out is what Compose built.
//
// Coverage rather than ink: a border a thousand cells across shown in a
// thousand pixels puts several cells in one pixel, and plotting them as a flat
// on/off makes the dense parts and the thin parts look the same. Counting how
// many cells land in each pixel and shading by that is what makes the drawing
// read at a size where a cell is smaller than a pixel — which is the size these
// borders are actually used at.
func (l Layout) Image(boxW, boxH int) *image.Gray {
	img := image.NewGray(image.Rect(0, 0, boxW, boxH))
	if boxW <= 0 || boxH <= 0 {
		return img
	}
	g := l.Grid()
	x0, y0, x1, y1 := l.inkBox(g)
	w, h := x1-x0, y1-y0
	if w <= 0 || h <= 0 {
		return img
	}
	// One scale for both axes, so the border keeps its proportions; the same
	// choice FitFont makes, for the same reason.
	s := math.Min(float64(boxW)/w, float64(boxH)/h)
	offX := (float64(boxW) - w*s) / 2
	offY := (float64(boxH) - h*s) / 2

	t := l.Turn * math.Pi / 180
	cos, sin := math.Cos(t), math.Sin(t)
	cover := make([]int, boxW*boxH)
	most := 0
	for r := range g {
		for c := range g[r] {
			if a, ok := Arms(g[r][c]); !ok || a == 0 {
				continue
			}
			fx, fy := float64(c)+0.5, float64(r)+0.5
			rx, ry := fx*cos-fy*sin, fx*sin+fy*cos
			px := int((rx-x0)*s + offX)
			py := int((ry-y0)*s + offY)
			if px < 0 || py < 0 || px >= boxW || py >= boxH {
				continue
			}
			i := py*boxW + px
			cover[i]++
			if cover[i] > most {
				most = cover[i]
			}
		}
	}
	if most == 0 {
		return img
	}
	// A square root rather than a straight ratio. Coverage is very long-tailed
	// — a knot where the path doubles back can hold fifty times what a plain
	// stroke does — and dividing by the peak leaves every ordinary stroke at
	// two or three percent gray, which is to say invisible.
	for i, v := range cover {
		if v == 0 {
			continue
		}
		f := math.Sqrt(float64(v) / float64(most))
		lum := 60 + f*195
		img.Pix[i] = color.Gray{Y: uint8(math.Min(255, lum))}.Y
	}
	return img
}
