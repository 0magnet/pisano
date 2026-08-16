//go:build !js

// The Bubble Tea driver. It is one of two, and it is deliberately thin: it
// translates Bubble Tea's messages into the four calls in model.go and hands
// back the string the model rendered. The other driver, in the wasm build,
// makes the same four calls from a shell applet's raw input loop.
//
// The build tag is what lets the model itself compile for js/wasm. Bubble Tea
// has no port there — it wants to listen for terminal resizes and suspend the
// process, neither of which a browser has — so the driver stays behind, and
// only the driver.

package tui

import (
	"fmt"
	"os"
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
	_, err := tea.NewProgram(teaModel{New(opt)}).Run()
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
