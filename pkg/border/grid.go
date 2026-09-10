package border

import "strings"

// Grid is a rectangular block of box characters — a figure, or a whole frame.
// Rows are kept equal length so every transform is total.
type Grid [][]rune

// NewGrid pads a set of lines into a rectangle.
func NewGrid(lines []string) Grid {
	w := 0
	for _, l := range lines {
		if n := len([]rune(l)); n > w {
			w = n
		}
	}
	g := make(Grid, len(lines))
	for i, l := range lines {
		rr := []rune(l)
		g[i] = make([]rune, w)
		for j := range g[i] {
			if j < len(rr) {
				g[i][j] = rr[j]
			} else {
				g[i][j] = Blank
			}
		}
	}
	return g
}

// Blank grid of a given size.
func BlankGrid(w, h int) Grid {
	g := make(Grid, h)
	for i := range g {
		g[i] = make([]rune, w)
		for j := range g[i] {
			g[i][j] = Blank
		}
	}
	return g
}

// W and H are the grid's size in cells.
func (g Grid) W() int {
	if len(g) == 0 {
		return 0
	}
	return len(g[0])
}
func (g Grid) H() int { return len(g) }

// String renders the grid, trailing blanks trimmed.
func (g Grid) String() string {
	var b strings.Builder
	for i, row := range g {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(strings.TrimRight(string(row), string(Blank)))
	}
	return b.String()
}

// Lines is String split back up, for callers that want the rows.
func (g Grid) Lines() []string { return strings.Split(g.String(), "\n") }

// RotateCW turns the grid a quarter turn clockwise — cells AND glyphs, so a
// corner still points the way it did relative to the figure.
func (g Grid) RotateCW() Grid {
	h, w := g.H(), g.W()
	out := BlankGrid(h, w)
	for i := 0; i < w; i++ {
		for j := 0; j < h; j++ {
			out[i][j] = RotateGlyphCW(g[h-1-j][i])
		}
	}
	return out
}

// Rotate turns the grid n quarter turns clockwise (negative for anticlockwise).
func (g Grid) Rotate(n int) Grid {
	n = ((n % 4) + 4) % 4
	out := g
	for i := 0; i < n; i++ {
		out = out.RotateCW()
	}
	return out
}

// MirrorH and MirrorV reflect the grid, glyphs included.
func (g Grid) MirrorH() Grid {
	h, w := g.H(), g.W()
	out := BlankGrid(w, h)
	for i := 0; i < h; i++ {
		for j := 0; j < w; j++ {
			out[i][j] = MirrorGlyphH(g[i][w-1-j])
		}
	}
	return out
}

func (g Grid) MirrorV() Grid {
	h, w := g.H(), g.W()
	out := BlankGrid(w, h)
	for i := 0; i < h; i++ {
		for j := 0; j < w; j++ {
			out[i][j] = MirrorGlyphV(g[h-1-i][j])
		}
	}
	return out
}

// Stamp draws src onto g at (x,y), JOINING what is already there rather than
// replacing it. See MergeGlyph: overwriting is what turns a crossing into a
// break.
func (g Grid) Stamp(src Grid, x, y int) {
	for i := range src {
		for j := range src[i] {
			if src[i][j] == Blank {
				continue
			}
			yy, xx := y+i, x+j
			if yy < 0 || yy >= g.H() || xx < 0 || xx >= g.W() {
				continue
			}
			g[yy][xx] = MergeGlyph(g[yy][xx], src[i][j])
		}
	}
}

// RuleH and RuleV draw a straight run, joining as they go. A border is a
// rectangle of line with figures threaded onto it; these draw the line.
func (g Grid) RuleH(x0, x1, y int) {
	for x := x0; x < x1; x++ {
		if y >= 0 && y < g.H() && x >= 0 && x < g.W() {
			g[y][x] = MergeGlyph(g[y][x], '─')
		}
	}
}

