package commands

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/0magnet/pisano/pkg/border"
)

var borderCmd = func() *cobra.Command {
	var (
		across, down, corner    int
		acrossP, downP, cornerP int
		cols, rows              int
		out                     string
		fitW, fitH              float64
		font                    float64
		list                    string
		maxMod                  int
		limit                   int
	)
	cmd := &cobra.Command{
		Use:   "border",
		Short: "build a border out of turtle figures",
		Long: `Compose a border from three figures: one running across the top and bottom,
one running down the sides, and one closed figure at the corners.

A figure only makes a run if its path TRAVELS — ends somewhere other than where
it began — so that copies laid one travel apart advance along the edge instead
of piling up. Most traveling figures go diagonally, and a diagonal run is laid
flat by turning the whole border; that turn cannot be done in a grid, so an
angled border is written as HTML and an upright one can also be printed here.

The corners are closed figures on purpose. A closed path has no loose ends, so a
run can stop against it without leaving a stub hanging in the air.

The gap between a corner and the first piece of a run is one travel vector, so
a corner much smaller than the run's travel leaves air between them. Pair a
big-traveling run with a corner of comparable size, or raise --cols/--rows so
the travel is a smaller share of the edge.

--list prints the catalog instead of building anything: every distinct figure,
folded together under rotation and reflection, since otherwise the even moduli
alone repeat one drawing hundreds of times.`,
		Example: `  pisano border --list travel
  pisano border --list closed
  pisano border --across 9 --down 23 --corner 37 --cols 9 --rows 3 -o border.html
  pisano border --across 33 --down 43 --corner 5 --fit 900x400 -o border.html`,
		Args: cobra.NoArgs,
	}
	cmd.Flags().IntVar(&across, "across", 9, "modulus for the top and bottom runs")
	cmd.Flags().IntVar(&down, "down", 0, "modulus for the side runs; 0 picks one square to --across")
	cmd.Flags().IntVar(&acrossP, "across-passes", 2, "times round the period for the across figure; the figure changes with it")
	cmd.Flags().IntVar(&downP, "down-passes", 2, "times round the period for the down figure")
	cmd.Flags().IntVar(&cornerP, "corner-passes", 2, "times round the period for the corner figure")
	cmd.Flags().IntVar(&corner, "corner", 0, "modulus for the corners; 0 picks a closed figure big enough for the runs")
	cmd.Flags().IntVar(&cols, "cols", 8, "copies of the across figure between the corners")
	cmd.Flags().IntVar(&rows, "rows", 3, "copies of the down figure between the corners")
	cmd.Flags().StringVarP(&out, "out", "o", "", "write HTML here (default: stdout, as text when the border is upright)")
	cmd.Flags().Float64Var(&font, "font", 0, "font size in px; 0 fits the border to --fit")
	cmd.Flags().StringVar(&list, "list", "", "print the catalog instead: all, closed, travel, diagonal")
	cmd.Flags().IntVar(&maxMod, "max-mod", 3000, "highest modulus to consider when listing")
	cmd.Flags().IntVar(&limit, "limit", 60, "how many figures --list prints")
	var cornerScale float64
	cmd.Flags().Float64Var(&cornerScale, "corner-scale", 3,
		"how many times the widest run the corner must be; 10 is an order of magnitude, but only small runs leave room for it")
	var fit string
	cmd.Flags().StringVar(&fit, "fit", "900x400", "box the border is scaled to fit, WxH in px")

	cmd.RunE = func(cc *cobra.Command, _ []string) error {
		if _, err := fmt.Sscanf(fit, "%gx%g", &fitW, &fitH); err != nil {
			return fmt.Errorf("--fit wants WxH in pixels, e.g. 900x400: %w", err)
		}
		if list != "" {
			return printCatalog(cc.OutOrStdout(), list, maxMod, limit)
		}
		pick := func(m, passes int, what string) (border.Figure, error) {
			f, ok := border.OfPasses(m, passes)
			if !ok {
				return f, fmt.Errorf("no %s figure for modulus %d at %d passes", what, m, passes)
			}
			return f, nil
		}
		a, err := pick(across, acrossP, "across")
		if err != nil {
			return err
		}
		var d border.Figure
		if down == 0 {
			// Pick a partner whose travel is a quarter turn from the across
			// figure.s. Getting this wrong is the usual way to end up with a
			// parallelogram, and the drawings give no hint of it.
			cands := border.SquarePartners(a, border.Catalog(3, maxMod), 1)
			if len(cands) == 0 {
				return fmt.Errorf("no figure below modulus %d runs square to mod %d; try another --across", maxMod, across)
			}
			d = cands[0]
			fmt.Fprintf(cc.ErrOrStderr(), //nolint:errcheck // a note, not output
				"picked mod %d for the sides: %dx%d, travel %+d,%+d, square to mod %d\n",
				d.Mod, d.W(), d.H(), d.DX, d.DY, across)
		} else {
			var err error
			d, err = pick(down, downP, "down")
			if err != nil {
				return err
			}
		}
		var c border.Figure
		if corner == 0 {
			// A corner must be at least as big as the widest run, or the runs
			// swallow it and it stops reading as the place they end.
			var ok bool
			c, ok = border.CornerFor(a, d, border.Catalog(3, maxMod), cornerScale)
			if !ok {
				return fmt.Errorf("no closed figure below modulus %d is big enough to corner these runs; raise --max-mod", maxMod)
			}
			fmt.Fprintf(cc.ErrOrStderr(), //nolint:errcheck // a note, not output
				"picked mod %d for the corners: %dx%d, %d points\n", c.Mod, c.W(), c.H(), c.Points)
		} else {
			var err error
			c, err = pick(corner, cornerP, "corner")
			if err != nil {
				return err
			}
			if !c.Closed {
				fmt.Fprintf(cc.ErrOrStderr(), //nolint:errcheck // a note, not output
					"note: mod %d is an open figure, so the corners will have loose ends\n", corner)
			}
			if big := math.Max(a.Extent(), d.Extent()); c.Extent() < big {
				fmt.Fprintf(cc.ErrOrStderr(), //nolint:errcheck // a note, not output
					"note: the corner (extent %.0f) is smaller than the widest run (%.0f), so the runs will swallow it\n",
					c.Extent(), big)
			}
		}
		l, err := border.Compose(border.Spec{Across: a, Down: d, Corner: c, Cols: cols, Rows: rows})
		if err != nil {
			return err
		}
		if !l.Square(1) {
			fmt.Fprintf(cc.ErrOrStderr(), //nolint:errcheck // a note, not output //nolint:errcheck // a note, not output
				"note: the down run lands at %.1f° rather than 90°, so the two runs are not square to each other\n",
				l.DownAngle)
		}
		// An upright border is a grid, and a grid can be printed. An angled one
		// only exists once something can turn it, which here is CSS.
		if out == "" && l.Upright(1) {
			fmt.Fprintln(cc.OutOrStdout(), l.Grid().String()) //nolint:errcheck // stdout
			return nil
		}
		frag := l.HTML(border.HTMLOptions{
			FontPx: font, FitW: fitW, FitH: fitH,
			Color: "#3d8fb8",
		})
		page := border.Page(fmt.Sprintf("pisano border %d/%d/%d", across, down, corner), frag)
		if out == "" {
			fmt.Fprint(cc.OutOrStdout(), page) //nolint:errcheck // stdout
			return nil
		}
		return os.WriteFile(out, []byte(page), 0o600)
	}
	return cmd
}()

