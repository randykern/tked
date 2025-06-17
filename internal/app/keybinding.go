package app

import "github.com/gdamore/tcell/v2"

type KeyBinding struct {
	Key     tcell.Key
	Mod     tcell.ModMask
	Command Command
}

type KeyBindings struct {
	KeyBindings []KeyBinding
}

func (k *KeyBindings) GetCommandForKey(key tcell.Key, mod tcell.ModMask) Command {
	for _, binding := range k.KeyBindings {
		if binding.Key == key && binding.Mod == mod {
			return binding.Command
		}
	}
	return nil
}

func DefaultKeyBindings() KeyBindings {
	return KeyBindings{
		KeyBindings: []KeyBinding{
			{tcell.KeyEscape, tcell.ModNone, GetCommand("exit")},
			/*
				{tcell.KeyCtrlZ, tcell.ModCtrl, GetCommand("undo")},
				{tcell.KeyCtrlR, tcell.ModCtrl, GetCommand("redo")},
				{tcell.KeyCtrlS, tcell.ModCtrl, GetCommand("save")},
				{tcell.KeyCtrlO, tcell.ModCtrl, GetCommand("open")},
				{tcell.KeyUp, tcell.ModNone, GetCommand("up")},
				{tcell.KeyDown, tcell.ModNone, GetCommand("down")},
				{tcell.KeyLeft, tcell.ModNone, GetCommand("left")},
				{tcell.KeyRight, tcell.ModNone, GetCommand("right")},
				{tcell.KeyBackspace, tcell.ModNone, GetCommand("backspace")},
				{tcell.KeyDelete, tcell.ModNone, GetCommand("delete")},
			*/
		},
	}
}
