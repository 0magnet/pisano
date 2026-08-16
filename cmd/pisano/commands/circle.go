package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/0magnet/pisano/pkg/pisano"
)

var circleCmd = func() *cobra.Command {
	var (
		mod    string
		out    string
		split  string
		cell   int
		cols   int
		labels bool
		points bool
		rings  bool
		limit  int
		theme  string
	)
	cmd := &cobra.Command{
		Use:   "circle",
		Short: "render circular chord designs",
		Long: `Space m points evenly around a circle, label them with the possible
remainders, and join them in the order the remainders appear.

One modulus gives a single figure; a range gives the contact sheet, which is
where the families become visible — the near-duplicate designs at Fibonacci
moduli 8/21/55 and 13/34/89 are only obvious side by side.`,
		Example: `  pisano circle --mod 1-40 -o sheet.svg
  pisano circle --mod 8,21,55 --cols 3 --cell 260 -o family.svg
  pisano circle --mod 1-24 --mul 3 -o x3.svg
  pisano circle --seq nat --mod 3-9 -o polygons.svg
  pisano circle --mod 1-60 -o page.html`,
		Args: cobra.NoArgs,
	}
	seq := addSeqFlags(cmd)
	cmd.Flags().StringVarP(&mod, "mod", "m", "1-40", "modulus, a range like 1-40, or a list like 8,21,55")
	cmd.Flags().StringVarP(&out, "out", "o", "-", "output .svg or .html file, or - for stdout")
	cmd.Flags().StringVar(&split, "split", "", "write one file per design into this directory")
	cmd.Flags().IntVar(&cell, "cell", 180, "pixels per design")
	cmd.Flags().IntVar(&cols, "cols", 0, "designs per row; 0 for a square sheet")
	cmd.Flags().BoolVar(&labels, "labels", true, "caption each design")
	cmd.Flags().BoolVar(&points, "points", true, "mark the points around each circle")
	cmd.Flags().BoolVar(&rings, "rings", true, "draw the circle itself")
	cmd.Flags().IntVarP(&limit, "cap", "c", 0, "term limit for sequences that may not repeat")
	cmd.Flags().StringVar(&theme, "theme", "auto", "palette: auto, light, dark")

	cmd.RunE = func(_ *cobra.Command, _ []string) error {
		s, err := seq()
		if err != nil {
			return err
		}
		th, err := pisano.ParseTheme(theme)
		if err != nil {
			return err
		}
		ms, err := parseRange(mod)
		if err != nil {
			return err
		}

		opt := pisano.SVGOptions{
			Sheet:  pisano.Sheet{Cell: cell, Cols: cols, Labels: labels, Theme: th},
			Points: points,
			Rings:  rings,
		}
		designs := make([]pisano.Design, 0, len(ms))
		for _, m := range ms {
			designs = append(designs, pisano.Circular(pisano.Compute(s, m, limit)))
		}

		if split != "" {
			one := opt
			one.Cols = 1
			for _, d := range designs {
				name := filepath.Join(split, fmt.Sprintf("%s-mod%d.svg", slug(d.Seq), d.Modulus))
				f, closeFn, err := create(name)
				if err != nil {
					return err
				}
				err = pisano.WriteSVG(f, []pisano.Design{d}, one)
				closeFn()
				if err != nil {
					return err
				}
			}
			fmt.Fprintf(os.Stderr, "wrote %d file(s) to %s\n", len(designs), split)
			return nil
		}

		f, closeFn, err := create(out)
		if err != nil {
			return err
		}
		defer closeFn()

		if isHTML(out) {
			err = pisano.WriteHTML(f, fmt.Sprintf("%s circular designs", designs[0].Seq),
				[]pisano.Section{{Tiles: pisano.CircleTiles(designs, opt), Min: cell}})
		} else {
			err = pisano.WriteSVG(f, designs, opt)
		}
		if err != nil {
			return err
		}
		if out != "-" {
			fmt.Fprintf(os.Stderr, "wrote %d design(s) to %s\n", len(designs), out)
		}
		return nil
	}
	return cmd
}()
