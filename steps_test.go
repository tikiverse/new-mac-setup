package main

import "testing"

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

func TestNewModelHidesTestingCategory(t *testing.T) {
	m := newModel(&AppState{Steps: make(map[string]StepStatus)})
	for _, c := range m.categories {
		if c == "Testing" {
			t.Fatal("newModel should not expose the Testing category by default")
		}
	}
}
