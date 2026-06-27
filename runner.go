package main

import (
	"bufio"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// interactiveCommand builds a single shell command that runs all of a step's
// commands in sequence (&&-chained), for steps that need the real terminal —
// e.g. those that prompt for a sudo password. It inherits the terminal's
// stdin/stdout/stderr when run via tea.ExecProcess.
func interactiveCommand(step Step) *exec.Cmd {
	c := exec.Command("bash", "-c", strings.Join(step.Commands, " && "))
	c.Env = append(c.Environ(), "HOMEBREW_NO_AUTO_UPDATE=1")
	return c
}

// ansiRe matches the common ANSI escape sequences (CSI color/cursor and OSC)
// so streamed command output renders as clean text in the TUI.
var ansiRe = regexp.MustCompile("\x1b\\[[0-9;?]*[ -/]*[@-~]|\x1b\\][^\x07]*(\x07|\x1b\\\\)|\x1b[@-Z\\-_]")

func stripANSI(s string) string {
	return ansiRe.ReplaceAllString(s, "")
}

// streamCommand runs a single shell command, invoking onLine for each output
// segment as it arrives. Segments are split on \n and \r; a segment whose
// previous delimiter was a bare \r is reported with replace=true so the caller
// can overwrite the previous line (mirroring terminal carriage-return behavior).
// stdout and stderr are merged. It returns the command's exit error, if any.
func streamCommand(cmd string, onLine func(text string, replace bool)) error {
	c := exec.Command("bash", "-c", cmd)
	c.Env = append(c.Environ(), "HOMEBREW_NO_AUTO_UPDATE=1")

	pr, pw, err := os.Pipe()
	if err != nil {
		return err
	}
	c.Stdout = pw
	c.Stderr = pw

	if err := c.Start(); err != nil {
		pw.Close()
		pr.Close()
		return err
	}
	pw.Close() // close the parent's copy; the child keeps the write end

	reader := bufio.NewReader(pr)
	var buf []byte
	prevCR := false
	emit := func(replaceNext bool) {
		onLine(stripANSI(string(buf)), prevCR)
		buf = buf[:0]
		prevCR = replaceNext
	}

	for {
		b, e := reader.ReadByte()
		if e != nil {
			if len(buf) > 0 {
				onLine(stripANSI(string(buf)), prevCR)
			}
			break
		}
		switch b {
		case '\n':
			emit(false)
		case '\r':
			// Treat \r\n as a single newline rather than an overwrite.
			if nb, _ := reader.Peek(1); len(nb) == 1 && nb[0] == '\n' {
				reader.ReadByte()
				emit(false)
			} else {
				emit(true)
			}
		default:
			buf = append(buf, b)
		}
	}

	pr.Close()
	return c.Wait()
}
