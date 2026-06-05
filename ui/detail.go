package ui

import (
	"fmt"
	"image/color"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/NolanFogarty/courtside/backend"
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

	// aggregate stats shown in the comparison bar
	fgPct, fg3Pct float64
	ftm, fta      int
	reb, ast, to  int
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

// statRow is one row of the team-comparison table. away/home are the display
// strings; awayKey/homeKey are the numeric values used to decide which side to
// highlight. When lowerBetter is true the smaller value is the "leader".
type statRow struct {
	label            string
	away, home       string
	awayKey, homeKey float64
	lowerBetter      bool
}

// leader reports which side to highlight. Ties highlight neither.
func (r statRow) leader() (away, home bool) {
	if r.awayKey == r.homeKey {
		return false, false
	}
	awayWins := r.awayKey > r.homeKey
	if r.lowerBetter {
		awayWins = r.awayKey < r.homeKey
	}
	return awayWins, !awayWins
}

// ---- styles --------------------------------------------------------------

var (
	awayColor    = lipgloss.Color("39")  // blue
	homeColor    = lipgloss.Color("203") // red
	mutedColor   = lipgloss.Color("245")
	dimColor     = lipgloss.Color("240")
	accentColor  = lipgloss.Color("214")
	headerBarSty = lipgloss.NewStyle().Bold(true).Padding(0, 1).
			Border(lipgloss.RoundedBorder()).BorderForeground(dimColor)
	panelSty = lipgloss.NewStyle().Padding(0, 1).
			Border(lipgloss.RoundedBorder()).BorderForeground(dimColor)
	colHeaderSty = lipgloss.NewStyle().Bold(true).Foreground(mutedColor)
	dimRowSty    = lipgloss.NewStyle().Foreground(dimColor)
	mutedSty     = lipgloss.NewStyle().Foreground(mutedColor)
	accentSty    = lipgloss.NewStyle().Bold(true).Foreground(accentColor)
	hintSty      = lipgloss.NewStyle().Foreground(mutedColor)
	errSty       = lipgloss.NewStyle().Bold(true).Foreground(homeColor)
)

// stat-table column widths
const (
	wName = 14
	wNum  = 4 // PTS, AST, REB, BLK, TO
	wPM   = 5 // +/-
)

// ---- model ---------------------------------------------------------------

type detail struct {
	gameID        string
	game          gameDetail
	loading       bool
	err           error
	scroll        int // play-by-play scroll offset (0 = newest at top)
	width, height int
}

// newDetail seeds the header from the list selection (team names and scores are
// known immediately) and marks the box score / play-by-play as still loading.
func newDetail(g backend.Game, width, height int) detail {
	d := detail{gameID: g.GameId, loading: true, width: width, height: height}
	d.game.away = teamBox{name: g.AwayTeam, score: g.AwayScore}
	d.game.home = teamBox{name: g.HomeTeam, score: g.HomeScore}
	return d
}

// gameDetailMsg carries the result of the async GetGameDetail fetch.
type gameDetailMsg struct {
	detail backend.GameDetail
	err    error
}

func (m detail) Init() tea.Cmd {
	id := m.gameID
	return func() tea.Msg {
		d, err := backend.GetGameDetail(id)
		return gameDetailMsg{detail: d, err: err}
	}
}

func (m detail) Update(msg tea.Msg) (detail, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case gameDetailMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.game = toGameDetail(msg.detail)
		}
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

// toGameDetail maps the backend's GameDetail into the view's render types.
func toGameDetail(d backend.GameDetail) gameDetail {
	g := gameDetail{away: toTeamBox(d.Away), home: toTeamBox(d.Home)}
	g.plays = make([]playEvent, len(d.Plays))
	for i, p := range d.Plays {
		g.plays[i] = playEvent{period: p.Period, clock: p.Clock, team: p.Team, desc: p.Desc}
	}
	return g
}

func toTeamBox(t backend.TeamDetail) teamBox {
	tb := teamBox{
		name: t.Name, tricode: t.Tricode, score: t.Score,
		fgPct: t.FGPct, fg3Pct: t.FG3Pct, ftm: t.FTM, fta: t.FTA,
		reb: t.Reb, ast: t.Ast, to: t.To,
	}
	tb.players = make([]playerStat, len(t.Players))
	for i, p := range t.Players {
		tb.players[i] = playerStat{
			name: p.Name, pts: p.Pts, ast: p.Ast, reb: p.Reb,
			blk: p.Blk, to: p.To, plusMinus: p.PlusMinus,
		}
	}
	return tb
}

// maxScroll is the furthest the play-by-play feed can scroll back.
func (m detail) maxScroll() int {
	max := len(m.game.plays) - m.pbpVisible()
	if max < 0 {
		return 0
	}
	return max
}

// leftColumnHeight is the rendered height of the left column (both player
// tables plus the team-stats bar between them).
func (m detail) leftColumnHeight() int {
	return lipgloss.Height(m.renderPlayerColumn())
}

// pbpVisible is how many play-by-play lines show at once: enough to fill the
// play-by-play box to the full height of the left column.
func (m detail) pbpVisible() int {
	v := m.leftColumnHeight() - 4 // box border (2) + title & divider (2)
	if v < 4 {
		v = 4
	}
	return v
}

// ---- view ----------------------------------------------------------------

func (m detail) View() tea.View {
	hFrame, vFrame := docStyle.GetFrameSize()
	width := m.width - hFrame
	if width < 40 {
		width = 40
	}

	var body string
	switch {
	case m.err != nil:
		body = lipgloss.JoinVertical(lipgloss.Left, m.renderHeader(width), "",
			lipgloss.PlaceHorizontal(width, lipgloss.Center,
				errSty.Render("Failed to load game: "+m.err.Error())))
	case m.loading:
		body = lipgloss.JoinVertical(lipgloss.Left, m.renderHeader(width), "",
			lipgloss.PlaceHorizontal(width, lipgloss.Center,
				mutedSty.Render("Loading game data…")))
	default:
		body = lipgloss.JoinVertical(lipgloss.Left, m.renderHeader(width), "", m.renderMain(width))
	}
	hint := hintSty.Render("↑/k up • ↓/j down • q back")

	// Push the hint to the bottom of the screen with a spacer that fills the
	// leftover height between the body and the hint.
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

// renderHeader is a centered scoreboard bar using the same away/score/home
// layout as the list view.
func (m detail) renderHeader(width int) string {
	g := m.game
	line := fmt.Sprintf("%-13s %3d - %-3d %13s",
		g.away.name, g.away.score, g.home.score, g.home.name)
	bar := headerBarSty.Render(line)
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, bar)
}

// renderMain lays out the two-column body: the player tables (with the team
// stats bar between them) stacked on the left, and the play-by-play feed
// filling the right, splitting the width down the middle.
func (m detail) renderMain(width int) string {
	const gap = 2
	colW := (width - gap) / 2

	left := m.renderPlayerColumn()
	right := m.renderInfoColumn(colW, lipgloss.Height(left))

	// Right-align the (fixed-width) left column so it sits against the center
	// divide; the play-by-play box fills the right half. Contents stay left-
	// aligned.
	leftCol := lipgloss.PlaceHorizontal(colW, lipgloss.Right, left)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftCol, "  ", right)
}

