<div align="center">
  <h1>🚀 neossh</h1>
  <p><b>An interactive, keyboard-driven SSH manager for your terminal</b></p>
  <p><i>Created by <a href="https://github.com/Adembc">Adembc</a> • Maintained & developed by <a href="https://github.com/WhiteRoseLK">WhiteRoseLK</a> & the community</i></p>
</div>

<div align="center">

[![GitHub release](https://img.shields.io/github/v/release/WhiteRoseLK/neossh?style=flat-square)](https://github.com/WhiteRoseLK/neossh/releases)
[![License](https://img.shields.io/github/license/WhiteRoseLK/neossh?style=flat-square)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/WhiteRoseLK/neossh?style=flat-square)](go.mod)

</div>

---

> [!NOTE]
> **neossh is a direct continuation and actively maintained fork of [lazyssh](https://github.com/Adembc/lazyssh)**, originally created by [Adembc](https://github.com/Adembc).
> All core credit for the foundational idea, architecture, and original implementation belongs to **Adembc**.
> This project exists to continue active development, integrate valuable community PRs, and provide ongoing maintenance.

> [!TIP]
> 📚 **Complete Documentation**: Looking for in-depth topic guides, advanced configuration, and complete references? Check out the [**neossh Wiki**](https://github.com/WhiteRoseLK/neossh/wiki).

---

## 💡 About neossh

**neossh** is an interactive, keyboard-driven SSH manager for your terminal. With neossh, you can quickly navigate, search, connect, manage, and configure servers defined in your `~/.ssh/config` without remembering IP addresses or dealing with complex SSH commands.

### 🌟 Key Highlights

- ⌨️ **Keyboard-Driven TUI**: Intuitive, high-performance terminal interface with fuzzy search, panel switching (<kbd>0</kbd>–<kbd>3</kbd>), and instant navigation.
- 🔍 **Advanced Filter Syntax**: Filter servers by `tag:`, `user:`, `host:`, `port:`, `status:`, and `group:` with negative exclusions (e.g. `-tag:staging`).
- 📁 **Dual-Pane & External SFTP Manager**: Built-in WinSCP-style file manager (<kbd>Ctrl+F</kbd>) or launch external tools (<kbd>F</kbd>) like Yazi, Ranger, FileZilla, Cyberduck, and Dolphin.
- 🔗 **Port Forwarding Assistant**: Interactive wizard (<kbd>f</kbd>) for Local (`-L`), Remote (`-R`), and Dynamic SOCKS5 (`-D`) tunnels with saved favorite profiles per host.
- 📜 **SSH Certificates & Auto-Renewal**: Certificate validity status badges (`✓ Valid`, `⚠ Expiring soon`, `✗ Expired`) and automatic on-demand renewal before connecting (`step`, `vault`, `tsh`).
- 🛡️ **Config Safety & Non-Destructive Writes**: Atomic saves, rolling backups, and recursive OpenSSH `Include` directive support.
- 📦 **Backup & Sanitized Export**: Archive configuration into `.tar.gz` bundles with `--sanitize` to strip private key paths and comments for safe team sharing.
- 🚀 **Dotfiles Synchronization (chezmoi)**: 1-click deployment of personal dotfiles (<kbd>D</kbd>) to remote hosts via streaming tar archive with zero remote dependencies.
- 🧭 **Companion Tools Onboarding Wizard**: Automated host package manager detection (`brew`, `apt`, `dnf`, `pacman`, `zypper`, `apk`, `winget`, `scoop`, `choco`) and 1-click installation of native companion tools (`chezmoi`, `yazi`, `ssh-copy-id`) on first launch or via `neossh --setup`.
- 🪝 **Pre-Connect Hooks**: Automated local scripts before connecting (VPN bring-up, Wake-on-LAN, token refresh).
- 🔑 **Secure Password Authentication**: Automated delivery via `sshpass` backed by native OS keyring (Keychain, Secret Service, Credential Manager) or AES-256-GCM local vault.
- 🎨 **Themes & Internationalization**: Dark, Light, and System themes (<kbd>T</kbd>), with multilingual UI support (English, French, Simplified Chinese).

---

## 📷 Screenshots

<details>
<summary>Click to expand screenshots</summary>

### Startup
<img src="./docs/loader.png" alt="Startup screen" />

### Server List & Active Sessions
<img src="./docs/list server.png" alt="Server list and active sessions" />

### Fuzzy Search
<img src="./docs/search.png" alt="Fuzzy search" />

### Add / Edit Server Form
<img src="./docs/add server.png" alt="Add server form" />

### SCP & SSH Command Generator
<img src="./docs/ssh.png" alt="SCP and SSH command generator" />

</details>

---

## 🚀 Quick Install

### Homebrew (macOS & Linux) — Official Tap

```bash
brew tap WhiteRoseLK/tap
brew trust WhiteRoseLK/tap
brew install neossh
```

### Go Install

```bash
go install github.com/WhiteRoseLK/neossh/cmd@latest
```

### Other Installation Options

| Platform / Method | Command / Link |
|---|---|
| **Arch Linux (AUR)** | `yay -S neossh` or `paru -S neossh` |
| **Windows (Scoop)** | `scoop install https://raw.githubusercontent.com/WhiteRoseLK/neossh/main/scoop/neossh.json` |
| **Pre-compiled Binaries** | Download from [GitHub Releases](https://github.com/WhiteRoseLK/neossh/releases/latest) |
| **Build from Source** | `git clone https://github.com/WhiteRoseLK/neossh.git && cd neossh && make build` |

👉 *Full instructions and options in the **[Installation Guide](https://github.com/WhiteRoseLK/neossh/wiki/Installation)**.*

---

## 💻 Quick Usage & Essential Shortcuts

### Common CLI Commands

```bash
# Launch interactive TUI
neossh

# Run interactive companion tools setup wizard
neossh --setup

# Connect directly to a server without TUI picker
neossh -c my-server

# Quick-launch SFTP file manager
neossh --sftp web-prod

# Launch pre-filtered view (useful for screen shares)
neossh prod

# Open in safe read-only viewer mode
neossh -r

# Connect and exit automatically when the SSH session ends
neossh -c -x my-server
```

### Essential Keyboard Shortcuts

| Key | Action |
|:---:|---|
| <kbd>Enter</kbd> | SSH into selected server |
| <kbd>/</kbd> or <kbd>0</kbd> | Fuzzy search by alias, hostname, IP, user, or tags |
| <kbd>a</kbd> / <kbd>e</kbd> / <kbd>d</kbd> | Add / Edit / Delete server configuration |
| <kbd>y</kbd> / <kbd>v</kbd> | Clone selected server / Paste SSH command from clipboard |
| <kbd>D</kbd> | Deploy and sync personal dotfiles to remote server via `chezmoi archive` |
| <kbd>f</kbd> | Interactive SSH port forwarding & tunnel assistant |
| <kbd>F</kbd> / <kbd>Ctrl+F</kbd> | Launch SFTP session (external tool) / Built-in dual-pane file manager |
| <kbd>c</kbd> / <kbd>o</kbd> / <kbd>M</kbd> | Copy SSH command / Generate SCP templates / Mount via SSHFS |
| <kbd>G</kbd> / <kbd>W</kbd> | Parallel ping all servers / Toggle periodic background ping watch |
| <kbd>t</kbd> / <kbd>T</kbd> | Edit server tags in SSH config comments / Toggle color theme |
| <kbd>Tab</kbd> / <kbd>Shift+Tab</kbd> | Cycle focus between Search, Servers, Active Sessions, and Details |
| <kbd>q</kbd> | Quit neossh |

👉 *Complete tables in **[CLI Reference](https://github.com/WhiteRoseLK/neossh/wiki/CLI-Reference)** and **[Keybindings Reference](https://github.com/WhiteRoseLK/neossh/wiki/Keybindings-Reference)**.*

---

## 📖 Documentation Index

All in-depth documentation is organized in our [**GitHub Wiki**](https://github.com/WhiteRoseLK/neossh/wiki):

| Category | Wiki Pages |
|---|---|
| **Getting Started** | [Installation](https://github.com/WhiteRoseLK/neossh/wiki/Installation) • [Quick Start](https://github.com/WhiteRoseLK/neossh/wiki/Quick-Start) • [Companion Tools Setup](https://github.com/WhiteRoseLK/neossh/wiki/Companion-Tools) |
| **Guides** | [Server Management](https://github.com/WhiteRoseLK/neossh/wiki/Server-Management) • [Search & Navigation](https://github.com/WhiteRoseLK/neossh/wiki/Search-and-Navigation) • [SSH Configuration & Safety](https://github.com/WhiteRoseLK/neossh/wiki/SSH-Configuration) • [Key Management](https://github.com/WhiteRoseLK/neossh/wiki/Key-Management) • [Dotfiles Sync (chezmoi)](https://github.com/WhiteRoseLK/neossh/wiki/Dotfiles-Sync) • [SFTP & File Transfer](https://github.com/WhiteRoseLK/neossh/wiki/SFTP-and-File-Transfer) • [Port Forwarding & Tunnels](https://github.com/WhiteRoseLK/neossh/wiki/Port-Forwarding-and-Tunnels) • [SSHFS Remote Mounts](https://github.com/WhiteRoseLK/neossh/wiki/SSHFS-Remote-Mounts) • [Backup & Export](https://github.com/WhiteRoseLK/neossh/wiki/Backup-and-Export) • [Pre-Connect Hooks](https://github.com/WhiteRoseLK/neossh/wiki/Pre-Connect-Hooks) • [Password Auth (sshpass)](https://github.com/WhiteRoseLK/neossh/wiki/Password-Authentication) • [SSH Certificates](https://github.com/WhiteRoseLK/neossh/wiki/SSH-Certificates) • [Tags & Comments](https://github.com/WhiteRoseLK/neossh/wiki/Tags-and-Config-Comments) • [Internationalization (i18n)](https://github.com/WhiteRoseLK/neossh/wiki/Internationalization) |
| **References** | [CLI Flags & Options](https://github.com/WhiteRoseLK/neossh/wiki/CLI-Reference) • [Keybindings Reference](https://github.com/WhiteRoseLK/neossh/wiki/Keybindings-Reference) • [Configuration Files Layout](https://github.com/WhiteRoseLK/neossh/wiki/Configuration-Files) • [Shell Autocompletion](https://github.com/WhiteRoseLK/neossh/wiki/Shell-Completion) • [Security Policy](https://github.com/WhiteRoseLK/neossh/wiki/Security) |

---

## 🤝 Contributing & Community Standards

Contributions are welcome! We strive to foster an open, welcoming, and inclusive community:

- **[Contributing Guide](CONTRIBUTING.md)**: Development setup, testing, conventions, and pull request guidelines.
- **[Code of Conduct](CODE_OF_CONDUCT.md)**: Our standards based on the Contributor Covenant v2.1.
- **[Security Policy](SECURITY.md)**: Responsible vulnerability disclosure and security architecture.

---

## 📄 License & Attribution

This project is licensed under the [Apache-2.0 License](LICENSE).

### Credits & Acknowledgments

#### Original Creator

- **[Adembc](https://github.com/Adembc)** — Creator of [lazyssh](https://github.com/Adembc/lazyssh). All core credit for the foundational idea and architecture belongs to him.

#### Upstream & Community Contributors

Full credit to all upstream and community contributors whose ideas, issues, and code made this continuation possible:

<p align="center">
  <a href="https://github.com/WhiteRoseLK/neossh/graphs/contributors">
    <img src="https://contrib.rocks/image?repo=WhiteRoseLK/neossh" alt="Contributors" />
  </a>
</p>
