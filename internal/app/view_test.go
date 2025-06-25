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

func TestViewDeleteRuneSelection(t *testing.T) {
	v := NewView("", rope.NewRope("hello"))
	sel := []Selection{{StartRow: 0, StartCol: 1, EndRow: 0, EndCol: 3}}
	v.SetSelections(sel)
	v.SetCursor(0, 3)
	v.DeleteRune(false)
	if got := v.Buffer().Contents().String(); got != "hlo" {
		t.Fatalf("expected 'hlo' got %q", got)
	}
	row, col := v.Cursor()
	if row != 0 || col != 1 {
		t.Fatalf("expected cursor (0,1) got (%d,%d)", row, col)
	}
	if len(v.Selections()) != 0 {
		t.Fatalf("expected selections cleared")
	}
}

// Comprehensive edge case tests for View.InsertRune function

func TestViewInsertRune_EmptyBuffer(t *testing.T) {
	// Test inserting into empty buffer at cursor (0,0)
	v := NewView("", rope.NewRope(""))
	v.InsertRune('a')
	if got := v.Buffer().Contents().String(); got != "a" {
		t.Fatalf("expected 'a' got %q", got)
	}
	row, col := v.Cursor()
	if row != 0 || col != 1 {
		t.Fatalf("expected cursor (0,1) got (%d,%d)", row, col)
	}
}

func TestViewInsertRune_BeginningOfLine(t *testing.T) {
	// Test inserting at beginning of non-empty line
	v := NewView("", rope.NewRope("hello"))
	v.SetCursor(0, 0) // cursor at beginning
	v.InsertRune('X')
	if got := v.Buffer().Contents().String(); got != "Xhello" {
		t.Fatalf("expected 'Xhello' got %q", got)
	}
	row, col := v.Cursor()
	if row != 0 || col != 1 {
		t.Fatalf("expected cursor (0,1) got (%d,%d)", row, col)
	}
}

func TestViewInsertRune_EndOfLine(t *testing.T) {
	// Test inserting at end of line
	v := NewView("", rope.NewRope("hello"))
	v.SetCursor(0, 5) // cursor at end of line
	v.InsertRune('X')
	if got := v.Buffer().Contents().String(); got != "helloX" {
		t.Fatalf("expected 'helloX' got %q", got)
	}
	row, col := v.Cursor()
	if row != 0 || col != 6 {
		t.Fatalf("expected cursor (0,6) got (%d,%d)", row, col)
	}
}

func TestViewInsertRune_MiddleOfLine(t *testing.T) {
	// Test inserting in middle of line
	v := NewView("", rope.NewRope("hello"))
	v.SetCursor(0, 2) // cursor between 'e' and 'l'
	v.InsertRune('X')
	if got := v.Buffer().Contents().String(); got != "heXllo" {
		t.Fatalf("expected 'heXllo' got %q", got)
	}
	row, col := v.Cursor()
	if row != 0 || col != 3 {
		t.Fatalf("expected cursor (0,3) got (%d,%d)", row, col)
	}
}

func TestViewInsertRune_EmptyLine(t *testing.T) {
	// Test inserting on empty line in middle of buffer
	v := NewView("", rope.NewRope("line1\n\nline3"))
	v.SetCursor(1, 0) // cursor on empty line
	v.InsertRune('X')
	if got := v.Buffer().Contents().String(); got != "line1\nX\nline3" {
		t.Fatalf("expected 'line1\nX\nline3' got %q", got)
	}
	row, col := v.Cursor()
	if row != 1 || col != 1 {
		t.Fatalf("expected cursor (1,1) got (%d,%d)", row, col)
	}
}

func TestViewInsertRune_NewlineCharacter(t *testing.T) {
	// Test inserting newline character
	v := NewView("", rope.NewRope("hello"))
	v.SetCursor(0, 2) // cursor between 'e' and 'l'
	v.InsertRune('\n')
	if got := v.Buffer().Contents().String(); got != "he\nllo" {
		t.Fatalf("expected 'he\nllo' got %q", got)
	}
	row, col := v.Cursor()
	if row != 1 || col != 0 {
		t.Fatalf("expected cursor (1,0) after newline got (%d,%d)", row, col)
	}
}

