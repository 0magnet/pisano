package pisano

import (
	"bytes"
	"strings"
	"testing"
)

func tiles(n int) []Tile {
	out := make([]Tile, n)
	for i := range out {
		out[i] = Tile{
			Caption: "figure",
			Body:    `<circle cx="10" cy="10" r="5"/>`,
			Size:    100,
		}
	}
	return out
}

func TestParseTheme(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want Theme
	}{
		{"", ThemeAuto},
		{"auto", ThemeAuto},
		{"light", ThemeLight},
		{"dark", ThemeDark},
	} {
		got, err := ParseTheme(tc.in)
		if err != nil {
			t.Errorf("ParseTheme(%q): %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("ParseTheme(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

// An unknown name has to say what the choices are, since it comes straight
// from a command line flag.
func TestParseThemeRejectsAnUnknownName(t *testing.T) {
	for _, in := range []string{"Dark", "sepia", "1", " light"} {
		got, err := ParseTheme(in)
		if err == nil {
			t.Errorf("ParseTheme(%q) was accepted as %v", in, got)
			continue
		}
		for _, want := range []string{"auto", "light", "dark"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("the error for %q does not offer %q: %v", in, want, err)
			}
		}
	}
}

// A zero Sheet has to be usable, since that is what a caller that only wants
// the defaults passes.
func TestZeroSheetHasUsableGeometry(t *testing.T) {
	var s Sheet
	if got := s.TileSize(); got <= 0 {
		t.Errorf("TileSize = %v on a zero sheet", got)
	}
	if got := s.cell(); got <= 0 {
		t.Errorf("cell = %v on a zero sheet", got)
	}
	if got := s.Stroke(); got <= 0 {
		t.Errorf("Stroke = %v on a zero sheet", got)
	}
	if got := s.cols(9); got <= 0 {
		t.Errorf("cols = %v on a zero sheet", got)
	}
}

// The tile draws inside the cell with a margin, so it must be smaller than the
// cell or neighboring figures touch.
func TestTileFitsInsideItsCell(t *testing.T) {
	for _, cell := range []int{0, 60, 180, 512} {
		s := Sheet{Cell: cell}
		if got, cellSize := s.TileSize(), float64(s.cell()); got >= cellSize {
			t.Errorf("cell %d: the tile is %v inside a cell of %v", cell, got, cellSize)
		}
	}
}

// The stroke has a floor so a figure drawn small still has a visible line.
func TestStrokeNeverVanishes(t *testing.T) {
	for _, cell := range []int{1, 10, 60, 1000} {
		if got := (Sheet{Cell: cell}).Stroke(); got < 0.6 {
			t.Errorf("cell %d gives a stroke of %v, below the 0.6 floor", cell, got)
		}
	}
}

// Left to itself the sheet picks a roughly square grid, which is what makes a
// page of figures readable rather than a single long row.
func TestColsPicksARoughlySquareGrid(t *testing.T) {
	var s Sheet
	for _, n := range []int{1, 4, 9, 16, 30, 100} {
		c := s.cols(n)
		if c < 1 {
			t.Errorf("%d tiles gave %d columns", n, c)
			continue
		}
		if c > n {
			t.Errorf("%d tiles gave %d columns, more than there are tiles", n, c)
		}
		rows := (n + c - 1) / c
		// Square enough: neither dimension more than three times the other.
		if c > 3*rows || rows > 3*c {
			t.Errorf("%d tiles laid out %d by %d, which is not roughly square", n, c, rows)
		}
	}
}

func TestColsHonorsAnExplicitSetting(t *testing.T) {
	s := Sheet{Cols: 3}
	if got := s.cols(10); got != 3 {
		t.Errorf("cols = %d, want the 3 that was asked for", got)
	}
}

// ── WriteSVG ─────────────────────────────────────────────────────────────────

func TestWriteSVGProducesAWellFormedDocument(t *testing.T) {
	var buf bytes.Buffer
	if err := (Sheet{Cell: 120}).WriteSVG(&buf, tiles(4)); err != nil {
		t.Fatalf("WriteSVG: %v", err)
	}
	out := buf.String()
	if !strings.HasPrefix(out, "<svg") {
		t.Errorf("the document does not start with an svg element:\n%.80s", out)
	}
	if !strings.Contains(out, "</svg>") {
		t.Error("the svg element is not closed")
	}
	if o, c := strings.Count(out, "<g "), strings.Count(out, "</g>"); o != c {
		t.Errorf("%d groups opened, %d closed", o, c)
	}
	if n := strings.Count(out, "<circle"); n != 4 {
		t.Errorf("%d of the 4 tile bodies made it into the sheet", n)
	}
}

// Nothing to draw is an error rather than an empty document: a sheet with no
// figures on it is a mistake upstream, and writing one hides it.
func TestWriteSVGRefusesAnEmptySheet(t *testing.T) {
	var buf bytes.Buffer
	if err := (Sheet{}).WriteSVG(&buf, nil); err == nil {
		t.Error("a sheet with no tiles was written")
	}
}

// The size in the header has to match the grid actually drawn, or the figures
// are cropped or float in empty space.
func TestWriteSVGSizeGrowsWithTheTileCount(t *testing.T) {
	// Two rows of two must produce a taller document than one row.
	var two, four bytes.Buffer
	if err := (Sheet{Cell: 100, Cols: 2}).WriteSVG(&two, tiles(2)); err != nil {
		t.Fatal(err)
	}
	if err := (Sheet{Cell: 100, Cols: 2}).WriteSVG(&four, tiles(4)); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(two.String(), `height="100"`) {
		t.Errorf("one row is not 100 tall:\n%.120s", two.String())
	}
	if !strings.Contains(four.String(), `height="200"`) {
		t.Errorf("two rows are not 200 tall:\n%.120s", four.String())
	}
}

func TestWriteSVGIsStable(t *testing.T) {
	var a, b bytes.Buffer
	s := Sheet{Cell: 120, Labels: true}
	if err := s.WriteSVG(&a, tiles(3)); err != nil {
		t.Fatal(err)
	}
	if err := s.WriteSVG(&b, tiles(3)); err != nil {
		t.Fatal(err)
	}
	if a.String() != b.String() {
		t.Error("two writes of the same sheet differed")
	}
}

// Labels are what make a sheet readable as a reference rather than a picture.
func TestLabelsPutTheCaptionInTheDocument(t *testing.T) {
	var with, without bytes.Buffer
	if err := (Sheet{Cell: 120, Labels: true}).WriteSVG(&with, tiles(1)); err != nil {
		t.Fatal(err)
	}
	if err := (Sheet{Cell: 120}).WriteSVG(&without, tiles(1)); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(with.String(), "figure") {
		t.Error("a labeled sheet does not contain the caption")
	}
	if strings.Contains(without.String(), "figure") {
		t.Error("an unlabeled sheet contains the caption anyway")
	}
}

// The three themes must actually differ, or the fixed choices that exist for
// conversion to an image are pointless.
func TestThemesRenderDifferently(t *testing.T) {
	out := map[Theme]string{}
	for _, th := range []Theme{ThemeAuto, ThemeLight, ThemeDark} {
		var buf bytes.Buffer
		if err := (Sheet{Cell: 100, Theme: th}).WriteSVG(&buf, tiles(1)); err != nil {
			t.Fatal(err)
		}
		out[th] = buf.String()
	}
	if out[ThemeLight] == out[ThemeDark] {
		t.Error("the light and dark themes render identically")
	}
	if out[ThemeAuto] == out[ThemeLight] {
		t.Error("the automatic theme renders the same as the fixed light one")
	}
}

// Auto follows the viewer's setting, which can only be a media query — and a
// media query is exactly what does not survive conversion to an image.
func TestAutoThemeUsesAMediaQueryAndTheFixedOnesDoNot(t *testing.T) {
	render := func(th Theme) string {
		var buf bytes.Buffer
		if err := (Sheet{Cell: 100, Theme: th}).WriteSVG(&buf, tiles(1)); err != nil {
			t.Fatal(err)
		}
		return buf.String()
	}
	if !strings.Contains(render(ThemeAuto), "prefers-color-scheme") {
		t.Error("the automatic theme has no media query, so it cannot follow the viewer")
	}
	for _, th := range []Theme{ThemeLight, ThemeDark} {
		if strings.Contains(render(th), "prefers-color-scheme") {
			t.Errorf("a fixed theme carries a media query, which will not survive being converted")
		}
	}
}

// ── WriteHTML ────────────────────────────────────────────────────────────────

func TestWriteHTMLIsASelfContainedPage(t *testing.T) {
	var buf bytes.Buffer
	secs := []Section{{Title: "first", Tiles: tiles(2)}}
	if err := WriteHTML(&buf, "a page", secs); err != nil {
		t.Fatalf("WriteHTML: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"<html", "</html>", "a page", "first", "<svg"} {
		if !strings.Contains(out, want) {
			t.Errorf("the page has no %q", want)
		}
	}
	// Self-contained means nothing to fetch: every figure is inline SVG.
	if strings.Contains(out, "<img") || strings.Contains(out, "src=\"http") {
		t.Error("the page references something it would have to fetch")
	}
}

func TestWriteHTMLWithSeveralSections(t *testing.T) {
	var buf bytes.Buffer
	secs := []Section{
		{Title: "alpha", Tiles: tiles(1)},
		{Title: "beta", Tiles: tiles(2)},
	}
	if err := WriteHTML(&buf, "t", secs); err != nil {
		t.Fatalf("WriteHTML: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"alpha", "beta"} {
		if !strings.Contains(out, want) {
			t.Errorf("the page is missing the %q section", want)
		}
	}
	if n := strings.Count(out, "<circle"); n != 3 {
		t.Errorf("%d of the 3 figures made it onto the page", n)
	}
}
