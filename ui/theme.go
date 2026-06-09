package ui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

type theme struct {
	name     string
	muted    color.Color
	dim      color.Color
	accent   color.Color
	border   color.Color
	helpKey  color.Color
	helpDesc color.Color
	helpSep  color.Color
}

var darkTheme = theme{
	name:     "dark",
	muted:    lipgloss.Color("245"),
	dim:      lipgloss.Color("240"),
	accent:   lipgloss.Color("247"),
	border:   lipgloss.Color("240"),
	helpKey:  lipgloss.Color("#626262"),
	helpDesc: lipgloss.Color("#4A4A4A"),
	helpSep:  lipgloss.Color("#3C3C3C"),
}

var lightTheme = theme{
	name:     "light",
	muted:    lipgloss.Color("#555555"),
	dim:      lipgloss.Color("#999999"),
	accent:   lipgloss.Color("#B8860B"),
	border:   lipgloss.Color("#CCCCCC"),
	helpKey:  lipgloss.Color("#555555"),
	helpDesc: lipgloss.Color("#777777"),
	helpSep:  lipgloss.Color("#AAAAAA"),
}

var currentTheme = darkTheme

func toggleTheme() {
	if currentTheme.name == "dark" {
		currentTheme = lightTheme
	} else {
		currentTheme = darkTheme
	}
	applyTheme()
}

func applyTheme() {
	t := currentTheme
	mutedColor = t.muted
	dimColor = t.dim
	accentColor = t.accent

	headerBarSty = lipgloss.NewStyle().Bold(true).Padding(0, 1).
		Border(lipgloss.RoundedBorder()).BorderForeground(t.border)
	panelSty = lipgloss.NewStyle().Padding(0, 1).
		Border(lipgloss.RoundedBorder()).BorderForeground(t.border)
	colHeaderSty = lipgloss.NewStyle().Bold(true).Foreground(t.muted)
	dimRowSty = lipgloss.NewStyle().Foreground(t.dim)
	mutedSty = lipgloss.NewStyle().Foreground(t.muted)
	accentSty = lipgloss.NewStyle().Bold(true).Foreground(t.accent)
	hintSty = lipgloss.NewStyle().Foreground(t.muted)
	errSty = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#C8102E"))
	helpKeySty = lipgloss.NewStyle().Foreground(t.helpKey)
	helpDescSty = lipgloss.NewStyle().Foreground(t.helpDesc)
	helpSepSty = lipgloss.NewStyle().Foreground(t.helpSep)
	helpSty = lipgloss.NewStyle().PaddingLeft(2)
}
