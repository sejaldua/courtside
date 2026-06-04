package ui

import (
	"fmt"
	"image/color"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// ---- data ----------------------------------------------------------------
//
// Dummy data for now. The shapes mirror what the SDK's BoxScoreTraditionalV3
// and PlayByPlayV3 endpoints provide, so real data can slot in later.

type playerStat struct {
	name                              string
	pts, ast, reb, blk, to, plusMinus int
}

// recorded reports whether the player has any box-score contribution. Players
// with an empty line are shown dimmed.
func (p playerStat) recorded() bool {
	return p.pts != 0 || p.ast != 0 || p.reb != 0 || p.blk != 0 || p.to != 0
}

type teamBox struct {
	name    string
	tricode string
	score   int
	players []playerStat
}

type playEvent struct {
	period int
	clock  string // "8:44"
	team   string // tricode; empty for neutral events
	desc   string
}

type gameDetail struct {
	away, home teamBox
	plays      []playEvent // newest first
}

// ---- styles --------------------------------------------------------------

var (
	awayColor    = lipgloss.Color("39")  // blue
	homeColor    = lipgloss.Color("203") // red
	mutedColor   = lipgloss.Color("245")
	dimColor     = lipgloss.Color("240")
	headerBarSty = lipgloss.NewStyle().Bold(true).Padding(0, 1).
			Border(lipgloss.RoundedBorder()).BorderForeground(dimColor)
	panelSty = lipgloss.NewStyle().Padding(0, 1).
			Border(lipgloss.RoundedBorder()).BorderForeground(dimColor)
	colHeaderSty = lipgloss.NewStyle().Bold(true).Foreground(mutedColor)
	dimRowSty    = lipgloss.NewStyle().Foreground(dimColor)
	hintSty      = lipgloss.NewStyle().Foreground(mutedColor)
)

// stat-table column widths
const (
	wName = 14
	wNum  = 4 // PTS, AST, REB, BLK, TO
	wPM   = 5 // +/-
)

// ---- model ---------------------------------------------------------------

type detail struct {
	game          gameDetail
	scroll        int // play-by-play scroll offset (0 = newest at top)
	width, height int
}

func newDetail(width, height int) detail {
	return detail{game: dummyGame(), width: width, height: height}
}

func (m detail) Init() tea.Cmd {
	return nil
}

func (m detail) Update(msg tea.Msg) (detail, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "k":
			if m.scroll > 0 {
				m.scroll--
			}
		case "down", "j":
			if m.scroll < m.maxScroll() {
				m.scroll++
			}
		}
	}
	return m, nil
}

// maxScroll is the furthest the play-by-play feed can scroll back.
func (m detail) maxScroll() int {
	max := len(m.game.plays) - m.pbpVisible()
	if max < 0 {
		return 0
	}
	return max
}

// pbpVisible is how many play-by-play lines fit at once, given the terminal
// height left over after the header and stat tables.
func (m detail) pbpVisible() int {
	v := m.height - 24 // rough reserve for header + tables + chrome
	if v < 6 {
		v = 6
	}
	if v > 30 {
		v = 30
	}
	return v
}

// ---- view ----------------------------------------------------------------

func (m detail) View() tea.View {
	hFrame, _ := docStyle.GetFrameSize()
	width := m.width - hFrame
	if width < 40 {
		width = 40
	}

	sections := lipgloss.JoinVertical(
		lipgloss.Left,
		m.renderHeader(width),
		"",
		m.renderTables(width),
		"",
		m.renderPlayByPlay(width),
		"",
		hintSty.Render("q back · ↑/↓ or j/k scroll"),
	)

	v := tea.NewView(docStyle.Render(sections))
	v.AltScreen = true
	return v
}

// renderHeader is a centered scoreboard bar using the same away/score/home
// layout as the list view.
func (m detail) renderHeader(width int) string {
	g := m.game
	line := fmt.Sprintf("%-13s %3d - %-3d %13s",
		g.away.name, g.away.score, g.home.score, g.home.name)
	bar := headerBarSty.Render(line)
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, bar)
}

