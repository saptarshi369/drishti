package claude

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/saptarshi369/drishti/internal/model"
)

// settingsFile is the subset of settings.json / settings.local.json we read.
type settingsFile struct {
	Hooks                  map[string][]hookGroup `json:"hooks"`
	DisableBundledSkills   bool                   `json:"disableBundledSkills"`
	SkillOverrides         map[string]string      `json:"skillOverrides"`
	DisabledMcpjsonServers []string               `json:"disabledMcpjsonServers"`
	EnabledMcpjsonServers  []string               `json:"enabledMcpjsonServers"`
	EnableAllProjectMcp    bool                   `json:"enableAllProjectMcpServers"`
	EnabledPlugins         map[string]bool        `json:"enabledPlugins"`
	OutputStyle            string                 `json:"outputStyle"`
	ClaudeMdExcludes       []string               `json:"claudeMdExcludes"`
	Permissions            permissionsBlock       `json:"permissions"`
}

// permissionsBlock is the subset of the settings "permissions" object the audit
// reads. It maps directly to Claude Code's permissions JSON shape.
type permissionsBlock struct {
	Deny        []string `json:"deny"`
	Allow       []string `json:"allow"`
	Ask         []string `json:"ask"`
	DefaultMode string   `json:"defaultMode"`
}

// hookGroup is one matcher entry under an event: a matcher plus its handlers.
type hookGroup struct {
	Matcher string        `json:"matcher"`
	Hooks   []hookHandler `json:"hooks"`
}

// hookHandler is a single hook action inside a hookGroup.
type hookHandler struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

// ParseSettings extracts hook InventoryItems, the resolution Toggles, and the
// ScopePermissions from one settings file. Hooks across an event/matcher are
// flattened to one item per handler; the display name is "<event> · <matcher>"
// (matcher omitted when empty). Events are processed in sorted order for
// deterministic output. The ScopePermissions captures the deny/allow/ask/
// defaultMode fields plus any setting keys whose values looked like secrets per
// sm (values are inspected then discarded — only key paths are returned).
func ParseSettings(content []byte, scope model.Scope, relPath string, sm SecretMatcher) ([]model.InventoryItem, model.Toggles, model.ScopePermissions, error) {
	var f settingsFile
	if err := json.Unmarshal(content, &f); err != nil {
		return nil, model.Toggles{}, model.ScopePermissions{}, err
	}

	// Collect event names and sort for deterministic output order.
	events := make([]string, 0, len(f.Hooks))
	for ev := range f.Hooks {
		events = append(events, ev)
	}
	sort.Strings(events)

	// Build one InventoryItem per hook handler, grouped by event + matcher.
	var hooks []model.InventoryItem
	for _, ev := range events {
		for _, grp := range f.Hooks[ev] {
			// Name is "<event>" when no matcher, "<event> · <matcher>" when set.
			name := ev
			if grp.Matcher != "" {
				name = ev + " · " + grp.Matcher
			}
			for _, h := range grp.Hooks {
				hooks = append(hooks, model.InventoryItem{
					AgentCode: "claude",
					Category:  model.CatHook,
					Name:      name,
					Scope:     scope,
					RelPath:   relPath,
					Enabled:   true,
					// Attrs stores hook-specific fields so the schema stays flat
					// (see model.InventoryItem for the convention).
					Attrs: map[string]string{
						"event":   ev,
						"matcher": grp.Matcher,
						"command": h.Command,
					},
				})
			}
		}
	}

	// enabledPlugins maps "name@marketplace" -> bool. We emit one InventoryItem
	// per entry (keeping disabled ones so the Plugins category can show them) and
	// also keep the enabled-only slice for any toggle consumers. Names are sorted
	// for deterministic output.
	pluginNames := make([]string, 0, len(f.EnabledPlugins))
	for k := range f.EnabledPlugins {
		pluginNames = append(pluginNames, k)
	}
	sort.Strings(pluginNames)
	var plugins []string
	for _, name := range pluginNames {
		on := f.EnabledPlugins[name]
		if on {
			plugins = append(plugins, name)
		}
		hooks = append(hooks, model.InventoryItem{
			AgentCode: "claude",
			Category:  model.CatPlugin,
			Name:      name,
			Scope:     scope,
			RelPath:   relPath,
			Enabled:   on,
			Attrs:     map[string]string{"marketplace": marketplaceOf(name)},
		})
	}

	tg := model.Toggles{
		DisableBundledSkills:   f.DisableBundledSkills,
		SkillOverrides:         f.SkillOverrides,
		DisabledMcpjsonServers: f.DisabledMcpjsonServers,
		EnabledMcpjsonServers:  f.EnabledMcpjsonServers,
		EnableAllProjectMcp:    f.EnableAllProjectMcp,
		EnabledPlugins:         plugins,
		OutputStyle:            f.OutputStyle,
		ClaudeMdExcludes:       f.ClaudeMdExcludes,
	}

	// Walk the raw document for string values that look like secrets.
	// We re-decode into a generic map so we can traverse all keys recursively,
	// including ones not in settingsFile. The error is ignored: if Unmarshal
	// already succeeded above for settingsFile, it will succeed here too.
	var raw map[string]any
	_ = json.Unmarshal(content, &raw)
	perms := model.ScopePermissions{
		Scope:             scope,
		RelPath:           relPath,
		Deny:              f.Permissions.Deny,
		Allow:             f.Permissions.Allow,
		Ask:               f.Permissions.Ask,
		DefaultMode:       f.Permissions.DefaultMode,
		SecretSettingKeys: scanSecretKeys("", raw, sm),
	}
	return hooks, tg, perms, nil
}

