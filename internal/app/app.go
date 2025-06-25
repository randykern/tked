package app

import (
	"os"
	"slices"

	"github.com/gdamore/tcell/v2"

	"tked/internal/lsp"
	"tked/internal/tklog"
)

type App interface {
	// OpenFile opens a file and adds a new view for it.
	OpenFile(filename string) error
	// Run starts the application and enters the event loop.
	Run(screen tcell.Screen)
	// Settings returns the editor settings instance.
	Settings() Settings
	// LoadSettings loads the settings from the given file.
	LoadSettings(filename string) error
	// GetStatusBar returns the status bar instance.
	GetStatusBar() StatusBar
	// GetCurrentView returns the current view.
	GetCurrentView() View
	// SetCurrentView sets the current view.
	SetCurrentView(view View)
	// GetViews returns all the views in the application.
	Views() []View
	// CloseView closes the given view. Returns true if the view was closed,
	// false if the user cancelled the close.
	CloseView(view View) bool
}

type app struct {
	views       []View
	statusBar   StatusBar
	tabBar      TabBar
	currentView int
	settings    Settings
}

func (a *app) OpenFile(filename string) error {
	var view View
	if filename == "" {
		view = NewView("", nil)
	} else {
		file, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer file.Close()

		view, err = NewViewFromReader(filename, file)
		if err != nil {
			return err
		}
	}

	// Resize the view to match the current view's size
	width, height := a.GetCurrentView().Size()
	view.Resize(height, width)

	// If the current view is empty, replace it with the new one
	currentView := a.GetCurrentView()
	if currentView.Buffer().GetFilename() == "" && !currentView.Buffer().IsDirty() {
		a.views[a.currentView] = view // replace the current view with the new one
	} else {
		a.views = append(a.views, view)  // add the new view to the end of the list
		a.currentView = len(a.views) - 1 // set the current view to the new one
	}

	return nil
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
		view.Resize(height-1, width)
	}

	// Draw initial status bar
	a.statusBar.SetScreen(screen) // status bar needs to know the screen to draw on
	a.statusBar.Draw(a.GetCurrentView())

	// Draw initial tab bar
	if a.tabBar != nil {
		a.tabBar.SetScreen(screen)
		a.tabBar.Draw(a.views, a.currentView)
	}

	// Draw initial view
	a.GetCurrentView().Draw(screen, 1, 0)

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
			if a.handleKey(ev) {
				break eventLoop
			}
		case *tcell.EventMouse:
			a.handleMouse(ev)
		}

		screen.Clear()
		if a.tabBar != nil {
			a.tabBar.Draw(a.views, a.currentView)
		}
		a.GetCurrentView().Draw(screen, 1, 0)
		a.statusBar.Draw(a.GetCurrentView())
	}

	for _, view := range a.views {
		view.Buffer().Close()
	}

	lsp.ShutdownAll()
}

func (a *app) Settings() Settings { return a.settings }

func (a *app) LoadSettings(filename string) error {
	settings, err := NewSettingsFromFile(filename)
	if err != nil {
		return err
	}
	a.settings = settings

	return nil
}

func (a *app) GetStatusBar() StatusBar {
	if a.statusBar == nil {
		tklog.Panic("status bar is nil") // this is a bug not an error!
	}

	return a.statusBar
}

func (a *app) GetCurrentView() View {
	if a.currentView < 0 || a.currentView >= len(a.views) || a.views[a.currentView] == nil {
		tklog.Panic("no active view") // this is a bug not an error!
	}

	return a.views[a.currentView]
}

func (a *app) SetCurrentView(view View) {
	a.currentView = slices.Index(a.views, view)
	if a.currentView == -1 {
		a.currentView = 0
	}
}

func (a *app) Views() []View { return a.views }

func (a *app) handleResize(screen tcell.Screen) {
	// TODO: This will have to be smarter about resizing views- not all are full screen
	width, height := screen.Size()
	for _, view := range a.views {
		view.Resize(height-1, width)
	}

	if a.tabBar != nil {
		a.tabBar.Draw(a.views, a.currentView)
	}

	screen.Sync()
}

