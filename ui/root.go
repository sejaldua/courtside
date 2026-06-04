package ui

import (
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/NolanFogarty/courtside/backend"
)

const (
	listview int = iota
	detailview
)

type root struct {
	current       int
	list          gamelist
	detail        detail
	width, height int
}

func newRoot(games []backend.Game) root {
	return root{
		current: listview,
		list:    newGamesList(games),
	}
}

func (r root) Init() tea.Cmd {
	return r.list.Init()
}

func (r root) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Quit from anywhere
	if key, ok := msg.(tea.KeyPressMsg); ok && key.String() == "ctrl+c" {
		return r, tea.Quit
	}

	// Keep both sub-views sized regardless of which is active.
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		r.width, r.height = ws.Width, ws.Height
		updated, cmd := r.list.Update(ws)
		r.list = updated.(gamelist)
		r.detail, _ = r.detail.Update(ws)
		return r, cmd
	}

	switch r.current {
	case listview:
		// Enter on a selected item navigates to the detail screen,
		// unless the list is currently being filtered.
		if key, ok := msg.(tea.KeyPressMsg); ok && key.String() == "enter" &&
			r.list.list.FilterState() != list.Filtering {
			if _, ok := r.list.list.SelectedItem().(item); ok {
				r.detail = newDetail(r.width, r.height)
				r.current = detailview
				return r, nil
			}
		}

		updated, cmd := r.list.Update(msg)
		r.list = updated.(gamelist)
		return r, cmd

	case detailview:
		if key, ok := msg.(tea.KeyPressMsg); ok {
			switch key.String() {
			case "esc", "q":
				r.current = listview
				return r, nil
			}
		}

		var cmd tea.Cmd
		r.detail, cmd = r.detail.Update(msg)
		return r, cmd
	}

	return r, nil
}

func (r root) View() tea.View {
	switch r.current {
	case detailview:
		return r.detail.View()
	default:
		return r.list.View()
	}
}
