package backend

var testing bool = true

func TestList() []Game {
	var testlist []Game

	testlist = []Game{
			{
					GameId: "001",
					HomeTeam: "Lakers",
					AwayTeam: "Celtics",
					GameClock: "Q1 08:45",
					HomeScore: 110,
					AwayScore: 100,
			},
			{
					GameId: "002",
					HomeTeam: "Bulls",
					AwayTeam: "Raptors",
					GameClock: "Q2 05:12",
					HomeScore: 95,
					AwayScore: 92,
			},
			{
					GameId: "003",
					HomeTeam: "Warriors",
					AwayTeam: "Pacers",
					GameClock: "Q3 01:20",
					HomeScore: 105,
					AwayScore: 98,
			},
	}
	return testlist
}

