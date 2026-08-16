package tui

import (
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/0magnet/pisano/pkg/pisano"
)

// frames drives the model headlessly through the same four calls both drivers
// make, so the tests cover what the terminal and the browser actually run
// rather than one driver's idea of it.
func frames(t *testing.T, opt Options, w, h, n int) Model {
	t.Helper()
	m := New(opt)
	m.Resize(w, h)
	for i := 0; i < n; i++ {
		m.Advance()
	}
	return m
}

func (m Model) tick(n int) Model {
	for i := 0; i < n; i++ {
		m.Advance()
	}
	return m
}

func (m Model) press(k rune) Model {
	m.Key(string(k))
	return m
}

// The view must always be exactly as tall as the terminal, or the footer walks
// up the screen as the drawing changes size.
func TestViewFillsTheScreen(t *testing.T) {
	for _, opt := range []Options{
		{Mod: 25, Speed: 8},
		{Mod: 10, Speed: 8},
		{Mod: 1},
		{Mod: 10, Circle: true},
		{Mod: 25, Render: "braille", Cam: "fit"},
		{Mod: 25, Trail: "comet"},
		{Mod: 25, Cam: "scroll"},
		{Mod: 25, Cam: "scroll", Render: "braille"},
		{Mod: 25, Cam: "page"},
		{NoMod: true},
	} {
		for _, size := range [][2]int{{80, 24}, {40, 12}, {200, 60}} {
			m := frames(t, opt, size[0], size[1], 20)
			got := strings.Count(m.Frame(), "\n") + 1
			if got != size[1] {
				t.Errorf("mod %d %dx%d: view is %d lines, want %d",
					opt.Mod, size[0], size[1], got, size[1])
			}
		}
	}
}

// The walk must never finish. A closed path is what would once have stopped,
// so it is the one to check: the figure repeats, but the walk behind it keeps
// going and the lap count keeps climbing.
func TestClosedPathKeepsWalking(t *testing.T) {
	m := frames(t, Options{Mod: 11, Speed: 16}, 80, 24, 5)
	if !m.shape.Closed {
		t.Fatal("fib mod 11 should be closed")
	}
	before := m.walk.pass
	m = m.tick(200)
	if m.walk.pass <= before {
		t.Errorf("lap count stuck at %d after 200 ticks", m.walk.pass)
	}
}

// Re-walking a closed path has to actually recolour it, or there is nothing to
// watch: the geometry is identical lap to lap, so colour is the only thing that
// can change.
func TestRewalkingRecolours(t *testing.T) {
	m := frames(t, Options{Mod: 11, Speed: 4}, 80, 24, 5)
	seen := map[string]bool{}
	for i := 0; i < 300; i++ {
		m = m.tick(1)
		seen[m.segColor(len(m.ptPass)-1)] = true
	}
	if len(seen) < 2 {
		t.Errorf("the head kept one colour across %d laps", m.walk.pass)
	}
}

// A closed path with the whole-circuit trail holds exactly one circuit, so the
// figure stands still and complete however long it runs.
func TestWholeTrailHoldsOneCircuit(t *testing.T) {
	m := frames(t, Options{Mod: 11, Speed: 8, Trail: "whole"}, 80, 24, 5)
	want := m.movesPerPass*m.shape.Periods + 1
	m = m.tick(400)
	if slack := max(64, want/8); len(m.pts) > want+slack {
		t.Errorf("held %d points, want about %d", len(m.pts), want)
	}
	if m.dropped == 0 {
		t.Error("nothing was ever aged out of the trail")
	}
}

// An open path must keep growing and keep drifting; that is the whole point.
func TestOpenPathKeepsDrifting(t *testing.T) {
	m := frames(t, Options{Mod: 25, Speed: 16}, 80, 24, 5)
	if m.shape.Closed {
		t.Fatal("fib mod 25 should be open")
	}
	first := m.pts[len(m.pts)-1]
	m = m.tick(300)
	last := m.pts[len(m.pts)-1]
	if first == last {
		t.Error("the head never moved")
	}
	if m.walk.pass < 2 {
		t.Errorf("only %d passes walked", m.walk.pass)
	}
}

