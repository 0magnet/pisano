package border

import "testing"

// Turning a figure has to turn everything a border reads: the drawing, the
// travel, and where the path starts. Turning only the picture composes a frame
// from a figure that travels one way and draws another.
func TestRotateCWTurnsTravelAndStart(t *testing.T) {
	f, ok := OfPasses(13, 1)
	if !ok {
		t.Fatal("no figure for modulus 13")
	}
	if f.DX != 0 || f.DY != -4 {
		t.Fatalf("modulus 13 travels %+d,%+d, expected +0,-4", f.DX, f.DY)
	}
	r := f.RotateCW()
	if r.DX != 4 || r.DY != 0 {
		t.Errorf("turned once it travels %+d,%+d, want +4,+0", r.DX, r.DY)
	}
	if r.W() != f.H() || r.H() != f.W() {
		t.Errorf("turned it is %dx%d, want %dx%d", r.W(), r.H(), f.H(), f.W())
	}
	// The start has to land on the same mark it was on before.
	if a, ok := Arms(r.Grid[r.StartY][r.StartX]); !ok || a == 0 {
		t.Errorf("start %d,%d is blank after turning", r.StartX, r.StartY)
	}
	// Four turns is the identity.
	if s := f.Rotate(4); s.Grid.String() != f.Grid.String() ||
		s.DX != f.DX || s.DY != f.DY || s.StartX != f.StartX || s.StartY != f.StartY {
		t.Error("four quarter turns is not the identity")
	}
	// And one modulus can serve both directions: square by construction.
	if !r.SquareTo(f, 0.001) {
		t.Errorf("a figure and its own quarter turn are not square: %.2f vs %.2f",
			r.Angle(), f.Angle())
	}
}

// The far run has to be the near one REFLECTED, or a frame whose two runs are
// the same drawing has its ornament turned into the frame along the bottom and
// out of it along the top.
//
// Only the two runs the mirror swaps are compared. A run is a stack of copies
// of one figure, so it cannot be symmetric along its OWN length unless the
// figure is, and the frame's mirror maps each side run onto itself — that
// residue is a property of the figure, not of the composition.
func TestFarRunsAreReflections(t *testing.T) {
	d, ok := OfPasses(13, 1)
	if !ok {
		t.Fatal("no figure for modulus 13")
	}
	c, ok := OfPasses(31, 4)
	if !ok {
		t.Fatal("no figure for modulus 31 at four passes")
	}
	across := d.RotateCW()
	l, err := Compose(Spec{Across: across, Down: d, Corner: c, Cols: 12, Rows: 5})
	if err != nil {
		t.Fatal(err)
	}
	g, _ := l.GridJoins()
	x0, y0, x1, y1 := inkBounds(g)
	// Clear of the corners, where the runs are all there is.
	for r := y0; r <= y0+across.H()+1; r++ {
		for col := x0 + c.W() + 2; col <= x1-c.W()-2; col++ {
			if got, want := g[r][col], MirrorGlyphV(g[y0+y1-r][col]); got != want {
				t.Fatalf("top and bottom runs differ at %d,%d: %q vs %q",
					col, r, string(got), string(want))
			}
		}
	}
	for r := y0 + c.H() + 2; r <= y1-c.H()-2; r++ {
		for col := x0; col <= x0+d.W()+1; col++ {
			if got, want := g[r][col], MirrorGlyphH(g[r][x0+x1-col]); got != want {
				t.Fatalf("left and right runs differ at %d,%d: %q vs %q",
					col, r, string(got), string(want))
			}
		}
	}
}

// inkBounds is the box the marks occupy, ignoring the composition's margin.
func inkBounds(g Grid) (x0, y0, x1, y1 int) {
	x0, y0, x1, y1 = 1<<30, 1<<30, -1, -1
	for r := range g {
		for c := range g[r] {
			if a, ok := Arms(g[r][c]); !ok || a == 0 {
				continue
			}
			x0, x1 = min(x0, c), max(x1, c)
			y0, y1 = min(y0, r), max(y1, r)
		}
	}
	return
}
