package pisano

import (
	"reflect"
	"testing"
)

// The published Pisano periods for the Fibonacci numbers, OEIS A001175.
var knownPeriods = []int{
	1: 1, 2: 3, 3: 8, 4: 6, 5: 20, 6: 24, 7: 16, 8: 12, 9: 24, 10: 60,
	11: 10, 12: 24, 13: 28, 14: 48, 15: 40, 16: 24, 17: 36, 18: 24, 19: 18,
	20: 60, 21: 16, 22: 30, 23: 48, 24: 24, 25: 100, 26: 84, 27: 72, 28: 48,
	29: 14, 30: 120,
}

func TestFibonacciPeriods(t *testing.T) {
	fib := Fibonacci()
	for m, want := range knownPeriods {
		if m == 0 {
			continue
		}
		p := Compute(fib, m, 0)
		if !p.Bounded {
			t.Errorf("mod %d: no period found", m)
			continue
		}
		if got := p.Len(); got != want {
			t.Errorf("mod %d: period %d, want %d", m, got, want)
		}
		if len(p.Head) != 0 {
			t.Errorf("mod %d: Fibonacci is purely periodic, got run-in %v", m, p.Head)
		}
	}
}

// The two blocks the video reads out on screen.
func TestPeriodTerms(t *testing.T) {
	for _, tc := range []struct {
		m    int
		want []int
	}{
		{3, []int{0, 1, 1, 2, 0, 2, 2, 1}},
		{4, []int{0, 1, 1, 2, 3, 1}},
	} {
		p := Compute(Fibonacci(), tc.m, 0)
		if !reflect.DeepEqual(p.Terms, tc.want) {
			t.Errorf("fib mod %d = %v, want %v", tc.m, p.Terms, tc.want)
		}
	}
}

// The video says the number of zeros in the block is always 1, 2 or 4. It is a
// stated result rather than one it proves, so it is worth holding the code to.
func TestZeroCountIsOneTwoOrFour(t *testing.T) {
	fib := Fibonacci()
	for m := 1; m <= 500; m++ {
		z := Compute(fib, m, 0).Zeros()
		if z != 1 && z != 2 && z != 4 {
			t.Errorf("mod %d: %d zeros, want 1, 2 or 4", m, z)
		}
	}
}

// π(m) ≤ 6m, with equality exactly at m = 2·5^k.
func TestPeriodBound(t *testing.T) {
	fib := Fibonacci()
	for m := 1; m <= 500; m++ {
		if p := Compute(fib, m, 0); p.Len() > 6*m {
			t.Errorf("mod %d: period %d exceeds 6m", m, p.Len())
		}
	}
}

func TestLucasAndTriangularArePeriodic(t *testing.T) {
	for _, s := range []Sequence{Lucas(), Triangular(), Scaled(2), Scaled(3)} {
		for m := 1; m <= 120; m++ {
			if p := Compute(s, m, 0); !p.Bounded {
				t.Errorf("%s mod %d: no period found", s.Name(), m)
			}
		}
	}
}

// Scaling the sequence by k reproduces the plain designs at every kth modulus:
// k·Fib mod k·m is k times Fib mod m, so the residues are the same figures
// scaled up. This is the relation behind the video's "design m/k" labelling.
func TestScaledReproducesPlainAtEveryKth(t *testing.T) {
	for k := 2; k <= 4; k++ {
		for m := 1; m <= 40; m++ {
			plain := Compute(Fibonacci(), m, 0)
			scaled := Compute(Scaled(k), k*m, 0)
			if plain.Len() != scaled.Len() {
				t.Fatalf("k=%d m=%d: period %d vs %d", k, m, plain.Len(), scaled.Len())
			}
			for i := range plain.Terms {
				if scaled.Terms[i] != k*plain.Terms[i] {
					t.Fatalf("k=%d m=%d term %d: %d, want %d",
						k, m, i, scaled.Terms[i], k*plain.Terms[i])
				}
			}
		}
	}
}

func TestPrimesNeverReportAPeriod(t *testing.T) {
	for m := 2; m <= 30; m++ {
		if p := Compute(Primes(), m, 200); p.Bounded {
			t.Errorf("primes mod %d claimed a period", m)
		}
	}
}

// The video points out that the mod 5 design happens to use every diagonal.
func TestModFiveUsesEveryChord(t *testing.T) {
	d := Circular(Compute(Fibonacci(), 5, 0))
	if !d.Complete() {
		t.Errorf("fib mod 5: %d chords, want all %d", len(d.Chords), 5*4/2)
	}
}

// Unreduced, the Fibonacci numbers run even, odd, odd, even, odd, odd, ... with
// only the leading term actually zero, so the turtle reads them as one stay
// followed by left, left, right forever. That is the cycle the video describes,
// and it closes.
func TestUnreducedFibonacciIsLeftLeftRight(t *testing.T) {
	p, err := UnreducedPeriod(Fibonacci())
	if err != nil {
		t.Fatal(err)
	}
	if want := []int{MoveStay}; !reflect.DeepEqual(p.Head, want) {
		t.Errorf("run-in %v, want %v", p.Head, want)
	}
	if want := []int{MoveLeft, MoveLeft, MoveRight}; !reflect.DeepEqual(p.Terms, want) {
		t.Fatalf("cycle %v, want %v (left, left, right)", p.Terms, want)
	}

	s := Classify(p.Terms)
	if !s.Closed || s.Periods != 4 {
		t.Errorf("unreduced fib path: %v, want closed after 4 passes", s)
	}

	pts, _ := PathOf(p, s.Periods)
	if end := pts[len(pts)-1]; end != (Pt{}) {
		t.Errorf("path ended at %v, want the origin", end)
	}
}

