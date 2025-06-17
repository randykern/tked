package app

import "testing"

func TestSettingsTabWidth(t *testing.T) {
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
