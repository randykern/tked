package app

import (
	"os"
	"testing"

	"tked/internal/rope"
)

func TestBufferDelete(t *testing.T) {
	b := NewBuffer("", rope.NewRope("hello"))
	b.Delete(1, 2) // remove 'e'
	got := b.Contents().String()
	if got != "hllo" {
		t.Fatalf("expected %q got %q", "hllo", got)
	}
}

func TestBufferInsert(t *testing.T) {
	b := NewBuffer("", rope.NewRope(""))

	// inserting into empty buffer
	b.Insert(0, "world")
	got := b.Contents().String()
	if got != "world" {
		t.Fatalf("expected %q got %q", "world", got)
	}
	if !b.IsDirty() {
		t.Fatalf("expected dirty buffer after insert")
	}

	// insert at beginning
	b.Insert(0, "hello ")
	got = b.Contents().String()
	if got != "hello world" {
		t.Fatalf("expected %q got %q", "hello world", got)
	}

	// insert beyond end should append
	b.Insert(100, " end")
	got = b.Contents().String()
	if got != "hello world end" {
		t.Fatalf("expected %q got %q", "hello world end", got)
	}

	// insert with negative index should prepend
	b.Insert(-10, "start ")
	got = b.Contents().String()
	if got != "start hello world end" {
		t.Fatalf("expected %q got %q", "start hello world end", got)
	}
}

func TestBufferSave(t *testing.T) {
	f, err := os.CreateTemp("", "tked_test_*.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer os.Remove(f.Name())

	b := NewBuffer(f.Name(), rope.NewRope("hello"))
	n, err := b.Write(f)
	f.Close()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != int64(len("hello")) {
		t.Fatalf("expected %d bytes written got %d", len("hello"), n)
	}
	data, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "hello" {
		t.Fatalf("expected file contents %q got %q", "hello", string(data))
	}
	if b.IsDirty() {
		t.Fatalf("expected buffer to be clean after save")
	}
}

func TestNewBufferLoadsFile(t *testing.T) {
	tmp, err := os.CreateTemp("", "bufferload*.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer os.Remove(tmp.Name())
	tmp.WriteString("data")
	tmp.Close()

	f, err := os.Open(tmp.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer f.Close()

	b, err := NewBufferFromReader(tmp.Name(), f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.Contents().String() != "data" {
		t.Fatalf("expected contents 'data' got %q", b.Contents().String())
	}
	if b.IsDirty() {
		t.Fatalf("new buffer should not be dirty")
	}
	if b.GetFilename() != tmp.Name() {
		t.Fatalf("filename not set")
	}
}
