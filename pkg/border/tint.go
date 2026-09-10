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
// Colors are pisano's own, from PassColors, so a bordered page and a pisano
// sheet in the same theme are drawn from the same six.

// Tint is a palette index per cell, or -1 where there is no mark.
type Tint [][]int

// TintAlong walks the grid's largest connected line and colors each cell by how
// far along it that cell falls, cycling the palette once per circuit.
//
// Cells not on that line — a stray piece Prune left, or a figure the border
// does not reach — come back -1 rather than being forced into a color, so a
// caller can see there was something there to explain.
func (g Grid) TintAlong(colors int) Tint {
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
	lab, n := g.label()
	if n == 0 {
		return t
	}
	size := make([]int, n)
	for r := range lab {
		for _, v := range lab[r] {
			if v >= 0 {
				size[v]++
			}
		}
	}
	best := 0
	for i, s := range size {
		if s > size[best] {
			best = i
		}
	}
	// Walk it in traversal order. Depth first along the line, so consecutive
	// cells of the border come out consecutive — a breadth-first walk would
	// spread out from the start in both directions at once and color the two
	// halves of the frame the same, which is not what going round it looks
	// like.
	var start [2]int
	found := false
	for r := 0; r < h && !found; r++ {
		for c := 0; c < w; c++ {
			if lab[r][c] == best {
				start, found = [2]int{r, c}, true
				break
			}
		}
	}
	if !found {
		return t
	}
	linked := func(r, c, dr, dc int, mine, theirs Arm) bool {
		r2, c2 := r+dr, c+dc
		if r2 < 0 || r2 >= h || c2 < 0 || c2 >= w {
			return false
		}
		a, ok1 := Arms(g[r][c])
		b, ok2 := Arms(g[r2][c2])
		return ok1 && ok2 && a&mine != 0 && b&theirs != 0
	}
	seen := make([][]bool, h)
	for i := range seen {
		seen[i] = make([]bool, w)
	}
	order := make([][2]int, 0, size[best])
	stack := [][2]int{start}
	for len(stack) > 0 {
		p := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		r, c := p[0], p[1]
		if seen[r][c] {
			continue
		}
		seen[r][c] = true
		order = append(order, p)
		if linked(r, c, -1, 0, Up, Down) {
			stack = append(stack, [2]int{r - 1, c})
		}
		if linked(r, c, 1, 0, Down, Up) {
			stack = append(stack, [2]int{r + 1, c})
		}
		if linked(r, c, 0, -1, Left, Right) {
			stack = append(stack, [2]int{r, c - 1})
		}
		if linked(r, c, 0, 1, Right, Left) {
			stack = append(stack, [2]int{r, c + 1})
		}
	}
	for i, p := range order {
		t[p[0]][p[1]] = i * colors / max(1, len(order)) % colors
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
		p := pisano.PassColors(pisano.ThemeDark)
		colors = p[:]
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
