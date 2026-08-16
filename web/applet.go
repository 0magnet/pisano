//go:build js && wasm

package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/0magnet/sh/v3/interp"
	"github.com/0magnet/websh/shell"

	"github.com/0magnet/pisano/pkg/pisano"
	"github.com/0magnet/pisano/pkg/tui"
)

// registerPisano adds the pisano command to the shell.
//
// The subcommands are parsed here rather than through the cobra tree the native
// binary uses, for one reason: writing a file. Natively -o means os.Create; here
// it has to mean the shell's virtual filesystem, so that a design generated in
// the browser can be handed to websh's own `download` applet. Everything the
// commands actually do is the same package either way.
func registerPisano() {
	shell.RegisterApplet("pisano", "Pisano period designs (try: pisano tui)",
		func(ctx context.Context, s *shell.Shell, hc *interp.HandlerContext, args []string) int {
			return runPisano(ctx, s, hc, args)
		})
}

const pisanoUsage = `pisano — designs from reducing an integer sequence modulo m

  pisano tui [--mod N] [--seq S]      watch a design draw itself, full screen
  pisano period --mod N               the repeating block
  pisano turtle --mod N [--reps R]    the left/right walk, as box drawing
  pisano circle --mod RANGE -o FILE   write an SVG to the filesystem
  pisano sweep [--max N]              does the zero count predict symmetry?

  --seq  fib (default), lucas, tri, nat, prime
  --mod  a number, a range like 1-40, or a list like 8,21,55

Files written by "circle" land in the virtual filesystem; use "download" to
pull one out to your machine.
`

