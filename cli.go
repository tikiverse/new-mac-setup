package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const usage = `mac-setup — a-la-carte Mac setup

Usage:
  mac-setup                      Launch the interactive TUI
  mac-setup --dry-run            Launch the TUI in dry-run mode
  mac-setup <step-id>            Run a single step directly in this terminal
  mac-setup <step-id> --done     Mark the step as done (no run)
  mac-setup <step-id> --undone   Mark the step as not done
  mac-setup <step-id> --copy     Copy the step's command(s) to the clipboard
  mac-setup <step-id> --dry-run  Print the step's command(s) without running

Flags:
  -n, --dry-run   Print commands instead of executing them
  -h, --help      Show this help

Step ids are shown next to each step in the TUI.
`

type cliAction int

const (
	actionRun cliAction = iota
	actionDone
	actionUndone
	actionCopy
)

type cliOptions struct {
	stepID string
	action cliAction
	dryRun bool
	help   bool
}

// parseArgs parses CLI arguments. Flags may appear before or after the step id.
func parseArgs(args []string) (cliOptions, error) {
	var o cliOptions
	actionSet := false
	setAction := func(a cliAction) error {
		if actionSet {
			return fmt.Errorf("only one of --done, --undone, --copy may be given")
		}
		o.action = a
		actionSet = true
		return nil
	}

	for _, a := range args {
		switch a {
		case "--done":
			if err := setAction(actionDone); err != nil {
				return o, err
			}
		case "--undone":
			if err := setAction(actionUndone); err != nil {
				return o, err
			}
		case "--copy":
			if err := setAction(actionCopy); err != nil {
				return o, err
			}
		case "-n", "--dry-run":
			o.dryRun = true
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
	return o, nil
}

// runDirect performs a single-step action from the CLI and returns an exit code.
func runDirect(opts cliOptions) int {
	step, ok := StepByID(opts.stepID)
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown step id: %q\n", opts.stepID)
		fmt.Fprintln(os.Stderr, "Run mac-setup with no arguments to browse step ids.")
		return 1
	}
	state := LoadState()

	switch opts.action {
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

	case actionUndone:
		delete(state.Steps, step.ID)
		if err := state.Save(); err != nil {
			fmt.Fprintf(os.Stderr, "Save failed: %v\n", err)
			return 1
		}
		fmt.Printf("Marked %s as not done.\n", step.ID)
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
	state.Steps[step.ID] = StatusCompleted
	if err := state.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "Save failed: %v\n", err)
		return 1
	}
	fmt.Printf("\n✓ %s done.\n", step.ID)
	return 0
}

// runStepDirect runs a step's commands with the terminal's real stdin/stdout/
// stderr, so interactive prompts (sudo, installers) and live output work
// natively. Returns the first command's error, if any.
func runStepDirect(step Step, dryRun bool) error {
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
