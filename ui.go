package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ── Screens ────────────────────────────────────────────────────────────────

type screen int

const (
	screenCategories screen = iota
	screenStepSelect
	screenCategoryRun
)

// ── Messages ───────────────────────────────────────────────────────────────

// runCategoryMsg triggers sequential execution of a category's selected steps.
type runCategoryMsg struct {
	category string
	steps    []Step
}

// stepResultMsg is sent after each step completes during a category run.
type stepResultMsg struct {
	step   Step
	result RunResult
	manual bool // true if this was a manual step (just acknowledged)
}

// manualStepMsg indicates a manual step needs user acknowledgement.
type manualStepMsg struct {
	step Step
}

// categoryDoneMsg signals all steps in the category have been run.
type categoryDoneMsg struct{}

// ── Model ──────────────────────────────────────────────────────────────────

type model struct {
	screen screen
	state  *AppState
	width  int
	height int

	// Category selection
	categories []string
	catCursor  int

	// Step selection (per-category drill-down)
	stepSelectCat    string
	stepSelectSteps  []Step
	stepSelectCursor int
	stepSelected     map[string]bool

	// Reset confirmation
	confirmReset bool

	// Category run state
	runCategory   string          // category currently being run
	runSteps      []Step          // steps to run in this category
	runLog        []runLogEntry   // log of completed steps
	runIndex      int             // index of currently running step
	runWaitManual bool            // waiting for Enter on a manual step
	runManualStep *Step           // the manual step we're waiting on
	runDone       bool            // all steps finished

	// Mode
	dryRun bool
}

type runLogEntry struct {
	name   string
	status string // "ok", "fail", "manual", "skip"
}

