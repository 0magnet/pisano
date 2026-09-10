package border

import (
	"fmt"
	"html"
	"math"
	"strings"
)

// The character cell, as a fraction of the font size.
//
// A monospace advance is very close to 0.6em, and setting line-height to the
// same fraction makes the cell SQUARE — which is what lets a figure be turned
// by 45 degrees and still be the figure. At the default line-height a cell is
// about twice as tall as wide, so every turned figure would come out sheared.
const CellEm = 0.6

// FitFont returns the font size, in pixels, at which a block of the given size
// in cells fills a box of the given size in pixels.
//
// This is the "zoom out on the big ones" knob: a figure forty cells wide and
// one six cells wide can be shown at the same apparent size by giving each the
// font that fits its own extent, rather than a single font that makes one
// microscopic and the other enormous.
func FitFont(cellsW, cellsH int, boxW, boxH float64) float64 {
	if cellsW <= 0 || cellsH <= 0 {
		return 0
	}
	cell := math.Min(boxW/float64(cellsW), boxH/float64(cellsH))
	return cell / CellEm
}

// HTMLOptions controls the rendered fragment.
type HTMLOptions struct {
	FontPx  float64 // font size; 0 means fit to FitW/FitH
	FitW    float64 // target box, used when FontPx is 0
	FitH    float64
	Color   string // the runs
	Corner  string // the corner figures; empty means the same as Color
	Class   string // class for the wrapper div
	Content string // HTML dropped in the middle of the frame
}

// RotatedBox is the smallest box, in cells, holding every piece once turned.
//
// Rotating the LATTICE bounds is not the same thing and gives a box that is
// both too big and the wrong shape: the pieces do not fill their lattice
// rectangle, and each piece has its own extent that the turn moves about. This
// takes the four corners of every piece's own box, turns them, and keeps the
// extremes — so a wide border gets a wide box rather than a square one with the
// drawing stranded in the top of it.
func (l Layout) RotatedBox() (minX, minY, maxX, maxY float64) {
	if len(l.Places) == 0 {
		return
	}
	t := l.Turn * math.Pi / 180
	cos, sin := math.Cos(t), math.Sin(t)
	minX, minY = math.MaxFloat64, math.MaxFloat64
	maxX, maxY = -math.MaxFloat64, -math.MaxFloat64
	for _, p := range l.Places {
		for _, c := range [4][2]float64{
			{float64(p.GX), float64(p.GY)},
			{float64(p.GX + p.Grid.W()), float64(p.GY)},
			{float64(p.GX), float64(p.GY + p.Grid.H())},
			{float64(p.GX + p.Grid.W()), float64(p.GY + p.Grid.H())},
		} {
			x, y := c[0]*cos-c[1]*sin, c[0]*sin+c[1]*cos
			minX, minY = math.Min(minX, x), math.Min(minY, y)
			maxX, maxY = math.Max(maxX, x), math.Max(maxY, y)
		}
	}
	return
}

// RotatedExtent is RotatedBox as a width and height in cells.
func (l Layout) RotatedExtent() (w, h float64) {
	minX, minY, maxX, maxY := l.RotatedBox()
	return maxX - minX, maxY - minY
}

// HTML renders the layout as a self-contained fragment: one positioned div with
// a <pre> per piece.
//
// Every piece carries the SAME rotation and sits at its lattice position turned
// into screen space, with transform-origin at its top-left so that point is its
// grid origin. That is what holds the pieces in phase — rotating each about its
// own center gives them a common angle but an arbitrary offset, and the
// lattices then almost line up, which reads worse than not lining up at all.
func (l Layout) HTML(opt HTMLOptions) string {
	font := opt.FontPx
	if font <= 0 {
		w, h := l.RotatedExtent()
		font = FitFont(int(math.Ceil(w)), int(math.Ceil(h)), opt.FitW, opt.FitH)
	}
	cell := font * CellEm
	color := opt.Color
	if color == "" {
		color = "currentColor"
	}
	corner := opt.Corner
	if corner == "" {
		corner = color
	}
	t := l.Turn * math.Pi / 180
	cos, sin := math.Cos(t), math.Sin(t)

	// Place everything, then shift so the drawing starts at the origin: the
	// rotation moves pieces to negative coordinates as often as not.
	// One box for both the wrapper and the offsets, so the drawing starts at the
	// wrapper's origin and the wrapper is exactly as big as the drawing.
	bx, by, bx2, by2 := l.RotatedBox()
	type placed struct {
		x, y float64
		p    Placement
	}
	var ps []placed
	for _, p := range l.Places {
		u, v := float64(p.GX)*cell, float64(p.GY)*cell
		ps = append(ps, placed{u*cos - v*sin, u*sin + v*cos, p})
	}
	minX, minY := bx*cell, by*cell

	var b strings.Builder
	cls := opt.Class
	if cls == "" {
		cls = "pisano-border"
	}
	fmt.Fprintf(&b, `<div class="%s" style="position:relative;width:%.1fpx;height:%.1fpx">`,
		html.EscapeString(cls), (bx2-bx)*cell, (by2-by)*cell)
	for _, q := range ps {
		col := color
		if q.p.Role == RoleCorner {
			col = corner
		}
		fmt.Fprintf(&b,
			`<pre style="position:absolute;margin:0;white-space:pre;transform-origin:0 0;`+
				`font:%.3fpx/%.2f ui-monospace,monospace;color:%s;left:%.2fpx;top:%.2fpx;`+
				`transform:rotate(%.4fdeg)">%s</pre>`,
			font, CellEm, col, q.x-minX, q.y-minY, l.Turn, html.EscapeString(q.p.Grid.String()))
	}
	if opt.Content != "" {
		fmt.Fprintf(&b, `<div style="position:absolute;inset:0;display:flex;align-items:center;`+
			`justify-content:center">%s</div>`, opt.Content)
	}
	b.WriteString(`</div>`)
	return b.String()
}

// Page wraps fragments in a minimal standalone document, for looking at.
func Page(title string, fragments ...string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "<!doctype html>\n<meta charset=\"utf-8\">\n<title>%s</title>\n",
		html.EscapeString(title))
	b.WriteString("<style>body{background:#0b0e13;color:#3d8fb8;font:13px system-ui;" +
		"padding:18px;margin:0;display:flex;flex-wrap:wrap;gap:24px;align-items:flex-start}</style>\n")
	for _, f := range fragments {
		b.WriteString(f)
		b.WriteByte('\n')
	}
	return b.String()
}
