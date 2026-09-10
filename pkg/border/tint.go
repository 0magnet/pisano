package border

import (
	"fmt"
	"html"
	"strings"

	"github.com/0magnet/pisano/pkg/pisano"
)

// Coloring a composed border.
//
// pisano tints a figure as it WALKS it — the tinter is handed each step in
// order, and every mode is some question about that step: which pass laid it
// down, which way the turtle faced, how many times it has been here. A composed
// border has none of that. It is stamped from figures that were walked
// separately, cut, trimmed and joined, and by the time it is a grid the order
// the pen went in is gone.
//
// So the order is recovered rather than remembered. The border is one closed
// line — that is the whole claim the composition makes and the thing Components
// checks — so walking it from any cell gives a traversal, and coloring by how
// far along that traversal a cell is gives pisano's TintAge: the palette once
// per circuit, showing the order the figure was drawn in. It is the one mode
// that survives losing the pen, and on a frame it is the one that reads: the
// color travels round the border.
//
// Colors are pisano's own, from PassHex: the six its turtle command tints with
// in a terminal, so a bordered page is the colors the tool draws. The SVG
// sheets have a different palette and it is the wrong one here.

// Tint is a palette index per cell, or -1 where there is no mark.
type Tint [][]int

// TintByCopy colors each piece of the border by which copy of its run it is,
// stepping through the palette — which is what pisano does.
//
// pisano tints a figure by the PASS that laid a step down: one solid color per
// pass, stepping through the six. Run `pisano turtle --mod 13 -r12` and that is
// what comes out — four rows of red, then four of blue, then four of yellow,
// one motif per color. A copy of a run figure IS a pass, so a border colored
// this way is the same drawing the terminal makes.
//
// It replaces coloring by position along the walked line, which was an attempt
// to reconstruct pisano's TintAge from a grid that had forgotten the pen. That
// banded the border in six sweeps and looked nothing like the tool: the bands
// fell wherever the traversal happened to be, cutting across motifs instead of
// following them.
//
// The four corners share the first color. A corner is one closed figure walked
// four times, and pisano would give those four passes four colors; that needs
// per-step data the composition does not keep, and four corners in four colors
// would read as an accident rather than as a frame.
func (l Layout) TintByCopy(g Grid, colors int) Tint {
	if colors < 1 {
		colors = 1
	}
	h, w := g.H(), g.W()
	t := make(Tint, h)
	for i := range t {
		t[i] = make([]int, w)
		for j := range t[i] {
			t[i][j] = -1
		}
	}
	minX, minY, _, _ := l.Bounds()
	const pad = 2
	marked := func(x, y int) bool {
		if y < 0 || y >= h || x < 0 || x >= w {
			return false
		}
		a, ok := Arms(g[y][x])
		return ok && a != 0
	}
	// In the order GridJoins stamps: corners, then the runs over them. Where
	// two pieces share a cell the later one wins, which is the same rule the
	// canvas uses when a path crosses itself.
	paint := func(want Role) {
		for _, p := range l.Places {
			if p.Role != want {
				continue
			}
			idx := 0
			if p.Role != RoleCorner {
				idx = ((p.Seq % colors) + colors) % colors
			}
			ox, oy := p.GX-minX+pad, p.GY-minY+pad
			for r := range p.Grid {
				for c := range p.Grid[r] {
					if a, ok := Arms(p.Grid[r][c]); !ok || a == 0 {
						continue
					}
					if x, y := ox+c, oy+r; marked(x, y) {
						t[y][x] = idx
					}
				}
			}
		}
	}
	paint(RoleCorner)
	paint(RoleAcross)
	paint(RoleDown)
	// The short segments drawn out of a corner, and anything Join added, belong
	// to no piece. They take a neighbor's color rather than none, so a join does
	// not show up as a gap in the coloring.
	for r := range g {
		for c := range g[r] {
			if !marked(c, r) || t[r][c] >= 0 {
				continue
			}
			for _, d := range [4][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
				y, x := r+d[0], c+d[1]
				if y >= 0 && y < h && x >= 0 && x < w && t[y][x] >= 0 {
					t[r][c] = t[y][x]
					break
				}
			}
			if t[r][c] < 0 {
				t[r][c] = 0
			}
		}
	}
	return t
}

// TintHTML renders the grid as HTML, one span per run of a color.
//
// A span per run rather than per cell, which is the difference between a few
// hundred elements and a hundred thousand: the color changes only where the
// walk crosses into the next sixth of the circuit, so the runs are long.
//
// Colors default to pisano's dark pass palette, so a border on a page and a
// pisano sheet are drawn from the same six.
func (g Grid) TintHTML(t Tint, colors []string) string {
	if len(colors) == 0 {
		colors = pisano.PassHex()
	}
	var b strings.Builder
	cur, open := -2, false
	closeSpan := func() {
		if open {
			b.WriteString("</span>")
			open = false
		}
	}
	for r := range g {
		for c := range g[r] {
			want := -1
			if t != nil && r < len(t) && c < len(t[r]) {
				want = t[r][c]
			}
			if ch := g[r][c]; ch == Blank {
				// A blank carries no color, and forcing one would end the run
				// on every gap in the drawing.
				want = cur
			}
			if want != cur {
				closeSpan()
				cur = want
				if cur >= 0 {
					fmt.Fprintf(&b, `<span style="color:%s">`, html.EscapeString(colors[cur%len(colors)]))
					open = true
				}
			}
			b.WriteString(html.EscapeString(string(g[r][c])))
		}
		if r+1 < len(g) {
			b.WriteByte('\n')
		}
	}
	closeSpan()
	return b.String()
}
