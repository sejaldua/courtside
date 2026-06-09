package backend

import (
	"context"
	"fmt"
	"strings"
	"time"

	nba "github.com/NolanFogarty/nba-sdk"
	"github.com/NolanFogarty/nba-sdk/stats"
)

// GetGamesForDate returns every game on the given date ("YYYY-MM-DD") — past
// results, in-progress, and future scheduled games — from stats.nba.com's
// scoreboardv3 endpoint, which returns full team names, scores, and status
// directly. Today is routed to the live CDN scoreboard (the same source used at
// startup) so the navigated "today" view matches the initial one.
func GetGamesForDate(date string) ([]Game, error) {
	if isTodayET(date) {
		return GetTodaysGames()
	}

	d, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, err
	}

	client := nba.NewClient()
	resp, err := client.Stats.ScoreboardV3(context.Background(), d)
	if err != nil {
		return nil, err
	}

	games := make([]Game, 0, len(resp.Scoreboard.Games))
	for _, g := range resp.Scoreboard.Games {
		games = append(games, scoreboardGame(g))
	}
	return games, nil
}

// scoreboardGame converts a scoreboardv3 game into our internal Game, deriving
// the display clock from the game's status.
func scoreboardGame(g stats.ScoreboardV3Game) Game {
	var clock string
	switch g.GameStatus {
	case statusLive:
		clock = fmt.Sprintf("Q%d %s", g.Period, parseGameClock(g.GameClock))
	case statusFinal:
		clock = "Final"
	default:
		clock = strings.TrimSpace(g.GameStatusText)
	}

	homePeriods := make([]int, len(g.HomeTeam.Periods))
	for i, p := range g.HomeTeam.Periods {
		homePeriods[i] = p.Score
	}
	awayPeriods := make([]int, len(g.AwayTeam.Periods))
	for i, p := range g.AwayTeam.Periods {
		awayPeriods[i] = p.Score
	}

	return Game{
		GameId:      g.GameID,
		HomeTeam:    g.HomeTeam.TeamName,
		AwayTeam:    g.AwayTeam.TeamName,
		HomeTricode: g.HomeTeam.TeamTricode,
		AwayTricode: g.AwayTeam.TeamTricode,
		HomeScore:   g.HomeTeam.Score,
		AwayScore:   g.AwayTeam.Score,
		HomePeriods: homePeriods,
		AwayPeriods: awayPeriods,
		GameClock:   clock,
		status:      g.GameStatus,
	}
}

// isTodayET reports whether the given "YYYY-MM-DD" date is today (ET).
func isTodayET(date string) bool {
	d, err := time.Parse("2006-01-02", date)
	if err != nil {
		return false
	}
	now := time.Now()
	if loc, err := time.LoadLocation("America/New_York"); err == nil {
		now = now.In(loc)
	}
	y, m, day := now.Date()
	return d.Year() == y && d.Month() == m && d.Day() == day
}
