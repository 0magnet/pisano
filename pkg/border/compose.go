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

	// Corners at the four lattice points that bound the runs.
	add(s.Corner.Grid, 0, 0, RoleCorner)
	add(s.Corner.Grid, s.Cols*ax, s.Cols*ay, RoleCorner)
	add(s.Corner.Grid, s.Rows*dx, s.Rows*dy, RoleCorner)
	add(s.Corner.Grid, s.Cols*ax+s.Rows*dx, s.Cols*ay+s.Rows*dy, RoleCorner)

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
	for _, p := range l.Places {
		g.Stamp(p.Grid, p.GX-minX+pad, p.GY-minY+pad)
	}
	return g, g.Join()
}
