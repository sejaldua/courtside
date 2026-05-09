package main

import (
	"fmt"
	"os"

	"github.com/NolanFogarty/courtside/backend"
	"github.com/NolanFogarty/courtside/ui"
)

func mergeGames(live, scheduled []backend.Game) []backend.Game {
	seen := make(map[string]bool)
	merged := make([]backend.Game, 0, len(live)+len(scheduled))
	for _, g := range live {
		seen[g.GameId] = true
		merged = append(merged, g)
	}
	for _, g := range scheduled {
		if !seen[g.GameId] {
			merged = append(merged, g)
		}
	}
	return merged
}

func main() {
	games := mergeGames(backend.GetLiveGames(), backend.GetScheduledGames())

	if err := ui.Run(games); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
