package border

// Making a composed border into one closed line.
//
// The runs join themselves: consecutive copies sit one travel vector apart, and
// that is exactly the offset which puts one copy's path end on the next copy's
// path start, so a run of any length is a single line. Measured across moduli
// from 9 to 2269, every run comes out as one component.
//
// The corners do not, and cannot. A closed figure has no loose ends to offer,
// and a figure whose path never reaches the edge of its own box — which is most
// of them — has nothing at the boundary to meet anyway. So the joins have to be
// drawn. Join finds what is still separate and draws the shortest connector
// between the pieces, in LATTICE space, so the connector turns with everything
// else and stays part of the same drawing.

// label numbers the cells of each connected piece, -1 where there is nothing.
func (g Grid) label() ([][]int, int) {
	h, w := g.H(), g.W()
	lab := make([][]int, h)
	for i := range lab {
		lab[i] = make([]int, w)
		for j := range lab[i] {
			lab[i][j] = -1
		}
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
	// Flooded with an explicit stack, not by recursion.
	//
	// A border is one long connected line, so a recursive flood recurses once
	// per cell of it — thousands deep on a frame of any size. That is fine on a
	// host and fatal in a browser: tinygo's wasm stack is a few tens of
	// kilobytes, and the store's page decorator hit runtime.runtimeFatal here,
	// through nilPanic, on the first frame it tried to draw. An explicit stack
	// puts the same walk on the heap, where the size is not a limit.
	n := 0
	var stack [][2]int
	flood := func(r0, c0 int) {
		stack = append(stack[:0], [2]int{r0, c0})
		for len(stack) > 0 {
			p := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			r, c := p[0], p[1]
			if lab[r][c] != -1 {
				continue
			}
			lab[r][c] = n
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
	}
	for i := 0; i < h; i++ {
		for j := 0; j < w; j++ {
			if a, ok := Arms(g[i][j]); ok && a != 0 && lab[i][j] == -1 {
				flood(i, j)
				n++
			}
		}
	}
	return lab, n
}

// connect draws an L from (r1,c1) to (r2,c2), merging arms so it joins what it
// meets rather than cutting through it.
func (g Grid) connect(r1, c1, r2, c2 int) {
	lo, hi := c1, c2
	if hi < lo {
		lo, hi = hi, lo
	}
	g.RuleH(lo, hi+1, r1)
	lo, hi = r1, r2
	if hi < lo {
		lo, hi = hi, lo
	}
	g.RuleV(lo, hi+1, c2)
}

// Join draws connectors until the whole drawing is a single line, and reports
// how many it needed. Zero means the pieces already met.
//
// Each pass connects the CLOSEST pair of cells belonging to two different
// pieces, which keeps the added line short and puts it where the pieces were
// nearly touching anyway — at the corners, in practice, since that is the only
// place a border comes apart.
func (g Grid) Join() int {
	added := 0
	for {
		lab, n := g.label()
		if n <= 1 {
			return added
		}
		// Nearest cell of another piece, by a breadth-first walk out from piece
		// zero — linear in the grid.
		//
		// Comparing every cell of one piece against every cell of the others is
		// the obvious way and is quadratic in the number of marks, which on a
		// border a thousand cells on a side means billions of comparisons per
		// join. It ran in well under a second on the small test borders and did
		// not finish at all on a real one.
		h, w := g.H(), g.W()
		type qi struct{ r, c, sr, sc int }
		seen := make([][]bool, h)
		for i := range seen {
			seen[i] = make([]bool, w)
		}
		var q []qi
		for r := 0; r < h; r++ {
			for c := 0; c < w; c++ {
				if lab[r][c] == 0 {
					q = append(q, qi{r, c, r, c})
					seen[r][c] = true
				}
			}
		}
		found := false
		var br1, bc1, br2, bc2 int
		for i := 0; i < len(q) && !found; i++ {
			cur := q[i]
			for _, d := range [4][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
				r2, c2 := cur.r+d[0], cur.c+d[1]
				if r2 < 0 || r2 >= h || c2 < 0 || c2 >= w || seen[r2][c2] {
					continue
				}
				if lab[r2][c2] > 0 {
					br1, bc1, br2, bc2 = cur.sr, cur.sc, r2, c2
					found = true
					break
				}
				seen[r2][c2] = true
				q = append(q, qi{r2, c2, cur.sr, cur.sc})
			}
		}
		if !found {
			return added // nothing reachable to join to
		}
		g.connect(br1, bc1, br2, bc2)
		added++
	}
}

// Prune erases every connected piece smaller than minCells, and reports how
// many cells it removed.
//
// The cut that stops a run at a corner is a straight line, and a wandering path
// can cross a straight line many times — so what is left outside the corner is
// usually the run, plus a few crumbs the cut sheared off it. They are not part
// of the border and they are not worth reconnecting: drawing a connector to
// each is how a frame ends up with thirty straight lines through it, which is
// the one thing a woven border must not have.
//
// Sized against the smallest whole piece rather than by a constant, so it
// scales with the figures and can never eat one. Anything at least half a piece
// is left alone for Join to deal with, on the grounds that something that big
// going missing would be a bug worth seeing rather than tidying away.
func (g Grid) Prune(minCells int) int {
	lab, n := g.label()
	if n <= 1 {
		return 0
	}
	size := make([]int, n)
	for r := range lab {
		for _, v := range lab[r] {
			if v >= 0 {
				size[v]++
			}
		}
	}
	biggest := 0
	for i, s := range size {
		if s > size[biggest] {
			biggest = i
		}
	}
	removed := 0
	for r := range lab {
		for c, v := range lab[r] {
			if v < 0 || v == biggest || size[v] >= minCells {
				continue
			}
			g[r][c] = Blank
			removed++
		}
	}
	return removed
}
