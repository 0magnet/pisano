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
	n := 0
	var flood func(r, c int)
	flood = func(r, c int) {
		if lab[r][c] != -1 {
			return
		}
		lab[r][c] = n
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
		// Cells of piece 0, and of everything else, then the nearest pair.
		bestD := 1 << 30
		var br1, bc1, br2, bc2 int
		for r1 := range lab {
			for c1 := range lab[r1] {
				if lab[r1][c1] != 0 {
					continue
				}
				for r2 := range lab {
					for c2 := range lab[r2] {
						if lab[r2][c2] <= 0 {
							continue
						}
						d := abs(r1-r2) + abs(c1-c2)
						if d < bestD {
							bestD, br1, bc1, br2, bc2 = d, r1, c1, r2, c2
						}
					}
				}
			}
		}
		if bestD == 1<<30 {
			return added // nothing to join to; leave it rather than loop
		}
		g.connect(br1, bc1, br2, bc2)
		added++
	}
}