func runPisano(ctx context.Context, s *shell.Shell, hc *interp.HandlerContext, args []string) int {
	// websh hands an applet its arguments with the command name already
	// stripped, so args[0] is the subcommand.
	if len(args) == 0 {
		fmt.Fprint(hc.Stdout, pisanoUsage)
		return 0
	}
	sub, rest := args[0], args[1:]

	fs := flag.NewFlagSet("pisano "+sub, flag.ContinueOnError)
	fs.SetOutput(hc.Stderr)
	var (
		seqName = fs.String("seq", "fib", "sequence: fib, lucas, tri, nat, prime")
		mul     = fs.Int("mul", 1, "multiply the Fibonacci sequence by this")
		mod     = fs.String("mod", "", "modulus, range or list")
		out     = fs.String("o", "", "output file in the virtual filesystem")
		reps    = fs.Int("reps", 3, "passes to draw for paths that never close")
		max     = fs.Int("max", 300, "highest modulus for sweep")
		cell    = fs.Int("cell", 180, "pixels per design")
		cols    = fs.Int("cols", 0, "designs per row")
		speed   = fs.Int("speed", 4, "tui: steps drawn per frame")
		cycle   = fs.Duration("cycle", 0, "tui: step to the next modulus this often")
		circle  = fs.Bool("circle", false, "tui: start on the circular design")
	)
	if err := fs.Parse(rest); err != nil {
		return 2
	}

	seq, err := buildSeq(*seqName, *mul)
	if err != nil {
		fmt.Fprintln(hc.Stderr, "pisano:", err)
		return 1
	}

	modOr := func(def string) string {
		if *mod == "" {
			return def
		}
		return *mod
	}

	switch sub {
	case "help", "-h", "--help":
		fmt.Fprint(hc.Stdout, pisanoUsage)
		return 0

	case "tui", "watch":
		// 25 rather than 10: an open path, which is the one worth watching.
		return runViewer(ctx, s, hc, tui.Options{
			Seq: *seqName, Mul: *mul, Mod: atoiOr(*mod, 25),
			Speed: *speed, Cycle: *cycle, Circle: *circle,
		})

	case "period":
		ms, err := parseRange(modOr("10"))
		if err != nil {
			fmt.Fprintln(hc.Stderr, "pisano:", err)
			return 1
		}
		for _, n := range ms {
			p := pisano.Compute(seq, n, 0)
			fmt.Fprintln(hc.Stdout, p)
			if len(p.Head) > 0 {
				fmt.Fprintf(hc.Stdout, "  run-in: %v\n", p.Head)
			}
			fmt.Fprintf(hc.Stdout, "  terms:  %v\n", p.Terms)
		}
		return 0

	case "turtle":
		ms, err := parseRange(modOr("25"))
		if err != nil {
			fmt.Fprintln(hc.Stderr, "pisano:", err)
			return 1
		}
		for i, n := range ms {
			if i > 0 {
				fmt.Fprintln(hc.Stdout)
			}
			p := pisano.Compute(seq, n, 0)
			art, shape := pisano.RenderTurtle(p, pisano.TurtleOptions{Reps: *reps, Colorize: true})
			fmt.Fprintf(hc.Stdout, "%s mod %d — cycle %d, %s\n\n", p.Seq, n, p.Len(), shape)
			fmt.Fprint(hc.Stdout, art)
		}
		return 0

	case "circle":
		ms, err := parseRange(modOr("1-40"))
		if err != nil {
			fmt.Fprintln(hc.Stderr, "pisano:", err)
			return 1
		}
		designs := make([]pisano.Design, 0, len(ms))
		for _, n := range ms {
			designs = append(designs, pisano.Circular(pisano.Compute(seq, n, 0)))
		}
		opt := pisano.SVGOptions{
			Sheet:  pisano.Sheet{Cell: *cell, Cols: *cols, Labels: true},
			Points: true, Rings: true,
		}
		if *out == "" {
			return svgTo(hc.Stdout, designs, opt, hc)
		}
		f, err := s.FS.Create(absPath(s, *out))
		if err != nil {
			fmt.Fprintln(hc.Stderr, "pisano:", err)
			return 1
		}
		code := svgTo(f, designs, opt, hc)
		if cerr := f.Close(); cerr != nil {
			fmt.Fprintln(hc.Stderr, "pisano:", cerr)
			return 1
		}
		if code == 0 {
			fmt.Fprintf(hc.Stderr, "wrote %d design(s) to %s\n", len(designs), *out)
		}
		return code

	case "sweep":
		rows := pisano.Sweep(seq, pisano.SweepOptions{Max: *max})
		fmt.Fprintln(hc.Stdout, "zeros vs symmetry of the circular design")
		fmt.Fprintln(hc.Stdout, "----------------------------------------")
		fmt.Fprint(hc.Stdout, pisano.ZeroSymmetry(rows))
		fmt.Fprintln(hc.Stdout)
		fmt.Fprintln(hc.Stdout, "turtle path shape")
		fmt.Fprintln(hc.Stdout, "-----------------")
		fmt.Fprint(hc.Stdout, pisano.TurtleShapes(rows))
		return 0
	}

	fmt.Fprintf(hc.Stderr, "pisano: unknown command %q\n\n%s", sub, pisanoUsage)
	return 2
}

func svgTo(w io.Writer, d []pisano.Design, opt pisano.SVGOptions, hc *interp.HandlerContext) int {
	if err := pisano.WriteSVG(w, d, opt); err != nil {
		fmt.Fprintln(hc.Stderr, "pisano:", err)
		return 1
	}
	return 0
}

func absPath(s *shell.Shell, p string) string {
	if strings.HasPrefix(p, "/") {
		return p
	}
	return s.Dir() + "/" + p
}

// --- the viewer -------------------------------------------------------------

