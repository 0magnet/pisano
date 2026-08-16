// The Bubble Tea driver. It is thin on purpose: it translates Bubble Tea's
// messages into the four calls in model.go and hands back the string the model
// rendered.
//
// It used to be one of two drivers, behind a !js build tag, because Bubble Tea
// did not build for WebAssembly — it wanted to listen for terminal resizes and
// suspend the process, and a browser has neither. 0magnet/bubbletea's wasm
// branch answers both (the host knows the size; there is nothing to suspend),
// so there is one driver again and the browser runs the same event loop the
// terminal does.
//
// What still differs by platform is not the loop but who owns the terminal:
// natively the program takes stdin and stdout, and in a browser the host hands
// it a reader, a writer and a size. That is what Program leaves to the caller.

package tui

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(FrameInterval, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// teaModel adapts Model to Bubble Tea's interface.
type teaModel struct{ m Model }

func (t teaModel) Init() tea.Cmd { return tick() }

func (t teaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.m.Resize(msg.Width, msg.Height)

	case tickMsg:
		t.m.Advance()
		return t, tick()

	case tea.KeyPressMsg:
		// v2 splits presses from releases; only presses drive anything.
		if !t.m.Key(msg.String()) {
			return t, tea.Quit
		}
	}
	return t, nil
}

func (t teaModel) View() tea.View {
	var v tea.View
	v.AltScreen = true
	v.SetContent(t.m.Frame())
	return v
}

// Program builds the viewer as a Bubble Tea program, without running it.
//
// Callers supply whatever the program should read, write and believe about its
// size. Natively that is stdin, stdout and the tty, which Run fills in; in a
// browser it is the host's pipes and its own idea of how many cells it has,
// since nothing else can know.
func Program(opt Options, teaOpts ...tea.ProgramOption) *tea.Program {
	return tea.NewProgram(teaModel{New(opt)}, teaOpts...)
}
