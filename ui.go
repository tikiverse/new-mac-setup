package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
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

// streamLineMsg carries one line of streamed output from the running step.
// When replace is true, it overwrites the previous line (a \r progress update).
type streamLineMsg struct {
	line    string
	replace bool
}

// stepFinishedMsg is sent when a step's commands finish (or fail).
type stepFinishedMsg struct {
	step   Step
	err    error
	output string // full captured output of the step
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
	runCategory     string         // category currently being run
	runSteps        []Step         // steps to run in this category
	runLog          []runLogEntry  // log of completed steps
	runIndex        int            // index of currently running step
	runWaitManual   bool           // waiting for Enter on a manual step
	runManualStep   *Step          // the manual step we're waiting on
	runWaitFail     bool           // paused on a failed step, awaiting retry/skip/abort
	runFailStep     *Step          // the step that failed
	runFailOutput   string         // captured output of the failed step
	runFailCounts   map[string]int // per-step failure count within the current run
	runDone         bool           // all steps finished
	runReturnScreen screen         // screen to return to when the run finishes
	runStream       chan tea.Msg   // channel of streamed output for the running step
	runLines        []string       // accumulated output lines for the current run
	runViewport     viewport.Model // scrollable view of streamed output

	// Mode
	dryRun bool
	debug  bool // show the Testing category
}

type runLogEntry struct {
	name   string
	status string // "ok", "fail", "manual", "skip"
}

func newModel(state *AppState) model {
	cats := visibleCategories(false)

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
		runViewport:  viewport.New(80, 14),
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
		w := msg.Width - 4
		if w < 20 {
			w = 20
		}
		h := msg.Height / 2
		if h < 6 {
			h = 6
		} else if h > 16 {
			h = 16
		}
		m.runViewport.Width = w
		m.runViewport.Height = h
		return m, nil

	case streamLineMsg:
		if msg.replace && len(m.runLines) > 0 {
			m.runLines[len(m.runLines)-1] = msg.line
		} else {
			m.runLines = append(m.runLines, msg.line)
		}
		m.runViewport.SetContent(strings.Join(m.runLines, "\n"))
		m.runViewport.GotoBottom()
		return m, waitForStream(m.runStream)

	case stepFinishedMsg:
		if msg.err != nil {
			// Pause and let the user decide: retry, skip, or abort.
			s := msg.step
			m.runWaitFail = true
			m.runFailStep = &s
			m.runFailOutput = msg.output
			if m.runFailCounts == nil {
				m.runFailCounts = make(map[string]int)
			}
			m.runFailCounts[s.ID]++
			return m, nil
		}
		m.state.Steps[msg.step.ID] = StatusCompleted
		m.runLog = append(m.runLog, runLogEntry{name: msg.step.Name, status: "ok"})
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

	// Left/Right arrows (and vim h/l) mirror Esc/Enter. Inside the step list,
	// Right/l selects (toggles) the step under the cursor instead, so Left = back
	// and Right = select.
	switch key {
	case "left", "h":
		key = "esc"
	case "right", "l":
		if m.screen == screenStepSelect {
			key = " "
		} else {
			key = "enter"
		}
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
		case "L":
			// Launch just the step under the cursor, run to completion.
			if m.stepSelectCursor > 0 {
				return m.startSingleStepRun(m.stepSelectSteps[m.stepSelectCursor-1])
			}
		case "backspace", "esc":
			m.screen = screenCategories
		case "enter":
			m.screen = screenCategories
		case "q":
			return m, tea.Quit
		}

	case screenCategoryRun:
		if m.runWaitFail {
			step := *m.runFailStep
			switch key {
			case "r", "R":
				// Retry the same step (runIndex still points at it).
				m.runWaitFail = false
				m.runFailStep = nil
				m.runFailOutput = ""
				return m.runCurrentStep()
			case "s", "S":
				// Skip: record the choice and move on.
				m.runWaitFail = false
				m.runFailStep = nil
				m.runFailOutput = ""
				m.state.Steps[step.ID] = StatusSkipped
				m.runLog = append(m.runLog, runLogEntry{name: step.Name, status: "skip"})
				m.saveState()
				return m.runNextStep()
			case "a", "A":
				// Abort: record the failure and stop the run.
				m.runWaitFail = false
				m.runFailStep = nil
				m.runFailOutput = ""
				m.state.Steps[step.ID] = StatusFailed
				m.runLog = append(m.runLog, runLogEntry{name: step.Name, status: "fail"})
				m.saveState()
				m.runDone = true
				return m, nil
			case "q":
				m.saveState()
				return m, tea.Quit
			}
			return m, nil
		}
		if m.runWaitManual {
			if key == "enter" {
				// Acknowledge manual step and advance.
				step := *m.runManualStep
				m.runWaitManual = false
				m.runManualStep = nil
				m.state.Steps[step.ID] = StatusCompleted
				m.runLog = append(m.runLog, runLogEntry{name: step.Name, status: "manual"})
				m.saveState()
				return m.runNextStep()
			}
			if key == "q" {
				m.saveState()
				return m, tea.Quit
			}
			return m, nil
		}
		if m.runDone {
			if key == "enter" || key == "esc" {
				m.screen = m.runReturnScreen
				return m, nil
			}
			if key == "q" {
				m.saveState()
				return m, tea.Quit
			}
			return m, nil
		}
		if key == "q" {
			m.saveState()
			return m, tea.Quit
		}
		// Streaming a step: forward keys to the viewport for scrollback.
		var cmd tea.Cmd
		m.runViewport, cmd = m.runViewport.Update(msg)
		return m, cmd
	}

	return m, nil
}

