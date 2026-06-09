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
	min                               string
	pts, ast, reb, blk, to, plusMinus int

	// expanded detail
	fgm, fga, tpm, tpa, ftm, fta int
	oreb, dreb, stl, pf          int
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

	// expanded detail
	fgm, fga, tpm, tpa       int
	oreb, dreb, stl, blk, pf int
}

type playEvent struct {
	period    int
	clock     string // "8:44"
	team      string // tricode; empty for neutral events
	desc      string
	scoreAway int
	scoreHome int
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
	accentColor  = lipgloss.Color("247")
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

	// help-bar styling matched to the bubbles list help: muted key/desc/separator
	// colors and the list's left padding, so all key hints look the same.
	helpKeySty  = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262"))
	helpDescSty = lipgloss.NewStyle().Foreground(lipgloss.Color("#4A4A4A"))
	helpSepSty  = lipgloss.NewStyle().Foreground(lipgloss.Color("#3C3C3C"))
	helpSty     = lipgloss.NewStyle().PaddingLeft(2)
)

// renderHints renders key/description pairs in the same muted colors, separator,
// and left padding as the bubbles list's help bar.
func renderHints(pairs ...[2]string) string {
	sep := helpSepSty.Render(" • ")
	parts := make([]string, len(pairs))
	for i, p := range pairs {
		parts[i] = helpKeySty.Render(p[0]) + " " + helpDescSty.Render(p[1])
	}
	return helpSty.Render(strings.Join(parts, sep))
}

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
	scheduled     bool   // game hasn't tipped off yet; no stats to fetch
	tipoff        string // start time, shown when scheduled
	live          bool   // game in progress; auto-refresh its data
	expanded      bool   // show detailed (expanded) stats
	scroll        int    // play-by-play scroll offset (0 = newest at top)
	width, height int
}

// newDetail seeds the header from the list selection (team names and scores are
// known immediately). A game that hasn't started has no box score / play-by-play
// to fetch, so we show its tip-off time instead of loading.
func newDetail(g backend.Game, width, height int) detail {
	d := detail{gameID: g.GameId, live: g.IsLive(), width: width, height: height}
	d.game.away = teamBox{name: g.AwayTeam, score: g.AwayScore}
	d.game.home = teamBox{name: g.HomeTeam, score: g.HomeScore}
	if g.NotStarted() {
		d.scheduled = true
		d.tipoff = g.GameClock
	} else {
		d.loading = true
	}
	return d
}

// refreshCmd re-fetches the box score and play-by-play, but only while the game
// is live. Existing scroll position and expanded state are preserved by Update.
func (m detail) refreshCmd() tea.Cmd {
	if !m.live {
		return nil
	}
	id := m.gameID
	return func() tea.Msg {
		d, err := backend.GetGameDetail(id)
		return gameDetailMsg{detail: d, err: err}
	}
}

// gameDetailMsg carries the result of the async GetGameDetail fetch.
type gameDetailMsg struct {
	detail backend.GameDetail
	err    error
}

func (m detail) Init() tea.Cmd {
	if m.scheduled {
		return nil // nothing to fetch until the game starts
	}
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
		wasLoading := m.loading
		m.loading = false
		if msg.err != nil {
			// Only surface an error on the initial load; ignore transient
			// auto-refresh failures and keep showing the last good data.
			if wasLoading {
				m.err = msg.err
			}
		} else {
			m.game = toGameDetail(msg.detail)
			m.err = nil
		}
	case tea.KeyPressMsg:
		switch msg.String() {
		case "o":
			m.expanded = !m.expanded
			if m.scroll > m.maxScroll() {
				m.scroll = m.maxScroll()
			}
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
		g.plays[i] = playEvent{
			period: p.Period, clock: p.Clock, team: p.Team, desc: p.Desc,
			scoreAway: p.ScoreAway, scoreHome: p.ScoreHome,
		}
	}
	return g
}