func (g Grid) RuleV(y0, y1, x int) {
	for y := y0; y < y1; y++ {
		if y >= 0 && y < g.H() && x >= 0 && x < g.W() {
			g[y][x] = MergeGlyph(g[y][x], '│')
		}
	}
}

// Canonical is the grid's smallest form under the eight symmetries of the
// square. Two figures share it exactly when one is the other turned or
// mirrored, which is what makes a catalog of distinct figures possible —
// without it the even moduli alone look like hundreds of different drawings
// when they are a handful seen from different angles.
func (g Grid) Canonical() string {
	best := ""
	cur := g
	for m := 0; m < 2; m++ {
		for r := 0; r < 4; r++ {
			if s := cur.String(); best == "" || s < best {
				best = s
			}
			cur = cur.RotateCW()
		}
		cur = cur.MirrorH()
	}
	return best
}

// Stubs are the cells with exactly one arm: the loose ends of the path. An open
// figure has two; a closed one has none, which is why a closed figure makes a
// corner that a run can simply stop against.
func (g Grid) Stubs() [][2]int {
	var out [][2]int
	for i := range g {
		for j := range g[i] {
			a, ok := Arms(g[i][j])
			if !ok || a == 0 {
				continue
			}
			if a&(a-1) == 0 {
				out = append(out, [2]int{i, j})
			}
		}
	}
	return out
}

