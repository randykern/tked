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

func New() App {
	return &app{
		buffers: []Buffer{},
		views:   []View{},
	}
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

func drawBox(s tcell.Screen, x1, y1, x2, y2 int, style tcell.Style, text string) {
	if y2 < y1 {
		y1, y2 = y2, y1
	}
	if x2 < x1 {
		x1, x2 = x2, x1
	}

	// Fill background
	for row := y1; row <= y2; row++ {
		for col := x1; col <= x2; col++ {
			s.SetContent(col, row, ' ', nil, style)
		}
	}

	// Draw borders
	for col := x1; col <= x2; col++ {
		s.SetContent(col, y1, tcell.RuneHLine, nil, style)
		s.SetContent(col, y2, tcell.RuneHLine, nil, style)
	}
	for row := y1 + 1; row < y2; row++ {
		s.SetContent(x1, row, tcell.RuneVLine, nil, style)
		s.SetContent(x2, row, tcell.RuneVLine, nil, style)
	}

	// Only draw corners if necessary
	if y1 != y2 && x1 != x2 {
		s.SetContent(x1, y1, tcell.RuneULCorner, nil, style)
		s.SetContent(x2, y1, tcell.RuneURCorner, nil, style)
		s.SetContent(x1, y2, tcell.RuneLLCorner, nil, style)
		s.SetContent(x2, y2, tcell.RuneLRCorner, nil, style)
	}

	drawText(s, x1+1, y1+1, x2-1, y2-1, style, text)
}

func drawStatusBar(a *app, s tcell.Screen) {
	filename := "Untitled"
	if len(a.views) > 0 {
		filename = a.views[0].Buffer().GetFilename()
	}

	width, height := s.Size()
	cursorRow, cursorCol := a.views[0].Cursor()
	drawText(s, 0, height-1, width-1, height-1, tcell.StyleDefault.Foreground(tcell.ColorWhite), filename)
	drawText(s, len(filename), height-1, width-1, height-1, tcell.StyleDefault.Foreground(tcell.ColorWhite), fmt.Sprintf(": %d %d", cursorRow, cursorCol))
}

func drawView(v View, s tcell.Screen) {
	// TODO: Erase the viewport?

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
			bufferCol += 4
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
	boxStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorPurple)

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

	// Here's how to get the screen size when you need it.
	// xmax, ymax := s.Size()

	// Here's an example of how to inject a keystroke where it will
	// be picked up by the next PollEvent call.  Note that the
	// queue is LIFO, it has a limited length, and PostEvent() can
	// return an error.
	// s.PostEvent(tcell.NewEventKey(tcell.KeyRune, rune('a'), 0))

	// Event loop
	ox, oy := -1, -1
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
						if row > 0 {
							row--
						}
					case tcell.KeyDown:
						row++
					case tcell.KeyLeft:
						if col > 0 {
							col--
						}
					case tcell.KeyRight:
						col++
					}
					a.views[0].SetCursor(row, col)
				}
			}
		case *tcell.EventMouse:
			x, y := ev.Position()

			switch ev.Buttons() {
			case tcell.Button1, tcell.Button2:
				if ox < 0 {
					ox, oy = x, y // record location when click started
				}

			case tcell.ButtonNone:
				if ox >= 0 {
					label := fmt.Sprintf("%d,%d to %d,%d", ox, oy, x, y)
					drawBox(screen, ox, oy, x, y, boxStyle, label)
					ox, oy = -1, -1
				}
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

	return nil
}
