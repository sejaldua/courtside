package ui

import (
	"fmt"
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/NolanFogarty/courtside/backend"
)

// ---- model ---------------------------------------------------------------

type standings struct {
	loading       bool
	err           error
	data          backend.Standings
	width, height int
}

func newStandings(width, height int) standings {
	return standings{loading: true, width: width, height: height}
}

// standingsLoadedMsg carries the result of the async GetStandings fetch.
type standingsLoadedMsg struct {
	data backend.Standings
	err  error
}

func (m standings) Init() tea.Cmd {
	return func() tea.Msg {
		d, err := backend.GetStandings()
		return standingsLoadedMsg{data: d, err: err}
	}
}

func (m standings) Update(msg tea.Msg) (standings, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case standingsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.data = msg.data
		}
	}
	return m, nil
}

// ---- view ----------------------------------------------------------------

func (m standings) View() tea.View {
	hFrame, vFrame := docStyle.GetFrameSize()
	width := m.width - hFrame
	if width < 40 {
		width = 40
	}

	titleText := "NBA Standings"
	if m.data.Season != "" {
		titleText += " · " + m.data.Season
	}
	title := lipgloss.PlaceHorizontal(width, lipgloss.Center, headerBarSty.Render(titleText))

	var main string
	switch {
	case m.err != nil:
		main = lipgloss.PlaceHorizontal(width, lipgloss.Center,
			errSty.Render("Failed to load standings: "+m.err.Error()))
	case m.loading:
		main = lipgloss.PlaceHorizontal(width, lipgloss.Center,
			mutedSty.Render("Loading standings…"))
	default:
		east := renderConference("EASTERN", awayColor, m.data.East)
		west := renderConference("WESTERN", homeColor, m.data.West)
		tables := lipgloss.JoinHorizontal(lipgloss.Top, east, "  ", west)
		main = lipgloss.PlaceHorizontal(width, lipgloss.Center, tables)
	}

	body := lipgloss.JoinVertical(lipgloss.Left, title, "", main)
	hint := renderHints([2]string{"q", "back"})

	spacerH := m.height - vFrame - lipgloss.Height(body) - lipgloss.Height(hint)
	if spacerH < 1 {
		spacerH = 1
	}
	spacer := lipgloss.NewStyle().Height(spacerH).Render("")

	sections := lipgloss.JoinVertical(lipgloss.Left, body, spacer, hint)
	v := tea.NewView(docStyle.Render(sections))
	v.AltScreen = true
	return v
}

// standings-table column widths
const (
	wRank   = 2
	wClinch = 2 // clinch marker (x, y, e, …)
	wTeam   = 16
	wWL     = 3
	wPct    = 5
	wGB     = 4
	wL10    = 5
	wStrk   = 4
)

func renderConference(title string, teamColor color.Color, rows []backend.Standing) string {
	header := pad("#", wRank, true) + " " + pad("", wClinch, false) + " " +
		pad("TEAM", wTeam, false) + " " +
		pad("W", wWL, true) + " " + pad("L", wWL, true) + " " +
		pad("PCT", wPct, true) + " " + pad("GB", wGB, true) + " " +
		pad("L10", wL10, true) + " " + pad("STRK", wStrk, true)

	titleLine := lipgloss.PlaceHorizontal(lipgloss.Width(header), lipgloss.Center,
		lipgloss.NewStyle().Bold(true).Foreground(teamColor).Render(title))

	lines := []string{
		titleLine,
		colHeaderSty.Render(header),
	}
	for _, s := range rows {
		lines = append(lines,
			pad(num(s.Rank), wRank, true)+" "+
				accentSty.Render(pad(s.Clinch, wClinch, false))+" "+
				pad(truncate(s.Team, wTeam), wTeam, false)+" "+
				pad(num(s.Wins), wWL, true)+" "+
				pad(num(s.Losses), wWL, true)+" "+
				pad(pctStr(s.PCT), wPct, true)+" "+
				pad(gbStr(s.GamesBack), wGB, true)+" "+
				pad(s.L10, wL10, true)+" "+
				pad(s.Streak, wStrk, true))
	}

	return panelSty.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

// pctStr renders a win percentage NBA-style: ".652", "1.000".
func pctStr(p float64) string {
	s := fmt.Sprintf("%.3f", p)
	return strings.TrimPrefix(s, "0")
}

// gbStr renders games-back: "-" for the conference leader, else one decimal.
func gbStr(gb float64) string {
	if gb == 0 {
		return "-"
	}
	return fmt.Sprintf("%.1f", gb)
}