// Components counts the separate pieces of the drawing. Two neighboring cells
// are joined only when BOTH point at each other, so this measures the line, not
// the pixels: one component means the border really is continuous.
func (g Grid) Components() int {
	h, w := g.H(), g.W()
	if h == 0 {
		return 0
	}
	seen := make([][]bool, h)
	for i := range seen {
		seen[i] = make([]bool, w)
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
	var flood func(r, c int)
	flood = func(r, c int) {
		if seen[r][c] {
			return
		}
		seen[r][c] = true
		if linked(r, c, -1, 0, Up, Down) {
			flood(r-1, c)
		}
		if linked(r, c, 1, 0, Down, Up) {
			flood(r+1, c)
		}
		if linked(r, c, 0, -1, Left, Right) {
			flood(r, c-1)
		}
		if linked(r, c, 0, 1, Right, Left) {
			flood(r, c+1)
		}
	}
	n := 0
	for i := range g {
		for j := range g[i] {
			if a, ok := Arms(g[i][j]); !ok || a == 0 || seen[i][j] {
				continue
			}
			n++
			flood(i, j)
		}
	}
	return n
}

// StampOutside draws src at (x,y) but skips any cell inside one of the given
// rectangles, so the drawing is cut off where they are rather than running
// underneath them.
//
// This is how a run stops at a corner. Laying the run down whole and putting
// the corner on top leaves the run's ink inside the corner, showing through it
// and past it — the corner reads as something sitting ON the run instead of the
// place the run ends.
func (g Grid) StampOutside(src Grid, x, y int, keepOut []Rect) {
	for i := range src {
		for j := range src[i] {
			if src[i][j] == Blank {
				continue
			}
			yy, xx := y+i, x+j
			if yy < 0 || yy >= g.H() || xx < 0 || xx >= g.W() {
				continue
			}
			blocked := false
			for _, r := range keepOut {
				if r.Contains(xx, yy) {
					blocked = true
					break
				}
			}
			if blocked {
				continue
			}
			g[yy][xx] = MergeGlyph(g[yy][xx], src[i][j])
		}
	}
}

// Rect is a half-open cell rectangle.
type Rect struct{ X0, Y0, X1, Y1 int }

// Contains reports whether a cell is inside.
func (r Rect) Contains(x, y int) bool { return x >= r.X0 && x < r.X1 && y >= r.Y0 && y < r.Y1 }

// Grow expands a rectangle by n cells on every side.
func (r Rect) Grow(n int) Rect { return Rect{r.X0 - n, r.Y0 - n, r.X1 + n, r.Y1 + n} }

// RotSymmetry is the order of the figure's rotational symmetry: 4 when a
// quarter turn leaves it unchanged, 2 when only a half turn does, 1 otherwise.
//
// A corner is seen four times, once per corner of the frame, each time mirrored
// or turned. A figure with four-fold symmetry looks the same in all of them, so
// the frame reads as one design rather than four rotations of a motif — which
// is what makes a symmetric figure worth hunting for even though it is rarer.
//
// Four-fold symmetry needs a square grid, since a quarter turn swaps the sides.
func (g Grid) RotSymmetry() int {
	s := g.String()
	if g.W() == g.H() && g.RotateCW().String() == s {
		return 4
	}
	if g.Rotate(2).String() == s {
		return 2
	}
	return 1
}

// MirrorAxes counts the reflections that leave the figure unchanged: across,
// down, and the two diagonals. Four means every reflection, which together with
// four-fold rotation is the full symmetry of the square.
func (g Grid) MirrorAxes() int {
	n := 0
	s := g.String()
	if g.MirrorH().String() == s {
		n++
	}
	if g.MirrorV().String() == s {
		n++
	}
	if g.W() == g.H() {
		// The diagonals, as a quarter turn composed with a reflection.
		if g.RotateCW().MirrorH().String() == s {
			n++
		}
		if g.RotateCW().MirrorV().String() == s {
			n++
		}
	}
	return n
}

// Symmetry is the size of the figure's symmetry group, 1 to 8 — the number of
// the eight ways of turning and flipping a square that leave it alone. Eight is
// as symmetric as a figure on a grid can be.
func (g Grid) Symmetry() int {
	s := g.String()
	n := 0
	cur := g
	for m := 0; m < 2; m++ {
		for r := 0; r < 4; r++ {
			if cur.String() == s {
				n++
			}
			cur = cur.RotateCW()
		}
		cur = cur.MirrorH()
	}
	return n
}

// Silhouette is the region a figure occupies: for every row, the span from its
// leftmost to its rightmost mark, and likewise down every column.
//
// A figure's bounding BOX is not its shape — most of the box is empty — so
// clipping a run against the box cuts it off in blank space well before the
// figure, and clipping against the marks alone barely cuts it at all. The
// silhouette is the outline the eye reads as the figure's extent, and cutting
// there is what makes a run look stopped BY the corner.
func (g Grid) Silhouette() [][]bool {
	h, w := g.H(), g.W()
	m := make([][]bool, h)
	for i := range m {
		m[i] = make([]bool, w)
	}
	inked := func(r, c int) bool {
		a, ok := Arms(g[r][c])
		return ok && a != 0
	}
	for r := 0; r < h; r++ {
		lo, hi := -1, -1
		for c := 0; c < w; c++ {
			if inked(r, c) {
				if lo < 0 {
					lo = c
				}
				hi = c
			}
		}
		for c := lo; c >= 0 && c <= hi; c++ {
			m[r][c] = true
		}
	}
	// Intersected with the same by column, so a figure shaped like a cross does
	// not swallow the empty quadrants between its arms.
	for c := 0; c < w; c++ {
		lo, hi := -1, -1
		for r := 0; r < h; r++ {
			if inked(r, c) {
				if lo < 0 {
					lo = r
				}
				hi = r
			}
		}
		for r := 0; r < h; r++ {
			if r < lo || r > hi || lo < 0 {
				m[r][c] = false
			}
		}
	}
	return m
}

// StampClipped draws src at (x,y) but drops any cell falling inside the mask,
// EXCEPT where the destination already carries a mark — there the two merge, so
// the run ends joined to what stopped it rather than merely touching it.
func (g Grid) StampClipped(src Grid, x, y int, mask [][]bool, mx, my int) {
	inMask := func(gx, gy int) bool {
		r, c := gy-my, gx-mx
		return r >= 0 && r < len(mask) && c >= 0 && c < len(mask[r]) && mask[r][c]
	}
	for i := range src {
		for j := range src[i] {
			if src[i][j] == Blank {
				continue
			}
			yy, xx := y+i, x+j
			if yy < 0 || yy >= g.H() || xx < 0 || xx >= g.W() {
				continue
			}
			if a, ok := Arms(g[yy][xx]); inMask(xx, yy) && (!ok || a == 0) {
				continue
			}
			g[yy][xx] = MergeGlyph(g[yy][xx], src[i][j])
		}
	}
}

// Inside is the clear rectangle in the middle of the grid — where whatever the
// border is around goes.
//
// A border is only useful if something can be put in it, and the caller cannot
// work this out from the Spec: the frame's size falls out of the figures, the
// runs are woven so their inner edge is ragged rather than straight, and the
// corners reach further in than the runs do. The only reliable answer is to
// look at the marks.
//
// Grown from the middle, one side at a time, and each side stops for good the
// first time the strip it would add is not empty. Shrinking the whole grid down
// instead does not work, and the way it fails is worth recording: the first
// pass measures rows that lie INSIDE the top and bottom runs, where there is
// ink all the way to the middle, so the left and right edges are dragged to the
// center line and the rectangle collapses before the vertical bounds have been
// found. There is no ordering that fixes that — the rows to measure are the
// ones inside the answer.
func (g Grid) Inside() Rect {
	w, h := g.W(), g.H()
	if w == 0 || h == 0 {
		return Rect{}
	}
	inked := func(x, y int) bool {
		if y < 0 || y >= len(g) || x < 0 || x >= len(g[y]) {
			return false
		}
		a, ok := Arms(g[y][x])
		return ok && a != 0
	}
	cx, cy := w/2, h/2
	if inked(cx, cy) {
		return Rect{cx, cy, cx, cy}
	}
	r := Rect{cx, cy, cx + 1, cy + 1}
	// Four sides, each with a flag saying whether it can still move. Rotating
	// between them rather than exhausting one at a time keeps the rectangle
	// roughly centered, which matters because the caller is going to put
	// something in the middle of it.
	open := [4]bool{true, true, true, true}
	for open[0] || open[1] || open[2] || open[3] {
		for side := range 4 {
			if !open[side] {
				continue
			}
			clear := true
			switch side {
			case 0: // left
				for y := r.Y0; y < r.Y1 && clear; y++ {
					clear = r.X0 > 0 && !inked(r.X0-1, y)
				}
			case 1: // right
				for y := r.Y0; y < r.Y1 && clear; y++ {
					clear = r.X1 < w && !inked(r.X1, y)
				}
			case 2: // top
				for x := r.X0; x < r.X1 && clear; x++ {
					clear = r.Y0 > 0 && !inked(x, r.Y0-1)
				}
			case 3: // bottom
				for x := r.X0; x < r.X1 && clear; x++ {
					clear = r.Y1 < h && !inked(x, r.Y1)
				}
			}
			if !clear {
				open[side] = false
				continue
			}
			switch side {
			case 0:
				r.X0--
			case 1:
				r.X1++
			case 2:
				r.Y0--
			case 3:
				r.Y1++
			}
		}
	}
	return r
}

// Overlay writes lines of text into the grid, replacing whatever is there.
//
// Replacing, not merging: content laid over a border has to cover it, and the
// arm algebra that joins two figures would happily weld a table rule to a run.
// Anything past the edge is dropped rather than wrapped, because a border sized
// to its content is the caller's problem and silently rewrapping hides it.
func (g Grid) Overlay(x, y int, lines []string) {
	for i, line := range lines {
		r := y + i
		if r < 0 || r >= len(g) {
			continue
		}
		for j, ch := range []rune(line) {
			if c := x + j; c >= 0 && c < len(g[r]) {
				g[r][c] = ch
			}
		}
	}
}
