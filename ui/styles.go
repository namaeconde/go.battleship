// ui/styles.go
package ui

import "github.com/charmbracelet/lipgloss"

// Ocean Deep colour palette.
var (
	colorBg      = lipgloss.Color("#071020")
	colorTitleBg = lipgloss.Color("#0a1e3d")
	colorAccent  = lipgloss.Color("#5baee8")
	colorBorder  = lipgloss.Color("#1a4a80")
	colorWater   = lipgloss.Color("#1a5a8a")
	colorShip    = lipgloss.Color("#5baee8")
	colorHit     = lipgloss.Color("#ff6b6b")
	colorMiss    = lipgloss.Color("#2a4a6a")
	colorLabel   = lipgloss.Color("#3a7ab0")
	colorStatus  = lipgloss.Color("#a8c8e8")

	// Animation frame colours — hit
	colorAnimHit0 = lipgloss.Color("#ffff00") // bright yellow
	colorAnimHit1 = lipgloss.Color("#ff8800") // orange
	colorAnimHit2 = lipgloss.Color("#ff4444") // red

	// Animation frame colours — miss
	colorAnimMiss0 = lipgloss.Color("#ffffff") // bright white
	colorAnimMiss1 = lipgloss.Color("#aaaaaa") // light grey
	colorAnimMiss2 = lipgloss.Color("#555555") // mid grey
)

var (
	titleStyle = lipgloss.NewStyle().
			Background(colorTitleBg).
			Foreground(colorAccent).
			Bold(true).
			Padding(0, 2)

	phaseBadgeStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccent).
			Padding(0, 1)

	boardPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)

	boardTitleStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	messageStyle = lipgloss.NewStyle().
			Foreground(colorStatus)

	actionBarStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(colorBorder).
			Padding(0, 1)

	shipIntactStyle = lipgloss.NewStyle().
			Foreground(colorAccent)

	shipSunkStyle = lipgloss.NewStyle().
			Foreground(colorHit).
			Strikethrough(true)

	shipUnplacedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#333a4a"))
)
