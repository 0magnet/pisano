package border

import "testing"

// The classification a border depends on: which figures close, which travel,
// and which travel on the diagonal.
func TestFigureClassification(t *testing.T) {
	for _, tc := range []struct {
		mod             int
		closed, travels bool
	}{
		{5, true, false},  // a closed 3x3
		{8, true, false},  // the closed diamond
		{9, false, true},  // the diagonal cascade
		{23, false, true}, // travels +6,-6
	} {
		f, ok := Of(tc.mod)
		if !ok {
			t.Fatalf("mod %d: no figure", tc.mod)
		}
		if f.Closed != tc.closed {
			t.Errorf("mod %d: Closed=%v, want %v", tc.mod, f.Closed, tc.closed)
		}
		if f.Travels() != tc.travels {
			t.Errorf("mod %d: Travels=%v, want %v (travel %d,%d)", tc.mod, f.Travels(), tc.travels, f.DX, f.DY)
		}
		if f.Closed && f.Travels() {
			t.Errorf("mod %d: a closed path cannot also travel", tc.mod)
		}
		if len(f.Grid.Stubs()) != map[bool]int{true: 0, false: 2}[f.Closed] {
			t.Errorf("mod %d: %d stubs for closed=%v", tc.mod, len(f.Grid.Stubs()), f.Closed)
		}
	}
}

// mod 9 and mod 23 travel at 45 degrees; a wandering figure does not.
func TestDiagonalDetection(t *testing.T) {
	for _, m := range []int{9, 23, 33, 43} {
		f, _ := Of(m)
		if !f.Diagonal(0.25) {
			t.Errorf("mod %d should be diagonal (travel %d,%d)", m, f.DX, f.DY)
		}
	}
	f, _ := Of(8) // closed, goes nowhere
	if f.Diagonal(0.25) {
		t.Error("a closed figure should not count as diagonal")
	}
}

// The catalog must fold orientations together, which is most of its value.
func TestCatalogDeduplicates(t *testing.T) {
	all := Catalog(3, 400)
	if len(all) == 0 {
		t.Fatal("empty catalog")
	}
	seen := map[string]int{}
	for _, f := range all {
		k := f.Grid.Canonical()
		seen[k]++
		if seen[k] > 1 {
			t.Fatalf("catalog contains the same figure twice (mod %d)", f.Mod)
		}
	}
	// The even moduli are the case that proves it is doing something: one
	// figure recurs at dozens of them.
	var biggest int
	for _, f := range all {
		if len(f.Also) > biggest {
			biggest = len(f.Also)
		}
	}
	if biggest < 10 {
		t.Errorf("expected one figure to recur at many moduli; the most any recurs is %d", biggest+1)
	}
	t.Logf("moduli 3..400: %d distinct figures; the most repeated appears at %d moduli", len(all), biggest+1)
}
