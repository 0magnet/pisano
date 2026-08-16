package commands

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/0magnet/pisano/pkg/pisano"
)

// sheetSpec is one section of the gallery, kept as data so the same figures can
// go to the page and to a standalone file without being built twice.
type sheetSpec struct {
	file    string
	sec     pisano.Section
	designs []pisano.Design
	periods []pisano.Period
}

var galleryCmd = func() *cobra.Command {
	var (
		out    string
		svgDir string
		cell   int
		theme  string
	)
	cmd := &cobra.Command{
		Use:   "gallery",
		Short: "build one HTML page holding every figure",
		Long: `Build a single self-contained page holding every figure, in the order the
constructions are usually introduced, and write the same figures out as
standalone SVG sheets.

The page is for looking at; the files are for using.`,
		Example: `  pisano gallery
  pisano gallery -o site/designs.html --svgdir site/svg --cell 240`,
		Args: cobra.NoArgs,
	}
	cmd.Flags().StringVarP(&out, "out", "o", "out/index.html", "HTML page to write")
	cmd.Flags().StringVar(&svgDir, "svgdir", "out/svg", "directory for standalone SVG sheets; empty to skip")
	cmd.Flags().IntVar(&cell, "cell", 190, "pixels per figure")
	cmd.Flags().StringVar(&theme, "theme", "auto", "palette for the SVG sheets: auto, light, dark")

	cmd.RunE = func(cc *cobra.Command, _ []string) error {
		stdout, stderr := cc.OutOrStdout(), cc.ErrOrStderr()
		h := hostFrom(cc.Context())
		th, err := pisano.ParseTheme(theme)
		if err != nil {
			return err
		}
		return buildGallery(h, stdout, stderr, out, svgDir, cell, th)
	}
	return cmd
}()

