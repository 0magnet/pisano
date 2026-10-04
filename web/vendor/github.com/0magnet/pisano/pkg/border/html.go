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
	FontPx float64 // font size; 0 means fit to FitW/FitH
	FitW   float64 // target box, used when FontPx is 0
	FitH   float64
	// Color of the whole border.
	//
	// One color, because the border is one element: it is composed and joined
	// into a single grid and turned once, which is what makes it a single line
	// with nothing to keep in phase. Coloring the corners separately would mean
	// splitting it up again, and the joins would fall between the pieces.
	Color   string
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

// HTML renders the border as a self-contained fragment: ONE turned <pre>.
//
// A piece per element was the wrong shape. Each had to be positioned in screen
// space and rotated about its own origin, and any error there put the cell
// grids out of phase — pieces sharing an angle but not a lattice, which reads
// worse than no alignment at all. Worse, the elements carried the RAW figures,
// so whatever joining had been worked out in lattice space was not in the
// picture: the fragment showed something other than what was measured.
//
// Composing to one grid and turning that removes both problems at once. There
// is one element, so there is nothing to keep in phase, and the thing rendered
// is the thing Components() counted.
func (l Layout) HTML(opt HTMLOptions) string {
	g := l.Grid()
	ix0, iy0, ix1, iy1 := l.inkBox(g)
	font := opt.FontPx
	if font <= 0 {
		font = FitFont(int(math.Ceil(ix1-ix0)), int(math.Ceil(iy1-iy0)), opt.FitW, opt.FitH)
	}
	cell := font * CellEm
	color := opt.Color
	if color == "" {
		color = "currentColor"
	}
	cls := opt.Class
	if cls == "" {
		cls = "pisano-border"
	}
	// The wrapper is the INKED extent, so it is the size of the border rather
	// than of the grid the border was composed in.
	minX, minY := ix0*cell, iy0*cell
	maxX, maxY := ix1*cell, iy1*cell
	var b strings.Builder
	fmt.Fprintf(&b, `<div class="%s" style="position:relative;width:%.1fpx;height:%.1fpx">`,
		html.EscapeString(cls), maxX-minX, maxY-minY)
	fmt.Fprintf(&b,
		`<pre style="position:absolute;margin:0;white-space:pre;transform-origin:0 0;`+
			`font:%.3fpx/%.2f ui-monospace,monospace;color:%s;left:%.2fpx;top:%.2fpx;`+
			`transform:rotate(%.4fdeg)">%s</pre>`,
		font, CellEm, color, -minX, -minY, l.Turn, html.EscapeString(g.String()))
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

// inkBox is the turned extent of the cells that actually carry a mark, in
// cells.
//
// The grid's own box is the wrong thing to measure. A border composed from
// diagonal runs fills a RHOMBUS inside a square grid, and turning that square
// gives a box half again as big with the drawing floating in the middle of it —
// a wrapper sized from it is mostly empty, and --fit then scales to the empty
// part rather than to the border.
func (l Layout) inkBox(g Grid) (minX, minY, maxX, maxY float64) {
	t := l.Turn * math.Pi / 180
	cos, sin := math.Cos(t), math.Sin(t)
	minX, minY = math.MaxFloat64, math.MaxFloat64
	maxX, maxY = -math.MaxFloat64, -math.MaxFloat64
	seen := false
	for r := range g {
		for c := range g[r] {
			if a, ok := Arms(g[r][c]); !ok || a == 0 {
				continue
			}
			seen = true
			for _, p := range [4][2]float64{
				{float64(c), float64(r)}, {float64(c + 1), float64(r)},
				{float64(c), float64(r + 1)}, {float64(c + 1), float64(r + 1)},
			} {
				x, y := p[0]*cos-p[1]*sin, p[0]*sin+p[1]*cos
				minX, minY = math.Min(minX, x), math.Min(minY, y)
				maxX, maxY = math.Max(maxX, x), math.Max(maxY, y)
			}
		}
	}
	if !seen {
		return 0, 0, 0, 0
	}
	return
}
