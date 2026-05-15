package backend

import (
	"github.com/poteto0/go-nba-sdk/api"
	"github.com/poteto0/go-nba-sdk/api/live"
	"github.com/poteto0/go-nba-sdk/types"
)

type PlayerStats struct {
	name       string
	minutes    string
	points     int
	assists    int
	rebTotal   int
	rebOff     int
	rebDef     int
	steals     int
	blocks     int
	turnovers  int
	plusMinus  int
	fgMade     int
	fgAttempts int
	tpMade     int
	tpAttempts int
	ftMade     int
	ftAttempts int
	fouls      int
}

type GameBoxScore struct {
	gameId        string
	homeTeam      string
	awayTeam      string
	homeScore     int
	awayScore     int
	period        int
	gameClock     string
	arena         string
	attendance    int
	homeTeamStats []PlayerStats
	awayTeamStats []PlayerStats
}

type Play struct {
	period      int
	clock       string
	description string
	actionType  string
	playerName  string
	teamName    string
	scoreHome   string
	scoreAway   string
}

type PlayByPlay struct {
	gameId string
	plays  []Play
}

func derefInt(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}

func derefFloat(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}

func GetGameBoxScore(gameId string) GameBoxScore {
	provider := api.NewProvider(nil)
	result := live.GetBoxScore(provider, &types.BoxScoreParams{
		GameID: gameId,
	})
	game := result.Contents.Game

	homeStats := make([]PlayerStats, 0, 20)
	for _, p := range *game.HomeTeam.Players {
		homeStats = append(homeStats, PlayerStats{
			name:       p.Name,
			minutes:    p.Statistics.MinutesClock(),
			points:     derefInt(p.Statistics.Pts),
			assists:    derefInt(p.Statistics.Ast),
			rebTotal:   derefInt(p.Statistics.Reb),
			rebOff:     derefInt(p.Statistics.OReb),
			rebDef:     derefInt(p.Statistics.DReb),
			steals:     derefInt(p.Statistics.Stl),
			blocks:     derefInt(p.Statistics.Blk),
			turnovers:  derefInt(p.Statistics.Tov),
			plusMinus:  int(derefFloat(p.Statistics.PlusMinus)),
			fgMade:     derefInt(p.Statistics.FgM),
			fgAttempts: derefInt(p.Statistics.FgA),
			tpMade:     derefInt(p.Statistics.Fg3M),
			tpAttempts: derefInt(p.Statistics.Fg3A),
			ftMade:     derefInt(p.Statistics.FtM),
			ftAttempts: derefInt(p.Statistics.FtA),
			fouls:      derefInt(p.Statistics.PF),
		})
	}

	awayStats := make([]PlayerStats, 0, 20)
	for _, p := range *game.AwayTeam.Players {
		awayStats = append(awayStats, PlayerStats{
			name:       p.Name,
			minutes:    p.Statistics.MinutesClock(),
			points:     derefInt(p.Statistics.Pts),
			assists:    derefInt(p.Statistics.Ast),
			rebTotal:   derefInt(p.Statistics.Reb),
			rebOff:     derefInt(p.Statistics.OReb),
			rebDef:     derefInt(p.Statistics.DReb),
			steals:     derefInt(p.Statistics.Stl),
			blocks:     derefInt(p.Statistics.Blk),
			turnovers:  derefInt(p.Statistics.Tov),
			plusMinus:  int(derefFloat(p.Statistics.PlusMinus)),
			fgMade:     derefInt(p.Statistics.FgM),
			fgAttempts: derefInt(p.Statistics.FgA),
			tpMade:     derefInt(p.Statistics.Fg3M),
			tpAttempts: derefInt(p.Statistics.Fg3A),
			ftMade:     derefInt(p.Statistics.FtM),
			ftAttempts: derefInt(p.Statistics.FtA),
			fouls:      derefInt(p.Statistics.PF),
		})
	}

	return GameBoxScore{
		gameId:        game.GameId,
		homeTeam:      game.HomeTeam.TeamName,
		awayTeam:      game.AwayTeam.TeamName,
		homeScore:     game.HomeTeam.Score,
		awayScore:     game.AwayTeam.Score,
		period:        game.Period,
		gameClock:     parseGameClock(game.GameClock),
		arena:         game.Arena.ArenaName,
		homeTeamStats: homeStats,
		awayTeamStats: awayStats,
	}
}

func GetPlayByPlay(gameId string) PlayByPlay {
	provider := api.NewProvider(nil)
	result := live.GetPlayByPlay(provider, &types.PlayByPlayParams{
		GameID: gameId,
	})
	game := result.Contents.Game

	plays := make([]Play, 0, len(game.Actions))
	for _, a := range game.Actions {
		plays = append(plays, Play{
			period:      a.Period,
			clock:       parseGameClock(a.Clock),
			description: a.Description,
			actionType:  a.ActionType,
			playerName:  a.PlayerName,
			teamName:    a.TeamTricode,
			scoreHome:   a.ScoreHome,
			scoreAway:   a.ScoreAway,
		})
	}
	return PlayByPlay{
		gameId: game.GameID,
		plays:  plays,
	}
}
