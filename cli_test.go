package main

import "testing"

func TestParseArgs(t *testing.T) {
	cases := []struct {
		name   string
		args   []string
		want   cliOptions
		errish bool
	}{
		{"empty", nil, cliOptions{}, false},
		{"id only", []string{"finder-path-bar"}, cliOptions{stepID: "finder-path-bar", action: actionRun}, false},
		{"id then done", []string{"finder-path-bar", "--done"}, cliOptions{stepID: "finder-path-bar", action: actionDone}, false},
		{"flag then id", []string{"--copy", "finder-path-bar"}, cliOptions{stepID: "finder-path-bar", action: actionCopy}, false},
		{"undone", []string{"x", "--undone"}, cliOptions{stepID: "x", action: actionUndone}, false},
		{"dry-run shorthand", []string{"x", "-n"}, cliOptions{stepID: "x", action: actionRun, dryRun: true}, false},
		{"help", []string{"--help"}, cliOptions{help: true}, false},
		{"conflicting actions", []string{"x", "--done", "--copy"}, cliOptions{}, true},
		{"unknown flag", []string{"x", "--nope"}, cliOptions{}, true},
		{"two ids", []string{"a", "b"}, cliOptions{}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseArgs(tc.args)
			if tc.errish {
				if err == nil {
					t.Fatalf("expected an error for %v", tc.args)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("parseArgs(%v) = %+v, want %+v", tc.args, got, tc.want)
			}
		})
	}
}

func TestClipboardPayload(t *testing.T) {
	cmd := Step{Commands: []string{"a", "b"}}
	if got := clipboardPayload(cmd); got != "a\nb" {
		t.Fatalf("command payload = %q, want %q", got, "a\nb")
	}
	man := Step{ManualInstructions: "do the thing"}
	if got := clipboardPayload(man); got != "do the thing" {
		t.Fatalf("manual payload = %q, want %q", got, "do the thing")
	}
}

func TestStepByID(t *testing.T) {
	if _, ok := StepByID("finder-path-bar"); !ok {
		t.Fatal("expected finder-path-bar to exist")
	}
	if _, ok := StepByID("does-not-exist"); ok {
		t.Fatal("expected lookup miss for unknown id")
	}
}

func TestRunDirectDoneUndone(t *testing.T) {
	t.Setenv("HOME", t.TempDir()) // isolate state.json under a temp home
	const id = "finder-path-bar"

	if code := runDirect(cliOptions{stepID: id, action: actionDone}); code != 0 {
		t.Fatalf("--done returned %d", code)
	}
	if s := LoadState(); s.Steps[id] != StatusCompleted {
		t.Fatalf("expected %s completed, got %q", id, s.Steps[id])
	}

	if code := runDirect(cliOptions{stepID: id, action: actionUndone}); code != 0 {
		t.Fatalf("--undone returned %d", code)
	}
	if s := LoadState(); s.Steps[id] != "" {
		t.Fatalf("expected %s cleared, got %q", id, s.Steps[id])
	}
}

func TestRunDirectUnknownID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if code := runDirect(cliOptions{stepID: "nope", action: actionDone}); code != 1 {
		t.Fatalf("expected exit 1 for unknown id, got %d", code)
	}
}

func TestRunDirectManualStepIsNotMarked(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	// accessibility-zoom is a manual (instruction-only) step.
	const id = "accessibility-zoom"
	if code := runDirect(cliOptions{stepID: id, action: actionRun}); code != 0 {
		t.Fatalf("manual run returned %d", code)
	}
	if s := LoadState(); s.Steps[id] == StatusCompleted {
		t.Fatal("a manual step should not be auto-marked done by a direct run")
	}
}
