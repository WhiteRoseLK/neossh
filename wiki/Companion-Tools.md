# Companion Tools Setup & Onboarding

`neossh` integrates seamlessly with specialized terminal companion tools to provide an elevated SSH management and workflow experience.

To ensure a frictionless experience across macOS, Linux, and Windows, `neossh` includes automated package manager detection and an interactive setup wizard available at first launch or anytime via the `--setup` CLI flag.

---

## 🛠️ Integrated Companion Tools

| Tool | Role in `neossh` | Primary Keybinding |
|---|---|:---:|
| **[chezmoi](https://www.chezmoi.io/)** | 1-click dotfiles synchronization to remote servers via in-memory tar stream | <kbd>D</kbd> |
| **[yazi](https://yazi-rs.github.io/)** | Lightning-fast asynchronous terminal file manager for visual SFTP exploration | <kbd>Shift+F</kbd> |
| **`ssh-copy-id`** | Standard OpenSSH utility to copy and install your public key on remote servers | <kbd>Shift+K</kbd> |

---

## 🧭 First-Launch Onboarding Wizard

When launching `neossh` for the first time, if any recommended companion tools are missing from your `$PATH`, an interactive onboarding dialog appears:

1. **System Scan**: Automatically detects your host operating system and active package manager.
2. **Component Status**: Displays a live checklist showing which companion tools are installed (`✓`) and which are missing (`✗`).
3. **Interactive Selection**: Choose which tools you wish to install using checkbox selectors.
4. **Automated Installation**: Executes installation commands via your system package manager with live progress indicators.
5. **Preference Persistence**: Once completed or skipped, `neossh` records `"first_run_completed": true` in `~/.neossh/settings.json` so the prompt does not reappear on subsequent runs.

---

## ⚡ On-Demand Setup Wizard (`neossh --setup`)

You can relaunch the companion tools setup wizard at any time from your terminal:

```bash
neossh --setup
```

This command directly opens the interactive setup wizard, re-scans installed tools, and allows upgrading or installing newly desired companion tools without affecting existing configurations.

---

## 📦 Supported Package Managers

`neossh` automatically probes and detects the appropriate package manager for your environment in the following order:

| Operating System | Detected Package Managers | Priority |
|---|---|:---:|
| **macOS** | Homebrew (`brew`) | 1 |
| **Debian / Ubuntu** | `apt` / `apt-get` | 1 |
| **Fedora / RHEL** | `dnf` / `yum` | 1 |
| **Arch Linux / Manjaro** | `pacman` | 1 |
| **openSUSE** | `zypper` | 1 |
| **Alpine Linux** | `apk` | 1 |
| **Windows** | `winget`, `scoop`, `choco` | 1, 2, 3 |
| **Any (Fallback)** | `go install` | Fallback |

---

## 📖 Manual Installation Guide

If you prefer installing companion tools manually or work in an air-gapped environment:

### chezmoi
```bash
# macOS (Homebrew)
brew install chezmoi

# Debian / Ubuntu
sudo apt install chezmoi

# Arch Linux
sudo pacman -S chezmoi

# Fedora
sudo dnf install chezmoi

# Windows (Scoop / Winget)
scoop install chezmoi
# or
winget install twpayne.chezmoi

# Direct binary install
sh -c "$(curl -fsLS get.chezmoi.io)"
```

### yazi
```bash
# macOS (Homebrew)
brew install yazi ffmpeg sevenzip jq poppler fd ripgrep fzf zoxide imagemagick font-symbols-only-nerd-font

# Arch Linux
sudo pacman -S yazi

# Debian / Ubuntu (via Cargo or prebuilt binary)
cargo install --locked yazi-fm yazi-cli

# Windows (Scoop / Winget)
scoop install yazi
# or
winget install sxyazi.yazi
```

### ssh-copy-id
- **Linux**: Typically bundled with `openssh-client`.
- **macOS**: `brew install ssh-copy-id` (or bundled with Homebrew OpenSSH).
- **Windows**: Built-in PowerShell equivalents or OpenSSH tools.
