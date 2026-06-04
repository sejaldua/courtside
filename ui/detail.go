package ui

import (
	tea "charm.land/bubbletea/v2"
)

type detail struct {
	title string
}

func newDetail(title string) detail {
	return detail{title: title}
}

func (m detail) Init() tea.Cmd {
	return nil
}

func (m detail) Update(msg tea.Msg) (detail, tea.Cmd) {
	return m, nil
}

func (m detail) View() tea.View {
	body := docStyle.Render(
		m.title + "\n\n" +
			"Quarter: 3\n" +
			"Time left: 04:21\n\n" +
			"Top scorer: Player X — 28 pts\n\n" +
			"(press esc to go back)",
	)
	v := tea.NewView(body)
	v.AltScreen = true
	return v
}
