package app

type View interface {
	Buffer() Buffer
	TopLeft() (int, int)
}

type view struct {
	buffer Buffer
	top    int
	left   int
}

func (v *view) Buffer() Buffer {
	return v.buffer
}

func (v *view) TopLeft() (int, int) {
	return v.top, v.left
}

func NewView(buffer Buffer) View {
	return &view{
		buffer: buffer,
		top:    0,
		left:   0,
	}
}
