package app

import (
	"github.com/gdamore/tcell/v2"
)

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
		{tcell.KeyCtrlD, tcell.ModCtrl, GetCommand("exit")},
		{tcell.KeyCtrlZ, tcell.ModCtrl, GetCommand("undo")},
		{tcell.KeyCtrlR, tcell.ModCtrl, GetCommand("redo")},
		{tcell.KeyCtrlN, tcell.ModCtrl, GetCommand("new")},
		{tcell.KeyCtrlW, tcell.ModCtrl, GetCommand("saveAs")},
		{tcell.KeyCtrlS, tcell.ModCtrl, GetCommand("save")},
		{tcell.KeyCtrlO, tcell.ModCtrl, GetCommand("open")},
		{tcell.KeyCtrlQ, tcell.ModCtrl, GetCommand("close")},
		{tcell.KeyUp, tcell.ModNone, GetCommand("up")},
		{tcell.KeyUp, tcell.ModShift, GetCommand("up")},
		{tcell.KeyDown, tcell.ModNone, GetCommand("down")},
		{tcell.KeyDown, tcell.ModShift, GetCommand("down")},
		{tcell.KeyLeft, tcell.ModNone, GetCommand("left")},
		{tcell.KeyLeft, tcell.ModShift, GetCommand("left")},
		{tcell.KeyRight, tcell.ModNone, GetCommand("right")},
		{tcell.KeyRight, tcell.ModShift, GetCommand("right")},
		{tcell.KeyBackspace, tcell.ModNone, GetCommand("backspace")},
		{tcell.KeyBackspace2, tcell.ModNone, GetCommand("backspace")},
		{tcell.KeyDelete, tcell.ModNone, GetCommand("delete")},
		{tcell.KeyPgUp, tcell.ModNone, GetCommand("pageup")},
		{tcell.KeyPgDn, tcell.ModNone, GetCommand("pagedown")},
		{tcell.KeyRight, tcell.ModAlt, GetCommand("nextView")},
		{tcell.KeyLeft, tcell.ModAlt, GetCommand("prevView")},
	})
}
