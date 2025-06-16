package app

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

func (v *view) SetTopLeft(top, left int) {
	v.top = max(0, top)
	v.left = max(0, left)
}

func (v *view) Cursor() (int, int) {
	return v.cursorRow, v.cursorCol
}

func (v *view) SetCursor(row, col int) {
	v.cursorRow = max(0, row)
	v.cursorCol = max(0, col)
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
