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
	_, g, in, ok := FitAround(Spec{Across: across, Down: down, Corner: corner}, needW, needH)
	if !ok {
		return nil, fmt.Errorf("border: no frame holds %dx%d; the corners may be bigger than the frame would be", needW, needH)
	}
	w, h := in.X1-in.X0, in.Y1-in.Y0
	g.Overlay(in.X0+(w-needW)/2+pad, in.Y0+(h-len(lines))/2, lines)
	return g, nil
}

// FitBox is the largest frame that fits inside a box of the given size in
// characters, with the layout that made it, and whether one fits at all.
//
// The call a decorator makes: it has a box on a page, measured in character
// cells, and wants a frame drawn to fill it. Frame is the other way round — it
// has content and grows a border to hold it — and neither can be had from the
// other, because a frame's size is quantised by the figures and lands where it
// lands.
//
// Solved rather than searched. A frame's width is linear in the number of
// across copies — each advances it by exactly one travel and nothing else
// moves — so two compositions give the slope and the count falls out, with at
// most a step back for rounding.
//
// That is not a nicety. This runs per box on every resize, in a browser, and
// every trial composition lays out every piece of the frame. Growing one copy
// at a time would be a hundred of them for a wide box; even a dozen ran tinygo
// out of memory on a page with sixty-nine boxes.
func FitBox(across, down, corner Figure, cols, rows int) (Layout, Grid, bool) {
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
		return Layout{}, nil, false
	}
	w2, h2, ok := size(2, 2)
	if !ok {
		return Layout{}, nil, false
	}
	// Exactly linear, so two probes are the whole calculation. Each extra copy
	// advances the frame by one travel and nothing else moves, which is what
	// makes this arithmetic rather than a search — and the search mattered:
	// every trial composition lays out every piece, and under tinygo in a
	// browser a dozen of those per box ran the heap out.
	pick := func(want, at1, step int) int {
		if step <= 0 {
			return 1
		}
		return max(1, 1+(want-at1)/step)
	}
	c := pick(cols, w1, w2-w1)
	r := pick(rows, h1, h2-h1)
	w, h, ok := size(c, r)
	if !ok {
		return Layout{}, nil, false
	}
	// One step back apiece, in case the division rounded the wrong way.
	for w > cols && c > 1 {
		c--
		w, h, ok = size(c, r)
		if !ok {
			return Layout{}, nil, false
		}
	}
	for h > rows && r > 1 {
		r--
		w, h, ok = size(c, r)
		if !ok {
			return Layout{}, nil, false
		}
	}
	if w > cols || h > rows {
		return Layout{}, nil, false
	}
	l, err := Compose(Spec{Across: across, Down: down, Corner: corner, Cols: c, Rows: r})
	if err != nil {
		return Layout{}, nil, false
	}
	// The layout comes back too, because a caller that wants the frame colored
	// needs it: a tint is per COPY of a run, and only the layout knows which
	// copy a cell came from.
	return l, l.Grid(), true
}

// FitAround is the smallest frame whose CLEAR MIDDLE holds a box of the given
// size in characters, with the layout that made it and where that middle is.
//
// The difference from FitBox is the whole point. FitBox fits the frame inside
// the box, which puts the border on top of whatever is there; this fits the box
// inside the frame, which puts the border around it. A caller with a table to
// surround wants this one, and lines the middle up with the table rather than
// lining the frame up with it.
//
// The template's Cols and Rows are ignored — they are what this works out — but
// everything else in it is honored, so a caller can ask for a detached frame or
// the other corner bias and get one.
//
// Grown rather than solved, unlike FitBox, because the middle cannot be
// predicted from the counts: the runs are woven so their inner edge is ragged,
// the corners reach further in than the runs, and how far each run is held off
// its corners depends on where the crossing of its wave falls. The first guess
// comes from the travels and is usually within a copy, so this is a nudge and
// not a search.
func FitAround(template Spec, cols, rows int) (Layout, Grid, Rect, bool) {
	stepA := math.Hypot(float64(template.Across.DX), float64(template.Across.DY))
	stepD := math.Hypot(float64(template.Down.DX), float64(template.Down.DY))
	if stepA == 0 || stepD == 0 {
		return Layout{}, nil, Rect{}, false
	}
	s := template
	s.Cols = max(1, int(math.Ceil(float64(cols)/stepA)))
	s.Rows = max(1, int(math.Ceil(float64(rows)/stepD)))
	const tries = 40
	for range tries {
		l, err := Compose(s)
		if err != nil {
			return Layout{}, nil, Rect{}, false
		}
		g := l.Grid()
		in := g.Inside()
		w, h := in.X1-in.X0, in.Y1-in.Y0
		if w >= cols && h >= rows {
			return l, g, in, true
		}
		if w < cols {
			s.Cols++
		}
		if h < rows {
			s.Rows++
		}
	}
	return Layout{}, nil, Rect{}, false
}
