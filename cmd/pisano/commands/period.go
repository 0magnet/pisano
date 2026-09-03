package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/0magnet/pisano/pkg/pisano"
)

var periodCmd = func() *cobra.Command {
	var (
		mod   string
		limit int
	)
	cmd := &cobra.Command{
		Use:   "period",
		Short: "print the repeating block for a modulus",
		Long: `Print the repeating block of a sequence reduced modulo m.

The Fibonacci numbers mod m always repeat, and the length of the repeat is the
Pisano period. The primes never do, and say so rather than reporting the scan
limit as though it were a period.`,
		Example: `  pisano period --mod 10
  pisano period --mod 1-12
  pisano period --seq prime --mod 10 --cap 20`,
		Args: cobra.NoArgs,
	}
	seq := addSeqFlags(cmd)
	cmd.Flags().StringVarP(&mod, "mod", "m", "10", "modulus, or a range like 1-40")
	cmd.Flags().IntVarP(&limit, "cap", "c", 0, "term limit for sequences that may not repeat")

	cmd.RunE = func(cc *cobra.Command, _ []string) error {
		stdout := cc.OutOrStdout()
		_ = stdout
		s, err := seq()
		if err != nil {
			return err
		}
		ms, err := parseRange(mod)
		if err != nil {
			return err
		}
		for _, m := range ms {
			p := pisano.Compute(s, m, limit)
			fmt.Fprintln(stdout, p) //nolint:errcheck
			if len(p.Head) > 0 {
				fmt.Fprintf(stdout, "  run-in: %v\n", p.Head) //nolint:errcheck
			}
			fmt.Fprintf(stdout, "  terms:  %v\n", p.Terms) //nolint:errcheck
		}
		return nil
	}
	return cmd
}()
