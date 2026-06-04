package ui

import (
	"fmt"
	"time"

	"charm.land/bubbles/v2/key"
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
	list    list.Model
	day     time.Time // the date currently being viewed
	seq     int       // bumped on each day change; debounces rapid navigation
	loading bool
	err     error
}

// daySettledMsg fires after a short pause following a day change; the fetch only
// runs if seq still matches (i.e. no newer keypress superseded it).
type daySettledMsg struct{ seq int }

// gamesLoadedMsg carries the result of an async day fetch.
type gamesLoadedMsg struct {
	day   time.Time
	games []backend.Game
	err   error
}

// fetchDebounce is how long to wait after the last day change before fetching.
const fetchDebounce = 250 * time.Millisecond

func (m gamelist) Init() tea.Cmd {
	return nil
}

func (m gamelist) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		// Step the viewed date, unless the user is typing a filter.
		if m.list.FilterState() != list.Filtering {
			switch msg.String() {
			case "left", "h":
				return m.gotoDay(-1)
			case "right", "l":
				return m.gotoDay(1)
			}
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	case daySettledMsg:
		if msg.seq != m.seq {
			return m, nil // superseded by a newer day change; don't fetch
		}
		return m, fetchGames(m.day)
	case gamesLoadedMsg:
		if !sameDay(msg.day, m.day) {
			return m, nil // a newer day was requested; ignore stale result
		}
		m.loading = false
		m.err = msg.err
		var cmd tea.Cmd
		if msg.err == nil {
			cmd = m.applyGames(msg.games)
		}
		m.list.Title = m.title()
		return m, cmd
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// gotoDay moves the viewed date by delta days immediately, but defers the fetch
// behind a debounce so spamming the key doesn't hammer the API.
func (m gamelist) gotoDay(delta int) (gamelist, tea.Cmd) {
	m.day = m.day.AddDate(0, 0, delta)
	m.loading = true
	m.err = nil
	m.seq++
	m.list.Title = m.title()
	seq := m.seq
	return m, tea.Tick(fetchDebounce, func(time.Time) tea.Msg {
		return daySettledMsg{seq: seq}
	})
}

func fetchGames(day time.Time) tea.Cmd {
	return func() tea.Msg {
		games, err := backend.GetGamesForDate(day.Format("2006-01-02"))
		return gamesLoadedMsg{day: day, games: games, err: err}
	}
}

// applyGames replaces the list's items with the games for the current day.
func (m *gamelist) applyGames(games []backend.Game) tea.Cmd {
	items := make([]list.Item, len(games))
	for i, g := range games {
		items[i] = formatGame(g)
	}
	return m.list.SetItems(items)
}

func (m gamelist) title() string {
	label := m.day.Format("Mon Jan 2")
	if sameDay(m.day, nbaToday()) {
		label = "Today · " + label
	}
	t := "NBA Scores — " + label
	switch {
	case m.loading:
		t += "  (loading…)"
	case m.err != nil:
		t += "  (failed to load)"
	}
	return t
}

func sameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.YearDay() == b.YearDay()
}

// nbaToday returns the current date in US Eastern time, which is what the NBA's
// schedule endpoints key game dates on. Falls back to local time if the zone
// database is unavailable.
func nbaToday() time.Time {
	if loc, err := time.LoadLocation("America/New_York"); err == nil {
		return time.Now().In(loc)
	}
	return time.Now()
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

	m := gamelist{
		list: list.New(items, list.NewDefaultDelegate(), 0, 0),
		day:  nbaToday(),
	}

	// Repurpose ←/→ and h/l for day navigation; keep paging on pgup/pgdown.
	m.list.KeyMap.PrevPage.SetKeys("pgup", "b")
	m.list.KeyMap.NextPage.SetKeys("pgdown", "f")
	dayKeys := []key.Binding{
		key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "prev day")),
		key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "next day")),
	}
	m.list.AdditionalShortHelpKeys = func() []key.Binding { return dayKeys }
	m.list.AdditionalFullHelpKeys = func() []key.Binding { return dayKeys }

	m.list.Title = m.title()
	return m
}

func Run(games []backend.Game) error {
	m := newRoot(games)
	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}
