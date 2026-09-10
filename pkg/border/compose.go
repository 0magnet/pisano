package border

import (
	"fmt"
	"math"
)

// Spec says what a border is made of. Counts are in COPIES, not pixels: the
// frame's size falls out of the figures rather than being imposed on them,
// which is what keeps every piece on one lattice.
type Spec struct {
	Across Figure // runs along the top and bottom
	Down   Figure // runs down the left and right
	Corner Figure // sits at the four corners; closed figures make the best ones
	Cols   int    // copies of Across between the corners
	Rows   int    // copies of Down between the corners
	Bias   Bias   // which side of a corner's middle the runs meet it on
	// Detached leaves the runs and the corners as separate pieces: no segment
	// is drawn from a corner out to the run that stops at it, and no connector
	// is drawn to close the frame. The border is then four runs and four
	// corners that meet by proximity rather than by line, which is a look worth
	// having and not a failure to join.
	Detached bool
}

// Bias says which side of a corner's middle the runs come in on.
//
// There is always a side, and this is why: a closed corner's middle is the
// hole its path encircles, and an even-sided corner cannot be centered on an
// odd-width run in any case — the two middles fall half a cell apart. So the
// runs meet a corner just inside or just outside its middle, and the only
// thing left to get right is doing it the same way at all four.
type Bias int

const (
	// Inward puts the runs on the frame's side of each corner's middle, so
	// the corners stand slightly proud of the runs. It is the default
	// because it is what a picture frame does.
	Inward Bias = iota
	// Outward puts the runs on the outside, flushing the frame's outer edge
	// and letting the corners reach in.
	Outward
)

// Role says which part of the frame a piece is.
type Role string

const (
	RoleCorner Role = "corner"
	RoleAcross Role = "across"
	RoleDown   Role = "down"
)

// Placement is one piece at an integer cell position in the UNROTATED lattice.
// Keeping positions on that lattice is what holds every piece in phase: they
// are not separate rotated objects that share an angle, they are one grid seen
// at an angle.
type Placement struct {
	Grid   Grid
	GX, GY int
	Role   Role
	// MeetX, MeetY are where a corner expects its runs to arrive, in the
	// piece's own cells. GridJoins needs them to draw the short segment from
	// the corner's outermost mark out to the run waiting beyond it; they are
	// meaningless on a run.
	MeetX, MeetY int
}

// Layout is a composed border: where every piece goes, and the single rotation
// that turns the whole lattice to lie along the frame's edges.
type Layout struct {
	Places []Placement
	Turn   float64 // degrees clockwise, applied to every piece alike
	Spec   Spec
	// DownAngle is where the Down figure ends up under Turn. It should be 90:
	// square to the Across run. Anything else means the two runs cannot share a
	// lattice, and the border will look subtly out of register however good
	// each run looks alone.
	DownAngle float64
}

