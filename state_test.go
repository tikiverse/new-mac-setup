package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadStateMissingFileIsNotAnError(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	s, err := LoadState()
	if err != nil {
		t.Fatalf("a first run should not report an error, got %v", err)
	}
	if s.Steps == nil {
		t.Fatal("expected an initialized Steps map")
	}
}

func TestLoadStateRoundTrip(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	seed := &AppState{
		Steps:         map[string]StepStatus{"finder-path-bar": StatusCompleted, "key-repeat": StatusSkipped},
		SelectedSteps: map[string]bool{"finder-path-bar": true},
	}
	if err := seed.Save(); err != nil {
		t.Fatal(err)
	}
	got, err := LoadState()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Steps["finder-path-bar"] != StatusCompleted || got.Steps["key-repeat"] != StatusSkipped {
		t.Fatalf("statuses did not survive a round trip: %v", got.Steps)
	}
}

// A corrupt file must be reported and preserved — never silently discarded,
// which is what let a bad read turn into permanent data loss.
func TestLoadStateCorruptFileIsReportedAndKept(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.MkdirAll(stateDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	// A truncated write, as an interrupted save would leave behind.
	if err := os.WriteFile(statePath(), []byte(`{"steps":{"finder-path-`), 0o644); err != nil {
		t.Fatal(err)
	}

	s, err := LoadState()
	if err == nil {
		t.Fatal("expected an error for an unparseable state file")
	}
	if s == nil || s.Steps == nil {
		t.Fatal("expected a usable fresh state alongside the error")
	}

	matches, _ := filepath.Glob(statePath() + ".corrupt-*")
	if len(matches) != 1 {
		t.Fatalf("expected the damaged file to be kept, found %v", matches)
	}
	kept, readErr := os.ReadFile(matches[0])
	if readErr != nil {
		t.Fatal(readErr)
	}
	if !strings.Contains(string(kept), "finder-path-") {
		t.Fatalf("quarantined file lost its contents: %q", kept)
	}
	if !strings.Contains(err.Error(), matches[0]) {
		t.Errorf("error should name the kept file, got %q", err)
	}

	// The salvaged copy must survive the next save.
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(matches[0]); err != nil {
		t.Fatalf("a save destroyed the quarantined file: %v", err)
	}
}

// Save must not leave a partially written state file behind.
func TestSaveIsAtomic(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	prev := &AppState{Steps: map[string]StepStatus{"finder-path-bar": StatusCompleted}}
	if err := prev.Save(); err != nil {
		t.Fatal(err)
	}

	next := &AppState{Steps: map[string]StepStatus{"key-repeat": StatusFailed}}
	if err := next.Save(); err != nil {
		t.Fatal(err)
	}

	// No temp files may survive a successful save.
	leftovers, _ := filepath.Glob(filepath.Join(stateDir(), "state-*.tmp"))
	if len(leftovers) != 0 {
		t.Fatalf("save left temp files behind: %v", leftovers)
	}

	// Whatever is at the real path must always be complete and parseable.
	got, err := LoadState()
	if err != nil {
		t.Fatalf("state file was not parseable after save: %v", err)
	}
	if got.Steps["key-repeat"] != StatusFailed {
		t.Fatalf("unexpected contents: %v", got.Steps)
	}

	info, err := os.Stat(statePath())
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o644 {
		t.Errorf("state file mode = %v, want 0644", info.Mode().Perm())
	}
}
