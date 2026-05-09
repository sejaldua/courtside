package backend

import (
	"fmt"
	"testing"
)

func TestGetGames(t *testing.T) {
	games := GetGames()

	if games == nil {
		t.Fatal("GetGames() returned nil")
	}

	if len(games) == 0 {
		t.Fatal("GetGames() returned empty slice")
	}

	fmt.Printf("\n=== Games Retrieved (%d total) ===\n", len(games))
	for i, game := range games {
		fmt.Printf("%d. %s vs %s | %d-%d | %s\n", i+1, game.HomeTeam, game.AwayTeam, game.HomeScore, game.AwayScore, game.GameClock)

		if game.HomeTeam == "" {
			t.Errorf("Game %d: HomeTeam is empty", i)
		}
		if game.AwayTeam == "" {
			t.Errorf("Game %d: AwayTeam is empty", i)
		}
		if game.GameClock == "" {
			t.Errorf("Game %d: GameClock is empty", i)
		}
	}
}
