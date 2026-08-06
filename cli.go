package main

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

// appName is the installed executable's name, used in help text and in the
// TUI's "run this step by itself" hint.
const appName = "mac-setup"

const usage = `mac-setup — a-la-carte Mac setup

Usage:
  mac-setup                      Launch the interactive TUI
  mac-setup --dry-run            Launch the TUI in dry-run mode
  mac-setup --debug              Launch the TUI with the hidden Testing category
  mac-setup <step-id>            Show the step's metadata and command(s)
  mac-setup <step-id> --run      Run the step directly in this terminal
  mac-setup <step-id> --done     Mark the step as done (no run)
  mac-setup <step-id> --reset    Mark the step as not done (clear its status)
  mac-setup <step-id> --copy     Copy the step's command(s) to the clipboard

Flags:
  --run           Execute the step (with --dry-run, print without running)
  -n, --dry-run   With --run, print commands instead of executing them
  --debug         Show the hidden Testing category in the TUI
  -h, --help      Show this help

Step ids are shown next to each step in the TUI.
`

type cliAction int

const (
	actionShow cliAction = iota // default: print metadata + command, no execution
	actionRun
	actionDone
	actionReset
	actionCopy
)

type cliOptions struct {
	stepID string
	action cliAction
	dryRun bool
	debug  bool
	help   bool
}

// parseArgs parses CLI arguments. Flags may appear before or after the step id.
func parseArgs(args []string) (cliOptions, error) {
	var o cliOptions
	actionSet := false
	setAction := func(a cliAction) error {
		if actionSet {
			return fmt.Errorf("only one of --run, --done, --reset, --copy may be given")
		}
		o.action = a
		actionSet = true
		return nil
	}

	for _, a := range args {
		switch a {
		case "--run":
			if err := setAction(actionRun); err != nil {
				return o, err
			}
		case "--done":
			if err := setAction(actionDone); err != nil {
				return o, err
			}
		case "--reset":
			if err := setAction(actionReset); err != nil {
				return o, err
			}
		case "--copy":
			if err := setAction(actionCopy); err != nil {
				return o, err
			}
		case "-n", "--dry-run":
			o.dryRun = true
		case "--debug":
			o.debug = true
		case "-h", "--help":
			o.help = true
		default:
			if strings.HasPrefix(a, "-") {
				return o, fmt.Errorf("unknown flag: %s", a)
			}
			if o.stepID != "" {
				return o, fmt.Errorf("only one step id may be given (got %q and %q)", o.stepID, a)
			}
			o.stepID = a
		}
	}

	// The action flags all operate on a single step, so they are meaningless
	// without a step id. Reject them instead of falling through to the TUI,
	// which would silently ignore what the user asked for. Each action is
	// listed by name so that TUI-only flags added later are unaffected.
	if o.stepID == "" {
		switch o.action {
		case actionRun, actionDone, actionReset, actionCopy:
			return o, fmt.Errorf("%s requires a step id", actionFlag(o.action))
		}
	}

	return o, nil
}

// actionFlag returns the flag that selects an action, for use in error text.
func actionFlag(a cliAction) string {
	switch a {
	case actionRun:
		return "--run"
	case actionDone:
		return "--done"
	case actionReset:
		return "--reset"
	case actionCopy:
		return "--copy"
	default:
		return ""
	}
}

// runDirect performs a single-step action from the CLI and returns an exit code.
func runDirect(opts cliOptions) int {
	step, ok := StepByID(opts.stepID)
	if !ok {
		printUnknownStepID(opts.stepID)
		return 1
	}
	state, err := LoadState()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}

	switch opts.action {
	case actionShow:
		return showStep(step, state)

	case actionCopy:
		payload := clipboardPayload(step)
		if err := copyToClipboard(payload); err != nil {
			fmt.Fprintf(os.Stderr, "Copy failed: %v\n", err)
			return 1
		}
		fmt.Printf("Copied %s to the clipboard:\n\n%s\n", step.ID, payload)
		return 0

	case actionDone:
		state.Steps[step.ID] = StatusCompleted
		if err := state.Save(); err != nil {
			fmt.Fprintf(os.Stderr, "Save failed: %v\n", err)
			return 1
		}
		fmt.Printf("Marked %s as done.\n", step.ID)
		return 0

	case actionReset:
		delete(state.Steps, step.ID)
		if err := state.Save(); err != nil {
			fmt.Fprintf(os.Stderr, "Save failed: %v\n", err)
			return 1
		}
		fmt.Printf("Reset %s to not done.\n", step.ID)
		return 0
	}

	// actionRun: execute the step directly in this terminal.
	if step.IsManual() {
		fmt.Printf("%s is a manual step:\n\n%s\n\n", step.ID, step.ManualInstructions)
		fmt.Printf("When finished, mark it done with: mac-setup %s --done\n", step.ID)
		return 0
	}

	if err := runStepDirect(step, opts.dryRun); err != nil {
		state.Steps[step.ID] = StatusFailed
		state.Save()
		fmt.Fprintf(os.Stderr, "\n✗ %s failed: %v\n", step.ID, err)
		return 1
	}

	if opts.dryRun {
		return 0
	}
	refreshShellEnv()
	state.Steps[step.ID] = StatusCompleted
	if err := state.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "Save failed: %v\n", err)
		return 1
	}
	fmt.Printf("\n✓ %s done.\n", step.ID)
	if step.Note != "" {
		fmt.Printf("\n%s %s\n", styleWarning.Render("Note:"), styleManual.Render(step.Note))
	}
	return 0
}

