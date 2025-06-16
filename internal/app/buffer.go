package app

import (
	"os"
	"tked/internal/rope"
)

type Buffer interface {
	GetFilename() string
	Contents() rope.Rope
	IsDirty() bool
	// Insert returns a new Buffer with text inserted at the specified index.
	Insert(idx int, text string) Buffer
}

type buffer struct {
	filename string
	contents rope.Rope
	dirty    bool
}

func (b *buffer) GetFilename() string {
	return b.filename
}

func (b *buffer) Contents() rope.Rope {
	return b.contents
}

func (b *buffer) IsDirty() bool {
	return b.dirty
}

func (b *buffer) Insert(idx int, text string) Buffer {
	if b == nil {
		return nil
	}
	if idx < 0 {
		idx = 0
	}
	if idx > b.contents.Len() {
		idx = b.contents.Len()
	}
	nb := &buffer{
		filename: b.filename,
		contents: b.contents.Insert(idx, text),
		dirty:    true,
	}
	return nb
}

func NewBuffer(filename string) (Buffer, error) {
	contents := rope.NewRope("")

	if filename != "" {
		file, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		contents, err = rope.NewFromReader(file)
		if err != nil {
			return nil, err
		}
	}

	return &buffer{
		filename: filename,
		contents: contents,
		dirty:    false,
	}, nil
}
