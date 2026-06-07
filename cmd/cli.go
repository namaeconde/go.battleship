// cmd/cli.go
package cmd

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"go.battleship/game"
	"go.battleship/network"
	"go.battleship/ui"
)

var serverURL string

var rootCmd = &cobra.Command{
	Use:   "battleship",
	Short: "Terminal Battleship",
}

var hostCmd = &cobra.Command{
	Use:   "host",
	Short: "Create a new game and wait for opponent",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateServerURL(); err != nil {
			return err
		}
		conn, gameID, err := network.CreateGame(context.Background(), serverURL)
		if err != nil {
			return fmt.Errorf("creating game: %w", err)
		}
		fmt.Printf("Game created! Share this code with your opponent: %s\n", gameID)
		fmt.Println("Waiting for opponent to join...")

		if _, err = conn.Receive(); err != nil {
			return fmt.Errorf("waiting for opponent: %w", err)
		}
		fmt.Println("Opponent joined! Starting game...")

		RunGame(conn, true)
		return nil
	},
}

var joinCmd = &cobra.Command{
	Use:   "join GAME_ID",
	Short: "Join an existing game",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateServerURL(); err != nil {
			return err
		}
		conn, err := network.JoinGame(context.Background(), serverURL, args[0])
		if err != nil {
			return fmt.Errorf("joining game: %w", err)
		}
		fmt.Printf("Joined game %s! Starting game...\n", args[0])
		RunGame(conn, false)
		return nil
	},
}

func init() {
	godotenv.Load()
	defaultServer := os.Getenv("SERVER_URL")
	hostCmd.Flags().StringVar(&serverURL, "server", defaultServer, "gRPC relay server URL (or set SERVER_URL env var)")
	joinCmd.Flags().StringVar(&serverURL, "server", defaultServer, "gRPC relay server URL (or set SERVER_URL env var)")
	rootCmd.AddCommand(hostCmd, joinCmd)
}

func validateServerURL() error {
	if serverURL == "" {
		return fmt.Errorf("server URL is required: use --server flag or set SERVER_URL environment variable")
	}
	return nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func RunGame(conn game.NetworkConn, host bool) {
	m := ui.NewModel(conn, host)
	p := tea.NewProgram(m, tea.WithAltScreen())
	m.SetProgram(p)
	defer m.Close()
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "battleship: %v\n", err)
		os.Exit(1)
	}
}
