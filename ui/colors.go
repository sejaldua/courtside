package ui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// teamColorHex maps NBA team tricodes to their primary brand color.
var teamColorHex = map[string]string{
	"ATL": "#E03A3E",
	"BOS": "#007A33",
	"BKN": "#000000",
	"CHA": "#1D1160",
	"CHI": "#CE1141",
	"CLE": "#6F263D",
	"DAL": "#00538C",
	"DEN": "#0E2240",
	"DET": "#C8102E",
	"GSW": "#1D428A",
	"HOU": "#CE1141",
	"IND": "#002D62",
	"LAC": "#C8102E",
	"LAL": "#552583",
	"MEM": "#5D76A9",
	"MIA": "#98002E",
	"MIL": "#00471B",
	"MIN": "#0C2340",
	"NOP": "#0C2340",
	"NYK": "#F58426",
	"OKC": "#007AC1",
	"ORL": "#0077C0",
	"PHI": "#006BB6",
	"PHX": "#1D1160",
	"POR": "#E03A3E",
	"SAC": "#5A2D81",
	"SAS": "#C4CED4",
	"TOR": "#CE1141",
	"UTA": "#002B5C",
	"WAS": "#002B5C",
}

// WNBA team tricodes to their primary brand color.
var wnbaColorHex = map[string]string{
	"ATL":  "#CC092F",
	"CHI":  "#418FDE",
	"CON":  "#F05023",
	"DAL":  "#C4D600",
	"GS":   "#552583",
	"IND":  "#002D62",
	"LVA":  "#000000",
	"LA":   "#552583",
	"MIN":  "#236192",
	"NYL":  "#6ECEB2",
	"PHX":  "#CB6015",
	"SEA":  "#2C5234",
	"WAS":  "#E31837",
}

func teamColor(tricode string) color.Color {
	if hex, ok := teamColorHex[tricode]; ok {
		return lipgloss.Color(hex)
	}
	if hex, ok := wnbaColorHex[tricode]; ok {
		return lipgloss.Color(hex)
	}
	return lipgloss.Color("247")
}
