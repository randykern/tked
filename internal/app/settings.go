package app

// Settings defines configurable editor options.
type Settings interface {
	// TabWidth returns the current tab width in spaces.
	TabWidth() int
	// SetTabWidth updates the tab width. Values <= 0 are ignored.
	SetTabWidth(width int)
}

// settings is the default implementation of Settings.
type settings struct {
	tabWidth int
}

// NewSettings creates a new Settings instance with default values.
func NewSettings() Settings {
	return &settings{tabWidth: 4}
}

func (s *settings) TabWidth() int { return s.tabWidth }

func (s *settings) SetTabWidth(width int) {
	if width <= 0 {
		return
	}
	s.tabWidth = width
}
