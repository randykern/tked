package app

import "github.com/gdamore/tcell/v2"

type KeyBinding struct {
	Key     tcell.Key
	Mod     tcell.ModMask
	Command Command
}

type keyCombo struct {
	key tcell.Key
	mod tcell.ModMask
}

type KeyBindings struct {
	bindings map[keyCombo]Command
}

func NewKeyBindings(b []KeyBinding) KeyBindings {
	kb := KeyBindings{bindings: make(map[keyCombo]Command, len(b))}
	for _, binding := range b {
		kb.bindings[keyCombo{binding.Key, binding.Mod}] = binding.Command
	}
	return kb
}

func (k KeyBindings) GetCommandForKey(key tcell.Key, mod tcell.ModMask) Command {
	return k.bindings[keyCombo{key, mod}]
}

func DefaultKeyBindings() KeyBindings {
	return NewKeyBindings([]KeyBinding{
		{tcell.KeyEscape, tcell.ModNone, GetCommand("exit")},
		{tcell.KeyCtrlZ, tcell.ModCtrl, GetCommand("undo")},
		{tcell.KeyCtrlR, tcell.ModCtrl, GetCommand("redo")},
		{tcell.KeyCtrlS, tcell.ModCtrl, GetCommand("save")},
		{tcell.KeyCtrlO, tcell.ModCtrl, GetCommand("open")},
		{tcell.KeyUp, tcell.ModNone, GetCommand("up")},
		{tcell.KeyDown, tcell.ModNone, GetCommand("down")},
		{tcell.KeyLeft, tcell.ModNone, GetCommand("left")},
		{tcell.KeyRight, tcell.ModNone, GetCommand("right")},
		{tcell.KeyBackspace, tcell.ModNone, GetCommand("backspace")},
		{tcell.KeyBackspace2, tcell.ModNone, GetCommand("backspace")},
		{tcell.KeyDelete, tcell.ModNone, GetCommand("delete")},
		{tcell.KeyPgUp, tcell.ModNone, GetCommand("pageup")},
		{tcell.KeyPgDn, tcell.ModNone, GetCommand("pagedown")},
	})
}
