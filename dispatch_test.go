package main

import "testing"

func TestOpenRequestURL_NoEmailFallsBackToOpenURL(t *testing.T) {
	var openURLCalledWith string
	withFixtureLocalState(t, fixtureLocalState)
	origOpenURL := openURL
	openURL = func(line string) (string, error) { openURLCalledWith = line; return "", nil }
	t.Cleanup(func() { openURL = origOpenURL })

	_, err := openRequestURL(openRequest{URL: "https://example.com/x"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if openURLCalledWith != "https://example.com/x" {
		t.Errorf("openURL called with %q, want the request URL", openURLCalledWith)
	}
}

func TestOpenRequestURL_NonResolvingEmailFallsBackToOpenURL(t *testing.T) {
	var openURLCalledWith string
	withFixtureLocalState(t, fixtureLocalState)
	origOpenURL := openURL
	openURL = func(line string) (string, error) { openURLCalledWith = line; return "", nil }
	t.Cleanup(func() { openURL = origOpenURL })

	origFindChrome := findChromeBinary
	findChromeBinary = func() string { t.Fatal("must not search for Chrome when the email never resolves"); return "" }
	t.Cleanup(func() { findChromeBinary = origFindChrome })

	_, err := openRequestURL(openRequest{URL: "https://example.com/x", ChromeProfileEmail: "nobody@example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if openURLCalledWith != "https://example.com/x" {
		t.Errorf("openURL called with %q, want the request URL", openURLCalledWith)
	}
}

func TestOpenRequestURL_NoChromeBinaryFallsBackToOpenURL(t *testing.T) {
	var openURLCalledWith string
	withFixtureLocalState(t, fixtureLocalState)
	origOpenURL := openURL
	openURL = func(line string) (string, error) { openURLCalledWith = line; return "", nil }
	t.Cleanup(func() { openURL = origOpenURL })

	origFindChrome := findChromeBinary
	findChromeBinary = func() string { return "" }
	t.Cleanup(func() { findChromeBinary = origFindChrome })

	origLaunch := launchChromeAtProfile
	launchChromeAtProfile = func(bin, dir, url string) error {
		t.Fatal("must not launch when no Chrome binary was found")
		return nil
	}
	t.Cleanup(func() { launchChromeAtProfile = origLaunch })

	_, err := openRequestURL(openRequest{URL: "https://example.com/x", ChromeProfileEmail: "retso.huang@ikala.ai"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if openURLCalledWith != "https://example.com/x" {
		t.Errorf("openURL called with %q, want the request URL", openURLCalledWith)
	}
}

func TestOpenRequestURL_UnreadableLocalStateFallsBackToOpenURL(t *testing.T) {
	var openURLCalledWith string
	withFixtureLocalState(t, "") // no Local State file at the fixture path
	origOpenURL := openURL
	openURL = func(line string) (string, error) { openURLCalledWith = line; return "", nil }
	t.Cleanup(func() { openURL = origOpenURL })

	_, err := openRequestURL(openRequest{URL: "https://example.com/x", ChromeProfileEmail: "retso.huang@ikala.ai"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if openURLCalledWith != "https://example.com/x" {
		t.Errorf("openURL called with %q, want the request URL", openURLCalledWith)
	}
}

func TestOpenRequestURL_ResolvingEmailLaunchesChromeDirectlyWithNoOpenURLFallback(t *testing.T) {
	withFixtureLocalState(t, fixtureLocalState)

	origFindChrome := findChromeBinary
	findChromeBinary = func() string { return "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" }
	t.Cleanup(func() { findChromeBinary = origFindChrome })

	var gotBin, gotDir, gotURL string
	origLaunch := launchChromeAtProfile
	launchChromeAtProfile = func(bin, dir, url string) error {
		gotBin, gotDir, gotURL = bin, dir, url
		return nil
	}
	t.Cleanup(func() { launchChromeAtProfile = origLaunch })

	openURLCalled := false
	origOpenURL := openURL
	openURL = func(line string) (string, error) { openURLCalled = true; return "", nil }
	t.Cleanup(func() { openURL = origOpenURL })

	_, err := openRequestURL(openRequest{URL: "https://example.com/x", ChromeProfileEmail: "retso.huang@ikala.ai"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if openURLCalled {
		t.Error("openURL fallback must not fire when the direct Chrome launch succeeds")
	}
	if gotBin == "" {
		t.Error("expected the resolved Chrome binary to be passed to the launcher")
	}
	if gotDir != "Profile 1" {
		t.Errorf("--profile-directory value = %q, want the Local State-resolved \"Profile 1\"", gotDir)
	}
	if gotURL != "https://example.com/x" {
		t.Errorf("url = %q, want the request URL", gotURL)
	}
}
