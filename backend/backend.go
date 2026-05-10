package backend

import (
	"fmt"
	"time"
	"github.com/poteto0/go-nba-sdk/gns"
	"github.com/poteto0/go-nba-sdk/types"
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
	GameId		string
	HomeTeam  string
	AwayTeam  string
	GameClock string
	HomeScore int
	AwayScore int
}

func currentSeason() string {
	now := time.Now()
	year := now.Year()
	// nba season starts in october
	if now.Month() < time.October {
		// e.g. May 2026 → "2025-26"
		return fmt.Sprintf("%d-%02d", year-1, year%100)
	}
	// e.g. October 2026 → "2026-27"
	return fmt.Sprintf("%d-%02d", year, (year+1)%100)
}

func GetScheduledGames() []Game {
	client := gns.NewClient(nil)
	result := client.Stats.GetScheduleLeagueV2(
		&types.ScheduleLeagueV2Params{
			LeagueID: "00",
			Season: currentSeason(),
		},
	)

	if result.Error != nil {
		panic(result.Error)
	}

	//today := time.Now().Format("01/02/2006 12:00:00 AM")
	gamelist := make([]Game, 0, 20)

	for _, gameDate := range result.Contents.LeagueSchedule.GameDates {
		gamedate, _ := time.Parse("01/02/2006 15:04:05", gameDate.GameDate)
		now := time.Now()
		isToday := gamedate.Year() == now.Year() &&
				gamedate.Month() == now.Month() &&
				gamedate.Day() == now.Day()
		if isToday {
			for _, game := range gameDate.Games {
				// Skip live games — this endpoint doesn't provide quarter/time left;
				// use GetLiveGames() instead
				if game.GameStatus == 2 { // TODO: confirm '2' actually means game is live
					// 1 = not started, 2 = live, 3 = final?
					continue
				}
				gameclock := game.GameStatusText

				gamelist = append(gamelist, Game{
					GameId:    game.GameID,
					HomeTeam:  game.HomeTeam.TeamName,
					AwayTeam:  game.AwayTeam.TeamName,
					GameClock: gameclock,
					HomeScore: game.HomeTeam.Score,
					AwayScore: game.AwayTeam.Score,
				})
			}
		}
	}
	return gamelist
}

func GetLiveGames() []Game {
	client := gns.NewClient(nil)
	result := client.Live.GetScoreBoard(nil)

	games := result.Contents.Scoreboard.Games
	gamelist := make([]Game, 0, 20)

	for _, game := range games {
		var gameclock string
		if !game.IsGameStart() {
			gameclock = game.GameEt.Format("3:04 pm ET")
		} else if game.IsFinished() {
			gameclock = fmt.Sprintf("Final")
		} else {
			gameclock = fmt.Sprintf("Q%d %s", game.Period, parseGameClock(game.GameClock))
		}

		gamelist = append(gamelist, Game {
			GameId: 	 game.GameId,
			HomeTeam:  game.HomeTeam.TeamName,
			AwayTeam:  game.AwayTeam.TeamName,
			GameClock: gameclock,
			HomeScore: game.HomeTeam.Score,
			AwayScore: game.AwayTeam.Score,
		})
	}
	return gamelist
}
