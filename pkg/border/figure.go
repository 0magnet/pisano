package border

import (
	"math"
	"sort"

	"github.com/0magnet/pisano/pkg/pisano"
)

// Figure is one turtle figure, with the properties a border needs.
type Figure struct {
	Mod    int   // the modulus that drew it (the smallest, when several do)
	Passes int   // how many times round the period; the figure changes with it
	Also   []int // the other moduli producing the same figure, up to orientation
	Grid   Grid  // the drawing
	DX, DY int   // net travel of the path, in cells: where it ends up
	Closed bool  // the path returns to where it started
	// StartX, StartY are where the path BEGINS, in cells from the grid origin —
	// which is where a run's line enters the figure, and so the point another
	// piece has to be lined up with. Placing by the grid corner instead makes
	// runs meet a corner off to one side of it.
	StartX, StartY int
	Points         int // path length, a fair measure of how busy it looks
}

// W and H in cells.
func (f Figure) W() int { return f.Grid.W() }
func (f Figure) H() int { return f.Grid.H() }

// Travels reports whether the figure goes somewhere rather than wandering back.
// Only a traveling figure chains into a run: copies laid one travel apart
// advance along the edge instead of piling up.
func (f Figure) Travels() bool { return f.DX != 0 || f.DY != 0 }

