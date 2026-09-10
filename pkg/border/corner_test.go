package border

import "testing"

// tuiSpec is the frame that has to work in a terminal: modulus 17 travels
// -4,+0 and 13 travels +0,-4, so nothing needs turning, and 31 at four passes
// is closed, 8x8, and has the full symmetry of the square.
func tuiSpec(t *testing.T, bias Bias) Spec {
	t.Helper()
	get := func(m, p int) Figure {
		f, ok := OfPasses(m, p)
		if !ok {
			t.Fatalf("no figure for modulus %d at %d passes", m, p)
		}
		return f
	}
	return Spec{Across: get(17, 1), Down: get(13, 1), Corner: get(31, 4),
		Cols: 12, Rows: 5, Bias: bias}
}

// The frame has to come out as one line with NO connectors of Join's choosing:
// every meeting is designed, drawn by reachOut along the row and column the
// corner was hung on.
func TestTUIFrameNeedsNoConnectors(t *testing.T) {
	l, err := Compose(tuiSpec(t, Inward))
	if err != nil {
		t.Fatal(err)
	}
	if !l.Upright(1) || !l.Square(1) {
		t.Fatalf("turn %.1f, down angle %.1f: this frame must be upright and square",
			l.Turn, l.DownAngle)
	}
	g, joins := l.GridJoins()
	if joins != 0 {
		t.Errorf("%d connectors drawn, want none", joins)
	}
	if n := g.Components(); n != 1 {
		t.Errorf("%d pieces, want one closed line", n)
	}
}

// The four corners must be reflections of each other. One offset used at all
// four is the same offset in ABSOLUTE terms, which tucks the top run inside its
// corners and hangs the bottom one outside them — the frame then looks lopsided
// however good each piece is on its own.
func TestCornersMirrorEachOther(t *testing.T) {
	for _, bias := range []Bias{Inward, Outward} {
		l, err := Compose(tuiSpec(t, bias))
		if err != nil {
			t.Fatal(err)
		}
		var xs, ys []int
		w, h := 0, 0
		for _, p := range l.Places {
			if p.Role != RoleCorner {
				continue
			}
			xs = append(xs, p.MeetX)
			ys = append(ys, p.MeetY)
			w, h = p.Grid.W(), p.Grid.H()
		}
		if len(xs) != 4 {
			t.Fatalf("%d corners, want 4", len(xs))
		}
		// Two distinct meeting columns, two rows, each pair reflecting.
		seen := map[int]bool{}
		for _, x := range xs {
			seen[x] = true
		}
		if len(seen) != 2 {
			t.Errorf("bias %v: meeting columns %v, want two distinct", bias, xs)
		}
		for x := range seen {
			if !seen[w-1-x] {
				t.Errorf("bias %v: column %d has no reflection %d in %v", bias, x, w-1-x, xs)
			}
		}
		seen = map[int]bool{}
		for _, y := range ys {
			seen[y] = true
		}
		if len(seen) != 2 {
			t.Errorf("bias %v: meeting rows %v, want two distinct", bias, ys)
		}
		for y := range seen {
			if !seen[h-1-y] {
				t.Errorf("bias %v: row %d has no reflection %d in %v", bias, y, h-1-y, ys)
			}
		}
	}
}

// Inward and Outward have to differ, and differ by swapping sides: a corner
// whose runs come in below its middle one way comes in above it the other.
func TestBiasSwapsSides(t *testing.T) {
	in, err := Compose(tuiSpec(t, Inward))
	if err != nil {
		t.Fatal(err)
	}
	out, err := Compose(tuiSpec(t, Outward))
	if err != nil {
		t.Fatal(err)
	}
	first := func(l Layout) Placement {
		for _, p := range l.Places {
			if p.Role == RoleCorner {
				return p
			}
		}
		t.Fatal("no corner")
		return Placement{}
	}
	a, b := first(in), first(out)
	w, h := a.Grid.W(), a.Grid.H()
	if a.MeetX == b.MeetX && a.MeetY == b.MeetY {
		t.Fatalf("bias changes nothing: both meet at %d,%d", a.MeetX, a.MeetY)
	}
	// A closed corner's middle is its hole, so the two biases straddle it.
	midX, midY := float64(w-1)/2, float64(h-1)/2
	if (float64(a.MeetX) > midX) == (float64(b.MeetX) > midX) {
		t.Errorf("both biases meet on the same side of column %.1f: %d and %d",
			midX, a.MeetX, b.MeetX)
	}
	if (float64(a.MeetY) > midY) == (float64(b.MeetY) > midY) {
		t.Errorf("both biases meet on the same side of row %.1f: %d and %d",
			midY, a.MeetY, b.MeetY)
	}
}