func newModel(state *AppState) model {
	cats := Categories()

	stepSel := make(map[string]bool)
	if state.SelectedSteps != nil {
		for _, step := range AllSteps() {
			stepSel[step.ID] = state.SelectedSteps[step.ID]
		}
	} else {
		for _, step := range AllSteps() {
			stepSel[step.ID] = true
		}
	}

	return model{
		screen:       screenCategories,
		state:        state,
		categories:   cats,
		stepSelected: stepSel,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

// ── Update ─────────────────────────────────────────────────────────────────

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case stepResultMsg:
		entry := runLogEntry{name: msg.step.Name, status: "ok"}
		if msg.manual {
			entry.status = "manual"
		} else if msg.result.Err != nil {
			entry.status = "fail"
			m.state.Steps[msg.step.ID] = StatusFailed
		} else {
			m.state.Steps[msg.step.ID] = StatusCompleted
		}
		m.runLog = append(m.runLog, entry)
		m.saveState()
		// Continue to next step
		return m.runNextStep()

	case manualStepMsg:
		m.runWaitManual = true
		m.runManualStep = &msg.step
		return m, nil

	case categoryDoneMsg:
		m.runDone = true
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if key == "ctrl+c" {
		m.saveState()
		return m, tea.Quit
	}

	switch m.screen {
	case screenCategories:
		if m.confirmReset {
			switch key {
			case "y", "Y":
				cat := m.categories[m.catCursor]
				for _, s := range AllSteps() {
					if s.Category == cat {
						delete(m.state.Steps, s.ID)
					}
				}
				m.saveState()
				m.confirmReset = false
			default:
				m.confirmReset = false
			}
			return m, nil
		}
		switch key {
		case "up", "k":
			if m.catCursor > 0 {
				m.catCursor--
			}
		case "down", "j":
			if m.catCursor < len(m.categories)-1 {
				m.catCursor++
			}
		case "enter", " ":
			cat := m.categories[m.catCursor]
			m.stepSelectCat = cat
			m.stepSelectSteps = nil
			for _, s := range AllSteps() {
				if s.Category == cat {
					m.stepSelectSteps = append(m.stepSelectSteps, s)
				}
			}
			m.stepSelectCursor = 0
			m.screen = screenStepSelect
		case "R":
			m.confirmReset = true
		case "G":
			return m.startCategoryRun()
		case "q":
			return m, tea.Quit
		}

	case screenStepSelect:
		if m.confirmReset {
			switch key {
			case "y", "Y":
				for _, s := range m.stepSelectSteps {
					delete(m.state.Steps, s.ID)
				}
				m.saveState()
				m.confirmReset = false
			default:
				m.confirmReset = false
			}
			return m, nil
		}
		totalRows := len(m.stepSelectSteps) + 1
		switch key {
		case "up", "k":
			if m.stepSelectCursor > 0 {
				m.stepSelectCursor--
			}
		case "down", "j":
			if m.stepSelectCursor < totalRows-1 {
				m.stepSelectCursor++
			}
		case " ":
			if m.stepSelectCursor == 0 {
				allOn := m.allStepsSelectedInCat()
				for _, s := range m.stepSelectSteps {
					m.stepSelected[s.ID] = !allOn
				}
			} else {
				step := m.stepSelectSteps[m.stepSelectCursor-1]
				m.stepSelected[step.ID] = !m.stepSelected[step.ID]
			}
		case "R":
			m.confirmReset = true
		case "G":
			m.screen = screenCategories
			return m.startCategoryRun()
		case "backspace", "esc":
			m.screen = screenCategories
		case "enter":
			m.screen = screenCategories
		case "q":
			return m, tea.Quit
		}

	case screenCategoryRun:
		if m.runWaitManual {
			if key == "enter" {
				// Acknowledge manual step
				step := *m.runManualStep
				m.runWaitManual = false
				m.runManualStep = nil
				m.state.Steps[step.ID] = StatusCompleted
				return m, func() tea.Msg {
					return stepResultMsg{step: step, manual: true}
				}
			}
			if key == "q" {
				m.saveState()
				return m, tea.Quit
			}
			return m, nil
		}
		if m.runDone {
			if key == "enter" || key == "esc" {
				m.screen = screenCategories
				return m, nil
			}
			if key == "q" {
				m.saveState()
				return m, tea.Quit
			}
		}
		if key == "q" {
			m.saveState()
			return m, tea.Quit
		}

	}

	return m, nil
}

// startCategoryRun gathers selected steps for the current category and starts running them.
func (m model) startCategoryRun() (tea.Model, tea.Cmd) {
	cat := m.categories[m.catCursor]
	var steps []Step
	for _, s := range AllSteps() {
		if s.Category == cat && m.stepSelected[s.ID] {
			// Skip already completed/skipped steps
			status := m.state.Steps[s.ID]
			if status == StatusCompleted || status == StatusSkipped {
				continue
			}
			steps = append(steps, s)
		}
	}

	if len(steps) == 0 {
		// Nothing to run — category already done or nothing selected
		return m, nil
	}

	m.screen = screenCategoryRun
	m.runCategory = cat
	m.runSteps = steps
	m.runLog = nil
	m.runIndex = 0
	m.runDone = false
	m.runWaitManual = false
	m.runManualStep = nil

	// Start running the first step
	return m.runCurrentStep()
}

// runCurrentStep dispatches the current step (automated or manual).
func (m model) runCurrentStep() (tea.Model, tea.Cmd) {
	if m.runIndex >= len(m.runSteps) {
		return m, func() tea.Msg { return categoryDoneMsg{} }
	}

	step := m.runSteps[m.runIndex]

	// Manual step: no commands, just instructions
	if step.ManualInstructions != "" && len(step.Commands) == 0 {
		return m, func() tea.Msg { return manualStepMsg{step: step} }
	}

	// Automated step: run commands
	dryRun := m.dryRun
	return m, func() tea.Msg {
		var result RunResult
		if dryRun {
			result = DryRunCommands(step.Commands)
		} else {
			result = RunCommands(step.Commands)
		}
		return stepResultMsg{step: step, result: result}
	}
}

// runNextStep advances to the next step and dispatches it.
func (m model) runNextStep() (tea.Model, tea.Cmd) {
	m.runIndex++
	return m.runCurrentStep()
}

// saveState syncs selection and persists state to disk.
func (m model) saveState() {
	m.state.SelectedSteps = make(map[string]bool)
	for id, sel := range m.stepSelected {
		if sel {
			m.state.SelectedSteps[id] = true
		}
	}
	m.state.Save()
}

// navigateCategory moves to the next or previous category within the step select screen.
func (m *model) navigateCategory(dir int) {
	idx := -1
	for i, cat := range m.categories {
		if cat == m.stepSelectCat {
			idx = i
			break
		}
	}
	idx += dir
	if idx < 0 || idx >= len(m.categories) {
		return
	}
	m.stepSelectCat = m.categories[idx]
	m.catCursor = idx
	m.stepSelectSteps = nil
	for _, s := range AllSteps() {
		if s.Category == m.stepSelectCat {
			m.stepSelectSteps = append(m.stepSelectSteps, s)
		}
	}
	m.stepSelectCursor = 0
}

func (m model) allStepsSelectedInCat() bool {
	for _, s := range m.stepSelectSteps {
		if !m.stepSelected[s.ID] {
			return false
		}
	}
	return true
}

// isCategoryDone returns true if all selected steps in the category are completed/skipped.
func (m model) isCategoryDone(cat string) bool {
	hasSelected := false
	for _, s := range AllSteps() {
		if s.Category == cat && m.stepSelected[s.ID] {
			hasSelected = true
			status := m.state.Steps[s.ID]
			if status != StatusCompleted && status != StatusSkipped {
				return false
			}
		}
	}
	return hasSelected
}

// ── View ───────────────────────────────────────────────────────────────────

func (m model) View() string {
	switch m.screen {
	case screenCategories:
		return m.viewCategories()
	case screenStepSelect:
		return m.viewStepSelect()
	case screenCategoryRun:
		return m.viewCategoryRun()
	}
	return ""
}

func (m model) viewCategories() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(styleTitle.Render("  Mac Setup"))
	if m.dryRun {
		b.WriteString("  " + styleWarning.Render("[DRY RUN]"))
	}
	b.WriteString("\n\n")

	for i, cat := range m.categories {
		cursor := "  "
		if i == m.catCursor {
			cursor = styleTitle.Render("▸ ")
		}

		total := 0
		selected := 0
		for _, s := range AllSteps() {
			if s.Category == cat {
				total++
				if m.stepSelected[s.ID] {
					selected++
				}
			}
		}

		done := m.isCategoryDone(cat)

		var icon string
		var label string
		if done {
			icon = styleSuccess.Render("✓")
			label = styleSuccess.Render(fmt.Sprintf("%s (%d/%d) done", cat, selected, total))
		} else {
			icon = styleDim.Render("·")
			label = styleUnselected.Render(fmt.Sprintf("%s (%d/%d)", cat, selected, total))
		}

		b.WriteString(fmt.Sprintf("  %s%s %s\n", cursor, icon, label))
	}

	b.WriteString("\n")
	if m.confirmReset {
		b.WriteString(styleWarning.Render("  Reset all checkmarks for this category? [y] Yes  [n] No"))
		b.WriteString("\n")
	} else {
		b.WriteString(help("  [G] Run category"))
		b.WriteString("\n")
		b.WriteString(help("  [↑/k] Up  [↓/j] Down  [Enter] Choose steps"))
		b.WriteString("\n")
		b.WriteString(help("  [q] Quit  [R] Reset checkmarks"))
		b.WriteString("\n")
	}
	return b.String()
}

