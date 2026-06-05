package backend

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	nba "github.com/NolanFogarty/nba-sdk"
	"github.com/NolanFogarty/nba-sdk/stats"
)

// PlayerLine is one player's box-score line.
type PlayerLine struct {
	Name string
	Min  string // minutes played

	Pts, Ast, Reb, Blk, To int
	PlusMinus              int

	// expanded detail
	FGM, FGA, TPM, TPA, FTM, FTA int
	OReb, DReb, Stl, PF          int
}

// TeamDetail is one team's full box score plus the aggregate stats shown in the
// comparison bar.
type TeamDetail struct {
	Name, Tricode string
	Score         int
	Players       []PlayerLine

	FGPct, FG3Pct float64 // shooting percentages, 0-100
	FTM, FTA      int     // free throws made / attempted
	Reb, Ast, To  int

	// expanded detail
	FGM, FGA, TPM, TPA       int
	OReb, DReb, Stl, Blk, PF int
}

// PlayLine is a single play-by-play event. Team is the tricode, empty for
// neutral events (timeouts, period boundaries, etc.).
type PlayLine struct {
	Period int
	Clock  string // "MM:SS"
	Team   string
	Desc   string
}

// GameDetail is everything the detail screen needs for one game.
type GameDetail struct {
	Away, Home TeamDetail
	Plays      []PlayLine // newest first
}

// GetGameDetail fetches the traditional box score and play-by-play for a game
// and projects them into a GameDetail.
func GetGameDetail(gameID string) (GameDetail, error) {
	client := nba.NewClient()
	ctx := context.Background()

	box, err := client.Stats.BoxScoreTraditionalV3(ctx, gameID)
	if err != nil {
		return GameDetail{}, err
	}
	pbp, err := client.Stats.PlayByPlayV3(ctx, gameID)
	if err != nil {
		return GameDetail{}, err
	}

	d := GameDetail{
		Away: toTeamDetail(box.BoxScoreTraditional.AwayTeam),
		Home: toTeamDetail(box.BoxScoreTraditional.HomeTeam),
	}

	// Actions come oldest-first; the feed wants newest-first.
	actions := pbp.Game.Actions
	d.Plays = make([]PlayLine, 0, len(actions))
	for i := len(actions) - 1; i >= 0; i-- {
		a := actions[i]
		d.Plays = append(d.Plays, PlayLine{
			Period: a.Period,
			Clock:  parseGameClock(a.Clock),
			Team:   a.TeamTricode,
			Desc:   a.Description,
		})
	}
	return d, nil
}

func toTeamDetail(t stats.BoxTeam) TeamDetail {
	s := t.Statistics
	td := TeamDetail{
		Name:    t.TeamName,
		Tricode: t.TeamTricode,
		Score:   s.Points,
		FGPct:   pct(s.FieldGoalsMade, s.FieldGoalsAttempted),
		FG3Pct:  pct(s.ThreePointersMade, s.ThreePointersAttempted),
		FTM:     s.FreeThrowsMade,
		FTA:     s.FreeThrowsAttempted,
		Reb:     s.ReboundsTotal,
		Ast:     s.Assists,
		To:      s.Turnovers,
		FGM:     s.FieldGoalsMade,
		FGA:     s.FieldGoalsAttempted,
		TPM:     s.ThreePointersMade,
		TPA:     s.ThreePointersAttempted,
		OReb:    s.ReboundsOffensive,
		DReb:    s.ReboundsDefensive,
		Stl:     s.Steals,
		Blk:     s.Blocks,
		PF:      s.FoulsPersonal,
	}

	for _, p := range t.Players {
		name := p.NameI
		if name == "" {
			name = p.FamilyName
		}
		ps := p.Statistics
		td.Players = append(td.Players, PlayerLine{
			Name:      name,
			Min:       fmtMinutes(ps.Minutes),
			Pts:       ps.Points,
			Ast:       ps.Assists,
			Reb:       ps.ReboundsTotal,
			Blk:       ps.Blocks,
			To:        ps.Turnovers,
			PlusMinus: int(ps.PlusMinusPoints),
			FGM:       ps.FieldGoalsMade,
			FGA:       ps.FieldGoalsAttempted,
			TPM:       ps.ThreePointersMade,
			TPA:       ps.ThreePointersAttempted,
			FTM:       ps.FreeThrowsMade,
			FTA:       ps.FreeThrowsAttempted,
			OReb:      ps.ReboundsOffensive,
			DReb:      ps.ReboundsDefensive,
			Stl:       ps.Steals,
			PF:        ps.FoulsPersonal,
		})
	}
	return td
}

// fmtMinutes extracts whole minutes from the box score's minutes field, which
// may be ISO ("PT34M12.00S") or "MM:SS".
func fmtMinutes(raw string) string {
	if raw == "" {
		return "0"
	}
	var m int
	if _, err := fmt.Sscanf(raw, "PT%dM", &m); err == nil {
		return strconv.Itoa(m)
	}
	if i := strings.IndexByte(raw, ':'); i >= 0 {
		return raw[:i]
	}
	return raw
}

// pct returns made/attempted as a 0-100 percentage, or 0 when none attempted.
func pct(made, attempted int) float64 {
	if attempted == 0 {
		return 0
	}
	return float64(made) / float64(attempted) * 100
}
