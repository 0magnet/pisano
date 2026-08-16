package tui

import (
	"fmt"
	"math"
	"strings"

	"github.com/0magnet/pisano/pkg/pisano"
)

// passColors tint successive circuits. On a closed path the geometry repeats
// exactly, so colour is the only thing that changes between one lap and the
// next — which is what turns a finished figure into something still worth
// watching.
var passColors = []string{
	"\x1b[36m", "\x1b[35m", "\x1b[32m", "\x1b[33m", "\x1b[34m", "\x1b[31m",
}

// segColor tints the segment ending at point i by the pass that drew it.
func (m Model) segColor(i int) string {
	if !m.color || i >= len(m.ptPass) {
		return ""
	}
	return passColors[m.ptPass[i]%len(passColors)]
}

// Frame renders the whole screen as a plain string. It is the only output the
// model produces, so every driver draws the same picture.
func (m Model) Frame() string {
	if m.quit {
		return ""
	}
	if m.w == 0 || m.h == 0 {
		return "starting..."
	}
	return m.frame()
}

func (m Model) frame() string {
	head := m.header()
	foot := m.footer()
	bodyH := m.h - height(head) - height(foot)
	if bodyH < 1 {
		bodyH = 1
	}

	var body string
	if m.help {
		body = m.helpText()
	} else {
		body = m.canvas(m.w, bodyH)
	}
	// The chrome is clipped to the terminal width: a wrapped header or footer
	// would occupy more rows than bodyH budgeted for, and the drawing would be
	// pushed off the top of the screen.
	return clip(head, m.w) + "\n" + padTo(body, bodyH) + "\n" + clipLines(foot, m.w)
}

// clipLines clips every line of a multi-line block.
func clipLines(s string, w int) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = clip(l, w)
	}
	return strings.Join(lines, "\n")
}

func (m Model) header() string {
	what := fmt.Sprintf("%s mod %d", m.period.Seq, m.mod)
	if m.noMod {
		what = m.period.Seq
	}

	var detail string
	if m.view == viewCircle {
		sym := "asymmetrical"
		if n := len(m.design.Reflections()); n == 1 {
			sym = "1 mirror axis"
		} else if n > 1 {
			sym = fmt.Sprintf("%d mirror axes", n)
		}
		if !m.period.Bounded {
			sym = "no period"
		}
		detail = fmt.Sprintf("π=%d · %d zeros · %d chords · %s",
			m.period.Len(), m.period.Zeros(), len(m.design.Chords), sym)
	} else {
		detail = fmt.Sprintf("%d moves · %s", m.period.Len(), m.shape)
	}
	return m.title(what) + m.dim("  "+detail)
}

var camNames = [...]string{"auto", "fit", "follow", "scroll", "page"}

func (m Model) footer() string {
	state := "paused"
	if m.playing {
		state = "playing"
	}
	// The circular design is always braille — chords at arbitrary angles have
	// nowhere to live on a character grid — so reporting the box setting there
	// would name a renderer that is not running.
	modeName := [...]string{"box", "braille"}[m.mode]
	if m.view == viewCircle {
		modeName = "braille"
	}
	camName := camNames[m.cam]
	if m.cam == camAuto {
		camName = "auto/" + camNames[m.camera()]
	}
	viewName := [...]string{"turtle", "circle"}[m.view]

	var progress string
	if m.view == viewCircle {
		progress = fmt.Sprintf("step %d/%d", m.head, m.period.Len())
	} else {
		progress = fmt.Sprintf("lap %d · %d pts · %s",
			m.walk.pass, len(m.pts), trails[m.trailIx].name)
		if m.dropped > 0 {
			progress += fmt.Sprintf(" (%d aged out)", m.dropped)
		}
	}

	if m.cycleEvery > 0 {
		left := float64(m.cycleEvery-m.cycleFor) * frame.Seconds()
		state = fmt.Sprintf("%s · next mod in %.0fs", state, left)
	}

	status := m.dim(fmt.Sprintf("%s · %s · %s · %s · ×%d · %s",
		viewName, modeName, camName, progress, m.speed, state))
	if m.note != "" {
		status = m.warn(m.note)
	}

	// The full hint line is about ninety cells; on a narrower terminal it would
	// be clipped to something misleading, so drop to the essentials and let ?
	// carry the rest.
	var keys string
	if m.w < 96 {
		keys = m.dim(
			m.key("space") + " play  " +
				m.key("←→") + " mod  " +
				m.key("o") + " open  " +
				m.key("?") + " help  " +
				m.key("q") + " quit")
	} else {
		keys = m.dim(
			m.key("space") + " play  " +
				m.key("←→") + " mod  " +
				m.key("↑↓") + " speed  " +
				m.key("o") + " next-open  " +
				m.key("a") + " auto  " +
				m.key("t") + " trail  " +
				m.key("s") + " seq  " +
				m.key("v") + " view  " +
				m.key("m") + " render  " +
				m.key("f") + " cam  " +
				m.key("?") + " help  " +
				m.key("q") + " quit")
	}

	return status + "\n" + keys
}

