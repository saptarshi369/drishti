package services

import (
	"github.com/saptarshi369/drishti/internal/model"
	"github.com/saptarshi369/drishti/internal/resolve"
	"github.com/saptarshi369/drishti/internal/security"
	claude "github.com/saptarshi369/drishti/internal/sources/claude"
)

// InventoryStore is the narrow slice of the store that RefreshInventory needs.
// Defining a minimal interface here (rather than depending on the concrete
// *store.Store) keeps the inventory service trivially testable: any struct that
// provides these methods satisfies the interface. This is the standard Go
// "accept interfaces, return concretes" idiom for testability (§10, spec).
type InventoryStore interface {
	ReplaceInventory(agent, projectRoot string, items []model.InventoryItem) error
	ReplaceResolved(agent, projectRoot string, resolved []model.ResolvedItem) error
	// ReplaceSecurityFindings atomically replaces all security findings for one
	// (agent, projectRoot) pair. Added in Task 9 so the security rule engine
	// output is persisted alongside inventory/resolved data.
	ReplaceSecurityFindings(agent, projectRoot string, findings []model.Finding) error
}

// RefreshInventory is the service-layer orchestrator for the inventory pipeline:
// discover raw items → persist raw → resolve → persist resolved → evaluate
// security rules → persist findings. It is NOT an ingest/watcher step; it is
// the service function called on demand (startup, file-change event) per spec §10.
//
// For each location, RefreshInventory:
//  1. Builds a SecretMatcher from the rules' configured keyword/prefix lists.
//  2. Calls claude.Discover to collect raw InventoryItems and SecurityInputs from
//     the filesystem.
//  3. Writes the raw items to the store (keyed by "claude" agent + ProjectRoot).
//  4. Calls resolve.Resolve to compute the effective/precedence view.
//  5. Writes the resolved items to the store.
//  6. Calls rules.Evaluate on the SecurityInputs and persists the findings.
//
// A location that yields zero items is normal success (user may have no skills).
// Any hard error (DB write failure, unexpected discover error) is returned
// immediately; subsequent locations are not processed.
func RefreshInventory(st InventoryStore, locs []claude.Locations, rules security.Rules) error {
	// Build the SecretMatcher once from the rules' configured lists. When the lists
	// are empty, SecretMatcher falls back to its built-in detection heuristics, so
	// passing empty slices is safe. Building it once outside the loop avoids
	// recomputing the same matcher for every location.
	sm := claude.SecretMatcher{
		Keywords: rules.SecretKeywords(),
		Prefixes: rules.SecretPrefixes(),
	}

	for _, loc := range locs {
		// Discover returns (items, toggles, securityInputs, err). The toggles feed
		// resolve so disable/enable overrides (e.g. settings.json) are reflected in
		// the effective status. SecurityInputs carries the privacy-safe audit data
		// (permission shapes, MCP env key names, plugin sources) for the rule engine.
		items, tg, sec, err := claude.Discover(loc, sm)
		if err != nil {
			return err
		}

		// ProjectRoot is "" for the user-global location; the store uses that
		// as the key for user-scoped rows. Per spec §10, each project gets its
		// own resolved view so we must pass the root here.
		root := loc.ProjectRoot

		if err := st.ReplaceInventory("claude", root, items); err != nil {
			return err
		}

		// resolve.Resolve is a pure function: given raw items + toggles, it
		// returns the materialized "effective" view with precedence trails.
		resolved := resolve.Resolve(items, tg)
		if err := st.ReplaceResolved("claude", root, resolved); err != nil {
			return err
		}

		// Evaluate all enabled security rules against the parsed inputs, then
		// persist the findings. An empty finding slice is a valid (clean) result.
		findings := rules.Evaluate(sec)
		if err := st.ReplaceSecurityFindings("claude", root, findings); err != nil {
			return err
		}
	}
	return nil
}