// Compose lays out a border.
//
// The rotation is whatever turns the Across figure's travel horizontal. Every
// piece — runs and corners alike — is turned by that same angle, which is the
// difference between pieces that merely share an orientation and pieces whose
// cell grids actually line up.
func Compose(s Spec) (Layout, error) {
	if s.Cols < 1 || s.Rows < 1 {
		return Layout{}, fmt.Errorf("border: need at least one copy each way, got %dx%d", s.Cols, s.Rows)
	}
	if !s.Across.Travels() {
		return Layout{}, fmt.Errorf("border: the across figure (mod %d) does not travel, so copies of it pile up instead of running", s.Across.Mod)
	}
	if !s.Down.Travels() {
		return Layout{}, fmt.Errorf("border: the down figure (mod %d) does not travel", s.Down.Mod)
	}
	turn := -math.Atan2(float64(s.Across.DY), float64(s.Across.DX))
	down := math.Atan2(float64(s.Down.DY), float64(s.Down.DX)) + turn

	l := Layout{
		Turn:      turn * 180 / math.Pi,
		DownAngle: down * 180 / math.Pi,
		Spec:      s,
	}
	add := func(g Grid, gx, gy int, r Role) {
		l.Places = append(l.Places, Placement{Grid: g, GX: gx, GY: gy, Role: r})
	}
	ax, ay := s.Across.DX, s.Across.DY
	dx, dy := s.Down.DX, s.Down.DY

	// Corners sit where the two runs' center LINES cross, and each of the four
	// is a mirror of the first.
	//
	// The crossing is the point both runs head away from, so it is the one
	// place a corner can sit and have both runs aimed at it. What it cannot be
	// is the corner's exact middle, because a closed corner's middle is the
	// hole its path encircles — MeetRow and MeetCol pick the marked row and
	// column just to one side of that hole, and the corner is hung so those
	// land on the crossing.
	//
	// Mirroring is the other half, and skipping it is what made the finished
	// frame look lopsided even when every piece was right. One offset used at
	// all four corners is the same offset in ABSOLUTE terms, so if the top run
	// sits half a cell below its corners' middles then so does the bottom run —
	// which puts the top one inside its corners and the bottom one outside.
	// Reflecting the corner grid and its meeting point together makes the four
	// corners reflections of each other, which is what a frame is.
	crossX, crossY := runCross(s.Across, s.Down)
	// Which way is "inward" depends on where the opposite corners are, and that
	// is a property of the travels, not of the page: a down figure traveling
	// +0,-4 builds its frame UPWARD from the first corner, so the run that ends
	// up along the top of the picture is the one at the far end. The frame's
	// middle sits at half the sum of the two spans, so the sign of that sum is
	// the direction to lean.
	inward := s.Bias == Inward
	below := (s.Cols*ay+s.Rows*dy > 0) == inward
	right := (s.Cols*ax+s.Rows*dx > 0) == inward
	mr := s.Corner.Grid.MeetRow(below)
	mc := s.Corner.Grid.MeetCol(right)
	cw, ch := s.Corner.Grid.W(), s.Corner.Grid.H()
	for _, q := range [4]struct{ flipX, flipY bool }{
		{false, false}, {true, false}, {false, true}, {true, true},
	} {
		cg := s.Corner.Grid
		mx, my := mc, mr
		if q.flipX {
			cg, mx = cg.MirrorH(), cw-1-mc
		}
		if q.flipY {
			cg, my = cg.MirrorV(), ch-1-mr
		}
		px, py := crossX, crossY
		if q.flipX {
			px += s.Cols * ax
			py += s.Cols * ay
		}
		if q.flipY {
			px += s.Rows * dx
			py += s.Rows * dy
		}
		l.Places = append(l.Places, Placement{
			Grid: cg, GX: px - mx, GY: py - my, Role: RoleCorner,
			MeetX: mx, MeetY: my,
		})
	}

	// Runs strictly between them, one travel apart, and the far one placed by
	// REFLECTING the near one rather than by copying it along.
	//
	// Cols copies span corner to corner: copy k starts one travel short of
	// where copy k+1 does, and the last reaches into the far corner. Going one
	// further, to a copy starting AT the far corner, overshoots by a whole
	// figure — a figure is about as wide as it travels, so that copy would hang
	// its entire width outside the frame.
	//
	// The reflection is the same point as mirroring the corners, and it shows
	// once both runs are the same drawing: a figure's ornament points one way,
	// so a translated copy along the bottom has its tips turned into the frame
	// while the top's are turned out of it.
	//
	// Mirroring the GRID in place is not enough, and measuring is the only way
	// to know — it left 540 cells disagreeing with the frame's own mirror. A
	// mirrored grid occupies the columns it always did, which is not where the
	// reflection of the near run lands; the piece has to be positioned by
	// reflecting it about the frame's middle, which is the crossing plus half
	// of both spans. Twice that middle is an integer even when the middle is
	// not, so the arithmetic stays exact.
	twoCX := 2*crossX + s.Cols*ax + s.Rows*dx
	twoCY := 2*crossY + s.Cols*ay + s.Rows*dy
	// Only when the frame's own axes are the lattice's. A frame turned to lie
	// along a diagonal has its mirror lines turned with it, and a turned line
	// is not one a grid can be reflected in — the far run is then laid down by
	// translation, as it always was.
	square := (ax == 0 || ay == 0) && (dx == 0 || dy == 0)
	for k := 0; k < s.Cols; k++ {
		add(s.Across.Grid, k*ax, k*ay, RoleAcross)
		if square {
			add(s.Across.Grid.MirrorV(), k*ax, twoCY-k*ay-(s.Across.H()-1), RoleAcross)
			continue
		}
		add(s.Across.Grid, k*ax+s.Rows*dx, k*ay+s.Rows*dy, RoleAcross)
	}
	for m := 0; m < s.Rows; m++ {
		add(s.Down.Grid, m*dx, m*dy, RoleDown)
		if square {
			add(s.Down.Grid.MirrorH(), twoCX-m*dx-(s.Down.W()-1), m*dy, RoleDown)
			continue
		}
		add(s.Down.Grid, s.Cols*ax+m*dx, s.Cols*ay+m*dy, RoleDown)
	}
	return l, nil
}

