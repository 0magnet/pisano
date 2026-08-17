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

	desk.Register(desk.App{
		Name:      "term",
		Maximized: true,
		Title:     "terminal",
		Help:      "a shell with the pisano commands",
		Width:     780,
		Height:    470,
		Open: func([]string) (desk.Pane, error) {
			return term.New(greeting, "pisano"), nil
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