func (m Model) helpText() string {
	rows := [][2]string{
		{"space", "play / pause"},
		{"← → h l", "modulus down / up"},
		{"H L pgup pgdn", "modulus by ten"},
		{"↑ ↓ k j", "speed double / halve"},
		{"o", "jump to the next modulus whose path never closes"},
		{"a", "step to the next modulus on its own every five seconds"},
		{"t", "trail: whole circuit, long, short, comet"},
		{"0", "no modulus at all — read the terms themselves"},
		{"s", "next sequence: fib, lucas, tri, nat, prime"},
		{"[ ]", "multiply the Fibonacci sequence by less / more"},
		{"v", "turtle path or circular design"},
		{"m", "box-drawing characters or braille dots"},
		{"f", "camera: auto, fit, follow, scroll the least it can, page"},
		{"c", "colour on / off"},
		{"r", "restart the walk"},
		{"q", "quit"},
	}
	var b strings.Builder
	b.WriteString(m.title("keys") + "\n\n")
	for _, r := range rows {
		b.WriteString("  " + m.paint(sgrKey, fmt.Sprintf("%-14s", r[0])) +
			m.dim(r[1]) + "\n")
	}
	b.WriteString("\n" + m.dim(
		"The walk never finishes. A closed path is simply walked again, retracing the\n"+
			"same figure in the next colour, so the loop sweeps round instead of sitting\n"+
			"there done. An open one drifts forever and the camera chases its head, with\n"+
			"the oldest of the path dropped as it goes.\n\n"+
			"Which of the two happens is decided by the net turn over one pass: a quarter\n"+
			"or three-quarter turn closes after four passes, a half turn after two, and no\n"+
			"net turn closes only if the drift is also zero. Press o for the next open one.") + "\n")
	return b.String()
}

// canvas draws the current figure into a w by h character box.
func (m Model) canvas(w, h int) string {
	if m.view == viewCircle {
		return m.circleCanvas(w, h)
	}
	if m.mode == modeBraille {
		return m.turtleBraille(w, h)
	}
	return m.turtleBox(w, h)
}

func bounds(pts []pisano.Pt) (minX, minY, maxX, maxY int) {
	if len(pts) == 0 {
		return 0, 0, 0, 0
	}
	minX, minY = pts[0].X, pts[0].Y
	maxX, maxY = minX, minY
	for _, p := range pts {
		minX, maxX = min(minX, p.X), max(maxX, p.X)
		minY, maxY = min(minY, p.Y), max(maxY, p.Y)
	}
	return
}

// turtleBox draws the path at its true size: one lattice step is one row or two
// columns, because a terminal cell is about twice as tall as it is wide.
//
// There is no scale to choose, so the camera is the only thing that can move.
// Following the head and letting the canvas clip whatever falls outside is what
// makes an endless path watchable here — the drawing scrolls rather than
// shrinking, and the characters keep joining up at full size however long the
// walk runs.
func (m Model) turtleBox(w, h int) string {
	if len(m.pts) < 2 {
		return ""
	}
	var x0, y0 int
	switch m.camera() {
	case camScroll, camPage:
		x0, y0 = m.camX*2, m.camY
	case camFollow:
		head := m.pts[len(m.pts)-1]
		x0, y0 = head.X*2-w/2, head.Y-h/2
	default:
		minX, minY, maxX, maxY := bounds(m.pts)
		x0 = (minX + maxX) - w/2 // minX+maxX is the centre, already doubled
		y0 = (minY+maxY)/2 - h/2
	}

	c := pisano.NewCanvas(x0, y0, x0+w-1, y0+h-1)
	for i := 0; i < len(m.pts)-1; i++ {
		a, b := m.pts[i], m.pts[i+1]
		c.Segment(a.X*2, a.Y, b.X*2, b.Y, m.segColor(i+1))
	}
	return c.String(m.color)
}

