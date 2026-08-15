package updater

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"regexp"
)

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
	latestStableVersion func() (string, error)
}

func Run(config Config) error {
	config = config.withDefaults()

	original, fileInfo, err := readGoMod(config.GoModPath)
	if err != nil {
		return err
	}

	goDirective, ok := findDirective(original, goLineRE)
	if !ok {
		return fmt.Errorf(`could not find a "go" directive in %s`, config.GoModPath)
	}

	latestVersion, err := config.hooks.latestStableVersion()
	if err != nil {
		return err
	}

	previousVersion := goDirective.version
	comparison, err := semverCompare(previousVersion, latestVersion)
	if err != nil {
		return err
	}

	changed := comparison < 0
	updatedVersion := previousVersion
	if changed {
		updated := replaceDirective(original, goDirective, fmt.Sprintf("%sgo %s", goDirective.indent, latestVersion))

		if config.UpdateToolchain {
			if toolchainDirective, ok := findDirective(updated, toolchainLineRE); ok {
				updated = replaceDirective(updated, toolchainDirective, fmt.Sprintf("%stoolchain go%s", toolchainDirective.indent, latestVersion))
			}
		}

		if err := os.WriteFile(config.GoModPath, []byte(updated), fileInfo.Mode().Perm()); err != nil {
			return fmt.Errorf("write go.mod: %w", err)
		}

		updatedVersion = latestVersion

		config.Logger.Info(
			"updated go.mod",
			"path", config.GoModPath,
			"previous_version", previousVersion,
			"updated_version", updatedVersion,
		)
	} else {
		config.Logger.Info(
			"go.mod is already on the latest stable Go version",
			"path", config.GoModPath,
			"updated_version", updatedVersion,
		)
	}

	return writeOutputs(config.GitHubOutput, changed, previousVersion, updatedVersion)
}

func readGoMod(path string) (string, os.FileInfo, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return "", nil, fmt.Errorf("read go.mod: %w", err)
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		return "", nil, fmt.Errorf("stat go.mod: %w", err)
	}

	return string(contents), fileInfo, nil
}

func (config Config) withDefaults() Config {
	if config.hooks.latestStableVersion == nil {
		config.hooks.latestStableVersion = func() (string, error) {
			return latestStableVersion(config)
		}
	}

	return config
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

func writeOutputs(path string, changed bool, previousVersion, updatedVersion string) error {
	if path == "" {
		return nil
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open GitHub output file: %w", err)
	}

	if _, err := fmt.Fprintf(
		file,
		"changed=%t\nprevious-version=%s\nupdated-version=%s\n",
		changed,
		previousVersion,
		updatedVersion,
	); err != nil {
		return fmt.Errorf("write GitHub outputs: %w", err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("close GitHub output file: %w", err)
	}

	return nil
}
