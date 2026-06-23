package main

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// sendKey simulates a key press and returns the updated model.
func sendKey(m tea.Model, key string) tea.Model {
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
	return updated
}

func sendSpecialKey(m tea.Model, keyType tea.KeyType) tea.Model {
	updated, _ := m.Update(tea.KeyMsg{Type: keyType})
	return updated
}

func TestGRunsOnlyCurrentCategory(t *testing.T) {
	state := &AppState{Steps: make(map[string]StepStatus)}
	m := newModel(state)
	m.dryRun = true

	var tm tea.Model = m

	if m.screen != screenCategories {
		t.Fatalf("expected categories screen, got %d", m.screen)
	}

	// Cursor is on first category. Press G to run it.
	firstCat := m.categories[0]
	var firstCatSteps []Step
	for _, s := range AllSteps() {
		if s.Category == firstCat {
			firstCatSteps = append(firstCatSteps, s)
		}
	}

	tm = sendKey(tm, "G")
	m = tm.(model)

	if m.screen != screenCategoryRun {
		t.Fatalf("expected category run screen, got %d", m.screen)
	}

	if len(m.runSteps) != len(firstCatSteps) {
		t.Fatalf("expected %d steps in run, got %d", len(firstCatSteps), len(m.runSteps))
	}
	for _, s := range m.runSteps {
		if s.Category != firstCat {
			t.Fatalf("step %q has category %q, expected %q", s.ID, s.Category, firstCat)
		}
	}

	t.Logf("PASS: G runs only %d steps from %q", len(m.runSteps), firstCat)
}

func TestRestartPreservesSelection(t *testing.T) {
	state := &AppState{
		Steps:         make(map[string]StepStatus),
		SelectedSteps: make(map[string]bool),
	}

	cats := Categories()
	firstCat := cats[0]
	allSteps := AllSteps()

	for _, s := range allSteps {
		if s.Category == firstCat {
			state.SelectedSteps[s.ID] = true
		}
	}
	state.Steps[allSteps[0].ID] = StatusCompleted

	m := newModel(state)
	m.dryRun = true

	// Should go straight to categories
	if m.screen != screenCategories {
		t.Fatalf("expected categories screen, got %d", m.screen)
	}

	// Step selection should be restored from state
	selectedCount := 0
	for _, s := range allSteps {
		if m.stepSelected[s.ID] {
			selectedCount++
			if s.Category != firstCat {
				t.Fatalf("step %q (category %q) should not be selected", s.ID, s.Category)
			}
		}
	}

	var firstCatCount int
	for _, s := range allSteps {
		if s.Category == firstCat {
			firstCatCount++
		}
	}

	if selectedCount != firstCatCount {
		t.Fatalf("expected %d selected steps, got %d", firstCatCount, selectedCount)
	}

	t.Logf("PASS: Restart preserved selection (%d steps from %q)", selectedCount, firstCat)
}

func TestCategoryRunSequential(t *testing.T) {
	state := &AppState{Steps: make(map[string]StepStatus)}
	m := newModel(state)
	m.dryRun = true

	var tm tea.Model = m

	// G to run first category
	tm = sendKey(tm, "G")
	m = tm.(model)

	firstCat := m.categories[0]
	var firstCatSteps []Step
	for _, s := range AllSteps() {
		if s.Category == firstCat {
			firstCatSteps = append(firstCatSteps, s)
		}
	}

	if m.screen != screenCategoryRun {
		t.Fatalf("expected category run screen, got %d", m.screen)
	}

	// Simulate step results flowing through
	for i, step := range firstCatSteps {
		m = tm.(model)
		if m.screen != screenCategoryRun {
			t.Fatalf("step %d: expected category run screen, got %d", i, m.screen)
		}

		if step.ManualInstructions != "" && len(step.Commands) == 0 {
			tm, _ = tm.Update(manualStepMsg{step: step})
			m = tm.(model)
			if !m.runWaitManual {
				t.Fatalf("step %d: expected runWaitManual", i)
			}
			tm = sendSpecialKey(tm, tea.KeyEnter)
			m = tm.(model)
		} else {
			tm, _ = tm.Update(stepResultMsg{step: step, result: RunResult{Output: "(dry)"}})
			m = tm.(model)
		}
	}

	// After all steps, should get categoryDoneMsg
	tm, _ = tm.Update(categoryDoneMsg{})
	m = tm.(model)

	if !m.runDone {
		t.Fatal("expected runDone to be true")
	}

	for _, step := range firstCatSteps {
		status := m.state.Steps[step.ID]
		if status != StatusCompleted {
			t.Fatalf("step %q should be completed, got %q", step.ID, status)
		}
	}

	// Press Enter to return to categories
	tm = sendSpecialKey(tm, tea.KeyEnter)
	m = tm.(model)

	if m.screen != screenCategories {
		t.Fatalf("expected categories screen after run, got %d", m.screen)
	}

	if !m.isCategoryDone(firstCat) {
		t.Fatal("expected first category to be done")
	}

	t.Logf("PASS: Ran %d steps sequentially, category marked done", len(firstCatSteps))
}

func TestCategoryRunSkipsAlreadyDone(t *testing.T) {
	state := &AppState{Steps: make(map[string]StepStatus)}
	m := newModel(state)
	m.dryRun = true

	firstCat := m.categories[0]
	var firstCatSteps []Step
	for _, s := range AllSteps() {
		if s.Category == firstCat {
			firstCatSteps = append(firstCatSteps, s)
		}
	}

	// Mark first 3 as completed
	for i, s := range firstCatSteps {
		if i < 3 {
			m.state.Steps[s.ID] = StatusCompleted
		}
	}

	var tm tea.Model = m

	// Press G
	tm = sendKey(tm, "G")
	m = tm.(model)

	if m.screen != screenCategoryRun {
		t.Fatalf("expected category run screen, got %d", m.screen)
	}

	expected := len(firstCatSteps) - 3
	if len(m.runSteps) != expected {
		t.Fatalf("expected %d steps to run (skipping 3 done), got %d", expected, len(m.runSteps))
	}

	t.Logf("PASS: Category run skipped 3 already-done steps, running %d", expected)
}

