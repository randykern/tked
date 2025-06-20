package app

import (
	"testing"

	"tked/internal/rope"
)

func TestViewInsertUndoRedo(t *testing.T) {
	r := rope.NewRope("")
	v := NewView("", r)
	v.InsertRune('a')
	v.InsertRune('b')
	if got := v.Buffer().Contents().String(); got != "ab" {
		t.Fatalf("expected ab got %q", got)
	}
	v.Buffer().Undo()
	if got := v.Buffer().Contents().String(); got != "a" {
		t.Fatalf("after undo expected a got %q", got)
	}
	v.Buffer().Redo()
	if got := v.Buffer().Contents().String(); got != "ab" {
		t.Fatalf("after redo expected ab got %q", got)
	}
}

func TestViewDeleteRune(t *testing.T) {
	r := rope.NewRope("")
	v := NewView("", r)
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

func TestViewDeleteRuneNewline(t *testing.T) {
	r := rope.NewRope("a\nb")
	v := NewView("", r)
	v.SetCursor(1, 0)
	v.DeleteRune(false)
	if got := v.Buffer().Contents().String(); got != "ab" {
		t.Fatalf("backspace newline failed, got %q", got)
	}
	row, col := v.Cursor()
	if row != 0 || col != 1 {
		t.Fatalf("expected cursor 0,1 got %d,%d", row, col)
	}

	r2 := rope.NewRope("a\nb")
	v2 := NewView("", r2)
	v2.SetCursor(0, 1)
	v2.DeleteRune(true)
	if got := v2.Buffer().Contents().String(); got != "ab" {
		t.Fatalf("delete newline failed, got %q", got)
	}
}
