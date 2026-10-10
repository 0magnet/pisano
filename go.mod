module github.com/0magnet/pisano

go 1.26.6

require (
	github.com/spf13/cobra v1.10.2
	github.com/spf13/pflag v1.0.10
)

require (
	github.com/0magnet/bubbletea/v2 v2.0.9-0.20261004194435-f5a3ab217440
	github.com/0magnet/calvin v0.0.1-0.20261010111553-d1d2be37be6a
)

require (
	github.com/charmbracelet/colorprofile v0.4.3 // indirect
	github.com/charmbracelet/ultraviolet v0.0.0-20261008173134-6b8d4baf91b4 // indirect
	github.com/charmbracelet/x/ansi v0.11.9 // indirect
	github.com/charmbracelet/x/term v0.2.2 // indirect
	github.com/charmbracelet/x/termios v0.1.1 // indirect
	github.com/charmbracelet/x/windows v0.2.2 // indirect
	github.com/clipperhouse/displaywidth v0.11.0 // indirect
	github.com/clipperhouse/uax29/v2 v2.7.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/lucasb-eyer/go-colorful v1.4.1 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	// Held here: from v0.0.27 go-runewidth builds a width lookup table in a
	// package init, and TinyGo evaluates package initialisers at compile time
	// with an interpreter that gives up on it — "interp: running for more than
	// 3m0s, timing out". Nothing here needs what the newer versions added, and
	// the browser build is not worth losing over it. charmbracelet/x/ansi
	// follows it down: v0.11.8 is the release that requires the newer one.
	github.com/mattn/go-runewidth v0.0.31 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/xo/terminfo v1.2.0 // indirect
	golang.org/x/sync v0.24.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
)
