package app

import (
	"io"
	"os"
	"path/filepath"

	"github.com/gdamore/tcell/v2"

	"tked/internal/rope"
)

type View interface {
	// Returns the buffer in this view.
	Buffer() Buffer

	// TODO: It might be nice to remove this. It is only used for drawing
	// and scrolling to ensure the cursor is visible.
	// Size returns the number of rows and columns visible in the view.
	Size() (int, int)
	// Resize updates the number of rows and columns visible in the view.
	Resize(rows, cols int)

	// TopLeft returns the top row and left column offsets.
	TopLeft() (int, int)
	// SetTopLeft updates the view's top row and left column offsets.
	SetTopLeft(top, left int)

	// Cursor returns the current cursor position as row and column indexes.
	Cursor() (int, int)
	// SetCursor updates the current cursor position and moves the viewport
	// to ensure the cursor is visible.
	SetCursor(row, col int)

	// InsertRune inserts a rune into the buffer at the cursor position.
	InsertRune(r rune)
	// DeleteRune deletes a rune. When forward is true it deletes the rune
	// under the cursor (Delete key behaviour). Otherwise it deletes the
	// rune before the cursor (Backspace behaviour).
	DeleteRune(forward bool)

	// Draw renders the view's contents on the provided screen.
	// topOffset and leftOffset specify where to start drawing on the screen.
	Draw(screen tcell.Screen, tabWidth, topOffset, leftOffset int)

	// Save writes the buffer contents to disk using the filename. If fileanme
	// is empty, save uses the existing filename if set, otherwise it returns an error.
	Save(filename string) error
}

type view struct {
	buffer Buffer
	width  int
	height int
	top    int
	left   int
}

type cursor struct {
	row int
	col int
}

func (v *view) Buffer() Buffer {
	return v.buffer
}

func (v *view) Size() (int, int) {
	return v.height, v.width
}

func (v *view) Resize(rows, cols int) {
	v.height = max(1, rows)
	v.width = max(1, cols)
}

func (v *view) TopLeft() (int, int) {
	return v.top, v.left
}

func (v *view) SetTopLeft(top, left int) {
	v.top = max(0, top)
	v.left = max(0, left)
}

func (v *view) Cursor() (int, int) {
	c := v.buffer.GetProperty(cursorProp).(*cursor)
	return c.row, c.col
}

func (v *view) SetCursor(row, col int) {
	c := &cursor{
		row: max(0, row),
		col: max(0, col),
	}
	v.buffer.SetProperty(cursorProp, c)

	// Adjust the viewport to ensure the cursor is visible.
	v.ensureCursorVisible()
}

func (v *view) InsertRune(r rune) {
	cursorRow, cursorCol := v.Cursor()
	idx := indexForRowCol(v.buffer.Contents(), cursorRow, cursorCol)
	v.buffer.Insert(idx, string(r))

	if r == '\n' {
		cursorRow++
		cursorCol = 0
	} else {
		cursorCol++
	}
	v.SetCursor(cursorRow, cursorCol)
}

func (v *view) DeleteRune(forward bool) {
	cursorRow, cursorCol := v.Cursor()
	idx := indexForRowCol(v.buffer.Contents(), cursorRow, cursorCol)

	if forward {
		v.buffer.Delete(idx, idx+1)
		// Cursor doesn't move in this case
	} else {
		// Cursor moves back one character, handling the case where it was at the start of a line
		cursorCol--
		if cursorCol < 0 {
			cursorRow--
			if cursorRow < 0 {
				cursorRow = 0
			} else {
				// Set cursorCol to the end of the previous line

				// Start at the beginning of the previous line
				cursorCol = 0

				// Scan to the end of the line
				r := v.buffer.Contents()
				lineLen := 0
				for idxStartOfLine := indexForRowCol(r, cursorRow, cursorCol); ; lineLen++ {
					b, ok := r.Index(idxStartOfLine + lineLen)
					if !ok || b == '\n' {
						break
					}
				}
				cursorCol = lineLen
			}
		}

		// Actaully delete the character now- we don't do it before
		// because we need to know the length of the previous line
		v.buffer.Delete(idx-1, idx)
		v.SetCursor(cursorRow, cursorCol)
	}
}

