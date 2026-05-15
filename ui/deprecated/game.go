package ui

import (
	tea "charm.land/bubbletea/v2"
)

type gamestats struct{}

func (m *gamestats) Init() tea.Cmd {
	return nil
}

func (m *gamestats) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			return m, nil
		}
	}
	return m, nil
}

func (m *gamestats) View() tea.View {
	v := tea.NewView("hello screen2")
	v.AltScreen = true
	return v
}
