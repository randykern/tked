package app

import (
	"os"
	"tked/internal/rope"
)

type Buffer interface {
	GetFilename() string
	Contents() rope.Rope
	IsDirty() bool
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
