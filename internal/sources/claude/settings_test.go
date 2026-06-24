package claude

import (
	"fmt"
	"strings"
	"testing"

	"github.com/saptarshi369/drishti/internal/model"
)

func TestParseSettings_HooksAndToggles(t *testing.T) {
	content := []byte(`{
	  "hooks": {
	    "PreToolUse": [
	      {"matcher": "Bash", "hooks": [{"type":"command","command":"guard-rm.sh"}]}
	    ]
	  },
	  "disableBundledSkills": true,
	  "disabledMcpjsonServers": ["postgres"],
	  "skillOverrides": {"pdf": "off"}
	}`)
	hooks, tg, _, err := ParseSettings(content, model.ScopeUser, "settings.json", SecretMatcher{})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(hooks) != 1 {
		t.Fatalf("got %d hooks, want 1", len(hooks))
	}
	h := hooks[0]
	if h.Category != model.CatHook || h.Name != "PreToolUse · Bash" {
		t.Errorf("hook name = %q (cat %s)", h.Name, h.Category)
	}
	if h.Attrs["command"] != "guard-rm.sh" {
		t.Errorf("hook command = %q", h.Attrs["command"])
	}
	if !tg.DisableBundledSkills || tg.SkillOverrides["pdf"] != "off" {
		t.Errorf("toggles = %+v", tg)
	}
	if len(tg.DisabledMcpjsonServers) != 1 || tg.DisabledMcpjsonServers[0] != "postgres" {
		t.Errorf("disabledMcp = %v", tg.DisabledMcpjsonServers)
	}
}

func TestMergeToggles_PrecedenceAndUnion(t *testing.T) {
	high := model.Toggles{DisableBundledSkills: true, SkillOverrides: map[string]string{"a": "off"}, DisabledMcpjsonServers: []string{"x"}}
	low := model.Toggles{SkillOverrides: map[string]string{"a": "on", "b": "name-only"}, DisabledMcpjsonServers: []string{"y"}}
	got := MergeToggles(high, low)
	if !got.DisableBundledSkills {
		t.Error("bool: true should stick")
	}
	if got.SkillOverrides["a"] != "off" || got.SkillOverrides["b"] != "name-only" {
		t.Errorf("map merge = %v", got.SkillOverrides)
	}
	if len(got.DisabledMcpjsonServers) != 2 {
		t.Errorf("union = %v, want 2 entries", got.DisabledMcpjsonServers)
	}
}

func TestParseSettings_PluginsAndNewToggles(t *testing.T) {
	js := []byte(`{
		"outputStyle": "Explanatory",
		"claudeMdExcludes": ["**/other-team/CLAUDE.md"],
		"enabledPlugins": {"github@official": true, "legacy@community": false}
	}`)
	items, tg, _, err := ParseSettings(js, model.ScopeUser, "settings.json", SecretMatcher{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var plugins []model.InventoryItem
	for _, it := range items {
		if it.Category == model.CatPlugin {
			plugins = append(plugins, it)
		}
	}
	if len(plugins) != 2 {
		t.Fatalf("want 2 plugin items, got %d", len(plugins))
	}
	if plugins[0].Name != "github@official" || !plugins[0].Enabled {
		t.Errorf("plugin[0] = %q enabled=%v", plugins[0].Name, plugins[0].Enabled)
	}
	if plugins[1].Name != "legacy@community" || plugins[1].Enabled {
		t.Errorf("plugin[1] = %q enabled=%v", plugins[1].Name, plugins[1].Enabled)
	}
	if plugins[0].Attrs["marketplace"] != "official" {
		t.Errorf("marketplace = %q", plugins[0].Attrs["marketplace"])
	}
	if tg.OutputStyle != "Explanatory" {
		t.Errorf("OutputStyle = %q", tg.OutputStyle)
	}
	if len(tg.ClaudeMdExcludes) != 1 || tg.ClaudeMdExcludes[0] != "**/other-team/CLAUDE.md" {
		t.Errorf("ClaudeMdExcludes = %v", tg.ClaudeMdExcludes)
	}
}

func TestMergeToggles_OutputStyleAndExcludes(t *testing.T) {
	// in[0] is highest precedence (local), then project, then user.
	local := model.Toggles{ClaudeMdExcludes: []string{"a"}}
	project := model.Toggles{OutputStyle: "Explanatory", ClaudeMdExcludes: []string{"b"}}
	user := model.Toggles{OutputStyle: "Learning", ClaudeMdExcludes: []string{"a", "c"}}
	got := MergeToggles(local, project, user)
	// Highest-precedence non-empty OutputStyle wins: project beats user.
	if got.OutputStyle != "Explanatory" {
		t.Errorf("OutputStyle = %q, want Explanatory", got.OutputStyle)
	}
	// ClaudeMdExcludes are unioned (deduped), input order.
	want := []string{"a", "b", "c"}
	if len(got.ClaudeMdExcludes) != len(want) {
		t.Fatalf("ClaudeMdExcludes = %v, want %v", got.ClaudeMdExcludes, want)
	}
	for i := range want {
		if got.ClaudeMdExcludes[i] != want[i] {
			t.Errorf("ClaudeMdExcludes[%d] = %q, want %q", i, got.ClaudeMdExcludes[i], want[i])
		}
	}
}

func TestParseSettings_PermissionsAndSecretKeys(t *testing.T) {
	doc := []byte(`{
		"permissions": {"deny": ["Read(.env)"], "allow": ["Bash(*)"], "ask": [], "defaultMode": "bypassPermissions"},
		"anthropicApiKey": "sk-FAKEFAKEFAKE000"
	}`)
	_, _, perms, err := ParseSettings(doc, model.ScopeUser, "settings.json", SecretMatcher{})
	if err != nil {
		t.Fatal(err)
	}
	if perms.DefaultMode != "bypassPermissions" {
		t.Fatalf("DefaultMode = %q", perms.DefaultMode)
	}
	if len(perms.Deny) != 1 || perms.Deny[0] != "Read(.env)" {
		t.Fatalf("Deny = %v", perms.Deny)
	}
	if len(perms.Allow) != 1 || perms.Allow[0] != "Bash(*)" {
		t.Fatalf("Allow = %v", perms.Allow)
	}
	if len(perms.SecretSettingKeys) != 1 || perms.SecretSettingKeys[0] != "anthropicApiKey" {
		t.Fatalf("SecretSettingKeys = %v, want [anthropicApiKey]", perms.SecretSettingKeys)
	}
	// Privacy: the value must not leak into ScopePermissions output.
	if strings.Contains(fmt.Sprintf("%+v", perms), "sk-FAKEFAKEFAKE000") {
		t.Fatal("secret value leaked into ParseSettings output")
	}
}
