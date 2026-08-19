package pisano

import "testing"

// sequences the 3D walk is exercised over. The unreduced and unbounded ones are
// included on purpose: they are the awkward cases, with a run-in or no cycle at
// all.
var seqs3 = []struct {
	name  string
	build func() Sequence
}{
	{"fib", Fibonacci},
	{"lucas", Lucas},
	{"tri", Triangular},
	{"nat", Naturals},
	{"prime", Primes},
}

// TestClassify3MatchesWalkingIt is the whole claim: that the shape of a path in
// space can be read off one pass, without walking any further.
//
// Walk the number of passes Classify3 predicts and see whether the turtle is
// back at the origin facing the way it started. A path it calls closed must be;
// one it does not must not be.
func TestClassify3MatchesWalkingIt(t *testing.T) {
	for _, s := range seqs3 {
		for m := 1; m <= 300; m++ {
			p := Compute(s.build(), m, 200)
			if !p.Bounded || len(p.Terms) == 0 {
				continue
			}
			sh := Classify3(p.Terms)

			// Compose the pass motion Periods times, from the identity.
			r, drift := pass3(p.Terms)
			frame, pos := IdentityFrame3, Pt3{}
			for i := 0; i < sh.Periods; i++ {
				pos = pos.Add(frame.Apply(drift))
				frame = frame.Compose(r)
			}
			home := pos.isZero() && frame == IdentityFrame3
			if home != sh.Closed {
				t.Fatalf("%s mod %d: Classify3 says %v, walking %d passes gives home=%v",
					s.name, m, sh, sh.Periods, home)
			}
		}
	}
}

// TestClosureIsTheAxialDrift states the theorem the classification rests on: a
// path closes exactly when the drift has no component along the rotation axis.
// Axial is that accumulated component, so it is zero precisely for closed
// figures.
func TestClosureIsTheAxialDrift(t *testing.T) {
	for m := 1; m <= 500; m++ {
		p := Compute(Fibonacci(), m, 0)
		sh := Classify3(p.Terms)
		if sh.Closed != sh.Axial.isZero() {
			t.Errorf("mod %d: closed=%v but axial drift is (%d,%d,%d)",
				m, sh.Closed, sh.Axial.X, sh.Axial.Y, sh.Axial.Z)
		}
	}
}

// TestOrderIsAlwaysARotationOfTheCube — 1, 2, 3 or 4 and nothing else. A frame
// that produced anything else would not be a rotation, which would mean a turn
// somewhere is reflecting the figure.
func TestOrderIsAlwaysARotationOfTheCube(t *testing.T) {
	seen := map[int]int{}
	for _, s := range seqs3 {
		for m := 1; m <= 500; m++ {
			p := Compute(s.build(), m, 200)
			if !p.Bounded || len(p.Terms) == 0 {
				continue
			}
			k := Classify3(p.Terms).Periods
			switch k {
			case 1, 2, 3, 4:
				seen[k]++
			default:
				t.Fatalf("%s mod %d: rotation of order %d", s.name, m, k)
			}
		}
	}
	// Three has no counterpart in the plane, where orders are 1, 2 and 4. Its
	// absence would mean the turtle never reaches a body-diagonal rotation, and
	// so is not really using the group.
	if seen[3] == 0 {
		t.Error("no three-fold path found; the walk is not reaching the body diagonals")
	}
	for _, k := range []int{1, 2, 4} {
		if seen[k] == 0 {
			t.Errorf("no path of order %d", k)
		}
	}
}

// TestAllThreeOutcomesOccur. Closed and open exist in the plane; helical does
// not, and is the point of going to three dimensions.
func TestAllThreeOutcomesOccur(t *testing.T) {
	var closed, drifting, helical int
	for m := 1; m <= 500; m++ {
		sh := Classify3(Compute(Fibonacci(), m, 0).Terms)
		switch {
		case sh.Closed:
			closed++
		case sh.Helical:
			helical++
		default:
			drifting++
		}
	}
	if closed == 0 || drifting == 0 || helical == 0 {
		t.Errorf("closed %d, drifting %d, helical %d — want some of each",
			closed, drifting, helical)
	}
}

// TestFiguresAreGenuinelyThreeDimensional. A rule that produced flat figures
// would be a lifted plane wearing a third coordinate, which is not worth
// having.
func TestFiguresAreGenuinelyThreeDimensional(t *testing.T) {
	flat := 0
	for m := 3; m <= 300; m++ {
		p := Compute(Fibonacci(), m, 0)
		sh := Classify3(p.Terms)
		if Span3(Path3(p, sh.Passes(2))) < 3 {
			flat++
			if flat <= 3 {
				t.Logf("mod %d stays flat", m)
			}
		}
	}
	if flat > 0 {
		t.Errorf("%d moduli produced flat figures", flat)
	}
}

