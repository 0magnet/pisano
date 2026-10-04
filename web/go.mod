module github.com/0magnet/pisano/web

go 1.26.6

require (
	github.com/0magnet/afero v1.15.1-0.20261003211811-482680d00992
	github.com/0magnet/desk v0.0.1-0.20261004200012-7ac6a12489c1
	github.com/0magnet/desk/panes v0.0.1-0.20261004200012-7ac6a12489c1
	github.com/0magnet/pisano v0.0.0-20261004112647-764f40cafdc3
	github.com/0magnet/sh/v3 v3.13.2-0.20261004194540-aa2d6e4a31a5
	github.com/0magnet/websh v0.0.1-0.20261004183953-6a1b7c7a18ec
	github.com/charmbracelet/colorprofile v0.4.3
)

require (
	github.com/0magnet/bubbletea/v2 v2.0.9-0.20261004194435-f5a3ab217440
	github.com/0magnet/calvin v0.0.0
)

require (
	github.com/0magnet/u-root v0.16.1-0.20261003214924-44e47b732754 // indirect
	github.com/0magnet/winbox-go v0.0.0 // indirect
	github.com/0magnet/xterm-go v0.0.1-0.20261004020305-36b45f096b30 // indirect
	github.com/benhoyt/goawk v1.32.0 // indirect
	github.com/charmbracelet/ultraviolet v0.0.0-20261001125412-878653296cfd // indirect
	github.com/charmbracelet/x/ansi v0.11.8 // indirect
	github.com/charmbracelet/x/term v0.2.2 // indirect
	github.com/charmbracelet/x/termios v0.1.1 // indirect
	github.com/charmbracelet/x/windows v0.2.2 // indirect
	github.com/clipperhouse/displaywidth v0.11.0 // indirect
	github.com/clipperhouse/uax29/v2 v2.7.0 // indirect
	github.com/dustin/go-humanize v1.1.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/itchyny/gojq v0.12.19 // indirect
	github.com/itchyny/timefmt-go v0.1.9 // indirect
	github.com/lucasb-eyer/go-colorful v1.4.1 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	// Held here: from v0.0.27 go-runewidth builds a width lookup table in a
	// package init, and TinyGo evaluates package initialisers at compile time
	// with an interpreter that gives up on it — "interp: running for more than
	// 3m0s, timing out". Nothing here needs what the newer versions added, and
	// the browser build is not worth losing over it. charmbracelet/x/ansi
	// follows it down: v0.11.8 is the release that requires the newer one.
	github.com/mattn/go-runewidth v0.0.30 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/spf13/cobra v1.10.2 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/xo/terminfo v1.2.0 // indirect
	golang.org/x/sync v0.23.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/term v0.46.0 // indirect
	golang.org/x/text v0.42.0 // indirect
)
