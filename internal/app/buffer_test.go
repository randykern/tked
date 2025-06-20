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

func TestBufferIndexForRow(t *testing.T) {
	b := NewBuffer("", rope.NewRope("a\nb\nc"))

	idx, row := b.IndexForRow(0)
	if idx != 0 || row != 0 {
		t.Fatalf("expected row 0 index 0 got row %d index %d", row, idx)
	}

	idx, row = b.IndexForRow(1)
	if idx != 2 || row != 1 {
		t.Fatalf("expected row 1 index 2 got row %d index %d", row, idx)
	}

	idx, row = b.IndexForRow(10)
	if idx != 4 || row != 2 {
		t.Fatalf("expected last row index 4 got row %d index %d", row, idx)
	}

	idx, row = b.IndexForRow(-5)
	if idx != 0 || row != 0 {
		t.Fatalf("expected negative row clamp to 0 got row %d index %d", row, idx)
	}
}

func TestBufferUndoRedoAndProperties(t *testing.T) {
	b := NewBuffer("", rope.NewRope("a"))
	prop := RegisterBufferProperty()
	b.SetProperty(prop, "initial")

	b.Insert(1, "b")
	b.SetProperty(prop, "after insert")

	b.Delete(0, 1)

	b.Undo()
	if b.Contents().String() != "ab" {
		t.Fatalf("undo delete failed, got %q", b.Contents().String())
	}
	if b.GetProperty(prop) != "after insert" {
		t.Fatalf("property not restored on undo")
	}

	b.Undo()
	if b.Contents().String() != "a" {
		t.Fatalf("undo insert failed, got %q", b.Contents().String())
	}
	if b.GetProperty(prop) != "initial" {
		t.Fatalf("property not restored after second undo")
	}

	b.Redo()
	if b.Contents().String() != "ab" {
		t.Fatalf("redo insert failed, got %q", b.Contents().String())
	}
	if b.GetProperty(prop) != "after insert" {
		t.Fatalf("property not restored after redo")
	}
}

func TestBufferOnChange(t *testing.T) {
	b := NewBuffer("", rope.NewRope("abc"))
	calls := 0
	var lastStart, lastEnd int
	reg := b.OnChange(func(_ Buffer, start, end int, _ any) {
		calls++
		lastStart = start
		lastEnd = end
	}, nil)

	b.Insert(1, "x")
	if calls != 1 || lastStart != 1 || lastEnd != 2 {
		t.Fatalf("callback values unexpected: calls %d start %d end %d", calls, lastStart, lastEnd)
	}

	b.Delete(0, 1)
	if calls != 2 || lastStart != 0 || lastEnd != 1 {
		t.Fatalf("callback after delete unexpected: calls %d start %d end %d", calls, lastStart, lastEnd)
	}

	reg.Remove()
	b.Insert(0, "y")
	if calls != 2 {
		t.Fatalf("callback not removed")
	}
	_ = lastStart
	_ = lastEnd
}
