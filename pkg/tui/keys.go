package tui

import (
	"bufio"
	"errors"
	"io"
)

// ReadKey reads one keystroke from a raw terminal stream and names it the way
// Model.Key expects — which is the way Bubble Tea names them, so both drivers
// speak one vocabulary and a binding added to the model works in either.
//
// It returns an empty name for input it does not recognise, so a caller can
// simply skip it. The error is only ever the reader's.
//
// This lives here rather than in the driver that needs it because escape
// sequence parsing is exactly the sort of thing that is wrong in ways nobody
// notices until an arrow key does something surprising, and a driver compiled
// only for js/wasm cannot be tested on the machine writing it.
func ReadKey(r *bufio.Reader) (string, error) {
	c, err := r.ReadByte()
	if err != nil {
		return "", err
	}

	switch {
	case c == 0x03:
		return "ctrl+c", nil

	case c == 0x1b:
		// Escape alone is "esc"; escape followed by [ or O introduces an
		// arrow or a page key. A lone escape at end of input is common
		// enough — a user pressing it — that it must not be an error.
		b1, err := r.ReadByte()
		if errors.Is(err, io.EOF) {
			return "esc", nil
		}
		if err != nil {
			return "", err
		}
		if b1 != '[' && b1 != 'O' {
			return "esc", nil
		}
		b2, err := r.ReadByte()
		if err != nil {
			return "", err
		}
		switch b2 {
		case 'A':
			return "up", nil
		case 'B':
			return "down", nil
		case 'C':
			return "right", nil
		case 'D':
			return "left", nil
		case 'H':
			return "home", nil
		case 'F':
			return "end", nil
		case '5', '6':
			// ESC [ 5 ~ and ESC [ 6 ~ — the tilde has to be eaten or it
			// arrives next as a stray keystroke.
			if t, err := r.ReadByte(); err == nil && t != '~' {
				_ = r.UnreadByte()
			}
			if b2 == '5' {
				return "pgup", nil
			}
			return "pgdown", nil
		}
		return "", nil

	case c >= 0x20 && c < 0x7f:
		return string(rune(c)), nil
	}
	return "", nil
}