// TestPath3AgreesWithWalk3 — one is the other's points, and a renderer will use
// whichever is handier.
func TestPath3AgreesWithWalk3(t *testing.T) {
	for _, m := range []int{1, 2, 5, 10, 11, 25, 47} {
		p := Compute(Fibonacci(), m, 0)
		steps := Walk3(p, 2)
		pts := Path3(p, 2)
		if len(pts) != len(steps)+1 {
			t.Fatalf("mod %d: %d points for %d steps", m, len(pts), len(steps))
		}
		for i, s := range steps {
			if pts[i] != s.From || pts[i+1] != s.To {
				t.Fatalf("mod %d step %d: %v→%v but points are %v,%v",
					m, i, s.From, s.To, pts[i], pts[i+1])
			}
		}
	}
}

// TestStepsAreUnitMoves. Every move is one unit along an axis; anything else
// means a turn has gone wrong and the figure is off the lattice.
func TestStepsAreUnitMoves(t *testing.T) {
	for _, s := range seqs3 {
		for m := 1; m <= 120; m++ {
			p := Compute(s.build(), m, 200)
			for _, st := range Walk3(p, 3) {
				d := Pt3{st.To.X - st.From.X, st.To.Y - st.From.Y, st.To.Z - st.From.Z}
				n := abs3(d.X) + abs3(d.Y) + abs3(d.Z)
				if n != 1 {
					t.Fatalf("%s mod %d: step %v→%v is not a unit move", s.name, m, st.From, st.To)
				}
			}
		}
	}
}

// TestFrameStaysRightHanded. Right is heading × up, and every turn has to keep
// it that way — a turn that reflected the frame would break the rigid-motion
// argument the classification depends on.
func TestFrameStaysRightHanded(t *testing.T) {
	cross := func(a, b Pt3) Pt3 {
		return Pt3{a.Y*b.Z - a.Z*b.Y, a.Z*b.X - a.X*b.Z, a.X*b.Y - a.Y*b.X}
	}
	for m := 1; m <= 200; m++ {
		p := Compute(Fibonacci(), m, 0)
		for _, st := range Walk3(p, 2) {
			if got := cross(st.Dir.H, st.Dir.U); got != st.Dir.R {
				t.Fatalf("mod %d: H×U = %v but R = %v", m, got, st.Dir.R)
			}
		}
	}
}

// TestUnboundedIsWalkedOnce. A sequence with no cycle has a truncated prefix,
// not a repeating block; walking it twice would draw a loop it does not have.
func TestUnboundedIsWalkedOnce(t *testing.T) {
	p := Compute(Primes(), 10, 40)
	if p.Bounded {
		t.Skip("the primes mod 10 turned out to repeat; the test needs a case that does not")
	}
	one, three := len(Walk3(p, 1)), len(Walk3(p, 3))
	if one != three {
		t.Errorf("asking for three passes of an unbounded sequence gave %d steps, one gave %d",
			three, one)
	}
}

// TestRunInIsWalkedAndMarked. A sequence with a run-in has terms before the
// cycle; they are part of the figure and belong to no pass. Reducing modulo m
// never produces one — the unreduced sequences do, because their opening zero
// never recurs once the values are saturating rather than wrapping.
func TestRunInIsWalkedAndMarked(t *testing.T) {
	// The primes unreduced: a run-in of [2], which moves, unlike the opening
	// zero the other sequences begin with. A zero run-in term is walked and
	// correctly produces no step, so it would prove nothing here.
	p, err := UnreducedPeriod(Primes())
	if err != nil {
		t.Fatalf("unreduced primes: %v", err)
	}
	if len(p.Head) == 0 || p.Head[0] == 0 {
		t.Fatalf("unreduced primes gave run-in %v; the test needs one that moves", p.Head)
	}

	steps := Walk3(p, 2)
	var runIn, inPass int
	for _, s := range steps {
		if s.Pass < 0 {
			runIn++
		} else {
			inPass++
		}
	}
	if runIn == 0 {
		t.Errorf("a run-in of %v produced no step marked as one", p.Head)
	}
	if inPass == 0 {
		t.Error("the repeating block produced no steps")
	}

	// The run-in comes first and happens once, however many passes are asked
	// for: it is a prefix, not part of the cycle.
	for i, s := range steps {
		if s.Pass < 0 && i >= runIn {
			t.Fatalf("step %d is run-in but comes after the block started", i)
		}
	}
	if again := Walk3(p, 5); countRunIn(again) != runIn {
		t.Errorf("five passes gave %d run-in steps, two gave %d", countRunIn(again), runIn)
	}
}

func countRunIn(steps []Step3) int {
	n := 0
	for _, s := range steps {
		if s.Pass < 0 {
			n++
		}
	}
	return n
}

