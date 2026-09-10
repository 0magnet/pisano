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
}

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

	// Corners sit where the two runs' LINES CROSS, centered on that point.
	//
	// A run does not enter its figure at the figure's corner: its path starts
	// at StartX,StartY, so the line a run draws along the top of the frame is
	// the line through that point in the direction it travels. The two runs
	// meeting at a frame corner are two such lines, and they cross at one
	// place. Putting the corner figure's middle there is what makes both runs
	// head INTO it: each is aimed at the middle and stops at the first mark it
	// meets, which is the mark nearest the middle along that line.
	//
	// Aiming at the nearest inked cell to the middle instead — which is what
	// this did before — lines the corner up with ONE of the runs and lets the
	// other arrive wherever it happens to. On a ring-shaped corner that put
	// both runs onto the same side of the ring, so the frame's corner read as
	// a bend in one line with an ornament stuck to it.
	crossX, crossY := runCross(s.Across, s.Down)
	cx := crossX - s.Corner.Grid.W()/2
	cy := crossY - s.Corner.Grid.H()/2
	add(s.Corner.Grid, cx, cy, RoleCorner)
	add(s.Corner.Grid, s.Cols*ax+cx, s.Cols*ay+cy, RoleCorner)
	add(s.Corner.Grid, s.Rows*dx+cx, s.Rows*dy+cy, RoleCorner)
	add(s.Corner.Grid, s.Cols*ax+s.Rows*dx+cx, s.Cols*ay+s.Rows*dy+cy, RoleCorner)

	// Runs strictly between them, one travel apart.
	// Runs from the first corner up to the last, so the final copy STARTS one
	// travel before the far corner and reaches into it — as much of the run as
	// can meet the corner does, and none of it carries on past.
	//
	// Going one further, to a copy starting AT the far corner, overshoots by a
	// whole figure: a figure is about as wide as it travels, so that copy hangs
	// its entire width outside the frame.
	for k := 0; k < s.Cols; k++ {
		add(s.Across.Grid, k*ax, k*ay, RoleAcross)
		add(s.Across.Grid, k*ax+s.Rows*dx, k*ay+s.Rows*dy, RoleAcross)
	}
	for m := 0; m < s.Rows; m++ {
		add(s.Down.Grid, m*dx, m*dy, RoleDown)
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
	// Corners first, then the runs clipped against them: a run keeps every
	// cell outside a corner and every cell that lands ON one of its marks, so
	// it is cut off at the corner's outline and joined where it meets a
	// stroke. The corners are where the runs STOP, and this is what makes
	// them look it.
	//
	// Two earlier attempts are worth knowing about, because both look
	// plausible. Drawing the runs whole and stamping the corners over them
	// leaves run ink inside and beyond the corner, so the corner reads as an
	// ornament laid on a line that carries on underneath. Clipping against
	// the corner's bounding BOX instead cuts the run off in blank space well
	// short of any mark — which is what used to force connectors to be drawn,
	// and a drawn connector looks like the straight line it is.
	//
	// ONE mask covering every corner, not one pass per corner: clipping a run
	// separately against each would draw, on the pass for corner B, exactly
	// the cells corner A had just excluded, and nothing would be cut at all.
	blocked := make([][]bool, g.H())
	for i := range blocked {
		blocked[i] = make([]bool, g.W())
	}
	for _, p := range l.Places {
		if p.Role != RoleCorner {
			continue
		}
		ox, oy := p.GX-minX+pad, p.GY-minY+pad
		g.Stamp(p.Grid, ox, oy)
		sil := p.Grid.Silhouette()
		for r := range sil {
			for c := range sil[r] {
				if !sil[r][c] {
					continue
				}
				if yy, xx := oy+r, ox+c; yy >= 0 && yy < g.H() && xx >= 0 && xx < g.W() {
					blocked[yy][xx] = true
				}
			}
		}
	}
	for _, p := range l.Places {
		if p.Role != RoleCorner {
			g.StampClipped(p.Grid, p.GX-minX+pad, p.GY-minY+pad, blocked, 0, 0)
		}
	}
	return g, g.Join()
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
// Each run is a line rather than a ray for this purpose: the crossing is
// usually a little way BACK from where either run starts, which is exactly
// right — that is the point both runs are heading away from, and so the point
// a corner has to be centered on for them to look like they came out of it.
//
// Parallel runs have no crossing. That is not a real border — the sides would
// lie along the top — and Square already reports it, so the across figure's own
// start is a harmless answer to give back.
func runCross(across, down Figure) (x, y int) {
	ax, ay := float64(across.DX), float64(across.DY)
	dx, dy := float64(down.DX), float64(down.DY)
	det := dx*ay - ax*dy
	if det == 0 {
		return across.StartX, across.StartY
	}
	ex := float64(across.StartX - down.StartX)
	ey := float64(across.StartY - down.StartY)
	t := (dx*ey - dy*ex) / det
	return across.StartX - int(math.Round(t*ax)), across.StartY - int(math.Round(t*ay))
}
