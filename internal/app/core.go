package app

import (
	"github.com/gdamore/tcell/v2"
)

type App interface {
	// OpenFile opens a file and adds a new view for it.
	OpenFile(filename string) error
	// Run starts the application and enters the event loop.
	Run(screen tcell.Screen)
	// Settings returns the editor settings instance.
	Settings() Settings
	// GetStatusBar returns the status bar instance.
	GetStatusBar() StatusBar
}

type app struct {
	views       []View
	statusBar   StatusBar
	currentView int
	settings    Settings
	keyBindings KeyBindings
}

func (a *app) scrollBy(lines int) {
	if len(a.views) == 0 {
		return
	}

	top, left := a.views[a.currentView].TopLeft()
	a.views[a.currentView].SetTopLeft(max(0, top+lines), left)
}

func (a *app) Settings() Settings { return a.settings }

func drawText(s tcell.Screen, x1, y1, x2, y2 int, style tcell.Style, text string) {
	row := y1
	col := x1
	for _, r := range text {
		s.SetContent(col, row, r, nil, style)
		col++
		if col >= x2 {
			row++
			col = x1
		}
		if row > y2 {
			break
		}
	}
}

func drawView(v View, s tcell.Screen, tabWidth int) {
	screenWidth, screenHeight := s.Size()

	index := 0

	viewTop, viewLeft := v.TopLeft()

	bufferRow := 0
	bufferCol := 0

	// TODO: MB characeter sets

	for {
		r, ok := v.Buffer().Contents().Index(index)
		if !ok {
			break
		}
		index++

		// Instead of special logic for text that shouldn't be drawn, we will always just move forward
		// a character at a time, but suppress the drawing of the character if it is outside
		// the viewport.

		if r == '\n' {
			bufferRow++
			bufferCol = 0
			continue
		} else if r == '\t' {
			if tabWidth <= 0 {
				tabWidth = 4
			}
			if bufferCol%tabWidth == 0 {
				bufferCol += tabWidth
			} else {
				bufferCol += tabWidth - bufferCol%tabWidth
			}
			continue
		}

		if bufferRow >= viewTop && bufferCol >= viewLeft && bufferRow < viewTop+screenHeight-1 && bufferCol < viewLeft+screenWidth-1 {
			s.SetContent(bufferCol-viewLeft, bufferRow-viewTop, rune(r), nil, tcell.StyleDefault)
		}

		bufferCol++
	}

	cursorRow, cursorCol := v.Cursor()
	if cursorRow >= viewTop && cursorRow < viewTop+screenHeight-1 && cursorCol >= viewLeft && cursorCol < viewLeft+screenWidth-1 {
		s.ShowCursor(cursorCol-viewLeft, cursorRow-viewTop)
	} else {
		s.HideCursor()
	}
}

func (a *app) Run(screen tcell.Screen) {
	defStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)

	// Initialize screen
	screen.SetStyle(defStyle)
	screen.EnableMouse()
	screen.EnablePaste()
	screen.Clear()

	// Draw initial status bar
	if a.statusBar != nil {
		var view View
		if len(a.views) > 0 {
			view = a.views[a.currentView]
		}
		a.statusBar.Draw(screen, view)
	}

	// Draw initial view
	if len(a.views) > 0 {
		drawView(a.views[a.currentView], screen, a.settings.TabWidth())
	}

	// Event loop
eventLoop:
	for {
		// Update screen
		screen.Show()

		// Poll event
		ev := screen.PollEvent()

		// Process event
		switch ev := ev.(type) {
		case *tcell.EventResize:
			a.handleResize(screen)
		case *tcell.EventKey:
			if a.handleKey(screen, ev) {
				break eventLoop
			}
		case *tcell.EventMouse:
			a.handleMouse(ev)
		}

		screen.Clear()
		drawView(a.views[a.currentView], screen, a.settings.TabWidth())
		if a.statusBar != nil {
			a.statusBar.Draw(screen, a.views[a.currentView])
		}
	}
}

func (a *app) handleResize(screen tcell.Screen) {
	screen.Sync()
}

func (a *app) handleKey(screen tcell.Screen, ev *tcell.EventKey) bool {
	if ev.Key() == tcell.KeyRune {
		if a.views[a.currentView] != nil {
			a.views[a.currentView].InsertRune(ev.Rune())
			adjustViewport(a.views[a.currentView], screen)
		}
	} else {
		command := a.keyBindings.GetCommandForKey(ev.Key(), ev.Modifiers())
		if command != nil {
			ret, err := command.Execute(a, a.views[a.currentView], screen, ev)
			if err != nil {
				if a.statusBar != nil {
					a.statusBar.Errorf(screen, "Error executing command: %v", err)
				}
			}
			return ret
		}
	}

	return false
}

func (a *app) handleMouse(ev *tcell.EventMouse) {
	x, y := ev.Position()

	switch ev.Buttons() {
	case tcell.Button1:
		top, left := a.views[a.currentView].TopLeft()
		a.views[a.currentView].SetCursor(top+y, left+x)
	case tcell.WheelUp:
		a.scrollBy(-1)
	case tcell.WheelDown:
		a.scrollBy(1)
	}
}

func (a *app) OpenFile(filename string) error {
	buffer, err := NewBuffer(filename)
	if err != nil {
		return err
	}

	view := NewView(buffer)

	// If the current view is empty, replace it with the new one
	if a.views[a.currentView].Buffer().GetFilename() == "" && !a.views[a.currentView].Buffer().IsDirty() {
		a.views[a.currentView] = view
	} else {
		a.views = append(a.views, view)
		a.currentView = len(a.views) - 1
	}

	return nil
}

func (a *app) GetStatusBar() StatusBar {
	return a.statusBar
}

func NewApp() (App, error) {
	buffer, err := NewBuffer("")
	if err != nil {
		return nil, err
	}

	return &app{
		views:       []View{NewView(buffer)},
		statusBar:   NewStatusBar(),
		currentView: 0,
		settings:    NewSettings(),
		keyBindings: DefaultKeyBindings(),
	}, nil
}
