package pisano

import "testing"

// The axis direction has to satisfy the thing that defines it: the rotation
// leaves it alone. Checking that against every modulus is a far better test
// than any handful of expected vectors, because it is the property the caller
// depends on rather than a transcription of today's output.
func TestAxisDirIsFixedByTheTurn(t *testing.T) {
	for m := 2; m <= 300; m++ {
		p := Compute(Fibonacci(), m, 1<<20)
		if !p.Bounded || len(p.Terms) == 0 {
			continue
		}
		s := Classify3(p.Terms)
		d := s.AxisDir()
		if d.isZero() {
			// Only legitimate when the figure neither turns nor drifts, which
			// is a path that closes on the spot.
			if s.Turn != IdentityFrame3 || !s.Drift.isZero() {
				t.Errorf("m=%d: no axis direction, but turn=%v drift=%v", m, s.Turn, s.Drift)
			}
			continue
		}
		if got := s.Turn.Apply(d); got != d {
			t.Errorf("m=%d: turn moves the axis: %v -> %v", m, d, got)
		}
	}
}

// A helical figure advances ALONG its axis — that is what makes it a screw
// rather than a rotation and a slide that disagree. So the advance and the
// direction must be parallel.
func TestAxisDirIsParallelToTheAdvance(t *testing.T) {
	for m := 2; m <= 300; m++ {
		p := Compute(Fibonacci(), m, 1<<20)
		if !p.Bounded || len(p.Terms) == 0 {
			continue
		}
		s := Classify3(p.Terms)
		if !s.Helical || s.Axial.isZero() {
			continue
		}
		if !parallel3(s.AxisDir(), s.Axial) {
			t.Errorf("m=%d: axis %v is not parallel to the advance %v",
				m, s.AxisDir(), s.Axial)
		}
	}
}

// Primitive means two parallel axes are the same vector, not two multiples of
// one — otherwise comparing them needs a division every time.
func TestAxisDirIsPrimitive(t *testing.T) {
	for m := 2; m <= 300; m++ {
		p := Compute(Fibonacci(), m, 1<<20)
		if !p.Bounded || len(p.Terms) == 0 {
			continue
		}
		d := Classify3(p.Terms).AxisDir()
		if d.isZero() {
			continue
		}
		if g := gcd3(gcd3(abs3(d.X), abs3(d.Y)), abs3(d.Z)); g != 1 {
			t.Errorf("m=%d: axis %v has common factor %d", m, d, g)
		}
	}
}

// The crystallographic restriction, checked rather than asserted: a rotation
// that maps the cubic lattice to itself has order 1, 2, 3 or 4. This is what
// makes the figure seen end-on 1-, 2-, 3- or 4-fold and nothing else, so it is
// worth a test of its own.
func TestTurnOrderIsCrystallographic(t *testing.T) {
	seen := map[int]bool{}
	for _, seq := range []Sequence{Fibonacci(), Lucas(), Scaled(2), Scaled(3)} {
		for m := 2; m <= 300; m++ {
			p := Compute(seq, m, 1<<20)
			if !p.Bounded || len(p.Terms) == 0 {
				continue
			}
			o := Classify3(p.Terms).Turn.Order()
			seen[o] = true
			switch o {
			case 1, 2, 3, 4:
			default:
				t.Fatalf("%s m=%d: rotation of order %d cannot preserve a cubic lattice",
					seq.Name(), m, o)
			}
		}
	}
	for _, want := range []int{1, 2, 3, 4} {
		if !seen[want] {
			t.Errorf("order %d never occurred; the survey is not exercising the whole group", want)
		}
	}
}
