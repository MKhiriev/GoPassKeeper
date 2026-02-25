package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	up       key.Binding
	down     key.Binding
	left     key.Binding
	right    key.Binding
	enter    key.Binding
	esc      key.Binding
	tab      key.Binding
	backtab  key.Binding
	quit     key.Binding
	logout   key.Binding
	newItem  key.Binding
	sync     key.Binding
	edit     key.Binding
	delete   key.Binding
	copy     key.Binding
	copyUser key.Binding
	yes      key.Binding
	no       key.Binding
}

var keys = keyMap{
	up:       key.NewBinding(key.WithKeys("up", "k")),
	down:     key.NewBinding(key.WithKeys("down", "j")),
	left:     key.NewBinding(key.WithKeys("left", "h")),
	right:    key.NewBinding(key.WithKeys("right", "l")),
	enter:    key.NewBinding(key.WithKeys("enter")),
	esc:      key.NewBinding(key.WithKeys("esc")),
	tab:      key.NewBinding(key.WithKeys("tab")),
	backtab:  key.NewBinding(key.WithKeys("shift+tab")),
	quit:     key.NewBinding(key.WithKeys("q", "ctrl+c")),
	logout:   key.NewBinding(key.WithKeys("l")),
	newItem:  key.NewBinding(key.WithKeys("n")),
	sync:     key.NewBinding(key.WithKeys("s")),
	edit:     key.NewBinding(key.WithKeys("e")),
	delete:   key.NewBinding(key.WithKeys("d")),
	copy:     key.NewBinding(key.WithKeys("c")),
	copyUser: key.NewBinding(key.WithKeys("u")),
	yes:      key.NewBinding(key.WithKeys("y")),
	no:       key.NewBinding(key.WithKeys("n")),
}
