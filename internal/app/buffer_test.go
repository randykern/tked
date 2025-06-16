package app

import "testing"

func TestBufferDelete(t *testing.T) {
	b, err := NewBuffer("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b = b.Insert(0, "hello")
	b = b.Delete(1, 2) // remove 'e'
	got := b.Contents().String()
	if got != "hllo" {
		t.Fatalf("expected %q got %q", "hllo", got)
	}
}

func TestBufferInsert(t *testing.T) {
	b, err := NewBuffer("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// inserting into empty buffer
	b2 := b.Insert(0, "world")
	if got := b.Contents().String(); got != "" {
		t.Fatalf("original buffer modified: %q", got)
	}
	if got := b2.Contents().String(); got != "world" {
		t.Fatalf("expected %q got %q", "world", got)
	}
	if !b2.IsDirty() {
		t.Fatalf("expected dirty buffer after insert")
	}

	// insert at beginning
	b3 := b2.Insert(0, "hello ")
	if got := b3.Contents().String(); got != "hello world" {
		t.Fatalf("expected %q got %q", "hello world", got)
	}

	// insert beyond end should append
	b4 := b3.Insert(100, " end")
	if got := b4.Contents().String(); got != "hello world end" {
		t.Fatalf("expected %q got %q", "hello world end", got)
	}

	// insert with negative index should prepend
	b5 := b3.Insert(-10, "start ")
	if got := b5.Contents().String(); got != "start hello world" {
		t.Fatalf("expected %q got %q", "start hello world", got)
	}
}
