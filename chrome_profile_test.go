package main

import (
	"os"
	"path/filepath"
	"testing"
)

const fixtureLocalState = `{
  "profile": {
    "info_cache": {
      "Default": {"user_name": ""},
      "Profile 1": {"user_name": "retso.huang@ikala.ai"},
      "Profile 3": {"user_name": "retsohuang@gmail.com"}
    }
  }
}`

func withFixtureLocalState(t *testing.T, content string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "Local State")
	if content != "" {
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	orig := chromeLocalStatePath
	chromeLocalStatePath = func() (string, error) { return path, nil }
	t.Cleanup(func() { chromeLocalStatePath = orig })
}

func TestResolveChromeProfileDir_KnownEmailMatches(t *testing.T) {
	withFixtureLocalState(t, fixtureLocalState)
	dir, ok := resolveChromeProfileDir("retso.huang@ikala.ai")
	if !ok || dir != "Profile 1" {
		t.Errorf("got (%q, %v), want (\"Profile 1\", true)", dir, ok)
	}
}

func TestResolveChromeProfileDir_NonMatchingEmailReturnsNoKey(t *testing.T) {
	withFixtureLocalState(t, fixtureLocalState)
	dir, ok := resolveChromeProfileDir("nobody@example.com")
	if ok || dir != "" {
		t.Errorf("got (%q, %v), want (\"\", false) — a non-matching email must never guess", dir, ok)
	}
}

func TestResolveChromeProfileDir_EmptyEmailReturnsNoKey(t *testing.T) {
	withFixtureLocalState(t, fixtureLocalState)
	dir, ok := resolveChromeProfileDir("")
	if ok || dir != "" {
		t.Errorf("got (%q, %v), want (\"\", false)", dir, ok)
	}
}

func TestResolveChromeProfileDir_UnreadableLocalStateReturnsNoKey(t *testing.T) {
	withFixtureLocalState(t, "") // no file written at the fixture path
	dir, ok := resolveChromeProfileDir("retso.huang@ikala.ai")
	if ok || dir != "" {
		t.Errorf("got (%q, %v), want (\"\", false)", dir, ok)
	}
}
