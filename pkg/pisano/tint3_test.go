package pisano

import "testing"

// The tint modes in space. What is being checked is that they mean the same
// thing they mean in the plane: the plane is the case where nothing moves along
// z, so a walk that stays flat has to tint identically either way, and the
// modes with memory have to recognize a step in space as the same piece of path
// however it is walked.

// step3 builds a step in space between two adjacent points, giving it the
// frame the turtle would have been in to make that move.
func step3(from, to Pt3, pass, index int) Step3 {
	h := Pt3{to.X - from.X, to.Y - from.Y, to.Z - from.Z}
	// Any frame with this heading will do — only H is read.
	return Step3{From: from, To: to, Term: 1, Pass: pass, Index: index,
		Dir: Frame3{H: h, U: Pt3{0, 1, 0}, R: Pt3{0, 0, 1}}}
}

// TestFlatWalksTintTheSameEitherWay is the control. Walk a figure in the plane
// and the same figure embedded in space, and every mode has to agree step for
// step — otherwise "the same seven modes" is not what was carried over.
func TestFlatWalksTintTheSameEitherWay(t *testing.T) {
	p := Compute(Fibonacci(), 25, 0)
	steps := Walk(p, 3)

	for _, mode := range TintModes() {
		flat := NewTinter(mode, 6, len(steps))
		spatial := NewTinter(mode, 6, len(steps))
		for i, s := range steps {
			want := flat.Tint(s)
			got := spatial.Tint3(Step3{
				From: Pt3{s.From.X, s.From.Y, 0}, To: Pt3{s.To.X, s.To.Y, 0},
				Term: s.Term, Pass: s.Pass, Index: s.Index,
				Dir: Frame3{H: Pt3{s.To.X - s.From.X, s.To.Y - s.From.Y, 0}},
			})
			if mode == TintHeading {
				continue // four compass points against six directions; checked below
			}
			if got != want {
				t.Fatalf("%v step %d: flat gave %d, spatial gave %d", mode, i, want, got)
			}
		}
	}
}

// TestStepModeRemembersASegmentInSpace is TintStep's whole point, in space: a
// piece of path walked again moves forward through the palette, and one walked
// against the way it went before moves back.
func TestStepModeRemembersASegmentInSpace(t *testing.T) {
	tn := NewTinter(TintStep, 6, 10)
	a, b := Pt3{0, 0, 0}, Pt3{0, 0, 1}
	if got := tn.Tint3(step3(a, b, 0, 0)); got != 0 {
		t.Fatalf("first walk gave %d, want 0", got)
	}
	if got := tn.Tint3(step3(a, b, 4, 1)); got != 1 {
		t.Errorf("walking it again gave %d, want 1", got)
	}
	if got := tn.Tint3(step3(b, a, 4, 2)); got != 0 {
		t.Errorf("walking it back gave %d, want 0", got)
	}
}

// TestSegmentsAlongDifferentAxesAreDifferentPath — three steps leaving the same
// point along the three axes must not be confused for one another, which is the
// thing a key that forgot the axis would get wrong.
func TestSegmentsAlongDifferentAxesAreDifferentPath(t *testing.T) {
	tn := NewTinter(TintVisits, 6, 10)
	o := Pt3{4, 4, 4}
	for i, to := range []Pt3{{5, 4, 4}, {4, 5, 4}, {4, 4, 5}} {
		if got := tn.Tint3(step3(o, to, 0, i)); got != 0 {
			t.Errorf("step along axis %d gave %d visits-1, want 0 — it was counted as one already walked", i, got)
		}
	}
	if got := tn.Tint3(step3(o, Pt3{5, 4, 4}, 0, 3)); got != 1 {
		t.Errorf("second walk of the x step gave %d, want 1", got)
	}
}

// TestHeadingNamesSixDirections — in the plane a heading has four values, in
// space six, and opposite directions must not share a color or the mode says
// nothing about which way the turtle went.
func TestHeadingNamesSixDirections(t *testing.T) {
	tn := NewTinter(TintHeading, 6, 10)
	seen := map[int]Pt3{}
	o := Pt3{0, 0, 0}
	for _, to := range []Pt3{{1, 0, 0}, {-1, 0, 0}, {0, 1, 0}, {0, -1, 0}, {0, 0, 1}, {0, 0, -1}} {
		got := tn.Tint3(step3(o, to, 0, 0))
		if prev, dup := seen[got]; dup {
			t.Errorf("heading %v and %v both tinted %d", prev, to, got)
		}
		seen[got] = to
	}
	if len(seen) != 6 {
		t.Errorf("six directions produced %d colors, want 6", len(seen))
	}
}

// TestTintingAWholeSpatialWalkIsTotal — every step of a real walk gets a color
// in range (or the run-in's -1), for every mode. A tint that fell off the end
// of the palette would show up here as an index no palette has.
func TestTintingAWholeSpatialWalkIsTotal(t *testing.T) {
	const n = 6
	for _, m := range []int{2, 8, 11, 25, 47} {
		p := Compute(Fibonacci(), m, 0)
		steps := Walk3(p, Classify3(p.Terms).Passes(3))
		if len(steps) == 0 {
			t.Fatalf("mod %d walked nowhere", m)
		}
		for _, mode := range TintModes() {
			tn := NewTinter(mode, n, len(steps))
			for i, s := range steps {
				got := tn.Tint3(s)
				if got < -1 || got >= n {
					t.Fatalf("mod %d %v step %d gave %d, outside a palette of %d", m, mode, i, got, n)
				}
				if got == -1 && s.Pass >= 0 {
					t.Fatalf("mod %d %v step %d is in pass %d yet went uncolored", m, mode, i, s.Pass)
				}
			}
		}
	}
}

// TestVisitsSeesSpaceIsRoomier — the reason the third dimension is worth having
// is that the walk stops running over itself. Visits should find markedly less
// retracing in space than in the plane on the same modulus.
func TestVisitsSeesSpaceIsRoomier(t *testing.T) {
	p := Compute(Fibonacci(), 25, 0)

	flat := NewTinter(TintVisits, 6, 1)
	var flatRevisits int
	for _, s := range Walk(p, 3) {
		if flat.Tint(s) > 0 {
			flatRevisits++
		}
	}
	spatial := NewTinter(TintVisits, 6, 1)
	var spatialRevisits int
	for _, s := range Walk3(p, 3) {
		if spatial.Tint3(s) > 0 {
			spatialRevisits++
		}
	}
	if spatialRevisits >= flatRevisits {
		t.Errorf("mod 25 retraced %d steps in space and %d in the plane; space was expected to be roomier",
			spatialRevisits, flatRevisits)
	}
}
