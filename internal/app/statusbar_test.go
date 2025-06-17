package app

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestStatusBarInputEnter(t *testing.T) {
	screen := tcell.NewSimulationScreen("")
	screen.Init()
	screen.SetSize(20, 5)
	sb := NewStatusBar()
	done := make(chan struct{})
	var val string
	var ok bool
	go func() {
		val, ok = sb.Input(screen, "file: ")
		close(done)
	}()
	screen.InjectKey(tcell.KeyRune, 't', tcell.ModNone)
	screen.InjectKey(tcell.KeyRune, 'x', tcell.ModNone)
	screen.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)
	<-done
	if !ok || val != "tx" {
		t.Fatalf("expected tx true got %q %v", val, ok)
	}
}

func TestStatusBarInputEsc(t *testing.T) {
	screen := tcell.NewSimulationScreen("")
	screen.Init()
	screen.SetSize(20, 5)
	sb := NewStatusBar()
	done := make(chan struct{})
	var val string
	var ok bool
	go func() {
		val, ok = sb.Input(screen, "file: ")
		close(done)
	}()
	screen.InjectKey(tcell.KeyRune, 'a', tcell.ModNone)
	screen.InjectKey(tcell.KeyEscape, 0, tcell.ModNone)
	<-done
	if ok || val != "" {
		t.Fatalf("expected cancel got %q %v", val, ok)
	}
}
