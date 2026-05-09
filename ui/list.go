package main

import (
	"fmt"
	"os"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/NolanFogarty/courtside/backend"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type model struct {
	list list.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m model) View() tea.View {
	v := tea.NewView(docStyle.Render(m.list.View()))
	v.AltScreen = true
	return v
}

func formatGame(g backend.Game) item {
	scoreboard := fmt.Sprintf("%-13s %3d - %-3d %13s", g.AwayTeam, g.AwayScore, g.HomeScore, g.HomeTeam)
	pad := (len(scoreboard) - len(g.GameClock)) / 2
	clock := fmt.Sprintf("%*s", pad+len(g.GameClock), g.GameClock)
	return item{title: scoreboard, desc: clock}
}

func mergeGames(live, scheduled []backend.Game) []backend.Game {
	seen := make(map[string]bool)
	merged := make([]backend.Game, 0, len(live)+len(scheduled))
	for _, g := range live {
		seen[g.GameId] = true
		merged = append(merged, g)
	}
	for _, g := range scheduled {
		if !seen[g.GameId] {
			merged = append(merged, g)
		}
	}
	return merged
}

func main() {
	games := mergeGames(backend.GetLiveGames(), backend.GetScheduledGames())

	items := make([]list.Item, len(games))
	for i, g := range games {
		items[i] = formatGame(g)
	}

	m := model{list: list.New(items, list.NewDefaultDelegate(), 0, 0)}
	m.list.Title = "NBA Scores"

	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
