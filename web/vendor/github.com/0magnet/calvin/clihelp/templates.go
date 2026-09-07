package clihelp

// The templates below are skywire's, so a command styled here renders the same
// way as one styled there. Both support cobra command groups (cobra >= 1.6):
// with groups defined, subcommands render under their group's title and any
// ungrouped ones fall into "Additional Commands:"; with none, the flat
// "Available Commands:" listing is produced.
//
// The non-grouped path MUST use `range .Commands` rather than a $variable, so
// coloredcobra's regex can find and wrap .Name and .Short in its style
// functions. The grouped path calls the style functions explicitly, because
// that regex cannot match a $variable or a dynamic group title. cc.Init
// registers the functions globally, so both paths render identically.
const helpTemplateNoUsage = `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}{{if .HasAvailableSubCommands}}{{if eq (len .Groups) 0}}{{HeadingStyle "Available Commands:"}}{{range .Commands}}{{if and (ne .Name "completion") .IsAvailableCommand}}
  {{rpad (CommandStyle .Name) (sum .NamePadding 12) }} {{CmdShortStyle .Short}}{{end}}{{end}}
{{else}}{{$cmds := .Commands}}{{range $group := .Groups}}
{{HeadingStyle $group.Title}}{{range $cmds}}{{if and (eq .GroupID $group.ID) (ne .Name "completion") .IsAvailableCommand}}
  {{rpad (CommandStyle .Name) (sum .NamePadding 12) }} {{CmdShortStyle .Short}}{{end}}{{end}}
{{end}}{{if not .AllChildCommandsHaveGroup}}
{{HeadingStyle "Additional Commands:"}}{{range $cmds}}{{if and (eq .GroupID "") (ne .Name "completion") .IsAvailableCommand}}
  {{rpad (CommandStyle .Name) (sum .NamePadding 12) }} {{CmdShortStyle .Short}}{{end}}{{end}}
{{end}}{{end}}

{{end}}{{if .HasAvailableLocalFlags}}{{HeadingStyle "Flags:"}}
{{FlagStyle .LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

{{HeadingStyle "Global Flags:"}}
{{FlagStyle .InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}
`

// help is used as usage template (coloredcobra will add colors to it).
// IMPORTANT: the non-grouped path MUST use `range .Commands` (not a
// $variable) so coloredcobra's regex can find and wrap `.Name` /
// `.Short` in style functions. The grouped path uses explicit style
// function calls (HeadingStyle, CommandStyle, CmdShortStyle) because
// cc's regex can't match the `$cmds` variable or dynamic group
// titles. The functions are registered globally by cc.Init, so both
// paths render identically.
const help = `{{if gt (len .Aliases) 0}}{{.NameAndAliases}}{{end}}{{if .HasAvailableSubCommands}}{{if eq (len .Groups) 0}}Available Commands:{{range .Commands}}{{if and (ne .Name "completion") .IsAvailableCommand}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{$cmds := .Commands}}{{range $group := .Groups}}
{{HeadingStyle $group.Title}}{{range $cmds}}{{if and (eq .GroupID $group.ID) (ne .Name "completion") .IsAvailableCommand}}
  {{rpad (CommandStyle .Name) (sum .NamePadding 12)}} {{CmdShortStyle .Short}}{{end}}{{end}}
{{end}}{{if not .AllChildCommandsHaveGroup}}
{{HeadingStyle "Additional Commands:"}}{{range $cmds}}{{if and (eq .GroupID "") (ne .Name "completion") .IsAvailableCommand}}
  {{rpad (CommandStyle .Name) (sum .NamePadding 12)}} {{CmdShortStyle .Short}}{{end}}{{end}}
{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}
`

const helpUsage = `Usage:
  {{.UseLine}}

` + help
