package app

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestNewKeyBindings(t *testing.T) {
	commands = make(map[string]Command)
	registerCommand("exit", &CommandExit{})
	kb := NewKeyBindings([]KeyBinding{{Key: tcell.KeyCtrlA, Mod: tcell.ModCtrl, Command: GetCommand("exit")}})
	if kb.GetCommandForKey(tcell.KeyCtrlA, tcell.ModCtrl) == nil {
		t.Fatalf("expected command for key")
	}
}

func TestDefaultKeyBindingsIncludesExit(t *testing.T) {
	commands = make(map[string]Command)
	registerCommands()
	kb := DefaultKeyBindings()
	if kb.GetCommandForKey(tcell.KeyCtrlD, tcell.ModCtrl) == nil {
		t.Fatalf("expected exit binding present")
	}
}
