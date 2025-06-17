package app

import "testing"

func TestViewInsertUndoRedo(t *testing.T) {
	b, _ := NewBuffer("")
	v := NewView(b)
	v.InsertRune('a')
	v.InsertRune('b')
	if got := v.Buffer().Contents().String(); got != "ab" {
		t.Fatalf("expected ab got %q", got)
	}
	v.Undo()
	if got := v.Buffer().Contents().String(); got != "a" {
		t.Fatalf("after undo expected a got %q", got)
	}
	v.Redo()
	if got := v.Buffer().Contents().String(); got != "ab" {
		t.Fatalf("after redo expected ab got %q", got)
	}
}

func TestViewDeleteRune(t *testing.T) {
	b, _ := NewBuffer("")
	v := NewView(b)
	v.InsertRune('a')
	v.InsertRune('b')
	v.DeleteRune(false)
	if got := v.Buffer().Contents().String(); got != "a" {
		t.Fatalf("expected a got %q", got)
	}
	v.DeleteRune(true)
	if got := v.Buffer().Contents().String(); got != "a" {
		t.Fatalf("forward delete at end should not change, got %q", got)
	}
}
