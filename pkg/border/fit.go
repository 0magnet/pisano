package border

import (
	"fmt"
	"math"
)

// Frame builds the smallest border that holds the given lines, and puts them
// in it.
//
// This is the call a page makes. Everything else in the package works in copies
// — how many times the across figure repeats, how many the down one does —
// which is the right unit for the composition and the wrong one for a caller,
// who has a block of text and wants a border round it. The counts are found by
// growing until the clear middle is big enough.
//
// Growing rather than solving. The clear middle cannot be predicted from the
// counts: the runs are woven so their inner edge is ragged, the corners reach
// further in than the runs do, and how far the runs are held off their corners
// depends on where the crossing of each run's wave falls. All of that is only
// known once the thing is built, so it is built and then measured. A frame
// costs about a millisecond and this takes a handful of tries.
func Frame(across, down, corner Figure, lines []string, pad int) (Grid, error) {
	if pad < 0 {
		pad = 0
	}
	needW, needH := 0, len(lines)+2*pad
	for _, l := range lines {
		needW = max(needW, len([]rune(l))+2*pad)
	}
	// A first guess from the travels, then grow whichever way is short. The
	// guess is usually within a copy or two, so this is a nudge and not a
	// search.
	stepA := math.Hypot(float64(across.DX), float64(across.DY))
	stepD := math.Hypot(float64(down.DX), float64(down.DY))
	if stepA == 0 || stepD == 0 {
		return nil, fmt.Errorf("border: a frame needs two figures that travel, got %+d,%+d and %+d,%+d",
			across.DX, across.DY, down.DX, down.DY)
	}
	cols := max(1, int(math.Ceil(float64(needW)/stepA)))
	rows := max(1, int(math.Ceil(float64(needH)/stepD)))
	const tries = 40
	for range tries {
		l, err := Compose(Spec{Across: across, Down: down, Corner: corner,
			Cols: cols, Rows: rows})
		if err != nil {
			return nil, err
		}
		g := l.Grid()
		in := g.Inside()
		w, h := in.X1-in.X0, in.Y1-in.Y0
		if w >= needW && h >= needH {
			y := in.Y0 + (h-len(lines))/2
			g.Overlay(in.X0+(w-needW)/2+pad, y, lines)
			return g, nil
		}
		if w < needW {
			cols++
		}
		if h < needH {
			rows++
		}
	}
	return nil, fmt.Errorf("border: %d copies each way still will not hold %dx%d; the corners may be bigger than the frame",
		cols, needW, needH)
}

// FitBox is the largest frame that fits inside a box of the given size in
// characters, and whether one does at all.
//
// The call a decorator makes: it has a box on a page, measured in character
// cells, and wants a frame drawn to fill it. Frame is the other way round — it
// has content and grows a border to hold it — and neither can be had from the
// other, because a frame's size is quantised by the figures and lands where it
// lands.
//
// Solved rather than searched. A frame's width is linear in the number of
// across copies, since each copy advances it by exactly one travel, so two
// compositions give the slope and the intercept and the count falls out; a
// couple of steps either way then settle the rounding. That matters because
// this runs per box on every resize, in a browser: growing one copy at a time
// would be a hundred compositions for a wide box, and this is four.
func FitBox(across, down, corner Figure, cols, rows int) (Grid, bool) {
	size := func(c, r int) (w, h int, ok bool) {
		l, err := Compose(Spec{Across: across, Down: down, Corner: corner, Cols: c, Rows: r})
		if err != nil {
			return 0, 0, false
		}
		minX, minY, maxX, maxY := l.Bounds()
		// Matching the margin GridJoins leaves, so this measures what will
		// actually be drawn.
		return maxX - minX + 4, maxY - minY + 4, true
	}
	w1, h1, ok := size(1, 1)
	if !ok {
		return nil, false
	}
	w2, h2, ok := size(2, 2)
	if !ok {
		return nil, false
	}
	solve := func(want, at1, at2 int) int {
		step := at2 - at1
		if step <= 0 {
			return 1
		}
		return max(1, 1+(want-at1)/step)
	}
	c := solve(cols, w1, w2)
	r := solve(rows, h1, h2)
	// Two or three steps of settling: the intercept is exact but the rounding
	// is not, and a frame one copy too wide is worse than one copy too narrow.
	for range 4 {
		w, h, ok := size(c, r)
		if !ok {
			return nil, false
		}
		grew := false
		if w > cols && c > 1 {
			c, grew = c-1, true
		} else if w2, _, _ := size(c+1, r); w2 <= cols {
			c, grew = c+1, true
		}
		if h > rows && r > 1 {
			r, grew = r-1, true
		} else if _, h2, _ := size(c, r+1); h2 <= rows {
			r, grew = r+1, true
		}
		if !grew {
			break
		}
	}
	w, h, ok := size(c, r)
	if !ok || w > cols || h > rows {
		return nil, false
	}
	l, err := Compose(Spec{Across: across, Down: down, Corner: corner, Cols: c, Rows: r})
	if err != nil {
		return nil, false
	}
	return l.Grid(), true
}
