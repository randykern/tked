package app

import (
	"slices"

	"github.com/gdamore/tcell/v2"

	"tked/internal/tklog"
)

type CommandExit struct{}

func (c *CommandExit) Name() string { return "exit" }

func (c *CommandExit) Execute(app App, ev *tcell.EventKey) (bool, error) {
	for _, view := range app.Views() {
		if view.Buffer().IsDirty() {
			answer, ok := app.GetStatusBar().Input("You have unsaved changes. Quit anyway? (y/n): ")
			if !ok {
				return false, nil
			}
			if answer == "y" {
				return true, nil
			}
			return false, nil
		}
	}
	return true, nil
}

type CommandUndo struct{}

func (c *CommandUndo) Name() string { return "undo" }

func (c *CommandUndo) Execute(app App, ev *tcell.EventKey) (bool, error) {
	view := app.GetCurrentView()
	if view != nil {
		view.Buffer().Undo()
	}
	return false, nil
}

type CommandRedo struct{}

func (c *CommandRedo) Name() string { return "redo" }

func (c *CommandRedo) Execute(app App, ev *tcell.EventKey) (bool, error) {
	view := app.GetCurrentView()
	if view != nil {
		view.Buffer().Redo()
	}
	return false, nil
}

type CommandNewFile struct{}

func (c *CommandNewFile) Name() string { return "newFile" }

func (c *CommandNewFile) Execute(app App, ev *tcell.EventKey) (bool, error) {
	app.OpenFile("")
	return false, nil
}

type CommandSave struct{}

func (c *CommandSave) Name() string { return "save" }

func (c *CommandSave) Execute(app App, ev *tcell.EventKey) (bool, error) {
	view := app.GetCurrentView()
	if view != nil {
		filename := view.Buffer().GetFilename()
		if filename == "" {
			var ok bool
			filename, ok = app.GetStatusBar().Input("Save as: ")
			if !ok {
				return false, nil
			}
		}
		if err := view.Save(filename); err != nil {
			return false, err
		}
	}
	return false, nil
}

type CommandSaveAs struct{}

func (c *CommandSaveAs) Name() string { return "saveAs" }

func (c *CommandSaveAs) Execute(app App, ev *tcell.EventKey) (bool, error) {
	view := app.GetCurrentView()
	if view != nil {
		filename, ok := app.GetStatusBar().Input("Save as: ")
		if !ok {
			return false, nil
		}
		if err := view.Save(filename); err != nil {
			return false, err
		} else {
			view.Buffer().SetFilename(filename)
		}
	}
	return false, nil
}

type CommandOpen struct{}

func (c *CommandOpen) Name() string { return "open" }

func (c *CommandOpen) Execute(app App, ev *tcell.EventKey) (bool, error) {
	filename, ok := app.GetStatusBar().Input("Open file: ")
	if ok && filename != "" {
		if err := app.OpenFile(filename); err != nil {
			app.GetStatusBar().Errorf("Error opening file: %v", err)
		}
	}

	return false, nil
}

type CommandClose struct{}

func (c *CommandClose) Name() string { return "close" }

func (c *CommandClose) Execute(app App, ev *tcell.EventKey) (bool, error) {
	onlyOneView := len(app.Views()) == 1
	closed := app.CloseView(app.GetCurrentView())
	if !closed {
		return false, nil
	}

	if onlyOneView {
		return true, nil
	}

	return false, nil
}

type CommandMove struct {
	dRow int
	dCol int
}

func (c *CommandMove) Name() string {
	switch {
	case c.dRow == -1 && c.dCol == 0:
		return "up"
	case c.dRow == 1 && c.dCol == 0:
		return "down"
	case c.dRow == 0 && c.dCol == -1:
		return "left"
	case c.dRow == 0 && c.dCol == 1:
		return "right"
	}
	return "move"
}

func (c *CommandMove) Execute(app App, ev *tcell.EventKey) (bool, error) {
	view := app.GetCurrentView()
	row, col := view.Cursor()
	row += c.dRow
	col += c.dCol
	row = max(0, row)
	col = max(0, col)
	view.SetCursor(row, col)
	return false, nil
}

type CommandBackspace struct{}

func (c *CommandBackspace) Name() string { return "backspace" }

func (c *CommandBackspace) Execute(app App, ev *tcell.EventKey) (bool, error) {
	view := app.GetCurrentView()
	view.DeleteRune(false)
	return false, nil
}

type CommandDelete struct{}

func (c *CommandDelete) Name() string { return "delete" }

func (c *CommandDelete) Execute(app App, ev *tcell.EventKey) (bool, error) {
	view := app.GetCurrentView()
	view.DeleteRune(true)
	return false, nil
}

type CommandPageUp struct{}

func (c *CommandPageUp) Name() string { return "pageup" }

func (c *CommandPageUp) Execute(app App, ev *tcell.EventKey) (bool, error) {
	view := app.GetCurrentView()
	_, height := view.Size()
	row, col := view.Cursor()
	page := height - 1
	scrollBy(view, -page)
	row = max(0, row-page)
	view.SetCursor(row, col)
	return false, nil
}

type CommandPageDown struct{}

func (c *CommandPageDown) Name() string { return "pagedown" }

func (c *CommandPageDown) Execute(app App, ev *tcell.EventKey) (bool, error) {
	view := app.GetCurrentView()
	_, height := view.Size()
	row, col := view.Cursor()
	page := height - 1
	scrollBy(view, page)
	row += page
	view.SetCursor(row, col)
	return false, nil
}

type CommandNextView struct{}

func (c *CommandNextView) Name() string { return "nextView" }

func (c *CommandNextView) Execute(app App, ev *tcell.EventKey) (bool, error) {
	nextview(app, 1)
	return false, nil
}

type CommandPrevView struct{}

func (c *CommandPrevView) Name() string { return "prevView" }

func (c *CommandPrevView) Execute(app App, ev *tcell.EventKey) (bool, error) {
	nextview(app, -1)
	return false, nil
}

func nextview(app App, direction int) {
	views := app.Views()
	if len(views) > 1 {
		idx := slices.Index(views, app.GetCurrentView())
		if idx == -1 {
			tklog.Panic("current view not found in views") // bug not an error!
		}

		idx += direction
		if idx >= len(views) {
			idx = 0 // wrap around
		} else if idx < 0 {
			idx = len(views) - 1 // wrap around
		}
		app.SetCurrentView(views[idx])
	}
}

func scrollBy(view View, lines int) {
	top, left := view.TopLeft()
	top = max(0, top+lines)
	view.SetTopLeft(top, left)
}

func registerCommands() {
	if len(commands) > 0 {
		return // already registered
	}

	registerCommand("exit", &CommandExit{})
	registerCommand("undo", &CommandUndo{})
	registerCommand("redo", &CommandRedo{})
	registerCommand("new", &CommandNewFile{})
	registerCommand("save", &CommandSave{})
	registerCommand("saveAs", &CommandSaveAs{})
	registerCommand("open", &CommandOpen{})
	registerCommand("close", &CommandClose{})
	registerCommand("up", &CommandMove{dRow: -1})
	registerCommand("down", &CommandMove{dRow: 1})
	registerCommand("left", &CommandMove{dCol: -1})
	registerCommand("right", &CommandMove{dCol: 1})
	registerCommand("backspace", &CommandBackspace{})
	registerCommand("delete", &CommandDelete{})
	registerCommand("pageup", &CommandPageUp{})
	registerCommand("pagedown", &CommandPageDown{})
	registerCommand("nextView", &CommandNextView{})
	registerCommand("prevView", &CommandPrevView{})
}
