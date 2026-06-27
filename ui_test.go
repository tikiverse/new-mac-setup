package main

import (
	"errors"
	"fmt"
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
			tm, _ = tm.Update(stepFinishedMsg{step: step})
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
	tm, _ = tm.Update(stepFinishedMsg{step: step, err: errors.New("exit 1"), output: "boom: command failed"})
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
	if m.state.Steps[step.ID] != StatusSkipped {
		t.Fatalf("skip should record the step skipped, got %q", m.state.Steps[step.ID])
	}
	if m.runIndex == 0 {
		t.Fatal("skip should advance past the failed step")
	}
	var logged bool
	for _, e := range m.runLog {
		if e.name == step.Name && e.status == "skip" {
			logged = true
		}
	}
	if !logged {
		t.Fatal("skip should add a skip entry to the run log")
	}
	t.Logf("PASS: skip records skipped status and advances")
}

func TestFailurePauseTracksRetryCount(t *testing.T) {
	tm, step := startFailedRun(t)
	if got := tm.(model).runFailCounts[step.ID]; got != 1 {
		t.Fatalf("expected 1 failure recorded, got %d", got)
	}

	// Retry, then fail again: the in-run counter should climb.
	tm = sendKey(tm, "r")
	tm, _ = tm.Update(stepFinishedMsg{step: step, err: errors.New("exit 1"), output: "still broken"})
	m := tm.(model)

	if !m.runWaitFail {
		t.Fatal("expected to be paused on failure again")
	}
	if got := m.runFailCounts[step.ID]; got != 2 {
		t.Fatalf("expected 2 failures this run, got %d", got)
	}
	if !contains(m.View(), "failed 2 times this run") {
		t.Fatal("failure pause should show the in-run retry count")
	}
	t.Logf("PASS: in-run retry count tracked across retries")
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

func TestRunReoffersFailedAndSkipped(t *testing.T) {
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

	// Mark one completed (should be skipped), one failed and one skipped
	// (both should be re-offered).
	m.state.Steps[firstCatSteps[0].ID] = StatusCompleted
	m.state.Steps[firstCatSteps[1].ID] = StatusFailed
	m.state.Steps[firstCatSteps[2].ID] = StatusSkipped

	var tm tea.Model = m
	tm = sendKey(tm, "G")
	m = tm.(model)

	if m.screen != screenCategoryRun {
		t.Fatalf("expected category run screen, got %d", m.screen)
	}
	if m.isCategoryDone(firstCat) {
		t.Fatal("category with a skipped/failed step should not be marked done")
	}

	got := make(map[string]bool)
	for _, s := range m.runSteps {
		got[s.ID] = true
	}
	if got[firstCatSteps[0].ID] {
		t.Fatal("completed step should not be re-offered")
	}
	if !got[firstCatSteps[1].ID] {
		t.Fatal("previously failed step should be re-offered")
	}
	if !got[firstCatSteps[2].ID] {
		t.Fatal("previously skipped step should be re-offered")
	}
	t.Logf("PASS: failed/skipped steps re-offered, completed skipped")
}

func TestCategoryProgressFraction(t *testing.T) {
	state := &AppState{Steps: make(map[string]StepStatus)}
	m := newModel(state)

	firstCat := m.categories[0]
	var steps []Step
	for _, s := range AllSteps() {
		if s.Category == firstCat {
			steps = append(steps, s)
		}
	}
	if len(steps) < 4 {
		t.Fatalf("need >=4 steps in %q for this test", firstCat)
	}
	total := len(steps)

	// 2 completed, 1 skipped, the rest pending: done over (total − skipped).
	m.state.Steps[steps[0].ID] = StatusCompleted
	m.state.Steps[steps[1].ID] = StatusCompleted
	m.state.Steps[steps[2].ID] = StatusSkipped

	out := m.viewCategories()
	want := fmt.Sprintf("%d/%d + 1 skip", 2, total-1)
	if !contains(out, want) {
		t.Fatalf("expected category line to show %q, got:\n%s", want, out)
	}

	// Everything resolved (all completed, one skipped) → shows "done".
	for _, s := range steps {
		m.state.Steps[s.ID] = StatusCompleted
	}
	m.state.Steps[steps[0].ID] = StatusSkipped
	out = m.viewCategories()
	wantDone := fmt.Sprintf("%d/%d + 1 skip) done", total-1, total-1)
	if !contains(out, wantDone) {
		t.Fatalf("expected complete category to show %q, got:\n%s", wantDone, out)
	}
}

func TestArrowKeysMirrorEnterEsc(t *testing.T) {
	state := &AppState{Steps: make(map[string]StepStatus)}
	m := newModel(state)
	var tm tea.Model = m

	// Right behaves like Enter: drill into the first category.
	tm = sendSpecialKey(tm, tea.KeyRight)
	if got := tm.(model).screen; got != screenStepSelect {
		t.Fatalf("Right should act like Enter (enter category); screen=%d", got)
	}

	// Left behaves like Esc: go back to the categories screen.
	tm = sendSpecialKey(tm, tea.KeyLeft)
	if got := tm.(model).screen; got != screenCategories {
		t.Fatalf("Left should act like Esc (go back); screen=%d", got)
	}
}

func TestVimKeysMirrorEnterEsc(t *testing.T) {
	state := &AppState{Steps: make(map[string]StepStatus)}
	m := newModel(state)
	var tm tea.Model = m

	// l behaves like Enter (vim right): drill into the first category.
	tm = sendKey(tm, "l")
	if got := tm.(model).screen; got != screenStepSelect {
		t.Fatalf("l should act like Enter (enter category); screen=%d", got)
	}

	// h behaves like Esc (vim left): go back to the categories screen.
	tm = sendKey(tm, "h")
	if got := tm.(model).screen; got != screenCategories {
		t.Fatalf("h should act like Esc (go back); screen=%d", got)
	}
}

func TestRightSelectsInsideCategory(t *testing.T) {
	state := &AppState{Steps: make(map[string]StepStatus)}
	m := newModel(state)
	var tm tea.Model = m

	// Enter the first category's step list, then move to the first step.
	tm = sendSpecialKey(tm, tea.KeyEnter)
	tm = sendKey(tm, "j")
	m = tm.(model)
	if m.screen != screenStepSelect {
		t.Fatalf("expected step-select screen, got %d", m.screen)
	}
	step := m.stepSelectSteps[0]
	before := m.stepSelected[step.ID]

	// Right toggles the step (acts like Space) and stays in the list.
	tm = sendSpecialKey(tm, tea.KeyRight)
	m = tm.(model)
	if m.screen != screenStepSelect {
		t.Fatalf("Right inside a category should not navigate away; screen=%d", m.screen)
	}
	if m.stepSelected[step.ID] == before {
		t.Fatal("Right should toggle the step selection inside a category")
	}

	// l (vim) toggles it back.
	tm = sendKey(tm, "l")
	m = tm.(model)
	if m.stepSelected[step.ID] != before {
		t.Fatal("l should also toggle the step selection")
	}

	// Left still goes back to the categories screen.
	tm = sendSpecialKey(tm, tea.KeyLeft)
	if got := tm.(model).screen; got != screenCategories {
		t.Fatalf("Left should still go back; screen=%d", got)
	}
}

func TestLaunchSingleStep(t *testing.T) {
	state := &AppState{Steps: make(map[string]StepStatus)}
	m := newModel(state)
	m.dryRun = true
	var tm tea.Model = m

	// Enter the first category and move to its first step.
	tm = sendSpecialKey(tm, tea.KeyEnter)
	tm = sendKey(tm, "j")
	m = tm.(model)
	step := m.stepSelectSteps[0]

	// Shift+L launches just that step.
	tm = sendKey(tm, "L")
	m = tm.(model)
	if m.screen != screenCategoryRun {
		t.Fatalf("L should start a run; screen=%d", m.screen)
	}
	if len(m.runSteps) != 1 || m.runSteps[0].ID != step.ID {
		t.Fatalf("L should run exactly the cursor step, got %d steps", len(m.runSteps))
	}

	// When the run finishes, it returns to the step list (not categories).
	tm, _ = tm.Update(categoryDoneMsg{})
	tm = sendSpecialKey(tm, tea.KeyEnter)
	if got := tm.(model).screen; got != screenStepSelect {
		t.Fatalf("after a single-step launch, should return to the step list; screen=%d", got)
	}
}

func TestAdminStepUsesTerminalHandoff(t *testing.T) {
	// The mas installs and Zoom need a sudo password, so they must be admin.
	for _, id := range []string{"zoom-install", "things-install", "fantastical-install", "amphetamine-install"} {
		s, ok := StepByID(id)
		if !ok || !s.RequiresAdmin {
			t.Fatalf("%s should exist and be RequiresAdmin", id)
		}
	}

	step, _ := StepByID("zoom-install")
	m := newModel(&AppState{Steps: make(map[string]StepStatus)})
	m.runSteps = []Step{step}
	m.runIndex = 0

	// Real run: hands off to the terminal (a command), not the stream channel.
	updated, cmd := m.runCurrentStep()
	if cmd == nil {
		t.Fatal("admin step should return a command (terminal handoff)")
	}
	if updated.(model).runStream != nil {
		t.Fatal("admin step should not start the streaming channel")
	}

	// Dry-run: falls through to the streaming print path (never sudo).
	m.dryRun = true
	updated, _ = m.runCurrentStep()
	if updated.(model).runStream == nil {
		t.Fatal("dry-run admin step should use the streaming path, not handoff")
	}
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
