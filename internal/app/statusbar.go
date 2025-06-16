package app

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

// StatusBar describes the behaviour of a status bar component.
type StatusBar interface {
	// Draw renders the status bar for the provided view onto the provided screen.
	Draw(s tcell.Screen, v View)
	// Input displays a prompt on the status bar and returns the entered value.
	// The boolean return is false if the prompt was cancelled with Esc.
	Input(s tcell.Screen, prompt string) (string, bool)
}

type statusBar struct{}

// NewStatusBar creates a new status bar instance.
func NewStatusBar() StatusBar {
	return &statusBar{}
}

// Input displays a prompt on the status bar and collects user input. The
// second return value will be false if the user pressed Esc to cancel the
// prompt.
func (sb *statusBar) Input(s tcell.Screen, prompt string) (string, bool) {
	input := []rune{}
	for {
		width, height := s.Size()
		// Clear the status line
		for x := 0; x < width; x++ {
			s.SetContent(x, height-1, ' ', nil, tcell.StyleDefault)
		}
		drawText(s, 0, height-1, width-1, height-1, tcell.StyleDefault.Foreground(tcell.ColorWhite), prompt+string(input))
		s.Show()

		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEnter:
				return string(input), true
			case tcell.KeyEscape:
				return "", false
			case tcell.KeyBackspace, tcell.KeyBackspace2:
				if len(input) > 0 {
					input = input[:len(input)-1]
				}
			case tcell.KeyRune:
				input = append(input, ev.Rune())
			}
		case *tcell.EventResize:
			s.Sync()
		}
	}
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
