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

	// Corners at the four lattice points that bound the runs — CENTERED on the
	// point the runs converge at, not hung off it by a grid corner.
	//
	// A run's line enters its figure at StartX,StartY, so at a lattice point the
	// line is at P + start. Placing the corner's own grid origin at P instead
	// leaves the runs arriving somewhere along the corner's edge rather than
	// heading into the middle of it, which reads as the runs passing BY the
	// corner rather than ending in it.
	cx := s.Across.StartX - s.Corner.W()/2
	cy := s.Across.StartY - s.Corner.H()/2
	add(s.Corner.Grid, cx, cy, RoleCorner)
	add(s.Corner.Grid, s.Cols*ax+cx, s.Cols*ay+cy, RoleCorner)
	add(s.Corner.Grid, s.Rows*dx+cx, s.Rows*dy+cy, RoleCorner)
	add(s.Corner.Grid, s.Cols*ax+s.Rows*dx+cx, s.Cols*ay+s.Rows*dy+cy, RoleCorner)

	// Runs strictly between them, one travel apart.
	for k := 1; k < s.Cols; k++ {
		add(s.Across.Grid, k*ax, k*ay, RoleAcross)
		add(s.Across.Grid, k*ax+s.Rows*dx, k*ay+s.Rows*dy, RoleAcross)
	}
	for m := 1; m < s.Rows; m++ {
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
// means the figures met on their own.
func (l Layout) GridJoins() (Grid, int) {
	minX, minY, maxX, maxY := l.Bounds()
	// A margin, so a connector that wants to step outside the pieces has room.
	const pad = 2
	g := BlankGrid(maxX-minX+2*pad, maxY-minY+2*pad)
	// The corners are where the runs STOP, so their footprints are kept clear
	// and the runs are cut off at them. Drawing the runs whole and covering them
	// with the corner leaves run ink inside and beyond it, which makes the
	// corner look like an ornament laid on a continuous run rather than the
	// place that run ends.
	var keepOut []Rect
	for _, p := range l.Places {
		if p.Role != RoleCorner {
			continue
		}
		keepOut = append(keepOut, Rect{
			X0: p.GX - minX + pad, Y0: p.GY - minY + pad,
			X1: p.GX - minX + pad + p.Grid.W(), Y1: p.GY - minY + pad + p.Grid.H(),
		})
	}
	for _, p := range l.Places {
		if p.Role == RoleCorner {
			continue
		}
		g.StampOutside(p.Grid, p.GX-minX+pad, p.GY-minY+pad, keepOut)
	}
	for _, p := range l.Places {
		if p.Role == RoleCorner {
			g.Stamp(p.Grid, p.GX-minX+pad, p.GY-minY+pad)
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
