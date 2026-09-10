// Package border builds borders out of Pisano turtle figures.
//
// A figure is a closed or open path drawn in box-drawing characters. Laying
// several of them end to end into a frame needs three things the drawing itself
// does not carry: which way the figure travels, whether it can be turned
// without its corners pointing the wrong way, and whether the result is one
// continuous line or a scatter of pieces. This package answers those, composes
// the frame, and renders it either as a grid (for a terminal) or as a placement
// list (for HTML, where a figure can be turned by an angle a grid cannot hold).
package border

// The arm algebra.
//
// Every box-drawing character is exactly a set of arms — up, right, down, left.
// Working in arms rather than in characters is what makes turning and joining
// exact: a quarter turn permutes the arms, a junction unions them, and the
// character is looked back up at the end. Nothing is approximated and there is
// no table of special cases.

// Arm is one of the four directions a box character can reach.
type Arm uint8

// The four arms, as bits so a character is a set.
const (
	Up Arm = 1 << iota
	Right
	Down
	Left
)

// Blank is the empty cell: no arms.
const Blank = ' '

var (
	runeArms = map[rune]Arm{
		Blank: 0,
		'│':   Up | Down, '─': Left | Right,
		'┌': Right | Down, '┐': Left | Down, '└': Up | Right, '┘': Up | Left,
		'├': Up | Right | Down, '┤': Up | Left | Down,
		'┬': Left | Right | Down, '┴': Left | Right | Up,
		'┼': Up | Right | Down | Left,
		'╵': Up, '╶': Right, '╷': Down, '╴': Left,
	}
	armsRune = func() map[Arm]rune {
		m := map[Arm]rune{}
		for r, a := range runeArms {
			if _, seen := m[a]; !seen {
				m[a] = r
			}
		}
		return m
	}()
)

// Arms reports the arms of a box character, and whether it is one at all.
// Anything that is not a box character is left alone by every operation here.
func Arms(r rune) (Arm, bool) {
	a, ok := runeArms[r]
	return a, ok
}

// Glyph is the box character with exactly these arms, or the blank when there
// is none — which cannot happen for any subset of four arms, but a caller
// building an arm set by hand deserves a defined answer.
func Glyph(a Arm) rune {
	if r, ok := armsRune[a]; ok {
		return r
	}
	return Blank
}

// mapArms rewrites a character by sending each of its arms somewhere.
func mapArms(r rune, f func(Arm) Arm) rune {
	a, ok := runeArms[r]
	if !ok {
		return r
	}
	var out Arm
	for _, arm := range []Arm{Up, Right, Down, Left} {
		if a&arm != 0 {
			out |= f(arm)
		}
	}
	return Glyph(out)
}

func armRotCW(a Arm) Arm {
	switch a {
	case Up:
		return Right
	case Right:
		return Down
	case Down:
		return Left
	}
	return Up
}

func armMirrorH(a Arm) Arm { // left <-> right
	switch a {
	case Left:
		return Right
	case Right:
		return Left
	}
	return a
}

func armMirrorV(a Arm) Arm { // up <-> down
	switch a {
	case Up:
		return Down
	case Down:
		return Up
	}
	return a
}

// RotateGlyphCW turns one character a quarter turn clockwise.
func RotateGlyphCW(r rune) rune { return mapArms(r, armRotCW) }

// MirrorGlyphH and MirrorGlyphV reflect one character.
func MirrorGlyphH(r rune) rune { return mapArms(r, armMirrorH) }
func MirrorGlyphV(r rune) rune { return mapArms(r, armMirrorV) }

// MergeGlyph joins two characters into the junction that carries both sets of
// arms. Drawing one over another loses what was underneath — a rule crossing a
// corner would replace it rather than becoming a tee — so anything that lays
// one figure across another goes through here.
func MergeGlyph(old, add rune) rune {
	if old == Blank || old == 0 {
		return add
	}
	if add == Blank {
		return old
	}
	a, ok1 := runeArms[old]
	b, ok2 := runeArms[add]
	if !ok1 || !ok2 {
		return add
	}
	return Glyph(a | b)
}
