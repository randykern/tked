package app

import "tked/internal/rope"

type View interface {
	// Buffer returns the buffer that the view is displaying.
	Buffer() Buffer
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
	// Undo reverts the last editing action.
	Undo()
	// Redo reapplies an undone editing action.
	Redo()
}

type viewState struct {
	buffer    Buffer
	top       int
	left      int
	cursorRow int
	cursorCol int
}

type view struct {
	states []viewState
	curr   int
	width  int
	height int
}

func (v *view) Buffer() Buffer {
	if len(v.states) == 0 {
		return nil
	}
	return v.states[v.curr].buffer
}

func (v *view) Size() (int, int) {
	return v.height, v.width
}

func (v *view) Resize(rows, cols int) {
	v.height = max(1, rows)
	v.width = max(1, cols)
}

func (v *view) TopLeft() (int, int) {
	if len(v.states) == 0 {
		return 0, 0
	}
	s := v.states[v.curr]
	return s.top, s.left
}

func (v *view) SetTopLeft(top, left int) {
	if len(v.states) == 0 {
		return
	}
	s := &v.states[v.curr]
	s.top = max(0, top)
	s.left = max(0, left)
}

func (v *view) Cursor() (int, int) {
	if len(v.states) == 0 {
		panic("view has no states") // this is a bug not an error
	}
	s := v.states[v.curr]
	return s.cursorRow, s.cursorCol
}

func (v *view) SetCursor(row, col int) {
	if len(v.states) == 0 {
		panic("view has no states") // this is a bug not an error
	}
	s := &v.states[v.curr]
	s.cursorRow = max(0, row)
	s.cursorCol = max(0, col)

	// Adjust the viewport to ensure the cursor is visible.
	ensureCursorVisible(v)
}

func (v *view) InsertRune(r rune) {
	if len(v.states) == 0 {
		return
	}

	curr := v.states[v.curr]
	idx := bufferIndexAt(curr.buffer.Contents(), curr.cursorRow, curr.cursorCol)
	newBuf := curr.buffer.Insert(idx, string(r))

	newRow := curr.cursorRow
	newCol := curr.cursorCol
	if r == '\n' {
		newRow++
		newCol = 0
	} else {
		newCol++
	}

	newState := viewState{
		buffer:    newBuf,
		top:       curr.top,
		left:      curr.left,
		cursorRow: newRow,
		cursorCol: newCol,
	}
	v.states = append([]viewState{newState}, v.states[v.curr:]...)
	v.curr = 0

	// Adjust the viewport to ensure the cursor is visible.
	ensureCursorVisible(v)
}

func (v *view) DeleteRune(forward bool) {
	if len(v.states) == 0 {
		return
	}

	curr := v.states[v.curr]
	idx := bufferIndexAt(curr.buffer.Contents(), curr.cursorRow, curr.cursorCol)

	var start, end int
	var ch byte
	var ok bool
	newRow := curr.cursorRow
	newCol := curr.cursorCol

	if forward {
		_, ok = curr.buffer.Contents().Index(idx)
		if !ok {
			return
		}
		start = idx
		end = idx + 1
	} else {
		if idx == 0 {
			return
		}
		ch, _ = curr.buffer.Contents().Index(idx - 1)
		start = idx - 1
		end = idx
		if ch == '\n' {
			newRow--
			colCount := 0
			scanIdx := start - 1
			for scanIdx >= 0 {
				b, ok := curr.buffer.Contents().Index(scanIdx)
				if !ok {
					break
				}
				if b == '\n' {
					break
				}
				colCount++
				scanIdx--
			}
			newCol = colCount
		} else {
			newCol--
		}
		if newRow < 0 {
			newRow = 0
		}
		if newCol < 0 {
			newCol = 0
		}
	}

	newBuf := curr.buffer.Delete(start, end)

	newState := viewState{
		buffer:    newBuf,
		top:       curr.top,
		left:      curr.left,
		cursorRow: newRow,
		cursorCol: newCol,
	}
	v.states = append([]viewState{newState}, v.states[v.curr:]...)
	v.curr = 0

	// Adjust the viewport to ensure the cursor is visible.
	ensureCursorVisible(v)
}

func (v *view) Undo() {
	if v.curr+1 < len(v.states) {
		v.curr++
	}
}

func (v *view) Redo() {
	if v.curr > 0 {
		v.curr--
	}
}

func NewView(buffer Buffer) View {
	return &view{
		states: []viewState{
			{
				buffer:    buffer,
				top:       0,
				left:      0,
				cursorRow: 0,
				cursorCol: 0,
			},
		},
		curr:   0,
		width:  80,
		height: 24,
	}
}

func bufferIndexAt(r rope.Rope, row, col int) int {
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

func ensureCursorVisible(v *view) {
	s := &v.states[v.curr]
	if s.cursorRow < s.top {
		s.top = s.cursorRow
	} else if s.cursorRow >= s.top+v.height-1 {
		s.top = s.cursorRow - v.height + 2
	}

	if s.cursorCol < s.left {
		s.left = s.cursorCol
	} else if s.cursorCol >= s.left+v.width-1 {
		s.left = s.cursorCol - (v.width - 2)
	}
}
