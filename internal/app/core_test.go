package app

import (
	"os"
	"testing"
)

func TestNewAppAndOpenFile(t *testing.T) {
	commands = make(map[string]Command)
	RegisterCommands()
	aInt, err := NewApp()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	a := aInt.(*app)
	if len(a.views) != 1 {
		t.Fatalf("expected 1 view")
	}
	tmp, _ := os.CreateTemp("", "file*.txt")
	os.WriteFile(tmp.Name(), []byte("hi"), 0644)
	defer os.Remove(tmp.Name())
	if err := a.OpenFile(tmp.Name()); err != nil {
		t.Fatalf("open error: %v", err)
	}
	if len(a.views) != 1 {
		t.Fatalf("empty view should be replaced")
	}
	v := a.getCurrentView()
	if v.Buffer().GetFilename() != tmp.Name() {
		t.Fatalf("filename not set")
	}
	if v.Buffer().Contents().String() != "hi" {
		t.Fatalf("content not loaded")
	}
}
