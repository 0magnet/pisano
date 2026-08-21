module github.com/0magnet/pisano/web

go 1.26

replace github.com/0magnet/pisano => ../

require (
	charm.land/bubbletea/v2 v2.0.9
	github.com/0magnet/afero v1.15.1-0.20260816202415-9f9d46a34dcd
	github.com/0magnet/desk v0.0.0-20260821232155-7ded1c828135
	github.com/0magnet/desk/panes v0.0.0-20260821223104-b7a16e4c27a2
	github.com/0magnet/pisano v0.0.0-20260821223114-79da5c869208
	github.com/0magnet/sh/v3 v3.13.2-0.20260818190530-13d0024da85c
	github.com/0magnet/websh v0.0.0-20260821231944-8cefc6a09852
	github.com/charmbracelet/colorprofile v0.4.3
)

require (
	github.com/0magnet/u-root v0.16.1-0.20260814161052-156e0b67262b // indirect
	github.com/0magnet/winbox-go v0.0.0-20260821223041-b2d40b5b492d // indirect
	github.com/0magnet/xterm-go v0.0.0-20260821223040-7fc35994fbca // indirect
	github.com/benhoyt/goawk v1.31.0 // indirect
	github.com/charmbracelet/ultraviolet v0.0.0-20260812204455-68fa937c71be // indirect
	github.com/charmbracelet/x/ansi v0.11.8 // indirect
	github.com/charmbracelet/x/term v0.2.2 // indirect
	github.com/charmbracelet/x/termios v0.1.1 // indirect
	github.com/charmbracelet/x/windows v0.2.2 // indirect
	github.com/clipperhouse/displaywidth v0.11.0 // indirect
	github.com/clipperhouse/uax29/v2 v2.7.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/itchyny/gojq v0.12.19 // indirect
	github.com/itchyny/timefmt-go v0.1.8 // indirect
	github.com/lucasb-eyer/go-colorful v1.4.1 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	// Held here: from v0.0.27 go-runewidth builds a width lookup table in a
	// package init, and TinyGo evaluates package initialisers at compile time
	// with an interpreter that gives up on it — "interp: running for more than
	// 3m0s, timing out". Nothing here needs what the newer versions added, and
	// the browser build is not worth losing over it. charmbracelet/x/ansi
	// follows it down: v0.11.8 is the release that requires the newer one.
	github.com/mattn/go-runewidth v0.0.28 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/spf13/cobra v1.10.2 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/xo/terminfo v1.0.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/term v0.45.0 // indirect
	golang.org/x/text v0.41.0 // indirect
)

replace charm.land/bubbletea/v2 => github.com/0magnet/bubbletea/v2 v2.0.9-0.20260816230205-5aaf8ac0d36c