// minSubstringQuery is the shortest input allowed to match on containment
// alone. Below it a query is too weak to mean anything — a lone "s" appears in
// 65 of the ids — so short inputs are held to the stronger prefix and typo
// tests instead.
const minSubstringQuery = 3

// maxSuggestions bounds the printed list. It sits above what any realistic
// query matches (the widest, "finder", finds 7) so it only fires on a fragment
// like "install" that matches every install step, and the remainder is counted
// rather than dropped silently.
const maxSuggestions = 10

// printUnknownStepID reports an id that matched no step, offering the closest
// ids it can find. Everything goes to stderr so a caller piping stdout still
// sees why nothing happened.
func printUnknownStepID(id string) {
	fmt.Fprintf(os.Stderr, "Unknown step id: %q\n", id)

	matches := suggestStepIDs(id)
	rest := 0
	if len(matches) > maxSuggestions {
		rest = len(matches) - maxSuggestions
		matches = matches[:maxSuggestions]
	}

	switch len(matches) {
	case 0:
	case 1:
		fmt.Fprintf(os.Stderr, "\nDid you mean %s?\n", styleWarning.Render(matches[0]))
	default:
		fmt.Fprintln(os.Stderr, "\nDid you mean one of these?")
		width := 0
		for _, m := range matches {
			if len(m) > width {
				width = len(m)
			}
		}
		for _, m := range matches {
			// Pad before styling: the ANSI codes lipgloss adds would otherwise
			// be counted as width by a %-*s verb.
			pad := strings.Repeat(" ", width-len(m))
			name := ""
			if step, ok := StepByID(m); ok {
				name = step.Name
			}
			fmt.Fprintf(os.Stderr, "  %s%s  %s\n", styleWarning.Render(m), pad, styleDescription.Render(name))
		}
		// Count what was left out, so the list never reads as exhaustive.
		if rest > 0 {
			fmt.Fprintf(os.Stderr, "  %s\n", styleDim.Render(fmt.Sprintf("… and %d more", rest)))
		}
	}

	fmt.Fprintln(os.Stderr, "\nRun mac-setup with no arguments to browse step ids.")
}

// suggestStepIDs returns every step id close enough to input to be worth
// offering, best match first, or nothing when the input resembles no id at all.
//
// Every match is returned; the caller decides how many to print. The tests for
// what counts as a match are kept strict enough that realistic queries stay
// well inside that limit, since withholding a match risks hiding the very id
// the user was reaching for.
//
// Matching is case-insensitive. An id containing what was typed ranks above a
// misspelling of one, since "tailscale" is a person naming the thing they want
// rather than fumbling the keys. Equally good matches keep the order they are
// declared in, which is the order the TUI lists them and the order they are
// meant to be run — so "syncthing" offers syncthing-install before the pairing
// step that follows it. Misspellings are bounded by edit distance so an
// unrelated id is never offered.
func suggestStepIDs(input string) []string {
	query := strings.ToLower(strings.TrimSpace(input))
	if query == "" {
		return nil
	}

	type candidate struct {
		id   string
		rank int // kind of match; lower is better
		dist int // edit distance, for ranking misspellings against each other
		idx  int // position in AllSteps, so ties keep declaration order
	}
	var found []candidate

	for i, s := range AllSteps() {
		id := strings.ToLower(s.ID)
		switch {
		case id == query: // right id, wrong case
			found = append(found, candidate{s.ID, 0, 0, i})
		case strings.HasPrefix(id, query):
			found = append(found, candidate{s.ID, 1, 0, i})
		case len(query) >= minSubstringQuery && strings.Contains(id, query):
			found = append(found, candidate{s.ID, 2, 0, i})
		default:
			if d := editDistance(id, query); d <= typoBudget(query) {
				found = append(found, candidate{s.ID, 3, d, i})
			}
		}
	}

	sort.Slice(found, func(i, j int) bool {
		a, b := found[i], found[j]
		if a.rank != b.rank {
			return a.rank < b.rank
		}
		if a.dist != b.dist {
			return a.dist < b.dist
		}
		return a.idx < b.idx
	})

	ids := make([]string, len(found))
	for i, c := range found {
		ids[i] = c.id
	}
	return ids
}

