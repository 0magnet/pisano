package border

import (
	"math"
	"strings"
	"testing"
)

// tuiSpec is the frame that has to work in a terminal: modulus 17 travels
// -4,+0 and 13 travels +0,-4, so nothing needs turning, and 31 at four passes
// is closed, 8x8, and has the full symmetry of the square.
func tuiSpec(t *testing.T, bias Bias) Spec {
	t.Helper()
	get := func(m, p int) Figure {
		f, ok := OfPasses(m, p)
		if !ok {
			t.Fatalf("no figure for modulus %d at %d passes", m, p)
		}
		return f
	}
	return Spec{Across: get(17, 1), Down: get(13, 1), Corner: get(31, 4),
		Cols: 12, Rows: 5, Bias: bias}
}

// The frame has to come out as one line with NO connectors of Join's choosing:
// every meeting is designed, drawn by reachOut along the row and column the
// corner was hung on.
func TestTUIFrameNeedsNoConnectors(t *testing.T) {
	l, err := Compose(tuiSpec(t, Inward))
	if err != nil {
		t.Fatal(err)
	}
	if !l.Upright(1) || !l.Square(1) {
		t.Fatalf("turn %.1f, down angle %.1f: this frame must be upright and square",
			l.Turn, l.DownAngle)
	}
	g, joins := l.GridJoins()
	if joins != 0 {
		t.Errorf("%d connectors drawn, want none", joins)
	}
	if n := g.Components(); n != 1 {
		t.Errorf("%d pieces, want one closed line", n)
	}
}

// The four corners must be reflections of each other. One offset used at all
// four is the same offset in ABSOLUTE terms, which tucks the top run inside its
// corners and hangs the bottom one outside them — the frame then looks lopsided
// however good each piece is on its own.
func TestCornersMirrorEachOther(t *testing.T) {
	for _, bias := range []Bias{Inward, Outward} {
		l, err := Compose(tuiSpec(t, bias))
		if err != nil {
			t.Fatal(err)
		}
		var xs, ys []int
		w, h := 0, 0
		for _, p := range l.Places {
			if p.Role != RoleCorner {
				continue
			}
			xs = append(xs, p.MeetX)
			ys = append(ys, p.MeetY)
			w, h = p.Grid.W(), p.Grid.H()
		}
		if len(xs) != 4 {
			t.Fatalf("%d corners, want 4", len(xs))
		}
		// Two distinct meeting columns, two rows, each pair reflecting.
		seen := map[int]bool{}
		for _, x := range xs {
			seen[x] = true
		}
		if len(seen) != 2 {
			t.Errorf("bias %v: meeting columns %v, want two distinct", bias, xs)
		}
		for x := range seen {
			if !seen[w-1-x] {
				t.Errorf("bias %v: column %d has no reflection %d in %v", bias, x, w-1-x, xs)
			}
		}
		seen = map[int]bool{}
		for _, y := range ys {
			seen[y] = true
		}
		if len(seen) != 2 {
			t.Errorf("bias %v: meeting rows %v, want two distinct", bias, ys)
		}
		for y := range seen {
			if !seen[h-1-y] {
				t.Errorf("bias %v: row %d has no reflection %d in %v", bias, y, h-1-y, ys)
			}
		}
	}
}