func (m model) viewStepSelect() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(styleTitle.Render(fmt.Sprintf("  %s — select steps", m.stepSelectCat)))
	if m.dryRun {
		b.WriteString("  " + styleWarning.Render("[DRY RUN]"))
	}
	b.WriteString("\n\n")

	// Row 0: Select All
	{
		cursor := "  "
		if m.stepSelectCursor == 0 {
			cursor = styleTitle.Render("▸ ")
		}
		allOn := m.allStepsSelectedInCat()
		check := styleSelected.Render("●")
		if !allOn {
			check = styleSkipped.Render("○")
		}
		label := styleCategory.Render("Select All")
		b.WriteString(fmt.Sprintf("  %s%s %s\n", cursor, check, label))
	}

	b.WriteString(styleDim.Render("  ──────────────────────────────────────") + "\n")

	for i, step := range m.stepSelectSteps {
		row := i + 1
		cursor := "  "
		if m.stepSelectCursor == row {
			cursor = styleTitle.Render("▸ ")
		}

		status := m.state.Steps[step.ID]
		var check string
		if status == StatusCompleted {
			check = styleSuccess.Render("✓")
		} else if m.stepSelected[step.ID] {
			check = styleSelected.Render("●")
		} else {
			check = styleSkipped.Render("○")
		}

		name := styleUnselected.Render(step.Name)
		desc := styleDescription.Render(" — " + step.Description)
		if status == StatusCompleted {
			name = styleSuccess.Render(step.Name)
			desc = styleDescription.Render(" — " + step.Description)
		} else if !m.stepSelected[step.ID] {
			name = styleSkipped.Render(step.Name)
			desc = styleSkipped.Render(" — " + step.Description)
		}

		manual := ""
		if step.ManualInstructions != "" && len(step.Commands) == 0 {
			manual = styleWarning.Render(" ✋")
		}
		b.WriteString(fmt.Sprintf("  %s%s %s%s%s\n", cursor, check, name, manual, desc))
	}

	b.WriteString("\n")
	if m.confirmReset {
		b.WriteString(styleWarning.Render("  Reset all checkmarks for this category? [y] Yes  [n] No"))
		b.WriteString("\n")
	} else {
		b.WriteString(help("  [G] Run category"))
		b.WriteString("\n")
		b.WriteString(help("  [↑/k] Up  [↓/j] Down  [Space] Toggle"))
		b.WriteString("\n")
		b.WriteString(help("  [Esc] Back  [R] Reset checkmarks"))
		b.WriteString("\n")
	}
	return b.String()
}

