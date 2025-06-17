package app

import (
	"os"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestNewAppAndOpenFile(t *testing.T) {
	commands = make(map[string]Command)
	registerCommands()
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

func TestHandleKeyEnter(t *testing.T) {
	commands = make(map[string]Command)
	registerCommands()
	aInt, err := NewApp()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	a := aInt.(*app)
	screen := tcell.NewSimulationScreen("")
	screen.Init()
	screen.SetSize(20, 5)
	ev := tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
	a.handleKey(screen, ev)
	got := a.getCurrentView().Buffer().Contents().String()
	if got != "\n" {
		t.Fatalf("expected newline got %q", got)
	}
}
