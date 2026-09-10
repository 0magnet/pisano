package border

import "testing"

// Four quarter turns is where you started. If the arm permutation is wrong this
// is the first thing that breaks, and it breaks silently otherwise — a figure
// that is subtly mis-turned still looks like a figure.
func TestRotateFourTimesIsIdentity(t *testing.T) {
	g := NewGrid([]string{"┌┐", "└┼╴", "┌┼┐┌┐"})
	want := g.String()
	if got := g.Rotate(4).String(); got != want {
		t.Errorf("four turns changed the figure:\ngot\n%s\nwant\n%s", got, want)
	}
}

// Each character must turn into the next one round, not stay put.
func TestGlyphsTurnWithTheGrid(t *testing.T) {
	for _, tc := range []struct{ in, want rune }{
		{'┌', '┐'}, {'┐', '┘'}, {'┘', '└'}, {'└', '┌'},
		{'─', '│'}, {'│', '─'},
		{'├', '┬'}, {'┬', '┤'}, {'┤', '┴'}, {'┴', '├'},
		{'┼', '┼'}, {'╵', '╶'}, {'╶', '╷'}, {'╷', '╴'}, {'╴', '╵'},
		{' ', ' '}, {'x', 'x'}, // not box characters: left alone
	} {
		if got := RotateGlyphCW(tc.in); got != tc.want {
			t.Errorf("RotateGlyphCW(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// Merging is what keeps a crossing a crossing.
func TestMergeGlyphMakesJunctions(t *testing.T) {
	for _, tc := range []struct{ old, add, want rune }{
		{'─', '│', '┼'},
		{'┌', '─', '┬'},
		{'└', '─', '┴'},
		{'│', '╶', '├'},
		{' ', '─', '─'},
		{'─', ' ', '─'},
		{'┼', '─', '┼'},
	} {
		if got := MergeGlyph(tc.old, tc.add); got != tc.want {
			t.Errorf("MergeGlyph(%q,%q) = %q, want %q", tc.old, tc.add, got, tc.want)
		}
	}
}

// Mirroring twice is the identity, and a mirror is not a rotation.
func TestMirrorsAreInvolutions(t *testing.T) {
	g := NewGrid([]string{"┌─┐", "│ ╵", "└╴ "})
	if g.MirrorH().MirrorH().String() != g.String() {
		t.Error("MirrorH twice is not the identity")
	}
	if g.MirrorV().MirrorV().String() != g.String() {
		t.Error("MirrorV twice is not the identity")
	}
	if g.MirrorH().String() == g.String() {
		t.Error("this figure is not H-symmetric; mirroring should change it")
	}
}

// Canonical must collapse every orientation of one figure and separate two
// different ones.
func TestCanonicalIdentifiesOrientations(t *testing.T) {
	g := NewGrid([]string{"┌─┐", "│ ╵", "└╴ "})
	want := g.Canonical()
	cur := g
	for i := 0; i < 4; i++ {
		cur = cur.RotateCW()
		if got := cur.Canonical(); got != want {
			t.Errorf("turn %d has a different canonical form", i+1)
		}
	}
	if g.MirrorH().Canonical() != want {
		t.Error("the mirror has a different canonical form")
	}
	other := NewGrid([]string{"┼┼", "┼┼"})
	if other.Canonical() == want {
		t.Error("two different figures share a canonical form")
	}
}

// Stubs are the loose ends: an open path has two, a closed one none.
func TestStubsFindLooseEnds(t *testing.T) {
	open := NewGrid([]string{"╶─┐", "  ╵"})
	if got := len(open.Stubs()); got != 2 {
		t.Errorf("open path has %d stubs, want 2", got)
	}
	closed := NewGrid([]string{"┌┐", "└┘"})
	if got := len(closed.Stubs()); got != 0 {
		t.Errorf("closed path has %d stubs, want 0", got)
	}
}

// Components measures the drawing, not the pixels: cells that merely touch are
// not joined unless they point at each other.
func TestComponentsCountsJoinedPieces(t *testing.T) {
	one := NewGrid([]string{"┌┐", "└┘"})
	if got := one.Components(); got != 1 {
		t.Errorf("a closed loop is %d pieces, want 1", got)
	}
	// Two loops side by side, touching but not joined.
	two := NewGrid([]string{"┌┐┌┐", "└┘└┘"})
	if got := two.Components(); got != 2 {
		t.Errorf("two adjacent loops are %d pieces, want 2", got)
	}
	// A rule laid across them joins them into one.
	two.RuleH(0, 4, 0)
	if got := two.Components(); got != 1 {
		t.Errorf("after threading a rule they are %d pieces, want 1", got)
	}
}
