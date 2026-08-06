package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
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
//
// A missing file is the normal first-run case and is not an error. Any other
// failure returns a fresh state *and* a non-nil error: callers must surface it
// rather than silently starting over, because the next Save would otherwise
// overwrite recoverable progress. An unparseable file is quarantined alongside
// the original so a later save cannot destroy it.
func LoadState() (*AppState, error) {
	fresh := func() *AppState { return &AppState{Steps: make(map[string]StepStatus)} }

	data, err := os.ReadFile(statePath())
	if os.IsNotExist(err) {
		return fresh(), nil
	}
	if err != nil {
		return fresh(), fmt.Errorf("reading %s: %w", statePath(), err)
	}

	var s AppState
	if err := json.Unmarshal(data, &s); err != nil {
		kept, keepErr := quarantine()
		if keepErr != nil {
			return fresh(), fmt.Errorf("%s is corrupt (%w); it could not be set aside: %v", statePath(), err, keepErr)
		}
		return fresh(), fmt.Errorf("%s is corrupt (%w); the damaged file was kept at %s", statePath(), err, kept)
	}

	if s.Steps == nil {
		s.Steps = make(map[string]StepStatus)
	}
	return &s, nil
}

// quarantine renames an unreadable state file out of the way and returns its
// new path, so the next Save writes a clean file without destroying evidence.
func quarantine() (string, error) {
	dst := statePath() + ".corrupt-" + time.Now().Format("20060102-150405")
	if err := os.Rename(statePath(), dst); err != nil {
		return "", err
	}
	return dst, nil
}

// Save atomically writes the state to disk. It writes a temp file in the same
// directory and renames it into place, so an interrupted save leaves the
// previous state intact rather than a half-written file.
func (s *AppState) Save() error {
	dir := stateDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, "state-*.json.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op once the rename below succeeds

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	// Flush to disk before the rename so a crash can't leave an empty file
	// under the real name.
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	// CreateTemp makes the file 0600; match the previous WriteFile mode.
	if err := os.Chmod(tmpName, 0o644); err != nil {
		return err
	}
	return os.Rename(tmpName, statePath())
}

// Reset clears all progress.
func (s *AppState) Reset() {
	s.Steps = make(map[string]StepStatus)
	s.SelectedSteps = nil
}
