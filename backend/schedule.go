package backend

import (
	"context"
	"encoding/json"
	"strings"

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
			name := rowString(row, idx, "TEAM_NAME")
			if name == "" {
				name = rowString(row, idx, "TEAM_NICKNAME")
			}
			if lines[gid] == nil {
				lines[gid] = make(map[int]teamLine)
			}
			lines[gid][tid] = teamLine{name: name, pts: rowInt(row, idx, "PTS")}
		}
	}

	games := make([]Game, 0, len(resp.Games))
	for _, g := range resp.Games {
		home := lines[g.GameID][g.HomeTeamID]
		away := lines[g.GameID][g.VisitorTeamID]
		games = append(games, Game{
			GameId:    g.GameID,
			HomeTeam:  home.name,
			AwayTeam:  away.name,
			HomeScore: home.pts,
			AwayScore: away.pts,
			GameClock: strings.TrimSpace(g.GameStatusText),
		})
	}
	return games, nil
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
