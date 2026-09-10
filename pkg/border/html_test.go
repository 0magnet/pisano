package border

import (
	"math"
	"strings"
	"testing"
)

// The fit knob: a wide figure gets a small font and a small one a large font,
// so both fill the same box. That is the whole point of it.
func TestFitFontScalesInversely(t *testing.T) {
	wide := FitFont(40, 10, 300, 300)
	narrow := FitFont(6, 6, 300, 300)
	if !(wide < narrow) {
		t.Errorf("a 40-cell figure got font %.2f and a 6-cell one %.2f; the wide one should be smaller", wide, narrow)
	}
	// And the fitted block really does fit.
	if got := 40 * wide * CellEm; got > 300.5 {
		t.Errorf("fitted width %.1f exceeds the 300px box", got)
	}
	if FitFont(0, 5, 100, 100) != 0 {
		t.Error("a zero-width block should not produce a font size")
	}
}

// A block turned 45 degrees needs its diagonal, not its width.
func TestRotatedExtentGrowsWithTheTurn(t *testing.T) {
	f9, _ := Of(9)
	f23, _ := Of(23)
	f37, _ := Of(37)
	l, err := Compose(Spec{Across: f9, Down: f23, Corner: f37, Cols: 5, Rows: 3})
	if err != nil {
		t.Fatal(err)
	}
	minX, minY, maxX, maxY := l.Bounds()
	flat := math.Max(float64(maxX-minX), float64(maxY-minY))
	w, h := l.RotatedExtent()
	if w < flat*0.9 && h < flat*0.9 {
		t.Errorf("rotated extent %.0fx%.0f is smaller than the unrotated %.0f; the turn was not accounted for", w, h, flat)
	}
}

// The fragment must carry one rotation for every piece and put each at its own
// lattice point — a shared angle with per-piece origins is the bug this guards.
func TestHTMLUsesOneRotationAndLatticePositions(t *testing.T) {
	f9, _ := Of(9)
	f23, _ := Of(23)
	f37, _ := Of(37)
	l, err := Compose(Spec{Across: f9, Down: f23, Corner: f37, Cols: 4, Rows: 2})
	if err != nil {
		t.Fatal(err)
	}
	frag := l.HTML(HTMLOptions{FitW: 600, FitH: 300, Color: "#3d8fb8", Corner: "#7fd6a0"})
	if n := strings.Count(frag, "<pre"); n != len(l.Places) {
		t.Errorf("fragment has %d pieces, layout has %d", n, len(l.Places))
	}
	if n := strings.Count(frag, "transform-origin:0 0"); n != len(l.Places) {
		t.Error("every piece must rotate about its top-left, which is its grid origin")
	}
	turn := strings.Count(frag, "rotate(135.0000deg)")
	if turn != len(l.Places) {
		t.Errorf("%d pieces carry the layout rotation, want all %d", turn, len(l.Places))
	}
	if !strings.Contains(frag, "#7fd6a0") {
		t.Error("corner color was not applied")
	}
}
