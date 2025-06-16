package app

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

// StatusBar describes the behaviour of a status bar component.
type StatusBar interface {
	// Draw renders the status bar onto the provided screen.
	Draw(s tcell.Screen)
}

type statusBar struct {
	a *app
}

// NewStatusBar creates a new status bar bound to the given app.
func NewStatusBar(a *app) StatusBar {
	return &statusBar{a: a}
}

// Draw renders the current status bar.
func (sb *statusBar) Draw(s tcell.Screen) {
	filename := "Untitled"
	cursor := ": 1 1"
	dirty := ""
	if len(sb.a.views) > 0 {
		filename = sb.a.views[0].Buffer().GetFilename()
		if filename == "" {
			filename = "Untitled"
		}

		cursorRow, cursorCol := sb.a.views[0].Cursor()
		cursor = fmt.Sprintf(": %d %d", cursorRow+1, cursorCol+1)

		if sb.a.views[0].Buffer().IsDirty() {
			dirty = "*"
		}
	}

	width, height := s.Size()
	drawText(s, 0, height-1, width-1, height-1, tcell.StyleDefault.Foreground(tcell.ColorWhite), filename+dirty)
	drawText(s, len(filename)+len(dirty), height-1, width-1, height-1, tcell.StyleDefault.Foreground(tcell.ColorWhite), cursor)
}
