package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/faisal-memon/update-go-mod-action/internal/updater"
)

func main() {
	config := updater.Config{
		GoModPath:       getenv("INPUT_GO_MOD_PATH", "go.mod"),
		UpdateToolchain: parseBool(getenv("INPUT_UPDATE_TOOLCHAIN", "false")),
		GitHubOutput:    os.Getenv("GITHUB_OUTPUT"),
	}

	if err := updater.Run(config); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func getenv(name, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	return value
}

func parseBool(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
