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
