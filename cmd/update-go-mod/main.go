// Package main runs the update-go-mod GitHub Action.
package main

import (
	"log/slog"
	"os"
	"strings"

	"github.com/faisal-memon/update-go-mod-action/internal/updater"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	if err := updater.Run(updater.Config{
		GoModPath:       getenv("INPUT_GO_MOD_PATH", "go.mod"),
		UpdateToolchain: parseBool(getenv("INPUT_UPDATE_TOOLCHAIN", "false")),
		GitHubOutput:    os.Getenv("GITHUB_OUTPUT"),
		Logger:          log,
	}); err != nil {
		log.Error("update go.mod", "error", err)
		os.Exit(1)
	}
}

// getenv returns an environment variable value or a fallback when it is unset.
func getenv(name, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	return value
}

// parseBool accepts common truthy input values used by GitHub Actions.
func parseBool(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
