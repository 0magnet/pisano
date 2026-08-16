// Package flags pkg/flags/flags.go
package flags

import (
	cc "github.com/ivanpirog/coloredcobra"
	"github.com/spf13/cobra"
)

// InitFlags sets the help templates and colour styling on a command tree.
// Pass usage true to show the "Usage:" line above the command listing.
func InitFlags(cmd *cobra.Command, usage bool) {
	var helpflag bool
	if usage {
		cmd.SetUsageTemplate(helpUsage)
	} else {
		cmd.SetUsageTemplate(help)
		cmd.SetHelpTemplate(helpTemplateNoUsage)
	}
	cmd.PersistentFlags().BoolVarP(&helpflag, "help", "h", false, "show help menu")
	cmd.PersistentFlags().MarkHidden("help") //nolint:errcheck,gosec

	initColoredCobra(cmd)
}

// InitStyle applies the templates and colours without touching flags, for a
// subcommand root that already inherits --help from its parent.
func InitStyle(cmd *cobra.Command) {
	cmd.SetUsageTemplate(helpUsage)
	initColoredCobra(cmd)
}

func initColoredCobra(cmd *cobra.Command) {
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

const helpTemplateNoUsage = `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}
{{end}}` + help

// help is the body shared by both templates. coloredcobra finds `.Name` and
// `.Short` by regex to wrap them in style functions, so the plain
// `range .Commands` form has to stay exactly as it is — pulling the command
// list into a template variable would silently lose the colouring.
//
// Every block opens with its own leading newline and closes without a trailing
// one, so a command with no subcommands or no examples does not leave a gap
// where they would have gone.
const help = `{{if gt (len .Aliases) 0}}
Aliases:
  {{.NameAndAliases}}
{{end}}{{if .HasExample}}
Examples:
{{.Example}}
{{end}}{{if .HasAvailableSubCommands}}
Available Commands:{{range .Commands}}{{if and (ne .Name "completion") .IsAvailableCommand}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}
{{end}}{{if .HasAvailableLocalFlags}}
Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}
{{end}}{{if .HasAvailableInheritedFlags}}
Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}
{{end}}`

const helpUsage = `Usage:
  {{.UseLine}}
` + help
