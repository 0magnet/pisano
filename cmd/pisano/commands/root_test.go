package commands

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/0magnet/pisano/pkg/tui"
)

// nopFiles swallows output so a command that writes a file can run in a test
// without one.
type nopFiles struct{}

type nopWriter struct{ io.Writer }

func (nopWriter) Close() error { return nil }

func (nopFiles) Create(string) (io.WriteCloser, error) { return nopWriter{io.Discard}, nil }

func (nopFiles) MkdirAll(string, os.FileMode) error { return nil }

// runWith executes one command line the way a host does, and reports the
// viewer options it asked for, if it asked for the viewer.
func runWith(t *testing.T, args ...string) (tui.Options, int, string) {
	t.Helper()
	var got tui.Options
	var out bytes.Buffer
	h := &Host{
		Files:     nopFiles{},
		Dir:       func() string { return "/" },
		RunViewer: func(o tui.Options) error { got = o; return nil },
	}
	code := Run(WithHost(context.Background(), h), args, &out, &out)
	return got, code, out.String()
}

// TestFlagsDoNotSurviveTheCommand is the browser's bug, in a test.
//
// The tree is built once at init and its flag values live in the closures that
// registered them, so a host that runs many commands in one process inherits
// the last run's flags. Natively the process exits and nothing shows; in a page
// `pisano tui --cycle 5s` left --cycle set, and the next `pisano tui --mod 109`
// stepped through the moduli on its own.
func TestFlagsDoNotSurviveTheCommand(t *testing.T) {
	if opt, code, out := runWith(t, "tui", "--cycle", "5s"); code != 0 {
		t.Fatalf("first run failed: %d %s", code, out)
	} else if opt.Cycle != 5*time.Second {
		t.Fatalf("--cycle 5s gave %v", opt.Cycle)
	}

	opt, code, out := runWith(t, "tui", "--mod", "109")
	if code != 0 {
		t.Fatalf("second run failed: %d %s", code, out)
	}
	if opt.Mod != 109 {
		t.Errorf("--mod 109 gave mod %d", opt.Mod)
	}
	if opt.Cycle != 0 {
		t.Errorf("--cycle leaked into the next command: %v", opt.Cycle)
	}
}

// TestDefaultsAreRestoredForEveryFlag checks the reset is general rather than a
// patch for the one flag that was noticed.
func TestDefaultsAreRestoredForEveryFlag(t *testing.T) {
	if _, code, out := runWith(t,
		"tui", "--seq", "lucas", "--mul", "3", "--speed", "9",
		"--render", "braille", "--cam", "page", "--trail", "comet", "--tint", "age",
		"--circle", "--paused", "--mono", "--no-mod", "--max-points", "42",
	); code != 0 {
		t.Fatalf("first run failed: %d %s", code, out)
	}

	opt, code, out := runWith(t, "tui")
	if code != 0 {
		t.Fatalf("second run failed: %d %s", code, out)
	}
	want := tui.Options{
		Seq: "fib", Mul: 1, Mod: 25, Speed: 4, MaxPts: 60000,
		Render: "box", Cam: "auto", Trail: "whole", Tint: "step",
	}
	if opt != want {
		t.Errorf("a bare tui inherited flags\n got %+v\nwant %+v", opt, want)
	}
}

// TestEachRunGetsItsOwnHost covers the quieter half of the same fault.
//
// cobra copies the root's context onto a subcommand only if the subcommand has
// none, so the second run of a command keeps the first run's context — and with
// it the first run's filesystem, working directory and terminal. In a page that
// means two open terminals writing each other's files.
func TestEachRunGetsItsOwnHost(t *testing.T) {
	first := 0
	one := &Host{
		Files: nopFiles{}, Dir: func() string { return "/one" },
		RunViewer: func(tui.Options) error { first++; return nil },
	}
	if code := Run(WithHost(context.Background(), one), []string{"tui"}, io.Discard, io.Discard); code != 0 {
		t.Fatalf("first run failed: %d", code)
	}

	second := 0
	two := &Host{
		Files: nopFiles{}, Dir: func() string { return "/two" },
		RunViewer: func(tui.Options) error { second++; return nil },
	}
	if code := Run(WithHost(context.Background(), two), []string{"tui"}, io.Discard, io.Discard); code != 0 {
		t.Fatalf("second run failed: %d", code)
	}

	if first != 1 || second != 1 {
		t.Errorf("the second run went to the first run's host: first=%d second=%d", first, second)
	}
}

// TestOutputPathDoesNotLeak is the same fault where it would silently write to
// the wrong file rather than merely animate.
func TestOutputPathDoesNotLeak(t *testing.T) {
	if _, code, out := runWith(t, "circle", "--mod", "1-4", "-o", "s.svg"); code != 0 {
		t.Fatalf("first run failed: %d %s", code, out)
	}
	_, code, out := runWith(t, "circle", "--mod", "1-4")
	if code != 0 {
		t.Fatalf("second run failed: %d %s", code, out)
	}
	if !strings.Contains(out, "<svg") {
		t.Errorf("-o leaked: the design went to a file instead of stdout\n%s", out)
	}
}