// Diagonal reports whether the travel is near 45 degrees, within tol (0.25 is
// a reasonable default). These are the figures worth turning: chained they
// climb, and turning the strip lays them along an edge that no unturned figure
// could follow.
func (f Figure) Diagonal(tol float64) bool {
	ax, ay := abs(f.DX), abs(f.DY)
	if ax < 3 || ay < 3 {
		return false
	}
	lo, hi := ax, ay
	if lo > hi {
		lo, hi = hi, lo
	}
	return float64(lo)/float64(hi) >= 1-tol
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// Of builds the figure for one modulus, walking the period twice.
func Of(m int) (Figure, bool) { return OfPasses(m, 2) }

// OfPasses builds the figure for one modulus, walking the period `passes` times.
//
// The pass count is not a detail: it is half the design. The same modulus draws
// a different figure at each count, and whether the path CLOSES depends on it —
// modulus 1019 is an open 50x45 wanderer at two passes and a closed, perfectly
// half-turn-symmetric 60x60 at four. Fixing it at two, as this package first
// did, hides most of what the sequence can draw, and hides the large symmetric
// figures entirely.
//
// ok is false when the sequence has no bounded period within the term limit, or
// the path is too short to draw — neither makes a usable figure.
func OfPasses(m, passes int) (Figure, bool) {
	if passes < 1 {
		passes = 1
	}
	p := pisano.Compute(pisano.Fibonacci(), m, termLimit)
	if !p.Bounded {
		return Figure{}, false
	}
	pts, _ := pisano.PathOf(p, passes)
	if len(pts) < 2 {
		return Figure{}, false
	}
	minX, minY, maxX, maxY := pts[0].X, pts[0].Y, pts[0].X, pts[0].Y
	for _, q := range pts {
		if q.X < minX {
			minX = q.X
		}
		if q.Y < minY {
			minY = q.Y
		}
		if q.X > maxX {
			maxX = q.X
		}
		if q.Y > maxY {
			maxY = q.Y
		}
	}
	cv := pisano.NewCanvas(minX, minY, maxX, maxY)
	for i := 0; i+1 < len(pts); i++ {
		cv.Segment(pts[i].X, pts[i].Y, pts[i+1].X, pts[i+1].Y, "")
	}
	g := NewGrid(splitTrim(cv.String(false)))
	if g.H() == 0 {
		return Figure{}, false
	}
	first, last := pts[0], pts[len(pts)-1]
	return Figure{
		Mod:    m,
		Passes: passes,
		Grid:   g,
		DX:     last.X - first.X,
		DY:     last.Y - first.Y,
		Closed: first.X == last.X && first.Y == last.Y,
		StartX: first.X - minX,
		StartY: first.Y - minY,
		Points: len(pts),
	}, true
}

// termLimit is how far Compute looks for a repeat. The Pisano period of m can
// reach 6m, so this covers moduli into the tens of thousands; a sequence that
// has not repeated by then is reported unbounded rather than guessed at.
const termLimit = 200000

// Catalog is every distinct figure for moduli in [lo,hi], deduplicated up to
// rotation and reflection, smallest first.
//
// Deduplication matters more than it sounds. Of the moduli up to 3000, the even
// ones draw the same handful of figures over and over — one shape recurs at
// more than two hundred of them — so a list that did not fold orientations
// together would be mostly repeats.
func Catalog(lo, hi int) []Figure { return CatalogPasses(lo, hi, DefaultPasses) }

// DefaultPasses is the set of pass counts a catalog walks when none is given.
//
// Four is in it for a reason: several of the largest symmetric figures only
// exist there. Modulus 1399 at four passes is a 96x96 with four-fold rotational
// symmetry, and at two passes it is nothing of the sort — a search fixed at two
// concludes, wrongly, that no closed figure above 6x6 has more than a half turn.
var DefaultPasses = []int{1, 2, 3, 4, 6}

// CatalogPasses is Catalog over the given pass counts.
func CatalogPasses(lo, hi int, passes []int) []Figure {
	seen := map[string]*Figure{}
	for m := lo; m <= hi; m++ {
		for _, ps := range passes {
			f, ok := OfPasses(m, ps)
			if !ok {
				continue
			}
			k := f.Grid.Canonical()
			if prev, dup := seen[k]; dup {
				prev.Also = append(prev.Also, m)
				continue
			}
			cp := f
			seen[k] = &cp
		}
	}
	out := make([]Figure, 0, len(seen))
	for _, f := range seen {
		out = append(out, *f)
	}
	sort.Slice(out, func(i, j int) bool {
		ai, aj := out[i].W()*out[i].H(), out[j].W()*out[j].H()
		if ai != aj {
			return ai < aj
		}
		return out[i].Mod < out[j].Mod
	})
	return out
}

func splitTrim(s string) []string {
	// Trailing newlines from the canvas would add empty rows to every figure.
	for len(s) > 0 && s[len(s)-1] == '\n' {
		s = s[:len(s)-1]
	}
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	return append(out, s[start:])
}

// Angle is the direction the path travels, in degrees, measured the way the
// screen measures: x to the right, y down. Zero for a figure that goes nowhere.
func (f Figure) Angle() float64 {
	if !f.Travels() {
		return 0
	}
	return math.Atan2(float64(f.DY), float64(f.DX)) * 180 / math.Pi
}

// SquareTo reports whether this figure can run down the side of a border whose
// top is `across`: the two travels must be a quarter turn apart, within tol
// degrees.
//
// This is the difference between a border and a parallelogram. Both runs are
// turned by the same angle — that is what keeps them on one lattice — so if
// their travels are not perpendicular to begin with, no rotation will make the
// frame square, and the result leans.
func (f Figure) SquareTo(across Figure, tol float64) bool {
	if !f.Travels() || !across.Travels() {
		return false
	}
	d := math.Mod(math.Abs(f.Angle()-across.Angle()), 180)
	return math.Abs(d-90) <= tol
}

// SquarePartners returns the figures in cat that can run down the side of a
// border topped by `across`, nearest-sized first — so the pick that will look
// balanced against it comes up before the one that dwarfs it.
//
// Choosing a partner by hand is the step that most often produces a leaning
// border, because whether two travels are perpendicular is not something the
// drawings show.
func SquarePartners(across Figure, cat []Figure, tol float64) []Figure {
	var out []Figure
	for _, f := range cat {
		if f.SquareTo(across, tol) {
			out = append(out, f)
		}
	}
	// Ranked by how close the TRAVEL is, not the area.
	//
	// Travel is the spacing of a run, so a partner that travels one cell where
	// the top travels six gives sides six times as dense and a frame that is all
	// width and no height. Area was the wrong key: a figure can be large and
	// still barely move.
	ref := math.Hypot(float64(across.DX), float64(across.DY))
	span := func(f Figure) float64 { return math.Hypot(float64(f.DX), float64(f.DY)) }
	sort.Slice(out, func(i, j int) bool {
		di, dj := math.Abs(span(out[i])-ref), math.Abs(span(out[j])-ref)
		if di != dj {
			return di < dj
		}
		return out[i].Mod < out[j].Mod
	})
	return out
}

// Extent is the figure's size along its longest diagonal, in cells — what it
// occupies once turned, which is the fair way to compare two figures that will
// be turned by the same angle.
func (f Figure) Extent() float64 {
	return math.Hypot(float64(f.W()), float64(f.H()))
}

// CornerFor picks a corner figure for a border whose runs are these.
//
// Closed, so it has no loose ends of its own where a run arrives. At least
// atLeast times the BAND of the widest run — the thickness of the stripe it
// draws — or the runs swallow it and it stops reading as the place they end. And as symmetric as can be had at that
// size — a corner is seen four times, once per corner, each time turned or
// mirrored, so a symmetric figure makes the frame read as one design instead of
// four rotations of a motif.
//
// Size and symmetry pull against each other, and hard. Of the closed figures
// under modulus 6000 there are exactly THREE with the full symmetry of the
// square, and they are 3x3, 4x4 and 6x6 — the most symmetric corner available
// is also nearly the smallest. Two-fold symmetry, a half turn, is where the big
// figures are: 321 of them, up to 160x188. So this takes the most symmetric
// figure that meets the size, rather than the most symmetric outright, and a
// caller asking for a large corner is choosing the half turn whether or not it
// knows it.
func CornerFor(across, down Figure, cat []Figure, atLeast float64) (Figure, bool) {
	need := math.Max(across.Band(), down.Band()) * atLeast
	var best Figure
	bestSym := -1
	for _, f := range cat {
		if !f.Closed || f.Extent() < need {
			continue
		}
		s := f.Grid.Symmetry()
		switch {
		case s > bestSym:
			best, bestSym = f, s
		case s == bestSym && f.Extent() < best.Extent():
			// Among equally symmetric, the smallest that still qualifies: a
			// corner should terminate the runs, not overwhelm them.
			best = f
		}
	}
	return best, bestSym >= 0
}

// Band is how thick the stripe a run lays down is: the figure's extent measured
// ACROSS its travel, in cells.
//
// This, not Extent, is the number a corner has to beat. A run figure is long in
// the direction it travels — that is what traveling means — and its bounding
// diagonal is mostly that length, so sizing a corner against Extent asks it to
// be bigger than the whole repeating unit rather than bigger than the line the
// run draws. The two differ by a lot: a figure 68 cells along and 7 across has
// an Extent of 68 and a Band of 7, and a corner ten times the Band is a corner
// that reads as the end of the line, while ten times the Extent does not exist.
func (f Figure) Band() float64 {
	t := math.Hypot(float64(f.DX), float64(f.DY))
	if t == 0 {
		return f.Extent()
	}
	// The unit normal to the travel; the projection onto it is the distance
	// either side of the run's own axis.
	nx, ny := -float64(f.DY)/t, float64(f.DX)/t
	lo, hi := math.MaxFloat64, -math.MaxFloat64
	for r := range f.Grid {
		for c := range f.Grid[r] {
			if a, ok := Arms(f.Grid[r][c]); !ok || a == 0 {
				continue
			}
			d := float64(c)*nx + float64(r)*ny
			lo, hi = math.Min(lo, d), math.Max(hi, d)
		}
	}
	if hi < lo {
		return 0
	}
	return hi - lo
}

// Axis is a point on the run's center line, in the figure's own grid.
//
// Not StartX,StartY, which is where the PATH happens to begin. Those are
// different things and the gap between them is visible in the finished frame:
// modulus 13 is five cells wide and its path starts at column 4, the right-hand
// edge, so lining a corner up on the start puts the corner two cells off the
// stripe the run actually draws. Measured on a 17/13/31 frame, that was the
// verticals sitting a cell and a half left of their corners.
//
// Across the travel it is the middle of the ink, not the average of it: a run's
// stripe is bounded by its extremes, and a figure with a dense knot at one edge
// has a centroid pulled into the knot while the stripe it draws is unmoved.
// Along the travel the average is right, and harmless either way — every point
// on the line is as good as any other, and Compose only ever intersects it with
// another line.
func (f Figure) Axis() (x, y float64) {
	t := math.Hypot(float64(f.DX), float64(f.DY))
	if t == 0 {
		return float64(f.W()-1) / 2, float64(f.H()-1) / 2
	}
	ux, uy := float64(f.DX)/t, float64(f.DY)/t
	nx, ny := -uy, ux
	lo, hi := math.MaxFloat64, -math.MaxFloat64
	su, n := 0.0, 0
	for r := range f.Grid {
		for c := range f.Grid[r] {
			if a, ok := Arms(f.Grid[r][c]); !ok || a == 0 {
				continue
			}
			fc, fr := float64(c), float64(r)
			d := fc*nx + fr*ny
			lo, hi = math.Min(lo, d), math.Max(hi, d)
			su += fc*ux + fr*uy
			n++
		}
	}
	if n == 0 {
		return 0, 0
	}
	mid, avg := (lo+hi)/2, su/float64(n)
	return mid*nx + avg*ux, mid*ny + avg*uy
}