func (m model) viewCategoryRun() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(styleTitle.Render(fmt.Sprintf("  Running: %s", m.runCategory)))
	if m.dryRun {
		b.WriteString("  " + styleWarning.Render("[DRY RUN]"))
	}
	b.WriteString("\n\n")

	// Show completed log entries
	for _, entry := range m.runLog {
		var icon string
		switch entry.status {
		case "ok":
			icon = styleSuccess.Render("✓")
		case "fail":
			icon = styleError.Render("✗")
		case "manual":
			icon = styleWarning.Render("✋")
		}
		b.WriteString(fmt.Sprintf("  %s %s\n", icon, styleDescription.Render(entry.name)))
	}

	// Show current state
	if m.runWaitManual && m.runManualStep != nil {
		step := *m.runManualStep
		b.WriteString("\n")
		b.WriteString(styleWarning.Render(fmt.Sprintf("  ✋ %s", step.Name)) + "\n")
		b.WriteString("\n")
		for _, line := range strings.Split(step.ManualInstructions, "\n") {
			b.WriteString(styleManual.Render("  "+line) + "\n")
		}
		b.WriteString("\n")
		b.WriteString(help("  Press [Enter] when done  •  [q] Quit"))
		b.WriteString("\n")
	} else if m.runDone {
		b.WriteString("\n")

		// Count results
		okCount, failCount := 0, 0
		for _, e := range m.runLog {
			switch e.status {
			case "ok", "manual":
				okCount++
			case "fail":
				failCount++
			}
		}

		if failCount > 0 {
			b.WriteString(styleWarning.Render(fmt.Sprintf("  Done — %d completed, %d failed", okCount, failCount)) + "\n")
		} else {
			b.WriteString(styleSuccess.Render(fmt.Sprintf("  Done — %d steps completed", okCount)) + "\n")
		}
		b.WriteString("\n")
		b.WriteString(help("  Press [Enter] to return  •  [q] Quit"))
		b.WriteString("\n")
	} else if m.runIndex < len(m.runSteps) {
		// Currently running an automated step
		step := m.runSteps[m.runIndex]
		b.WriteString(fmt.Sprintf("  %s %s\n",
			styleDim.Render("⋯"),
			styleDim.Render(step.Name)))
	}

	return b.String()
}


// help renders a help line with dim text and yellow [keys].
func help(s string) string {
	var b strings.Builder
	i := 0
	for i < len(s) {
		bracketStart := strings.IndexByte(s[i:], '[')
		if bracketStart == -1 {
			b.WriteString(styleHelp.Render(s[i:]))
			break
		}
		if bracketStart > 0 {
			b.WriteString(styleHelp.Render(s[i : i+bracketStart]))
		}
		bracketEnd := strings.IndexByte(s[i+bracketStart:], ']')
		if bracketEnd == -1 {
			b.WriteString(styleHelp.Render(s[i:]))
			break
		}
		keyEnd := i + bracketStart + bracketEnd + 1
		b.WriteString(styleWarning.Render(s[i+bracketStart : keyEnd]))
		i = keyEnd
		continue
	}
	return b.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

