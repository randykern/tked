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
	// LoadSettings loads the settings from the given file.
	LoadSettings(filename string) error
}

type app struct {
	views       []View
	statusBar   StatusBar
	currentView int
	settings    Settings
}

func (a *app) getCurrentView() View {
	if a.currentView < 0 || a.currentView >= len(a.views) || a.views[a.currentView] == nil {
		panic("no active view") // this is a bug not an error!
	}

	return a.views[a.currentView]
}

func (a *app) Settings() Settings { return a.settings }

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

	// Get the initial screen size and update each view to match
	width, height := screen.Size()
	for _, view := range a.views {
		view.Resize(height, width)
	}

	// Draw initial status bar
	a.statusBar.SetScreen(screen) // status bar needs to know the screen to draw on
	a.statusBar.Draw(a.getCurrentView())

	// Draw initial view
	drawView(a.getCurrentView(), screen, a.settings.TabWidth())

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
		drawView(a.getCurrentView(), screen, a.settings.TabWidth())
		a.statusBar.Draw(a.getCurrentView())
	}
}

func (a *app) handleResize(screen tcell.Screen) {
	// TODO: This will have to be smarter about resizing views- not all are full screen
	width, height := screen.Size()
	for _, view := range a.views {
		view.Resize(height, width)
	}

	screen.Sync()
}

func (a *app) handleKey(screen tcell.Screen, ev *tcell.EventKey) bool {
	if ev.Key() == tcell.KeyRune || ev.Key() == tcell.KeyEnter {
		view := a.getCurrentView()
		r := ev.Rune()
		if ev.Key() == tcell.KeyEnter {
			r = '\n'
		}
		view.InsertRune(r)
	} else {
		command := a.settings.KeyBindings().GetCommandForKey(ev.Key(), ev.Modifiers())
		if command != nil {
			ret, err := command.Execute(a, a.getCurrentView(), screen, ev)
			if err != nil {
				a.statusBar.Errorf("Error executing command: %v", err)
			}
			return ret
		}
	}

	return false
}

func (a *app) handleMouse(ev *tcell.EventMouse) {
	x, y := ev.Position()
	view := a.getCurrentView()

	switch ev.Buttons() {
	case tcell.Button1:
		top, left := view.TopLeft()
		view.SetCursor(top+y, left+x)
	case tcell.WheelUp:
		scrollBy(view, -1)
	case tcell.WheelDown:
		scrollBy(view, 1)
	}
}

func (a *app) OpenFile(filename string) error {
	buffer, err := NewBuffer(filename)
	if err != nil {
		return err
	}

	view := NewView(buffer)

	// Resize the view to match the current view's size
	width, height := a.getCurrentView().Size()
	view.Resize(height, width)

	// If the current view is empty, replace it with the new one
	currentBuffer := a.getCurrentView().Buffer()
	if currentBuffer.GetFilename() == "" && !currentBuffer.IsDirty() {
		a.views[a.currentView] = view // replace the current view with the new one
	} else {
		a.views = append(a.views, view)  // add the new view to the end of the list
		a.currentView = len(a.views) - 1 // set the current view to the new one
	}

	return nil
}

func (a *app) GetStatusBar() StatusBar {
	if a.statusBar == nil {
		panic("status bar is nil") // this is a bug not an error!
	}

	return a.statusBar
}

func (a *app) LoadSettings(filename string) error {
	settings, err := NewSettingsFromFile(filename)
	if err != nil {
		return err
	}
	a.settings = settings

	return nil
}

func NewApp() (App, error) {
	registerCommands()

	buffer, err := NewBuffer("")
	if err != nil {
		return nil, err
	}

	return &app{
		views:       []View{NewView(buffer)},
		statusBar:   NewStatusBar(),
		currentView: 0,
		settings:    NewSettings(),
	}, nil
}