// Square reports whether the two runs are square to each other, within tol
// degrees. A layout that is not square still renders; it just will not look
// like a rectangle.
func (l Layout) Square(tol float64) bool {
	d := math.Mod(math.Abs(l.DownAngle), 180)
	return math.Abs(d-90) <= tol
}

// Upright reports whether the whole layout sits on the axis, needing no
// rotation — which is exactly when it can be drawn into a grid and printed to
// a terminal. An angled layout has to go to HTML, where a rotation is available.
func (l Layout) Upright(tol float64) bool {
	t := math.Mod(math.Abs(l.Turn), 90)
	return t <= tol || 90-t <= tol
}

// Bounds is the lattice extent of the placed pieces, in cells.
func (l Layout) Bounds() (minX, minY, maxX, maxY int) {
	if len(l.Places) == 0 {
		return
	}
	minX, minY = math.MaxInt32, math.MaxInt32
	maxX, maxY = math.MinInt32, math.MinInt32
	for _, p := range l.Places {
		if p.GX < minX {
			minX = p.GX
		}
		if p.GY < minY {
			minY = p.GY
		}
		if p.GX+p.Grid.W() > maxX {
			maxX = p.GX + p.Grid.W()
		}
		if p.GY+p.Grid.H() > maxY {
			maxY = p.GY + p.Grid.H()
		}
	}
	return
}

// Grid composes the whole border into one grid, in lattice space, and joins it
// into a single line.
//
// This is the drawing; the layout's rotation is only how it is presented. So
// connectivity is settled here, once, for both renderers — an angled border is
// this same grid turned, and turning cannot break a line.
//
// The runs need no help: consecutive copies sit one travel apart, which is
// exactly the offset that lands one copy's path end on the next one's start.
// What needs joining is the corners, where a closed figure has no loose end to
// offer.
func (l Layout) Grid() Grid {
	g, _ := l.GridJoins()
	return g
}

// GridJoins is Grid, also reporting how many connectors had to be drawn. Zero
// means the figures met on their own, which is the aim: a connector is a
// straight line drawn between two figures, and it looks like one.
func (l Layout) GridJoins() (Grid, int) {
	minX, minY, maxX, maxY := l.Bounds()
	// A margin, so a connector that wants to step outside the pieces has room.
	const pad = 2
	g := BlankGrid(maxX-minX+2*pad, maxY-minY+2*pad)

	// Corners first, then each run cut ACROSS ITSELF where a corner stands in
	// its way, then a short segment out of the corner to meet it.
	//
	// The cut is a straight line square to the run's travel, at the near edge of
	// the corner's box, and getting there took three wrong answers. Stamping
	// the corners over whole runs leaves run ink inside and beyond them, so a
	// corner reads as an ornament laid on a line that carries on underneath.
	// Cutting to the corner's OUTLINE instead follows the figure's own ragged
	// shape, which shreds a thick run into fragments that reach round the
	// corner on both sides — a fifty-cell band came apart into forty pieces,
	// each needing a connector of its own. Keeping run cells that coincide with
	// the corner's ink repairs that beautifully on a sparse corner and fails
	// completely on a dense one: modulus 31 at four passes carries 117 marks in
	// 64 cells, so nearly the whole run coincides with something and survives
	// inside the corner, which is exactly the look of passing straight through.
	//
	// A straight cut has none of those problems. It leaves the run with a clean
	// square end whatever the corner's shape, it cannot divide the run into
	// pieces, and it works the same for a band one cell thick and one fifty
	// cells thick.
	//
	// Each run is cut only by the direction it travels in. A corner blocks the
	// across run over the span it occupies ALONG the across travel, and blocks
	// the down run over its span along the down travel; the two masks have to
	// be separate, or the top run would be cut off by the corners at the far
	// side of the frame as well.
	ua := unit(l.Spec.Across.DX, l.Spec.Across.DY)
	ud := unit(l.Spec.Down.DX, l.Spec.Down.DY)
	var aSpans, dSpans [][2]float64
	for _, p := range l.Places {
		if p.Role != RoleCorner {
			continue
		}
		ox, oy := p.GX-minX+pad, p.GY-minY+pad
		g.Stamp(p.Grid, ox, oy)
		aSpans = append(aSpans, boxSpan(ox, oy, p.Grid.W(), p.Grid.H(), ua))
		dSpans = append(dSpans, boxSpan(ox, oy, p.Grid.W(), p.Grid.H(), ud))
	}
	for _, p := range l.Places {
		if p.Role == RoleCorner {
			continue
		}
		u, spans := ua, aSpans
		if p.Role == RoleDown {
			u, spans = ud, dSpans
		}
		g.stampCut(p.Grid, p.GX-minX+pad, p.GY-minY+pad, u, spans)
	}
	// Loose ends go BEFORE the segments are drawn, so a segment reaches what is
	// left rather than what was about to be rubbed out.
	//
	// A straight cut through a woven run leaves a fringe: cells whose arms were
	// reaching for the cells the cut took, now pointing at nothing. Trim takes
	// exactly one cell off each, which is the whole of the fringe and none of
	// the run.
	g.Trim()
	if l.Spec.Detached {
		g.Prune(smallestPiece(l.Places) / 2)
		g.Smooth()
		return g, 0
	}
	for _, p := range l.Places {
		if p.Role == RoleCorner {
			reachOut(g, p.Grid, p.GX-minX+pad, p.GY-minY+pad, p.MeetX, p.MeetY)
		}
	}
	g.Prune(smallestPiece(l.Places) / 2)
	// Smoothing goes last of all. Join and Prune both leave dead arms behind —
	// one by drawing into a cell, the other by emptying one — so anything that
	// ran before them would have its work undone.
	n := g.Join()
	g.Smooth()
	return g, n
}

