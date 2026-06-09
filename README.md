# Courtside

Terminal UI for following NBA games. Live scores, box scores, play-by-play, standings — all from the command line.

## What it does

Open your terminal, run `courtside`, and you get today's scoreboard. Arrow into any game for the full box score and a scrollable play-by-play feed. Navigate between days to check past results. Live games auto-refresh every 15 seconds.

The detail view shows Dean Oliver's Four Factors (eFG%, TOV%, ORB%, FT Rate) in the team comparison bar, a plaintextsports-inspired ASCII game flow chart showing scoring runs, and a lead changes/ties summary.

## Data

Pulls from the NBA's public JSON endpoints (cdn.nba.com for live data, stats.nba.com for historical) via [`nba-sdk`](https://github.com/NolanFogarty/nba-sdk). No API key needed.

## Install

```bash
go install github.com/NolanFogarty/courtside@latest
```

Or build from source:

```bash
git clone https://github.com/NolanFogarty/courtside.git
cd courtside
go build -o courtside
sudo mv courtside /usr/local/bin/
```

## Keys

| Key | Action |
| --- | --- |
| `↑/k`, `↓/j` | Navigate / scroll |
| `enter` | Open game detail |
| `←/h`, `→/l` | Previous / next day |
| `d` | Jump to date |
| `s` | Standings |
| `o` | Toggle expanded stats |
| `t` | Toggle dark/light theme |
| `/` | Filter games |
| `q`/`esc` | Back / quit |

## Customizations

Forked from [NolanFogarty/courtside](https://github.com/NolanFogarty/courtside) and personalized with:

- Team-specific colors (each team renders in their brand color)
- Quarter linescores in the game list
- Four Factors + advanced team stats in the comparison bar
- ASCII game flow chart (plaintextsports-style dot chart)
- Lead changes / ties / biggest lead summary
- Player stat leaders row per team
- Dark/light theme toggle

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lipgloss](https://github.com/charmbracelet/lipgloss).
