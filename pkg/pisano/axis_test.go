package pisano

import "testing"

// Where the screw axis is.
//
// The defining property is the whole test: a point on the axis is carried by
// one pass of the motion to itself plus the advance along the axis, and nothing
// else is. Everything is checked scaled up by the denominator so it stays exact
// integer arithmetic — an axis that came out a rounding error away from right
// would be a slow wobble on screen, which is the thing it exists to remove.

// passMotion is what one pass of the period does to a point: turn, then shift.
func passMotion(s Shape3, p Pt3) Pt3 { return s.Turn.Apply(p).Add(s.Drift) }

func TestAxisIsCarriedAlongItself(t *testing.T) {
	for _, m := range []int{2, 3, 5, 8, 11, 16, 25, 47, 100, 233, 254} {
		p := Compute(Fibonacci(), m, 0)
		s := Classify3(p.Terms)
		num, den := s.Axis()
		if den < 1 {
			t.Fatalf("mod %d: denominator %d", m, den)
		}
		k := s.Periods

		// One pass moves the axis point by exactly the advance per pass, which
		// is Axial/Periods. Scaled by den·k so both sides are integers:
		//   den·k·(R·c + drift) == den·k·c + den·Axial
		// and R is linear, so R·(den·c) can be applied to the numerator.
		lhs := s.Turn.Apply(Pt3{num.X * k, num.Y * k, num.Z * k}).Add(
			Pt3{den * k * s.Drift.X, den * k * s.Drift.Y, den * k * s.Drift.Z})
		rhs := Pt3{
			num.X*k + den*s.Axial.X,
			num.Y*k + den*s.Axial.Y,
			num.Z*k + den*s.Axial.Z,
		}
		if lhs != rhs {
			t.Errorf("mod %d (%v): the axis point is not carried along the axis: %v vs %v",
				m, s, lhs, rhs)
		}
	}
}

// TestAxisRemovesThePrecession is what it is for. A point of the figure swings
// about the axis and advances along it — so its distance from the LINE never
// changes, however many passes go by. Measured from anywhere else it does, and
// that difference is the wobble a viewer sees when it centers on the wrong
// point: the figure orbiting an axis it failed to find.
//
// Distance from a line needs no tracking of the advance along it, since sliding
// along the line changes only the parallel part. Everything stays integer:
// (perpendicular distance · |a|)² is |d|²|a|² − (d·a)².
func TestAxisRemovesThePrecession(t *testing.T) {
	perp := func(d, a Pt3) int {
		dd := d.X*d.X + d.Y*d.Y + d.Z*d.Z
		aa := a.X*a.X + a.Y*a.Y + a.Z*a.Z
		da := d.X*a.X + d.Y*a.Y + d.Z*a.Z
		return dd*aa - da*da
	}
	checked := 0
	for m := 2; m <= 300; m++ {
		p := Compute(Fibonacci(), m, 0)
		s := Classify3(p.Terms)
		if !s.Helical {
			continue
		}
		checked++
		num, den := s.Axis()

		x := Pt3{den * 3, den * 5, den * 7} // a point of the figure, den-scaled
		want := perp(Pt3{x.X - num.X, x.Y - num.Y, x.Z - num.Z}, s.Axial)
		offAxis := perp(Pt3{x.X, x.Y, x.Z}, s.Axial) // measured from the origin instead
		for pass := 1; pass <= 3*s.Periods; pass++ {
			x = s.Turn.Apply(x).Add(Pt3{den * s.Drift.X, den * s.Drift.Y, den * s.Drift.Z})
			got := perp(Pt3{x.X - num.X, x.Y - num.Y, x.Z - num.Z}, s.Axial)
			if got != want {
				t.Fatalf("mod %d pass %d: distance from the axis moved, %d -> %d", m, pass, want, got)
			}
			if o := perp(x, s.Axial); o != offAxis {
				offAxis = -1 // the origin is not on the axis for this one, as expected
			}
		}
		if offAxis != -1 && want != 0 {
			t.Logf("mod %d: the origin happens to lie on the axis", m)
		}
	}
	if checked == 0 {
		t.Fatal("no helical figures to check")
	}
	t.Logf("checked %d helices", checked)
}

// TestAxisOfAPureTranslationIsTheOrigin — with no turn there is no axis to
// find, and a viewer that subtracts it must not be shifted by it.
func TestAxisOfAPureTranslationIsTheOrigin(t *testing.T) {
	s := Shape3{Periods: 1, Turn: IdentityFrame3, Drift: Pt3{3, -4, 5}, Axial: Pt3{3, -4, 5}}
	if num, den := s.Axis(); num != (Pt3{}) || den != 1 {
		t.Errorf("a pure translation gave axis %v/%d, want the origin", num, den)
	}
}

// TestClosedFiguresHaveAFixedPoint — a closed figure's motion has no advance at
// all, so its axis point is genuinely fixed: the motion carries it to itself.
func TestClosedFiguresHaveAFixedPoint(t *testing.T) {
	found := 0
	for m := 2; m <= 300; m++ {
		p := Compute(Fibonacci(), m, 0)
		s := Classify3(p.Terms)
		if !s.Closed || s.Turn == IdentityFrame3 {
			continue
		}
		found++
		num, den := s.Axis()
		got := s.Turn.Apply(num).Add(Pt3{den * s.Drift.X, den * s.Drift.Y, den * s.Drift.Z})
		if got != num {
			t.Errorf("mod %d: the fixed point moved, %v -> %v", m, num, got)
		}
	}
	if found == 0 {
		t.Fatal("no closed turning figures to check")
	}
}
