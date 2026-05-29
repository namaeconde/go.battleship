package cmd

import (
	"testing"
)

func TestParseArgs_HostMode(t *testing.T) {
	// Simulate command line arguments for host mode
	osArgs := []string{"battleship", "host", "-p", "8080"}
	config, err := ParseArgs(osArgs)

	if err != nil {
		t.Fatalf("ParseArgs failed: %v", err)
	}
	if !config.IsHost {
		t.Errorf("Expected IsHost to be true, got false")
	}
	if config.Port != "8080" {
		t.Errorf("Expected Port '8080', got '%s'", config.Port)
	}
	if config.RemoteAddress != "" {
		t.Errorf("Expected RemoteAddress to be empty, got '%s'", config.RemoteAddress)
	}
}

func TestParseArgs_JoinMode(t *testing.T) {
	// Simulate command line arguments for join mode
	osArgs := []string{"battleship", "join", "-a", "127.0.0.1:8080"}
	config, err := ParseArgs(osArgs)

	if err != nil {
		t.Fatalf("ParseArgs failed: %v", err)
	}
	if config.IsHost {
		t.Errorf("Expected IsHost to be false, got true")
	}
	if config.RemoteAddress != "127.0.0.1:8080" {
		t.Errorf("Expected RemoteAddress '127.0.0.1:8080', got '%s'", config.RemoteAddress)
	}
	if config.Port != "" {
		t.Errorf("Expected Port to be empty, got '%s'", config.Port)
	}
}

func TestParseArgs_InvalidMode(t *testing.T) {
	osArgs := []string{"battleship", "invalid"}
	_, err := ParseArgs(osArgs)
	if err == nil {
		t.Fatal("Expected error for invalid mode, got nil")
	}
	if err.Error() != "invalid mode: please specify 'host' or 'join'" {
		t.Errorf("Expected 'invalid mode' error, got '%v'", err.Error())
	}
}

func TestParseArgs_MissingAddressForJoin(t *testing.T) {
	osArgs := []string{"battleship", "join"}
	_, err := ParseArgs(osArgs)
	if err == nil {
		t.Fatal("Expected error for missing address in join mode, got nil")
	}
	if err.Error() != "address is required for join mode" {
		t.Errorf("Expected 'address is required' error, got '%v'", err.Error())
	}
}

func TestParseArgs_MissingPortForHost(t *testing.T) {
	osArgs := []string{"battleship", "host"}
	_, err := ParseArgs(osArgs)
	if err == nil {
		t.Fatal("Expected error for missing port in host mode, got nil")
	}
	if err.Error() != "port is required for host mode" {
		t.Errorf("Expected 'port is required' error, got '%v'", err.Error())
	}
}
