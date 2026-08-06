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
	Debug         bool // hidden from the TUI unless launched with --debug
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
		{
			ID:          "screenshots-dir",
			Category:    "System Preferences",
			Name:        "Change screenshots directory",
			Description: "Save screenshots to ~/Screenshots instead of Desktop.",
			Commands: []string{
				`mkdir -p ~/Screenshots`,
				`defaults write com.apple.screencapture location -string "${HOME}/Screenshots"`,
			},
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
			ID:          "finder-show-hidden",
			Category:    "Finder Settings",
			Name:        "Show hidden files and folders",
			Description: "Display dotfiles and other hidden items in Finder.",
			Commands: []string{
				`defaults write com.apple.finder AppleShowAllFiles -bool true`,
				`killall Finder`,
			},
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
		{
			ID:          "soundsource-install",
			Category:    "Media Apps",
			Name:        "Install SoundSource",
			Description: "Advanced audio control for Mac.",
			Commands:    []string{`brew install --cask soundsource`},
		},

		// ── Development ────────────────────────────────────────────────
		{
			ID:          "n-install",
			Category:    "Development",
			Name:        "Install Node.js via n",
			Description: "Install n version manager and latest Node.js.",
			Commands: []string{
				// -y auto-confirms the install prompt, which n-install otherwise
				// reads from /dev/tty — that prompt is unreachable (and the step
				// hangs/fails) when commands run through the app's captured,
				// non-interactive output pipe instead of a real terminal.
				`curl -L https://bit.ly/n-install | bash -s -- -y`,
			},
		},
		{
			ID:          "orbstack-install",
			Category:    "Development",
			Name:        "Install OrbStack",
			Description: "Lightweight, fast Docker-compatible container runtime.",
			Commands:    []string{`brew install --cask orbstack`},
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
		{
			ID:          "1password-cli-install",
			Category:    "Development",
			Name:        "Install 1Password CLI",
			Description: "op — read secrets from 1Password in scripts and the shell.",
			Commands:    []string{`brew install --cask 1password-cli`},
			Note:        "op needs the desktop app to unlock it: in 1Password, Settings → Developer → 'Integrate with 1Password CLI'. Until that is ticked, every op command prompts for your password.",
		},
		{
			ID:          "uv-install",
			Category:    "Development",
			Name:        "Install uv",
			Description: "Fast Python package installer and resolver.",
			Commands:    []string{`brew install uv`},
		},

		// ── Shell Setup ────────────────────────────────────────────────
		{
			ID:          "antidote-install",
			Category:    "Shell Setup",
			Name:        "Install antidote",
			Description: "Zsh plugin manager.",
			Commands:    []string{`brew install antidote`},
		},
		{
			ID:          "blexmono-nerd-font-install",
			Category:    "Shell Setup",
			Name:        "Install BlexMono Nerd Font",
			Description: "Patched monospace font with glyphs/icons for terminal prompts.",
			Commands:    []string{`brew install --cask font-blex-mono-nerd-font`},
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

		// ── Private Network ────────────────────────────────────────────
		{
			ID:          "tailscale-install",
			Category:    "Private Network",
			Name:        "Install Tailscale",
			Description: "Mesh VPN for accessing your other devices.",
			Commands:    []string{`brew install --cask tailscale`},
		},
		{
			ID:          "syncthing-install",
			Category:    "Private Network",
			Name:        "Install Syncthing",
			Description: "Continuous file sync between your devices; starts as a background service.",
			Commands: []string{
				`brew install syncthing`,
				`brew services start syncthing`,
			},
			Note: "Never copy key.pem from another machine — each device generates its own identity, and duplicate identities break the mesh.",
		},
		{
			ID:          "syncthing-setup",
			Category:    "Private Network",
			Name:        "Pair Syncthing with the always-on device",
			Description: "Introduce this Mac to the always-on device over Tailscale.",
			// One logical line per item — the run view wraps to the terminal.
			ManualInstructions: "1. Open http://localhost:8384\n" +
				"2. Add Remote Device → paste the always-on device's ID (kept as a secure note in 1Password, deliberately not in this repo)\n" +
				"3. Advanced tab → Addresses: tcp://<tailscale-id>:22000 instead of 'dynamic' (needs Tailscale connected)\n" +
				"4. On that device's Syncthing GUI, accept the prompt from this Mac",
		},
		{
			ID:          "syncthing-folder-configure",
			Category:    "Private Network",
			Name:        "Configure a Syncthing folder",
			Description: "Add a shared folder and pin the far side to a Tailscale address.",
			ManualInstructions: "1. In http://localhost:8384, click Add Folder\n" +
				"     Folder Label: pantry\n" +
				"     Folder ID:    pantry\n" +
				"     Folder Path:  ~/pantry (unless you keep it elsewhere)\n" +
				"2. Sharing tab → tick the always-on device, then Save\n" +
				"3. On the always-on device: Remote Devices → this Mac → Edit → Advanced tab → Addresses: tcp://<tailscale-id>:22000 instead of 'dynamic', using this Mac's Tailscale ID\n" +
				"4. Accept the folder share when it appears on the always-on device",
		},

		// ── Backup ─────────────────────────────────────────────────────
		{
			ID:          "restic-install",
			Category:    "Backup",
			Name:        "Install restic",
			Description: "Encrypted, deduplicated snapshot backups.",
			Commands:    []string{`brew install restic`},
			Note: "restic keeps no config of its own — a repository is created with `restic init -r <repo>` and unlocked only by the password you choose. " +
				"Store that password somewhere you can reach from a dead machine (1Password): without it the backups are unrecoverable.",
		},
		{
			ID:          "rclone-install",
			Category:    "Backup",
			Name:        "Install rclone",
			Description: "Move files to and from cloud storage remotes (also a restic backend).",
			Commands:    []string{`brew install rclone`},
			Note:        "Remotes are added interactively with `rclone config`; the credentials it writes live in ~/.config/rclone/rclone.conf, so treat that file as a secret.",
		},
		{
			ID:          "rsync-install",
			Category:    "Backup",
			Name:        "Install rsync (Homebrew)",
			Description: "macOS ships rsync 2.6.9 or openrsync; neither supports --append-verify.",
			Commands: []string{
				`brew install rsync`,
				// Probe the flag rather than parsing --version: the system binary is
				// rsync 2.6.9 on some macOS releases and openrsync on others, and the
				// two report versions differently. Either way this exits non-zero and
				// fails the step if the wrong rsync is on the path.
				`/opt/homebrew/bin/rsync --append-verify --version >/dev/null && echo "✓ /opt/homebrew/bin/rsync supports --append-verify"`,
			},
			Note: "Call it by full path — /opt/homebrew/bin/rsync — in scripts and LaunchAgents. " +
				"launchd runs agents with a minimal PATH (/usr/bin:/bin:/usr/sbin:/sbin) that excludes /opt/homebrew/bin, " +
				"so a bare `rsync` there silently resolves to the system one and quietly stops sending only the appended tail.",
		},

		// ── Supply Chain ───────────────────────────────────────────────
		// A compromised package release is usually caught and pulled within
		// days, so refusing versions younger than nine days means the community
		// does the smoke-testing instead of this machine. Each package manager
		// spells the same 9-day window differently.
		{
			ID:          "npm-cooldown",
			Category:    "Supply Chain",
			Name:        "npm: 9-day dependency cooldown",
			Description: "Refuse npm versions published in the last 9 days.",
			Commands: []string{
				`npm config set min-release-age 9 --location=user`,
				`npm config get min-release-age`,
			},
			Note: "npm counts in days. The setting lands in ~/.npmrc and covers transitive dependencies too. " +
				"Precedence is cli > env > project > user > global, so a single command can still opt out when you genuinely need a fresh release.",
		},
		{
			ID:          "pnpm-cooldown",
			Category:    "Supply Chain",
			Name:        "pnpm: 9-day dependency cooldown",
			Description: "Raise pnpm's minimum release age from its 1-day default to 9 days.",
			Commands: []string{
				// pnpm refuses --global config writes while its global bin directory
				// is missing from PATH, which is the state of any Mac where
				// `pnpm setup` has not been run. Creating that directory and putting
				// it on PATH for this one command is enough to get the write through.
				`mkdir -p "$HOME/Library/pnpm/bin" && PATH="$HOME/Library/pnpm/bin:$PATH" pnpm config set minimumReleaseAge 12960 --global`,
				`pnpm config get minimumReleaseAge`,
			},
			Note: "pnpm counts in minutes, so 12960 = 9 days; since v11 it already defaults to 1440 (one day). The value is stored in ~/Library/Preferences/pnpm/config.yaml. " +
				"Per-project settings live in pnpm-workspace.yaml, where minimumReleaseAgeExclude can exempt packages you need immediately.",
		},
		{
			ID:          "uv-cooldown",
			Category:    "Supply Chain",
			Name:        "uv: 9-day dependency cooldown",
			Description: "Limit uv to Python packages uploaded more than 9 days ago.",
			Commands: []string{
				`mkdir -p "$HOME/.config/uv" && touch "$HOME/.config/uv/uv.toml"`,
				// Prepend rather than append: an existing uv.toml may already open a
				// [table], and a key appended below one would be read as part of it.
				`grep -qs "^[[:space:]]*exclude-newer" "$HOME/.config/uv/uv.toml" && echo "exclude-newer is already set — leaving it alone" || { printf 'exclude-newer = "9 days"\n' | cat - "$HOME/.config/uv/uv.toml" > "$HOME/.config/uv/uv.toml.new" && mv "$HOME/.config/uv/uv.toml.new" "$HOME/.config/uv/uv.toml"; }`,
				// Resolving an empty requirements file is the cheapest way to make uv
				// parse the config it just wrote; a broken uv.toml fails the step here
				// rather than at the next real install.
				// Discard the resolution through a pipe rather than `-o /dev/null`:
				// uv writes its output file atomically via a sibling temp file, and
				// /dev/.tmpXXXX is not a path macOS lets it create.
				`if command -v uv >/dev/null; then uv pip compile /dev/null >/dev/null 2>&1 && echo "✓ uv reads the config"; else echo "(uv is not installed yet — the config is in place for when it is)"; fi`,
			},
			Note: "uv takes a relative duration, so the window rolls forward on its own rather than pinning a date. Units coarser than days are rejected — '9 days', not '1 week'. " +
				"Override per command with --exclude-newer, or exempt one package with --exclude-newer-package.",
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
