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
