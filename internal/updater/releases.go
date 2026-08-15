package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	goReleasesURL     = "https://go.dev/dl/?mode=json"
	goReleasesTimeout = 30 * time.Second
)

type Release struct {
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
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

		comparison, err := semverCompare(latest, version)
		if err != nil {
			return "", err
		}
		if comparison < 0 {
			latest = version
		}
	}

	if latest == "" {
		return "", fmt.Errorf("no stable Go releases found from go.dev")
	}

	return latest, nil
}

func loadReleases(config Config) ([]Release, error) {
	client := config.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: goReleasesTimeout}
	}

	response, err := client.Get(goReleasesURL)
	if err != nil {
		return nil, fmt.Errorf("fetch Go releases: %w", err)
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode < 200 || response.StatusCode > 299 {
		return nil, fmt.Errorf("fetch Go releases: unexpected HTTP status %s", response.Status)
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read Go releases response: %w", err)
	}

	var releases []Release
	if err := json.Unmarshal(data, &releases); err != nil {
		return nil, fmt.Errorf("parse Go releases JSON: %w", err)
	}

	return releases, nil
}