func toTeamBox(t backend.TeamDetail) teamBox {
	tb := teamBox{
		name: t.Name, tricode: t.Tricode, score: t.Score,
		fgPct: t.FGPct, fg3Pct: t.FG3Pct, ftm: t.FTM, fta: t.FTA,
		reb: t.Reb, ast: t.Ast, to: t.To,
		fgm: t.FGM, fga: t.FGA, tpm: t.TPM, tpa: t.TPA,
		oreb: t.OReb, dreb: t.DReb, stl: t.Stl, blk: t.Blk, pf: t.PF,
	}
	tb.players = make([]playerStat, len(t.Players))
	for i, p := range t.Players {
		tb.players[i] = playerStat{
			name: p.Name, min: p.Min, pts: p.Pts, ast: p.Ast, reb: p.Reb,
			blk: p.Blk, to: p.To, plusMinus: p.PlusMinus,
			fgm: p.FGM, fga: p.FGA, tpm: p.TPM, tpa: p.TPA, ftm: p.FTM, fta: p.FTA,
			oreb: p.OReb, dreb: p.DReb, stl: p.Stl, pf: p.PF,
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
// right panel after accounting for the game flow chart, summary, and borders.
func (m detail) pbpVisible() int {
	overhead := 4 // box border (2) + pbp title & divider (2)

	if len(m.game.plays) > 0 {
		// Estimate chart height: header(1) + blank(1) + away rows + axis(1) + home rows/label(1+)
		maxAway, maxHome := 0, 0
		for _, p := range m.game.plays {
			margin := p.scoreAway - p.scoreHome
			if margin > maxAway {
				maxAway = margin
			}
			if -margin > maxHome {
				maxHome = -margin
			}
		}
		awayRows := (maxAway + 2) / 3
		homeRows := (maxHome + 2) / 3
		if homeRows < 1 {
			homeRows = 1 // home label always shown
		}
		overhead += 2 + awayRows + 1 + homeRows + 1 // header, blank, away, axis, home, separator
	}

	// Summary: 3 lines + blank
	summary := computeFlowSummary(m.game.plays)
	if summary.biggestAway > 0 || summary.biggestHome > 0 {
		overhead += 4
	}

	v := m.leftColumnHeight() - overhead
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
	case m.scheduled:
		note := "Not started"
		if m.tipoff != "" {
			note = "Tip-off · " + m.tipoff
		}
		body = lipgloss.JoinVertical(lipgloss.Left, m.renderHeader(width), "",
			lipgloss.PlaceHorizontal(width, lipgloss.Center, mutedSty.Render(note)))
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

	// Only the loaded box score has scrollable / expandable content.
	hint := renderHints([2]string{"t", "theme"}, [2]string{"q", "back"})
	if !m.scheduled && !m.loading && m.err == nil {
		statsDesc := "more stats"
		if m.expanded {
			statsDesc = "less stats"
		}
		hint = renderHints(
			[2]string{"↑/k", "up"},
			[2]string{"↓/j", "down"},
			[2]string{"o", statsDesc},
			[2]string{"t", "theme"},
			[2]string{"q", "back"},
		)
	}

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
// layout as the list view, with per-team colors.
func (m detail) renderHeader(width int) string {
	g := m.game
	awayC := teamColorOrDefault(g.away.tricode, awayColor)
	homeC := teamColorOrDefault(g.home.tricode, homeColor)
	awaySty := lipgloss.NewStyle().Bold(true).Foreground(awayC)
	homeSty := lipgloss.NewStyle().Bold(true).Foreground(homeC)

	line := awaySty.Render(fmt.Sprintf("%-13s", g.away.name)) +
		fmt.Sprintf(" %3d - %-3d ", g.away.score, g.home.score) +
		homeSty.Render(fmt.Sprintf("%13s", g.home.name))
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
	leftW := lipgloss.Width(left)

	// Compact tables fit within half the screen, so keep the centered 50/50
	// feel by right-aligning the left column in its half. The wider expanded
	// tables claim the room they need and the feed takes whatever is left.
	if leftW <= colW {
		left = lipgloss.PlaceHorizontal(colW, lipgloss.Right, left)
		leftW = colW
	}
	rightW := width - leftW - gap
	if rightW < 24 {
		rightW = 24
	}
	right := m.renderInfoColumn(rightW, lipgloss.Height(left))

	return lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)
}

// renderPlayerColumn stacks the team-stats bar at the top, then the away and
// home player tables below. The team bar width matches the player tables.
func (m detail) renderPlayerColumn() string {
	awayC := teamColorOrDefault(m.game.away.tricode, awayColor)
	homeC := teamColorOrDefault(m.game.home.tricode, homeColor)
	away := m.renderTeamTable(m.game.away, awayC)
	home := m.renderTeamTable(m.game.home, homeC)
	tableW := lipgloss.Width(away)
	bar := m.renderTeamBar(tableW)
	return lipgloss.JoinVertical(lipgloss.Left, bar, "", away, "", home)
}

func teamColorOrDefault(tricode string, fallback color.Color) color.Color {
	if tricode == "" {
		return fallback
	}
	return teamColor(tricode)
}

// renderInfoColumn renders the game flow chart, lead/tie summary, and
// play-by-play stacked vertically in a bordered panel.
func (m detail) renderInfoColumn(w, height int) string {
	maxW := w - 4 // box border (2) + padding (2)
	if maxW < 24 {
		maxW = 24
	}

	sections := make([]string, 0, 4)

	// Game flow chart
	if chart := renderGameFlow(m.game.plays, m.game.away.tricode, m.game.home.tricode, maxW); chart != "" {
		sections = append(sections, chart)
	}

	// Lead/tie summary
	summary := computeFlowSummary(m.game.plays)
	if line := renderFlowSummary(summary, m.game.away.tricode, m.game.home.tricode); line != "" {
		sections = append(sections, line)
	}

	// Separator before play-by-play
	if len(sections) > 0 {
		sections = append(sections, "")
	}

	// Play-by-play
	sections = append(sections, m.renderPBPColumn(m.pbpWidth(maxW)))

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	target := height
	if nat := lipgloss.Height(content) + 2; target < nat {
		target = nat
	}
	return panelSty.Height(target).Render(content)
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
	if len(m.game.plays) == 0 {
		// Keep room for the "unavailable" note so the divider stays aligned.
		if nw := lipgloss.Width(pbpUnavailable); w < nw {
			w = nw
		}
	}
	if w > max {
		w = max
	}
	return w
}

// statCol is one numeric column of the player table: a header, a fixed width,
// and how to render a player's value.
type statCol struct {
	header string
	width  int
	value  func(p playerStat) string
}

func num(n int) string        { return fmt.Sprintf("%d", n) }
func ma(made, att int) string { return fmt.Sprintf("%d-%d", made, att) }

// playerCols returns the player-table columns for the current mode.
func (m detail) playerCols() []statCol {
	if m.expanded {
		return []statCol{
			{"MIN", 4, func(p playerStat) string { return p.min }},
			{"PTS", 4, func(p playerStat) string { return num(p.pts) }},
			{"FG", 6, func(p playerStat) string { return ma(p.fgm, p.fga) }},
			{"3P", 5, func(p playerStat) string { return ma(p.tpm, p.tpa) }},
			{"FT", 5, func(p playerStat) string { return ma(p.ftm, p.fta) }},
			{"REB", 4, func(p playerStat) string { return num(p.reb) }},
			{"AST", 4, func(p playerStat) string { return num(p.ast) }},
			{"STL", 4, func(p playerStat) string { return num(p.stl) }},
			{"BLK", 4, func(p playerStat) string { return num(p.blk) }},
			{"TO", 4, func(p playerStat) string { return num(p.to) }},
			{"PF", 4, func(p playerStat) string { return num(p.pf) }},
			{"+/-", 5, func(p playerStat) string { return signed(p.plusMinus) }},
		}
	}
	return []statCol{
		{"PTS", wNum, func(p playerStat) string { return num(p.pts) }},
		{"AST", wNum, func(p playerStat) string { return num(p.ast) }},
		{"REB", wNum, func(p playerStat) string { return num(p.reb) }},
		{"BLK", wNum, func(p playerStat) string { return num(p.blk) }},
		{"TO", wNum, func(p playerStat) string { return num(p.to) }},
		{"+/-", wPM, func(p playerStat) string { return signed(p.plusMinus) }},
	}
}

// tableWidth is the content width of a player table with the given columns:
// the name column plus each column preceded by a single-space separator.
func tableWidth(cols []statCol) int {
	w := wName
	for _, c := range cols {
		w += 1 + c.width
	}
	return w
}

// leadersLine builds a "★ Name: 34 PTS | Name: 11 AST | Name: 9 REB" summary.
func leadersLine(players []playerStat) string {
	if len(players) == 0 {
		return ""
	}

	type leader struct {
		name  string
		val   int
		label string
	}

	best := func(fn func(p playerStat) int, label string) leader {
		var top leader
		for _, p := range players {
			if v := fn(p); v > top.val {
				top = leader{name: shortName(p.name), val: v, label: label}
			}
		}
		return top
	}

	leaders := []leader{
		best(func(p playerStat) int { return p.pts }, "PTS"),
		best(func(p playerStat) int { return p.ast }, "AST"),
		best(func(p playerStat) int { return p.reb }, "REB"),
	}

	parts := make([]string, 0, 3)
	for _, l := range leaders {
		if l.val > 0 {
			parts = append(parts, fmt.Sprintf("%s: %d %s", l.name, l.val, l.label))
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return accentSty.Render("★ " + strings.Join(parts, " | "))
}

// shortName returns the last name (or full name if no space).
func shortName(name string) string {
	parts := strings.Fields(name)
	if len(parts) <= 1 {
		return name
	}
	// "L. James" style: first initial + last
	return string(parts[0][0]) + ". " + parts[len(parts)-1]
}

func (m detail) renderTeamTable(t teamBox, teamColor color.Color) string {
	cols := m.playerCols()

	title := lipgloss.NewStyle().Bold(true).Foreground(teamColor).
		Render(fmt.Sprintf("%s (%d)", t.name, t.score))

	hdr := pad("PLAYER", wName, false)
	for _, c := range cols {
		hdr += " " + pad(c.header, c.width, true)
	}

	players := append([]playerStat(nil), t.players...)
	sort.SliceStable(players, func(i, j int) bool {
		return players[i].pts > players[j].pts
	})

	rows := make([]string, 0, len(players)+3)
	rows = append(rows, title)
	if ldr := leadersLine(players); ldr != "" {
		rows = append(rows, ldr)
	}
	rows = append(rows, colHeaderSty.Render(hdr))
	for _, p := range players {
		row := pad(truncate(p.name, wName), wName, false)
		for _, c := range cols {
			row += " " + pad(c.value(p), c.width, true)
		}
		if !p.recorded() {
			row = dimRowSty.Render(row)
		}
		rows = append(rows, row)
	}

	return panelSty.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

// width of each team-stats bar stat column.
const barStatW = 6

// Dean Oliver's Four Factors helper computations.
func efgPct(fgm, tpm, fga int) float64 {
	if fga == 0 {
		return 0
	}
	return (float64(fgm) + 0.5*float64(tpm)) / float64(fga) * 100
}

func tovPct(to, fga, fta int) float64 {
	denom := float64(fga) + 0.44*float64(fta) + float64(to)
	if denom == 0 {
		return 0
	}
	return float64(to) / denom * 100
}

func orebPct(oreb, oppDreb int) float64 {
	total := oreb + oppDreb
	if total == 0 {
		return 0
	}
	return float64(oreb) / float64(total) * 100
}

func ftRate(ftm, fga int) float64 {
	if fga == 0 {
		return 0
	}
	return float64(ftm) / float64(fga) * 100
}

// barStatRows builds the columns shown in the horizontal team-stats bar.
// The number of stats is chosen to fill the player table width.
// Compact (7 stats): Four Factors + REB, AST, TO
// Expanded (12 stats): Shooting splits + Four Factors + counting stats
func (m detail) barStatRows() []statRow {
	a, h := m.game.away, m.game.home
	if m.expanded {
		// 12 stat columns to fill 79-char expanded player table width
		return []statRow{
			pctRow("FG%", a.fgPct, h.fgPct),
			pctRow("eFG%", efgPct(a.fgm, a.tpm, a.fga), efgPct(h.fgm, h.tpm, h.fga)),
			pctRow("3P%", a.fg3Pct, h.fg3Pct),
			pctRow("FTR", ftRate(a.ftm, a.fga), ftRate(h.ftm, h.fga)),
			{label: "TOV%", away: fmt.Sprintf("%.1f", tovPct(a.to, a.fga, a.fta)),
				home: fmt.Sprintf("%.1f", tovPct(h.to, h.fga, h.fta)),
				awayKey: tovPct(a.to, a.fga, a.fta), homeKey: tovPct(h.to, h.fga, h.fta),
				lowerBetter: true},
			pctRow("ORB%", orebPct(a.oreb, h.dreb), orebPct(h.oreb, a.dreb)),
			intRow("REB", a.reb, h.reb, false),
			intRow("AST", a.ast, h.ast, false),
			intRow("STL", a.stl, h.stl, false),
			intRow("BLK", a.blk, h.blk, false),
			intRow("TO", a.to, h.to, true),
			intRow("PF", a.pf, h.pf, true),
		}
	}
	// 7 stat columns to fill 50-char compact player table width
	return []statRow{
		pctRow("eFG%", efgPct(a.fgm, a.tpm, a.fga), efgPct(h.fgm, h.tpm, h.fga)),
		{label: "TOV%", away: fmt.Sprintf("%.1f", tovPct(a.to, a.fga, a.fta)),
			home: fmt.Sprintf("%.1f", tovPct(h.to, h.fga, h.fta)),
			awayKey: tovPct(a.to, a.fga, a.fta), homeKey: tovPct(h.to, h.fga, h.fta),
			lowerBetter: true},
		pctRow("ORB%", orebPct(a.oreb, h.dreb), orebPct(h.oreb, a.dreb)),
		pctRow("FTR", ftRate(a.ftm, a.fga), ftRate(h.ftm, h.fga)),
		intRow("REB", a.reb, h.reb, false),
		intRow("AST", a.ast, h.ast, false),
		intRow("TO", a.to, h.to, true),
	}
}

// maRow is a "made-attempted" comparison column; the team with more makes leads.
func maRow(label string, am, aa, hm, ha int) statRow {
	return statRow{label: label,
		away: fmt.Sprintf("%d-%d", am, aa), home: fmt.Sprintf("%d-%d", hm, ha),
		awayKey: float64(am), homeKey: float64(hm)}
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

// renderTeamBar renders the horizontal team-stats bar: a header row of column
// labels followed by one row per team. outerW is the total rendered width
// (including panelSty border/padding) to match.
func (m detail) renderTeamBar(outerW int) string {
	rows := m.barStatRows()
	// panelSty adds 2 chars of horizontal padding (Padding 0,1 on each side)
	// plus 2 chars of border (left + right) = 4 total.
	targetW := outerW - 4
	if targetW < 20 {
		targetW = tableWidth(m.playerCols())
	}

	// Size the TEAM column so that TEAM + stat columns = targetW.
	teamW := targetW - len(rows)*barStatW
	if teamW < 4 {
		teamW = 4
	}

	hdr := pad("TEAM", teamW, false)
	for _, r := range rows {
		hdr += pad(r.label, barStatW, true)
	}

	awayC := teamColorOrDefault(m.game.away.tricode, awayColor)
	homeC := teamColorOrDefault(m.game.home.tricode, homeColor)

	awayRow := m.barTeamRow(rows, m.game.away.tricode, awayC, true, teamW)
	homeRow := m.barTeamRow(rows, m.game.home.tricode, homeC, false, teamW)

	// Pad each row to targetW so the panel border aligns with the player tables.
	hdr = hdr + strings.Repeat(" ", max(0, targetW-lipgloss.Width(hdr)))
	awayRow = awayRow + strings.Repeat(" ", max(0, targetW-lipgloss.Width(awayRow)))
	homeRow = homeRow + strings.Repeat(" ", max(0, targetW-lipgloss.Width(homeRow)))

	content := lipgloss.JoinVertical(lipgloss.Left,
		colHeaderSty.Render(hdr),
		awayRow,
		homeRow,
	)
	return panelSty.Render(content)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// barTeamRow renders one team's row of the bar, highlighting the cells where
// this team leads using that team's color.
func (m detail) barTeamRow(rows []statRow, tricode string, tc color.Color, isAway bool, teamW int) string {
	teamSty := lipgloss.NewStyle().Bold(true).Foreground(tc)
	row := teamSty.Render(pad(tricode, teamW, false))
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
			cell = teamSty.Render(cell)
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

	if len(plays) == 0 {
		lines = append(lines, mutedSty.Render(pbpUnavailable))
	}
	for _, p := range plays[start:end] {
		lines = append(lines, renderPlay(p, m.game.home.tricode, w))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// pbpUnavailable is shown in the feed when no play-by-play could be loaded (the
// endpoint failed, or the game just tipped off).
const pbpUnavailable = "Play-by-play unavailable"

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
		tag := dimRowSty.Render(pad("---", wTag, false))
		return when + "  " + tag + "  " + dimRowSty.Italic(true).Render(truncate(p.desc, descW))
	}

	tagColor := teamColor(p.team)
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
