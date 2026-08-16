//go:build !js

package tui

import (
	"fmt"
	"os"
)

// Run opens the viewer on this terminal.
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
//
// This is the only part of the viewer that is native-only: a browser has no
// stdin to inspect and no terminal to refuse.
func Run(opt Options) error {
	if err := checkTerminal(); err != nil {
		return err
	}
	_, err := Program(opt).Run()
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
