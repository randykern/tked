package app

type View interface {
	Buffer() Buffer
	TopLeft() (int, int)
	// Cursor returns the current cursor position as row and column indexes.
	Cursor() (int, int)
	// SetCursor updates the current cursor position.
	SetCursor(row, col int)
}

type view struct {
	buffer    Buffer
	top       int
	left      int
	cursorRow int
	cursorCol int
}

func (v *view) Buffer() Buffer {
	return v.buffer
}

func (v *view) TopLeft() (int, int) {
	return v.top, v.left
}

func (v *view) Cursor() (int, int) {
	return v.cursorRow, v.cursorCol
}

func (v *view) SetCursor(row, col int) {
	v.cursorRow = row
	v.cursorCol = col
}

func NewView(buffer Buffer) View {
	return &view{
		buffer:    buffer,
		top:       0,
		left:      0,
		cursorRow: 0,
		cursorCol: 0,
	}
}
