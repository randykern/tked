package app

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

// StatusBar describes the behaviour of a status bar component.
type StatusBar interface {
	// Draw renders the status bar for the provided view onto the provided screen.
	Draw(s tcell.Screen, v View)
}

type statusBar struct{}

// NewStatusBar creates a new status bar instance.
func NewStatusBar() StatusBar {
	return &statusBar{}
}

// Draw renders the current status bar.
func (sb *statusBar) Draw(s tcell.Screen, v View) {
	filename := "Untitled"
	cursor := ": 1 1"
	dirty := ""
	if v != nil {
		filename = v.Buffer().GetFilename()
		if filename == "" {
			filename = "Untitled"
		}

		cursorRow, cursorCol := v.Cursor()
		cursor = fmt.Sprintf(": %d %d", cursorRow+1, cursorCol+1)

		if v.Buffer().IsDirty() {
			dirty = "*"
		}
	}

	width, height := s.Size()
	drawText(s, 0, height-1, width-1, height-1, tcell.StyleDefault.Foreground(tcell.ColorWhite), filename+dirty)
	drawText(s, len(filename)+len(dirty), height-1, width-1, height-1, tcell.StyleDefault.Foreground(tcell.ColorWhite), cursor)
}
