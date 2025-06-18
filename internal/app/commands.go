package app

import "github.com/gdamore/tcell/v2"

type CommandExit struct{}

func (c *CommandExit) Name() string { return "exit" }

func (c *CommandExit) Execute(app App, ev *tcell.EventKey) (bool, error) {
	return true, nil
}

type CommandUndo struct{}

func (c *CommandUndo) Name() string { return "undo" }

func (c *CommandUndo) Execute(app App, ev *tcell.EventKey) (bool, error) {
	view := app.GetCurrentView()
	if view != nil {
		view.Undo()
	}
	return false, nil
}

type CommandRedo struct{}

func (c *CommandRedo) Name() string { return "redo" }

func (c *CommandRedo) Execute(app App, ev *tcell.EventKey) (bool, error) {
	view := app.GetCurrentView()
	if view != nil {
		view.Redo()
	}
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
	registerCommand("save", &CommandSave{})
	registerCommand("open", &CommandOpen{})
	registerCommand("up", &CommandMove{dRow: -1})
	registerCommand("down", &CommandMove{dRow: 1})
	registerCommand("left", &CommandMove{dCol: -1})
	registerCommand("right", &CommandMove{dCol: 1})
	registerCommand("backspace", &CommandBackspace{})
	registerCommand("delete", &CommandDelete{})
	registerCommand("pageup", &CommandPageUp{})
	registerCommand("pagedown", &CommandPageDown{})
}