func TestGOnFullyDoneCategoryIsNoop(t *testing.T) {
	state := &AppState{Steps: make(map[string]StepStatus)}
	m := newModel(state)
	m.dryRun = true

	firstCat := m.categories[0]
	for _, s := range AllSteps() {
		if s.Category == firstCat {
			m.state.Steps[s.ID] = StatusCompleted
		}
	}

	var tm tea.Model = m

	tm = sendKey(tm, "G")
	m = tm.(model)

	if m.screen != screenCategories {
		t.Fatalf("expected to stay on categories (nothing to run), got %d", m.screen)
	}

	t.Logf("PASS: G on fully completed category is a no-op")
}

func TestStepSelectToggle(t *testing.T) {
	state := &AppState{Steps: make(map[string]StepStatus)}
	m := newModel(state)
	m.dryRun = true

	var tm tea.Model = m

	// Enter first category
	tm = sendSpecialKey(tm, tea.KeyEnter)
	m = tm.(model)

	if m.screen != screenStepSelect {
		t.Fatalf("expected step select screen, got %d", m.screen)
	}

	// All steps should be selected initially
	for _, s := range m.stepSelectSteps {
		if !m.stepSelected[s.ID] {
			t.Fatalf("expected step %q to be selected initially", s.ID)
		}
	}

	// Move to first step (row 1) and toggle it off
	tm = sendKey(tm, "j")
	tm = sendKey(tm, " ")
	m = tm.(model)

	firstStep := m.stepSelectSteps[0]
	if m.stepSelected[firstStep.ID] {
		t.Fatalf("expected step %q to be deselected after toggle", firstStep.ID)
	}

	// Go back to categories
	tm = sendSpecialKey(tm, tea.KeyEnter)
	m = tm.(model)

	if m.screen != screenCategories {
		t.Fatalf("expected categories screen, got %d", m.screen)
	}

	t.Logf("PASS: Step select toggle works")
}

// startFailedRun runs the first category, sends a failing result for its first
// step, and returns the paused model plus that step.
func startFailedRun(t *testing.T) (tea.Model, Step) {
	t.Helper()
	state := &AppState{Steps: make(map[string]StepStatus)}
	m := newModel(state)
	m.dryRun = true

	var tm tea.Model = m
	tm = sendKey(tm, "G")
	m = tm.(model)
	if m.screen != screenCategoryRun {
		t.Fatalf("expected category run screen, got %d", m.screen)
	}

	step := m.runSteps[0]
	failResult := RunResult{Output: "boom: command failed", Err: errors.New("exit 1")}
	tm, _ = tm.Update(stepResultMsg{step: step, result: failResult})
	m = tm.(model)

	if !m.runWaitFail {
		t.Fatal("expected runWaitFail after a failing step")
	}
	if m.runFailStep == nil || m.runFailStep.ID != step.ID {
		t.Fatal("expected runFailStep to be the failed step")
	}
	if m.state.Steps[step.ID] == StatusFailed {
		t.Fatal("step should not be recorded failed until the user resolves the pause")
	}
	return tm, step
}

func TestFailurePauseRetry(t *testing.T) {
	tm, step := startFailedRun(t)
	idxBefore := tm.(model).runIndex

	// Retry re-runs the same step: clears the pause, runIndex unchanged.
	tm = sendKey(tm, "r")
	m := tm.(model)

	if m.runWaitFail {
		t.Fatal("expected runWaitFail cleared after retry")
	}
	if m.runIndex != idxBefore {
		t.Fatalf("retry should re-run the same step (idx %d), got %d", idxBefore, m.runIndex)
	}
	if status := m.state.Steps[step.ID]; status == StatusFailed {
		t.Fatal("retry should not record the step as failed")
	}
	t.Logf("PASS: retry re-runs the failed step")
}

func TestFailurePauseSkip(t *testing.T) {
	tm, step := startFailedRun(t)

	tm = sendKey(tm, "s")
	m := tm.(model)

	if m.runWaitFail {
		t.Fatal("expected runWaitFail cleared after skip")
	}
	if m.state.Steps[step.ID] != StatusFailed {
		t.Fatalf("skip should record the step failed, got %q", m.state.Steps[step.ID])
	}
	if m.runIndex == 0 {
		t.Fatal("skip should advance past the failed step")
	}
	var logged bool
	for _, e := range m.runLog {
		if e.name == step.Name && e.status == "fail" {
			logged = true
		}
	}
	if !logged {
		t.Fatal("skip should add a fail entry to the run log")
	}
	t.Logf("PASS: skip records failure and advances")
}

func TestFailurePauseAbort(t *testing.T) {
	tm, step := startFailedRun(t)

	tm = sendKey(tm, "a")
	m := tm.(model)

	if m.runWaitFail {
		t.Fatal("expected runWaitFail cleared after abort")
	}
	if !m.runDone {
		t.Fatal("abort should end the run")
	}
	if m.state.Steps[step.ID] != StatusFailed {
		t.Fatalf("abort should record the step failed, got %q", m.state.Steps[step.ID])
	}
	t.Logf("PASS: abort records failure and stops the run")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
