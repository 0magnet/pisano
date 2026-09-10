package border

import (
	"strings"
	"testing"
)

// Every marked cell of the border must get a color, and no blank may get one.
func TestTintByCopyColorsTheWholeLine(t *testing.T) {
	d, ok := OfPasses(13, 1)
	if !ok {
		t.Fatal("no figure for modulus 13")
	}
	c, ok := OfPasses(31, 4)
	if !ok {
		t.Fatal("no figure for modulus 31 at four passes")
	}
	l, err := Compose(Spec{Across: d.RotateCW(), Down: d, Corner: c, Cols: 10, Rows: 4})
	if err != nil {
		t.Fatal(err)
	}
	g := l.Grid()
	tin := l.TintByCopy(g, 6)
	uncolored, seen := 0, map[int]bool{}
	for r := range g {
		for col := range g[r] {
			a, ok := Arms(g[r][col])
			marked := ok && a != 0
			switch {
			case marked && tin[r][col] < 0:
				uncolored++
			case marked:
				seen[tin[r][col]] = true
			case tin[r][col] >= 0:
				t.Fatalf("blank at %d,%d was given color %d", col, r, tin[r][col])
			}
		}
	}
	if uncolored != 0 {
		t.Errorf("%d marked cells got no color; the border should be one line", uncolored)
	}
	// The palette has to be used, not just its first entry.
	if len(seen) < 6 {
		t.Errorf("only %d of 6 colors used: %v", len(seen), seen)
	}
}

// The HTML has to carry the colors and stay one span per run, not per cell.
func TestTintHTMLGroupsRuns(t *testing.T) {
	d, ok := OfPasses(13, 1)
	if !ok {
		t.Fatal("no figure for modulus 13")
	}
	c, ok := OfPasses(31, 4)
	if !ok {
		t.Fatal("no figure for modulus 31 at four passes")
	}
	l, err := Compose(Spec{Across: d.RotateCW(), Down: d, Corner: c, Cols: 10, Rows: 4})
	if err != nil {
		t.Fatal(err)
	}
	g := l.Grid()
	out := g.TintHTML(l.TintByCopy(g, 6), nil)
	spans := strings.Count(out, "<span")
	if spans == 0 {
		t.Fatal("no spans emitted")
	}
	marks := 0
	for r := range g {
		for col := range g[r] {
			if a, ok := Arms(g[r][col]); ok && a != 0 {
				marks++
			}
		}
	}
	// Grouped, not one per cell. The bound is loose because a corner is now
	// four colors of its own rather than one, so the frame changes color more
	// often than the runs alone would.
	if spans > marks/2 {
		t.Errorf("%d spans for %d marks; runs are not being grouped", spans, marks)
	}
	if strings.Count(out, "</span>") != spans {
		t.Error("spans are not balanced")
	}
	// pisano's own palette by default.
	if !strings.Contains(out, "#00cdcd") {
		t.Error("the default palette is not pisano's terminal pass palette")
	}
}
