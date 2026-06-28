package main

// Step represents a single setup task.
type Step struct {
	ID                 string
	Category           string
	Name               string
	Description        string
	Commands           []string
	ManualInstructions string
	// Note is shown as a one-time acknowledgement after a command step
	// succeeds, for caveats the commands themselves can't convey (e.g. a
	// setting that requires logging out before it takes effect).
	Note          string
	RequiresAdmin bool
	Debug              bool // hidden from the TUI unless launched with --debug
}

// AllSteps returns the full ordered list of setup steps derived from the notion export.
func AllSteps() []Step {
	steps := []Step{
		// ── System Preferences ──────────────────────────────────────────
		{
			ID:          "key-repeat",
			Category:    "System Preferences",
			Name:        "Fast key repeat rate",
			Description: "Set key repeat to fastest setting (1) and shorten the delay before it starts.",
			Commands: []string{
				`defaults write NSGlobalDomain KeyRepeat -int 1`,
				`defaults write NSGlobalDomain InitialKeyRepeat -int 25`,
			},
			Note: "This is read at login — log out and back in for it to take effect.",
		},
		{
			ID:          "press-and-hold",
			Category:    "System Preferences",
			Name:        "Disable press-and-hold for accents",
			Description: "Enable key repeat instead of the accent character popup when holding a key.",
			Commands:    []string{`defaults write NSGlobalDomain ApplePressAndHoldEnabled -bool false`},
			Note:        "This is read at login — log out and back in for it to take effect.",
		},
		{
			ID:          "dock-active-only",
			Category:    "System Preferences",
			Name:        "Dock: show only active apps",
			Description: "Clear persistent dock icons so only running apps appear.",
			Commands: []string{
				`defaults write com.apple.dock persistent-apps -array '()'`,
				`killall Dock`,
			},
		},
		{
			ID:          "mission-control",
			Category:    "System Preferences",
			Name:        "Mission Control: disable auto-rearrange",
			Description: "Prevent Spaces from reordering based on recent use.",
			Commands:    []string{`defaults write com.apple.dock mru-spaces -int 0`},
		},
		{
			ID:          "expand-save-panel",
			Category:    "System Preferences",
			Name:        "Expand save panels by default",
			Description: "Always show the full save dialog instead of the compact one.",
			Commands: []string{
				`defaults write NSGlobalDomain NSNavPanelExpandedStateForSaveMode -bool true`,
				`defaults write NSGlobalDomain NSNavPanelExpandedStateForSaveMode2 -bool true`,
			},
		},
		{
			ID:          "printer-quit",
			Category:    "System Preferences",
			Name:        "Auto-quit printer app",
			Description: "Automatically quit the printer app once print jobs complete.",
			Commands:    []string{`defaults write com.apple.print.PrintingPrefs "Quit When Finished" -bool true`},
		},
		{
			ID:          "hide-desktop-icons",
			Category:    "System Preferences",
			Name:        "Hide desktop icons",
			Description: "Hide all icons on the desktop for a cleaner look.",
			Commands:    []string{`defaults write com.apple.finder CreateDesktop -bool false`},
		},
		{
			ID:          "accessibility-zoom",
			Category:    "System Preferences",
			Name:        "Enable Ctrl+scroll zoom",
			Description: "Use scroll gesture with Ctrl modifier to zoom the screen.",
			ManualInstructions: "Go to System Settings → Accessibility → Zoom\n" +
				"Enable 'Use scroll gesture with modifier keys to zoom'\n" +
				"Set modifier to ^ Control.",
		},

		// ── Homebrew & Terminal ─────────────────────────────────────────
		{
			ID:            "homebrew-install",
			Category:      "Homebrew & Terminal",
			Name:          "Install Homebrew",
			Description:   "Install Homebrew package manager (also installs Xcode CLI tools).",
			Commands:      []string{`/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`},
			RequiresAdmin: true,
		},
		{
			ID:          "homebrew-path",
			Category:    "Homebrew & Terminal",
			Name:        "Add Homebrew to PATH",
			Description: "Add brew shellenv to .zprofile so brew is available in new shells.",
			Commands: []string{
				`echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> ~/.zprofile`,
				`eval "$(/opt/homebrew/bin/brew shellenv)"`,
			},
		},
		{
			ID:          "homebrew-config",
			Category:    "Homebrew & Terminal",
			Name:        "Configure Homebrew",
			Description: "Update, upgrade, and disable analytics.",
			Commands: []string{
				`brew update`,
				`brew upgrade`,
				`brew analytics off`,
			},
		},
		{
			ID:          "mas-install",
			Category:    "Homebrew & Terminal",
			Name:        "Install mas (Mac App Store CLI)",
			Description: "Core prerequisite for Mac App Store installs (e.g. Things 3).",
			Commands:    []string{`brew install mas`},
		},

		// ── Browser ────────────────────────────────────────────────────
		{
			ID:          "chrome-install",
			Category:    "Browser",
			Name:        "Install Google Chrome",
			Description: "Install Chrome via Homebrew cask.",
			Commands:    []string{`brew install --cask google-chrome`},
		},
		{
			ID:          "1password-install",
			Category:    "Browser",
			Name:        "Install 1Password",
			Description: "Install 1Password via Homebrew cask.",
			Commands:    []string{`brew install --cask 1password`},
		},
		{
			ID:          "1password-setup",
			Category:    "Browser",
			Name:        "Set up 1Password",
			Description: "Log in to 1Password and install the Chrome extension.",
			ManualInstructions: "1. Open 1Password and sign in to your account\n" +
				"2. Install the 1Password Chrome extension:\n" +
				"   https://chrome.google.com/webstore/detail/1password/aeblfdkhhhdcdjpifhhbdiojplfjncoa\n" +
				"3. Enable in incognito mode: chrome://extensions/",
		},

		// ── Workflow Apps ──────────────────────────────────────────────
		{
			ID:          "alfred-install",
			Category:    "Workflow Apps",
			Name:        "Install Alfred",
			Description: "Spotlight replacement and productivity launcher.",
			Commands:    []string{`brew install --cask alfred`},
		},
		{
			ID:          "notion-install",
			Category:    "Workflow Apps",
			Name:        "Install Notion",
			Description: "Note-taking and workspace app.",
			Commands:    []string{`brew install --cask notion`},
		},
		{
			ID:          "vscode-install",
			Category:    "Workflow Apps",
			Name:        "Install Visual Studio Code",
			Description: "Code editor.",
			Commands:    []string{`brew install --cask visual-studio-code`},
		},
		{
			ID:          "iterm2-install",
			Category:    "Workflow Apps",
			Name:        "Install iTerm2",
			Description: "Terminal emulator replacement.",
			Commands:    []string{`brew install --cask iterm2`},
		},
		{
			ID:          "iterm2-setup",
			Category:    "Workflow Apps",
			Name:        "Configure iTerm2",
			Description: "Set iTerm2 appearance preferences.",
			ManualInstructions: "Open iTerm2 → Preferences:\n" +
				"  • Appearance → Theme: Minimal\n" +
				"  • Profiles → Colors: Pastel (Dark)\n" +
				"  • Set background to #1b1f22\n" +
				"  • Set blue to #0dc8ff",
		},
		{
			ID:            "amphetamine-install",
			Category:      "Workflow Apps",
			Name:          "Install Amphetamine",
			Description:   "Keep-awake utility from the Mac App Store (mas install needs admin + App Store sign-in).",
			Commands:      []string{`mas install 937984704`},
			RequiresAdmin: true,
		},
		{
			ID:          "rectangle-install",
			Category:    "Workflow Apps",
			Name:        "Install Rectangle",
			Description: "Window management with keyboard shortcuts.",
			Commands:    []string{`brew install --cask rectangle`},
		},
		{
			ID:          "rectangle-setup",
			Category:    "Workflow Apps",
			Name:        "Configure Rectangle",
			Description: "Grant permissions and enable launch at login.",
			ManualInstructions: "Open Rectangle:\n" +
				"  • Grant Accessibility permission when prompted\n" +
				"  • Enable 'Launch at Login'",
		},
		{
			ID:          "fantastical-install",
			Category:    "Workflow Apps",
			Name:        "Install Fantastical",
			Description: "Calendar app via Homebrew cask.",
			Commands:    []string{`brew install --cask fantastical`},
		},
		{
			ID:            "things-install",
			Category:      "Workflow Apps",
			Name:          "Install Things 3",
			Description:   "Task manager from the Mac App Store (mas install needs admin + App Store sign-in).",
			Commands:      []string{`mas install 904280696`},
			RequiresAdmin: true,
		},

		// ── Finder Settings ────────────────────────────────────────────
		{
			ID:          "finder-extensions",
			Category:    "Finder Settings",
			Name:        "Show all filename extensions",
			Description: "Always display file extensions in Finder.",
			Commands:    []string{`defaults write NSGlobalDomain AppleShowAllExtensions -bool true`},
		},
		{
			ID:          "finder-status-bar",
			Category:    "Finder Settings",
			Name:        "Show Finder status bar",
			Description: "Display the status bar at the bottom of Finder windows.",
			Commands:    []string{`defaults write com.apple.finder ShowStatusBar -bool true`},
		},
		{
			ID:          "finder-path-bar",
			Category:    "Finder Settings",
			Name:        "Show Finder path bar",
			Description: "Display the path bar at the bottom of Finder windows.",
			Commands:    []string{`defaults write com.apple.finder ShowPathbar -bool true`},
		},
		{
			ID:          "finder-library",
			Category:    "Finder Settings",
			Name:        "Show ~/Library folder",
			Description: "Unhide the Library folder in your home directory.",
			Commands:    []string{`chflags nohidden ~/Library`},
		},
		{
			ID:          "finder-default-home",
			Category:    "Finder Settings",
			Name:        "Set Finder default to Home",
			Description: "New Finder windows open to your home directory.",
			Commands: []string{
				`defaults write com.apple.finder NewWindowTarget PfHm`,
				`killall Finder`,
			},
		},
		{
			ID:          "finder-search-scope",
			Category:    "Finder Settings",
			Name:        "Search current folder by default",
			Description: "Finder searches the current folder instead of the whole Mac.",
			Commands:    []string{`defaults write com.apple.finder FXDefaultSearchScope -string "SCcf"`},
		},
		{
			ID:          "snap-to-grid",
			Category:    "Finder Settings",
			Name:        "Enable snap-to-grid for icons",
			Description: "Icons snap to a grid on the desktop and in icon views.",
			Commands: []string{
				`/usr/libexec/PlistBuddy -c "Set :DesktopViewSettings:IconViewSettings:arrangeBy grid" ~/Library/Preferences/com.apple.finder.plist`,
				`/usr/libexec/PlistBuddy -c "Set :FK_StandardViewSettings:IconViewSettings:arrangeBy grid" ~/Library/Preferences/com.apple.finder.plist`,
				`/usr/libexec/PlistBuddy -c "Set :StandardViewSettings:IconViewSettings:arrangeBy grid" ~/Library/Preferences/com.apple.finder.plist`,
			},
		},

		// ── Media Apps ─────────────────────────────────────────────────
		{
			ID:          "spotify-install",
			Category:    "Media Apps",
			Name:        "Install Spotify",
			Description: "Music streaming app.",
			Commands:    []string{`brew install --cask spotify`},
		},
		{
			ID:          "vlc-install",
			Category:    "Media Apps",
			Name:        "Install VLC",
			Description: "Universal media player.",
			Commands:    []string{`brew install --cask vlc`},
		},
		{
			ID:          "ffmpeg-install",
			Category:    "Media Apps",
			Name:        "Install ffmpeg",
			Description: "CLI tool for video/audio conversion.",
			Commands:    []string{`brew install ffmpeg`},
		},
		{
			ID:          "yt-dlp-install",
			Category:    "Media Apps",
			Name:        "Install yt-dlp",
			Description: "Download videos from YouTube and other sites (maintained youtube-dl fork).",
			Commands:    []string{`brew install yt-dlp`},
		},
		{
			ID:          "flux-install",
			Category:    "Media Apps",
			Name:        "Install f.lux",
			Description: "Adjusts screen color temperature at night.",
			Commands:    []string{`brew install --cask flux`},
		},
		{
			ID:            "zoom-install",
			Category:      "Media Apps",
			Name:          "Install Zoom",
			Description:   "Video conferencing app (cask uses a pkg installer, so it needs admin).",
			Commands:      []string{`brew install --cask zoom`},
			RequiresAdmin: true,
		},

		// ── Development ────────────────────────────────────────────────
		{
			ID:          "n-install",
			Category:    "Development",
			Name:        "Install Node.js via n",
			Description: "Install n version manager and latest Node.js.",
			Commands: []string{
				`curl -L https://bit.ly/n-install | bash`,
			},
		},
		{
			ID:          "docker-install",
			Category:    "Development",
			Name:        "Install Docker",
			Description: "Container runtime.",
			Commands:    []string{`brew install docker`},
		},
		{
			ID:          "brew-formulae",
			Category:    "Development",
			Name:        "Install CLI tools (brew formulae)",
			Description: "gh, fzf, ripgrep, jq, neovim, tmux, tree, httpie, tldr, mosh, pnpm, gron, just, llm, mcfly, slides, wifi-password, fastfetch, zoxide.",
			Commands: []string{
				`brew install gh fzf ripgrep jq neovim tmux tree httpie tldr mosh pnpm gron just llm mcfly slides wifi-password fastfetch zoxide`,
			},
		},

		// ── Shell Setup ────────────────────────────────────────────────
		{
			ID:          "antidote-install",
			Category:    "Shell Setup",
			Name:        "Install antidote",
			Description: "Zsh plugin manager.",
			Commands:    []string{`brew install antidote`},
		},

		// ── Keyboard ───────────────────────────────────────────────────
		{
			ID:          "hyperkey-install",
			Category:    "Keyboard",
			Name:        "Install Hyperkey",
			Description: "Caps Lock → Left Control, quick press = Escape; Hyper key on F4.",
			ManualInstructions: "1. Download and install from https://hyperkey.app/\n" +
				"2. Set the Hyper Key to F4\n" +
				"3. Remap Caps Lock → Left Control\n" +
				"4. Set Caps Lock quick press to Escape",
		},

		// ── Chrome Extensions ──────────────────────────────────────────
		{
			ID:          "chrome-extensions",
			Category:    "Chrome Extensions",
			Name:        "Install Chrome extensions",
			Description: "Manually install recommended Chrome extensions.",
			ManualInstructions: "Install these Chrome extensions:\n\n" +
				"  • uBlock Origin Lite — https://chromewebstore.google.com/detail/ublock-origin-lite/ddkjiahejlhfcafbddmgiahcphecmpfh\n" +
				"  • Vimium — https://chromewebstore.google.com/detail/vimium/dbepggeogbaibhgnhhndojpepiihcmeb\n" +
				"  • Old Reddit Redirect — https://chrome.google.com/webstore/detail/old-reddit-redirect/dneaehbmnbhcippjikoajpoabadpodje\n" +
				"  • Reddit Enhancement Suite — https://chrome.google.com/webstore/detail/reddit-enhancement-suite/kbmfpngjjgdllneeigpgjifpgocmfgmb\n" +
				"  • Instapaper — https://chrome.google.com/webstore/detail/instapaper/ldjkgaaoikpmhmkelcgkgacicjfbofhh\n" +
				"  • YouTube Playback Speed Control — https://chrome.google.com/webstore/detail/youtube-playback-speed-co/hdannnflhlmdablckfkjpleikpphncik\n" +
				"  • Also: Cold Turkey Blocker, Loom, OneTab, Readwise Highlighter, React DevTools",
		},
		{
			ID:          "chrome-flags",
			Category:    "Chrome Extensions",
			Name:        "Configure Chrome flags & settings",
			Description: "Disable hardware media key handling so Spotify isn't interrupted by YouTube.",
			ManualInstructions: "1. Open chrome://flags/#hardware-media-key-handling → Disable\n" +
				"2. To prevent Cmd+Shift+I opening Mail instead of DevTools:\n" +
				"   https://apple.stackexchange.com/a/108129",
		},

		// ── Additional Tweaks ──────────────────────────────────────────
		{
			ID:          "screenshots-dir",
			Category:    "Additional Tweaks",
			Name:        "Change screenshots directory",
			Description: "Save screenshots to ~/Screenshots instead of Desktop.",
			Commands: []string{
				`mkdir -p ~/Screenshots`,
				`defaults write com.apple.screencapture location -string "${HOME}/Screenshots"`,
			},
		},
		{
			ID:          "soundsource-install",
			Category:    "Additional Tweaks",
			Name:        "Install SoundSource",
			Description: "Advanced audio control for Mac.",
			Commands:    []string{`brew install --cask soundsource`},
		},

		// ── Testing ────────────────────────────────────────────────────
		// No-op steps with no side effects, used to exercise the run UI:
		// streaming, delays, progress bars, ANSI, failures, and so on.
		{
			ID:          "test-quick",
			Category:    "Testing",
			Name:        "Quick success",
			Description: "Prints two lines and exits 0. Baseline happy path.",
			Commands:    []string{`echo "starting…"; echo "done."`},
		},
		{
			ID:          "test-stream",
			Category:    "Testing",
			Name:        "Streaming output (drip)",
			Description: "Emits 10 lines, one every 0.3s, so you can watch them stream in.",
			Commands:    []string{`for i in $(seq 1 10); do echo "line $i of 10"; sleep 0.3; done`},
		},
		{
			ID:          "test-slow-start",
			Category:    "Testing",
			Name:        "Sleep 3s before output",
			Description: "Prints a line, sleeps 3s, then prints again (tests the waiting state).",
			Commands:    []string{`echo "working…"; sleep 3; echo "finished after 3s"`},
		},
		{
			ID:          "test-progress",
			Category:    "Testing",
			Name:        "Animated progress bar",
			Description: "A \\r-updated progress bar (tests carriage-return overwrite handling).",
			Commands:    []string{`for i in $(seq 1 20); do printf "\r[%-20s] %d%%" "$(printf '#%.0s' $(seq 1 $i))" $((i*5)); sleep 0.1; done; echo`},
		},
		{
			ID:          "test-spinner",
			Category:    "Testing",
			Name:        "Spinner animation",
			Description: "A \\r-updated spinner (more carriage-return overwrite testing).",
			Commands:    []string{`for n in $(seq 1 20); do for c in '|' '/' '-' '\'; do printf "\rworking %s" "$c"; sleep 0.08; done; done; printf "\rdone   \n"`},
		},
		{
			ID:          "test-colors",
			Category:    "Testing",
			Name:        "ANSI color output",
			Description: "Prints colored/bold text (tests ANSI stripping in the viewport).",
			Commands:    []string{`printf '\033[31mred\033[0m \033[32mgreen\033[0m \033[33myellow\033[0m \033[1mbold\033[0m\n'`},
		},
		{
			ID:          "test-stderr",
			Category:    "Testing",
			Name:        "Interleaved stdout + stderr",
			Description: "Writes to both streams (tests that they merge in order).",
			Commands:    []string{`echo "to stdout"; echo "to stderr" 1>&2; echo "more stdout"`},
		},
		{
			ID:          "test-burst",
			Category:    "Testing",
			Name:        "Burst of fast output",
			Description: "Prints 200 lines as fast as possible (tests throughput and scrollback).",
			Commands:    []string{`for i in $(seq 1 200); do echo "burst line $i"; done`},
		},
		{
			ID:          "test-200-lines-success",
			Category:    "Testing",
			Name:        "200 lines, then succeed",
			Description: "Prints 200 numbered lines, then exits successfully (tests long streamed output with vertical overflow and a clean done state).",
			Commands:    []string{`for i in $(seq 1 200); do echo "stream line $i of 200"; done`},
		},
		{
			ID:          "test-200-lines-fail",
			Category:    "Testing",
			Name:        "200 lines, then fail",
			Description: "Prints 200 numbered lines, then exits 1 (tests long streamed output preserved on failure after vertical overflow).",
			Commands:    []string{`for i in $(seq 1 200); do echo "stream line $i of 200"; done; echo "failing after 200 lines" 1>&2; exit 1`},
		},
		{
			ID:          "test-multi",
			Category:    "Testing",
			Name:        "Multiple commands",
			Description: "Three commands in sequence (tests the per-command $ headers).",
			Commands: []string{
				`echo "first command"`,
				`echo "second command"; sleep 1`,
				`echo "third command"`,
			},
		},
		{
			ID:          "test-interactive",
			Category:    "Testing",
			Name:        "Interactive prompt (needs stdin)",
			Description: "Runs `read` for a y/N answer — shows what happens with no TTY stdin.",
			Commands:    []string{`read -p "Continue? [y/N] " ans; echo "you answered: ${ans:-<no input>}"`},
		},
		{
			ID:          "test-fail",
			Category:    "Testing",
			Name:        "Always fails",
			Description: "Prints a line then exits 1 (tests the failure pause / retry / skip / abort).",
			Commands:    []string{`echo "about to fail…"; exit 1`},
		},
		{
			ID:          "test-fail-midway",
			Category:    "Testing",
			Name:        "Output then fail",
			Description: "Emits several lines, then fails (tests output capture on failure).",
			Commands:    []string{`echo "step 1 ok"; echo "step 2 ok"; echo "step 3 failing" 1>&2; exit 2`},
		},
		{
			ID:          "test-long-line",
			Category:    "Testing",
			Name:        "Very long single line",
			Description: "Prints a 300-char line (tests how the viewport wraps wide output).",
			Commands:    []string{`printf 'x%.0s' $(seq 1 300); echo`},
		},
		{
			ID:          "test-no-newline",
			Category:    "Testing",
			Name:        "Partial line, no trailing newline",
			Description: "Prints text with no final newline (tests flushing the last buffer at EOF).",
			Commands:    []string{`printf 'no trailing newline here'`},
		},
		{
			ID:          "test-no-output",
			Category:    "Testing",
			Name:        "No output, succeeds",
			Description: "Runs `true` with no output (edge case: empty viewport then done).",
			Commands:    []string{`true`},
		},
	}

	// The Testing category is a developer aid: its steps run no-ops to exercise
	// the run UI and are hidden from the TUI unless launched with --debug.
	for i := range steps {
		if steps[i].Category == "Testing" {
			steps[i].Debug = true
		}
	}
	return steps
}

// Categories returns the unique category names in order.
func Categories() []string {
	return visibleCategories(true)
}

// visibleCategories returns the unique category names in order, omitting
// categories made up solely of debug steps unless includeDebug is true.
func visibleCategories(includeDebug bool) []string {
	seen := map[string]bool{}
	var cats []string
	for _, s := range AllSteps() {
		if s.Debug && !includeDebug {
			continue
		}
		if !seen[s.Category] {
			seen[s.Category] = true
			cats = append(cats, s.Category)
		}
	}
	return cats
}

// StepByID returns the step with the given ID and whether it was found.
func StepByID(id string) (Step, bool) {
	for _, s := range AllSteps() {
		if s.ID == id {
			return s, true
		}
	}
	return Step{}, false
}

// IsManual reports whether the step is a manual (instruction-only) step.
func (s Step) IsManual() bool {
	return s.ManualInstructions != "" && len(s.Commands) == 0
}
