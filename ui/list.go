package ui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
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
	list     list.Model
	day      time.Time // the date currently being viewed
	seq      int       // bumped on each day change; debounces rapid navigation
	loading  bool
	err      error
	entering bool // "go to date" input mode
	input    textinput.Model
	height   int // terminal height, for bottom-anchoring the empty state
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
	// Keep the list sized in every mode.
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.height = ws.Height
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(ws.Width-h, ws.Height-v)
	}

	// While typing a date, the input owns every message (keys to edit, others
	// to drive the cursor blink).
	if m.entering {
		if key, ok := msg.(tea.KeyPressMsg); ok {
			if key.String() == "ctrl+c" {
				return m, tea.Quit
			}
			return m.updateDateInput(key)
		}
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		// Day navigation (unless typing a filter).
		if m.list.FilterState() != list.Filtering {
			switch msg.String() {
			case "d":
				return m.startDateInput()
			case "left", "h":
				return m.gotoDay(-1)
			case "right", "l":
				return m.gotoDay(1)
			}
		}
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
	next := m.day.AddDate(0, 0, delta)
	// Don't navigate past today (ET) — there are no future results to show.
	if delta > 0 && afterToday(next) {
		return m, nil
	}
	m.day = next
	m.loading = true
	m.err = nil
	m.seq++
	m.list.Title = m.title()
	seq := m.seq
	return m, tea.Tick(fetchDebounce, func(time.Time) tea.Msg {
		return daySettledMsg{seq: seq}
	})
}

// startDateInput enters "go to date" mode, focusing a fresh text input.
func (m gamelist) startDateInput() (gamelist, tea.Cmd) {
	m.entering = true
	m.input.Reset()
	cmd := m.input.Focus()
	return m, cmd
}

// updateDateInput handles keys while typing a date: Enter jumps, Esc cancels,
// everything else edits the field.
func (m gamelist) updateDateInput(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.entering = false
		m.input.Blur()
		return m, nil
	case "enter":
		day, ok := parseDate(m.input.Value())
		m.entering = false
		m.input.Blur()
		if !ok {
			return m, nil // invalid input; stay on the current day
		}
		return m.jumpToDay(day)
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// jumpToDay switches directly to an absolute date and fetches it immediately
// (no debounce — the user explicitly asked for this date).
func (m gamelist) jumpToDay(day time.Time) (gamelist, tea.Cmd) {
	m.day = day
	m.loading = true
	m.err = nil
	m.seq++ // invalidate any pending day-navigation debounce ticks
	m.list.Title = m.title()
	return m, fetchGames(m.day)
}

// parseDate accepts a few common date formats; year-less forms assume the
// current (ET) year.
func parseDate(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, false
	}
	for _, layout := range []string{"2006-01-02", "1/2/2006", "01/02/2006", "Jan 2 2006", "Jan 2, 2006"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	for _, layout := range []string{"1/2", "01/02", "Jan 2"} {
		if t, err := time.Parse(layout, s); err == nil {
			return time.Date(nbaToday().Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), true
		}
	}
	return time.Time{}, false
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

// afterToday reports whether t's calendar date is after today (ET). Both sides
// are normalized to their calendar date so the comparison ignores time of day
// and location differences.
func afterToday(t time.Time) bool {
	today := nbaToday()
	ty, tm, td := t.Date()
	ny, nm, nd := today.Date()
	return time.Date(ty, tm, td, 0, 0, 0, 0, time.UTC).
		After(time.Date(ny, nm, nd, 0, 0, 0, 0, time.UTC))
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
	// While entering a date, show the input prompt in the title area. Drop the
	// Title chip's background so the prompt and typed text render with one
	// consistent style (this mirrors how the list draws its own filter input).
	if m.entering {
		m.list.Styles.Title = lipgloss.NewStyle()
		m.list.Title = m.input.View()
	}

	var content string
	if len(m.list.Items()) == 0 {
		// Hide the list entirely when empty; show just the title and a single
		// message (the bubbles list would otherwise render "No items." twice).
		note := "No games on this date."
		if m.loading {
			note = "Loading…"
		}
		top := lipgloss.JoinVertical(
			lipgloss.Left,
			m.list.Styles.Title.Render(m.list.Title),
			"",
			lipgloss.NewStyle().Faint(true).Render(note),
		)
		hint := renderHints(
			[2]string{"←/h", "prev day"},
			[2]string{"→/l", "next day"},
			[2]string{"d", "date"},
			[2]string{"q", "quit"},
		)

		// Push the hint to the bottom of the screen.
		_, vFrame := docStyle.GetFrameSize()
		spacerH := m.height - vFrame - lipgloss.Height(top) - lipgloss.Height(hint)
		if spacerH < 1 {
			spacerH = 1
		}
		spacer := lipgloss.NewStyle().Height(spacerH).Render("")
		content = lipgloss.JoinVertical(lipgloss.Left, top, spacer, hint)
	} else {
		content = m.list.View()
	}

	v := tea.NewView(docStyle.Render(content))
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

	ti := textinput.New()
	ti.Prompt = "Go to date (YYYY-MM-DD): "
	ti.CharLimit = 16

	m := gamelist{
		list:  list.New(items, list.NewDefaultDelegate(), 0, 0),
		day:   nbaToday(),
		input: ti,
	}

	// Repurpose ←/→ and h/l for day navigation; keep paging on pgup/pgdown and
	// free up "d" (a default next-page key) for the date jump.
	m.list.KeyMap.PrevPage.SetKeys("pgup", "b")
	m.list.KeyMap.NextPage.SetKeys("pgdown", "f")
	dayKeys := []key.Binding{
		key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "prev day")),
		key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "next day")),
		key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "date")),
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
