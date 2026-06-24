// Package updater implements Drishti's NOTIFY-ONLY update surface. It never
// downloads or installs anything (spec D12). The opt-in GitHub Releases check
// (Check) is invoked only when the user enables auto_check or sends ?check=1;
// any failure degrades silently to CurrentStatus (spec §14).
package updater

import (
	"context"
	"encoding/json"
	"net/http"
	"runtime"
	"strconv"
	"strings"
)

// releaseURL is the GitHub Releases latest endpoint. A wrong or unreachable
// URL silently yields no-update (degrade, don't die, §14).
const releaseURL = "https://api.github.com/repos/saptarshi369/drishti/releases/latest"

// Status is the payload served at /api/update/status.
type Status struct {
	Current   string   `json:"current"`
	Latest    string   `json:"latest"`
	Available bool     `json:"available"`
	Commands  []string `json:"commands"`
}

// CurrentStatus returns the slice's stub status: current version, no known
// newer release, and copy-pasteable upgrade commands for this OS.
func CurrentStatus(version string) Status {
	return Status{
		Current:   version,
		Latest:    "",
		Available: false,
		Commands:  upgradeCommands(),
	}
}

// upgradeCommands returns the build-from-source commands for the running OS.
// We detect the environment to print the right commands; we NEVER execute them.
func upgradeCommands() []string {
	if runtime.GOOS == "windows" {
		return []string{"git pull", "go build -o drishti.exe ./cmd/drishti"}
	}
	return []string{"git pull", "go build -o drishti ./cmd/drishti"}
}

// Check queries GitHub for the latest release and returns a Status. It is the
// ONLY outbound call in Drishti and runs only when the user opts in via
// auto_check=true or ?check=1 (spec §6). Any network failure, bad status code,
// or unparseable body degrades silently to CurrentStatus — never errors, never
// panics (spec §14).
func Check(ctx context.Context, current string, client *http.Client) Status {
	return checkAt(ctx, current, client, releaseURL)
}

// checkAt is the internal implementation of Check with an injectable URL so
// tests can point at a local httptest server instead of GitHub. This is a
// common Go pattern: export the public API (Check) and test the mechanism
// through a small seam (checkAt).
func checkAt(ctx context.Context, current string, client *http.Client, url string) Status {
	// Build the GET request; attach the context so the caller can cancel it
	// (e.g. the handler passes a 5-second timeout context).
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		// Malformed URL or nil context — degrade silently.
		return CurrentStatus(current)
	}

	resp, err := client.Do(req)
	if err != nil {
		// Network failure (offline, DNS error, timeout) — degrade silently.
		return CurrentStatus(current)
	}
	defer func() { _ = resp.Body.Close() }() // Always drain + close; error is intentionally ignored per Go HTTP conventions.

	if resp.StatusCode != http.StatusOK {
		// GitHub may return 403 (rate-limit) or 404 (no releases yet).
		// Any non-200 is treated as "no update available".
		return CurrentStatus(current)
	}

	// We only need tag_name from the response; decode into an anonymous struct
	// to avoid pulling in the full GitHub API shape.
	var body struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil || body.TagName == "" {
		// Malformed JSON or missing field — degrade silently.
		return CurrentStatus(current)
	}

	// Return a full Status with the live latest tag and the comparison result.
	return Status{
		Current:   current,
		Latest:    body.TagName,
		Available: compareVersions(current, body.TagName),
		Commands:  upgradeCommands(),
	}
}

// compareVersions reports whether latest is a strictly newer semver than
// current. Returns false for "dev" builds or any unparseable input (§14).
func compareVersions(current, latest string) bool {
	cur, ok1 := parseSemver(current)
	lat, ok2 := parseSemver(latest)
	if !ok1 || !ok2 {
		return false
	}
	for i := 0; i < 3; i++ {
		if lat[i] != cur[i] {
			return lat[i] > cur[i]
		}
	}
	return false
}

// parseSemver parses "vX.Y.Z" (leading v optional) into a [3]int array.
// Returns ok=false for malformed input including "dev" or plain strings.
func parseSemver(s string) ([3]int, bool) {
	s = strings.TrimPrefix(s, "v")
	parts := strings.SplitN(s, ".", 3)
	if len(parts) != 3 {
		return [3]int{}, false
	}
	var out [3]int
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return [3]int{}, false
		}
		out[i] = n
	}
	return out, true
}
