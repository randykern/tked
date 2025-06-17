package app

import (
	"os"
	"path/filepath"

	"tked/internal/rope"
)

type Buffer interface {
	// Returns the filename of the buffer.
	GetFilename() string
	// Returns the contents of the buffer.
	Contents() rope.Rope
	// Returns true if the buffer has been modified since it was last saved.
	IsDirty() bool
	// Returns a new Buffer with text inserted at the specified index.
	// Insert returns a new Buffer with text inserted at the specified index.
	Insert(idx int, text string) Buffer
	// Delete returns a new Buffer with the specified range removed.
	Delete(start, end int) Buffer
	// Save writes the buffer contents to disk using the filename. It returns
	// an error if the write fails or no filename was specified.
	Save() error
}

type buffer struct {
	filename string
	contents rope.Rope
	dirty    bool
}

func (b *buffer) Save() error {
	if b.filename == "" {
		return os.ErrInvalid
	}

	dir, name := filepath.Split(b.filename)
	// Create a temporary file in the same directory so that os.Rename works
	// across filesystems.
	tmp, err := os.CreateTemp(dir, name+".tmp*")
	if err != nil {
		return err
	}
	// Write contents to the temporary file first.
	if _, err := rope.Write(tmp, b.contents); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}

	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return err
	}

	// Atomically replace the target file.
	if err := os.Rename(tmp.Name(), b.filename); err != nil {
		os.Remove(tmp.Name())
		return err
	}

	b.dirty = false
	return nil
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

func (b *buffer) Delete(start, end int) Buffer {
	if start < 0 {
		start = 0
	}
	if end > b.contents.Len() {
		end = b.contents.Len()
	}
	if start > end {
		start, end = end, start
	}
	if start == end {
		return b
	}
	nb := &buffer{
		filename: b.filename,
		contents: b.contents.Delete(start, end),
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
