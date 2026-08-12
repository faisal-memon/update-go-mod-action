package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"golang.org/x/mod/semver"
)

const goReleasesURL = "https://go.dev/dl/?mode=json"

var (
	goLineRE        = regexp.MustCompile(`(?m)^([ \t]*)go[ \t]+([^ \t\r\n]+)[ \t]*$`)
	toolchainLineRE = regexp.MustCompile(`(?m)^([ \t]*)toolchain[ \t]+go([^ \t\r\n]+)[ \t]*$`)
)

type Config struct {
	GoModPath       string
	UpdateToolchain bool
	GitHubOutput    string
	HTTPClient      *http.Client
	Logger          *slog.Logger
	hooks           hooks
}

type hooks struct {
	fetchReleases func() ([]byte, error)
}

type Release struct {
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
}

type Result struct {
	Changed         bool
	PreviousVersion string
	CurrentVersion  string
	LatestVersion   string
}

func Run(config Config) error {
	result, err := update(config)
	if err != nil {
		return err
	}

	if result.Changed {
		config.Logger.Info(
			"updated go.mod",
			"path", config.GoModPath,
			"previous_version", result.PreviousVersion,
			"latest_version", result.LatestVersion,
		)
	} else {
		config.Logger.Info(
			"go.mod is already on the latest stable Go version",
			"path", config.GoModPath,
			"current_version", result.PreviousVersion,
		)
	}

	return writeOutputs(config.GitHubOutput, result)
}

func update(config Config) (Result, error) {
	if config.GoModPath == "" {
		return Result{}, fmt.Errorf("go.mod path is required")
	}

	contents, err := os.ReadFile(config.GoModPath)
	if err != nil {
		return Result{}, fmt.Errorf("read go.mod: %w", err)
	}

	original := string(contents)
	goDirective, ok := findDirective(original, goLineRE)
	if !ok {
		return Result{}, fmt.Errorf(`could not find a "go" directive in %s`, config.GoModPath)
	}

	latestVersion, err := latestStableVersion(config)
	if err != nil {
		return Result{}, err
	}

	previousVersion := goDirective.version
	changed, err := versionLess(previousVersion, latestVersion)
	if err != nil {
		return Result{}, err
	}

	currentVersion := previousVersion
	if changed {
		updated := replaceDirective(original, goDirective, fmt.Sprintf("%sgo %s", goDirective.indent, latestVersion))

		if config.UpdateToolchain {
			if toolchainDirective, ok := findDirective(updated, toolchainLineRE); ok {
				updated = replaceDirective(updated, toolchainDirective, fmt.Sprintf("%stoolchain go%s", toolchainDirective.indent, latestVersion))
			}
		}

		if err := os.WriteFile(config.GoModPath, []byte(updated), 0o644); err != nil {
			return Result{}, fmt.Errorf("write go.mod: %w", err)
		}

		currentVersion = latestVersion
	}

	return Result{
		Changed:         changed,
		PreviousVersion: previousVersion,
		CurrentVersion:  currentVersion,
		LatestVersion:   latestVersion,
	}, nil
}

type directive struct {
	start   int
	end     int
	indent  string
	version string
}

func findDirective(contents string, expression *regexp.Regexp) (directive, bool) {
	match := expression.FindStringSubmatchIndex(contents)
	if match == nil {
		return directive{}, false
	}

	return directive{
		start:   match[0],
		end:     match[1],
		indent:  contents[match[2]:match[3]],
		version: contents[match[4]:match[5]],
	}, true
}

func replaceDirective(contents string, directive directive, replacement string) string {
	return contents[:directive.start] + replacement + contents[directive.end:]
}

func latestStableVersion(config Config) (string, error) {
	releases, err := loadReleases(config)
	if err != nil {
		return "", err
	}

	latest := ""
	for _, release := range releases {
		if !release.Stable || !strings.HasPrefix(release.Version, "go") {
			continue
		}

		version := strings.TrimPrefix(release.Version, "go")
		if latest == "" {
			latest = version
			continue
		}

		less, err := versionLess(latest, version)
		if err != nil {
			return "", err
		}
		if less {
			latest = version
		}
	}

	if latest == "" {
		return "", fmt.Errorf("no stable Go releases found from go.dev")
	}

	return latest, nil
}

func loadReleases(config Config) ([]Release, error) {
	var data []byte
	if config.hooks.fetchReleases != nil {
		var err error
		data, err = config.hooks.fetchReleases()
		if err != nil {
			return nil, err
		}
	} else {
		client := config.HTTPClient
		if client == nil {
			client = &http.Client{Timeout: 30 * time.Second}
		}

		response, err := client.Get(goReleasesURL)
		if err != nil {
			return nil, fmt.Errorf("fetch Go releases: %w", err)
		}
		defer response.Body.Close()

		if response.StatusCode < 200 || response.StatusCode > 299 {
			return nil, fmt.Errorf("fetch Go releases: unexpected HTTP status %s", response.Status)
		}

		data, err = io.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("read Go releases response: %w", err)
		}
	}

	var releases []Release
	if err := json.Unmarshal(data, &releases); err != nil {
		return nil, fmt.Errorf("parse Go releases JSON: %w", err)
	}

	return releases, nil
}

func versionLess(left, right string) (bool, error) {
	leftVersion, err := toSemver(left)
	if err != nil {
		return false, err
	}

	rightVersion, err := toSemver(right)
	if err != nil {
		return false, err
	}

	return semver.Compare(leftVersion, rightVersion) < 0, nil
}

func toSemver(version string) (string, error) {
	cleaned := strings.TrimSpace(version)
	if cleaned == "" {
		return "", fmt.Errorf("empty Go version")
	}

	cleaned = strings.TrimPrefix(cleaned, "go")
	if !strings.HasPrefix(cleaned, "v") {
		cleaned = "v" + cleaned
	}

	if !semver.IsValid(cleaned) {
		return "", fmt.Errorf("invalid Go version %q", version)
	}

	return cleaned, nil
}

func writeOutputs(path string, result Result) error {
	if path == "" {
		return nil
	}

	output := strings.Builder{}
	output.WriteString(fmt.Sprintf("changed=%t\n", result.Changed))
	output.WriteString(fmt.Sprintf("previous-version=%s\n", result.PreviousVersion))
	output.WriteString(fmt.Sprintf("current-version=%s\n", result.CurrentVersion))
	output.WriteString(fmt.Sprintf("latest-version=%s\n", result.LatestVersion))

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open GitHub output file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(output.String()); err != nil {
		return fmt.Errorf("write GitHub outputs: %w", err)
	}

	return nil
}
