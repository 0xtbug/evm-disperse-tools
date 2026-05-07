package update

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/0xtbug/evm-disperse-tools/internal/version"
)

const (
	githubAPIURL = "https://api.github.com/repos/0xtbug/evm-disperse-tools/releases/latest"
	httpTimeout  = 5 * time.Second
)

// Result holds the outcome of an update check.
type Result struct {
	CurrentVersion string
	LatestVersion  string
	HasUpdate      bool
	ReleaseURL     string
}

// githubRelease is the minimal JSON structure from the GitHub Releases API.
type githubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

// Check queries the GitHub Releases API and compares against the current version.
// Returns nil error with a valid Result even when no update is found.
func Check() (*Result, error) {
	current := version.Version
	if current == "" {
		current = "dev"
	}

	client := &http.Client{Timeout: httpTimeout}
	req, err := http.NewRequest("GET", githubAPIURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "evm-disperse-tools/"+current)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	latest := normalizeVersion(release.TagName)
	cur := normalizeVersion(current)

	return &Result{
		CurrentVersion: current,
		LatestVersion:  release.TagName,
		HasUpdate:      latest != cur && isGreater(latest, cur),
		ReleaseURL:     release.HTMLURL,
	}, nil
}

// normalizeVersion strips leading "v" prefix.
func normalizeVersion(v string) string {
	return strings.TrimPrefix(v, "v")
}

// isGreater performs a simple semver comparison (major.minor.patch).
// Returns true if a > b.
func isGreater(a, b string) bool {
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")

	// Pad to 3 parts
	for len(aParts) < 3 {
		aParts = append(aParts, "0")
	}
	for len(bParts) < 3 {
		bParts = append(bParts, "0")
	}

	for i := 0; i < 3; i++ {
		aNum, bNum := 0, 0
		fmt.Sscanf(aParts[i], "%d", &aNum)
		fmt.Sscanf(bParts[i], "%d", &bNum)
		if aNum > bNum {
			return true
		}
		if aNum < bNum {
			return false
		}
	}
	return false
}
