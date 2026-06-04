package ui

import (
	"fmt"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/NolanFogarty/courtside/backend"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type item struct {
	title, desc string
	game        backend.Game
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type gamelist struct {
	list list.Model
}

func (m gamelist) Init() tea.Cmd {
	return nil
}

func (m gamelist) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m gamelist) View() tea.View {
	// if m.current = "listview"
	// return v.listview..
	// if current = "detailedview"
	// return m.detailedview
	v := tea.NewView(docStyle.Render(m.list.View()))
	v.AltScreen = true
	return v
}

func formatGame(g backend.Game) item {
	scoreboard := fmt.Sprintf("%-13s %3d - %-3d %13s", g.AwayTeam, g.AwayScore, g.HomeScore, g.HomeTeam)
	pad := (len(scoreboard) - len(g.GameClock)) / 2
	clock := fmt.Sprintf("%*s", pad+len(g.GameClock), g.GameClock)
	return item{title: scoreboard, desc: clock, game: g}
}

func newGamesList(games []backend.Game) gamelist {
	items := make([]list.Item, len(games))
	for i, g := range games {
		items[i] = formatGame(g)
	}

	m := gamelist{list: list.New(items, list.NewDefaultDelegate(), 0, 0)}
	m.list.Title = "NBA Scores"

	return m
}

func Run(games []backend.Game) error {
	m := newRoot(games)
	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}
