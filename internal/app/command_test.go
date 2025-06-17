package app

import "testing"

func TestRegisterAndGetCommand(t *testing.T) {
	commands = make(map[string]Command)
	cmd := &CommandExit{}
	RegisterCommand(cmd.Name(), cmd)
	got := GetCommand("exit")
	if got != cmd {
		t.Fatalf("expected same command instance")
	}
}

func TestRegisterCommandDuplicatePanics(t *testing.T) {
	commands = make(map[string]Command)
	RegisterCommand("exit", &CommandExit{})
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic for duplicate register")
		}
	}()
	RegisterCommand("exit", &CommandExit{})
}

func TestGetCommandMissingPanics(t *testing.T) {
	commands = make(map[string]Command)
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic for missing command")
		}
	}()
	GetCommand("nosuch")
}
