package main

import (
	"bytes"
	"strings"
	"testing"
)

// These exercise the --dry-run and missing-consent paths only: both return
// before collect.Run is ever called, so they don't depend on a live
// Windows registry or PowerShell — unlike an actual collection run, which
// this package deliberately leaves to manual/host verification rather than
// faking out collect.All()'s real adapters in a unit test.

func TestRun_DryRunPrintsManifestAndExitsZero(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"--dry-run"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "os.version") {
		t.Errorf("dry-run output missing a known category id; got:\n%s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "never reads") {
		t.Errorf("dry-run output missing the never-collects disclosure; got:\n%s", stdout.String())
	}
}

func TestRun_WithoutConsentRefusesAndExitsNonZero(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{}, &stdout, &stderr)

	if code == 0 {
		t.Fatal("running with neither --dry-run nor --consent must not exit 0")
	}
	if !strings.Contains(stderr.String(), "refusing to run without --consent") {
		t.Errorf("missing the consent-refusal message; stderr:\n%s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "os.version") {
		t.Errorf("consent refusal must still show the manifest; stderr:\n%s", stderr.String())
	}
}

func TestRun_DryRunTakesPrecedenceOverMissingConsent(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"--dry-run"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (dry-run must not require --consent)", code)
	}
}