// Inward and Outward have to differ, and differ by swapping sides: a corner
// whose runs come in below its middle one way comes in above it the other.
func TestBiasSwapsSides(t *testing.T) {
	in, err := Compose(tuiSpec(t, Inward))
	if err != nil {
		t.Fatal(err)
	}
	out, err := Compose(tuiSpec(t, Outward))
	if err != nil {
		t.Fatal(err)
	}
	first := func(l Layout) Placement {
		for _, p := range l.Places {
			if p.Role == RoleCorner {
				return p
			}
		}
		t.Fatal("no corner")
		return Placement{}
	}
	a, b := first(in), first(out)
	w, h := a.Grid.W(), a.Grid.H()
	if a.MeetX == b.MeetX && a.MeetY == b.MeetY {
		t.Fatalf("bias changes nothing: both meet at %d,%d", a.MeetX, a.MeetY)
	}
	// A closed corner's middle is its hole, so the two biases straddle it.
	midX, midY := float64(w-1)/2, float64(h-1)/2
	if (float64(a.MeetX) > midX) == (float64(b.MeetX) > midX) {
		t.Errorf("both biases meet on the same side of column %.1f: %d and %d",
			midX, a.MeetX, b.MeetX)
	}
	if (float64(a.MeetY) > midY) == (float64(b.MeetY) > midY) {
		t.Errorf("both biases meet on the same side of row %.1f: %d and %d",
			midY, a.MeetY, b.MeetY)
	}
}

// Trim takes off the loose ends and nothing else. Two things have to survive
// it: the frame stays one line, and every piece stays where it was — a rule
// that erased crosses along a cut edge would punch a hole through the weave and
// part the run from its corner.
func TestTrimKeepsTheFrameWhole(t *testing.T) {
	l, err := Compose(tuiSpec(t, Inward))
	if err != nil {
		t.Fatal(err)
	}
	g, joins := l.GridJoins()
	if joins != 0 || g.Components() != 1 {
		t.Fatalf("%d connectors and %d pieces after trimming, want 0 and 1",
			joins, g.Components())
	}
	// Nothing left hanging by one link with an arm into the void.
	for r := range g {
		for c := range g[r] {
			a, ok := Arms(g[r][c])
			if !ok || a == 0 {
				continue
			}
			if live, dead := g.links(r, c, a); live <= 1 && dead > 0 {
				t.Fatalf("loose end still at %d,%d: %q", c, r, string(g[r][c]))
			}
		}
	}
}

// Detached is a look, not a failure to join: the runs stop at the corners and
// nothing is drawn between them, so the frame is eight pieces on purpose.
func TestDetachedDrawsNoSegments(t *testing.T) {
	s := tuiSpec(t, Inward)
	s.Detached = true
	l, err := Compose(s)
	if err != nil {
		t.Fatal(err)
	}
	g, joins := l.GridJoins()
	if joins != 0 {
		t.Errorf("%d connectors drawn, want none at all when detached", joins)
	}
	if n := g.Components(); n != 8 {
		t.Errorf("%d pieces, want 8: four runs and four corners", n)
	}
	// The corners have to come through untouched — no arm grafted on by a
	// segment that was not drawn.
	joined, err := Compose(tuiSpec(t, Inward))
	if err != nil {
		t.Fatal(err)
	}
	jg, _ := joined.GridJoins()
	if g.String() == jg.String() {
		t.Error("detached and joined frames are identical; the segments did nothing")
	}
}

// Nothing may stick out. Every arm on every character has to reach a character
// reaching back — no quarter-character spurs into empty space, whether the
// frame is joined into one line or left in pieces.
func TestNothingSticksOut(t *testing.T) {
	for _, detached := range []bool{false, true} {
		s := tuiSpec(t, Inward)
		s.Detached = detached
		l, err := Compose(s)
		if err != nil {
			t.Fatal(err)
		}
		g, _ := l.GridJoins()
		for r := range g {
			for c := range g[r] {
				a, ok := Arms(g[r][c])
				if !ok || a == 0 {
					continue
				}
				if _, dead := g.links(r, c, a); dead > 0 {
					t.Fatalf("detached=%v: %q at %d,%d has %d arm(s) reaching nothing",
						detached, string(g[r][c]), c, r, dead)
				}
			}
		}
		// And smoothing is a fixed point: a second pass must change nothing.
		if n := g.Smooth(); n != 0 {
			t.Errorf("detached=%v: smoothing again changed %d characters", detached, n)
		}
	}
}