func (v *view) Draw(screen tcell.Screen, tabWidth, topOffset, leftOffset int) {
	screenWidth, screenHeight := screen.Size()

	index := 0

	viewTop, viewLeft := v.TopLeft()

	bufferRow := 0
	bufferCol := 0

	// TODO: MB characeter sets

	for {
		r, ok := v.buffer.Contents().Index(index)
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
			screen.SetContent(leftOffset+bufferCol-viewLeft, topOffset+bufferRow-viewTop, rune(r), nil, tcell.StyleDefault)
		}

		bufferCol++
	}

	cursorRow, cursorCol := v.Cursor()
	if cursorRow >= viewTop && cursorRow < viewTop+screenHeight-1 && cursorCol >= viewLeft && cursorCol < viewLeft+screenWidth-1 {
		screen.ShowCursor(leftOffset+cursorCol-viewLeft, topOffset+cursorRow-viewTop)
	} else {
		screen.HideCursor()
	}
}

func (v *view) Save(filename string) error {
	if filename == "" {
		filename = v.buffer.GetFilename()
	}

	if filename == "" {
		return os.ErrInvalid
	}

	dir, name := filepath.Split(filename)

	// Create a temporary file in the same directory so that os.Rename works
	// across filesystems.
	tmp, err := os.CreateTemp(dir, name+".tmp*")
	if err != nil {
		return err
	}

	// Write contents to the temporary file first.
	if _, err := v.buffer.Write(tmp); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}

	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return err
	}

	// Atomically replace the target file.
	if err := os.Rename(tmp.Name(), filename); err != nil {
		os.Remove(tmp.Name())
		return err
	}

	return nil
}

func (v *view) ensureCursorVisible() {
	cursorRow, cursorCol := v.Cursor()
	if cursorRow < v.top {
		v.top = cursorRow
	} else if cursorRow >= v.top+v.height-1 {
		v.top = cursorRow - v.height + 2
	}

	if cursorCol < v.left {
		v.left = cursorCol
	} else if cursorCol >= v.left+v.width-1 {
		v.left = cursorCol - (v.width - 2)
	}
}

func (v *view) onBufferChange(buffer Buffer, start, end int, context any) {
	v.ensureCursorVisible()
}

// Create a new view with the given filename and contents. If contents is nil,
// an empty rope is used. The empty filename is used for unnamed views.
func NewView(filename string, contents rope.Rope) View {
	registerViewProperties()

	if contents == nil {
		contents = rope.NewRope("")
	}

	v := &view{
		buffer: NewBuffer(filename, contents),
		width:  80,
		height: 24,
		top:    0,
		left:   0,
	}
	v.SetCursor(0, 0)
	v.buffer.OnChange(v.onBufferChange, v)
	return v
}

// Create a new view with the given filename and contents read from the reader.
func NewViewFromReader(filename string, r io.Reader) (View, error) {
	contents, err := rope.NewFromReader(r)
	if err != nil {
		return nil, err
	}
	return NewView(filename, contents), nil
}

func indexForRowCol(r rope.Rope, row, col int) int {
	// TODO: This is very slow. We should use a more efficient algorithm.
	idx := 0
	currRow := 0
	currCol := 0
	for {
		if currRow == row && currCol == col {
			return idx
		}
		b, ok := r.Index(idx)
		if !ok {
			return idx
		}
		if b == '\n' {
			currRow++
			currCol = 0
		} else {
			currCol++
		}
		idx++
	}
}

var cursorProp PropKey

func registerViewProperties() {
	cursorProp = RegisterBufferProperty()
}
