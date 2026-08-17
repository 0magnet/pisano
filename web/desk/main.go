//go:build js && wasm

// Command desk is pisano on a desktop: generate a design with a command in one
// window, and open it in another.
//
// This is the arrangement the plain terminal page cannot do. There, a design
// written to the virtual filesystem can only be downloaded; here a second
// window can show it, because both windows are panes in one binary over one
// filesystem. Writing the file is the whole of the hand-off — `view` only has
// to name it.
//
//	pisano circle --mod 1-40 -o sheet.svg && view sheet.svg
//
// The dependency points this way round on purpose: pisano knows about desk,
// desk knows nothing about pisano. A project brings its own commands to the
// desktop rather than the desktop growing to accommodate them.
package main

import (
	"strings"
	"syscall/js"

	"github.com/0magnet/desk"
	"github.com/0magnet/desk/panes/files"
	"github.com/0magnet/desk/panes/term"
	"github.com/0magnet/desk/panes/viewer"

	"github.com/0magnet/websh/shell/browser"

	"github.com/0magnet/pisano/web/app"
)

// The terminal opens at a prompt and nothing else. What the page is and what
// to try are on the header above it, where they can be read without being
// scrolled away by the first command.
const greeting = ""

func main() {
	if el := js.Global().Get("document").Call("getElementById", "desktop"); el.Truthy() {
		desk.SetRoot(el)
	}

	// Commands the page was linked with. They are submitted once the shell is
	// up, as though typed, so what ran is on the screen rather than only in the
	// address bar.
	linked := linkedCommands()

	desk.Register(desk.App{
		Name:      "term",
		Maximized: true,
		Title:     "terminal",
		Help:      "a shell with the pisano commands",
		Width:     780,
		Height:    470,
		Open: func(args []string) (desk.Pane, error) {
			p := term.New(greeting, "pisano")
			if len(args) > 0 {
				return p.Run(strings.Join(args, " ")), nil
			}
			// Only the terminal the page opens with runs them. A second one,
			// opened from the panel later, starts at a clean prompt.
			p.Run(linked...)
			linked = nil
			return p, nil
		},
	})
	desk.Register(desk.App{
		Name:   "files",
		Title:  "files",
		Help:   "browse the filesystem",
		Width:  560,
		Height: 420,
		Open: func(args []string) (desk.Pane, error) {
			dir := ""
			if len(args) > 0 {
				dir = args[0]
			}
			return files.New(term.FS(), dir), nil
		},
	})
	viewer.Register(term.FS())

	browser.Register() // curl, download, upload, js, logs, pbcopy...
	app.Register()     // the pisano command itself

	desk.NewPanel()

	if _, err := desk.Launch("term"); err != nil {
		js.Global().Get("console").Call("error", err.Error())
	}
	select {}
}

// Limits on what a link may carry. A URL is not a script: these are generous
// for a demonstration and small enough that a malformed or hostile link cannot
// leave the page grinding through it.
const (
	maxLinkedCommands = 10
	maxLinkedLength   = 400
)

// linkedCommands reads the commands a link asked for.
//
//	https://0magnet.github.io/pisano/?run=pisano+turtle+--mod+25
//	…?run=pisano+circle+--mod+8,13,21,34+-o+s.svg+%26%26+view+s.svg
//	…?run=first&run=second          one parameter per line, in order
//
// The query string rather than the fragment, because a fragment is not sent
// anywhere and this is meant to be shared. Newlines are stripped so that one
// parameter is one line: a link that wants two commands says so twice, where
// it can be seen in the URL.
func linkedCommands() []string {
	loc := js.Global().Get("location")
	if !loc.Truthy() {
		return nil
	}
	params := js.Global().Get("URLSearchParams")
	if !params.Truthy() {
		return nil
	}
	all := params.New(loc.Get("search")).Call("getAll", "run")
	if !all.Truthy() {
		return nil
	}

	out := make([]string, 0, all.Length())
	for i := 0; i < all.Length() && len(out) < maxLinkedCommands; i++ {
		line := strings.TrimSpace(strings.NewReplacer("\r", " ", "\n", " ").Replace(all.Index(i).String()))
		if line == "" {
			continue
		}
		if len(line) > maxLinkedLength {
			line = line[:maxLinkedLength]
		}
		out = append(out, line)
	}
	return out
}
