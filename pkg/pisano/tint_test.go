package pisano

import "testing"

// step builds a step between two adjacent points, filling in the heading the
// turtle would have had to be on to make it.
func step(from, to Pt, term, pass, index int) Step {
	s := Step{From: from, To: to, Term: term, Pass: pass, Index: index}
	switch {
	case to.X > from.X:
		s.Dir = east
	case to.X < from.X:
		s.Dir = west
	case to.Y > from.Y:
		s.Dir = south
	default:
		s.Dir = north
	}
	return s
}

func TestStepModeColorsByThePassThatFirstWalksIt(t *testing.T) {
	tn := NewTinter(TintStep, 6, 10)
	if got := tn.Tint(step(Pt{0, 0}, Pt{1, 0}, 1, 0, 0)); got != 0 {
		t.Errorf("first walk on pass 0 gave %d", got)
	}
	if got := tn.Tint(step(Pt{1, 0}, Pt{2, 0}, 1, 3, 1)); got != 3 {
		t.Errorf("first walk on pass 3 gave %d", got)
	}
}

// TestWalkingItAgainMovesForward is the whole request: retracing a step should
// change its color, every time, rather than depending on how the pass counter
// happens to line up with the palette.
func TestWalkingItAgainMovesForward(t *testing.T) {
	tn := NewTinter(TintStep, 6, 10)
	a, b := Pt{0, 0}, Pt{1, 0}
	for i, want := range []int{0, 1, 2, 3, 4, 5, 0, 1} {
		// the pass number keeps rising and is ignored after the first walk
		if got := tn.Tint(step(a, b, 1, i*7, i)); got != want {
			t.Errorf("walk %d gave %d, want %d", i, got, want)
		}
	}
}

// TestWalkingItBackMovesBack is the other half of it: a step retraced against
// the way it went before moves the other way through the palette, so a path
// that doubles back shows where.
func TestWalkingItBackMovesBack(t *testing.T) {
	tn := NewTinter(TintStep, 6, 10)
	a, b := Pt{2, 5}, Pt{2, 6}
	if got := tn.Tint(step(a, b, 1, 0, 0)); got != 0 {
		t.Fatalf("first walk gave %d", got)
	}
	if got := tn.Tint(step(b, a, 1, 1, 1)); got != 5 {
		t.Errorf("walking it back gave %d, want 5 (one before the start)", got)
	}
	if got := tn.Tint(step(b, a, 1, 2, 2)); got != 0 {
		t.Errorf("walking back again gave %d, want 0", got)
	}
	if got := tn.Tint(step(a, b, 1, 3, 3)); got != 5 {
		t.Errorf("turning round again gave %d, want 5", got)
	}
}

// TestDirectionIsPerStep — two steps that share a point keep their own counters.
func TestDirectionIsPerStep(t *testing.T) {
	tn := NewTinter(TintStep, 6, 10)
	tn.Tint(step(Pt{0, 0}, Pt{1, 0}, 1, 0, 0))
	tn.Tint(step(Pt{1, 0}, Pt{1, 1}, 1, 0, 1))
	if got := tn.Tint(step(Pt{0, 0}, Pt{1, 0}, 1, 0, 2)); got != 1 {
		t.Errorf("the horizontal step gave %d, want 1", got)
	}
	if got := tn.Tint(step(Pt{1, 1}, Pt{1, 0}, 1, 0, 3)); got != 5 {
		t.Errorf("the vertical step walked back gave %d, want 5", got)
	}
}

func TestVisitsCounts(t *testing.T) {
	tn := NewTinter(TintVisits, 6, 10)
	a, b := Pt{0, 0}, Pt{0, 1}
	for i, want := range []int{0, 1, 2} {
		// walked back and forth: a visit is a visit whichever way it goes
		from, to := a, b
		if i%2 == 1 {
			from, to = b, a
		}
		if got := tn.Tint(step(from, to, 1, 0, i)); got != want {
			t.Errorf("visit %d gave %d, want %d", i+1, got, want)
		}
	}
}

func TestHeadingIsFourColors(t *testing.T) {
	tn := NewTinter(TintHeading, 6, 10)
	seen := map[int]bool{}
	o := Pt{0, 0}
	for _, to := range []Pt{{1, 0}, {0, 1}, {-1, 0}, {0, -1}} {
		seen[tn.Tint(step(o, to, 1, 0, 0))] = true
	}
	if len(seen) != 4 {
		t.Errorf("four headings gave %d colors: %v", len(seen), seen)
	}
}

// TestTurnIsTwoColors — odd turns left, even turns right, and nothing else
// makes a step.
func TestTurnIsTwoColors(t *testing.T) {
	tn := NewTinter(TintTurn, 6, 10)
	left := tn.Tint(step(Pt{0, 0}, Pt{1, 0}, 3, 0, 0))
	right := tn.Tint(step(Pt{0, 0}, Pt{1, 0}, 4, 0, 1))
	if left == right {
		t.Errorf("a left and a right turn came out the same color: %d", left)
	}
	if again := tn.Tint(step(Pt{5, 5}, Pt{5, 6}, 7, 0, 2)); again != left {
		t.Errorf("two odd terms gave %d and %d", left, again)
	}
}

