package backend

import (
	"context"
	"fmt"

	nba "github.com/NolanFogarty/nba-sdk"
	"github.com/NolanFogarty/nba-sdk/live"
)

// gameStatus values returned by the live scoreboard.
const (
	statusUpcoming = 1
	statusLive     = 2
	statusFinal    = 3
)

func parseGameClock(raw string) string {
	var mins int
	var secs float64
	if _, err := fmt.Sscanf(raw, "PT%dM%fS", &mins, &secs); err != nil {
		return raw
	}
	return fmt.Sprintf("%02d:%02d", mins, int(secs))
}

type Game struct {
	GameId    string
	HomeTeam  string
	AwayTeam  string
	GameClock string
	HomeScore int
	AwayScore int
	status    int // 1 = scheduled, 2 = live, 3 = final
}

// NotStarted reports whether the game is scheduled but hasn't tipped off yet
// (so there's no box score or play-by-play to fetch). GameClock holds the
// tip-off time in this state.
func (g Game) NotStarted() bool {
	return g.status == statusUpcoming
}

// toGame converts an SDK scoreboard game into our internal Game, deriving the
// display clock from the game's status.
func toGame(g live.Game) Game {
	var clock string
	switch g.GameStatus {
	case statusLive:
		clock = fmt.Sprintf("Q%d %s", g.Period, parseGameClock(g.GameClock))
	case statusFinal:
		clock = "Final"
	default:
		// Upcoming: GameStatusText already carries the tip-off time (e.g. "7:30 pm ET").
		clock = g.GameStatusText
	}

	return Game{
		GameId:    g.GameID,
		HomeTeam:  g.HomeTeam.TeamName,
		AwayTeam:  g.AwayTeam.TeamName,
		GameClock: clock,
		HomeScore: g.HomeTeam.Score,
		AwayScore: g.AwayTeam.Score,
		status:    g.GameStatus,
	}
}

// todaysGames fetches today's scoreboard and returns only the games whose
// status passes keep.
func todaysGames(keep func(status int) bool) []Game {
	client := nba.NewClient()
	resp, err := client.Live.Scoreboard(context.Background())
	if err != nil {
		panic(err)
	}

	games := make([]Game, 0, len(resp.Scoreboard.Games))
	for _, g := range resp.Scoreboard.Games {
		if keep(g.GameStatus) {
			games = append(games, toGame(g))
		}
	}
	return games
}

// GetScheduledGames returns today's games that are not currently in progress
// (upcoming and finished).
func GetScheduledGames() []Game {
	return todaysGames(func(status int) bool {
		return status != statusLive
	})
}

// GetLiveGames returns today's games that are currently in progress.
func GetLiveGames() []Game {
	return todaysGames(func(status int) bool {
		return status == statusLive
	})
}

// GetTodaysGames returns every game on today's live CDN scoreboard (upcoming,
// in-progress, and final) — the same source used at startup. GetGamesForDate
// routes today's date here so the navigated "today" view matches the initial
// one instead of using the staler scoreboardv2 schedule.
func GetTodaysGames() ([]Game, error) {
	client := nba.NewClient()
	resp, err := client.Live.Scoreboard(context.Background())
	if err != nil {
		return nil, err
	}
	games := make([]Game, 0, len(resp.Scoreboard.Games))
	for _, g := range resp.Scoreboard.Games {
		games = append(games, toGame(g))
	}
	return games, nil
}
