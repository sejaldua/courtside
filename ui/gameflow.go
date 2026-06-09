package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

type gameFlowSummary struct {
	leadChanges int
	timesTied   int
	biggestAway int
	biggestHome int
}

func computeFlowSummary(plays []playEvent) gameFlowSummary {
	if len(plays) == 0 {
		return gameFlowSummary{}
	}

	var s gameFlowSummary
	prevLeader := 0

	for i := len(plays) - 1; i >= 0; i-- {
		p := plays[i]
		if p.scoreAway == 0 && p.scoreHome == 0 {
			continue
		}

		diff := p.scoreHome - p.scoreAway
		var leader int
		switch {
		case diff > 0:
			leader = 1
		case diff < 0:
			leader = -1
		default:
			leader = 0
		}

		if leader == 0 && prevLeader != 0 {
			s.timesTied++
		} else if leader != 0 && prevLeader != 0 && leader != prevLeader {
			s.leadChanges++
		} else if leader != 0 && prevLeader == 0 {
			s.leadChanges++
		}
		prevLeader = leader

		if awayLead := p.scoreAway - p.scoreHome; awayLead > s.biggestAway {
			s.biggestAway = awayLead
		}
		if homeLead := p.scoreHome - p.scoreAway; homeLead > s.biggestHome {
			s.biggestHome = homeLead
		}
	}
	return s
}

func renderFlowSummary(s gameFlowSummary, awayTri, homeTri string) string {
	if s.biggestAway == 0 && s.biggestHome == 0 {
		return ""
	}

	bigAway := "N/A"
	if s.biggestAway > 0 {
		bigAway = fmt.Sprintf("%s: %d", awayTri, s.biggestAway)
	}
	bigHome := "N/A"
	if s.biggestHome > 0 {
		bigHome = fmt.Sprintf("%s: %d", homeTri, s.biggestHome)
	}

	lines := []string{
		fmt.Sprintf("Lead Changes: %d", s.leadChanges),
		fmt.Sprintf("Times Tied: %d", s.timesTied),
		fmt.Sprintf("Biggest Lead: %s, %s", bigAway, bigHome),
	}
	return mutedSty.Render(strings.Join(lines, "\n"))
}

func elapsedMinutes(period int, clock string) float64 {
	var m, s int
	fmt.Sscanf(clock, "%d:%d", &m, &s)
	remaining := float64(m) + float64(s)/60.0

	if period <= 4 {
		return float64(period-1)*12 + (12 - remaining)
	}
	return 48 + float64(period-5)*5 + (5 - remaining)
}

func totalGameMinutes(plays []playEvent) float64 {
	maxPeriod := 4
	for _, p := range plays {
		if p.period > maxPeriod {
			maxPeriod = p.period
		}
	}
	if maxPeriod <= 4 {
		return 48
	}
	return 48 + float64(maxPeriod-4)*5
}

