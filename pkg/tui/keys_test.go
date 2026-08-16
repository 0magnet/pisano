package tui

import (
	"bufio"
	"strings"
	"testing"
)

func TestReadKey(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want []string
	}{
		{"q", []string{"q"}},
		{" ", []string{" "}},
		{"\x03", []string{"ctrl+c"}},
		{"\x1b", []string{"esc"}},
		{"\x1b[A\x1b[B\x1b[C\x1b[D", []string{"up", "down", "right", "left"}},
		{"\x1bOA", []string{"up"}},
		{"\x1b[5~\x1b[6~", []string{"pgup", "pgdown"}},
		{"\x1b[H\x1b[F", []string{"home", "end"}},
		// the tilde must be eaten, or it arrives as a stray keystroke
		{"\x1b[5~q", []string{"pgup", "q"}},
		// an unrecognised sequence yields nothing but must not eat what follows
		{"\x1b[Zq", []string{"", "q"}},
		// control bytes other than ctrl+c are skipped
		{"\x01a", []string{"", "a"}},
	} {
		r := bufio.NewReader(strings.NewReader(tc.in))
		for i, want := range tc.want {
			got, err := ReadKey(r)
			if err != nil {
				t.Fatalf("%q key %d: %v", tc.in, i, err)
			}
			if got != want {
				t.Errorf("%q key %d = %q, want %q", tc.in, i, got, want)
			}
		}
	}
}

// Every name ReadKey can produce must be one the model actually binds, or the
// two drivers have quietly drifted apart.
func TestReadKeyNamesAreBound(t *testing.T) {
	produced := []string{
		"ctrl+c", "esc", "up", "down", "left", "right", "pgup", "pgdown",
		"q", " ", "a", "o", "s", "v", "m", "f", "t", "c", "r", "?", "h", "l",
		"H", "L", "k", "j", "0", "[", "]",
	}
	for _, name := range produced {
		before := frames(t, Options{Mod: 25}, 80, 24, 2)
		after := before
		after.Key(name)
		if name == "q" || name == "esc" || name == "ctrl+c" {
			if !after.quit {
				t.Errorf("%q did not quit", name)
			}
			continue
		}
		if after.quit {
			t.Errorf("%q quit unexpectedly", name)
		}
	}
}
