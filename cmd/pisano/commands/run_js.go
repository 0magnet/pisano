//go:build js

package commands

import "github.com/0magnet/pisano/pkg/tui"

// There is no default viewer in a browser: a program cannot help itself to a
// terminal that a page owns. A host supplies one or the tui command declines.
var nativeViewer func(tui.Options) error
