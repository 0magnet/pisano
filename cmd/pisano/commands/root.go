// Package commands cmd/pisano/commands/root.go
package commands

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

// RootCmd is the pisano command tree. It is exported so the root pisano.go can
// style it and so anything else that wants these subcommands can graft the
// whole tree onto its own root, which is how skywire composes its binaries.
var RootCmd = &cobra.Command{
	Use:   "pisano <command>",
	Short: "designs from reducing an integer sequence modulo m",
	Long: `pisano — the designs that fall out of reducing an integer sequence modulo m

Reduce a sequence mod m and the remainders repeat; the length of the repeat is
the Pisano period. Two constructions draw it:

  circle   space m points around a circle and join them in the order the
           remainders appear
  turtle   read each term as a turn — odd left, even right, zero neither — and
           walk it

Both consume the same period, so neither can disagree with the arithmetic.
Which sequence gets reduced is a flag: nothing about either construction is
particular to Fibonacci.`,
	SilenceErrors: true,
	SilenceUsage:  true,
	Run: func(cmd *cobra.Command, _ []string) {
		cmd.Help() //nolint:errcheck
	},
}

func init() {
	RootCmd.AddCommand(
		periodCmd,
		circleCmd,
		turtleCmd,
		sweepCmd,
		tuiCmd,
		galleryCmd,
	)
}

// Execute runs the command tree against the process.
func Execute() {
	if code := Run(context.Background(), os.Args[1:], os.Stdout, os.Stderr); code != 0 {
		os.Exit(code)
	}
}

// Run executes the command tree with the given arguments and writers, and
// returns an exit status. It is what a host that is not an operating system
// calls: a shell applet has arguments and a pair of pipes, not os.Args and a
// process to exit.
func Run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	RootCmd.SetArgs(args)
	RootCmd.SetOut(stdout)
	RootCmd.SetErr(stderr)
	if err := RootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintln(stderr, "pisano:", err)
		return 1
	}
	return 0
}
