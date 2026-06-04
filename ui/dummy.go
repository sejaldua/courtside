package ui

// teamStatRows returns the away-vs-home team comparison rows. For TO and PF the
// lower value is the "leader" (lowerBetter), otherwise the higher value leads.
func teamStatRows() []statRow {
	return []statRow{
		{label: "FG%", away: "40.0", home: "35.7", awayKey: 40.0, homeKey: 35.7},
		{label: "3P%", away: "33.3", home: "26.7", awayKey: 33.3, homeKey: 26.7},
		{label: "FT%", away: "0.0", home: "87.5", awayKey: 0.0, homeKey: 87.5},
		{label: "FTM", away: "0/0", home: "7/8", awayKey: 0, homeKey: 7},
		{label: "REB", away: "20", home: "17", awayKey: 20, homeKey: 17},
		{label: "AST", away: "7", home: "5", awayKey: 7, homeKey: 5},
		{label: "TO", away: "5", home: "3", awayKey: 5, homeKey: 3, lowerBetter: true},
		{label: "BLK", away: "2", home: "3", awayKey: 2, homeKey: 3},
		{label: "PF", away: "8", home: "5", awayKey: 8, homeKey: 5, lowerBetter: true},
		{label: "PAINT", away: "16", home: "12", awayKey: 16, homeKey: 12},
		{label: "BENCH", away: "10", home: "10", awayKey: 10, homeKey: 10},
	}
}

// dummyGame returns a fully populated game for the detail view until the
// BoxScoreTraditionalV3 / PlayByPlayV3 endpoints are wired in.
func dummyGame() gameDetail {
	return gameDetail{
		away: teamBox{
			name:    "Knicks",
			tricode: "NYK",
			score:   108,
			players: []playerStat{
				{name: "J. Brunson", pts: 31, ast: 7, reb: 4, blk: 0, to: 3, plusMinus: 12},
				{name: "K. Towns", pts: 24, ast: 3, reb: 13, blk: 1, to: 2, plusMinus: 9},
				{name: "M. Bridges", pts: 18, ast: 4, reb: 5, blk: 0, to: 1, plusMinus: 7},
				{name: "OG Anunoby", pts: 14, ast: 2, reb: 6, blk: 2, to: 0, plusMinus: 5},
				{name: "J. Hart", pts: 9, ast: 6, reb: 8, blk: 0, to: 1, plusMinus: 3},
				{name: "M. Robinson", pts: 6, ast: 0, reb: 7, blk: 2, to: 1, plusMinus: -2},
				{name: "D. DiVincenzo", pts: 0, ast: 0, reb: 0, blk: 0, to: 0, plusMinus: 0},
				{name: "A. Portis", pts: 4, ast: 1, reb: 3, blk: 0, to: 0, plusMinus: -1},
				{name: "M. Sensabaugh", pts: 2, ast: 0, reb: 1, blk: 0, to: 1, plusMinus: -3},
				{name: "T. Booker", pts: 0, ast: 1, reb: 0, blk: 0, to: 0, plusMinus: -1},
				{name: "J. McConnell", pts: 0, ast: 0, reb: 0, blk: 0, to: 0, plusMinus: 0},
				{name: "N. Powell", pts: 0, ast: 0, reb: 0, blk: 0, to: 0, plusMinus: 0},
			},
		},
		home: teamBox{
			name:    "Spurs",
			tricode: "SAS",
			score:   114,
			players: []playerStat{
				{name: "V. Wembanyama", pts: 29, ast: 5, reb: 14, blk: 6, to: 4, plusMinus: 15},
				{name: "D. Fox", pts: 26, ast: 9, reb: 3, blk: 0, to: 3, plusMinus: 10},
				{name: "D. Harper", pts: 17, ast: 4, reb: 4, blk: 1, to: 2, plusMinus: 8},
				{name: "H. Barnes", pts: 12, ast: 1, reb: 5, blk: 0, to: 0, plusMinus: 4},
				{name: "J. Sochan", pts: 8, ast: 3, reb: 6, blk: 1, to: 2, plusMinus: 2},
				{name: "C. Castle", pts: 7, ast: 5, reb: 2, blk: 0, to: 1, plusMinus: -1},
				{name: "L. Kornet", pts: 0, ast: 0, reb: 0, blk: 0, to: 0, plusMinus: 0},
				{name: "K. Johnson", pts: 5, ast: 1, reb: 2, blk: 0, to: 1, plusMinus: 1},
				{name: "P. Collins", pts: 3, ast: 0, reb: 2, blk: 1, to: 0, plusMinus: -2},
				{name: "S. Gay", pts: 2, ast: 1, reb: 1, blk: 0, to: 0, plusMinus: -1},
				{name: "Z. Duran", pts: 0, ast: 0, reb: 0, blk: 0, to: 0, plusMinus: 0},
				{name: "J. Champagnie", pts: 0, ast: 0, reb: 0, blk: 0, to: 0, plusMinus: 0},
			},
		},
		// newest first
		plays: []playEvent{
			{period: 4, clock: "0:32", team: "SAS", desc: "Wembanyama made dunk (29 PTS)"},
			{period: 4, clock: "0:48", team: "NYK", desc: "Brunson missed 3pt jump shot"},
			{period: 4, clock: "1:05", team: "", desc: "Knicks full timeout"},
			{period: 4, clock: "1:05", team: "SAS", desc: "Fox made free throw 2 of 2 (26 PTS)"},
			{period: 4, clock: "1:05", team: "SAS", desc: "Fox made free throw 1 of 2"},
			{period: 4, clock: "1:05", team: "", desc: "Shooting foul on Anunoby"},
			{period: 4, clock: "1:22", team: "NYK", desc: "Towns made layup (24 PTS)"},
			{period: 4, clock: "1:41", team: "SAS", desc: "Harper made layup (17 PTS)"},
			{period: 4, clock: "2:03", team: "NYK", desc: "Bridges made 3pt jump shot (18 PTS)"},
			{period: 4, clock: "2:30", team: "SAS", desc: "Wembanyama blocked Brunson's layup"},
			{period: 4, clock: "2:44", team: "", desc: "Spurs 20-second timeout"},
			{period: 4, clock: "3:01", team: "SAS", desc: "Sochan made jump shot (8 PTS)"},
			{period: 4, clock: "3:19", team: "NYK", desc: "Hart made layup (9 PTS)"},
			{period: 4, clock: "3:38", team: "NYK", desc: "Robinson defensive rebound"},
			{period: 4, clock: "3:38", team: "SAS", desc: "Fox missed 3pt jump shot"},
			{period: 4, clock: "4:00", team: "SAS", desc: "Castle made jump shot (7 PTS)"},
			{period: 4, clock: "4:21", team: "NYK", desc: "Brunson made step back 3pt (31 PTS)"},
			{period: 3, clock: "0:12", team: "SAS", desc: "Barnes made jump shot (12 PTS)"},
			{period: 3, clock: "0:40", team: "NYK", desc: "Anunoby made dunk (14 PTS)"},
			{period: 3, clock: "1:08", team: "", desc: "Personal foul on Sochan"},
			{period: 3, clock: "1:30", team: "SAS", desc: "Wembanyama made hook shot (23 PTS)"},
			{period: 3, clock: "1:55", team: "NYK", desc: "Towns made 3pt jump shot (22 PTS)"},
		},
	}
}