// turtleBraille draws the path as dots, scaling so the whole of it stays in
// frame however far it has drifted. A braille dot is half a cell wide and a
// quarter tall, and a terminal cell is roughly twice as tall as it is wide, so
// dots come out very nearly square — one scale serves both axes.
func (m Model) turtleBraille(w, h int) string {
	if len(m.pts) < 2 {
		return ""
	}
	b := pisano.NewBraille(w, h)
	pxW, pxH := b.Size()

	minX, minY, maxX, maxY := bounds(m.pts)
	spanX, spanY := float64(maxX-minX), float64(maxY-minY)

	var scale, cx, cy float64
	switch m.camera() {
	case camScroll, camPage:
		vw, vh := m.viewport()
		scale = brailleScale
		cx, cy = float64(m.camX)+float64(vw)/2, float64(m.camY)+float64(vh)/2
	case camFollow:
		scale = brailleScale
		head := m.pts[len(m.pts)-1]
		cx, cy = float64(head.X), float64(head.Y)
	default:
		scale = math.Inf(1)
		if spanX > 0 {
			scale = (float64(pxW) - 2) / spanX
		}
		if spanY > 0 {
			scale = math.Min(scale, (float64(pxH)-2)/spanY)
		}
		if math.IsInf(scale, 1) {
			scale = brailleScale
		}
		cx = float64(minX) + spanX/2
		cy = float64(minY) + spanY/2
	}

	at := func(p pisano.Pt) (int, int) {
		return int(math.Round((float64(p.X)-cx)*scale)) + pxW/2,
			int(math.Round((float64(p.Y)-cy)*scale)) + pxH/2
	}
	for i := 0; i < len(m.pts)-1; i++ {
		x0, y0 := at(m.pts[i])
		x1, y1 := at(m.pts[i+1])
		b.Line(x0, y0, x1, y1, m.segColor(i+1))
	}
	return b.String(m.color)
}

// circleCanvas traces the circular design chord by chord, in the order the
// remainders actually appear rather than as a finished set of lines. Watching
// it draw is the only way to see that order, which the finished figure
// discards.
//
// Braille only: these are chords at arbitrary angles, and a character grid
// cannot hold them.
func (m Model) circleCanvas(w, h int) string {
	terms := m.period.Terms
	if len(terms) < 2 || m.period.Modulus < 2 {
		return m.dim("  nothing to draw at this modulus")
	}
	b := pisano.NewBraille(w, h)
	pxW, pxH := b.Size()
	r := float64(min(pxW, pxH))/2 - 2
	cx, cy := float64(pxW)/2, float64(pxH)/2

	at := func(k int) (int, int) {
		x, y := pisano.Point(k, m.period.Modulus)
		return int(math.Round(cx + x*r)), int(math.Round(cy + y*r))
	}

	// The ring, so the points have something to sit on.
	if m.color {
		for i := 0; i < 720; i++ {
			t := 2 * math.Pi * float64(i) / 720
			b.Set(int(math.Round(cx+math.Sin(t)*r)), int(math.Round(cy-math.Cos(t)*r)),
				"\x1b[90m")
		}
	}

	for i := 0; i < min(m.head, len(terms)); i++ {
		a, z := terms[i], terms[(i+1)%len(terms)]
		if a == z {
			continue
		}
		col := ""
		if m.color {
			col = passColors[(i*len(passColors)/len(terms))%len(passColors)]
		}
		x0, y0 := at(a)
		x1, y1 := at(z)
		b.Line(x0, y0, x1, y1, col)
	}
	return b.String(m.color)
}

// padTo forces a block to exactly n lines so the footer never drifts up the
// screen as the drawing changes height.
func padTo(s string, n int) string {
	lines := strings.Split(s, "\n")
	if len(lines) > n {
		lines = lines[:n]
	}
	for len(lines) < n {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}
