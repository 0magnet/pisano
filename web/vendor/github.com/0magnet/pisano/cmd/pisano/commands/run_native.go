//go:build !js

package commands

import "github.com/0magnet/pisano/pkg/tui"

// Natively the viewer takes the terminal for itself, so a host that says
// nothing gets that.
var nativeViewer = tui.Run
