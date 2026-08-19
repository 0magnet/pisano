// Command axisscan tabulates what a 3D turtle path does, for every modulus.
//
// The question it answers: looking down the path's own axis — the way you would
// look at a screw end-on — what shape do you see, and does the path close?
//
// Both fall out of one fact. The turtle's frame only ever takes orientations
// that map the cubic lattice to itself, so after one pass through the period
// the whole path has been moved by a rigid motion of that lattice: a rotation
// about some axis together with a slide along it. That is a screw motion, and
// the rotations available to it form the chiral octahedral group, order 24,
// whose elements have order 1, 2, 3 or 4. So the figure seen down the axis is
// 1-, 2-, 3- or 4-fold and can be nothing else, and the path closes exactly
// when the slide is zero.
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/0magnet/pisano/pkg/pisano"
)

func main() {
	from := flag.Int("from", 2, "lowest modulus")
	to := flag.Int("to", 200, "highest modulus")
	mul := flag.Int("mul", 1, "multiplier k (the fib*k sequence)")
	seq := flag.String("seq", "fib", "sequence: fib, lucas")
	verbose := flag.Bool("v", false, "one line per modulus")
	flag.Parse()

	pick := func() pisano.Sequence {
		switch *seq {
		case "lucas":
			return pisano.Lucas()
		default:
			if *mul != 1 {
				return pisano.Scaled(*mul)
			}
			return pisano.Fibonacci()
		}
	}

	type key struct {
		order int
		kind  string
	}
	counts := map[key]int{}
	byOrder := map[int][]int{} // order -> moduli, for spotting the pattern
	closedBy := map[int][]int{}

	for m := *from; m <= *to; m++ {
		p := pisano.Compute(pick(), m, 1<<20)
		if !p.Bounded || len(p.Terms) == 0 {
			continue
		}
		s := pisano.Classify3(p.Terms)
		order := s.Turn.Order()
		kind := "open"
		switch {
		case s.Closed:
			kind = "closed"
		case s.Helical:
			kind = "helical"
		}
		counts[key{order, kind}]++
		byOrder[order] = append(byOrder[order], m)
		if s.Closed {
			closedBy[order] = append(closedBy[order], m)
		}
		if *verbose {
			num, den := s.Axis()
			fmt.Printf("m=%-4d order=%d %-8s axis=(%d,%d,%d)/%d period=%d\n",
				m, order, kind, num.X, num.Y, num.Z, den, s.Periods)
		}
	}

	fmt.Printf("\n%s, k=%d, moduli %d..%d\n", *seq, *mul, *from, *to)
	fmt.Println("rotation order seen down the axis  ×  what the path does")
	fmt.Printf("%-6s %-9s %-9s %-9s %s\n", "order", "closed", "helical", "open", "shape end-on")
	shapes := map[int]string{
		1: "no turn (straight/flat)",
		2: "two-lobed",
		3: "triangular / triad",
		4: "square",
	}
	orders := []int{}
	for o := range byOrder {
		orders = append(orders, o)
	}
	sort.Ints(orders)
	for _, o := range orders {
		fmt.Printf("%-6d %-9d %-9d %-9d %s\n", o,
			counts[key{o, "closed"}], counts[key{o, "helical"}], counts[key{o, "open"}],
			shapes[o])
	}
	if len(orders) == 0 {
		fmt.Fprintln(os.Stderr, "nothing computed")
	}
}
