// Package flags pkg/flags/flags.go
package flags

import (
	"github.com/spf13/cobra"
)

// InitFlags installs the help renderer on a command tree, and the hidden --help
// flag the tree needs to be asked for it. Pass usage true to show the "Usage:"
// line above the listing.
//
// cobra looks up HelpFunc and UsageFunc through the parent chain, so setting
// them on the root covers every subcommand.
func InitFlags(cmd *cobra.Command, usage bool) {
	var helpflag bool
	cmd.PersistentFlags().BoolVarP(&helpflag, "help", "h", false, "show help menu")
	cmd.PersistentFlags().MarkHidden("help") //nolint:errcheck,gosec

	InitStyle(cmd)
	if !usage {
		cmd.SetHelpFunc(func(c *cobra.Command, _ []string) {
			writeHelp(c.OutOrStdout(), c, false)
		})
	}
}

// InitStyle installs the renderer without touching flags, for a subcommand root
// that already inherits --help from its parent.
func InitStyle(cmd *cobra.Command) {
	cmd.SetHelpFunc(func(c *cobra.Command, _ []string) {
		writeHelp(c.OutOrStdout(), c, true)
	})
	cmd.SetUsageFunc(func(c *cobra.Command) error {
		writeUsage(c.OutOrStderr(), c, true)
		return nil
	})
}