func TestViewInsertRune_NewlineAtEndOfLine(t *testing.T) {
	// Test inserting newline at end of line
	v := NewView("", rope.NewRope("hello"))
	v.SetCursor(0, 5) // cursor at end
	v.InsertRune('\n')
	if got := v.Buffer().Contents().String(); got != "hello\n" {
		t.Fatalf("expected 'hello\n' got %q", got)
	}
	row, col := v.Cursor()
	if row != 1 || col != 0 {
		t.Fatalf("expected cursor (1,0) after newline got (%d,%d)", row, col)
	}
}

func TestViewInsertRune_NewlineAtBeginningOfLine(t *testing.T) {
	// Test inserting newline at beginning of line
	v := NewView("", rope.NewRope("hello"))
	v.SetCursor(0, 0)
	v.InsertRune('\n')
	if got := v.Buffer().Contents().String(); got != "\nhello" {
		t.Fatalf("expected '\nhello' got %q", got)
	}
	row, col := v.Cursor()
	if row != 1 || col != 0 {
		t.Fatalf("expected cursor (1,0) after newline got (%d,%d)", row, col)
	}
}

func TestViewInsertRune_TabCharacter(t *testing.T) {
	// Test inserting tab character
	v := NewView("", rope.NewRope("hello"))
	v.SetCursor(0, 2)
	v.InsertRune('\t')
	if got := v.Buffer().Contents().String(); got != "he\tllo" {
		t.Fatalf("expected 'he\tllo' got %q", got)
	}
	row, col := v.Cursor()
	// Tab expands to spaces, so cursor position depends on tab width
	// Default tab width is typically 8 or 4, cursor should advance by tab width
	if row != 0 {
		t.Fatalf("expected cursor row 0 got %d", row)
	}
	// Cursor should be positioned after the tab expansion
	if col < 3 { // Should be at least past the insertion point
		t.Fatalf("expected cursor col >= 3 got %d", col)
	}
}

func TestViewInsertRune_CursorBeyondEndOfLine(t *testing.T) {
	// Test cursor positioned beyond end of line
	v := NewView("", rope.NewRope("hi"))
	v.SetCursor(0, 10) // cursor beyond end of line
	v.InsertRune('X')
	if got := v.Buffer().Contents().String(); got != "hiX" {
		t.Fatalf("expected 'hiX' got %q", got)
	}
	row, col := v.Cursor()
	if row != 0 || col != 3 {
		t.Fatalf("expected cursor (0,3) got (%d,%d)", row, col)
	}
}

func TestViewInsertRune_CursorBeyondEndOfBuffer(t *testing.T) {
	// Test cursor positioned beyond end of buffer
	v := NewView("", rope.NewRope("line1\nline2"))
	v.SetCursor(10, 5) // row beyond end of buffer
	v.InsertRune('X')
	// Should insert at end of buffer
	if got := v.Buffer().Contents().String(); got != "line1\nline2X" {
		t.Fatalf("expected 'line1\nline2X' got %q", got)
	}
	row, col := v.Cursor()
	if row != 1 || col != 6 {
		t.Fatalf("expected cursor (1,6) got (%d,%d)", row, col)
	}
}

func TestViewInsertRune_MultipleEmptyLines(t *testing.T) {
	// Test inserting on multiple empty lines
	v := NewView("", rope.NewRope("\n\n\n"))
	v.SetCursor(1, 0) // middle empty line
	v.InsertRune('X')
	if got := v.Buffer().Contents().String(); got != "\nX\n\n" {
		t.Fatalf("expected '\nX\n\n' got %q", got)
	}
	row, col := v.Cursor()
	if row != 1 || col != 1 {
		t.Fatalf("expected cursor (1,1) got (%d,%d)", row, col)
	}
}

