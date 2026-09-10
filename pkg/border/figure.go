package border

import (
	"math"
	"sort"

	"github.com/0magnet/pisano/pkg/pisano"
)

// Figure is one turtle figure, with the properties a border needs.
type Figure struct {
	Mod    int   // the modulus that drew it (the smallest, when several do)
	Also   []int // the other moduli producing the same figure, up to orientation
	Grid   Grid  // the drawing
	DX, DY int   // net travel of the path, in cells: where it ends up
	Closed bool  // the path returns to where it started
	Points int   // path length, a fair measure of how busy it looks
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

// Of builds the figure for one modulus of the Fibonacci sequence.
//
// ok is false when the sequence has no bounded period within the term limit, or
// the path is too short to draw — neither makes a usable figure.
func Of(m int) (Figure, bool) {
	p := pisano.Compute(pisano.Fibonacci(), m, termLimit)
	if !p.Bounded {
		return Figure{}, false
	}
	pts, _ := pisano.PathOf(p, 2)
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
		Grid:   g,
		DX:     last.X - first.X,
		DY:     last.Y - first.Y,
		Closed: first.X == last.X && first.Y == last.Y,
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
func Catalog(lo, hi int) []Figure {
	seen := map[string]*Figure{}
	for m := lo; m <= hi; m++ {
		f, ok := Of(m)
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
