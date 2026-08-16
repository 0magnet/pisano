package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/0magnet/pisano/pkg/pisano"
)

var sweepCmd = func() *cobra.Command {
	var (
		max    int
		table  bool
		dupes  bool
		passes int
	)
	cmd := &cobra.Command{
		Use:   "sweep",
		Short: "tabulate periods, symmetry and path shape across many moduli",
		Long: `Measure a sequence across every modulus up to --max.

This is where the two open questions get tested. Every Fibonacci Pisano period
holds one, two or four zeros; four zeros appear to force a symmetrical circular
design and one zero to forbid it. A single counterexample would settle that, so
the sweep looks for one.

It also reports what decides whether a turtle path closes, and with --dupes
groups the moduli that trace the same figure as each other.`,
		Example: `  pisano sweep --max 3000
  pisano sweep --max 60 --dupes
  pisano sweep --max 40 --table`,
		Args: cobra.NoArgs,
	}
	seq := addSeqFlags(cmd)
	cmd.Flags().IntVarP(&max, "max", "n", 200, "highest modulus to examine")
	cmd.Flags().BoolVarP(&table, "table", "t", false, "print the per-modulus listing too")
	cmd.Flags().BoolVarP(&dupes, "dupes", "d", false, "group moduli whose turtle paths match")
	cmd.Flags().IntVar(&passes, "passes", 2, "passes to compare when grouping open paths")

	cmd.RunE = func(cc *cobra.Command, _ []string) error {
		stdout := cc.OutOrStdout()
		_ = stdout
		s, err := seq()
		if err != nil {
			return err
		}
		rows := pisano.Sweep(s, pisano.SweepOptions{Max: max, Dupes: dupes, Passes: passes})
		if table {
			fmt.Fprint(stdout, pisano.Table(rows))
			fmt.Fprintln(stdout)
		}
		fmt.Fprintln(stdout, "zeros vs symmetry of the circular design")
		fmt.Fprintln(stdout, "----------------------------------------")
		fmt.Fprint(stdout, pisano.ZeroSymmetry(rows))
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "turtle path shape")
		fmt.Fprintln(stdout, "-----------------")
		fmt.Fprint(stdout, pisano.TurtleShapes(rows))
		if dupes {
			fmt.Fprintln(stdout)
			fmt.Fprintln(stdout, "turtle paths shared between moduli")
			fmt.Fprintln(stdout, "----------------------------------")
			fmt.Fprint(stdout, pisano.Duplicates(rows))
		}
		return nil
	}
	return cmd
}()
