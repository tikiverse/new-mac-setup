package main

import "github.com/charmbracelet/lipgloss"

var (
	colorGreen   = lipgloss.Color("#a6e3a1")
	colorYellow  = lipgloss.Color("#f9e2af")
	colorRed     = lipgloss.Color("#f38ba8")
	colorBlue    = lipgloss.Color("#89b4fa")
	colorMauve   = lipgloss.Color("#cba6f7")
	colorDim     = lipgloss.Color("#6c7086")
	colorText    = lipgloss.Color("#cdd6f4")
	colorSubtext = lipgloss.Color("#a6adc8")

	styleTitle = lipgloss.NewStyle().
			Foreground(colorMauve).
			Bold(true)

	styleCategory = lipgloss.NewStyle().
			Foreground(colorBlue).
			Bold(true)

	styleDescription = lipgloss.NewStyle().
				Foreground(colorSubtext)

	styleCommand = lipgloss.NewStyle().
			Foreground(colorDim)

	styleSuccess = lipgloss.NewStyle().
			Foreground(colorGreen).
			Bold(true)

	styleWarning = lipgloss.NewStyle().
			Foreground(colorYellow).
			Bold(true)

	styleError = lipgloss.NewStyle().
			Foreground(colorRed).
			Bold(true)

	styleSkipped = lipgloss.NewStyle().
			Foreground(colorDim)

	styleHelp = lipgloss.NewStyle().
			Foreground(colorDim)

	styleProgressBar = lipgloss.NewStyle().
				Foreground(colorGreen)

	styleProgressBg = lipgloss.NewStyle().
			Foreground(colorDim)

	styleManual = lipgloss.NewStyle().
			Foreground(colorYellow)

	styleAdmin = lipgloss.NewStyle().
			Foreground(colorRed).
			Bold(true)

	styleSelected = lipgloss.NewStyle().
			Foreground(colorGreen)

	styleUnselected = lipgloss.NewStyle().
			Foreground(colorText)

	styleDim = lipgloss.NewStyle().
			Foreground(colorDim)
)

// statusStyle returns the lipgloss style used to render a persisted step status.
func statusStyle(s StepStatus) lipgloss.Style {
	switch s {
	case StatusCompleted:
		return styleSuccess
	case StatusFailed:
		return styleError
	case StatusSkipped:
		return styleWarning
	default:
		return styleDim
	}
}
