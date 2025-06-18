package app

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

// StatusBar describes the behaviour of a status bar component.
type StatusBar interface {
	// SetScreen sets the screen that the status bar will draw on.
	SetScreen(s tcell.Screen)
	// Draw renders the status bar for the provided view.
	Draw(v View)
	// Message displays a message on the status bar.
	Message(msg string)
	// Messagef formats the message and displays it on the status bar.
	Messagef(format string, args ...interface{})
	// Error displays an error message on the status bar until a key is pressed.
	Error(msg string)
	// Errorf formats the error message and displays it until a key is pressed.
	Errorf(format string, args ...interface{})
	// Input displays a prompt on the status bar and returns the entered value.
	// The boolean return is false if the prompt was cancelled with Esc.
	Input(prompt string) (string, bool)
}

type statusBar struct {
	screen tcell.Screen
}

// SetScreen sets the screen that the status bar will draw on.
func (sb *statusBar) SetScreen(s tcell.Screen) {
	if s == nil {
		panic("screen is nil")
	}
	sb.screen = s
}

// Draw renders the current status bar.
func (sb *statusBar) Draw(v View) {
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

	width, height := sb.screen.Size()
	sb.drawText(0, height-1, width-1, height-1, tcell.StyleDefault.Foreground(tcell.ColorWhite), filename+dirty)
	sb.drawText(len(filename)+len(dirty), height-1, width-1, height-1, tcell.StyleDefault.Foreground(tcell.ColorWhite), cursor)
}

// Message displays a message on the status bar.
func (sb *statusBar) Message(msg string) {
	sb.drawPrompt(msg, tcell.StyleDefault.Foreground(tcell.ColorWhite))
}

// Messagef formats the message and displays it on the status bar.
func (sb *statusBar) Messagef(format string, args ...interface{}) {
	sb.Message(fmt.Sprintf(format, args...))
}

// Error displays an error message on the status bar until a key is pressed.
func (sb *statusBar) Error(msg string) {
	sb.drawPrompt(msg, tcell.StyleDefault.Foreground(tcell.ColorRed))
}

// Errorf formats the error message and displays it until a key is pressed.
func (sb *statusBar) Errorf(format string, args ...interface{}) {
	sb.Error(fmt.Sprintf(format, args...))
}

// Input displays a prompt on the status bar and collects user input. The
// second return value will be false if the user pressed Esc to cancel the
// prompt.
func (sb *statusBar) Input(prompt string) (string, bool) {
	input := []rune{}
	for {
		width, height := sb.screen.Size()
		// Clear the status line
		for x := 0; x < width; x++ {
			sb.screen.SetContent(x, height-1, ' ', nil, tcell.StyleDefault)
		}
		sb.drawText(0, height-1, width-1, height-1, tcell.StyleDefault.Foreground(tcell.ColorWhite), prompt+string(input))
		sb.screen.Show()

		ev := sb.screen.PollEvent()
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
			sb.screen.Sync()
		}
	}
}

func (sb *statusBar) drawPrompt(msg string, style tcell.Style) {
	for {
		width, height := sb.screen.Size()
		sb.drawText(0, height-1, width-1, height-1, style, msg)
		sb.screen.Show()

		// TODO: this is a hack to get the message to display for a short time
		// and then disappear. It should be replaced with a more robust solution.
		ev := sb.screen.PollEvent()
		switch ev.(type) {
		case *tcell.EventKey, *tcell.EventMouse:
			return
		case *tcell.EventResize:
			sb.screen.Sync()
		}
	}
}

func (sb *statusBar) drawText(x1, y1, x2, y2 int, style tcell.Style, text string) {
	row := y1
	col := x1
	for _, r := range text {
		sb.screen.SetContent(col, row, r, nil, style)
		col++
		if col >= x2 {
			break
		}
	}
}

// NewStatusBar creates a new status bar instance.
func NewStatusBar() StatusBar {
	return &statusBar{
		screen: nil,
	}
}
