// Package commands cmd/pisano/commands/root.go
package commands

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
	prepare(RootCmd, ctx)
	RootCmd.SetArgs(args)
	RootCmd.SetOut(stdout)
	RootCmd.SetErr(stderr)
	if err := RootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintln(stderr, "pisano:", err)
		return 1
	}
	return 0
}

// prepare makes the tree ready to run again, which it is not by default.
//
// The tree is built once, at init, and it keeps two things between runs. A
// process that runs one command and exits never sees either; a browser runs a
// hundred in the same process, where both are wrong.
//
// The first is flag values, which live in the closures that registered them and
// stay set. That is why `pisano tui --cycle 5s` followed by `pisano tui --mod
// 109` still stepped through the moduli on its own: nothing had put --cycle
// back. It applies to every flag, not just that one.
//
// The second is worse and quieter. cobra copies the root's context onto the
// subcommand it dispatches to, but only if the subcommand has none —
//
//	if cmd.ctx == nil {
//		cmd.ctx = c.ctx
//	}
//
// — so the second run of a subcommand keeps the first run's context, and with
// it the first run's Host: its filesystem, its working directory, its terminal.
// Two terminals open in a page would have written each other's files.
func prepare(cmd *cobra.Command, ctx context.Context) {
	cmd.SetContext(ctx)
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if !f.Changed {
			return
		}
		// Set is the only way back: the default is kept as the string it was
		// registered with, which is exactly what the parser would have been
		// handed had the flag been given explicitly.
		_ = f.Value.Set(f.DefValue)
		f.Changed = false
	})
	for _, sub := range cmd.Commands() {
		prepare(sub, ctx)
	}
}
