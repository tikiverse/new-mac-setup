package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	dryRun := flag.Bool("n", false, "dry-run: print commands instead of executing them")
	dryRunLong := flag.Bool("dry-run", false, "dry-run: print commands instead of executing them")
	flag.Parse()

	state := LoadState()

	m := newModel(state)
	m.dryRun = *dryRun || *dryRunLong

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
