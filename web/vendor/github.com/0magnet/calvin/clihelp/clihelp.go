// Package clihelp gives a cobra command the house help menu: the program's
// name in calvin's ASCII font, the build it came from underneath, blue
// coloring, and the two flags that print what the toolchain recorded.
//
// It exists because the same help menu was being written out per repo and
// drifting. pisano carried a pkg/flags that had already lost the coloring and
// the banner; the skywire tree it was copied from has both. One package that
// every command calls keeps them the same by construction.
//
// Everything it reports comes from runtime/debug.BuildInfo, which the Go
// toolchain fills in on its own. Nothing here is injected with -ldflags: a
// binary built with `go build` describes itself, and one built any other way
// says so rather than claiming a version nobody set.
package clihelp

import (
	"fmt"
	"runtime/debug"
	"strings"

	cc "github.com/0magnet/coloredcobra"
	"github.com/spf13/cobra"

	"github.com/0magnet/calvin"
)

// bi is read once. debug.ReadBuildInfo only fails in a binary the toolchain
// did not stamp, which is a test binary or a linker doing something unusual;
// every accessor below is written to answer sensibly when it is nil.
var bi, biOK = debug.ReadBuildInfo()

// Version is the main module's version — a tag for an installed binary,
// "(devel)" for one built from a working tree. Empty when unknown.
func Version() string {
	if !biOK || bi == nil {
		return ""
	}
	return bi.Main.Version
}

// GoVersion is the toolchain that compiled the binary, e.g. "go1.27.0".
func GoVersion() string {
	if !biOK || bi == nil {
		return ""
	}
	return bi.GoVersion
}

// setting reads one of the key/value pairs the toolchain records — vcs.revision,
// vcs.time, vcs.modified and friends.
func setting(key string) string {
	if !biOK || bi == nil {
		return ""
	}
	for _, s := range bi.Settings {
		if s.Key == key {
			return s.Value
		}
	}
	return ""
}

// Commit is the VCS revision the build came from, or "" outside a repository.
func Commit() string { return setting("vcs.revision") }

// Dirty reports whether the tree had uncommitted changes at build time.
func Dirty() bool { return setting("vcs.modified") == "true" }

// Date is the commit timestamp the toolchain recorded, in RFC3339.
func Date() string { return setting("vcs.time") }

// BuildInfo is the raw record, for the -d flag.
func BuildInfo() *debug.BuildInfo {
	if !biOK {
		return nil
	}
	return bi
}

// describe is the block under the banner: what this is, and where it came
// from. Each line is omitted rather than filled with "unknown" — a help menu
// saying nothing about the commit is better than one asserting it has none.
func describe() string {
	var b strings.Builder
	if v := Version(); v != "" && v != "(devel)" {
		fmt.Fprintf(&b, "\nversion %s", v)
	} else if c := Commit(); c != "" {
		short := c
		if len(short) > 12 {
			short = short[:12]
		}
		if Dirty() {
			short += "-dirty"
		}
		fmt.Fprintf(&b, "\nbuilt from %s", short)
	}
	if d := Date(); d != "" {
		fmt.Fprintf(&b, "\nbuilt %s", d)
	}
	if g := GoVersion(); g != "" {
		fmt.Fprintf(&b, "\nbuilt with %s", g)
	}
	return b.String()
}

// Banner is the Long description for a command: name in the ASCII font, then
// the build. Pass the name the user types, which is not always cmd.Name() —
// a subcommand mounted under another binary keeps its own banner.
func Banner(name string) string {
	return calvin.AsciiFont(name) + describe()
}

// BannerWith is Banner followed by a blank line and text, for commands that
// want prose under the build lines.
func BannerWith(name, text string) string {
	return Banner(name) + "\n\n" + strings.TrimRight(text, "\n")
}

// Init gives cmd the house style: the banner as its Long, the blue coloring,
// the shared templates, and the -b/-d flags. Call it on the root command after
// its subcommands are attached, since the templates are inherited through the
// parent chain and the flags are only meaningful at the root.
//
// usage controls whether the "Usage:" line appears above the command listing.
func Init(cmd *cobra.Command, name string, usage bool) {
	cmd.Long = Banner(name)
	InitStyle(cmd, usage)
	InitFlags(cmd)
}

// InitStyle applies the templates and coloring without touching Long or the
// flags. Use it for a subcommand root that already inherits both.
func InitStyle(cmd *cobra.Command, usage bool) {
	if usage {
		cmd.SetUsageTemplate(helpUsage)
	} else {
		cmd.SetUsageTemplate(help)
		cmd.SetHelpTemplate(helpTemplateNoUsage)
	}
	// The templates have to be set before cc.Init, which colorizes whatever
	// templates the command is holding.
	cc.Init(&cc.Config{
		RootCmd:         cmd,
		Headings:        cc.HiBlue + cc.Bold,
		Commands:        cc.HiBlue + cc.Bold,
		CmdShortDescr:   cc.HiBlue,
		Example:         cc.HiBlue + cc.Italic,
		ExecName:        cc.HiBlue + cc.Bold,
		Flags:           cc.HiBlue + cc.Bold,
		FlagsDescr:      cc.HiBlue,
		NoExtraNewlines: true,
		NoBottomNewline: true,
	})
}

// InitFlags adds -b and -d, and hides cobra's --help so it does not sit in the
// listing describing itself. Both are skipped when the toolchain recorded
// nothing, so a binary that cannot answer does not advertise that it can.
func InitFlags(cmd *cobra.Command) {
	var helpFlag bool
	cmd.PersistentFlags().BoolVarP(&helpFlag, "help", "h", false, "show help menu")
	_ = cmd.PersistentFlags().MarkHidden("help")

	if BuildInfo() == nil {
		return
	}
	var showAll, showVer bool
	cmd.Flags().BoolVarP(&showAll, "info", "d", false, "print runtime/debug.BuildInfo")
	cmd.Flags().BoolVarP(&showVer, "bv", "b", false, "print the main module's version")

	// Wrap rather than replace: a command with its own Run keeps it, and one
	// without still answers the flags.
	inner := cmd.Run
	innerE := cmd.RunE
	cmd.Run = nil
	cmd.RunE = func(c *cobra.Command, args []string) error {
		switch {
		case showAll:
			fmt.Fprintln(c.OutOrStdout(), BuildInfo())
			return nil
		case showVer:
			fmt.Fprintln(c.OutOrStdout(), Version())
			return nil
		}
		switch {
		case innerE != nil:
			return innerE(c, args)
		case inner != nil:
			inner(c, args)
			return nil
		}
		return c.Help()
	}
}
