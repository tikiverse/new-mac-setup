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

### Running a single step from the CLI

Every step has an **id** (shown next to it in the TUI). Pass an id to act on
just that step directly in your terminal — handy for one-offs, and because the
command runs with the real terminal, interactive prompts (`sudo`, installers)
and live output work natively:

```
./mac-setup finder-path-bar            # run just this step in the terminal
./mac-setup finder-path-bar --dry-run  # print its command(s) without running
./mac-setup finder-path-bar --done     # mark as done (no run)
./mac-setup finder-path-bar --undone   # mark as not done
./mac-setup finder-path-bar --copy     # copy its command(s) to the clipboard
```

A direct run marks the step done on success (or failed on error); manual,
instruction-only steps are printed rather than executed.

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

When a step fails mid-run, the run pauses and shows the captured output so you can decide what to do:

| Key | Action |
|-----|--------|
| `r` | Retry the failed step |
| `s` | Skip it and continue |
| `a` | Abort the rest of the run |