// scanSecretKeys walks a decoded JSON value (the result of json.Unmarshal into
// any) and returns the dotted paths of every string leaf whose key and value
// look like a secret per sm. Only key paths are returned — the values are
// inspected then discarded (privacy default D8). Keys at each map level are
// visited in sorted order so the output is deterministic.
//
// Examples of returned paths: "anthropicApiKey", "env.API_TOKEN",
// "servers[0].password".
func scanSecretKeys(prefix string, v any, sm SecretMatcher) []string {
	var out []string
	switch t := v.(type) {
	case map[string]any:
		// Sort the keys so output order is stable across Go map iterations.
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			// Build the dotted path for this key.
			path := k
			if prefix != "" {
				path = prefix + "." + k
			}
			if s, ok := t[k].(string); ok {
				// Leaf string: test with SecretMatcher; keep path, discard value.
				if sm.Match(k, s) {
					out = append(out, path)
				}
				// Do not recurse — the value is a scalar, not a container.
				continue
			}
			// Non-string value: recurse so nested maps/arrays are also scanned.
			out = append(out, scanSecretKeys(path, t[k], sm)...)
		}
	case []any:
		// Recurse into each slice element, annotating the path with its index.
		for i, e := range t {
			out = append(out, scanSecretKeys(fmt.Sprintf("%s[%d]", prefix, i), e, sm)...)
		}
	}
	// Other types (number, bool, nil) hold no secrets; nothing appended.
	return out
}

// marketplaceOf extracts the marketplace suffix from a "name@marketplace" plugin
// key. Returns "" when there is no "@".
func marketplaceOf(key string) string {
	if i := strings.LastIndex(key, "@"); i >= 0 {
		return key[i+1:]
	}
	return ""
}

// MergeToggles combines per-scope toggles. The FIRST argument is the highest
// precedence. Booleans OR together (any scope that enables wins); map entries
// take the highest-precedence value per key; slices are unioned (deduplicated).
func MergeToggles(in ...model.Toggles) model.Toggles {
	out := model.Toggles{SkillOverrides: map[string]string{}}
	// seen maps track which slice values have already been added.
	seenDisabled, seenEnabled, seenPlugins := map[string]bool{}, map[string]bool{}, map[string]bool{}
	seenExcludes := map[string]bool{}

	// addUnique appends only values not yet present, using seen for O(1) lookup.
	addUnique := func(dst *[]string, seen map[string]bool, vals []string) {
		for _, v := range vals {
			if !seen[v] {
				seen[v] = true
				*dst = append(*dst, v)
			}
		}
	}

	// Apply lowest precedence first so higher scopes overwrite map keys.
	// We iterate in reverse so the first (highest-precedence) argument wins.
	for i := len(in) - 1; i >= 0; i-- {
		t := in[i]
		// Booleans: true from any scope sticks (OR semantics).
		if t.DisableBundledSkills {
			out.DisableBundledSkills = true
		}
		if t.EnableAllProjectMcp {
			out.EnableAllProjectMcp = true
		}
		// Map: write lowest first; higher scopes overwrite on next iteration.
		for k, v := range t.SkillOverrides {
			out.SkillOverrides[k] = v
		}
	}

	// Unions are precedence-independent; iterate forward for stable, input order.
	// A present allowlist anywhere yields a non-nil EnabledMcpjsonServers, which
	// is the signal resolveMCP uses to know an allowlist is active.
	for _, t := range in {
		addUnique(&out.DisabledMcpjsonServers, seenDisabled, t.DisabledMcpjsonServers)
		addUnique(&out.EnabledMcpjsonServers, seenEnabled, t.EnabledMcpjsonServers)
		addUnique(&out.EnabledPlugins, seenPlugins, t.EnabledPlugins)
		addUnique(&out.ClaudeMdExcludes, seenExcludes, t.ClaudeMdExcludes)
		// OutputStyle is a scalar: the first non-empty value wins. in[0] is the
		// highest-precedence scope, so keeping the first non-empty seen while
		// iterating forward yields highest-precedence-wins.
		if out.OutputStyle == "" && t.OutputStyle != "" {
			out.OutputStyle = t.OutputStyle
		}
	}
	return out
}
