package claude

import "testing"

func TestSecretMatcher_Match(t *testing.T) {
	m := SecretMatcher{} // empty → built-in defaults
	cases := []struct {
		name, key, value string
		want             bool
	}{
		{"value prefix sk-", "FOO", "sk-FAKEFAKEFAKE000", true},
		{"value prefix ghp_", "anything", "ghp_0000000000", true},
		{"secret key + opaque value", "API_KEY", "ABCDEFGHIJKLMNOP12345", true},
		{"secret key but pathy value (apiKeyHelper)", "apiKeyHelper", "/usr/local/bin/get_key.sh", false},
		{"secret key but short value", "token", "abc", false},
		{"plain pair", "command", "echo hello world", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := m.Match(c.key, c.value); got != c.want {
				t.Fatalf("Match(%q,%q)=%v want %v", c.key, c.value, got, c.want)
			}
		})
	}
}

func TestLooksOpaque(t *testing.T) {
	if looksOpaque("/path/with/slashes/and/length") {
		t.Fatal("path with slashes should not be opaque")
	}
	if !looksOpaque("ABCDEFGHIJKLMNOP12345") {
		t.Fatal("long no-space token should be opaque")
	}
	if looksOpaque("short") {
		t.Fatal("short string should not be opaque")
	}
}
