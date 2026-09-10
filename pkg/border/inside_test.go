package border

import "testing"

// The rectangle Inside reports has to be EMPTY, or content dropped into it
// lands on top of the border.
func TestInsideIsClear(t *testing.T) {
	// mod 17 travels -4,+0 and mod 13 travels +0,-4, so this frame is upright;
	// mod 31 at four passes is closed and has the full symmetry of the square.
	across, ok := OfPasses(17, 1)
	if !ok {
		t.Fatal("no figure for modulus 17")
	}
	down, ok := OfPasses(13, 1)
	if !ok {
		t.Fatal("no figure for modulus 13")
	}
	corner, ok := OfPasses(31, 4)
	if !ok {
		t.Fatal("no figure for modulus 31 at four passes")
	}
	if !corner.Closed {
		t.Fatal("mod 31 at four passes should be closed")
	}
	l, err := Compose(Spec{Across: across, Down: down, Corner: corner, Cols: 12, Rows: 5})
	if err != nil {
		t.Fatal(err)
	}
	g := l.Grid()
	in := g.Inside()
	if in.X1-in.X0 < 10 || in.Y1-in.Y0 < 4 {
		t.Fatalf("inside is %dx%d, too small to hold anything",
			in.X1-in.X0, in.Y1-in.Y0)
	}
	for y := in.Y0; y < in.Y1; y++ {
		for x := in.X0; x < in.X1; x++ {
			if a, ok := Arms(g[y][x]); ok && a != 0 {
				t.Fatalf("mark at %d,%d inside %v", x, y, in)
			}
		}
	}
}

// Overlay replaces what is there rather than merging with it: content laid over
// a border has to cover it.
func TestOverlayReplaces(t *testing.T) {
	g := BlankGrid(6, 3)
	g.RuleH(0, 6, 1)
	g.Overlay(1, 1, []string{"abcd"})
	if got := g.Lines()[1]; got != "─abcd─" {
		t.Errorf("row 1 is %q, want %q", got, "─abcd─")
	}
	// Past the edge is dropped, not wrapped.
	g.Overlay(4, 2, []string{"xyz"})
	if got := g.Lines()[2]; got != "    xy" {
		t.Errorf("row 2 is %q, want %q", got, "    xy")
	}
}
