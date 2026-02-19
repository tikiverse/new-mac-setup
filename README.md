# mac-setup

Interactive TUI for bootstrapping a fresh Mac. Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

Pick categories/steps, run or skip each one, quit anytime — progress is saved to `~/.mac-setup/state.json`.

## What it sets up

System prefs, Homebrew, browsers, dev tools (Node, Docker, CLI utilities), Finder tweaks, window management, media apps, Chrome extensions, and more. See `steps.go` for the full list.

## Usage

```
go run .            # interactive setup
go run . -n         # dry-run (prints commands, doesn't execute)
go run . --dry-run  # same
```

## Controls

| Key | Action |
|-----|--------|
| `j/k` | Navigate |
| `Space` | Toggle selection |
| `Enter` | Drill into category / run step |
| `G` | Start running selected steps |
| `s` | Skip step |
| `b` | Back |
| `q` | Quit (progress saved) |
