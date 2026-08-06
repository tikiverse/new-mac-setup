package main

import (
	"strings"
	"testing"
)

func TestTestingCategoryIsDebugOnly(t *testing.T) {
	// Test steps still exist in AllSteps so the CLI can find them by id,
	// and they are marked Debug.
	var found bool
	for _, s := range AllSteps() {
		if s.ID == "test-fail" {
			found = true
			if !s.Debug {
				t.Fatal("test-fail should be marked Debug")
			}
		}
	}
	if !found {
		t.Fatal("test-fail should exist in AllSteps")
	}

	// Hidden from the default (non-debug) category list.
	for _, c := range visibleCategories(false) {
		if c == "Testing" {
			t.Fatal("Testing category should be hidden without --debug")
		}
	}

	// Visible when debug is requested.
	var shown bool
	for _, c := range visibleCategories(true) {
		if c == "Testing" {
			shown = true
		}
	}
	if !shown {
		t.Fatal("Testing category should appear with --debug")
	}
}

// Step ids are the CLI's handle on a step and the key progress is stored
// under, so a duplicate would silently shadow another step's state.
func TestStepIDsAreUnique(t *testing.T) {
	seen := map[string]string{}
	for _, s := range AllSteps() {
		if prev, dup := seen[s.ID]; dup {
			t.Fatalf("duplicate step id %q, used by %q and %q", s.ID, prev, s.Name)
		}
		seen[s.ID] = s.Name
	}
}

// Every step needs an id, a category and something to do — a step with neither
// commands nor instructions would render as an empty run and mark itself done.
func TestStepsAreWellFormed(t *testing.T) {
	for _, s := range AllSteps() {
		switch {
		case s.ID == "":
			t.Errorf("step %q has no id", s.Name)
		case s.Category == "":
			t.Errorf("step %q has no category", s.ID)
		case s.Name == "":
			t.Errorf("step %q has no name", s.ID)
		case len(s.Commands) == 0 && s.ManualInstructions == "":
			t.Errorf("step %q has neither commands nor manual instructions", s.ID)
		}
	}
}

// The three cooldown steps express one decision — a 9-day window — in three
// different units, so changing one and forgetting the others is the easy
// mistake. Each package manager's spelling of "9 days" is pinned here.
func TestCooldownStepsShareOneWindow(t *testing.T) {
	want := map[string]string{
		"npm-cooldown":  "min-release-age 9",       // npm counts in days
		"pnpm-cooldown": "minimumReleaseAge 12960", // pnpm counts in minutes
		"uv-cooldown":   `exclude-newer = "9 days"`,
	}

	for id, substr := range want {
		step, ok := StepByID(id)
		if !ok {
			t.Errorf("step %q is missing", id)
			continue
		}
		if !strings.Contains(strings.Join(step.Commands, "\n"), substr) {
			t.Errorf("step %q should set a 9-day window (looking for %q) — if the window changed, change all three", id, substr)
		}
		// They are meant to be runnable as one category, without dragging the
		// Development installs along.
		if step.Category != "Supply Chain" {
			t.Errorf("step %q should be in Supply Chain, got %q", id, step.Category)
		}
	}
}

func TestNewModelHidesTestingCategory(t *testing.T) {
	m := newModel(&AppState{Steps: make(map[string]StepStatus)})
	for _, c := range m.categories {
		if c == "Testing" {
			t.Fatal("newModel should not expose the Testing category by default")
		}
	}
}
