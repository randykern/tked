package app

import (
	"os"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
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
	// GetViews returns all the views in the application.
	Views() []View
}

type app struct {
	views       []View
	statusBar   StatusBar
	tabBar      TabBar
	currentView int
	settings    Settings
	lsps        map[string]*lspClient
}

func (a *app) OpenFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	view, err := NewViewFromReader(filename, file)
	if err != nil {
		return err
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

	a.startLSP(filename)

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
	a.GetCurrentView().Draw(screen, a.settings.TabWidth(), 1, 0)

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
		a.GetCurrentView().Draw(screen, a.settings.TabWidth(), 1, 0)
		a.statusBar.Draw(a.GetCurrentView())
	}
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
		panic("status bar is nil") // this is a bug not an error!
	}

	return a.statusBar
}

func (a *app) GetCurrentView() View {
	if a.currentView < 0 || a.currentView >= len(a.views) || a.views[a.currentView] == nil {
		panic("no active view") // this is a bug not an error!
	}

	return a.views[a.currentView]
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
	if ev.Key() == tcell.KeyRune || ev.Key() == tcell.KeyEnter {
		view := a.GetCurrentView()
		r := ev.Rune()
		if ev.Key() == tcell.KeyEnter {
			r = '\n'
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
				a.closeView(idx)
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
		view.SetCursor(top+y-1, left+x)
	case tcell.WheelUp:
		scrollBy(view, -1)
	case tcell.WheelDown:
		scrollBy(view, 1)
	}
}

func (a *app) closeView(idx int) {
	if idx < 0 || idx >= len(a.views) {
		return
	}

	v := a.views[idx]
	if v.Buffer().IsDirty() {
		answer, ok := a.statusBar.Input("Save changes? (y/n): ")
		if !ok {
			return
		}
		if answer == "y" {
			filename := v.Buffer().GetFilename()
			if filename == "" {
				var ok bool
				filename, ok = a.statusBar.Input("Save as: ")
				if !ok {
					return
				}
			}
			if err := v.Save(filename); err != nil {
				a.statusBar.Errorf("Error saving file: %v", err)
				return
			}
		} else if answer != "n" {
			return
		}
	}

	a.views = append(a.views[:idx], a.views[idx+1:]...)
	if len(a.views) == 0 {
		a.views = []View{NewView("", nil)}
		a.currentView = 0
		return
	}
	if a.currentView >= len(a.views) {
		a.currentView = len(a.views) - 1
	} else if idx <= a.currentView && a.currentView > 0 {
		a.currentView--
	}
}

func NewApp() (App, error) {
	registerCommands()

	return &app{
		views:       []View{NewView("", nil)},
		statusBar:   NewStatusBar(),
		tabBar:      NewTabBar(),
		currentView: 0,
		settings:    NewSettings(),
		lsps:        make(map[string]*lspClient),
	}, nil
}

func (a *app) startLSP(filename string) {
	ext := filepath.Ext(filename)

	// TODO: Add support for other languages, and LSP mapping in the editor settings
	if ext == ".go" {
		client, err := newLSPClient("gopls", filename)
		if err != nil {
			return
		}
		if a.lsps == nil {
			panic("lsps is nil") // this is a bug not an error!
		}
		a.lsps[filename] = client
	}
}
