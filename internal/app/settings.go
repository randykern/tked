package app

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/pelletier/go-toml/v2"
)

// Settings defines configurable editor options.
type Settings interface {
	// TabWidth returns the current tab width in spaces.
	TabWidth() int
	// SetTabWidth updates the tab width. Values <= 0 are ignored.
	SetTabWidth(width int)
	// KeyBindings returns the current key bindings.
	KeyBindings() KeyBindings
	// Load reads settings from the provided TOML file.
	Load(filename string) error
	// Save writes the current settings to the provided TOML file.
	Save(filename string) error
}

// settings is the default implementation of Settings.
type settings struct {
	tabWidth    int
	keyBindings KeyBindings
}

// NewSettings creates a new Settings instance with default values.
func NewSettings() Settings {
	return &settings{
		tabWidth:    4,
		keyBindings: DefaultKeyBindings(),
	}
}

func (s *settings) TabWidth() int { return s.tabWidth }

func (s *settings) SetTabWidth(width int) {
	if width <= 0 {
		return
	}
	s.tabWidth = width
}

func (s *settings) KeyBindings() KeyBindings { return s.keyBindings }

func (s *settings) Load(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var cfg struct {
		TabWidth int `toml:"tab_width"`
		Bindings []struct {
			Key     int    `toml:"key"`
			Mod     uint32 `toml:"mod"`
			Command string `toml:"command"`
		} `toml:"key_bindings"`
	}

	if err := toml.Unmarshal(data, &cfg); err != nil {
		return err
	}

	if cfg.TabWidth > 0 {
		s.tabWidth = cfg.TabWidth
	}

	if len(cfg.Bindings) > 0 {
		bindings := make([]KeyBinding, 0, len(cfg.Bindings))
		seen := make(map[keyCombo]struct{}, len(cfg.Bindings))
		for _, b := range cfg.Bindings {
			kc := keyCombo{key: tcell.Key(b.Key), mod: tcell.ModMask(b.Mod)}
			if _, ok := seen[kc]; ok {
				return fmt.Errorf("duplicate binding for key %d mod %d", b.Key, b.Mod)
			}
			seen[kc] = struct{}{}
			bindings = append(bindings, KeyBinding{
				Key:     kc.key,
				Mod:     kc.mod,
				Command: GetCommand(b.Command),
			})
		}
		s.keyBindings = NewKeyBindings(bindings)
	}

	return nil
}

func (s *settings) Save(filename string) error {
	var cfg struct {
		TabWidth int `toml:"tab_width"`
		Bindings []struct {
			Key     int    `toml:"key"`
			Mod     uint32 `toml:"mod"`
			Command string `toml:"command"`
		} `toml:"key_bindings"`
	}

	cfg.TabWidth = s.tabWidth
	cfg.Bindings = make([]struct {
		Key     int    `toml:"key"`
		Mod     uint32 `toml:"mod"`
		Command string `toml:"command"`
	}, 0, len(s.keyBindings.bindings))

	for kc, cmd := range s.keyBindings.bindings {
		cfg.Bindings = append(cfg.Bindings, struct {
			Key     int    `toml:"key"`
			Mod     uint32 `toml:"mod"`
			Command string `toml:"command"`
		}{Key: int(kc.key), Mod: uint32(kc.mod), Command: cmd.Name()})
	}

	data, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}
