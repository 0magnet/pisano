package commands

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/0magnet/pisano/pkg/tui"
)

// FileSystem is where --out and --split write.
//
// It exists because the same commands run in two places that disagree about
// what a file is: a real one natively, and an entry in the shell's virtual
// filesystem in a browser, where os.Create writes nowhere anybody can reach.
// That disagreement, the working directory, and how the viewer gets a terminal
// are the whole of the difference between the two — which is why there is one
// command tree rather than two.
type FileSystem interface {
	Create(name string) (io.WriteCloser, error)
	MkdirAll(path string, perm os.FileMode) error
}

// Host is what the commands need from whatever is running them.
//
// It travels in the context rather than in package state because a browser can
// have two terminals open at once, each with its own filesystem view and its
// own idea of where it is; globals would let one command answer with the
// other's.
type Host struct {
	// Files is where output files go. Nil means the operating system's.
	Files FileSystem

	// Dir is the working directory relative paths resolve against. Nil means
	// the process's.
	Dir func() string

	// RunViewer starts the full-screen viewer. Nil means there is none, which
	// is the honest answer on a host that cannot give a program a terminal.
	//
	// It is here rather than shared because starting it is the one thing that
	// genuinely differs: natively the program takes stdin and stdout for
	// itself, and in a browser it is handed a reader, a writer and a size by
	// whatever draws the terminal.
	RunViewer func(tui.Options) error
}

type hostKey struct{}

// WithHost attaches a host to a context for Run to carry into the commands.
func WithHost(ctx context.Context, h *Host) context.Context {
	return context.WithValue(ctx, hostKey{}, h)
}

// hostFrom returns the context's host, falling back to the operating system.
func hostFrom(ctx context.Context) *Host {
	if ctx != nil {
		if h, ok := ctx.Value(hostKey{}).(*Host); ok && h != nil {
			return h.withDefaults()
		}
	}
	return (&Host{}).withDefaults()
}

func (h *Host) withDefaults() *Host {
	out := *h
	if out.Files == nil {
		out.Files = osFiles{}
	}
	if out.Dir == nil {
		out.Dir = osDir
	}
	if out.RunViewer == nil {
		out.RunViewer = nativeViewer // nil on a host without one
	}
	return &out
}

type osFiles struct{}

func (osFiles) Create(name string) (io.WriteCloser, error) { return os.Create(name) } //nolint:gosec // the path is the output file the caller named

func (osFiles) MkdirAll(path string, perm os.FileMode) error { return os.MkdirAll(path, perm) }

func osDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}

// resolve makes a path absolute the way this host means it.
func (h *Host) resolve(p string) string {
	if p == "" || p == "-" || filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(h.Dir(), p)
}
