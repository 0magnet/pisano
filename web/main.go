//go:build js && wasm

// pisano in the browser: the same designs, driven from a shell.
//
// websh supplies the terminal, the shell language and a virtual filesystem —
// all of it behind web.NewSession — and this binary adds the pisano command to
// it. The viewer is the identical tui.Model the desktop binary runs; see
// runViewer in applet.go, which makes the same four calls Bubble Tea makes
// there.
package main

import (
	"syscall/js"

	"github.com/0magnet/afero"

	"github.com/0magnet/websh/shell"
	"github.com/0magnet/websh/shell/browser"
	"github.com/0magnet/websh/web"
)

const greeting = "" +
	"\x1b[1;35mpisano\x1b[0m — Pisano period designs, in a shell, in your browser\r\n" +
	"\x1b[2mthe terminal and shell are \x1b[0m\x1b[1mwebsh\x1b[0m\x1b[2m; the designs are the same Go package the CLI uses\x1b[0m\r\n\r\n" +
	"try: \x1b[1mpisano tui\x1b[0m · \x1b[1mpisano turtle --mod 25\x1b[0m · \x1b[1mpisano sweep\x1b[0m\r\n" +
	"     \x1b[1mpisano circle --mod 1-40 -o sheet.svg && download sheet.svg\x1b[0m\r\n\r\n"

func main() {
	container := js.Global().Get("document").Call("getElementById", "terminal")

	vfs := afero.NewMemMapFs()
	if err := shell.Seed(vfs); err != nil {
		js.Global().Get("console").Call("error", "failed to seed filesystem: "+err.Error())
	}

	browser.Register() // curl, download, upload, js, logs, pbcopy...
	registerPisano()

	if _, err := web.NewSession(container, web.Options{
		FS:       vfs,
		Host:     "pisano",
		Greeting: greeting,
	}); err != nil {
		js.Global().Get("console").Call("error", err.Error())
		return
	}
	select {}
}
