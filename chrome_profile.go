package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// chromeLocalState is the subset of Chrome's `Local State` file this package
// reads: the profile directory keys and the account email each is signed
// into.
type chromeLocalState struct {
	Profile struct {
		InfoCache map[string]struct {
			UserName string `json:"user_name"`
		} `json:"info_cache"`
	} `json:"profile"`
}

// chromeLocalStatePath returns this machine's Chrome `Local State` path.
// Overridable so tests never touch a real Chrome installation.
var chromeLocalStatePath = func() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "Application Support", "Google", "Chrome", "Local State"), nil
}

// resolveChromeProfileDir matches email against each profile.info_cache
// entry's user_name in Chrome's Local State, returning that entry's
// directory key. Returns ("", false) — never an error — on an empty email,
// an unreadable or unparseable Local State, or no matching entry: the
// caller's contract is to fall back rather than fail the open.
func resolveChromeProfileDir(email string) (string, bool) {
	if email == "" {
		return "", false
	}
	path, err := chromeLocalStatePath()
	if err != nil {
		return "", false
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	var state chromeLocalState
	if err := json.Unmarshal(b, &state); err != nil {
		return "", false
	}
	for dir, info := range state.Profile.InfoCache {
		if info.UserName == email {
			return dir, true
		}
	}
	return "", false
}

// findChromeBinary locates the Google Chrome executable: LaunchServices
// resolution (via Spotlight metadata, which is backed by the LaunchServices
// registration) for the com.google.Chrome bundle identifier first, falling
// back to the two conventional install locations. Returns "" when none is
// found — the caller's contract is to fall back to the default browser
// rather than fail.
var findChromeBinary = func() string {
	if mdfind, err := exec.LookPath("mdfind"); err == nil {
		out, err := exec.Command(mdfind, "kMDItemCFBundleIdentifier == 'com.google.Chrome'").Output()
		if err == nil {
			for _, bundle := range strings.Split(strings.TrimSpace(string(out)), "\n") {
				if bundle == "" {
					continue
				}
				if candidate := filepath.Join(bundle, "Contents", "MacOS", "Google Chrome"); isExecutableFile(candidate) {
					return candidate
				}
			}
		}
	}
	for _, dir := range []string{"/Applications", filepath.Join(os.Getenv("HOME"), "Applications")} {
		if candidate := filepath.Join(dir, "Google Chrome.app", "Contents", "MacOS", "Google Chrome"); isExecutableFile(candidate) {
			return candidate
		}
	}
	return ""
}

func isExecutableFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// launchChromeAtProfile executes the Chrome binary directly — never through
// `open --args`, which silently no-ops against an already-running Chrome —
// with --profile-directory set to profileDir. profileDir must already be a
// directory key resolved from Local State; no caller-supplied text is ever
// forwarded to Chrome's command line. Non-blocking: Chrome may already be
// running (in which case this process forwards the request and exits fast)
// or may not be (in which case it becomes a long-lived process this daemon
// must not wait on).
var launchChromeAtProfile = func(chromeBinary, profileDir, url string) error {
	cmd := exec.Command(chromeBinary, "--profile-directory="+profileDir, url)
	return cmd.Start()
}
