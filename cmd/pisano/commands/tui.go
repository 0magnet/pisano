package commands

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/0magnet/pisano/pkg/tui"
)

var tuiCmd = func() *cobra.Command {
	var (
		seq    string
		mul    int
		mod    int
		noMod  bool
		speed  int
		maxPts int
		mono   bool
		render string
		cam    string
		trail  string
		circle bool
		paused bool
		cycle  time.Duration
	)
	cmd := &cobra.Command{
		Use:     "tui",
		Aliases: []string{"watch"},
		Short:   "watch the designs draw themselves",
		Long: `Open the designs in the alternate screen and watch them being drawn.

The walk never finishes. An open path drifts forever, so the camera follows its
head and the oldest of the drawing is dropped as it scrolls off — the figure
stays at its true size rather than shrinking to fit. A closed path is simply
walked again from where it left off, retracing exactly the same figure in the
next colour, so the loop sweeps round instead of sitting there done.

--cam scroll leaves the view where it is until the head reaches the edge, then
moves by the least it can to keep it in frame; --cam page does the same but
shoves the view most of a screen at a time, so it is still for longer and then
jumps once. --cam follow pins the head to the centre instead, which is exact but
means the whole drawing slides on every frame.

--trail sets how much stays on screen: a whole circuit, or a shorter tail that
follows the head like a comet.

Press o to jump to the next modulus whose path never closes, and space to hold
it still once it has drawn something worth looking at.

--cycle steps through the moduli on its own. Prefer it to a shell loop around
the whole program: a full-screen program puts the terminal in raw mode, where
Ctrl-C is a keystroke rather than a signal, so a loop of them is caught one
instance at a time. This command refuses to start without a terminal for the
same reason.`,
		Example: `  pisano tui
  pisano tui --mod 25
  pisano tui --seq lucas --mod 47 --speed 8
  pisano tui --circle --mod 10
  pisano tui --trail comet --mod 25
  pisano tui --cycle 5s --mod 1
  pisano tui --cam page --mod 25
  pisano tui --render braille --cam fit --mod 25`,
		Args: cobra.NoArgs,
	}
	cmd.Flags().StringVarP(&seq, "seq", "s", "fib", "sequence: fib, lucas, tri, nat, prime")
	cmd.Flags().IntVarP(&mul, "mul", "k", 1, "multiply the Fibonacci sequence by this")
	cmd.Flags().IntVarP(&mod, "mod", "m", 25, "modulus to start on")
	cmd.Flags().BoolVar(&noMod, "no-mod", false, "start with no modulus at all")
	cmd.Flags().IntVarP(&speed, "speed", "r", 4, "path steps drawn per frame")
	cmd.Flags().IntVar(&maxPts, "max-points", 60000, "stop extending an open path past this many points")
	cmd.Flags().BoolVar(&mono, "mono", false, "no colour")
	cmd.Flags().StringVar(&render, "render", "box", "renderer: box, braille")
	cmd.Flags().StringVar(&cam, "cam", "auto", "camera: auto, fit, follow, scroll, page")
	cmd.Flags().StringVar(&trail, "trail", "whole", "how much stays on screen: whole, long, short, comet")
	cmd.Flags().BoolVar(&circle, "circle", false, "start on the circular design rather than the turtle path")
	cmd.Flags().BoolVar(&paused, "paused", false, "start paused")
	cmd.Flags().DurationVar(&cycle, "cycle", 0, "step to the next modulus this often, e.g. 5s")

	cmd.RunE = func(_ *cobra.Command, _ []string) error {
		return tui.Run(tui.Options{
			Seq:    seq,
			Mul:    mul,
			Mod:    mod,
			NoMod:  noMod,
			Speed:  speed,
			MaxPts: maxPts,
			Mono:   mono,
			Render: render,
			Cam:    cam,
			Trail:  trail,
			Circle: circle,
			Paused: paused,
			Cycle:  cycle,
		})
	}
	return cmd
}()
