# mac-setup

Interactive TUI for bootstrapping a fresh Mac, a-la carte.

Pick categories/steps from a menu card, only picking what you decide — progress is saved to `~/.mac-setup/state.json`.

## What it sets up

System prefs, Homebrew, browsers, dev tools (Node, Docker, CLI utilities), Finder tweaks, window management, media apps, Chrome extensions, and more. See `steps.go` for the full list.

## Usage

```
./mac-setup            # interactive setup
./mac-setup --dry-run  # dry-run (prints commands, doesn't execute)
./mac-setup -n         # same, shorthand
go run .               # run directly from source
```

## Controls

| Key | Action |
|-----|--------|
| `↑/↓` or `j/k` | Navigate |
| `Enter` or `Space` | Drill into category |
| `Space` | Toggle step selection |
| `G` | Start running selected steps |
| `Backspace` or `Esc` | Back |
| `R` | Reset step state |
| `q` | Quit (progress saved) |