// TestAgeSweepsOncePerCircuit — the gradient has to complete exactly one sweep
// over a circuit, or the band does not meet itself when the figure closes.
func TestAgeSweepsOncePerCircuit(t *testing.T) {
	const span, n = 24, 6
	tn := NewTinter(TintAge, n, span)
	var seen []int
	for i := 0; i < span; i++ {
		seen = append(seen, tn.Tint(step(Pt{i, 0}, Pt{i + 1, 0}, 1, 0, i)))
	}
	if seen[0] != 0 {
		t.Errorf("the circuit starts at color %d", seen[0])
	}
	if seen[span-1] != n-1 {
		t.Errorf("the circuit ends at color %d, want %d", seen[span-1], n-1)
	}
	distinct := map[int]bool{}
	for _, c := range seen {
		distinct[c] = true
	}
	if len(distinct) != n {
		t.Errorf("a circuit used %d of %d colors", len(distinct), n)
	}
	// and it comes round again on the next circuit
	if got := tn.Tint(step(Pt{0, 0}, Pt{1, 0}, 1, 1, span)); got != 0 {
		t.Errorf("the next circuit started at color %d", got)
	}
}

func TestTintIndexWrapsBothWays(t *testing.T) {
	for _, tc := range []struct{ idx, want int }{
		{0, 0}, {5, 5}, {6, 0}, {7, 1}, {-1, 5}, {-2, 4}, {-6, 0}, {-7, 5},
	} {
		if got := TintIndex(tc.idx, 6); got != tc.want {
			t.Errorf("TintIndex(%d, 6) = %d, want %d", tc.idx, got, tc.want)
		}
	}
}

// TestAClosedFigureStepsThroughTheWholePalette is what the pass counter could
// not do. The geometry repeats every Shape.Periods circuits and the palette
// every six, so tinting by pass returned to a coloring it had already been in
// after 6/gcd(6, Periods) laps — three, for all but a handful of moduli.
func TestAClosedFigureStepsThroughTheWholePalette(t *testing.T) {
	const n = 6
	for _, mod := range []int{4, 5, 8, 11, 16} {
		p := Compute(Fibonacci(), mod, 0)
		shape := Classify(p.Terms)
		if !shape.Closed {
			t.Fatalf("mod %d does not close; the test wants a closed one", mod)
		}
		steps := Walk(p, shape.Periods*10)
		tn := NewTinter(TintStep, n, len(steps))
		first, _, _ := stepKey(steps[0].From, steps[0].To)
		var seen []int
		for _, s := range steps {
			idx := tn.Tint(s)
			if k, _, _ := stepKey(s.From, s.To); k == first {
				seen = append(seen, idx)
			}
		}
		if len(seen) < n {
			t.Fatalf("mod %d: the first step was only walked %d times", mod, len(seen))
		}
		distinct := map[int]bool{}
		for _, c := range seen[:n] {
			distinct[c] = true
		}
		if len(distinct) != n {
			t.Errorf("mod %d: %d colors over %d walks of one step, want %d (%v)",
				mod, len(distinct), n, n, seen[:n])
		}
	}
}

func TestParseTint(t *testing.T) {
	for _, name := range []string{"step", "pass", "visits", "heading", "turn", "term", "age"} {
		m, err := ParseTint(name)
		if err != nil {
			t.Errorf("%q: %v", name, err)
			continue
		}
		if m.String() != name {
			t.Errorf("%q round-tripped to %q", name, m.String())
		}
	}
	if _, err := ParseTint("stripes"); err == nil {
		t.Error("an unknown mode was accepted")
	}
	if m, err := ParseTint(""); err != nil || m != TintStep {
		t.Errorf("the empty flag gave %v, %v", m, err)
	}
}

// TestEveryModeStaysInThePalette guards against a mode returning an index no
// renderer can use. A mode being one color on a given figure is not a fault —
// pass is one color on a path that closes in a single pass, and visits is one
// color on a path that never crosses itself; that is the answer.
func TestEveryModeStaysInThePalette(t *testing.T) {
	for _, mode := range TintModes() {
		for _, mod := range []int{4, 10, 11, 25} {
			p := Compute(Fibonacci(), mod, 0)
			shape := Classify(p.Terms)
			steps := Walk(p, shape.Passes(3))
			tn := NewTinter(mode, 6, len(steps))
			for _, s := range steps {
				if idx := tn.Tint(s); idx < -1 || idx >= 6 {
					t.Fatalf("%v mod %d: color index %d out of range", mode, mod, idx)
				}
			}
		}
	}
}

// TestTheModesDisagree: they are different questions, so no two of them should
// give the same answer everywhere. Any one pair can coincide on any one figure
// — step and visits agree on a path that only ever retraces itself forwards —
// so the claim is over a set of figures, not a single one.
func TestTheModesDisagree(t *testing.T) {
	mods := []int{4, 5, 7, 10, 11, 16, 25, 47}
	colors := map[string][]int{}
	for _, mode := range TintModes() {
		var all []int
		for _, mod := range mods {
			p := Compute(Fibonacci(), mod, 0)
			shape := Classify(p.Terms)
			steps := Walk(p, shape.Passes(3))
			tn := NewTinter(mode, 6, len(steps))
			for _, s := range steps {
				all = append(all, tn.Tint(s))
			}
		}
		colors[mode.String()] = all
	}
	same := func(a, b []int) bool {
		if len(a) != len(b) {
			return false
		}
		for i := range a {
			if a[i] != b[i] {
				return false
			}
		}
		return true
	}
	for a, va := range colors {
		for b, vb := range colors {
			if a < b && same(va, vb) {
				t.Errorf("%s and %s color every figure identically", a, b)
			}
		}
	}
}
