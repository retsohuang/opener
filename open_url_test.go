package main

import (
	"runtime"
	"testing"
)

func TestOpenURL_DarwinLaunchesTheAbsoluteOpenBinary(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("the absolute-path fallback is Darwin-only")
	}

	var gotName string
	var gotArgs []string

	origRun := runOpenCommand
	runOpenCommand = func(name string, args ...string) (string, error) {
		gotName, gotArgs = name, args
		return "", nil
	}
	t.Cleanup(func() { runOpenCommand = origRun })

	if _, err := openURL("https://example.com/x"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotName != "/usr/bin/open" {
		t.Errorf("open command = %q, want the absolute /usr/bin/open", gotName)
	}
	if len(gotArgs) != 1 || gotArgs[0] != "https://example.com/x" {
		t.Errorf("open args = %v, want the request URL only", gotArgs)
	}
}