// Inset moves the cut by the number of cells asked for, and a whole travel step
// leaves the edge looking exactly as it did — that is the point of counting in
// cells rather than steps.
func TestInsetMovesTheCutByCells(t *testing.T) {
	d, ok := OfPasses(13, 1)
	if !ok {
		t.Fatal("no figure for modulus 13")
	}
	c, ok := OfPasses(31, 4)
	if !ok {
		t.Fatal("no figure for modulus 31 at four passes")
	}
	across := d.RotateCW()
	edge := func(inset int) (col int, shape string) {
		l, err := Compose(Spec{
			Across: across, Down: d, Corner: c, Cols: 12, Rows: 5,
			Detached: true, Inset: inset,
		})
		if err != nil {
			t.Fatal(err)
		}
		g, _ := l.GridJoins()
		// The leftmost across ink: the run band's rows, clear of the corner.
		x0, y0, _, _ := inkBounds(g)
		top, bot := y0, y0+across.H()-1
		col = 1 << 30
		for r := top; r <= bot; r++ {
			for x := x0 + c.W(); x < g.W(); x++ {
				if a, ok := Arms(g[r][x]); ok && a != 0 {
					if x < col {
						col = x
					}
					break
				}
			}
		}
		var b strings.Builder
		for r := top; r <= bot; r++ {
			b.WriteString(string(g[r][col : col+6]))
			b.WriteByte('\n')
		}
		return col, b.String()
	}
	c0, s0 := edge(0)
	c1, _ := edge(1)
	if c1 <= c0 {
		t.Errorf("inset 1 starts the run at column %d, not further in than %d", c1, c0)
	}
	// A whole travel step is four cells here, and it must repeat.
	step := across.DX
	cs, ss := edge(step)
	if ss != s0 {
		t.Errorf("inset of one whole step (%d cells) changed the edge:\n%s\nwant:\n%s", step, ss, s0)
	}
	if cs-c0 != step {
		t.Errorf("inset of %d cells moved the run %d columns, want %d", step, cs-c0, step)
	}
}

// The part of a run that meets a corner has to be the crossing of its wave, not
// a peak or a trough — the middle of the stripe against the corner, not its top
// or bottom edge.
//
// Measured after trimming, because that is the edge that survives, and the two
// differ: reading the figure column by column said crossing while the trimmed
// edge sat a whole cell above the center line.
func TestRunsMeetCornersAtTheCrossing(t *testing.T) {
	d, ok := OfPasses(13, 1)
	if !ok {
		t.Fatal("no figure for modulus 13")
	}
	c, ok := OfPasses(31, 4)
	if !ok {
		t.Fatal("no figure for modulus 31 at four passes")
	}
	across := d.RotateCW()
	for _, f := range []Figure{across, d} {
		lo, hi, ok := crossings(f)
		if !ok {
			t.Fatalf("modulus %d: no cut found", f.Mod)
		}
		offLo, _ := leadOffset(f, lo, true)
		offHi, _ := leadOffset(f, hi, false)
		// Half a cell is as close as an even-sided stripe can come to a center
		// line that falls between rows; a whole cell is a peak.
		if math.Abs(offLo) > 0.5 {
			t.Errorf("modulus %d: near end meets the corner %+.1f off its center line",
				f.Mod, offLo)
		}
		if math.Abs(offHi) > 0.5 {
			t.Errorf("modulus %d: far end meets the corner %+.1f off its center line",
				f.Mod, offHi)
		}
		// And the two ends alike, or the fault has only moved.
		if math.Abs(offLo+offHi) > 1e-9 && math.Abs(offLo-offHi) > 1e-9 {
			t.Errorf("modulus %d: the two ends meet their corners differently, %+.1f and %+.1f",
				f.Mod, offLo, offHi)
		}
	}
	l, err := Compose(Spec{Across: across, Down: d, Corner: c, Cols: 12, Rows: 5})
	if err != nil {
		t.Fatal(err)
	}
	g, joins := l.GridJoins()
	if joins != 0 || g.Components() != 1 {
		t.Errorf("%d connectors and %d pieces, want 0 and 1", joins, g.Components())
	}
}