// TestPassesFollowsTheShape — a closed figure is drawn for the circuit that
// closes it, and one that never closes for as many circuits as asked.
func TestPassesFollowsTheShape(t *testing.T) {
	for m := 2; m <= 200; m++ {
		sh := Classify3(Compute(Fibonacci(), m, 0).Terms)
		if sh.Closed {
			if got := sh.Passes(5); got != sh.Periods {
				t.Errorf("mod %d closed: Passes(5) = %d, want %d", m, got, sh.Periods)
			}
			continue
		}
		if got, want := sh.Passes(3), 3*sh.Periods; got != want {
			t.Errorf("mod %d open: Passes(3) = %d, want %d", m, got, want)
		}
		if got := sh.Passes(0); got != sh.Periods {
			t.Errorf("mod %d: Passes(0) = %d, want at least one circuit (%d)", m, got, sh.Periods)
		}
	}
}

// TestSmallAndDegenerateModuli — mod 1 reduces everything to zero, so nothing
// moves, and nothing should panic on the way to finding that out.
func TestSmallAndDegenerateModuli(t *testing.T) {
	for _, s := range seqs3 {
		for _, m := range []int{1, 2, 3} {
			p := Compute(s.build(), m, 200)
			sh := Classify3(p.Terms)
			pts := Path3(p, sh.Passes(2))
			if len(pts) == 0 {
				t.Errorf("%s mod %d: Path3 returned nothing", s.name, m)
			}
			_ = sh.String()
			lo, hi := Bounds3(pts)
			if lo.X > hi.X || lo.Y > hi.Y || lo.Z > hi.Z {
				t.Errorf("%s mod %d: bounds inverted", s.name, m)
			}
		}
	}
}

// TestClassify3IsStable — the same terms give the same answer, and no walk
// leaves state behind that the next one can see.
func TestClassify3IsStable(t *testing.T) {
	p := Compute(Fibonacci(), 47, 0)
	first := Classify3(p.Terms)
	for i := 0; i < 5; i++ {
		Walk3(p, 4)
		if got := Classify3(p.Terms); got != first {
			t.Fatalf("run %d: %v, first time %v", i, got, first)
		}
	}
}

// TestTheRuleReadsTheNextTerm pins the rule itself, not just the theorem it
// satisfies. Taking the plane from a second bit of the same term also gives a
// rigid motion and passes everything above — but it is a different figure, and
// a worse one, so something has to notice if it is ever swapped in.
//
// The property that decided it: the pair rule turns out of plane at every
// modulus, because every residue has a parity. A second bit does not exist at
// all when the modulus is 2, and is thin for small moduli generally.
func TestTheRuleReadsTheNextTerm(t *testing.T) {
	for m := 2; m <= 400; m++ {
		p := Compute(Fibonacci(), m, 0)
		moves, pitches := 0, 0
		turtle := NewTurtle3()
		for i := range p.Terms {
			up := turtle.Frame.U
			if turtle.Step(p.Terms[i], p.Terms[(i+1)%len(p.Terms)]) {
				moves++
				if turtle.Frame.U != up { // a pitch tips "up"; a yaw leaves it
					pitches++
				}
			}
		}
		if moves == 0 {
			continue
		}
		// A third is a floor, not a target — the measured share runs around a
		// half. Anything below this means the rule has stopped using the room
		// it was given.
		if share := float64(pitches) / float64(moves); share < 1.0/3 {
			t.Errorf("mod %d: only %.2f of moves leave the plane (%d of %d)",
				m, share, pitches, moves)
		}
	}
}

// TestSmallestModulusIsStillThreeDimensional is the sharpest case of the same
// thing. Fibonacci mod 2 is 0,1,1: two moves, and the rule has to get a turn
// out of the plane from them. A rule reading a second bit of the term cannot —
// there is no second bit in a residue below 2.
func TestSmallestModulusIsStillThreeDimensional(t *testing.T) {
	p := Compute(Fibonacci(), 2, 0)
	sh := Classify3(p.Terms)
	if got := Span3(Path3(p, sh.Passes(2))); got != 3 {
		t.Errorf("fib mod 2 occupies %d dimensions, want 3", got)
	}
}

// TestKnownPath is a golden: the first steps of a figure, so that any change to
// the turn conventions — which axis yaw is about, which way left goes, the
// handedness of the frame — has to be deliberate rather than accidental.
func TestKnownPath(t *testing.T) {
	want := []Pt3{
		{0, 0, 0}, {0, 1, 0}, {0, 1, -1}, {1, 1, -1}, {1, 1, -2},
		{1, 0, -2}, {2, 0, -2}, {2, -1, -2}, {2, -1, -1}, {3, -1, -1},
	}
	got := Path3(Compute(Fibonacci(), 10, 0), 1)
	if len(got) < len(want) {
		t.Fatalf("path has %d points, want at least %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("point %d is %v, want %v", i, got[i], want[i])
		}
	}
}