// renderPlayerColumn stacks the away player table, the horizontal team-stats
// bar, and the home player table vertically.
func (m detail) renderPlayerColumn() string {
	away := m.renderTeamTable(m.game.away, awayColor)
	bar := m.renderTeamBar()
	home := m.renderTeamTable(m.game.home, homeColor)
	return lipgloss.JoinVertical(lipgloss.Left, away, "", bar, "", home)
}

// renderInfoColumn renders the play-by-play box, sized to its own content width
// (bounded by w) and stretched to the full height of the left column.
func (m detail) renderInfoColumn(w, height int) string {
	maxW := w - 4 // box border (2) + padding (2)
	if maxW < 24 {
		maxW = 24
	}
	pbp := m.renderPBPColumn(m.pbpWidth(maxW))

	// Style.Height sets the box's total (border-inclusive) height, so target the
	// left column's height directly. Grow if the content would be taller.
	target := height
	if nat := lipgloss.Height(pbp) + 2; target < nat {
		target = nat
	}
	return panelSty.Height(target).Render(pbp)
}

// pbpWidth is the content width the play-by-play feed needs to show its longest
// event line without truncation, capped at max. Computed over all plays (not
// just the visible window) so the box width stays stable while scrolling.
func (m detail) pbpWidth(max int) int {
	maxDesc := 0
	for _, p := range m.game.plays {
		if dw := lipgloss.Width(p.desc); dw > maxDesc {
			maxDesc = dw
		}
	}
	w := wWhen + 2 + wTag + 2 + maxDesc
	if w > max {
		w = max
	}
	return w
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

// playerTableW is the width of a player-table row, matched by the team-stats
// bar so it spans the player tables exactly.
const playerTableW = wName + 5*wNum + wPM + 6 // 6 single-space separators

// team-stats bar layout: TEAM column + 3 shooting columns + 3
// team columns, summing to playerTableW.
const (
	barStatW    = 6
	barTeamW    = playerTableW - 6*barStatW
)

// barStatRows builds the six columns shown in the horizontal team-stats bar
// from the game's aggregate stats, in display order: shooting (FG%, 3P%, FTM)
// then team (REB, AST, TO).
func (m detail) barStatRows() []statRow {
	a, h := m.game.away, m.game.home
	return []statRow{
		pctRow("FG%", a.fgPct, h.fgPct),
		pctRow("3P%", a.fg3Pct, h.fg3Pct),
		ftmRow("FTM", a.ftm, a.fta, h.ftm, h.fta),
		intRow("REB", a.reb, h.reb, false),
		intRow("AST", a.ast, h.ast, false),
		intRow("TO", a.to, h.to, true), // fewer turnovers leads
	}
}

func pctRow(label string, a, h float64) statRow {
	return statRow{label: label,
		away: fmt.Sprintf("%.1f", a), home: fmt.Sprintf("%.1f", h),
		awayKey: a, homeKey: h}
}

func ftmRow(label string, am, aa, hm, ha int) statRow {
	return statRow{label: label,
		away: fmt.Sprintf("%d/%d", am, aa), home: fmt.Sprintf("%d/%d", hm, ha),
		awayKey: float64(am), homeKey: float64(hm)} // more made free throws leads
}

func intRow(label string, a, h int, lowerBetter bool) statRow {
	return statRow{label: label,
		away: fmt.Sprintf("%d", a), home: fmt.Sprintf("%d", h),
		awayKey: float64(a), homeKey: float64(h), lowerBetter: lowerBetter}
}

// renderTeamBar renders the horizontal team-stats bar that sits between the two
// player tables: a header row of column labels followed by one row per team.
func (m detail) renderTeamBar() string {
	rows := m.barStatRows()

	hdr := pad("TEAM", barTeamW, false)
	for i, r := range rows {
		if i == 3 { // gap separating the shooting and team groups
//			hdr += strings.Repeat(" ", barGroupGap)
		}
		hdr += pad(r.label, barStatW, true)
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		colHeaderSty.Render(hdr),
		m.barTeamRow(rows, m.game.away.tricode, awayColor, true),
		m.barTeamRow(rows, m.game.home.tricode, homeColor, false),
	)
	return panelSty.Render(content)
}

// barTeamRow renders one team's row of the bar, highlighting the cells where
// this team leads.
func (m detail) barTeamRow(rows []statRow, tricode string, teamColor color.Color, isAway bool) string {
	row := lipgloss.NewStyle().Bold(true).Foreground(teamColor).
		Render(pad(tricode, barTeamW, false))
	for _, r := range rows {
		val, leads := r.home, false
		awayLeads, homeLeads := r.leader()
		if isAway {
			val, leads = r.away, awayLeads
		} else {
			leads = homeLeads
		}
		cell := pad(val, barStatW, true)
		if leads {
			cell = accentSty.Render(cell)
		}
		row += cell
	}
	return row
}

// renderPBPColumn renders the scrollable play-by-play feed, newest first.
func (m detail) renderPBPColumn(w int) string {
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

	title := colHeaderSty.Render("PLAY-BY-PLAY")
	if start > 0 {
		title += hintSty.Render(fmt.Sprintf("   ↑ %d more", start))
	}
	lines := []string{title, dimRowSty.Render(strings.Repeat("─", w))}

	for _, p := range plays[start:end] {
		lines = append(lines, renderPlay(p, m.game.home.tricode, w))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// play-by-play line layout: "Q2 08:44" (8) + gap + "[NYK]" (5) + gap + desc
const (
	wWhen = 8
	wTag  = 5
)

func renderPlay(p playEvent, homeTricode string, width int) string {
	when := mutedSty.Render(pad(fmt.Sprintf("Q%d %s", p.period, mmss(p.clock)), wWhen, false))

	descW := width - wWhen - 2 - wTag - 2
	if descW < 4 {
		descW = 4
	}

	if p.team == "" {
		// Neutral event (timeout, jump ball, end of period) — fully muted.
		tag := dimRowSty.Render(pad("---", wTag, false))
		return when + "  " + tag + "  " + dimRowSty.Italic(true).Render(truncate(p.desc, descW))
	}

	tagColor := awayColor
	if p.team == homeTricode {
		tagColor = homeColor
	}
	tag := lipgloss.NewStyle().Bold(true).Foreground(tagColor).
		Render(pad("["+p.team+"]", wTag, false))
	return when + "  " + tag + "  " + truncate(p.desc, descW)
}

// mmss normalizes a "M:SS" clock to zero-padded "MM:SS".
func mmss(c string) string {
	var m, s int
	if _, err := fmt.Sscanf(c, "%d:%d", &m, &s); err != nil {
		return c
	}
	return fmt.Sprintf("%02d:%02d", m, s)
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