func (a *app) handleKey(ev *tcell.EventKey) bool {
	if ev.Key() == tcell.KeyRune || ev.Key() == tcell.KeyEnter || ev.Key() == tcell.KeyTab {
		view := a.GetCurrentView()
		r := ev.Rune()
		if ev.Key() == tcell.KeyEnter {
			r = '\n'
		} else if ev.Key() == tcell.KeyTab {
			r = '\t'
		}
		view.InsertRune(r)
	} else {
		command := a.settings.KeyBindings().GetCommandForKey(ev.Key(), ev.Modifiers())
		if command != nil {
			ret, err := command.Execute(a, ev)
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
	view := a.GetCurrentView()

	switch ev.Buttons() {
	case tcell.Button1:
		if a.tabBar != nil {
			if idx, ok := a.tabBar.CloseIndexAt(x, y); ok {
				a.CloseView(a.views[idx])
				return
			}
			if idx, ok := a.tabBar.ViewIndexAt(x, y); ok {
				if idx >= 0 && idx < len(a.views) {
					a.currentView = idx
					return
				}
			}
		}

		top, left := view.TopLeft()
		oldRow, oldCol := view.Cursor()

		if ev.Modifiers()&tcell.ModShift != 0 {
			aRow, aCol, ok := view.Anchor()
			if !ok {
				view.SetAnchor(oldRow, oldCol)
				aRow, aCol = oldRow, oldCol
			}
			view.SetCursor(top+y-1, left+x)
			row, col := view.Cursor()

			// Use the same selection logic as keyboard: always include character under anchor
			startRow, startCol, endRow, endCol := aRow, aCol, row, col
			if (aRow > row) || (aRow == row && aCol > col) {
				// Anchor is after cursor, so extend anchor by +1 col
				startRow, startCol = row, col
				endRow, endCol = aRow, aCol+1
			}
			sel := orderedSelection(startRow, startCol, endRow, endCol)
			view.SetSelections([]Selection{sel})
		} else {
			view.SetCursor(top+y-1, left+x)
			view.ClearAnchor()
			view.SetSelections(nil)
		}
	case tcell.WheelUp:
		scrollBy(view, -1)
	case tcell.WheelDown:
		scrollBy(view, 1)
	}
}

func (a *app) CloseView(v View) bool {
	idx := slices.Index(a.views, v)
	if idx == -1 {
		tklog.Panic("view not found") // this is a bug not an error!
	}

	if v.Buffer().IsDirty() {
		answer, ok := a.GetStatusBar().Input("Save changes? (y/n): ")
		if !ok {
			return false
		}
		if answer == "y" {
			filename := v.Buffer().GetFilename()
			if filename == "" {
				var ok bool
				filename, ok = a.GetStatusBar().Input("Save as: ")
				if !ok {
					return false
				}
			}
			if err := v.Save(filename); err != nil {
				a.GetStatusBar().Errorf("Error saving file: %v", err)
				return false
			}
		} else if answer != "n" {
			return false
		}
	}

	v.Buffer().Close()

	// remove the view from the list
	a.views = append(a.views[:idx], a.views[idx+1:]...)
	if len(a.views) == 0 {
		// Ensure we have at least one view open
		a.views = []View{NewView("", nil)}
		a.currentView = 0
	} else if a.currentView >= len(a.views) {
		a.currentView = len(a.views) - 1
	} else if idx <= a.currentView && a.currentView > 0 {
		a.currentView--
	}

	return true
}

var theApp App

func NewApp() (App, error) {
	if theApp != nil {
		tklog.Panic("app already created") // this is a bug not an error!
	}

	registerCommands()

	appObject := &app{
		views:       []View{},
		statusBar:   NewStatusBar(),
		tabBar:      NewTabBar(),
		currentView: 0,
		settings:    NewSettings(),
	}
	theApp = appObject
	appObject.views = []View{NewView("", nil)}

	return theApp, nil
}

func GetApp() App {
	if theApp == nil {
		tklog.Panic("app not created") // this is a bug not an error!
	}

	return theApp
}

func ResetApp() {
	theApp = nil
}
