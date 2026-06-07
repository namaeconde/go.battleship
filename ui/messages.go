// ui/messages.go
package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// tickMsg is sent on each animation tick.
type tickMsg struct{}

// tickCmd returns a tea.Cmd that fires a tickMsg after 100ms.
func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}