func buildGallery(h *Host, stdout, stderr io.Writer, out, svgDir string, cell int, theme pisano.Theme) error {
	// The page keeps the auto palette so it follows the reader's system, while
	// the standalone sheets take whatever was asked for — they get converted to
	// fixed images, and a media query cannot survive that.
	page := pisano.SVGOptions{
		Sheet:  pisano.Sheet{Cell: cell, Labels: true},
		Points: true,
		Rings:  true,
	}
	pageTurt := pisano.TurtleSVGOptions{
		Sheet:   pisano.Sheet{Cell: cell, Labels: true},
		Reps:    3,
		Rounded: true,
	}
	circ, turt := page, pageTurt
	circ.Theme, turt.Theme = theme, theme

	var sheets []sheetSpec

	circles := func(file, title, note string, s pisano.Sequence, mods []int) {
		var ds []pisano.Design
		for _, m := range mods {
			ds = append(ds, pisano.Circular(pisano.Compute(s, m, 0)))
		}
		sheets = append(sheets, sheetSpec{
			file:    file,
			designs: ds,
			sec: pisano.Section{
				Title: title, Note: note,
				Tiles: pisano.CircleTiles(ds, page), Min: cell,
			},
		})
	}
	turtles := func(file, title, note string, ps []pisano.Period) {
		sheets = append(sheets, sheetSpec{
			file:    file,
			periods: ps,
			sec: pisano.Section{
				Title: title, Note: note,
				Tiles: pisano.TurtleTiles(ps, pageTurt), Min: cell,
			},
		})
	}
	modPeriods := func(s pisano.Sequence, mods []int) []pisano.Period {
		var ps []pisano.Period
		for _, m := range mods {
			ps = append(ps, pisano.Compute(s, m, 0))
		}
		return ps
	}
	mixed := func(file, title, note string, ds []pisano.Design) {
		sheets = append(sheets, sheetSpec{
			file:    file,
			designs: ds,
			sec: pisano.Section{
				Title: title, Note: note,
				Tiles: pisano.CircleTiles(ds, page), Min: cell,
			},
		})
	}

	fib, lucas := pisano.Fibonacci(), pisano.Lucas()

	// --- circular designs ----------------------------------------------------

	circles("nat-1-9.svg", "The number line, moduli 1–9",
		"The reference case. Reduced mod m the naturals cycle through every "+
			"remainder in order, so the chords close up into a regular polygon "+
			"every time. Everything interesting below is a departure from this.",
		pisano.Naturals(), seqRange(1, 9))

	circles("fib-1-9.svg", "Fibonacci, moduli 1–9",
		"The same construction on the Fibonacci remainders. Mod 5 happens to use "+
			"every possible chord; mod 6 keeps a mirror axis; mod 8 has none.",
		fib, seqRange(1, 9))

	circles("fib-10.svg", "Fibonacci, modulus 10",
		"A period of 60, so far more chords than its neighbours, and still "+
			"symmetrical.", fib, []int{10})

	circles("fib-family-a.svg", "Fibonacci moduli 8, 21, 55",
		"Near-duplicates of one another. All three moduli are Fibonacci numbers, "+
			"all three periods hold two zeros, none of the three is symmetrical.",
		fib, []int{8, 21, 55})

	circles("fib-family-b.svg", "Fibonacci moduli 13, 34, 89",
		"The other family, alternating with the first as you count down the "+
			"Fibonacci numbers. Four zeros each, symmetrical each — one family "+
			"looks like the other copied and folded onto itself.",
		fib, []int{13, 34, 89})

	circles("lucas-family-a.svg", "Lucas moduli 7, 18, 47",
		"The same splitting into two families happens to the Lucas numbers, at "+
			"moduli that are themselves Lucas numbers.", lucas, []int{7, 18, 47})

	circles("lucas-family-b.svg", "Lucas moduli 11, 29, 76",
		"The second Lucas family.", lucas, []int{11, 29, 76})

	mixed("sequences-10.svg", "Five sequences, all at modulus 10",
		"Fibonacci, Lucas, triangular, the naturals, and the primes. Fibonacci is "+
			"the ingredient, not the recipe. The primes are the odd one out: their "+
			"residues never repeat, so there is no loop to close and the figure is "+
			"an open path — mostly landing on 1, 3, 7 and 9.",
		[]pisano.Design{
			pisano.Circular(pisano.Compute(fib, 10, 0)),
			pisano.Circular(pisano.Compute(lucas, 10, 0)),
			pisano.Circular(pisano.Compute(pisano.Triangular(), 10, 0)),
			pisano.Circular(pisano.Compute(pisano.Naturals(), 10, 0)),
			pisano.Circular(pisano.Compute(pisano.Primes(), 10, 0)),
		})

	for _, k := range []int{2, 3, 4} {
		circles(fmt.Sprintf("fib-x%d-1-24.svg", k),
			fmt.Sprintf("Fibonacci × %d, moduli 1–24", k),
			fmt.Sprintf("Multiplying the sequence through by %d makes the plain "+
				"designs reappear at every %dth modulus, with new ones interleaved "+
				"between them. That is the \"design m/%d\" labelling — m is still an "+
				"integer modulus, and there is no such thing as mod m/%d.", k, k, k, k),
			pisano.Scaled(k), seqRange(1, 24))
	}

	circles("fib-1-40.svg", "Fibonacci, moduli 1–40",
		"The contact sheet. Longer periods give busier figures, but length alone "+
			"does not decide whether one comes out symmetrical — the zero count does.",
		fib, seqRange(1, 40))

	circles("fib-41-80.svg", "Fibonacci, moduli 41–80",
		"Further out, where the near-duplicate families keep recurring.",
		fib, seqRange(41, 80))

	// --- turtle paths --------------------------------------------------------

	unreduced := func(s pisano.Sequence) pisano.Period {
		p, err := pisano.UnreducedPeriod(s)
		if err != nil {
			panic(err) // every sequence used here supports it
		}
		return p
	}

	turtles("turtle-unreduced.svg", "Turtle paths with no modulus at all",
		"Odd turns left and steps, even turns right and steps, zero does neither. "+
			"The naturals alternate odd and even and make a staircase. Fibonacci "+
			"runs odd, odd, even, so it turns left, left, right — and closes into "+
			"something that looks like a plus sign. The triangular numbers cancel "+
			"two lefts against two rights and drift instead.",
		[]pisano.Period{
			unreduced(pisano.Naturals()),
			unreduced(fib),
			unreduced(pisano.Triangular()),
			unreduced(lucas),
		})

	turtles("turtle-fib-1-40.svg", "Fibonacci turtle paths, moduli 1–40",
		"Now the remainders drive the turns, so a residue of zero also stops it. "+
			"Some paths close on themselves; others repeat a motif off to infinity, "+
			"drawn here for three passes. Which one happens is decided by the net "+
			"turn over a single pass, and nothing else.",
		modPeriods(fib, seqRange(1, 40)))

	turtles("turtle-fib-41-80.svg", "Fibonacci turtle paths, moduli 41–80",
		"The same, further out.", modPeriods(fib, seqRange(41, 80)))

	turtles("turtle-twins.svg", "Turtle paths shared between moduli",
		"These moduli trace the same figure as each other up to where it sits and "+
			"which way it faces: 6, 12, 16, 18, 24, 36 and 48 all draw one shape, "+
			"and 14, 28, 32, 42 and 46 another.",
		modPeriods(fib, []int{6, 12, 16, 18, 24, 36, 48, 14, 28, 32, 42, 46}))

	turtles("turtle-lucas-1-24.svg", "Lucas turtle paths, moduli 1–24",
		"The same walk over a different sequence.", modPeriods(lucas, seqRange(1, 24)))

	turtles("turtle-tri-1-24.svg", "Triangular turtle paths, moduli 1–24",
		"And over the triangular numbers, whose rule for making terms has nothing "+
			"to do with the other two.", modPeriods(pisano.Triangular(), seqRange(1, 24)))

	// --- write it all out ----------------------------------------------------

	secs := make([]pisano.Section, 0, len(sheets))
	for _, sh := range sheets {
		secs = append(secs, sh.sec)
	}

	f, closeFn, err := create(h, stdout, out)
	if err != nil {
		return err
	}
	err = pisano.WriteHTMLPage(f, "Pisano periods, drawn",
		`Every figure here is generated by <code>pisano gallery</code>. `+
			`For the interactive version — the same designs drawn live in a `+
			`shell compiled to WebAssembly — see <a href="term/">the terminal</a>.`,
		secs)
	closeFn()
	if err != nil {
		return err
	}
	fmt.Fprintf(stderr, "wrote %s (%d sections)\n", out, len(secs))

	if svgDir == "" {
		return nil
	}
	for _, sh := range sheets {
		g, closeG, err := create(h, stdout, filepath.Join(svgDir, sh.file))
		if err != nil {
			return err
		}
		if sh.designs != nil {
			err = pisano.WriteSVG(g, sh.designs, circ)
		} else {
			err = pisano.WriteTurtleSVG(g, sh.periods, turt)
		}
		closeG()
		if err != nil {
			return err
		}
	}
	fmt.Fprintf(stderr, "wrote %d SVG sheet(s) to %s\n", len(sheets), svgDir)
	return nil
}