// renderGameFlow builds a plaintextsports-style game flow chart.
// ':' = 3 points of lead (2 stacked dots), '.' = 1-2 point partial.
// Axis: +===========+===========+===========+===========+ (49 chars).
// Dots are color-coded by which team is winning at each minute.
func renderGameFlow(plays []playEvent, awayTri, homeTri string, width int) string {
	if len(plays) == 0 || width < 20 {
		return ""
	}

	numPeriods := 4
	for _, p := range plays {
		if p.period > numPeriods {
			numPeriods = p.period
		}
	}

	// 12 data columns per quarter, 48 total.
	// Axis = 49 chars: +===========+===========+===========+===========+
	// Leading '+' has no data; remaining 48 positions (11 '=' + 1 '+' per quarter) all get dots.
	labelW := 5
	chartW := 12 * numPeriods

	totalMin := totalGameMinutes(plays)
	margins := make([]int, chartW)
	filled := make([]bool, chartW)

	for i := len(plays) - 1; i >= 0; i-- {
		p := plays[i]
		if p.scoreAway == 0 && p.scoreHome == 0 {
			continue
		}
		elapsed := elapsedMinutes(p.period, p.clock)
		col := int(elapsed / totalMin * float64(chartW))
		if col >= chartW {
			col = chartW - 1
		}
		if col < 0 {
			col = 0
		}
		margins[col] = p.scoreAway - p.scoreHome
		filled[col] = true
	}

	lastMargin := 0
	for i := 0; i < chartW; i++ {
		if filled[i] {
			lastMargin = margins[i]
		} else {
			margins[i] = lastMargin
		}
	}

	lastDataCol := 0
	for i := len(plays) - 1; i >= 0; i-- {
		elapsed := elapsedMinutes(plays[i].period, plays[i].clock)
		col := int(elapsed / totalMin * float64(chartW))
		if col >= chartW {
			col = chartW - 1
		}
		if col > lastDataCol {
			lastDataCol = col
		}
	}

	const ptsPerRow = 3
	maxAway, maxHome := 0, 0
	for i := 0; i <= lastDataCol; i++ {
		if margins[i] > maxAway {
			maxAway = margins[i]
		}
		if -margins[i] > maxHome {
			maxHome = -margins[i]
		}
	}
	awayRows := (maxAway + ptsPerRow - 1) / ptsPerRow
	homeRows := (maxHome + ptsPerRow - 1) / ptsPerRow
	if awayRows < 1 && maxAway > 0 {
		awayRows = 1
	}
	if homeRows < 1 && maxHome > 0 {
		homeRows = 1
	}

	lines := make([]string, 0, awayRows+homeRows+5)

	// Header.
	axisW := chartW + 1 // 48 data positions + 1 leading '+' = 49 axis chars
	header := lipgloss.NewStyle().Bold(true).Render("Game Flow:")
	legend := mutedSty.Render("(2 dots = 3 points)")
	headerTotalW := axisW + labelW
	gap := headerTotalW - lipgloss.Width(header) - lipgloss.Width(legend)
	if gap < 2 {
		gap = 2
	}
	lines = append(lines, header+strings.Repeat(" ", gap)+legend)
	lines = append(lines, "")

	// Team colors for dot rendering.
	awayC := teamColorOrDefault(awayTri, awayColor)
	homeC := teamColorOrDefault(homeTri, homeColor)
	awaySty := lipgloss.NewStyle().Foreground(awayC)
	homeSty := lipgloss.NewStyle().Foreground(homeC)

	// Away rows (above axis): ':' = full 3-pt row, '.' = partial.
	// Dots colored in away team's color. One char per data column, offset by 1 for leading '+'.
	for row := awayRows; row >= 1; row-- {
		var line strings.Builder
		line.WriteString(strings.Repeat(" ", labelW))
		for col := 0; col < chartW; col++ {
			if col > lastDataCol {
				line.WriteString(" ")
				continue
			}
			m := margins[col]
			if m <= 0 {
				line.WriteString(" ")
				continue
			}
			fullRows := m / ptsPerRow
			remainder := m % ptsPerRow
			if row <= fullRows {
				line.WriteString(awaySty.Render(":"))
			} else if row == fullRows+1 && remainder > 0 {
				line.WriteString(awaySty.Render("."))
			} else {
				line.WriteString(" ")
			}
		}
		lines = append(lines, line.String())
	}

	// Axis line.
	awayLabel := lipgloss.NewStyle().Bold(true).Foreground(awayC).
		Render(fmt.Sprintf("%-4s", awayTri))
	homeLabel := lipgloss.NewStyle().Bold(true).Foreground(homeC).
		Render(fmt.Sprintf("%-4s", homeTri))

	var fullAxis strings.Builder
	fullAxis.WriteString("+")
	for q := 0; q < numPeriods; q++ {
		fullAxis.WriteString(strings.Repeat("=", 11))
		fullAxis.WriteString("+")
	}
	axisStr := fullAxis.String() // 49 chars total

	// Progress: lastDataCol (0-47) maps to axis position lastDataCol+1.
	progressIdx := lastDataCol + 2
	if progressIdx > len(axisStr) {
		progressIdx = len(axisStr)
	}

	playedSty := lipgloss.NewStyle().Bold(true)
	axisLine := awayLabel + " " +
		playedSty.Render(axisStr[:progressIdx]) +
		dimRowSty.Render(axisStr[progressIdx:])
	lines = append(lines, axisLine)

	// Home rows (below axis): ':' = full 3-pt row, '.' = partial.
	// Dots colored in home team's color. First row gets the home label.
	if homeRows == 0 {
		lines = append(lines, homeLabel)
	}
	for row := 1; row <= homeRows; row++ {
		var line strings.Builder
		if row == 1 {
			line.WriteString(homeLabel + " ")
		} else {
			line.WriteString(strings.Repeat(" ", labelW))
		}
		for col := 0; col < chartW; col++ {
			if col > lastDataCol {
				line.WriteString(" ")
				continue
			}
			m := margins[col]
			if m >= 0 {
				line.WriteString(" ")
				continue
			}
			homeLead := -m
			fullRows := homeLead / ptsPerRow
			remainder := homeLead % ptsPerRow
			if row <= fullRows {
				line.WriteString(homeSty.Render(":"))
			} else if row == fullRows+1 && remainder > 0 {
				line.WriteString(homeSty.Render("."))
			} else {
				line.WriteString(" ")
			}
		}
		lines = append(lines, line.String())
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
