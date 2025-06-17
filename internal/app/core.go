package app

import (
	"github.com/gdamore/tcell/v2"
)

type App interface {
	OpenFile(filename string) error
	Run(screen tcell.Screen)
	// Settings returns the editor settings instance.
	Settings() Settings
}

type app struct {
	views       []View
	statusBar   StatusBar
	currentView int
	settings    Settings
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
	switch ev.Key() {
	case tcell.KeyEscape:
		return a.handleEscape()
	case tcell.KeyCtrlZ:
		a.handleUndo()
	case tcell.KeyCtrlR:
		a.handleRedo()
	case tcell.KeyCtrlS:
		a.handleSave(screen)
	case tcell.KeyCtrlO:
		a.handleOpen(screen)
	case tcell.KeyUp:
		a.handleUp(screen)
	case tcell.KeyDown:
		a.handleDown(screen)
	case tcell.KeyLeft:
		a.handleLeft(screen)
	case tcell.KeyRight:
		a.handleRight(screen)
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		a.handleBackspace(screen)
	case tcell.KeyDelete:
		a.handleDelete(screen)
	case tcell.KeyPgUp:
		a.handlePageUp(screen)
	case tcell.KeyPgDn:
		a.handlePageDown(screen)
	case tcell.KeyRune:
		a.handleRune(screen, ev.Rune())
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

// handleEscape returns true to signal that the application should exit.
func (a *app) handleEscape() bool {
	return true
}

func (a *app) handleUndo() {
	if len(a.views) > 0 {
		a.views[a.currentView].Undo()
	}
}

func (a *app) handleRedo() {
	if len(a.views) > 0 {
		a.views[a.currentView].Redo()
	}
}

func (a *app) handleSave(screen tcell.Screen) {
	if len(a.views) > 0 {
		if err := a.views[a.currentView].Buffer().Save(); err != nil {
			if a.statusBar != nil {
				a.statusBar.Errorf(screen, "Error saving file: %v", err)
			}
		}
	}
}

func (a *app) handleOpen(screen tcell.Screen) {
	if a.statusBar != nil {
		filename, ok := a.statusBar.Input(screen, "Open file: ")
		if ok && filename != "" {
			if err := a.OpenFile(filename); err != nil {
				if a.statusBar != nil {
					a.statusBar.Errorf(screen, "Error opening file: %v", err)
				}
			}
		}
	}
}

func (a *app) handleClear(screen tcell.Screen) {
	screen.Clear()
}

func (a *app) adjustViewport(screen tcell.Screen) {
	width, height := screen.Size()
	top, left := a.views[a.currentView].TopLeft()
	row, col := a.views[a.currentView].Cursor()

	if row < top {
		top = row
	} else if row >= top+height-1 {
		top = row - height + 2
	}

	if col < left {
		left = col
	} else if col >= left+width-1 {
		left = col - (width - 2)
	}

	a.views[a.currentView].SetTopLeft(top, left)
}

func (a *app) moveCursor(screen tcell.Screen, dRow, dCol int) {
	if len(a.views) == 0 {
		return
	}
	row, col := a.views[a.currentView].Cursor()
	row += dRow
	col += dCol
	row = max(0, row)
	col = max(0, col)
	a.views[a.currentView].SetCursor(row, col)
	a.adjustViewport(screen)
}

func (a *app) handleUp(screen tcell.Screen)    { a.moveCursor(screen, -1, 0) }
func (a *app) handleDown(screen tcell.Screen)  { a.moveCursor(screen, 1, 0) }
func (a *app) handleLeft(screen tcell.Screen)  { a.moveCursor(screen, 0, -1) }
func (a *app) handleRight(screen tcell.Screen) { a.moveCursor(screen, 0, 1) }

func (a *app) handleBackspace(screen tcell.Screen) {
	if len(a.views) == 0 {
		return
	}
	a.views[a.currentView].DeleteRune(false)
	a.adjustViewport(screen)
}

func (a *app) handleDelete(screen tcell.Screen) {
	if len(a.views) == 0 {
		return
	}
	a.views[a.currentView].DeleteRune(true)
	a.adjustViewport(screen)
}

func (a *app) handleRune(screen tcell.Screen, r rune) {
	if len(a.views) == 0 {
		return
	}
	a.views[a.currentView].InsertRune(r)
	a.adjustViewport(screen)
}

func (a *app) handlePageUp(screen tcell.Screen) {
	if len(a.views) == 0 {
		return
	}
	_, height := screen.Size()
	row, col := a.views[a.currentView].Cursor()
	page := height - 1
	a.scrollBy(-page)
	row = max(0, row-page)
	a.views[a.currentView].SetCursor(row, col)
}

func (a *app) handlePageDown(screen tcell.Screen) {
	if len(a.views) == 0 {
		return
	}
	_, height := screen.Size()
	row, col := a.views[a.currentView].Cursor()
	page := height - 1
	a.scrollBy(page)
	row += page
	a.views[a.currentView].SetCursor(row, col)
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
	}, nil
}