// typoBudget is how many single-character edits an id may sit away from the
// input and still be worth offering. It scales with what was typed: one edit is
// a plausible slip in a short id, but allowing three would make every short id
// a match for every other.
func typoBudget(query string) int {
	switch n := len([]rune(query)); {
	case n <= 4:
		return 1
	case n <= 8:
		return 2
	default:
		return 3
	}
}

// editDistance returns the Levenshtein distance between a and b, counting one
// per inserted, deleted or substituted character.
func editDistance(a, b string) int {
	ar, br := []rune(a), []rune(b)
	// Only the previous row is needed to compute the next one.
	prev := make([]int, len(br)+1)
	cur := make([]int, len(br)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ar); i++ {
		cur[0] = i
		for j := 1; j <= len(br); j++ {
			sub := prev[j-1]
			if ar[i-1] != br[j-1] {
				sub++
			}
			cur[j] = min(cur[j-1]+1, min(prev[j]+1, sub))
		}
		prev, cur = cur, prev
	}
	return prev[len(br)]
}

// showStep prints a step's metadata and command(s) without executing anything.
// Colors are applied via lipgloss styles, which auto-degrade to plain text when
// stdout is not a terminal (and honor NO_COLOR).
func showStep(step Step, state *AppState) int {
	st := state.Steps[step.ID]
	fmt.Println(styleTitle.Render(step.ID))
	fmt.Printf("  Name:        %s\n", step.Name)
	fmt.Printf("  Category:    %s\n", styleCategory.Render(step.Category))
	if step.Description != "" {
		fmt.Printf("  Description: %s\n", styleDescription.Render(step.Description))
	}
	fmt.Printf("  Status:      %s\n", statusStyle(st).Render(statusLabel(st)))
	if step.RequiresAdmin {
		fmt.Printf("  Requires:    %s\n", styleAdmin.Render("admin (sudo)"))
	}
	manual := step.IsManual()
	if manual {
		fmt.Println("  Manual step — instructions:")
		for _, line := range strings.Split(step.ManualInstructions, "\n") {
			fmt.Printf("    %s\n", styleManual.Render(line))
		}
	} else {
		fmt.Println("  Command(s):")
		for _, cmd := range step.Commands {
			fmt.Printf("    %s\n", styleCommand.Render(cmd))
		}
	}

	// Contextual help: what you can do with this step.
	fmt.Println("\nActions:")
	if !manual {
		fmt.Printf("  mac-setup %s %s     run it directly in this terminal\n", step.ID, styleWarning.Render("--run"))
	}
	fmt.Printf("  mac-setup %s %s    only mark it done (no run)\n", step.ID, styleWarning.Render("--done"))
	fmt.Printf("  mac-setup %s %s   only mark it not done\n", step.ID, styleWarning.Render("--reset"))
	fmt.Printf("  mac-setup %s %s    only copy %s to clipboard\n", step.ID, styleWarning.Render("--copy"), copyTarget(step))
	return 0
}

// cliInvocation returns the command line that runs a step by itself from the
// terminal, e.g. "mac-setup tailscale --run".
func cliInvocation(step Step) string {
	return fmt.Sprintf("%s %s --run", appName, step.ID)
}

// copyTarget describes what --copy puts on the clipboard for a step.
func copyTarget(step Step) string {
	if step.IsManual() {
		return "instructions"
	}
	return "command(s)"
}

// statusLabel renders a persisted step status for display.
func statusLabel(s StepStatus) string {
	switch s {
	case StatusCompleted:
		return "completed"
	case StatusFailed:
		return "failed"
	case StatusSkipped:
		return "skipped"
	default:
		return "not run"
	}
}

// runStepDirect runs a step's commands with the terminal's real stdin/stdout/
// stderr, so interactive prompts (sudo, installers) and live output work
// natively. Returns the first command's error, if any.
func runStepDirect(step Step, dryRun bool) error {
	fmt.Printf("\n==> %s (%s)\n\n", step.Name, step.ID)
	for _, cmd := range step.Commands {
		fmt.Printf("$ %s\n", cmd)
		if dryRun {
			fmt.Println("  (dry run — skipped)")
			continue
		}
		c := exec.Command("bash", "-c", cmd)
		c.Env = append(c.Environ(), "HOMEBREW_NO_AUTO_UPDATE=1")
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			return err
		}
	}
	return nil
}

// clipboardPayload returns the text copied for a step: its commands (one per
// line) or, for a manual step, its instructions.
func clipboardPayload(step Step) string {
	if step.IsManual() {
		return step.ManualInstructions
	}
	return strings.Join(step.Commands, "\n")
}

// copyToClipboard pipes s into pbcopy (macOS).
func copyToClipboard(s string) error {
	c := exec.Command("pbcopy")
	c.Stdin = strings.NewReader(s)
	return c.Run()
}
