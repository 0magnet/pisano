package pisano

import "fmt"

// Pt is a lattice point on the turtle's grid. Y grows downward, matching the
// terminal.
type Pt struct{ X, Y int }

// Headings, clockwise on screen, so a right turn is +1 and a left turn is -1.
const (
	east = iota
	south
	west
	north
)

var stepBy = [4]Pt{east: {1, 0}, south: {0, 1}, west: {-1, 0}, north: {0, -1}}

// Turtle is a position and a heading. It starts facing right, and since every
// sequence here opens with a zero, the first term never moves it.
type Turtle struct {
	Pos Pt
	Dir int
}

// Step applies one term of the sequence: odd turns left then moves forward one
// unit, non-zero even turns right then moves forward one, and zero does
// neither. It reports whether the turtle moved.
func (t *Turtle) Step(term int) bool {
	if term == 0 {
		return false
	}
	if term%2 != 0 {
		t.Dir = (t.Dir + 3) % 4 // left
	} else {
		t.Dir = (t.Dir + 1) % 4 // right
	}
	d := stepBy[t.Dir]
	t.Pos = Pt{t.Pos.X + d.X, t.Pos.Y + d.Y}
	return true
}

// Shape classifies a turtle path exactly, without walking it.
//
// One pass of the period is a rigid motion of the plane: turn through some
// whole number of right angles, then translate. Repeating the period repeats
// that motion, so the path either closes or it does not, and which one is
// decided entirely by the net turn:
//
//   - a net turn of a quarter or three quarters is a rotation about a fixed
//     point, and four passes return to the start;
//   - a half turn is a rotation too, and two passes suffice;
//   - no net turn leaves a pure translation, which closes only if the
//     displacement is also zero — otherwise the path marches off forever in
//     that direction.
//
// The video observes both kinds and says it never looked into what decides
// which. This is what decides it.
type Shape struct {
	Closed  bool
	Periods int // passes needed to close; 0 when the path never closes
	Turn    int // net quarter turns per pass, 0..3
	Drift   Pt  // net displacement per pass
}

func (s Shape) String() string {
	if s.Closed {
		return fmt.Sprintf("closed after %d pass(es), net turn %d×90°", s.Periods, s.Turn)
	}
	return fmt.Sprintf("open, drifts (%d,%d) per pass", s.Drift.X, s.Drift.Y)
}

// Classify walks a single pass of the period and reads off its net motion.
func Classify(terms []int) Shape {
	t := Turtle{}
	for _, term := range terms {
		t.Step(term)
	}
	s := Shape{Turn: t.Dir % 4, Drift: t.Pos}
	switch s.Turn {
	case 0:
		if s.Drift == (Pt{}) {
			s.Closed, s.Periods = true, 1
		}
	case 2:
		s.Closed, s.Periods = true, 2
	default:
		s.Closed, s.Periods = true, 4
	}
	return s
}

// Path walks the period the given number of times and returns every lattice
// point visited, starting at the origin. A closed path is walked exactly as
// many times as it needs to close, so passes is only consulted for open ones.
func Path(terms []int, passes int) []Pt {
	if passes < 1 {
		passes = 1
	}
	t := Turtle{}
	pts := []Pt{t.Pos}
	for p := 0; p < passes; p++ {
		for _, term := range terms {
			if t.Step(term) {
				pts = append(pts, t.Pos)
			}
		}
	}
	return pts
}

// PathOf walks a whole period: the run-in once, then the repeating block as
// many times as asked. Walking Terms alone would be wrong for any sequence with
// a run-in — the unreduced ones have one, since their opening zero never
// recurs.
//
// It also returns, for each segment, which pass laid it down, so a renderer can
// tint the passes apart.
func PathOf(p Period, passes int) (pts []Pt, pass []int) {
	if passes < 1 {
		passes = 1
	}
	if !p.Bounded {
		// Terms is a truncated prefix, not a cycle. Repeating it would draw
		// a loop the sequence does not actually have.
		passes = 1
	}

	t := Turtle{}
	pts = []Pt{t.Pos}
	for _, term := range p.Head {
		if t.Step(term) {
			pts = append(pts, t.Pos)
			pass = append(pass, -1) // run-in, not part of any pass
		}
	}
	for i := 0; i < passes; i++ {
		for _, term := range p.Terms {
			if t.Step(term) {
				pts = append(pts, t.Pos)
				pass = append(pass, i)
			}
		}
	}
	return pts, pass
}

// Passes is how many times to walk the period for a given render: whatever it
// takes for a closed path, or the caller's repeat count for an open one.
func (s Shape) Passes(reps int) int {
	if s.Closed {
		return s.Periods
	}
	if reps < 1 {
		reps = 1
	}
	return reps
}

// TurtleOptions controls the terminal renderer.
type TurtleOptions struct {
	Reps     int  // passes to draw for paths that never close
	Colorize bool // tint each pass differently
}

// passColors tint successive passes so an open path shows how it repeats.
var passColors = []string{
	"\x1b[36m", "\x1b[35m", "\x1b[32m", "\x1b[33m", "\x1b[34m", "\x1b[31m",
}

// RenderTurtle draws the path for a period as terminal text.
func RenderTurtle(p Period, opt TurtleOptions) (string, Shape) {
	shape := Classify(p.Terms)
	pts, pass := PathOf(p, shape.Passes(opt.Reps))
	if len(pts) < 2 {
		return "", shape
	}

	// A horizontal unit spans two columns so the drawing comes out square in
	// a terminal cell grid.
	minX, minY := pts[0].X*2, pts[0].Y
	maxX, maxY := minX, minY
	for _, q := range pts {
		x, y := q.X*2, q.Y
		if x < minX {
			minX = x
		}
		if x > maxX {
			maxX = x
		}
		if y < minY {
			minY = y
		}
		if y > maxY {
			maxY = y
		}
	}

	c := NewCanvas(minX, minY, maxX, maxY)

	for i := 0; i < len(pts)-1; i++ {
		col := ""
		if opt.Colorize && pass[i] >= 0 {
			col = passColors[pass[i]%len(passColors)]
		}
		a, b := pts[i], pts[i+1]
		c.Segment(a.X*2, a.Y, b.X*2, b.Y, col)
	}
	return c.String(opt.Colorize), shape
}

// CanonicalPath reduces a path to a form invariant under translation and under
// the eight symmetries of the square, so two moduli that trace the same figure
// in a different place or orientation compare equal. The video notes that
// several moduli produce identical turtle designs; this is how Sweep finds
// them.
func CanonicalPath(pts []Pt) string {
	if len(pts) == 0 {
		return ""
	}
	best := ""
	for variant := 0; variant < 8; variant++ {
		v := make([]Pt, len(pts))
		for i, p := range pts {
			x, y := p.X, p.Y
			if variant&4 != 0 {
				x = -x // reflection
			}
			switch variant & 3 { // rotation by a multiple of 90°
			case 1:
				x, y = -y, x
			case 2:
				x, y = -x, -y
			case 3:
				x, y = y, -x
			}
			v[i] = Pt{x, y}
		}
		ox, oy := v[0].X, v[0].Y
		buf := make([]byte, 0, len(v)*6)
		for _, p := range v {
			buf = append(buf, fmt.Sprintf("%d,%d;", p.X-ox, p.Y-oy)...)
		}
		if s := string(buf); best == "" || s < best {
			best = s
		}
	}
	return best
}
