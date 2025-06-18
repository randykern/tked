package app

import "github.com/gdamore/tcell/v2"

// TabBar describes the behaviour of a tab bar component.
type TabBar interface {
	// SetScreen sets the screen that the tab bar will draw on.
	SetScreen(s tcell.Screen)
	// Draw renders tabs for each view highlighting the current one.
	Draw(views []View, current int)
	// ViewIndexAt returns the view index for the given screen coordinates.
	// The boolean return will be false if the coordinates are outside any tab.
	ViewIndexAt(x, y int) (int, bool)
	// CloseIndexAt returns the index of the view close button at the given
	// coordinates. The boolean return will be false if the coordinates are
	// outside any close button.
	CloseIndexAt(x, y int) (int, bool)
}

type tabPosition struct {
	start      int
	end        int
	closeStart int
	closeEnd   int
}

type tabBar struct {
	screen       tcell.Screen
	tabPositions []tabPosition
}

// SetScreen sets the screen that the tab bar will draw on.
func (tb *tabBar) SetScreen(s tcell.Screen) {
	if s == nil {
		panic("screen is nil")
	}
	tb.screen = s
}

// Draw renders the tab bar for the provided views.
func (tb *tabBar) Draw(views []View, current int) {
	width, _ := tb.screen.Size()
	tb.tabPositions = make([]tabPosition, len(views))

	col := 0
	for i, v := range views {
		title := v.Buffer().GetTitle()
		if title == "" {
			title = "Untitled"
		}
		if v.Buffer().IsDirty() {
			title += "*"
		}

		// add spaces around title and include a close button
		text := " " + title + " "
		titleRunes := []rune(text)
		closeStart := col + len(titleRunes)
		text += "X "
		start := col
		for _, r := range text {
			if col >= width {
				break
			}
			style := tcell.StyleDefault
			if i == current {
				style = style.Reverse(true)
			} else {
				style = style.Foreground(tcell.ColorWhite)
			}
			tb.screen.SetContent(col, 0, r, nil, style)
			col++
		}
		tb.tabPositions[i] = tabPosition{start: start, end: col, closeStart: closeStart, closeEnd: closeStart + 1}
		if col >= width {
			break
		}
	}
	// clear remaining space
	for ; col < width; col++ {
		tb.screen.SetContent(col, 0, ' ', nil, tcell.StyleDefault)
	}
}

// ViewIndexAt returns the view index for the given coordinates.
func (tb *tabBar) ViewIndexAt(x, y int) (int, bool) {
	if y != 0 {
		return -1, false
	}
	for i, pos := range tb.tabPositions {
		if x >= pos.start && x < pos.end {
			return i, true
		}
	}
	return -1, false
}

func (tb *tabBar) CloseIndexAt(x, y int) (int, bool) {
	if y != 0 {
		return -1, false
	}
	for i, pos := range tb.tabPositions {
		if x >= pos.closeStart && x < pos.closeEnd {
			return i, true
		}
	}
	return -1, false
}

// NewTabBar creates a new tab bar instance.
func NewTabBar() TabBar {
	return &tabBar{}
}
