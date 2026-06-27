package main

import "testing"

func TestStripANSI(t *testing.T) {
	cases := map[string]string{
		"\x1b[31mred\x1b[0m":          "red",
		"plain":                       "plain",
		"\x1b[1;32mbold green\x1b[0m": "bold green",
	}
	for in, want := range cases {
		if got := stripANSI(in); got != want {
			t.Fatalf("stripANSI(%q) = %q, want %q", in, got, want)
		}
	}
}

type line struct {
	text    string
	replace bool
}

func collect(t *testing.T, cmd string) ([]line, error) {
	t.Helper()
	var lines []line
	err := streamCommand(cmd, func(text string, replace bool) {
		lines = append(lines, line{text, replace})
	})
	return lines, err
}

func TestStreamCommandLines(t *testing.T) {
	lines, err := collect(t, "printf 'a\\nb\\n'")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lines) != 2 || lines[0].text != "a" || lines[1].text != "b" {
		t.Fatalf("got %+v, want a,b", lines)
	}
	if lines[0].replace || lines[1].replace {
		t.Fatalf("newline-delimited lines should not be replacements: %+v", lines)
	}
}

func TestStreamCommandCarriageReturn(t *testing.T) {
	// "x\ry\n": y overwrites x (a progress-style update).
	lines, err := collect(t, "printf 'x\\ry\\n'")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2: %+v", len(lines), lines)
	}
	if lines[0].text != "x" || lines[0].replace {
		t.Fatalf("first line = %+v, want {x false}", lines[0])
	}
	if lines[1].text != "y" || !lines[1].replace {
		t.Fatalf("second line = %+v, want {y true}", lines[1])
	}
}

func TestStreamCommandStripsANSI(t *testing.T) {
	lines, err := collect(t, "printf '\\033[31mred\\033[0m\\n'")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lines) != 1 || lines[0].text != "red" {
		t.Fatalf("got %+v, want [red]", lines)
	}
}

func TestStreamCommandExitError(t *testing.T) {
	lines, err := collect(t, "echo hi; exit 3")
	if err == nil {
		t.Fatal("expected a non-nil error for a failing command")
	}
	if len(lines) == 0 || lines[0].text != "hi" {
		t.Fatalf("expected to still capture output before failure, got %+v", lines)
	}
}