// unit is the travel direction, or +x for a figure that does not travel.
func unit(dx, dy int) [2]float64 {
	t := math.Hypot(float64(dx), float64(dy))
	if t == 0 {
		return [2]float64{1, 0}
	}
	return [2]float64{float64(dx) / t, float64(dy) / t}
}

// boxSpan is how far a piece reaches along one direction: the projections of
// its four box corners onto u, low and high.
func boxSpan(x, y, w, h int, u [2]float64) [2]float64 {
	lo, hi := math.MaxFloat64, -math.MaxFloat64
	for _, c := range [4][2]int{{x, y}, {x + w - 1, y}, {x, y + h - 1}, {x + w - 1, y + h - 1}} {
		d := float64(c[0])*u[0] + float64(c[1])*u[1]
		lo, hi = math.Min(lo, d), math.Max(hi, d)
	}
	return [2]float64{lo, hi}
}

// stampCut draws src at (x,y), dropping any cell whose position along u falls
// inside one of the spans — the straight square-ended cut described in
// GridJoins.
func (g Grid) stampCut(src Grid, x, y int, u [2]float64, spans [][2]float64) {
	for i := range src {
		for j := range src[i] {
			if src[i][j] == Blank {
				continue
			}
			yy, xx := y+i, x+j
			if yy < 0 || yy >= g.H() || xx < 0 || xx >= g.W() {
				continue
			}
			d := float64(xx)*u[0] + float64(yy)*u[1]
			blocked := false
			for _, s := range spans {
				if d >= s[0] && d <= s[1] {
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

// FitSpec chooses the copy counts that fill a box of the given size, at roughly
// the given cell size in pixels.
//
// This is the piece a border needs to be DYNAMIC. Everything else about a
// composition is fixed by the figures, but the counts depend on the box, and a
// box on a page is not known until it is measured — it changes with the
// content, the window and the font. Asking a caller for cols and rows works for
// a command run by hand and not at all for a page that resizes.
//
// After the turn, one copy of the across figure advances the length of its own
// travel along the top, and likewise the down figure down the side, so the
// counts are just the box divided by those lengths. Both are floored at one:
// a box too small for even a single copy still gets a border, just a cramped
// one, which is a better answer than none.
func FitSpec(across, down, corner Figure, boxW, boxH, cell float64) Spec {
	step := func(f Figure) float64 {
		s := math.Hypot(float64(f.DX), float64(f.DY)) * cell
		if s <= 0 {
			return cell
		}
		return s
	}
	// The corners sit at the ends of the runs and stick out past them, so the
	// runs have less room than the box. Subtracting one corner's turned extent
	// accounts for both ends together — half of it protrudes at each. Without
	// this every border came out a corner's width too big, consistently.
	corn := math.Hypot(float64(corner.W()), float64(corner.H())) * cell
	cols := int(math.Round(math.Max(boxW-corn, step(across)) / step(across)))
	rows := int(math.Round(math.Max(boxH-corn, step(down)) / step(down)))
	if cols < 1 {
		cols = 1
	}
	if rows < 1 {
		rows = 1
	}
	return Spec{Across: across, Down: down, Corner: corner, Cols: cols, Rows: rows}
}

// CellFor is the cell size, in pixels, at which this layout fills a box —
// the counterpart to FitSpec, for when the counts are already chosen.
func (l Layout) CellFor(boxW, boxH float64) float64 {
	g := l.Grid()
	x0, y0, x1, y1 := l.inkBox(g)
	w, h := x1-x0, y1-y0
	if w <= 0 || h <= 0 {
		return 0
	}
	return math.Min(boxW/w, boxH/h)
}

// runCross is the lattice point where the across run's line meets the down
// run's, in the corner where both begin.
//
// The lines are the runs' CENTER lines, from Figure.Axis — not the lines
// through StartX,StartY, which is only where each path happens to begin and can
// sit at the very edge of the stripe the run draws. Modulus 13 is the case that
// showed it: five cells wide, path starting at column 4, so a corner hung on
// the start sat two cells off the line the run actually draws.
//
// A line rather than a ray: the crossing is usually a little way BACK from
// where either run starts, which is exactly right — that is the point both runs
// head away from, and so the point a corner has to sit on.
//
// Parallel runs have no crossing. That is not a real border — the sides would
// lie along the top — and Square already reports it, so the across figure's own
// axis is a harmless answer to give back.
func runCross(across, down Figure) (x, y int) {
	ax, ay := float64(across.DX), float64(across.DY)
	dx, dy := float64(down.DX), float64(down.DY)
	pax, pay := across.Axis()
	det := dx*ay - ax*dy
	if det == 0 {
		return int(math.Round(pax)), int(math.Round(pay))
	}
	pdx, pdy := down.Axis()
	t := (dx*(pay-pdy) - dy*(pax-pdx)) / det
	return int(math.Round(pax - t*ax)), int(math.Round(pay - t*ay))
}

// reach is how far past a corner a connector will look for the run it belongs
// to, as a multiple of the corner's own size. The gap the clip leaves is a
// property of the corner's outline, so it scales with the corner rather than
// with the frame.
const reach = 1.0

// reachOut draws the short segment from a corner's outermost mark to the run
// waiting outside it, along the row and column the corner was hung on.
//
// This is the piece the hard clip makes necessary, and the piece that lets the
// clip be hard. Cutting a run at a corner's outline leaves it standing off a
// little — the outline is a rectangle and the figure inside it is not — so
// without this the two are near neighbors that never touch, and Join later
// bridges them with a connector chosen by distance rather than by design.
// Drawn here it goes where the frame means it to: straight along the marked row
// or column beside the corner's middle, which is the line the corner was hung
// on in the first place.
//
// All four directions are tried at every corner. Two of them have nothing to
// find and cost a short walk to discover it, which is cheaper than working out
// which of the four corners this is.
//
// Both ends of the segment are drawn, not just the cells between them. An arm
// has to be added at each end for the two to be linked — a run cell that
// happens to be a cross is joined either way, one that is an elbow pointing the
// wrong way is not, and a connector that works only when it lands on a cross is
// a connector that works most of the time.
func reachOut(g, cg Grid, ox, oy, mx, my int) int {
	lim := int(math.Round(reach*float64(cg.W()+cg.H())/2)) + 2
	marked := func(x, y int) bool {
		if y < 0 || y >= g.H() || x < 0 || x >= g.W() {
			return false
		}
		a, ok := Arms(g[y][x])
		return ok && a != 0
	}
	n := 0
	if lo, hi, ok := cg.EdgeInk(my, true); ok {
		for x := ox + hi + 1; x <= ox+hi+lim; x++ {
			if marked(x, oy+my) {
				g.RuleH(ox+hi, x+1, oy+my)
				n++
				break
			}
		}
		for x := ox + lo - 1; x >= ox+lo-lim; x-- {
			if marked(x, oy+my) {
				g.RuleH(x, ox+lo+1, oy+my)
				n++
				break
			}
		}
	}
	if lo, hi, ok := cg.EdgeInk(mx, false); ok {
		for y := oy + hi + 1; y <= oy+hi+lim; y++ {
			if marked(ox+mx, y) {
				g.RuleV(oy+hi, y+1, ox+mx)
				n++
				break
			}
		}
		for y := oy + lo - 1; y >= oy+lo-lim; y-- {
			if marked(ox+mx, y) {
				g.RuleV(y, oy+lo+1, ox+mx)
				n++
				break
			}
		}
	}
	return n
}

// smallestPiece is the ink count of the least of the layout's pieces, which is
// the scale Prune measures crumbs against.
func smallestPiece(ps []Placement) int {
	best := 0
	for _, p := range ps {
		n := 0
		for _, row := range p.Grid {
			for _, ch := range row {
				if a, ok := Arms(ch); ok && a != 0 {
					n++
				}
			}
		}
		if n > 0 && (best == 0 || n < best) {
			best = n
		}
	}
	return best
}