// Every trail setting has to bound memory, or a viewer left running overnight
// eats the machine.
func TestEveryTrailBoundsMemory(t *testing.T) {
	for _, tr := range trails {
		m := frames(t, Options{Mod: 25, Speed: 256, MaxPts: 5000, Trail: tr.name}, 80, 24, 300)
		limit := m.trailLen()
		if slack := max(64, limit/8); len(m.pts) > limit+slack {
			t.Errorf("trail %q held %d points past a limit of %d", tr.name, len(m.pts), limit)
		}
	}
}

// Following the head means the drawing is always under the camera. After a long
// drift the body must still have something in it — if the camera lagged, the
// path would have scrolled out of frame and left a blank screen.
func TestFollowKeepsTheHeadInFrame(t *testing.T) {
	m := frames(t, Options{Mod: 25, Speed: 32, Cam: "follow"}, 80, 24, 400)
	body := m.canvas(80, 20)
	if strings.TrimSpace(body) == "" {
		t.Fatal("the path drifted out of frame")
	}
	if !strings.ContainsAny(body, "─│┌┐└┘├┤┬┴┼") {
		t.Errorf("no box drawing in the body: %q", body)
	}
}

// Page mostly does not move at all: it should be stationary for a long stretch
// and then jump once.
func TestPageCameraHoldsStill(t *testing.T) {
	m := frames(t, Options{Mod: 25, Speed: 4, Cam: "page"}, 80, 24, 5)
	moved, ticks := 0, 600
	for i := 0; i < ticks; i++ {
		before := [2]int{m.camX, m.camY}
		m = m.tick(1)
		if before != [2]int{m.camX, m.camY} {
			moved++
		}
	}
	if moved == 0 {
		t.Fatal("the camera never moved at all across a drifting path")
	}
	if moved*10 > ticks {
		t.Errorf("camera moved on %d of %d frames; it should hold still between jumps",
			moved, ticks)
	}
}

// Holding still is only acceptable if the head is still on screen. Whenever the
// camera does move it must move far enough to keep it there.
func TestScrollKeepsHeadInFrame(t *testing.T) {
	for _, cam := range []string{"scroll", "page"} {
		for _, render := range []string{"box", "braille"} {
			m := frames(t, Options{Mod: 25, Speed: 6, Cam: cam, Render: render}, 80, 24, 5)
			for i := 0; i < 600; i++ {
				m = m.tick(1)
				vw, vh := m.viewport()
				head := m.pts[len(m.pts)-1]
				if head.X < m.camX || head.X > m.camX+vw-1 ||
					head.Y < m.camY || head.Y > m.camY+vh-1 {
					t.Fatalf("%s/%s: head %v outside viewport %dx%d at (%d,%d)",
						cam, render, head, vw, vh, m.camX, m.camY)
				}
			}
		}
	}
}

// Scroll must move by the least it can: never more than the head itself moved
// in that frame, which is at most one lattice step per term walked.
func TestScrollMovesTheMinimum(t *testing.T) {
	speed := 4
	m := frames(t, Options{Mod: 25, Speed: speed, Cam: "scroll"}, 80, 24, 5)
	for i := 0; i < 600; i++ {
		bx, by := m.camX, m.camY
		m = m.tick(1)
		if d := abs(m.camX - bx); d > speed {
			t.Fatalf("camera moved %d columns for %d steps of walking", d, speed)
		}
		if d := abs(m.camY - by); d > speed {
			t.Fatalf("camera moved %d rows for %d steps of walking", d, speed)
		}
	}
}

// Page is the opposite trade: it holds still far longer, so when it does move it
// has to cover most of a screen or it is just a noisier scroll.
func TestPageJumpsByPages(t *testing.T) {
	m := frames(t, Options{Mod: 25, Speed: 4, Cam: "page"}, 80, 24, 5)
	vw, _ := m.viewport()
	moved := 0
	for i := 0; i < 600; i++ {
		before := m.camX
		m = m.tick(1)
		if d := abs(m.camX - before); d != 0 {
			moved++
			if d < vw/2 {
				t.Fatalf("page shifted %d columns on a %d-wide viewport", d, vw)
			}
		}
	}
	if moved == 0 {
		t.Fatal("page never moved across a drifting path")
	}
}