func TestViewInsertRune_WithTabsExistingLine(t *testing.T) {
	// Test inserting on line that already contains tabs
	v := NewView("", rope.NewRope("\tindented"))
	// Set cursor at tab width position (tabs expand to spaces visually)
	// but when we set cursor at column 1, it should be after the tab character
	tabWidth := GetApp().Settings().TabWidth()
	v.SetCursor(0, tabWidth) // after the tab expansion
	v.InsertRune('X')
	if got := v.Buffer().Contents().String(); got != "\tXindented" {
		t.Fatalf("expected '\tXindented' got %q", got)
	}
	row, col := v.Cursor()
	if row != 0 {
		t.Fatalf("expected cursor row 0 got %d", row)
	}
	// Cursor should be positioned after the inserted character
	if col != tabWidth+1 {
		t.Fatalf("expected cursor col %d got %d", tabWidth+1, col)
	}
}

func TestViewInsertRune_SequentialInsertions(t *testing.T) {
	// Test multiple sequential insertions with cursor movement
	v := NewView("", rope.NewRope(""))

	// Insert characters and verify cursor moves correctly
	v.InsertRune('a')
	row, col := v.Cursor()
	if row != 0 || col != 1 {
		t.Fatalf("after 'a': expected cursor (0,1) got (%d,%d)", row, col)
	}

	v.InsertRune('b')
	row, col = v.Cursor()
	if row != 0 || col != 2 {
		t.Fatalf("after 'b': expected cursor (0,2) got (%d,%d)", row, col)
	}

	v.InsertRune('\n')
	row, col = v.Cursor()
	if row != 1 || col != 0 {
		t.Fatalf("after newline: expected cursor (1,0) got (%d,%d)", row, col)
	}

	v.InsertRune('c')
	row, col = v.Cursor()
	if row != 1 || col != 1 {
		t.Fatalf("after 'c': expected cursor (1,1) got (%d,%d)", row, col)
	}

	if got := v.Buffer().Contents().String(); got != "ab\nc" {
		t.Fatalf("expected 'ab\nc' got %q", got)
	}
}

func TestViewInsertRune_UnicodeCharacters(t *testing.T) {
	// Test inserting unicode characters
	v := NewView("", rope.NewRope("hello"))
	v.SetCursor(0, 2)
	v.InsertRune('🌟')
	if got := v.Buffer().Contents().String(); got != "he🌟llo" {
		t.Fatalf("expected 'he🌟llo' got %q", got)
	}
	row, col := v.Cursor()
	if row != 0 || col != 3 {
		t.Fatalf("expected cursor (0,3) got (%d,%d)", row, col)
	}
}

func TestViewInsertRune_CursorAtZeroZero(t *testing.T) {
	// Test cursor behavior when explicitly at (0,0)
	v := NewView("", rope.NewRope("existing"))
	v.SetCursor(0, 0)
	row, col := v.Cursor()
	if row != 0 || col != 0 {
		t.Fatalf("initial cursor should be (0,0) got (%d,%d)", row, col)
	}

	v.InsertRune('Z')
	if got := v.Buffer().Contents().String(); got != "Zexisting" {
		t.Fatalf("expected 'Zexisting' got %q", got)
	}
	row, col = v.Cursor()
	if row != 0 || col != 1 {
		t.Fatalf("expected cursor (0,1) got (%d,%d)", row, col)
	}
}

func TestViewInsertRune_LineEndingVariations(t *testing.T) {
	// Test with different line ending scenarios
	v := NewView("", rope.NewRope("line1\nline2\n"))
	v.SetCursor(2, 0) // cursor at beginning of empty line after line2
	v.InsertRune('X')
	if got := v.Buffer().Contents().String(); got != "line1\nline2\nX" {
		t.Fatalf("expected 'line1\nline2\nX' got %q", got)
	}
	row, col := v.Cursor()
	if row != 2 || col != 1 {
		t.Fatalf("expected cursor (2,1) got (%d,%d)", row, col)
	}
}

func TestViewSelections(t *testing.T) {
	v := NewView("", rope.NewRope("abc"))
	if len(v.Selections()) != 0 {
		t.Fatalf("expected no selections initially")
	}

	sel := []Selection{{StartRow: 0, StartCol: 1, EndRow: 0, EndCol: 2}}
	v.SetSelections(sel)

	got := v.Selections()
	if len(got) != 1 || got[0] != sel[0] {
		t.Fatalf("unexpected selections %#v", got)
	}
}