// printCatalog lists figures, filtered by what they are good for.
func printCatalog(w interface{ Write([]byte) (int, error) }, kind string, maxMod, limit int) error {
	all := border.Catalog(3, maxMod)
	var keep []border.Figure
	for _, f := range all {
		switch kind {
		case "all":
		case "closed":
			if !f.Closed {
				continue
			}
		case "travel":
			if !f.Travels() {
				continue
			}
		case "diagonal":
			if !f.Diagonal(0.25) {
				continue
			}
		default:
			return fmt.Errorf("--list wants one of: all, closed, travel, diagonal")
		}
		keep = append(keep, f)
	}
	sort.Slice(keep, func(i, j int) bool { return keep[i].Points < keep[j].Points })
	var b strings.Builder
	fmt.Fprintf(&b, "%d %s figures among moduli 3..%d (distinct up to rotation and reflection)\n\n",
		len(keep), kind, maxMod)
	fmt.Fprintf(&b, "%-7s %-9s %-7s %-9s %s\n", "mod", "size", "points", "travel", "also at")
	for i, f := range keep {
		if i >= limit {
			fmt.Fprintf(&b, "... and %d more\n", len(keep)-i)
			break
		}
		also := ""
		if n := len(f.Also); n > 0 {
			also = fmt.Sprintf("%d more moduli", n)
		}
		fmt.Fprintf(&b, "%-7d %-9s %-7d %-9s %s\n", f.Mod,
			fmt.Sprintf("%dx%d", f.W(), f.H()), f.Points,
			fmt.Sprintf("%+d,%+d", f.DX, f.DY), also)
	}
	_, err := w.Write([]byte(b.String()))
	return err
}
