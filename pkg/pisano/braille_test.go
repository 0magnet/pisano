package pisano

import (
	"strings"
	"testing"
)

// A braille cell is two dots across and four down, so the dot resolution is
// twice the width and four times the height in characters.
func TestBrailleSizeIsTwoByFourPerCharacter(t *testing.T) {
	b := NewBraille(10, 5)
	x, y := b.Size()
	if x != 20 || y != 20 {
		t.Errorf("a 10x5 canvas has %dx%d dots, want 20x20", x, y)
	}
}

// A canvas has to have somewhere to draw, so a zero or negative size becomes
// the smallest usable one rather than a canvas that panics on the first Set.
func TestNewBrailleClampsToAtLeastOneCell(t *testing.T) {
	for _, tc := range [][2]int{{0, 0}, {-5, 3}, {3, -5}, {-1, -1}} {
		b := NewBraille(tc[0], tc[1])
		x, y := b.Size()
		if x < 2 || y < 4 {
			t.Errorf("NewBraille(%d, %d) has %dx%d dots, want at least one cell", tc[0], tc[1], x, y)
		}
		b.Set(0, 0, "") // must not panic
	}
}

func TestSetLightsADot(t *testing.T) {
	b := NewBraille(2, 1)
	if got := b.String(false); strings.TrimRight(got, "\n") != "" {
		t.Errorf("a fresh canvas renders %q, want nothing", got)
	}
	b.Set(0, 0, "")
	if got := b.String(false); !strings.ContainsRune(got, '⠁') {
		t.Errorf("setting the top-left dot rendered %q", got)
	}
}

// Every dot in a cell has its own bit, so lighting all eight gives the full
// block rather than colliding and losing some.
func TestAllEightDotsInACellGiveTheFullBlock(t *testing.T) {
	b := NewBraille(1, 1)
	for y := 0; y < 4; y++ {
		for x := 0; x < 2; x++ {
			b.Set(x, y, "")
		}
	}
	if got := strings.TrimRight(b.String(false), "\n"); got != "⣿" {
		t.Errorf("all eight dots rendered %q, want the full block", got)
	}
}

// Dots must not bleed between cells: lighting one cell leaves its neighbor
// blank.
func TestDotsStayInTheirOwnCell(t *testing.T) {
	b := NewBraille(2, 1)
	b.Set(0, 0, "")
	line := strings.TrimRight(b.String(false), "\n")
	r := []rune(line)
	if len(r) != 1 {
		t.Fatalf("one lit cell rendered %d characters: %q", len(r), line)
	}
}

// Anything off the canvas is dropped, which is what makes the canvas a
// viewport onto a larger drawing rather than a bounds error.
func TestSetOutsideTheCanvasIsDropped(t *testing.T) {
	b := NewBraille(2, 2)
	dx, dy := b.Size()
	for _, p := range [][2]int{
		{-1, 0}, {0, -1}, {dx, 0}, {0, dy}, {dx + 100, dy + 100}, {-100, -100},
	} {
		b.Set(p[0], p[1], "") // must not panic
	}
	if got := strings.TrimSpace(b.String(false)); got != "" {
		t.Errorf("dots outside the canvas were drawn: %q", got)
	}
}

// Bresenham is what makes the arbitrary angles of the circular designs come
// out evenly stepped rather than in blocks.
func TestLineDrawsBetweenItsEnds(t *testing.T) {
	b := NewBraille(10, 3)
	b.Line(0, 0, 19, 11, "")
	out := b.String(false)
	if strings.TrimSpace(out) == "" {
		t.Fatal("a diagonal line drew nothing")
	}
	// Both ends have to be lit, or the line is short at one end.
	corner := NewBraille(10, 3)
	corner.Set(0, 0, "")
	corner.Set(19, 11, "")
	first := []rune(strings.SplitN(out, "\n", 2)[0])
	if len(first) == 0 || first[0] == ' ' {
		t.Errorf("the line does not start at its first end:\n%s", out)
	}
}

