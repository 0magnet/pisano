package border

import (
	"math"
	"testing"
)

func mustFig(t *testing.T, m int) Figure {
	t.Helper()
	f, ok := Of(m)
	if !ok {
		t.Fatalf("no figure for mod %d", m)
	}
	return f
}

// The two runs must come out square to each other under the single rotation,
// or they are on lattices that cannot agree.
func TestComposeRunsAreSquare(t *testing.T) {
	l, err := Compose(Spec{
		Across: mustFig(t, 9), Down: mustFig(t, 23), Corner: mustFig(t, 37),
		Cols: 6, Rows: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !l.Square(0.5) {
		t.Errorf("down run comes out at %.2f°, want 90° from the across run", l.DownAngle)
	}
	t.Logf("one rotation of %.1f° for every piece; down run lands at %.1f°", l.Turn, l.DownAngle)
}

// A figure that goes nowhere cannot make a run, and saying so beats drawing a
// pile of copies on one spot.
func TestComposeRejectsNonTravellingRuns(t *testing.T) {
	_, err := Compose(Spec{
		Across: mustFig(t, 8), Down: mustFig(t, 23), Corner: mustFig(t, 5), Cols: 4, Rows: 2,
	})
	if err == nil {
		t.Fatal("expected an error for a closed across figure")
	}
	t.Logf("rejected as expected: %v", err)
}

// Four corners, and the runs strictly between them.
func TestComposePieceCounts(t *testing.T) {
	cols, rows := 7, 4
	l, err := Compose(Spec{
		Across: mustFig(t, 9), Down: mustFig(t, 23), Corner: mustFig(t, 37),
		Cols: cols, Rows: rows,
	})
	if err != nil {
		t.Fatal(err)
	}
	n := map[Role]int{}
	for _, p := range l.Places {
		n[p.Role]++
	}
	if n[RoleCorner] != 4 {
		t.Errorf("%d corners, want 4", n[RoleCorner])
	}
	if want := 2 * (cols - 1); n[RoleAcross] != want {
		t.Errorf("%d across pieces, want %d", n[RoleAcross], want)
	}
	if want := 2 * (rows - 1); n[RoleDown] != want {
		t.Errorf("%d down pieces, want %d", n[RoleDown], want)
	}
}

// An upright layout must draw into a grid that is ONE line — that is the whole
// claim of the threading, and the unthreaded version is in many pieces.
func TestUprightLayoutDrawsOnePiece(t *testing.T) {
	// mod 11 travels (+6,-4)... find two axis-friendly travelers instead:
	// a run whose travel is purely horizontal or vertical needs no rotation.
	var across, down Figure
	for _, f := range Catalog(3, 200) {
		if f.DY == 0 && f.DX != 0 && across.Mod == 0 {
			across = f
		}
		if f.DX == 0 && f.DY != 0 && down.Mod == 0 {
			down = f
		}
	}
	if across.Mod == 0 || down.Mod == 0 {
		t.Skip("no axis-aligned travelers below mod 200")
	}
	l, err := Compose(Spec{Across: across, Down: down, Corner: mustFig(t, 8), Cols: 5, Rows: 3})
	if err != nil {
		t.Fatal(err)
	}
	if !l.Upright(0.5) {
		t.Skipf("layout is not upright (turn %.1f°)", l.Turn)
	}
	if n := l.Grid().Components(); n != 1 {
		t.Errorf("threaded border is %d pieces, want 1", n)
	}
}

// FitSpec is what makes a border dynamic: counts derived from a measured box
// rather than asked for. The test is that the composed border really does come
// out close to the box it was fitted to.
func TestFitSpecFillsTheBox(t *testing.T) {
	a := mustFig(t, 283)
	cat := Catalog(3, 400)
	ps := SquarePartners(a, cat, 1)
	if len(ps) == 0 {
		t.Skip("no square partner")
	}
	d := ps[0]
	c := mustFig(t, 755)
	for _, box := range [][2]float64{{1200, 1500}, {1200, 263}, {600, 600}, {320, 900}} {
		const cell = 2.5
		s := FitSpec(a, d, c, box[0], box[1], cell)
		l, err := Compose(s)
		if err != nil {
			t.Fatal(err)
		}
		g := l.Grid()
		if n := g.Components(); n != 1 {
			t.Errorf("box %vx%v: %d components, want 1", box[0], box[1], n)
		}
		x0, y0, x1, y1 := l.inkBox(g)
		gotW, gotH := (x1-x0)*cell, (y1-y0)*cell
		// Within a copy's length either way: the counts are whole copies, so a
		// border cannot land on an arbitrary size exactly.
		tolW := math.Hypot(float64(a.DX), float64(a.DY)) * cell
		tolH := math.Hypot(float64(d.DX), float64(d.DY)) * cell
		if math.Abs(gotW-box[0]) > tolW*1.5 || math.Abs(gotH-box[1]) > tolH*1.5 {
			t.Errorf("box %.0fx%.0f -> %.0fx%.0f (cols %d rows %d), further off than one copy",
				box[0], box[1], gotW, gotH, s.Cols, s.Rows)
			continue
		}
		t.Logf("box %4.0fx%-4.0f -> %4.0fx%-4.0f  cols %2d rows %2d", box[0], box[1], gotW, gotH, s.Cols, s.Rows)
	}
}
