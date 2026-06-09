package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	nba "github.com/NolanFogarty/nba-sdk"
	"github.com/NolanFogarty/nba-sdk/live"
	"github.com/NolanFogarty/nba-sdk/stats"
)

type League int

const (
	NBA League = iota
	WNBA
)

var CurrentLeague = NBA

func (l League) String() string {
	if l == WNBA {
		return "WNBA"
	}
	return "NBA"
}

func newClient() *nba.Client {
	if CurrentLeague == WNBA {
		return nba.NewClient(
			nba.WithStatsBaseURL("https://stats.wnba.com"),
			nba.WithLiveBaseURL("https://cdn.wnba.com"),
		)
	}
	return nba.NewClient()
}

const (
	wnbaCDN   = "https://cdn.wnba.com"
	wnbaStats = "https://stats.wnba.com"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

func wnbaGet(url string, out interface{}) error {
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Referer", "https://www.wnba.com")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("wnba: %s returned %d", url, resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func wnbaLiveScoreboard() (*live.ScoreboardResponse, error) {
	var out live.ScoreboardResponse
	err := wnbaGet(wnbaCDN+"/static/json/liveData/scoreboard/todaysScoreboard_10.json", &out)
	return &out, err
}

func wnbaScoreboardV3(date time.Time) (*stats.ScoreboardV3Response, error) {
	url := fmt.Sprintf("%s/stats/scoreboardv3?GameDate=%s&LeagueID=10", wnbaStats, date.Format("2006-01-02"))
	var out stats.ScoreboardV3Response
	err := wnbaGet(url, &out)
	return &out, err
}

func wnbaBoxScore(gameID string) (*live.BoxScoreResponse, error) {
	var out live.BoxScoreResponse
	err := wnbaGet(fmt.Sprintf("%s/static/json/liveData/boxscore/boxscore_%s.json", wnbaCDN, gameID), &out)
	return &out, err
}

func wnbaPlayByPlay(gameID string) (*live.PlayByPlayResponse, error) {
	var out live.PlayByPlayResponse
	err := wnbaGet(fmt.Sprintf("%s/static/json/liveData/playbyplay/playbyplay_%s.json", wnbaCDN, gameID), &out)
	return &out, err
}

func wnbaStandings(season string) (*stats.LeagueStandingsV3Response, error) {
	url := fmt.Sprintf("%s/stats/leaguestandingsv3?Season=%s&LeagueID=10&SeasonType=Regular+Season", wnbaStats, season)
	var out stats.LeagueStandingsV3Response
	err := wnbaGet(url, &out)
	return &out, err
}