// The point of both is that the head is free to wander inside the frame without
// costing a frame of movement. Follow has no such freedom — it moves whenever
// the head does — so either must move on strictly fewer frames than that.
func TestScrollMovesLessOftenThanTheHead(t *testing.T) {
	for _, cam := range []string{"scroll", "page"} {
		m := frames(t, Options{Mod: 25, Speed: 2, Cam: cam}, 80, 24, 5)
		camMoves, headMoves := 0, 0
		for i := 0; i < 600; i++ {
			bc := [2]int{m.camX, m.camY}
			bh := m.pts[len(m.pts)-1]
			m = m.tick(1)
			if bc != [2]int{m.camX, m.camY} {
				camMoves++
			}
			if bh != m.pts[len(m.pts)-1] {
				headMoves++
			}
		}
		if headMoves == 0 {
			t.Fatalf("%s: the head never moved", cam)
		}
		if camMoves >= headMoves {
			t.Errorf("%s: camera moved on %d frames against %d head moves",
				cam, camMoves, headMoves)
		}
	}
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// Every modulus at every sequence must render without panicking. Degenerate
// cases — one point, an all-zero period, a sequence that never repeats — are
// the ones that would.
func TestNoPanicAcrossModuli(t *testing.T) {
	for _, seq := range []string{"fib", "lucas", "tri", "nat", "prime"} {
		for mod := 1; mod <= 40; mod++ {
			for _, circle := range []bool{false, true} {
				m := frames(t, Options{Seq: seq, Mod: mod, Speed: 8, Circle: circle}, 60, 20, 6)
				if m.Frame() == "" {
					t.Errorf("%s mod %d: empty view", seq, mod)
				}
			}
		}
	}
}

// A period of all zeros moves the turtle nowhere. It must not spin.
func TestAllZeroPeriodTerminates(t *testing.T) {
	m := frames(t, Options{Mod: 1, Speed: 64}, 80, 24, 50)
	if len(m.pts) != 1 {
		t.Errorf("mod 1 drew %d points, want just the origin", len(m.pts))
	}
}

// o must land on a modulus whose path is genuinely open.
func TestNextOpenFindsAnOpenPath(t *testing.T) {
	m := frames(t, Options{Mod: 11, Speed: 1}, 80, 24, 1)
	if !m.shape.Closed {
		t.Fatal("expected to start on a closed path")
	}
	m = m.press('o')
	if m.shape.Closed {
		t.Errorf("o landed on mod %d, which is closed", m.mod)
	}
	if pisano.Classify(pisano.Compute(pisano.Fibonacci(), m.mod, 0).Terms).Closed {
		t.Errorf("mod %d disagrees with the library", m.mod)
	}
}

// Auto has to pick the mode that suits the figure: chase an open head, sit
// still for a closed loop.
func TestAutoCameraPicksByShape(t *testing.T) {
	open := frames(t, Options{Mod: 25}, 80, 24, 2)
	if open.camera() != camScroll {
		t.Error("auto should scroll an open path")
	}
	closed := frames(t, Options{Mod: 11}, 80, 24, 2)
	if closed.camera() != camFit {
		t.Error("auto should sit still on a closed path")
	}
}

// Toggling every mode in turn must leave a drawable state.
func TestTogglesStayDrawable(t *testing.T) {
	m := frames(t, Options{Mod: 25, Speed: 4}, 80, 24, 10)
	for _, k := range []rune{'v', 'm', 'f', 't', 'c', 's', ']', '[', '0', 'r', '?'} {
		m = m.press(k).tick(1)
		if got := strings.Count(m.Frame(), "\n") + 1; got != 24 {
			t.Errorf("after %q: view is %d lines, want 24", k, got)
		}
	}
}

// The braille canvas has to put dots where it is told.
func TestBrailleDots(t *testing.T) {
	b := pisano.NewBraille(1, 1)
	if got := b.String(false); got != "" {
		t.Errorf("empty canvas rendered %q", got)
	}
	for x := 0; x < 2; x++ {
		for y := 0; y < 4; y++ {
			b.Set(x, y, "")
		}
	}
	if got := b.String(false); got != "⣿" {
		t.Errorf("full cell rendered %q, want %q", got, "⣿")
	}

	b2 := pisano.NewBraille(1, 1)
	b2.Set(-1, 0, "")
	b2.Set(0, -1, "")
	b2.Set(2, 0, "")
	b2.Set(0, 4, "")
	if got := b2.String(false); got != "" {
		t.Errorf("out-of-range dots drew %q", got)
	}
}

func TestBrailleLine(t *testing.T) {
	b := pisano.NewBraille(4, 1)
	b.Line(0, 0, 7, 0, "")
	if got, want := b.String(false), "⠉⠉⠉⠉"; got != want {
		t.Errorf("horizontal line rendered %q, want %q", got, want)
	}
}

// The viewer must refuse to start without a terminal. Started by mistake from a
// pipeline it seizes the screen and puts it in raw mode, where Ctrl-C is a
// keystroke rather than a signal — so a shell loop of them is caught one
// instance at a time, which is exactly the trap worth closing.
func TestRunRefusesWithoutATerminal(t *testing.T) {
	err := Run(Options{Mod: 25})
	if err == nil {
		t.Fatal("Run started with a non-terminal stdin")
	}
	for _, want := range []string{"interactive terminal", "pisano turtle", "--cycle"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error is missing %q: %v", want, err)
		}
	}
}

