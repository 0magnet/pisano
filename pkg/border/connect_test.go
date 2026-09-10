package border

import (
	"math"
	"testing"
)

func sqrt(v float64) float64 { return math.Sqrt(v) }

// The guarantee: a composed border is ONE closed line, whatever figures went
// into it. Tested across the modulus range, high ones included, because a tool
// that only works for the moduli it was developed against is not a tool.
func TestEveryBorderIsOneLine(t *testing.T) {
	// Pairs chosen to vary travel size, figure size and corner shape.
	cases := []struct{ across, down, corner, cols, rows int }{
		{9, 23, 37, 9, 3},
		{33, 43, 323, 6, 2},
		{23, 9, 5, 8, 3},
		{9, 23, 755, 7, 3},
		{1009, 1051, 269, 5, 2},
		{2269, 33, 1525, 4, 2},
		{199, 139, 89, 6, 3},
		{161, 1009, 113, 5, 2},
		{2969, 2743, 55, 4, 2},
		{43, 2269, 8, 5, 3},
	}
	for _, tc := range cases {
		a, ok1 := Of(tc.across)
		d, ok2 := Of(tc.down)
		c, ok3 := Of(tc.corner)
		if !ok1 || !ok2 || !ok3 {
			t.Errorf("across %d down %d corner %d: a figure is missing", tc.across, tc.down, tc.corner)
			continue
		}
		l, err := Compose(Spec{Across: a, Down: d, Corner: c, Cols: tc.cols, Rows: tc.rows})
		if err != nil {
			t.Logf("across %-5d down %-5d corner %-5d: rejected - %v", tc.across, tc.down, tc.corner, err)
			continue
		}
		g, joins := l.GridJoins()
		n := g.Components()
		if n != 1 {
			t.Errorf("across %d down %d corner %d: %d components, want 1", tc.across, tc.down, tc.corner, n)
			continue
		}
		t.Logf("across %-5d down %-5d corner %-5d -> 1 line (%d connectors, %dx%d cells)",
			tc.across, tc.down, tc.corner, joins, g.W(), g.H())
	}
}

// Sweeping many combinations at once: the claim is that ANY accepted spec
// composes to one line, not that a hand-picked list does.
func TestBorderConnectivitySweep(t *testing.T) {
	cat := Catalog(3, 1200)
	var runs, corners []Figure
	for _, f := range cat {
		if f.Diagonal(0.25) && f.W() <= 20 && f.H() <= 20 && len(runs) < 8 {
			runs = append(runs, f)
		}
		if f.Closed && f.W() <= 14 && f.H() <= 14 && len(corners) < 5 {
			corners = append(corners, f)
		}
	}
	if len(runs) < 2 || len(corners) < 1 {
		t.Skip("not enough figures to sweep")
	}
	tried, bad := 0, 0
	for _, a := range runs {
		for _, d := range runs {
			for _, c := range corners {
				l, err := Compose(Spec{Across: a, Down: d, Corner: c, Cols: 5, Rows: 3})
				if err != nil {
					continue
				}
				tried++
				if n := l.Grid().Components(); n != 1 {
					bad++
					if bad <= 3 {
						t.Errorf("across %d down %d corner %d: %d components", a.Mod, d.Mod, c.Mod, n)
					}
				}
			}
		}
	}
	t.Logf("%d combinations composed; %d failed to close", tried, bad)
	if tried < 20 {
		t.Errorf("only %d combinations were tried; the sweep is not covering much", tried)
	}
}

// --down auto has to pick a partner that is BOTH square and comparably paced.
// Ranking by area picked figures that barely move — square, but with sides so
// short the frame was all width and no height.
func TestSquarePartnersMatchThePace(t *testing.T) {
	cat := Catalog(3, 400)
	for _, m := range []int{9, 33, 43} {
		a, ok := Of(m)
		if !ok || !a.Travels() {
			continue
		}
		ps := SquarePartners(a, cat, 1)
		if len(ps) == 0 {
			t.Errorf("mod %d: no square partner below 400", m)
			continue
		}
		best := ps[0]
		if !best.SquareTo(a, 1) {
			t.Errorf("mod %d: the top pick is not square to it", m)
		}
		ref := hyp(a.DX, a.DY)
		got := hyp(best.DX, best.DY)
		if got < ref/3 {
			t.Errorf("mod %d travels %.1f but the pick travels only %.1f; the sides will collapse",
				m, ref, got)
		}
		t.Logf("mod %-4d (travel %+d,%+d) -> mod %-4d (travel %+d,%+d)", m, a.DX, a.DY, best.Mod, best.DX, best.DY)
	}
}

func hyp(x, y int) float64 {
	fx, fy := float64(x), float64(y)
	return sqrt(fx*fx + fy*fy)
}
