package app

import (
	"os"
	"testing"

	"github.com/gdamore/tcell/v2"
)

type dummyApp struct {
	opened string
	sb     StatusBar
	view   View
}

func (d *dummyApp) OpenFile(name string) error { d.opened = name; return nil }
func (d *dummyApp) Run(tcell.Screen)           {}
func (d *dummyApp) Settings() Settings         { return NewSettings() }
func (d *dummyApp) GetStatusBar() StatusBar    { return d.sb }
func (d *dummyApp) LoadSettings(string) error  { return nil }
func (d *dummyApp) GetCurrentView() View       { return d.view }

type stubStatusBar struct{}

func (stubStatusBar) SetScreen(tcell.Screen)          {}
func (stubStatusBar) Draw(View)                       {}
func (stubStatusBar) Message(string)                  {}
func (stubStatusBar) Messagef(string, ...interface{}) {}
func (stubStatusBar) Error(string)                    {}
func (stubStatusBar) Errorf(string, ...interface{})   {}
func (stubStatusBar) Input(string) (string, bool)     { return "test.txt", true }

func TestCommandOpenExecute(t *testing.T) {
	commands = make(map[string]Command)
	registerCommands()
	d := &dummyApp{sb: stubStatusBar{}}
	screen := tcell.NewSimulationScreen("")
	screen.Init()
	cmd := &CommandOpen{}
	if _, err := cmd.Execute(d, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.opened != "test.txt" {
		t.Fatalf("expected file opened")
	}
}

func TestCommandMoveExecute(t *testing.T) {
	b, _ := NewBuffer("")
	v := NewView(b)
	d := &dummyApp{view: v}
	screen := tcell.NewSimulationScreen("")
	screen.Init()
	screen.SetSize(10, 10)
	v.SetCursor(0, 0)
	c := &CommandMove{dRow: 1, dCol: 1}
	if _, err := c.Execute(d, nil); err != nil {
		t.Fatalf("error: %v", err)
	}
	r, c2 := v.Cursor()
	if r != 1 || c2 != 1 {
		t.Fatalf("expected cursor 1,1 got %d,%d", r, c2)
	}
}

func TestCommandSaveExecute(t *testing.T) {
	tmp, _ := os.CreateTemp("", "cmdsave*.txt")
	tmp.Close()
	defer os.Remove(tmp.Name())
	b, _ := NewBuffer(tmp.Name())
	v := NewView(b.Insert(0, "data"))
	d := &dummyApp{view: v}
	screen := tcell.NewSimulationScreen("")
	screen.Init()
	cmd := &CommandSave{}
	if _, err := cmd.Execute(d, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, _ := os.ReadFile(tmp.Name())
	if string(data) != "data" {
		t.Fatalf("file not saved")
	}
}

func TestCommandMoveName(t *testing.T) {
	tests := []struct {
		cmd  CommandMove
		name string
	}{
		{CommandMove{dRow: -1}, "up"},
		{CommandMove{dRow: 1}, "down"},
		{CommandMove{dCol: -1}, "left"},
		{CommandMove{dCol: 1}, "right"},
		{CommandMove{dRow: 2, dCol: 2}, "move"},
	}

	for _, tt := range tests {
		if got := tt.cmd.Name(); got != tt.name {
			t.Fatalf("expected %s got %s", tt.name, got)
		}
	}
}