func TestLineIsTheSameBothWays(t *testing.T) {
	forward := NewBraille(12, 4)
	forward.Line(1, 1, 22, 14, "")
	backward := NewBraille(12, 4)
	backward.Line(22, 14, 1, 1, "")
	if forward.String(false) != backward.String(false) {
		t.Errorf("a line drawn backwards differs:\n%s\n---\n%s",
			forward.String(false), backward.String(false))
	}
}

func TestLineOfNoLengthIsASingleDot(t *testing.T) {
	b := NewBraille(4, 2)
	b.Line(3, 3, 3, 3, "")
	if got := strings.TrimSpace(b.String(false)); got == "" {
		t.Error("a line from a point to itself drew nothing")
	}
}

func TestHorizontalAndVerticalLines(t *testing.T) {
	h := NewBraille(6, 1)
	h.Line(0, 0, 11, 0, "")
	if got := strings.TrimRight(h.String(false), "\n"); len([]rune(got)) != 6 {
		t.Errorf("a full-width horizontal line covered %d cells, want 6: %q", len([]rune(got)), got)
	}

	v := NewBraille(1, 3)
	v.Line(0, 0, 0, 11, "")
	lines := strings.Split(strings.TrimRight(v.String(false), "\n"), "\n")
	if len(lines) != 3 {
		t.Errorf("a full-height vertical line covered %d rows, want 3", len(lines))
	}
}

// Trailing blanks are trimmed so the output pastes cleanly; a blank cell is a
// space rather than U+2800 so the terminal background shows through.
func TestBlankCellsAreSpacesAndTrailingOnesAreTrimmed(t *testing.T) {
	b := NewBraille(8, 1)
	b.Set(0, 0, "")
	b.Set(5, 0, "") // leaves a gap, then trailing blanks after it
	line := strings.TrimRight(b.String(false), "\n")

	if strings.ContainsRune(line, '⠀') {
		t.Error("a blank cell rendered as the blank braille pattern rather than a space")
	}
	if strings.HasSuffix(line, " ") {
		t.Errorf("trailing blanks were not trimmed: %q", line)
	}
	if !strings.Contains(line, " ") {
		t.Errorf("the gap between the two lit cells was lost: %q", line)
	}
}

// Color is written as escapes only when asked for, so the plain output can go
// into a file or a README.
func TestColorizeIsOptional(t *testing.T) {
	b := NewBraille(4, 1)
	b.Set(0, 0, "\x1b[31m")
	b.Set(3, 0, "\x1b[32m")

	if got := b.String(false); strings.Contains(got, "\x1b[") {
		t.Errorf("the plain rendering contains escapes: %q", got)
	}
	got := b.String(true)
	if !strings.Contains(got, "\x1b[31m") || !strings.Contains(got, "\x1b[32m") {
		t.Errorf("the colored rendering lost a color: %q", got)
	}
	if !strings.Contains(got, "\x1b[0m") {
		t.Errorf("the colored rendering never resets: %q", got)
	}
}

// A run of cells sharing a color says it once rather than per character,
// which is most of the bytes in a large drawing.
func TestARunOfOneColorIsWrittenOnce(t *testing.T) {
	b := NewBraille(6, 1)
	for x := 0; x < 12; x++ {
		b.Set(x, 0, "\x1b[31m")
	}
	if n := strings.Count(b.String(true), "\x1b[31m"); n != 1 {
		t.Errorf("a run of one color wrote it %d times, want once", n)
	}
}

func TestEveryRowIsPresentInTheOutput(t *testing.T) {
	b := NewBraille(3, 4)
	for row := 0; row < 4; row++ {
		b.Set(0, row*4, "")
	}
	if got := len(strings.Split(strings.TrimRight(b.String(false), "\n"), "\n")); got != 4 {
		t.Errorf("a four-row canvas rendered %d rows", got)
	}
}
