package main

import (
	"bytes"
	"os/exec"
	"strings"
	"time"
)

// RunResult holds the outcome of executing a step's commands.
type RunResult struct {
	Output string
	Err    error
}

// DryRunCommands simulates execution by printing commands without running them.
func DryRunCommands(commands []string) RunResult {
	var buf bytes.Buffer
	for _, cmd := range commands {
		buf.WriteString("$ " + cmd + "\n")
		buf.WriteString("  (dry run — skipped)\n\n")
	}
	// Small delay so the spinner is visible
	time.Sleep(300 * time.Millisecond)
	return RunResult{Output: strings.TrimRight(buf.String(), "\n")}
}

// RunCommands executes a list of shell commands sequentially.
// It stops on the first failure and returns combined output.
func RunCommands(commands []string) RunResult {
	var buf bytes.Buffer

	for _, cmd := range commands {
		buf.WriteString("$ " + cmd + "\n")

		c := exec.Command("bash", "-c", cmd)
		c.Env = append(c.Environ(), "HOMEBREW_NO_AUTO_UPDATE=1")
		out, err := c.CombinedOutput()
		buf.Write(out)
		if err != nil {
			buf.WriteString("\n✗ " + err.Error() + "\n")
			return RunResult{Output: buf.String(), Err: err}
		}
		buf.WriteString("\n")
	}

	return RunResult{Output: strings.TrimRight(buf.String(), "\n")}
}
