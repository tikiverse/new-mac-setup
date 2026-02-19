package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// StepStatus represents the state of a single step.
type StepStatus string

const (
	StatusPending   StepStatus = "pending"
	StatusCompleted StepStatus = "completed"
	StatusSkipped   StepStatus = "skipped"
	StatusFailed    StepStatus = "failed"
)

// AppState is persisted to disk between runs.
type AppState struct {
	Steps         map[string]StepStatus `json:"steps"`
	SelectedSteps map[string]bool       `json:"selected_steps,omitempty"`
}

func stateDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".mac-setup")
}

func statePath() string {
	return filepath.Join(stateDir(), "state.json")
}

// LoadState reads state from disk, returning a fresh state if none exists.
func LoadState() *AppState {
	data, err := os.ReadFile(statePath())
	if err != nil {
		return &AppState{Steps: make(map[string]StepStatus)}
	}
	var s AppState
	if err := json.Unmarshal(data, &s); err != nil {
		return &AppState{Steps: make(map[string]StepStatus)}
	}
	if s.Steps == nil {
		s.Steps = make(map[string]StepStatus)
	}
	return &s
}

// Save writes the state to disk.
func (s *AppState) Save() error {
	if err := os.MkdirAll(stateDir(), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(statePath(), data, 0o644)
}

// Reset clears all progress.
func (s *AppState) Reset() {
	s.Steps = make(map[string]StepStatus)
	s.SelectedSteps = nil
}