// renderTables renders the two team stat tables side by side.
func (m detail) renderTables(width int) string {
	gap := "  "
	away := m.renderTeamTable(m.game.away, awayColor)
	home := m.renderTeamTable(m.game.home, homeColor)
	tables := lipgloss.JoinHorizontal(lipgloss.Top, away, gap, home)
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, tables)
}

func (m detail) renderTeamTable(t teamBox, teamColor color.Color) string {
	title := lipgloss.NewStyle().Bold(true).Foreground(teamColor).
		Render(fmt.Sprintf("%s (%d)", t.name, t.score))

	header := colHeaderSty.Render(
		pad("PLAYER", wName, false) + " " +
			pad("PTS", wNum, true) + " " +
			pad("AST", wNum, true) + " " +
			pad("REB", wNum, true) + " " +
			pad("BLK", wNum, true) + " " +
			pad("TO", wNum, true) + " " +
			pad("+/-", wPM, true),
	)

	players := append([]playerStat(nil), t.players...)
	sort.SliceStable(players, func(i, j int) bool {
		return players[i].pts > players[j].pts
	})

	rows := make([]string, 0, len(players)+2)
	rows = append(rows, title, header)
	for _, p := range players {
		row := pad(truncate(p.name, wName), wName, false) + " " +
			pad(fmt.Sprintf("%d", p.pts), wNum, true) + " " +
			pad(fmt.Sprintf("%d", p.ast), wNum, true) + " " +
			pad(fmt.Sprintf("%d", p.reb), wNum, true) + " " +
			pad(fmt.Sprintf("%d", p.blk), wNum, true) + " " +
			pad(fmt.Sprintf("%d", p.to), wNum, true) + " " +
			pad(signed(p.plusMinus), wPM, true)
		if !p.recorded() {
			row = dimRowSty.Render(row)
		}
		rows = append(rows, row)
	}

	return panelSty.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

// renderPlayByPlay renders the scrollable feed, newest first.
func (m detail) renderPlayByPlay(width int) string {
	visible := m.pbpVisible()
	plays := m.game.plays
	start := m.scroll
	if start > m.maxScroll() {
		start = m.maxScroll()
	}
	end := start + visible
	if end > len(plays) {
		end = len(plays)
	}

	lines := make([]string, 0, visible+1)
	title := colHeaderSty.Render("PLAY-BY-PLAY")
	if start > 0 {
		title += hintSty.Render(fmt.Sprintf("   ↑ %d earlier", start))
	}
	lines = append(lines, title)

	for _, p := range plays[start:end] {
		lines = append(lines, renderPlay(p, m.game.home.tricode))
	}

	body := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return panelSty.Width(width - 2).Render(body)
}

func renderPlay(p playEvent, homeTricode string) string {
	when := fmt.Sprintf("Q%d %5s", p.period, p.clock)
	when = lipgloss.NewStyle().Foreground(mutedColor).Render(pad(when, 9, false))

	if p.team == "" {
		// Neutral event (timeout, foul, etc.) — muted, no team tag.
		return when + "  " + lipgloss.NewStyle().Foreground(dimColor).Italic(true).Render(p.desc)
	}

	tagColor := awayColor
	if p.team == homeTricode {
		tagColor = homeColor
	}
	tag := lipgloss.NewStyle().Bold(true).Foreground(tagColor).Render(fmt.Sprintf("[%s]", p.team))
	return when + "  " + pad(tag, lipgloss.Width(tag), false) + "  " + p.desc
}

// ---- helpers -------------------------------------------------------------

// pad left- or right-aligns s within width w (right=true right-aligns).
func pad(s string, w int, right bool) string {
	gap := w - lipgloss.Width(s)
	if gap <= 0 {
		return s
	}
	if right {
		return strings.Repeat(" ", gap) + s
	}
	return s + strings.Repeat(" ", gap)
}

func truncate(s string, w int) string {
	if len(s) <= w {
		return s
	}
	if w <= 1 {
		return s[:w]
	}
	return s[:w-1] + "…"
}

func signed(n int) string {
	if n > 0 {
		return fmt.Sprintf("+%d", n)
	}
	return fmt.Sprintf("%d", n)
}
