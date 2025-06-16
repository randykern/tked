package app

import "tked/internal/rope"

type View interface {
	// Buffer returns the buffer that the view is displaying.
	Buffer() Buffer
	// TopLeft returns the top row and left column offsets.
	TopLeft() (int, int)
	// SetTopLeft updates the view's top row and left column offsets.
	SetTopLeft(top, left int)
	// Cursor returns the current cursor position as row and column indexes.
	Cursor() (int, int)
	// SetCursor updates the current cursor position.
	SetCursor(row, col int)
	// InsertRune inserts a rune into the buffer at the cursor position.
	InsertRune(r rune)
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
}

func (v *view) Buffer() Buffer {
	if len(v.states) == 0 {
		return nil
	}
	return v.states[0].buffer
}

func (v *view) TopLeft() (int, int) {
	if len(v.states) == 0 {
		return 0, 0
	}
	s := v.states[0]
	return s.top, s.left
}

func (v *view) SetTopLeft(top, left int) {
	if len(v.states) == 0 {
		return
	}
	s := &v.states[0]
	s.top = max(0, top)
	s.left = max(0, left)
}

func (v *view) Cursor() (int, int) {
	if len(v.states) == 0 {
		return 0, 0
	}
	s := v.states[0]
	return s.cursorRow, s.cursorCol
}

func (v *view) SetCursor(row, col int) {
	if len(v.states) == 0 {
		return
	}
	s := &v.states[0]
	s.cursorRow = max(0, row)
	s.cursorCol = max(0, col)
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

func (v *view) InsertRune(r rune) {
	if len(v.states) == 0 {
		return
	}

	curr := v.states[0]
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
	v.states = append([]viewState{newState}, v.states...)
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
	}
}
