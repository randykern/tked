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
	v := a.GetCurrentView()
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
	a.statusBar.SetScreen(screen)

	ev := tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
	a.handleKey(ev)
	got := a.GetCurrentView().Buffer().Contents().String()
	if got != "\n" {
		t.Fatalf("expected newline got %q", got)
	}
}

func TestHandleMouseTabClick(t *testing.T) {
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
	a.statusBar.SetScreen(screen)
	a.tabBar.SetScreen(screen)

	a.views = append(a.views, NewView("second.txt", nil))
	for _, v := range a.views {
		v.Resize(4, 20)
	}
	a.tabBar.Draw(a.views, a.currentView)

	tb := a.tabBar.(*tabBar)
	x := tb.tabPositions[1].start
	ev := tcell.NewEventMouse(x, 0, tcell.Button1, tcell.ModNone)
	a.handleMouse(ev)
	if a.currentView != 1 {
		t.Fatalf("expected current view 1 got %d", a.currentView)
	}
}

type stubStatusBarClose struct{}

func (stubStatusBarClose) SetScreen(tcell.Screen)      {}
func (stubStatusBarClose) Draw(View)                   {}
func (stubStatusBarClose) Message(string)              {}
func (stubStatusBarClose) Messagef(string, ...any)     {}
func (stubStatusBarClose) Error(string)                {}
func (stubStatusBarClose) Errorf(string, ...any)       {}
func (stubStatusBarClose) Input(string) (string, bool) { return "n", true }

func TestHandleMouseTabClose(t *testing.T) {
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
	a.statusBar.SetScreen(screen)
	a.tabBar.SetScreen(screen)
	a.statusBar = stubStatusBarClose{}

	v2 := NewView("second.txt", nil)
	v2.InsertRune('a')
	a.views = append(a.views, v2)
	for _, v := range a.views {
		v.Resize(4, 20)
	}
	a.tabBar.Draw(a.views, a.currentView)

	tb := a.tabBar.(*tabBar)
	x := tb.tabPositions[1].closeStart
	ev := tcell.NewEventMouse(x, 0, tcell.Button1, tcell.ModNone)
	a.handleMouse(ev)
	if len(a.views) != 1 {
		t.Fatalf("expected view to be closed")
	}
	if a.currentView != 0 {
		t.Fatalf("expected current view 0 got %d", a.currentView)
	}
}
