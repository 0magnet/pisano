//go:build js && wasm

// Package app is the pisano command, as a websh applet.
//
// It runs the same cobra tree the desktop binary does. It used to parse its own
// flags, on the grounds that -o had to write into the shell's filesystem rather
// than the operating system's — which was true, and not a reason to have a
// second command line. The two drifted, as duplicated things do: the browser
// quietly lacked --theme, --split, --labels, --points and --rings, and had no
// gallery at all.
//
// What the browser actually needs is three answers the commands ask their host
// for: where files go, where "here" is, and how to get a terminal for the
// viewer. Those are supplied below; everything else is shared.
package app

import (
	"context"
	"fmt"
	"io"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"

	"github.com/0magnet/afero"
	"github.com/0magnet/sh/v3/interp"
	"github.com/0magnet/websh/shell"

	"github.com/0magnet/pisano/cmd/pisano/commands"
	"github.com/0magnet/pisano/pkg/flags"
	"github.com/0magnet/pisano/pkg/tui"
)

// Register adds the pisano command to the shell.
func Register() {
	// The same help styling the terminal binary uses. websh renders ANSI, so
	// there is no reason for the browser's help to be the plain one.
	flags.InitFlags(commands.RootCmd, true)

	shell.RegisterApplet("pisano", "Pisano period designs (try: pisano tui)",
		func(ctx context.Context, s *shell.Shell, hc *interp.HandlerContext, args []string) int {
			// websh strips the command name, and cobra expects it gone too.
			return commands.Run(
				commands.WithHost(ctx, hostFor(s, hc)),
				args, hc.Stdout, hc.Stderr)
		})
}

// hostFor answers the three questions the commands ask, for one invocation.
func hostFor(s *shell.Shell, hc *interp.HandlerContext) *commands.Host {
	return &commands.Host{
		Files:     vfs{s.FS},
		Dir:       s.Dir,
		RunViewer: viewer(s, hc),
	}
}

// vfs puts output files in the shell's filesystem, which is what makes
// `pisano circle -o s.svg && view s.svg` work: the file lands somewhere the
// other windows can see.
type vfs struct{ fs afero.Fs }

func (v vfs) Create(name string) (io.WriteCloser, error) { return v.fs.Create(name) }

func (v vfs) MkdirAll(path string, perm os.FileMode) error { return v.fs.MkdirAll(path, perm) }

// viewer runs the viewer as a Bubble Tea program over the shell's pipes.
//
// Natively Bubble Tea helps itself to stdin and stdout and asks the tty how big
// it is. Here it is told all three: the applet's reader and writer, and the
// size the terminal reports. Beyond that it is the same program — the same
// model, the same event loop — which is the point of the fork that made Bubble
// Tea build for WebAssembly at all.
func viewer(s *shell.Shell, hc *interp.HandlerContext) func(tui.Options) error {
	return func(opt tui.Options) error {
		cols, rows := 80, 24
		if s.Size != nil {
			cols, rows = s.Size()
		}
		// Raw input: no echo, no newline translation, and Ctrl+C arrives as a
		// keystroke for the program to honour rather than as a cancellation.
		if s.RawMode != nil {
			s.RawMode(true)
			defer s.RawMode(false)
		}

		p := tui.Program(opt,
			tea.WithInput(hc.Stdin),
			tea.WithOutput(crlf{hc.Stdout}),
			tea.WithWindowSize(cols, rows),
			// There are no signals to catch in a page.
			tea.WithoutSignalHandler(),
			// Bubble Tea profiles its output to decide how much colour to
			// emit, and an io.Pipe is not a terminal, so it would strip every
			// escape the model writes. The terminal on the other end of that
			// pipe is a real one; it is only this side that cannot tell.
			tea.WithColorProfile(colorprofile.ANSI256),
		)
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("tui: %w", err)
		}
		return nil
	}
}

// crlf translates line ends on the way to the terminal, which expects them the
// way a serial line did and not the way a Unix program writes them.
type crlf struct{ w io.Writer }

func (c crlf) Write(p []byte) (int, error) {
	n := len(p)
	out := make([]byte, 0, n+16)
	for i := 0; i < n; i++ {
		if p[i] == '\n' && (i == 0 || p[i-1] != '\r') {
			out = append(out, '\r')
		}
		out = append(out, p[i])
	}
	if _, err := c.w.Write(out); err != nil {
		return 0, err
	}
	return n, nil
}
