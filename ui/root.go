package ui

import (
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/NolanFogarty/courtside/backend"
)

const (
	listview int = iota
	detailview
)

// refreshInterval is how often live data (the list while viewing today, and an
// in-progress game's detail) is re-fetched.
const refreshInterval = 15 * time.Second

// refreshTickMsg is the periodic heartbeat that drives auto-refresh.
type refreshTickMsg struct{}

func refreshTick() tea.Cmd {
	return tea.Tick(refreshInterval, func(time.Time) tea.Msg { return refreshTickMsg{} })
}

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
	return tea.Batch(r.list.Init(), refreshTick())
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

	// Auto-refresh heartbeat: re-fetch whichever view is active and wants fresh
	// data, then reschedule. A single tick lineage avoids overlapping polls.
	if _, ok := msg.(refreshTickMsg); ok {
		var cmd tea.Cmd
		switch r.current {
		case listview:
			cmd = r.list.refreshCmd()
		case detailview:
			cmd = r.detail.refreshCmd()
		}
		if cmd != nil {
			return r, tea.Batch(cmd, refreshTick())
		}
		return r, refreshTick()
	}

	// Day navigation messages always belong to the list, even if the user has
	// since opened the detail view.
	switch msg.(type) {
	case gamesLoadedMsg, daySettledMsg:
		updated, cmd := r.list.Update(msg)
		r.list = updated.(gamelist)
		return r, cmd
	}

	switch r.current {
	case listview:
		// Enter on a selected item navigates to the detail screen, unless the
		// list is filtering or the user is typing a date to jump to.
		if key, ok := msg.(tea.KeyPressMsg); ok && key.String() == "enter" &&
			r.list.list.FilterState() != list.Filtering && !r.list.entering {
			if sel, ok := r.list.list.SelectedItem().(item); ok {
				r.detail = newDetail(sel.game, r.width, r.height)
				r.current = detailview
				return r, r.detail.Init() // kick off the async fetch
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
