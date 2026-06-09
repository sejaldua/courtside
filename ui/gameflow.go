package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

// gameFlowSummary computes lead changes, ties, and biggest leads from the
// play-by-play (which is stored newest-first).
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
	prevLeader := 0 // -1 away, 0 tie, 1 home

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

// renderFlowSummary renders the lead/tie summary as separate lines.
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

// elapsedMinutes converts a period + "MM:SS" clock into total elapsed minutes.
func elapsedMinutes(period int, clock string) float64 {
	var m, s int
	fmt.Sscanf(clock, "%d:%d", &m, &s)
	remaining := float64(m) + float64(s)/60.0

	if period <= 4 {
		return float64(period-1)*12 + (12 - remaining)
	}
	return 48 + float64(period-5)*5 + (5 - remaining)
}

// totalGameMinutes returns the full game length based on max period observed.
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

// renderGameFlow builds a plaintextsports-style ASCII game flow chart.
// Uses ':' (2 dots = 3 points) and '.' (1 dot = 1.5 points) stacked vertically.
// Away leads above the axis, home leads below.
func renderGameFlow(plays []playEvent, awayTri, homeTri string, width int) string {
	if len(plays) == 0 || width < 20 {
		return ""
	}

	// Determine number of periods for axis segments.
	numPeriods := 4
	for _, p := range plays {
		if p.period > numPeriods {
			numPeriods = p.period
		}
	}

	// Build axis: +=========+=========+...
	// Each segment gets equal width, bounded by the available space.
	labelW := 5 // "SAS  " or "NYK  " prefix
	axisContentW := width - labelW
	if axisContentW < 20 {
		axisContentW = 20
	}
	// Each segment includes leading '+' except the last which also has trailing '+'
	segW := (axisContentW - 1) / numPeriods // chars of '=' per segment
	if segW < 3 {
		segW = 3
	}
	chartW := segW * numPeriods // total columns of data (excluding '+' markers)

	// Build axis string.
	var axis strings.Builder
	for q := 0; q < numPeriods; q++ {
		axis.WriteString("+")
		axis.WriteString(strings.Repeat("=", segW))
	}
	axis.WriteString("+")
	axisStr := axis.String()

	// Collect margins at each time column. Positive = away leads.
	totalMin := totalGameMinutes(plays)
	margins := make([]int, chartW)
	filled := make([]bool, chartW)

	// Walk oldest-first.
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
		margins[col] = p.scoreAway - p.scoreHome
		filled[col] = true
	}

	// Forward-fill: carry last known margin through empty columns.
	last := 0
	lastFilled := -1
	for i := range margins {
		if filled[i] {
			last = margins[i]
			lastFilled = i
		} else if lastFilled >= 0 {
			margins[i] = last
		}
	}

	// Find last column with real data (don't draw past game progress).
	lastDataCol := 0
	for i := range filled {
		if filled[i] {
			lastDataCol = i
		}
	}

	// Compute max rows needed. Each row = 3 points (2 dots = ':').
	maxAway := 0
	maxHome := 0
	for i := 0; i <= lastDataCol; i++ {
		if margins[i] > maxAway {
			maxAway = margins[i]
		}
		if -margins[i] > maxHome {
			maxHome = -margins[i]
		}
	}
	awayRows := (maxAway + 2) / 3 // ceiling division: points -> rows of ':'
	homeRows := (maxHome + 2) / 3
	if awayRows < 1 && maxAway > 0 {
		awayRows = 1
	}
	if homeRows < 1 && maxHome > 0 {
		homeRows = 1
	}

	// Render the chart lines.
	lines := make([]string, 0, awayRows+homeRows+5)

	// Header.
	header := lipgloss.NewStyle().Bold(true).Render("Game Flow:")
	legend := mutedSty.Render("(2 dots = 3 points)")
	headerW := lipgloss.Width(axisStr) + labelW
	gap := headerW - lipgloss.Width(header) - lipgloss.Width(legend)
	if gap < 2 {
		gap = 2
	}
	lines = append(lines, header+strings.Repeat(" ", gap)+legend)
	lines = append(lines, "")

	// Away rows (top to bottom = largest lead to smallest).
	for row := awayRows; row >= 1; row-- {
		var line strings.Builder
		// Label on the axis row (row 1), but we put it on the axis line itself below.
		line.WriteString(strings.Repeat(" ", labelW))
		for col := 0; col < chartW; col++ {
			// Skip '+' column positions in the axis.
			if col > lastDataCol {
				line.WriteString(" ")
				continue
			}
			m := margins[col]
			if m <= 0 {
				line.WriteString(" ")
				continue
			}
			// How many full rows does this margin fill?
			fullRows := m / 3
			hasPartial := m%3 > 0

			if row <= fullRows {
				line.WriteString(":")
			} else if row == fullRows+1 && hasPartial {
				line.WriteString(".")
			} else {
				line.WriteString(" ")
			}
		}
		lines = append(lines, line.String())
	}

	// Axis line with team labels.
	awayC := teamColorOrDefault(awayTri, awayColor)
	homeC := teamColorOrDefault(homeTri, homeColor)

	awayLabel := lipgloss.NewStyle().Bold(true).Foreground(awayC).
		Render(fmt.Sprintf("%-4s", awayTri))
	homeLabel := lipgloss.NewStyle().Bold(true).Foreground(homeC).
		Render(fmt.Sprintf("%-4s", homeTri))

	axisLine := awayLabel + " " + dimRowSty.Render(axisStr)
	lines = append(lines, axisLine)

	// Home rows (top to bottom = smallest lead to largest).
	for row := 1; row <= homeRows; row++ {
		var line strings.Builder
		if row == 1 {
			// Put home label on first home row.
			line.WriteString(homeLabel)
			line.WriteString(" ")
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
			fullRows := homeLead / 3
			hasPartial := homeLead%3 > 0

			if row <= fullRows {
				line.WriteString(":")
			} else if row == fullRows+1 && hasPartial {
				line.WriteString(".")
			} else {
				line.WriteString(" ")
			}
		}
		lines = append(lines, line.String())
	}

	// If home never led, still show the label below axis.
	if homeRows == 0 {
		lines = append(lines, homeLabel)
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
