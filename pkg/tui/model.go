// Package pisano/pkg/tui is a terminal viewer for the designs, built on
// bubbletea and running in the alternate screen.
//
// The walk never finishes. An open path drifts forever by its nature, and a
// closed one is walked again from where it left off — retracing exactly the
// same figure, but in the next colour, so the loop sweeps round rather than
// sitting there finished. Either way the drawing keeps only its most recent
// stretch: the oldest of it is dropped as the turtle moves on, which is what
// lets it run indefinitely at a fixed size in a fixed amount of memory.
package tui

import (
	"fmt"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/0magnet/pisano/pkg/pisano"
)

// frame is the animation tick. Fast enough to look continuous, slow enough
// that the redraw cost stays invisible.
const frame = 40 * time.Millisecond

// holdFrames is how long the completed circular figure stays on screen before
// the trace starts over — about a second.
const holdFrames = 25

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(frame, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// View modes.
const (
	viewTurtle = iota
	viewCircle
	numViews
)

// Render modes. Box drawing is the default: one lattice step is one row or two
// columns, which is the figure at its true size and the reason the characters
// join up cleanly. Braille trades those joins for an arbitrary scale.
const (
	modeBox = iota
	modeBraille
	numModes
)

// Camera modes. Auto is the default because the right answer differs by figure
// and is not a matter of taste: an open path has a head that runs away from
// you, so the view has to keep up, while a closed one wants to sit still and be
// looked at.
//
// Follow, scroll and page all keep up, in increasing order of how much they
// leave alone:
//
//   - follow pins the head to the centre, so the entire drawing slides under it
//     on every frame. It is exact and unwatchable: nothing ever holds still, and
//     the head's wandering within a motif shakes the whole screen.
//   - scroll leaves the view where it is until the head reaches a margin of the
//     edge, then moves by the least it can to keep it inside. Wandering within
//     the motif moves nothing; only the net drift does, and only along the axis
//     it drifts on.
//   - page does the same but shoves the view most of a screen at a time, so it
//     is perfectly still for a long stretch and then jumps once.
const (
	camAuto = iota
	camFit
	camFollow
	camScroll
	camPage
	numCams
)

// brailleScale is how many dots one lattice step spans when the view is not
// scaling to fit.
const brailleScale = 6

// trails is how much of the path stays on screen, cycled with t. Zero means
// keep the whole figure: for a closed path that is exactly one full circuit, so
// the shape stands still while the colour sweeps around it; for an open path
// there is no such thing, so it falls back to the memory limit.
var trails = []struct {
	name string
	n    int
}{
	{"whole", 0},
	{"long", 1500},
	{"short", 400},
	{"comet", 90},
}

type seqEntry struct {
	name  string
	build func(mul int) pisano.Sequence
}

// sequences is the cycle order for the s key. Fibonacci first because it is
// the one the construction is usually shown with.
var sequences = []seqEntry{
	{"fib", func(mul int) pisano.Sequence {
		if mul > 1 {
			return pisano.Scaled(mul)
		}
		return pisano.Fibonacci()
	}},
	{"lucas", func(int) pisano.Sequence { return pisano.Lucas() }},
	{"tri", func(int) pisano.Sequence { return pisano.Triangular() }},
	{"nat", func(int) pisano.Sequence { return pisano.Naturals() }},
	{"prime", func(int) pisano.Sequence { return pisano.Primes() }},
}

// walker keeps a turtle mid-stride so the path can be extended one term at a
// time. The viewer runs indefinitely, so recomputing the path from the start
// on every frame is not an option; and stepping by whole passes would make the
// drawing jump rather than crawl.
type walker struct {
	t    pisano.Turtle
	p    pisano.Period
	i    int  // next index into the run currently being consumed
	head bool // still working through the run-in
	pass int  // completed passes of the repeating block
}

func newWalker(p pisano.Period) *walker {
	return &walker{p: p, head: len(p.Head) > 0}
}

// next consumes one term. It reports the turtle's position, whether the term
// moved it at all — a zero term does not — and which pass the term belonged
// to, which is what the colouring is keyed on.
func (w *walker) next() (pisano.Pt, bool, int) {
	pass := w.pass

	var term int
	switch {
	case w.head:
		term = w.p.Head[w.i]
		w.i++
		if w.i >= len(w.p.Head) {
			w.head, w.i = false, 0
		}
	case len(w.p.Terms) == 0:
		return w.t.Pos, false, pass
	default:
		term = w.p.Terms[w.i]
		w.i++
		if w.i >= len(w.p.Terms) {
			w.i, w.pass = 0, w.pass+1
		}
	}

	moved := w.t.Step(term)
	return w.t.Pos, moved, pass
}

// Model is the viewer state.
type Model struct {
	w, h int

	seqIdx int
	mul    int
	mod    int
	noMod  bool // read the sequence with no modulus at all

	period pisano.Period
	shape  pisano.Shape
	design pisano.Design

	// A zero term moves the turtle nowhere, so a pass contributes fewer
	// points than it has terms. movesPerPass is what turns a circuit count
	// into a point count.
	movesPerPass int

	walk    *walker
	pts     []pisano.Pt
	ptPass  []int // which pass laid down each point
	dropped int   // points aged out of the trail

	// head and hold belong to the circular view, which traces a fixed figure
	// rather than walking an open-ended one.
	head int
	hold int

	// The scrolling camera is the one piece of view state that has to persist
	// between frames: its whole point is to stay where it was put.
	camX, camY int
	camSet     bool

	playing bool
	speed   int

	// cycleEvery counts frames, not time: the tick is the only clock the
	// model has, and keeping it in frames means pausing pauses the cycling
	// too, which is what anyone would expect of a paused viewer.
	cycleEvery int
	cycleFor   int

	view    int
	mode    int
	cam     int
	trailIx int
	color   bool
	help    bool

	maxPts int
	note   string
	quit   bool
}

// Options configure the viewer's starting state.
type Options struct {
	Seq    string
	Mul    int
	Mod    int
	NoMod  bool
	Speed  int
	MaxPts int
	Mono   bool
	Render string // box, braille
	Cam    string // auto, fit, follow, scroll, page
	Trail  string // whole, long, short, comet
	Circle bool
	Paused bool
	Cycle  time.Duration // advance the modulus on its own every so often
}

// New builds a viewer.
func New(opt Options) Model {
	m := Model{
		mul:     max(1, opt.Mul),
		mod:     max(1, opt.Mod),
		noMod:   opt.NoMod,
		playing: !opt.Paused,
		speed:   max(1, opt.Speed),
		color:   !opt.Mono,
		maxPts:  opt.MaxPts,
	}
	if m.maxPts <= 0 {
		m.maxPts = 60000
	}
	for i, s := range sequences {
		if s.name == opt.Seq {
			m.seqIdx = i
		}
	}
	if opt.Render == "braille" {
		m.mode = modeBraille
	}
	switch opt.Cam {
	case "fit":
		m.cam = camFit
	case "follow":
		m.cam = camFollow
	case "scroll":
		m.cam = camScroll
	case "page":
		m.cam = camPage
	}
	for i, t := range trails {
		if t.name == opt.Trail {
			m.trailIx = i
		}
	}
	if opt.Circle {
		m.view = viewCircle
	}
	m.cycleEvery = int(opt.Cycle / frame)
	m.reload()
	return m
}

func (m Model) Init() tea.Cmd { return tick() }

// sequence builds the currently selected sequence.
func (m Model) sequence() pisano.Sequence {
	return sequences[m.seqIdx].build(m.mul)
}

// camera resolves auto to whichever mode suits the figure.
func (m Model) camera() int {
	if m.cam != camAuto {
		return m.cam
	}
	if m.shape.Closed {
		return camFit
	}
	return camScroll
}

// bodyH is the height of the drawing area: everything but the one-line header
// and the two-line footer.
func (m Model) bodyH() int { return max(1, m.h-3) }

// viewport is how much of the lattice fits on screen, in lattice steps. A cell
// is about twice as tall as it is wide, so box drawing spends two columns on a
// horizontal step and one row on a vertical one.
func (m Model) viewport() (int, int) {
	w, h := m.w, m.bodyH()
	if m.mode == modeBraille {
		return max(1, w*2/brailleScale), max(1, h*4/brailleScale)
	}
	return max(1, w/2), max(1, h)
}

// updateCamera moves the view, and only when it has to.
//
// The head is free to wander anywhere inside a margin of the edge without
// costing a single frame of movement — which is the whole difference from
// follow, where the head is pinned and everything else moves instead. Once it
// does cross the margin, scroll brings it back to that margin and no further,
// while page shoves the view on so the head lands at the opposite margin with a
// screen of room ahead of it.
func (m *Model) updateCamera() {
	cam := m.camera()
	if m.view != viewTurtle || (cam != camScroll && cam != camPage) || len(m.pts) == 0 {
		return
	}
	vw, vh := m.viewport()
	if vw < 4 || vh < 4 {
		return
	}
	head := m.pts[len(m.pts)-1]

	if !m.camSet {
		m.camX, m.camY = head.X-vw/2, head.Y-vh/2
		m.camSet = true
		return
	}

	mx, my := max(1, vw/8), max(1, vh/8)

	// near is where the head is put when it breaches the edge it is heading
	// for: the same margin it just crossed, or the far one for a page.
	nearX, farX := mx, vw-1-mx
	nearY, farY := my, vh-1-my
	if cam == camPage {
		nearX, farX = farX, nearX
		nearY, farY = farY, nearY
	}

	switch {
	case head.X < m.camX+mx:
		m.camX = head.X - nearX
	case head.X > m.camX+vw-1-mx:
		m.camX = head.X - farX
	}
	switch {
	case head.Y < m.camY+my:
		m.camY = head.Y - nearY
	case head.Y > m.camY+vh-1-my:
		m.camY = head.Y - farY
	}
}

// trailLen is how many points to keep. For a closed path "whole" is one full
// circuit exactly, so the figure stands complete and still while the colour
// moves through it. An open path has no full circuit, so it keeps as much as
// memory allows and lets the camera do the rest.
func (m Model) trailLen() int {
	n := trails[m.trailIx].n
	if n == 0 {
		if m.shape.Closed && m.movesPerPass > 0 {
			n = m.movesPerPass*m.shape.Periods + 1
		} else {
			n = m.maxPts
		}
	}
	return min(n, m.maxPts)
}

// reload recomputes everything that depends on the sequence or the modulus and
// restarts the walk.
func (m *Model) reload() {
	s := m.sequence()
	m.note = ""

	if m.noMod {
		p, err := pisano.UnreducedPeriod(s)
		if err != nil {
			// Only reachable if a sequence is added without unreduced
			// support; fall back rather than leaving the viewer blank.
			m.noMod = false
			m.note = err.Error()
			m.reload()
			return
		}
		m.period = p
	} else {
		m.period = pisano.Compute(s, m.mod, 0)
	}

	m.shape = pisano.Classify(m.period.Terms)
	m.design = pisano.Circular(m.period)
	m.movesPerPass = countMoves(m.period.Terms)
	m.restart()
}

func countMoves(terms []int) int {
	n := 0
	for _, t := range terms {
		if t != 0 {
			n++
		}
	}
	return n
}

// restart rewinds the walk without recomputing the arithmetic.
func (m *Model) restart() {
	m.walk = newWalker(m.period)
	m.pts = append(m.pts[:0], pisano.Pt{})
	m.ptPass = append(m.ptPass[:0], 0)
	m.dropped = 0
	m.head = 0
	m.hold = 0
	m.camSet = false

	// Draw one circuit up front so a closed figure is whole the moment it
	// appears, rather than crawling into existence every time the modulus
	// changes. An open path has nothing to complete, so it starts bare.
	if m.shape.Closed && m.movesPerPass > 0 {
		m.advance(m.movesPerPass * m.shape.Periods)
	}
}

// advance walks n terms and drops whatever has aged out of the trail.
func (m *Model) advance(n int) {
	if m.view == viewCircle {
		// The circular figure is fixed, so tracing it loops. It rests on the
		// finished figure for a moment first — without the pause the design
		// is complete for a single frame and then gone, which is the one
		// frame you actually wanted to look at.
		total := len(m.period.Terms)
		if total == 0 {
			return
		}
		if m.head >= total {
			m.hold++
			if m.hold > holdFrames {
				m.head, m.hold = 0, 0
			}
			return
		}
		m.head = min(total, m.head+n)
		return
	}

	for i := 0; i < n; i++ {
		p, moved, pass := m.walk.next()
		if !moved {
			continue
		}
		m.pts = append(m.pts, p)
		m.ptPass = append(m.ptPass, pass)
	}
	m.trim()
}

// trim discards the oldest of the path. It runs in batches rather than every
// frame: each trim shifts the whole slice down, so trimming one point at a
// time would copy the entire trail on every tick for the sake of one element.
func (m *Model) trim() {
	limit := m.trailLen()
	slack := max(64, limit/8)
	if len(m.pts) <= limit+slack {
		return
	}
	drop := len(m.pts) - limit
	m.pts = append(m.pts[:0], m.pts[drop:]...)
	m.ptPass = append(m.ptPass[:0], m.ptPass[drop:]...)
	m.dropped += drop
}

// cycle steps to the next modulus on its own once the dwell has elapsed. It is
// what a shell loop around the whole program would be reaching for, without the
// loop: one process, one terminal takeover, and a Ctrl-C that ends the whole
// thing rather than just the current modulus.
func (m *Model) cycle() {
	if m.cycleEvery <= 0 {
		return
	}
	m.cycleFor++
	if m.cycleFor < m.cycleEvery {
		return
	}
	m.cycleFor = 0
	m.noMod = false
	m.mod++
	m.reload()
}

// Run opens the viewer.
//
// The alternate screen is declared by the view rather than requested here, so
// the terminal it was launched from is left exactly as it was on exit.
//
// It refuses to start without an interactive terminal. A full-screen program
// takes the terminal over and puts it in raw mode, where Ctrl-C is no longer a
// signal but a keystroke the program itself has to honour — so one started by
// mistake, from a pipeline or a shell loop, is markedly harder to get out of
// than an ordinary command. Better to say so up front than to seize the screen
// and hope.
func Run(opt Options) error {
	if err := checkTerminal(); err != nil {
		return err
	}
	_, err := tea.NewProgram(New(opt)).Run()
	return err
}

func checkTerminal() error {
	for _, f := range []struct {
		name string
		file *os.File
	}{{"stdin", os.Stdin}, {"stdout", os.Stdout}} {
		fi, err := f.file.Stat()
		if err != nil {
			return fmt.Errorf("tui: cannot inspect %s: %w", f.name, err)
		}
		if fi.Mode()&os.ModeCharDevice == 0 {
			return fmt.Errorf(
				"tui needs an interactive terminal, but %s is a pipe or a file.\n"+
					"  For output you can redirect, use: pisano turtle --mod N -o out.svg\n"+
					"  To step through moduli in one session, use: pisano tui --cycle 5s",
				f.name)
		}
	}
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height

	case tickMsg:
		if m.playing {
			m.advance(m.speed)
			m.cycle()
		}
		cmd = tick()

	case tea.KeyPressMsg:
		// v2 splits presses from releases; only presses drive anything here.
		var next tea.Model
		next, cmd = m.onKey(msg)
		m = next.(Model)
	}
	m.updateCamera()
	return m, cmd
}

