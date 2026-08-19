package commands

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/0magnet/pisano/pkg/pisano"
)

var turtleCmd = func() *cobra.Command {
	var (
		mod    string
		out    string
		split  string
		reps   int
		cell   int
		cols   int
		labels bool
		bypass bool
		grid   bool
		plain  bool
		tint   string
		limit  int
		theme  string
	)
	cmd := &cobra.Command{
		Use:   "turtle",
		Short: "render left/right turtle paths",
		Long: `Read each term as an instruction: odd turns left and steps forward, even
turns right and steps forward, zero does neither.

With no --out the path is drawn in the terminal with box-drawing characters.
Give --out a .svg or .html name for vector output.

--mod 0 applies no modulus at all. That is not the same as --mod 2: mod 2 sends
every even term to the residue 0, which the turtle reads as "stay put", while
unreduced only a term that is genuinely zero stays. Unreduced Fibonacci is the
figure that looks like a plus sign.`,
		Example: `  pisano turtle --mod 10
  pisano turtle --mod 0
  pisano turtle --mod 1-40 -o paths.html
  pisano turtle --mod 25 --reps 6 --bypass --grid -o drift.svg
  pisano turtle --mod 8,21,55 --split web/ --labels=false`,
		Args: cobra.NoArgs,
	}
	seq := addSeqFlags(cmd)
	cmd.Flags().StringVarP(&mod, "mod", "m", "10", "modulus, a range like 1-20, or 0 for no modulus")
	cmd.Flags().StringVarP(&out, "out", "o", "", "output .svg or .html file; default draws in the terminal")
	cmd.Flags().StringVar(&split, "split", "", "write one file per design into this directory")
	cmd.Flags().IntVarP(&reps, "reps", "r", 3, "passes to draw for paths that never close")
	cmd.Flags().IntVar(&cell, "cell", 180, "pixels per design when writing a file")
	cmd.Flags().IntVar(&cols, "cols", 0, "designs per row; 0 for a square sheet")
	cmd.Flags().BoolVar(&labels, "labels", true, "caption each design")
	cmd.Flags().BoolVar(&bypass, "bypass", false, "give each pass its own color")
	cmd.Flags().BoolVar(&grid, "grid", false, "faint lattice behind each path")
	cmd.Flags().BoolVar(&plain, "plain", false, "no ANSI color in the terminal")
	cmd.Flags().StringVar(&tint, "tint", "step", "what a color means: "+pisano.TintNames())
	cmd.Flags().IntVarP(&limit, "cap", "c", 0, "term limit for sequences that may not repeat")
	cmd.Flags().StringVar(&theme, "theme", "auto", "palette: auto, light, dark")

	cmd.RunE = func(cc *cobra.Command, _ []string) error {
		stdout, stderr := cc.OutOrStdout(), cc.ErrOrStderr()
		h := hostFrom(cc.Context())
		s, err := seq()
		if err != nil {
			return err
		}
		th, err := pisano.ParseTheme(theme)
		if err != nil {
			return err
		}
		tintMode, err := pisano.ParseTint(tint)
		if err != nil {
			return err
		}
		periods, err := turtlePeriods(s, mod, limit)
		if err != nil {
			return err
		}

		if out == "" && split == "" {
			for i, p := range periods {
				if i > 0 {
					fmt.Fprintln(stdout) //nolint:errcheck
				}
				art, shape := pisano.RenderTurtle(p, pisano.TurtleOptions{
					Reps: reps, Colorize: !plain, Tint: tintMode,
				})
				what := p.Seq
				if p.Modulus > 0 {
					what = fmt.Sprintf("%s mod %d", p.Seq, p.Modulus)
				}
				fmt.Fprintf(stdout, "%s — cycle %d, %s\n\n", what, p.Len(), shape) //nolint:errcheck
				fmt.Fprint(stdout, art)                                            //nolint:errcheck
			}
			return nil
		}

		opt := pisano.TurtleSVGOptions{
			Sheet:   pisano.Sheet{Cell: cell, Cols: cols, Labels: labels, Theme: th},
			Reps:    reps,
			ByPass:  bypass,
			Tint:    tintMode,
			Grid:    grid,
			Rounded: true,
		}

		if split != "" {
			one := opt
			one.Cols = 1
			for _, p := range periods {
				name := filepath.Join(split,
					fmt.Sprintf("%s-turtle-mod%d.svg", slug(p.Seq), p.Modulus))
				f, closeFn, err := create(h, stdout, name)
				if err != nil {
					return err
				}
				err = pisano.WriteTurtleSVG(f, []pisano.Period{p}, one)
				closeFn()
				if err != nil {
					return err
				}
			}
			fmt.Fprintf(stderr, "wrote %d file(s) to %s\n", len(periods), split) //nolint:errcheck
			if out == "" {
				return nil
			}
		}

		f, closeFn, err := create(h, stdout, out)
		if err != nil {
			return err
		}
		defer closeFn()

		if isHTML(out) {
			err = pisano.WriteHTML(f, fmt.Sprintf("%s turtle paths", periods[0].Seq),
				[]pisano.Section{{Tiles: pisano.TurtleTiles(periods, opt), Min: cell}})
		} else {
			err = pisano.WriteTurtleSVG(f, periods, opt)
		}
		if err != nil {
			return err
		}
		if out != "-" {
			fmt.Fprintf(stderr, "wrote %d design(s) to %s\n", len(periods), out) //nolint:errcheck
		}
		return nil
	}
	return cmd
}()

// turtlePeriods resolves the --mod flag, where 0 means no reduction at all: the
// turtle reads the terms themselves, and only a term that is genuinely zero
// stays put. It is the reference figure the whole construction starts from.
func turtlePeriods(s pisano.Sequence, mod string, limit int) ([]pisano.Period, error) {
	if strings.TrimSpace(mod) == "0" {
		p, err := pisano.UnreducedPeriod(s)
		if err != nil {
			return nil, err
		}
		return []pisano.Period{p}, nil
	}
	ms, err := parseRange(mod)
	if err != nil {
		return nil, err
	}
	ps := make([]pisano.Period, 0, len(ms))
	for _, m := range ms {
		ps = append(ps, pisano.Compute(s, m, limit))
	}
	return ps, nil
}
