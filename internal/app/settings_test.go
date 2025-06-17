package app

import (
	"fmt"
	"os"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestSettingsTabWidth(t *testing.T) {
	commands = make(map[string]Command)
	registerCommands()
	s := NewSettings()
	if s.TabWidth() != 4 {
		t.Fatalf("default tab width should be 4, got %d", s.TabWidth())
	}
	s.SetTabWidth(8)
	if s.TabWidth() != 8 {
		t.Fatalf("expected tab width 8, got %d", s.TabWidth())
	}
	s.SetTabWidth(-1)
	if s.TabWidth() != 8 {
		t.Fatalf("negative value should not change tab width, got %d", s.TabWidth())
	}
}

func TestSettingsLoadDuplicateBinding(t *testing.T) {
	commands = make(map[string]Command)
	registerCommands()
	tmp, err := os.CreateTemp("", "config_*.toml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer os.Remove(tmp.Name())

	content := "" +
		"tab_width = 4\n" +
		"\n" +
		"[[key_bindings]]\n" +
		"key = %d\n" +
		"mod = %d\n" +
		"command = \"exit\"\n" +
		"\n" +
		"[[key_bindings]]\n" +
		"key = %d\n" +
		"mod = %d\n" +
		"command = \"undo\"\n"
	data := []byte(fmt.Sprintf(content, tcell.KeyEscape, tcell.ModNone, tcell.KeyEscape, tcell.ModNone))
	if err := os.WriteFile(tmp.Name(), data, 0644); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = NewSettingsFromFile(tmp.Name())
	if err == nil {
		t.Fatalf("expected error for duplicate binding, got nil")
	}
}