func (m Model) onKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "ctrl+c":
		m.quit = true
		return m, tea.Quit

	case " ":
		m.playing = !m.playing

	case "a":
		if m.cycleEvery > 0 {
			m.cycleEvery = 0
		} else {
			m.cycleEvery = int(5 * time.Second / frame)
		}
		m.cycleFor = 0

	case "right", "l":
		m.noMod = false
		m.mod++
		m.cycleFor = 0
		m.reload()
	case "left", "h":
		m.noMod = false
		if m.mod > 1 {
			m.mod--
		}
		m.cycleFor = 0
		m.reload()
	case "pgdown", "L":
		m.noMod = false
		m.mod += 10
		m.reload()
	case "pgup", "H":
		m.noMod = false
		m.mod = max(1, m.mod-10)
		m.reload()

	case "up", "k":
		m.speed = min(512, m.speed*2)
	case "down", "j":
		m.speed = max(1, m.speed/2)

	case "o":
		// Jump to the next modulus whose turtle path never closes.
		m.noMod = false
		m.nextOpen()

	case "s":
		m.seqIdx = (m.seqIdx + 1) % len(sequences)
		m.reload()
	case "]":
		m.mul++
		m.reload()
	case "[":
		m.mul = max(1, m.mul-1)
		m.reload()

	case "0":
		m.noMod = !m.noMod
		m.reload()

	case "v":
		m.view = (m.view + 1) % numViews
		m.head, m.hold = 0, 0
	case "m":
		m.mode = (m.mode + 1) % numModes
		m.camSet = false // the viewport is a different size in lattice steps
	case "f":
		m.cam = (m.cam + 1) % numCams
		m.camSet = false
	case "t":
		m.trailIx = (m.trailIx + 1) % len(trails)
		m.trim()
	case "c":
		m.color = !m.color
	case "r":
		m.restart()
	case "?":
		m.help = !m.help
	}
	return m, nil
}

// nextOpen walks forward to the next modulus with an open turtle path, giving
// up after a bounded search rather than scanning forever.
func (m *Model) nextOpen() {
	s := m.sequence()
	for i := 1; i <= 500; i++ {
		cand := m.mod + i
		if pisano.Classify(pisano.Compute(s, cand, 0).Terms).Closed {
			continue
		}
		m.mod = cand
		m.reload()
		return
	}
	m.note = fmt.Sprintf("no open path within 500 moduli of %d", m.mod)
}
