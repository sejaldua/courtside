package backend

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/NolanFogarty/nba-sdk/stats"
)

// Standing is one team's row in the conference standings.
type Standing struct {
	Rank      int
	Team      string // nickname
	Wins      int
	Losses    int
	PCT       float64
	GamesBack float64
	L10       string
	Streak    string
	Clinch    string // clinch/elimination marker (x, y, e, …), "" if none
}

// Standings holds the league standings split by conference, each sorted by
// playoff rank.
type Standings struct {
	Season string
	East   []Standing
	West   []Standing
}

// GetStandings fetches the current regular-season standings.
func GetStandings() (Standings, error) {
	season := currentSeason()

	var teams []stats.TeamStanding
	if CurrentLeague == WNBA {
		r, err := wnbaStandings(season)
		if err != nil {
			return Standings{}, err
		}
		teams = r.Standings
	} else {
		client := newClient()
		r, err := client.Stats.LeagueStandingsV3(context.Background(), season)
		if err != nil {
			return Standings{}, err
		}
		teams = r.Standings
	}

	out := Standings{Season: season}
	for _, t := range teams {
		s := Standing{
			Rank:      t.PlayoffRank,
			Team:      t.TeamName,
			Wins:      t.Wins,
			Losses:    t.Losses,
			PCT:       t.WinPCT,
			GamesBack: t.ConferenceGamesBack,
			L10:       t.L10,
			Streak:    t.StrCurrentStreak,
			Clinch:    clinchMark(t.ClinchIndicator),
		}
		switch t.Conference {
		case "East":
			out.East = append(out.East, s)
		case "West":
			out.West = append(out.West, s)
		}
	}

	sort.SliceStable(out.East, func(i, j int) bool { return out.East[i].Rank < out.East[j].Rank })
	sort.SliceStable(out.West, func(i, j int) bool { return out.West[i].Rank < out.West[j].Rank })
	return out, nil
}

// clinchMark reduces an NBA clinch indicator (e.g. " - x", " - e") to its
// short marker ("x", "e"), or "" if there is none.
func clinchMark(indicator string) string {
	f := strings.Fields(indicator)
	if len(f) == 0 {
		return ""
	}
	return f[len(f)-1]
}

// currentSeason returns the season string for the active league.
// NBA: "YYYY-YY" (season starts in October, e.g. "2025-26").
// WNBA: "YYYY" (season runs May-September within one calendar year, e.g. "2025").
func currentSeason() string {
	now := time.Now()
	if loc, err := time.LoadLocation("America/New_York"); err == nil {
		now = now.In(loc)
	}
	year := now.Year()

	if CurrentLeague == WNBA {
		return fmt.Sprintf("%d", year)
	}

	if now.Month() < time.October {
		return fmt.Sprintf("%d-%02d", year-1, year%100)
	}
	return fmt.Sprintf("%d-%02d", year, (year+1)%100)
}
