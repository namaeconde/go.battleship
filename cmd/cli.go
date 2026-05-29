package cmd

import (
	"errors"
	"flag"
	"fmt"
)

// Config holds the parsed command-line arguments.
type Config struct {
	IsHost        bool
	Port          string
	RemoteAddress string // IP:Port for join mode
}

// ParseArgs parses the command-line arguments.
func ParseArgs(args []string) (*Config, error) {
	if len(args) < 2 {
		return nil, errors.New("usage: battleship <mode> [-p <port>] [-a <address>]")
	}

	config := &Config{}
	mode := args[1]

	switch mode {
	case "host":
		config.IsHost = true
		hostCmd := flag.NewFlagSet("host", flag.ContinueOnError)
		hostPort := hostCmd.String("p", "", "Port to listen on (e.g., 8080)")
		hostCmd.Parse(args[2:])
		config.Port = *hostPort
		if config.Port == "" {
			return nil, errors.New("port is required for host mode")
		}
	case "join":
		config.IsHost = false
		joinCmd := flag.NewFlagSet("join", flag.ContinueOnError)
		joinAddress := joinCmd.String("a", "", "Remote host address (e.g., 127.0.0.1:8080)")
		joinCmd.Parse(args[2:])
		config.RemoteAddress = *joinAddress
		if config.RemoteAddress == "" {
			return nil, errors.New("address is required for join mode")
		}
	default:
		return nil, fmt.Errorf("invalid mode: please specify 'host' or 'join'")
	}

	return config, nil
}