// Reducing mod 2 is emphatically not the same walk: the term 2 is even and
// turns right unreduced, but is the residue 0 and stays put mod 2.
func TestModTwoDiffersFromUnreduced(t *testing.T) {
	u, err := UnreducedPeriod(Fibonacci())
	if err != nil {
		t.Fatal(err)
	}
	m2 := Compute(Fibonacci(), 2, 0)
	if reflect.DeepEqual(u.Terms, m2.Terms) {
		t.Errorf("unreduced and mod 2 gave the same cycle %v", u.Terms)
	}
	if got, want := m2.Terms, []int{0, 1, 1}; !reflect.DeepEqual(got, want) {
		t.Errorf("fib mod 2 = %v, want %v", got, want)
	}
}

func TestUnreducedOtherSequences(t *testing.T) {
	for _, s := range []Sequence{Lucas(), Triangular(), Primes(), Scaled(2)} {
		p, err := UnreducedPeriod(s)
		if err != nil {
			t.Errorf("%s: %v", s.Name(), err)
			continue
		}
		if !p.Bounded || p.Len() == 0 {
			t.Errorf("%s: no move cycle found", s.Name())
		}
		for _, m := range append(append([]int{}, p.Head...), p.Terms...) {
			if m != MoveStay && m != MoveLeft && m != MoveRight {
				t.Errorf("%s: move %d out of range", s.Name(), m)
			}
		}
	}
}

// The triangular numbers are 0, 1, 3, 6, 10, 15, 21, 28, ...: parity runs even,
// odd, odd, even and repeats with period four. Only the leading term is a
// genuine zero, so the turtle stays once and then cycles left, left, right,
// right — a four-move cycle, unlike Fibonacci's three.
func TestUnreducedTriangularParity(t *testing.T) {
	p, err := UnreducedPeriod(Triangular())
	if err != nil {
		t.Fatal(err)
	}
	if want := []int{MoveStay}; !reflect.DeepEqual(p.Head, want) {
		t.Errorf("run-in %v, want %v", p.Head, want)
	}
	want := []int{MoveLeft, MoveLeft, MoveRight, MoveRight}
	if !reflect.DeepEqual(p.Terms, want) {
		t.Errorf("triangular cycle %v, want %v", p.Terms, want)
	}

	// Two lefts and two rights cancel, and the walk drifts instead of closing.
	if s := Classify(p.Terms); s.Closed {
		t.Errorf("triangular path: %v, want open", s)
	}
}

// A closed path really does return to where it started, and an open one really
// does not. This is the claim Classify makes without walking the path, so it is
// checked here against actually walking it.
func TestClassifyMatchesTheWalk(t *testing.T) {
	fib := Fibonacci()
	for m := 1; m <= 200; m++ {
		terms := Compute(fib, m, 0).Terms
		s := Classify(terms)
		pts := Path(terms, s.Passes(4))
		end := pts[len(pts)-1]
		if s.Closed && end != (Pt{}) {
			t.Errorf("mod %d: classified closed after %d passes but ended at %v",
				m, s.Periods, end)
		}
		if !s.Closed && end == (Pt{}) {
			t.Errorf("mod %d: classified open but returned to the origin", m)
		}
	}
}

func TestCanonicalPathIgnoresPlacementAndOrientation(t *testing.T) {
	base := []Pt{{0, 0}, {1, 0}, {1, 1}, {2, 1}}

	moved := make([]Pt, len(base))
	for i, p := range base {
		moved[i] = Pt{p.X + 7, p.Y - 3}
	}
	rotated := make([]Pt, len(base))
	for i, p := range base {
		rotated[i] = Pt{-p.Y, p.X}
	}

	want := CanonicalPath(base)
	if got := CanonicalPath(moved); got != want {
		t.Errorf("translation changed the canonical form")
	}
	if got := CanonicalPath(rotated); got != want {
		t.Errorf("rotation changed the canonical form")
	}
	if CanonicalPath([]Pt{{0, 0}, {1, 0}}) == want {
		t.Errorf("different paths shared a canonical form")
	}
}

// Symmetry is a property of the chord set, so a design must agree with itself
// when its own reflection is applied.
func TestReflectionsAreConsistent(t *testing.T) {
	fib := Fibonacci()
	for m := 3; m <= 120; m++ {
		d := Circular(Compute(fib, m, 0))
		for _, s := range d.Reflections() {
			for _, c := range d.Chords {
				if !d.Has(pmod(s-c.A, m), pmod(s-c.B, m)) {
					t.Fatalf("mod %d axis %d: chord %v has no mirror", m, s, c)
				}
			}
		}
	}
}
