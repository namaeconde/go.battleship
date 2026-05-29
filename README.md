# Go Battleship

A 2-player networked terminal-based Battleship game implemented in Go.

## Features

*   Interactive terminal UI using `tcell`.
*   Peer-to-Peer (P2P) TCP networking for 2 players.
*   Ship placement and battle phases.

## How to Run

### Host
```bash
go run . host -p 8080
```

### Join
```bash
go run . join -a 127.0.0.1:8080
```

## Gameplay Instructions

1.  **Placement Phase**: Use arrow keys to navigate and place ships. Press `Enter` to place/set ready, `Spacebar` to remove a ship.
2.  **Battle Phase**: Use arrow keys to navigate the tracking board (right side). Press `Spacebar` to confirm a target (highlighted in orange), then `Enter` to fire. The host always fires first; players alternate turns.
3.  **Exit**: Press `Esc` or `Ctrl+C` at any time to exit.
