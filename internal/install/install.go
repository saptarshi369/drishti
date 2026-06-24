// Package install generates non-mutating settings.json proposals for the
// Drishti live-helper. It reads the user's existing ~/.claude/settings.json,
// merges Drishti's additions (currently just the statusLine entry), and
// returns the pretty-printed proposal for the user to copy and apply
// themselves. Drishti NEVER writes the user's settings.json — it only ever
// writes its own ~/.drishti/bin/ helper scripts (see EnsureHelperScripts).
// Design spec §1, §6.
package install

import (
	_ "embed" // required for //go:embed directives
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// statuslineHelperSH holds the embedded source of the statusline-helper.sh
// script. This is the authoritative installed copy — it is written into
// ~/.drishti/bin/ by EnsureHelperScripts so the path referenced in the
// proposed settings.json always exists before the user applies it.
//
//go:embed statusline-helper.sh
var statuslineHelperSH []byte

// ProposeSettingsJSON merges Drishti's statusLine entry into a copy of the
// user's existing settings.json. existing may be nil or empty (treated as
// {}) — the "fresh install" path. It returns:
//
//   - proposed: the pretty-printed merged JSON document.
//   - added: the list of top-level keys Drishti added (the additions-only diff).
//   - err: non-nil only when existing is non-empty but not valid JSON.
//
// The function is idempotent: re-running on an already-merged document adds
// nothing and returns an empty added slice. It is non-breaking by
// construction: it only ADDs keys into the user's existing map; it never
// removes or alters any existing keys.
func ProposeSettingsJSON(existing []byte, binDir string) (proposed []byte, added []string, err error) {
	// Start from an empty map; merge the user's existing document over it.
	doc := map[string]any{}
	if len(existing) > 0 {
		// Unmarshal fails loudly so the user never gets a silently corrupted
		// proposal from a pre-existing invalid file.
		if merr := json.Unmarshal(existing, &doc); merr != nil {
			return nil, nil, fmt.Errorf("existing settings.json is not valid JSON: %w", merr)
		}
	}

	// Add statusLine only when absent — idempotent guard prevents double-add.
	// binDir is the absolute path to ~/.drishti/bin/ where EnsureHelperScripts
	// writes the helper; filepath.Join produces the correct platform path.
	if _, ok := doc["statusLine"]; !ok {
		doc["statusLine"] = map[string]any{
			"type":    "command",
			"command": filepath.Join(binDir, "statusline-helper.sh"),
		}
		added = append(added, "statusLine")
	}

	// Marshal with two-space indentation so the user can read the proposal
	// before pasting it into their settings.json.
	out, merr := json.MarshalIndent(doc, "", "  ")
	if merr != nil {
		return nil, nil, merr
	}
	return out, added, nil
}

// EnsureHelperScripts writes Drishti's helper scripts into binDir (its OWN
// directory, e.g. ~/.drishti/bin/). It is idempotent: it creates binDir with
// 0o755 if absent, then writes statusline-helper.sh with 0o755 if the file
// does not already exist. It never overwrites an existing file — the script
// content is fixed at build time via the embedded copy (statuslineHelperSH).
func EnsureHelperScripts(binDir string) error {
	// MkdirAll is a no-op when the directory already exists. 0o755 matches the
	// standard executable-directory permission on macOS and Linux.
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return fmt.Errorf("create bin dir %s: %w", binDir, err)
	}

	// Write the statusline helper. O_EXCL causes os.OpenFile to return an
	// error if the file already exists — we interpret that as "already
	// installed" and treat it as success (idempotent).
	scriptPath := filepath.Join(binDir, "statusline-helper.sh")
	f, err := os.OpenFile(scriptPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o755)
	if err != nil {
		if os.IsExist(err) {
			// Already present — nothing to do.
			return nil
		}
		return fmt.Errorf("create statusline-helper.sh: %w", err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.Write(statuslineHelperSH); err != nil {
		return fmt.Errorf("write statusline-helper.sh: %w", err)
	}
	return nil
}
