# Go Battleship

A 2-player networked terminal-based Battleship game implemented in Go.

This is a sandbox project to test out gemini/claude when using superpowers extension.

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
3.  **Game Over**: Press `R` to play again (both players must agree) or `Q` to quit.
4.  **Exit**: Press `Esc` or `Ctrl+C` at any time to exit.

## Screenshots

### Battle Phase

Player 1 (left board: your ships, right board: your shots against opponent):

![Battle Phase - Player 1](battle_phase_1.png)

Player 2 perspective after firing:

![Battle Phase - Player 2](battle_phase_2.png)

### Game Over

Winner screen:

![Game Over - You Win](game_over_screen_2.png)

Loser screen:

![Game Over - You Lose](game_over_screen_1.png)

## Built With

## Authors

* **Namae Conde** - *Initial work* - [namaeconde][githublink]
* **Gemini/Claude**

[githublink]: https://github.com/namaeconde