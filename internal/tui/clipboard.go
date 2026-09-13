package tui

import (
	"os"

	"github.com/atotto/clipboard"
	"github.com/aymanbagabas/go-osc52/v2"
)

// copyToClipboard puts text on the system clipboard and reports how it got
// there, for the status line.
//
// Two routes, because neither works everywhere. The native one shells out to
// xclip, wl-copy or pbcopy and fails when none is installed — which is common
// on a bare server. OSC 52 asks the terminal itself to set the clipboard, which
// needs no tools and survives SSH, but some terminals disable it. Both are
// attempted so that whichever is available wins.
func copyToClipboard(text string) string {
	native := clipboard.WriteAll(text) == nil

	// Harmless where unsupported: a terminal ignores an escape it doesn't know.
	osc52.New(text).WriteTo(os.Stdout)

	if native {
		return "Copied to clipboard"
	}
	return "Copied — via the terminal, since no clipboard tool is installed"
}
