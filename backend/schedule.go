package backend

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	nba "github.com/NolanFogarty/nba-sdk"
)

// GetGamesForDate returns every game on the given date ("YYYY-MM-DD"), including
// past results and future scheduled games. It uses stats.nba.com's scoreboardv2
// endpoint, joining the GameHeader table (which game is home/away) with the
// LineScore table (team names and points).
func GetGamesForDate(date string) ([]Game, error) {
	client := nba.NewClient()
	resp, err := client.Stats.ScoreboardV2(context.Background(), date)
	if err != nil {
		return nil, err
	}

	// (gameID -> teamID -> name/points) from the LineScore result set.
	type teamLine struct {
		name string
		pts  int
	}
	lines := make(map[string]map[int]teamLine)
	for _, rs := range resp.ResultSets {
		if rs.Name != "LineScore" {
			continue
		}
		idx := headerIndex(rs.Headers)
		for _, row := range rs.RowSet {
			gid := rowString(row, idx, "GAME_ID")
			tid := rowInt(row, idx, "TEAM_ID")
			name := coalesce(rowString(row, idx, "TEAM_NAME"), rowString(row, idx, "TEAM_NICKNAME"))
			if lines[gid] == nil {
				lines[gid] = make(map[int]teamLine)
			}
			lines[gid][tid] = teamLine{name: name, pts: rowInt(row, idx, "PTS")}
		}
	}

	past := isPastDate(date)
	games := make([]Game, 0, len(resp.Games))
	for _, g := range resp.Games {
		home := lines[g.GameID][g.HomeTeamID]
		away := lines[g.GameID][g.VisitorTeamID]

		// LineScore can be sparse/missing for games scoreboardv2 reports as
		// scheduled; GAMECODE ("YYYYMMDD/VISHOME") is the fallback for identity.
		// Stale finished games get their real names from the box score below.
		awayCode, homeCode := parseMatchup(g.GameCode)
		game := Game{
			GameId:    g.GameID,
			HomeTeam:  coalesce(home.name, homeCode),
			AwayTeam:  coalesce(away.name, awayCode),
			HomeScore: home.pts,
			AwayScore: away.pts,
			GameClock: strings.TrimSpace(g.GameStatusText),
		}

		// scoreboardv2 sometimes still lists a finished game on a past date as
		// scheduled (0-0, tip-off time). Patch those from the authoritative box
		// score.
		if past && g.GameStatusID != statusFinal {
			reconcileFromBoxScore(client, &game)
		}

		games = append(games, game)
	}
	return games, nil
}

// reconcileFromBoxScore patches a game's score and status from the traditional
// box score. It leaves the game untouched if the box score can't be fetched or
// has no players (a game that wasn't actually played — e.g. postponed).
func reconcileFromBoxScore(client *nba.Client, game *Game) {
	box, err := client.Stats.BoxScoreTraditionalV3(context.Background(), game.GameId)
	if err != nil {
		return
	}
	bt := box.BoxScoreTraditional
	if len(bt.HomeTeam.Players) == 0 && len(bt.AwayTeam.Players) == 0 {
		return
	}
	game.HomeScore = bt.HomeTeam.Statistics.Points
	game.AwayScore = bt.AwayTeam.Statistics.Points
	game.GameClock = "Final"
	game.HomeTeam = coalesce(bt.HomeTeam.TeamName, game.HomeTeam)
	game.AwayTeam = coalesce(bt.AwayTeam.TeamName, game.AwayTeam)
}

// isPastDate reports whether the given "YYYY-MM-DD" date is before today (ET).
func isPastDate(date string) bool {
	d, err := time.Parse("2006-01-02", date)
	if err != nil {
		return false
	}
	now := time.Now()
	if loc, err := time.LoadLocation("America/New_York"); err == nil {
		now = now.In(loc)
	}
	y, m, day := now.Date()
	today := time.Date(y, m, day, 0, 0, 0, 0, time.UTC)
	return d.Before(today)
}

// parseMatchup splits a GAMECODE like "20241225/NYKSAS" into the visitor and
// home tricodes ("NYK", "SAS").
func parseMatchup(gameCode string) (away, home string) {
	_, codes, ok := strings.Cut(gameCode, "/")
	if !ok || len(codes) != 6 {
		return "", ""
	}
	return codes[:3], codes[3:]
}

// coalesce returns the first non-empty string.
func coalesce(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// headerIndex maps a result set's column names to their positions.
func headerIndex(headers []string) map[string]int {
	idx := make(map[string]int, len(headers))
	for i, h := range headers {
		idx[h] = i
	}
	return idx
}

func rowString(row []json.RawMessage, idx map[string]int, col string) string {
	i, ok := idx[col]
	if !ok || i >= len(row) {
		return ""
	}
	var s string
	_ = json.Unmarshal(row[i], &s)
	return s
}

func rowInt(row []json.RawMessage, idx map[string]int, col string) int {
	i, ok := idx[col]
	if !ok || i >= len(row) {
		return 0
	}
	var n int
	_ = json.Unmarshal(row[i], &n)
	return n
}
