package backend

import (
	"fmt"
	"github.com/poteto0/go-nba-sdk/gns"
)

type Game struct {
	HomeTeam  string
	AwayTeam  string
	GameClock string
	HomeScore int
	AwayScore int
}

func GetGames() []Game {
	client := gns.NewClient(nil)
	result := client.Live.GetScoreBoard(nil)
	games := result.Contents.Scoreboard.Games

	gamelist := make([]Game, 0, 20)

	for _, game := range games {
		var gameclock string
		if game.IsFinished() {
			gameclock = fmt.Sprintf("  Final  ")
		} else {
			gameclock = fmt.Sprintf("%dQ (%s)", game.Period, game.GameClock)
		}

		gamelist = append(gamelist, Game {
			HomeTeam: game.HomeTeam.TeamName,
			AwayTeam: game.AwayTeam.TeamName,
			GameClock: gameclock,
			HomeScore: game.HomeTeam.Score,
			AwayScore: game.AwayTeam.Score,
		})
	}
	return gamelist
}
