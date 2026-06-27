package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	opts, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		fmt.Fprint(os.Stderr, "\n"+usage)
		os.Exit(2)
	}
	if opts.help {
		fmt.Print(usage)
		return
	}
	if opts.stepID == "" {
		// No step id given: launch the interactive TUI.
		runTUI(opts.dryRun, opts.debug)
		return
	}
	// A step id was given: run that single step directly in this terminal.
	os.Exit(runDirect(opts))
}

// runTUI launches the full-screen interactive setup.
func runTUI(dryRun, debug bool) {
	state := LoadState()
	m := newModel(state)
	m.dryRun = dryRun
	if debug {
		m.debug = true
		m.categories = visibleCategories(true)
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
