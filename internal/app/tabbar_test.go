package app

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestTabBarViewIndexAt(t *testing.T) {
	tb := NewTabBar().(*tabBar)
	screen := tcell.NewSimulationScreen("")
	screen.Init()
	screen.SetSize(40, 5)
	tb.SetScreen(screen)
	views := []View{NewView("a.txt", nil), NewView("b.txt", nil)}
	tb.Draw(views, 0)

	if idx, ok := tb.ViewIndexAt(1, 0); !ok || idx != 0 {
		t.Fatalf("expected first tab index 0 got %d %v", idx, ok)
	}
	start := tb.tabPositions[1].start
	if idx, ok := tb.ViewIndexAt(start, 0); !ok || idx != 1 {
		t.Fatalf("expected second tab index 1 got %d %v", idx, ok)
	}
	if _, ok := tb.ViewIndexAt(0, 1); ok {
		t.Fatalf("expected no tab outside top row")
	}
}

func TestTabBarCloseIndexAt(t *testing.T) {
	tb := NewTabBar().(*tabBar)
	screen := tcell.NewSimulationScreen("")
	screen.Init()
	screen.SetSize(40, 5)
	tb.SetScreen(screen)
	views := []View{NewView("a.txt", nil), NewView("b.txt", nil)}
	tb.Draw(views, 0)

	pos := tb.tabPositions[1]
	if idx, ok := tb.CloseIndexAt(pos.closeStart, 0); !ok || idx != 1 {
		t.Fatalf("expected close index 1 got %d %v", idx, ok)
	}
}
