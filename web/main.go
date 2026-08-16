//go:build js && wasm

// pisano in the browser: the same designs, driven from a shell.
//
// websh supplies the terminal (xterm-go), the shell language (0magnet/sh) and a
// virtual filesystem; this binary adds the pisano command to it. The viewer is
// the identical tui.Model the desktop binary runs — see runViewer in applet.go,
// which makes the same four calls Bubble Tea makes there.
package main

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
	"syscall/js"

	"github.com/0magnet/afero"
	xterm "github.com/0magnet/xterm-go"
	"github.com/0magnet/xterm-go/vt"

	"github.com/0magnet/websh/shell"
	"github.com/0magnet/websh/shell/browser"
)

// termWriter converts LF to CRLF on the way to the terminal.
type termWriter struct{ term *xterm.Terminal }

func (w termWriter) Write(p []byte) (int, error) {
	w.term.WriteString(strings.ReplaceAll(string(p), "\n", "\r\n"))
	return len(p), nil
}

type session struct {
	term   *xterm.Terminal
	sh     *shell.Shell
	editor *shell.LineEditor

	running   bool
	rawInput  bool
	cancelRun context.CancelFunc
	stdinW    *io.PipeWriter
	lines     chan string
}

func (s *session) prompt() string {
	dir := s.sh.Dir()
	if strings.HasPrefix(dir, "/home/user") {
		dir = "~" + dir[len("/home/user"):]
	}
	if s.sh.Pending() {
		return "\x1b[1;33m>\x1b[0m "
	}
	return "\x1b[1;35mpisano\x1b[0m:\x1b[1;34m" + dir + "\x1b[0m$ "
}

func (s *session) writePrompt() { s.term.WriteString(s.prompt()) }

func main() {
	doc := js.Global().Get("document")
	container := doc.Call("getElementById", "terminal")

	opts := vt.NewOptions()
	opts.Scrollback = 2000
	term := xterm.New(opts)
	term.Open(container)
	term.Fit()
	if err := term.EnableWebGL(); err != nil {
		js.Global().Get("console").Call("log", "webgl unavailable, using DOM renderer: "+err.Error())
	}

	js.Global().Call("addEventListener", "resize", js.FuncOf(func(js.Value, []js.Value) any {
		term.Fit()
		return nil
	}))

	out := termWriter{term}
	stdinR, stdinW := io.Pipe()

	vfs := afero.NewMemMapFs()
	if err := shell.Seed(vfs); err != nil {
		term.WriteString("failed to seed filesystem: " + err.Error() + "\r\n")
	}

	sh, err := shell.New(vfs, stdinR, out, out)
	if err != nil {
		term.WriteString("failed to start shell: " + err.Error() + "\r\n")
		select {}
	}

	browser.Register() // curl, download, upload, js, logs, pbcopy...
	registerPisano(sh)
	if err := sh.PopulateBin(); err != nil {
		term.WriteString("failed to populate /bin: " + err.Error() + "\r\n")
	}

	s := &session{term: term, sh: sh, stdinW: stdinW, lines: make(chan string, 8)}
	s.editor = &shell.LineEditor{
		Echo: func(str string) { term.WriteString(str) },
		Redraw: func(content string, back int) {
			line := "\r\x1b[2K" + s.prompt() + content
			if back > 0 {
				line += fmt.Sprintf("\x1b[%dD", back)
			}
			term.WriteString(line)
		},
		Submit:    func(line string) { s.lines <- line },
		Interrupt: func() { s.sh.CancelPending(); s.writePrompt() },
		EOF:       func() { term.WriteString("\r\n"); s.writePrompt() },
		ClearScreen: func() {
			term.WriteString("\x1b[2J\x1b[H")
			s.writePrompt()
			term.WriteString(s.editor.Line())
		},
		Complete: func(word string, isFirstWord bool) []string {
			if isFirstWord && !strings.Contains(word, "/") {
				var names []string
				for _, n := range append(shell.AppletNames(), shellBuiltins...) {
					if strings.HasPrefix(n, word) {
						names = append(names, n)
					}
				}
				sort.Strings(names)
				return names
			}
			dir, base := filepath.Split(word)
			search := dir
			if !filepath.IsAbs(search) {
				search = filepath.Join(sh.Dir(), dir)
			}
			infos, err := afero.ReadDir(sh.FS, filepath.Clean(search))
			if err != nil {
				return nil
			}
			var names []string
			for _, info := range infos {
				if !strings.HasPrefix(info.Name(), base) {
					continue
				}
				cand := dir + info.Name()
				if info.IsDir() {
					cand += "/"
				}
				names = append(names, cand)
			}
			sort.Strings(names)
			return names
		},
	}
	if err := sh.UseHistory(s.editor.History, s.editor.ClearHistory); err != nil {
		term.WriteString("history unavailable: " + err.Error() + "\r\n")
	}

	// The viewer is a full-screen applet: it wants raw bytes and the size.
	sh.RawMode = func(on bool) { s.rawInput = on }
	sh.Size = func() (int, int) { return term.Core.Cols(), term.Core.Rows() }

	term.Core.OnData = func(data string) {
		if s.running {
			if s.rawInput {
				go shell.Write(s.stdinW, []byte(data))
				return
			}
			if strings.Contains(data, "\x03") {
				if s.cancelRun != nil {
					s.cancelRun()
				}
				return
			}
			term.WriteString(strings.ReplaceAll(data, "\r", "\r\n"))
			go shell.Write(s.stdinW, []byte(strings.ReplaceAll(data, "\r", "\n")))
			return
		}
		s.editor.Input(data)
	}

	go s.run()

	term.WriteString("\x1b[1;35mpisano\x1b[0m — Pisano period designs, in a shell, in your browser\r\n")
	term.WriteString("\x1b[2mthe terminal and shell are \x1b[0m\x1b[1mwebsh\x1b[0m\x1b[2m; the designs are the same Go package the CLI uses\x1b[0m\r\n\r\n")
	term.WriteString("try: \x1b[1mpisano tui\x1b[0m · \x1b[1mpisano turtle --mod 25\x1b[0m · \x1b[1mpisano sweep\x1b[0m\r\n")
	term.WriteString("     \x1b[1mpisano circle --mod 1-40 -o sheet.svg && download sheet.svg\x1b[0m\r\n\r\n")
	s.writePrompt()

	select {}
}

func (s *session) run() {
	for line := range s.lines {
		if !s.sh.Pending() {
			s.editor.AddHistory(line)
		}
		ctx, cancel := context.WithCancel(context.Background())
		s.cancelRun = cancel
		s.running = true

		_, err := s.sh.Run(ctx, line)

		s.running = false
		s.cancelRun = nil
		cancel()

		if err != nil {
			if msg := err.Error(); !strings.HasPrefix(msg, "exit status") {
				s.term.WriteString("pisano: " + strings.ReplaceAll(msg, "\n", "\r\n") + "\r\n")
			}
		}
		s.writePrompt()
	}
}

var shellBuiltins = []string{
	"cd", "pwd", "echo", "printf", "read", "exit", "export", "unset",
	"source", "test", "true", "false", "set", "shift", "local",
	"declare", "eval", "alias", "unalias", "type", "return", "break",
	"continue", "pushd", "popd", "dirs", "let", "getopts", "wait",
	"jobs", "kill", "disown", "fg", "bg", "enable", "compgen", "history",
	"builtin", "umask", "times", "trap", "shopt", "mapfile", "readarray",
}
