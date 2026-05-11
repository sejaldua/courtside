package ui

import (
	"fmt"
	tea "charm.land/bubbletea/v2"
)

const (
	listview int = iota
	gameview
)

type root struct {
	current int
	list gamelist
	game gamestats
}

func newRoot() root {
	return root{
		current: listview,
		list: newGamesList(),
	}
}

func (r root) Init() tea.Cmd {
	return r.list.Init()
}

func (r root) Update(msg. tea.Msg) (tea.Model, tea.Cmd) {
	// Quit immediately regardless of view
	if msg.(type) == tea.KeyPressMsg && msg.String() == "ctrl+c" {
		return r, tea.Quit
	}

	// switch msg := msg.(type) {
	// case tea.KeyPressMsg:
	// 	switch tea.KeyPressMsg:
	// 	case "ctrl+c":
	// 		return m, tea.Quit
	// }

	switch r.current { // what screen are we viewing?
	case listview:
		newlist, cmd := r.list.Update(msg)
		// if 'enter', switch to gamestats view
		// else nothing
	case gameview:
		newstats, cmd := r.gamestats.Update(msg)
		// if 'q' or 'esc', switch back to listview
		// else nothing
	}

	return r, nil
}

func (r root) View() tea.View {
	switch r.current {
	case listview:
		return r.list.View()
	case gameview:
		return r.game.View()
	}
	return ""
}
