package commands

import (
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/0magnet/pisano/pkg/pisano"
)

// addSeqFlags registers the sequence selection every command shares and returns
// the resolver. The flag values live in the closure rather than in package
// state so two commands can never quietly read each other's.
func addSeqFlags(cmd *cobra.Command) func() (pisano.Sequence, error) {
	var (
		name string
		k    int
	)
	cmd.Flags().StringVarP(&name, "seq", "s", "fib", "sequence: fib, lucas, tri, nat, prime")
	cmd.Flags().IntVarP(&k, "mul", "k", 1, "multiply the Fibonacci sequence by this")

	return func() (pisano.Sequence, error) {
		switch name {
		case "fib":
			if k != 1 {
				if k < 1 {
					return nil, fmt.Errorf("--mul must be positive")
				}
				return pisano.Scaled(k), nil
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
		default:
			return nil, fmt.Errorf("unknown sequence %q", name)
		}
	}
}

// parseRange accepts "10", "1-40", "1..40", or a comma-separated mixture of
// those. Naming moduli individually is how you put a family side by side: the
// designs at Fibonacci moduli 8,21,55 are near-duplicates of each other, and
// nothing about a contiguous range would show you that.
func parseRange(s string) ([]int, error) {
	if strings.Contains(s, ",") {
		var out []int
		for _, part := range strings.Split(s, ",") {
			if strings.TrimSpace(part) == "" {
				continue
			}
			ms, err := parseRange(part)
			if err != nil {
				return nil, err
			}
			out = append(out, ms...)
		}
		if len(out) == 0 {
			return nil, fmt.Errorf("no moduli in %q", s)
		}
		return out, nil
	}

	s = strings.TrimSpace(s)
	sep := ""
	switch {
	case strings.Contains(s, ".."):
		sep = ".."
	case strings.Contains(s, "-"):
		sep = "-"
	}
	if sep == "" {
		m, err := strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("bad modulus %q", s)
		}
		if m < 1 {
			return nil, fmt.Errorf("modulus must be at least 1")
		}
		return []int{m}, nil
	}

	parts := strings.SplitN(s, sep, 2)
	lo, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	hi, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil {
		return nil, fmt.Errorf("bad modulus range %q", s)
	}
	if lo < 1 || hi < lo {
		return nil, fmt.Errorf("bad modulus range %d-%d", lo, hi)
	}
	out := make([]int, 0, hi-lo+1)
	for m := lo; m <= hi; m++ {
		out = append(out, m)
	}
	return out, nil
}

// create opens an output path, or the command's own stdout for "-".
func create(h *Host, out io.Writer, path string) (io.Writer, func(), error) {
	if path == "-" || path == "" {
		return out, func() {}, nil
	}
	full := h.resolve(path)
	if dir := filepath.Dir(full); dir != "." && dir != "/" {
		if err := h.Files.MkdirAll(dir, 0o755); err != nil {
			return nil, nil, err
		}
	}
	f, err := h.Files.Create(full)
	if err != nil {
		return nil, nil, err
	}
	return f, func() { _ = f.Close() }, nil //nolint:errcheck,gosec // the callers cleanup; the file was only written to
}

func isHTML(path string) bool { return strings.EqualFold(filepath.Ext(path), ".html") }

func slug(s string) string {
	return strings.NewReplacer(" ", "-", "*", "x", "(", "", ")", "").Replace(s)
}

func seqRange(lo, hi int) []int {
	out := make([]int, 0, hi-lo+1)
	for m := lo; m <= hi; m++ {
		out = append(out, m)
	}
	return out
}
