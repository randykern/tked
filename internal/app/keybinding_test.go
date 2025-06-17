package app

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestNewKeyBindings(t *testing.T) {
	commands = make(map[string]Command)
	RegisterCommand("exit", &CommandExit{})
	kb := NewKeyBindings([]KeyBinding{{Key: tcell.KeyCtrlA, Mod: tcell.ModCtrl, Command: GetCommand("exit")}})
	if kb.GetCommandForKey(tcell.KeyCtrlA, tcell.ModCtrl) == nil {
		t.Fatalf("expected command for key")
	}
}

func TestDefaultKeyBindingsIncludesExit(t *testing.T) {
	commands = make(map[string]Command)
	RegisterCommands()
	kb := DefaultKeyBindings()
	if kb.GetCommandForKey(tcell.KeyEscape, tcell.ModNone) == nil {
		t.Fatalf("expected exit binding present")
	}
}