// runViewer drives the same tui.Model the desktop binary does. The difference
// is only in what feeds it: there, Bubble Tea; here, the shell's raw input and
// a ticker. Both make the same four calls, so the figure is the same figure.
func runViewer(ctx context.Context, s *shell.Shell, hc *interp.HandlerContext, opt tui.Options) int {
	cols, rows := 80, 24
	if s.Size != nil {
		cols, rows = s.Size()
	}

	m := tui.New(opt)
	m.Resize(cols, rows)

	if s.RawMode != nil {
		s.RawMode(true)
		defer s.RawMode(false)
	}
	out := hc.Stdout
	fmt.Fprint(out, "\x1b[?1049h\x1b[?25l") // alt screen, hide cursor
	defer fmt.Fprint(out, "\x1b[?25h\x1b[?1049l")

	// Keys are read on their own goroutine so the frame clock keeps running
	// while nothing is being typed. A blocking read here would stall the
	// animation between keystrokes.
	keys := make(chan string, 16)
	done := make(chan struct{})
	go readKeys(bufio.NewReader(hc.Stdin), keys, done)
	defer close(done)

	draw := func() {
		// The model returns a plain frame; the terminal wants the cursor
		// home and CRLF line ends.
		frame := strings.ReplaceAll(m.Frame(), "\n", "\r\n")
		fmt.Fprint(out, "\x1b[H\x1b[2J"+frame)
	}
	draw()

	ticker := time.NewTicker(tui.FrameInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return 130
		case k := <-keys:
			if !m.Key(k) {
				return 0
			}
			// The window can be resized while the viewer is up, and the
			// applet is never told; re-reading on input is cheap and picks
			// it up the moment anything is pressed.
			if s.Size != nil {
				if c, r := s.Size(); c != cols || r != rows {
					cols, rows = c, r
					m.Resize(c, r)
				}
			}
			draw()
		case <-ticker.C:
			m.Advance()
			draw()
		}
	}
}

// readKeys feeds the model's key names onto a channel. The parsing itself is
// tui.ReadKey, shared with nothing else today but tested natively — a driver
// that only compiles for js/wasm cannot be.
func readKeys(r *bufio.Reader, keys chan<- string, done <-chan struct{}) {
	for {
		name, err := tui.ReadKey(r)
		if err != nil {
			return
		}
		if name == "" {
			continue
		}
		select {
		case keys <- name:
		case <-done:
			return
		}
	}
}

// --- shared parsing ----------------------------------------------------------

func buildSeq(name string, mul int) (pisano.Sequence, error) {
	switch name {
	case "fib":
		if mul > 1 {
			return pisano.Scaled(mul), nil
		}
		return pisano.Fibonacci(), nil
	case "lucas":
		return pisano.Lucas(), nil
	case "tri":
		return pisano.Triangular(), nil
	case "nat":
		return pisano.Naturals(), nil
	case "prime":
		return pisano.Primes(), nil
	}
	return nil, fmt.Errorf("unknown sequence %q", name)
}

func atoiOr(s string, def int) int {
	if n, err := strconv.Atoi(strings.TrimSpace(s)); err == nil && n > 0 {
		return n
	}
	return def
}

// parseRange accepts "10", "1-40" or a comma-separated mixture.
func parseRange(s string) ([]int, error) {
	if strings.Contains(s, ",") {
		var all []int
		for _, part := range strings.Split(s, ",") {
			if strings.TrimSpace(part) == "" {
				continue
			}
			ms, err := parseRange(part)
			if err != nil {
				return nil, err
			}
			all = append(all, ms...)
		}
		return all, nil
	}
	s = strings.TrimSpace(s)
	if lo, hi, ok := strings.Cut(s, "-"); ok {
		a, err1 := strconv.Atoi(strings.TrimSpace(lo))
		b, err2 := strconv.Atoi(strings.TrimSpace(hi))
		if err1 != nil || err2 != nil || a < 1 || b < a {
			return nil, fmt.Errorf("bad modulus range %q", s)
		}
		out := make([]int, 0, b-a+1)
		for m := a; m <= b; m++ {
			out = append(out, m)
		}
		return out, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 {
		return nil, fmt.Errorf("bad modulus %q", s)
	}
	return []int{n}, nil
}
