package pisano

import "testing"

// The incremental walk has to be the same walk. Walk and Walk3 are the
// definition — they are what the tests of the figures themselves are written
// against — so a walker that disagrees with them anywhere is wrong, however
// plausible its own output looks.

func TestWalkerMatchesWalk(t *testing.T) {
	for _, m := range []int{1, 2, 3, 8, 10, 11, 25, 47, 100, 233} {
		p := Compute(Fibonacci(), m, 0)
		const passes = 4
		want := Walk(p, passes)

		w := NewWalker(p)
		var got []Step
		// Consume the run-in and `passes` passes worth of terms; a zero term
		// yields no step, so the count of terms is what has to line up, not the
		// count of steps.
		for i := 0; i < len(p.Head)+len(p.Terms)*passes; i++ {
			if s, ok := w.Next(); ok {
				got = append(got, s)
			}
		}
		if len(got) != len(want) {
			t.Fatalf("mod %d: walked %d steps, Walk gave %d", m, len(got), len(want))
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("mod %d step %d: walker %+v, Walk %+v", m, i, got[i], want[i])
			}
		}
	}
}

func TestWalker3MatchesWalk3(t *testing.T) {
	for _, m := range []int{1, 2, 3, 8, 10, 11, 25, 47, 100, 233} {
		p := Compute(Fibonacci(), m, 0)
		const passes = 4
		want := Walk3(p, passes)

		w := NewWalker3(p)
		var got []Step3
		for i := 0; i < len(p.Head)+len(p.Terms)*passes; i++ {
			if s, ok := w.Next(); ok {
				got = append(got, s)
			}
		}
		if len(got) != len(want) {
			t.Fatalf("mod %d: walked %d steps, Walk3 gave %d", m, len(got), len(want))
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("mod %d step %d: walker %+v, Walk3 %+v", m, i, got[i], want[i])
			}
		}
	}
}

// TestWalkerKeepsWalking is the whole reason it exists: it must never finish.
// A closed figure is the one to check, since it is the one that could plausibly
// stop having anything left to do.
func TestWalkerKeepsWalking(t *testing.T) {
	p := Compute(Fibonacci(), 10, 0) // closes after 4 passes
	w := NewWalker(p)
	for i := 0; i < 100000; i++ {
		w.Next()
	}
	if w.Pass() < 100 {
		t.Errorf("only %d passes after 100000 terms", w.Pass())
	}
	if w.Steps() < 1000 {
		t.Errorf("only %d steps after 100000 terms", w.Steps())
	}
	// And it is still walking the same figure, not drifting off it.
	if s, ok := w.Next(); !ok {
		t.Error("the walk stopped moving")
	} else if s.Index != w.Steps()-1 {
		t.Errorf("step index %d out of step with the count %d", s.Index, w.Steps())
	}
}

// TestUnboundedWalksOnceAndStops — a period that never repeated within the cap
// is a prefix, not a cycle. Repeating it would draw a loop the sequence does not
// have, so the walker stops instead.
func TestUnboundedWalksOnceAndStops(t *testing.T) {
	p := Period{Terms: []int{1, 1, 1, 1}, Bounded: false}
	w := NewWalker(p)
	n := 0
	for i := 0; i < 1000; i++ {
		if _, ok := w.Next(); ok {
			n++
		}
	}
	if n != len(p.Terms) {
		t.Errorf("walked %d steps of an unbounded prefix, want %d", n, len(p.Terms))
	}
}

// TestWalkerSurvivesAnEmptyPeriod — nothing to walk is not a crash.
func TestWalkerSurvivesAnEmptyPeriod(t *testing.T) {
	for i, p := range []Period{{}, {Bounded: true}, {Head: []int{1, 2}, Bounded: false}} {
		w, w3 := NewWalker(p), NewWalker3(p)
		for k := 0; k < 50; k++ {
			w.Next()
			w3.Next()
		}
		if w.Steps() > len(p.Head) || w3.Steps() > len(p.Head) {
			t.Errorf("period %d: walked past the end of it", i)
		}
	}
}
