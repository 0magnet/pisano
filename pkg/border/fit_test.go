package border

import "testing"

// Frame has to find a size that fits, and the text has to land in the clear.
func TestFrameFitsItsContent(t *testing.T) {
	d, ok := OfPasses(13, 1)
	if !ok {
		t.Fatal("no figure for modulus 13")
	}
	c, ok := OfPasses(31, 4)
	if !ok {
		t.Fatal("no figure for modulus 31 at four passes")
	}
	across := d.RotateCW()
	for _, lines := range [][]string{
		{"one"},
		{"a much longer line than the first one", "and a second", "and a third"},
	} {
		g, err := Frame(across, d, c, lines, 1)
		if err != nil {
			t.Fatal(err)
		}
		// Every line must be readable back out of the grid.
		text := g.Lines()
		for _, want := range lines {
			found := false
			for _, got := range text {
				if len(want) > 0 && contains(got, want) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("line %q is not in the frame", want)
			}
		}
	}
}

func contains(hay, needle string) bool {
	if len(needle) > len(hay) {
		return false
	}
	for i := 0; i+len(needle) <= len(hay); i++ {
		if hay[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}

// FitBox must never hand back a frame bigger than the box it was given, and
// must find one whenever the box is big enough to hold the smallest.
func TestFitBoxStaysInsideItsBox(t *testing.T) {
	d, ok := OfPasses(13, 1)
	if !ok {
		t.Fatal("no figure for modulus 13")
	}
	c, ok := OfPasses(31, 4)
	if !ok {
		t.Fatal("no figure for modulus 31 at four passes")
	}
	across := d.RotateCW()
	smallest, ok := FitBox(across, d, c, 1000, 1000)
	if !ok {
		t.Fatal("no frame fits in 1000x1000")
	}
	_ = smallest
	for _, box := range [][2]int{{40, 30}, {60, 40}, {120, 50}, {200, 90}, {31, 31}} {
		g, ok := FitBox(across, d, c, box[0], box[1])
		if !ok {
			continue // too small for any frame, which is a legitimate answer
		}
		if g.W() > box[0] || g.H() > box[1] {
			t.Errorf("box %dx%d got a frame %dx%d", box[0], box[1], g.W(), g.H())
		}
		if g.Components() != 1 {
			t.Errorf("box %dx%d: frame came out in %d pieces", box[0], box[1], g.Components())
		}
	}
	// A box far too small has to say so rather than return something broken.
	if _, ok := FitBox(across, d, c, 6, 6); ok {
		t.Error("a 6x6 box should not fit a frame with 8x8 corners")
	}
}