// Cycling has to actually advance the modulus, and only on its own clock.
func TestCycleAdvancesTheModulus(t *testing.T) {
	every := 10
	m := frames(t, Options{Mod: 5, Speed: 2, Cycle: time.Duration(every) * frame}, 80, 24, 0)
	if m.cycleEvery != every {
		t.Fatalf("cycleEvery is %d, want %d", m.cycleEvery, every)
	}
	m = m.tick(every - 1)
	if m.mod != 5 {
		t.Errorf("modulus moved to %d before the dwell elapsed", m.mod)
	}
	m = m.tick(1)
	if m.mod != 6 {
		t.Errorf("modulus is %d after the dwell, want 6", m.mod)
	}
	m = m.tick(every * 3)
	if m.mod != 9 {
		t.Errorf("modulus is %d after three more dwells, want 9", m.mod)
	}
}

// A paused viewer must not cycle either; the tick is the only clock there is.
func TestCycleStopsWhenPaused(t *testing.T) {
	m := frames(t, Options{Mod: 5, Cycle: 10 * frame, Paused: true}, 80, 24, 50)
	if m.mod != 5 {
		t.Errorf("modulus advanced to %d while paused", m.mod)
	}
}

// Changing the modulus by hand restarts the dwell, so a manual step does not
// get yanked away a frame later.
func TestManualStepResetsTheDwell(t *testing.T) {
	m := frames(t, Options{Mod: 5, Cycle: 10 * frame}, 80, 24, 9)
	m = m.press('l')
	if m.cycleFor != 0 {
		t.Errorf("dwell counter is %d after a manual step, want 0", m.cycleFor)
	}
	m = m.tick(9)
	if m.mod != 6 {
		t.Errorf("modulus is %d; the dwell did not restart", m.mod)
	}
}

// A full-screen program must never emit a line wider than the terminal. If it
// does the terminal wraps it, the chrome then occupies more rows than the
// layout budgeted, and the top of the drawing is pushed off the screen — which
// is exactly what happened in a browser at 700px before this was checked.
func TestNoLineExceedsTheWidth(t *testing.T) {
	for _, size := range [][2]int{{40, 12}, {60, 16}, {80, 24}, {96, 30}, {200, 60}} {
		for _, opt := range []Options{{Mod: 25}, {Mod: 11}, {Mod: 10, Circle: true}} {
			m := frames(t, opt, size[0], size[1], 30)
			for i, line := range strings.Split(m.Frame(), "\n") {
				if n := visibleWidth(line); n > size[0] {
					t.Errorf("%dx%d mod %d line %d is %d cells wide, max %d: %q",
						size[0], size[1], opt.Mod, i, n, size[0], line)
				}
			}
		}
	}
}

// visibleWidth counts cells, skipping escape sequences.
func visibleWidth(s string) int {
	n := 0
	for i := 0; i < len(s); {
		if s[i] == 0x1b {
			for i < len(s) && s[i] != 'm' {
				i++
			}
			i++
			continue
		}
		_, size := utf8.DecodeRuneInString(s[i:])
		n++
		i += size
	}
	return n
}

func TestClipPreservesStyleAndCloses(t *testing.T) {
	got := clip("\x1b[36mabcdef\x1b[0m", 3)
	if visibleWidth(got) != 3 {
		t.Errorf("clip to 3 gave width %d: %q", visibleWidth(got), got)
	}
	if !strings.HasSuffix(got, sgrReset) {
		t.Errorf("clipped styled text does not close its style: %q", got)
	}
	if s := "short"; clip(s, 40) != s {
		t.Errorf("clip shortened a string that already fit")
	}
}
