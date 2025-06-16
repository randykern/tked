package app

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

type App interface {
	OpenFile(filename string) error
	Run(screen tcell.Screen)
}

type app struct {
	buffers []Buffer
	views   []View
}

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

func drawStatusBar(a *app, s tcell.Screen) {
	filename := "Untitled"
	cursor := ": 1 1"
	if len(a.views) > 0 {
		filename = a.views[0].Buffer().GetFilename()
		if filename == "" {
			filename = "Untitled"
		}

		cursorRow, cursorCol := a.views[0].Cursor()
		cursor = fmt.Sprintf(": %d %d", cursorRow+1, cursorCol+1)
	}

	width, height := s.Size()
	drawText(s, 0, height-1, width-1, height-1, tcell.StyleDefault.Foreground(tcell.ColorWhite), filename)
	drawText(s, len(filename), height-1, width-1, height-1, tcell.StyleDefault.Foreground(tcell.ColorWhite), cursor)
}

func drawView(v View, s tcell.Screen) {
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
			// TODO: Tab width option
			// TODO: This doesn't work quite right. See line 8 of the sample text file.
			tabWidth := 4
			if bufferCol%tabWidth == 0 {
				bufferCol += tabWidth
			} else {
				bufferCol += tabWidth - bufferCol%4
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
	drawStatusBar(a, screen)

	// Draw initial buffer
	if len(a.views) > 0 {
		drawView(a.views[0], screen)
	}

	// Event loop
	for {
		// Update screen
		screen.Show()

		// Poll event
		ev := screen.PollEvent()

		// Process event
		switch ev := ev.(type) {
		case *tcell.EventResize:
			screen.Sync()
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
				return
			} else if ev.Key() == tcell.KeyCtrlL {
				screen.Sync()
			} else if ev.Rune() == 'C' || ev.Rune() == 'c' {
				screen.Clear()
			} else if ev.Key() == tcell.KeyUp || ev.Key() == tcell.KeyDown || ev.Key() == tcell.KeyLeft || ev.Key() == tcell.KeyRight {
				if len(a.views) > 0 {
					row, col := a.views[0].Cursor()
					switch ev.Key() {
					case tcell.KeyUp:
						row--
					case tcell.KeyDown:
						row++
					case tcell.KeyLeft:
						col--
					case tcell.KeyRight:
						col++
					}

					// Make sure we don't move the cursor before the start of the file
					row = max(0, row)
					col = max(0, col)

					// TODO: Moving the cursor past the end?

					// Adjust viewport if cursor moved outside visible area
					width, height := screen.Size()
					top, left := a.views[0].TopLeft()

					if row < top {
						// Moved out the top of the viewport- make the new top row the cursor row
						top = row
					} else if row >= top+height-1 {
						// Moved out the bottom of the viewport- make the new bottom row the cursor row
						top = row - height + 2
					}

					if col < left {
						left = col
					} else if col >= left+width-1 {
						left = col - (width - 2)
					}

					a.views[0].SetTopLeft(top, left)
					a.views[0].SetCursor(row, col)
				}
			}
		case *tcell.EventMouse:
			x, y := ev.Position()

			switch ev.Buttons() {
			case tcell.Button1:
				top, left := a.views[0].TopLeft()
				a.views[0].SetCursor(top+y, left+x)
			}
		}

		screen.Clear()
		drawView(a.views[0], screen)
		drawStatusBar(a, screen)
	}
}

func (a *app) OpenFile(filename string) error {
	buffer, err := NewBuffer(filename)
	if err != nil {
		return err
	}

	a.buffers = append(a.buffers, buffer)

	view := NewView(buffer)
	a.views = append(a.views, view)

	if a.buffers[0].GetFilename() == "" && !a.buffers[0].IsDirty() {
		a.buffers = a.buffers[1:]
		a.views = a.views[1:]
	}

	return nil
}

func New() (App, error) {
	buffer, err := NewBuffer("")
	if err != nil {
		return nil, err
	}

	return &app{
		buffers: []Buffer{buffer},
		views:   []View{NewView(buffer)},
	}, nil
}