// startCategoryRun gathers selected steps for the current category and starts running them.
func (m model) startCategoryRun() (tea.Model, tea.Cmd) {
	cat := m.categories[m.catCursor]
	var steps []Step
	for _, s := range AllSteps() {
		if s.Category == cat && m.stepSelected[s.ID] {
			// Only skip steps that fully completed; previously failed or
			// skipped steps are re-offered so the user is prompted again.
			if m.state.Steps[s.ID] == StatusCompleted {
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
	m.runFailCounts = make(map[string]int)
	m.runReturnScreen = screenCategories
	m.runLines = nil
	m.runViewport.SetContent("")

	// Start running the first step
	return m.runCurrentStep()
}

// startSingleStepRun runs a single step to completion, returning to the step
// list when finished (used by the "L" launch shortcut).
func (m model) startSingleStepRun(step Step) (tea.Model, tea.Cmd) {
	m.screen = screenCategoryRun
	m.runCategory = step.Category
	m.runSteps = []Step{step}
	m.runLog = nil
	m.runIndex = 0
	m.runDone = false
	m.runWaitManual = false
	m.runManualStep = nil
	m.runFailCounts = make(map[string]int)
	m.runReturnScreen = screenStepSelect
	m.runLines = nil
	m.runViewport.SetContent("")

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

	// Admin step: hand the terminal over (drop alt-screen) so sudo/installers
	// can prompt for a password. This bypasses streaming — the command owns the
	// terminal — then the TUI resumes. Dry-run falls through to the print path.
	if step.RequiresAdmin && !m.dryRun {
		s := step
		return m, tea.ExecProcess(interactiveCommand(s), func(err error) tea.Msg {
			out := ""
			if err != nil {
				out = "(this step ran in your terminal; output was not captured)"
			}
			return stepFinishedMsg{step: s, err: err, output: out}
		})
	}

	// Automated step: stream commands into the viewport.
	ch := make(chan tea.Msg, 256)
	m.runStream = ch
	dryRun := m.dryRun
	go streamStep(step, dryRun, ch)
	return m, waitForStream(ch)
}

// streamStep runs a step's commands, emitting each output line on ch followed
// by a final stepFinishedMsg, then closes ch.
func streamStep(step Step, dryRun bool, ch chan tea.Msg) {
	defer close(ch)
	var full strings.Builder
	record := func(text string) { full.WriteString(text + "\n") }

	for _, cmd := range step.Commands {
		ch <- streamLineMsg{line: "$ " + cmd}
		record("$ " + cmd)

		if dryRun {
			ch <- streamLineMsg{line: "  (dry run — skipped)"}
			record("  (dry run — skipped)")
			continue
		}

		err := streamCommand(cmd, func(text string, replace bool) {
			ch <- streamLineMsg{line: text, replace: replace}
			if !replace {
				record(text)
			}
		})
		if err != nil {
			ch <- streamLineMsg{line: "✗ " + err.Error()}
			record("✗ " + err.Error())
			ch <- stepFinishedMsg{step: step, err: err, output: full.String()}
			return
		}
	}
	ch <- stepFinishedMsg{step: step, output: full.String()}
}

// waitForStream returns a command that blocks until the next streamed message.
func waitForStream(ch chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return msg
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

// isCategoryDone reports whether every step in the category is resolved —
// completed or deliberately skipped, with nothing left pending or failed.
func (m model) isCategoryDone(cat string) bool {
	any := false
	for _, s := range AllSteps() {
		if s.Category != cat {
			continue
		}
		any = true
		switch m.state.Steps[s.ID] {
		case StatusCompleted, StatusSkipped:
			// resolved
		default:
			return false
		}
	}
	return any
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
	if m.debug {
		b.WriteString("  " + styleWarning.Render("[DEBUG]"))
	}
	b.WriteString("\n\n")

	for i, cat := range m.categories {
		cursor := "  "
		if i == m.catCursor {
			cursor = styleTitle.Render("▸ ")
		}

		// Progress is "done / (total − skipped)", with skipped steps set aside
		// and surfaced separately so they don't count against the category.
		total, doneCount, skipped := 0, 0, 0
		for _, s := range AllSteps() {
			if s.Category != cat {
				continue
			}
			total++
			switch m.state.Steps[s.ID] {
			case StatusCompleted:
				doneCount++
			case StatusSkipped:
				skipped++
			}
		}

		frac := fmt.Sprintf("%d/%d", doneCount, total-skipped)
		if skipped > 0 {
			frac += fmt.Sprintf(" + %d skip", skipped)
		}

		var icon string
		var label string
		if m.isCategoryDone(cat) {
			icon = styleSuccess.Render("✓")
			label = styleSuccess.Render(fmt.Sprintf("%s (%s) done", cat, frac))
		} else {
			icon = styleDim.Render("·")
			label = styleUnselected.Render(fmt.Sprintf("%s (%s)", cat, frac))
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
		hist := ""
		switch status {
		case StatusFailed:
			hist = styleError.Render(" (failed before)")
		case StatusSkipped:
			hist = styleWarning.Render(" (skipped before)")
		}
		id := styleDim.Render(" " + step.ID)
		b.WriteString(fmt.Sprintf("  %s%s %s%s%s%s%s\n", cursor, check, name, id, manual, hist, desc))
	}

	b.WriteString("\n")
	if m.confirmReset {
		b.WriteString(styleWarning.Render("  Reset all checkmarks for this category? [y] Yes  [n] No"))
		b.WriteString("\n")
	} else {
		b.WriteString(help("  [G] Run category  [L] Launch step"))
		b.WriteString("\n")
		b.WriteString(help("  [↑/k] Up  [↓/j] Down  [Space/→] Toggle"))
		b.WriteString("\n")
		b.WriteString(help("  [Esc/←] Back  [R] Reset checkmarks"))
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
		case "skip":
			icon = styleSkipped.Render("↷")
		case "manual":
			icon = styleWarning.Render("✋")
		}
		b.WriteString(fmt.Sprintf("  %s %s\n", icon, styleDescription.Render(entry.name)))
	}

	// Show current state
	if m.runWaitFail && m.runFailStep != nil {
		step := *m.runFailStep
		b.WriteString("\n")
		b.WriteString(styleError.Render(fmt.Sprintf("  ✗ %s failed", step.Name)) + "\n")
		if n := m.runFailCounts[step.ID]; n > 1 {
			b.WriteString(styleWarning.Render(fmt.Sprintf("  ↻ failed %d times this run", n)) + "\n")
		}
		// State is written only when the user resolves this pause, so it still
		// holds the status from a previous run here.
		switch m.state.Steps[step.ID] {
		case StatusFailed:
			b.WriteString(styleWarning.Render("  ↻ this step also failed on a previous run") + "\n")
		case StatusSkipped:
			b.WriteString(styleWarning.Render("  ↻ you skipped this step on a previous run") + "\n")
		}
		b.WriteString("\n")
		for _, line := range tailLines(m.runFailOutput, 12) {
			b.WriteString(styleDim.Render("  "+line) + "\n")
		}
		b.WriteString("\n")
		b.WriteString(help("  [r] Retry  •  [s] Skip  •  [a] Abort run  •  [q] Quit"))
		b.WriteString("\n")
	} else if m.runWaitManual && m.runManualStep != nil {
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
		okCount, failCount, skipCount := 0, 0, 0
		for _, e := range m.runLog {
			switch e.status {
			case "ok", "manual":
				okCount++
			case "fail":
				failCount++
			case "skip":
				skipCount++
			}
		}

		if failCount > 0 || skipCount > 0 {
			b.WriteString(styleWarning.Render(fmt.Sprintf("  Done — %d completed, %d failed, %d skipped", okCount, failCount, skipCount)) + "\n")
		} else {
			b.WriteString(styleSuccess.Render(fmt.Sprintf("  Done — %d steps completed", okCount)) + "\n")
		}
		b.WriteString("\n")
		b.WriteString(help("  Press [Enter] to return  •  [q] Quit"))
		b.WriteString("\n")
	} else if m.runIndex < len(m.runSteps) {
		// Currently streaming an automated step's output.
		step := m.runSteps[m.runIndex]
		b.WriteString(fmt.Sprintf("  %s %s\n\n",
			styleDim.Render("⋯"),
			styleDim.Render(step.Name)))
		b.WriteString(m.runViewport.View() + "\n\n")
		b.WriteString(help("  [↑/↓ PgUp/PgDn] Scroll  •  [q] Quit"))
		b.WriteString("\n")
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

// tailLines returns the last n non-trailing-empty lines of s, for bounded
// display of captured command output.
func tailLines(s string, n int) []string {
	s = strings.TrimRight(s, "\n")
	if s == "" {
		return []string{"(no output)"}
	}
	lines := strings.Split(s, "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return lines
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
