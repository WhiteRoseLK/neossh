<div align="center">
  <h1>🚀 neossh</h1>
  <p><b>An interactive, keyboard-driven SSH manager for your terminal</b></p>
  <p><i>Created by <a href="https://github.com/Adembc">Adembc</a> • Maintained & developed by <a href="https://github.com/WhiteRoseLK">WhiteRoseLK</a> & the community</i></p>
</div>

<div align="center">

[![GitHub release](https://img.shields.io/github/v/release/WhiteRoseLK/neossh?style=flat-square)](https://github.com/WhiteRoseLK/neossh/releases)
[![Documentation](https://img.shields.io/badge/docs-GitHub_Wiki-blue?style=flat-square)](https://github.com/WhiteRoseLK/neossh/wiki)
[![License](https://img.shields.io/github/license/WhiteRoseLK/neossh?style=flat-square)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/WhiteRoseLK/neossh?style=flat-square)](go.mod)
[![Code of Conduct](https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg?style=flat-square)](CODE_OF_CONDUCT.md)
[![Security Policy](https://img.shields.io/badge/security-policy-brightgreen?style=flat-square)](SECURITY.md)
[![Fork of](https://img.shields.io/badge/fork%20of-Adembc%2Flazyssh-blue?style=flat-square)](https://github.com/Adembc/lazyssh)

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
| **Getting Started** | [Installation](https://github.com/WhiteRoseLK/neossh/wiki/Installation) • [Quick Start](https://github.com/WhiteRoseLK/neossh/wiki/Quick-Start) |
| **Guides** | [Server Management](https://github.com/WhiteRoseLK/neossh/wiki/Server-Management) • [Search & Navigation](https://github.com/WhiteRoseLK/neossh/wiki/Search-and-Navigation) • [SSH Configuration & Safety](https://github.com/WhiteRoseLK/neossh/wiki/SSH-Configuration) • [Key Management](https://github.com/WhiteRoseLK/neossh/wiki/Key-Management) • [SFTP & File Transfer](https://github.com/WhiteRoseLK/neossh/wiki/SFTP-and-File-Transfer) • [Port Forwarding & Tunnels](https://github.com/WhiteRoseLK/neossh/wiki/Port-Forwarding-and-Tunnels) • [SSHFS Remote Mounts](https://github.com/WhiteRoseLK/neossh/wiki/SSHFS-Remote-Mounts) • [Backup & Export](https://github.com/WhiteRoseLK/neossh/wiki/Backup-and-Export) • [Pre-Connect Hooks](https://github.com/WhiteRoseLK/neossh/wiki/Pre-Connect-Hooks) • [Password Auth (sshpass)](https://github.com/WhiteRoseLK/neossh/wiki/Password-Authentication) • [SSH Certificates](https://github.com/WhiteRoseLK/neossh/wiki/SSH-Certificates) • [Tags & Comments](https://github.com/WhiteRoseLK/neossh/wiki/Tags-and-Config-Comments) • [Internationalization (i18n)](https://github.com/WhiteRoseLK/neossh/wiki/Internationalization) |
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

<p align="left">
  <a href="https://github.com/Adembc" title="Adembc (Original Creator)">
    <img src="https://github.com/Adembc.png?size=64" width="64" height="64" style="border-radius: 50%; vertical-align: middle; margin-right: 8px;" alt="Adembc" />
  </a>
  <b><a href="https://github.com/Adembc">Adembc</a></b> — Creator of <a href="https://github.com/Adembc/lazyssh">lazyssh</a>. All core credit for the foundational idea and architecture belongs to him.
</p>

#### Upstream & Community Contributors

Full credit to all upstream and community contributors whose ideas, issues, and code made this continuation possible:

<p align="center">
  <a href="https://github.com/DelphicOkami" title="DelphicOkami"><img src="https://github.com/DelphicOkami.png?size=48" width="48" height="48" style="border-radius: 50%; margin: 3px;" alt="DelphicOkami" /></a>
  <a href="https://github.com/malaiwah" title="malaiwah"><img src="https://github.com/malaiwah.png?size=48" width="48" height="48" style="border-radius: 50%; margin: 3px;" alt="malaiwah" /></a>
  <a href="https://github.com/aabichou" title="aabichou"><img src="https://github.com/aabichou.png?size=48" width="48" height="48" style="border-radius: 50%; margin: 3px;" alt="aabichou" /></a>
  <a href="https://github.com/barthofu" title="barthofu"><img src="https://github.com/barthofu.png?size=48" width="48" height="48" style="border-radius: 50%; margin: 3px;" alt="barthofu" /></a>
  <a href="https://github.com/omani" title="omani"><img src="https://github.com/omani.png?size=48" width="48" height="48" style="border-radius: 50%; margin: 3px;" alt="omani" /></a>
  <a href="https://github.com/yaronuliel" title="yaronuliel"><img src="https://github.com/yaronuliel.png?size=48" width="48" height="48" style="border-radius: 50%; margin: 3px;" alt="yaronuliel" /></a>
  <a href="https://github.com/natefabian18" title="natefabian18"><img src="https://github.com/natefabian18.png?size=48" width="48" height="48" style="border-radius: 50%; margin: 3px;" alt="natefabian18" /></a>
  <a href="https://github.com/Ferdyverse" title="Ferdyverse"><img src="https://github.com/Ferdyverse.png?size=48" width="48" height="48" style="border-radius: 50%; margin: 3px;" alt="Ferdyverse" /></a>
  <a href="https://github.com/Q0" title="Q0"><img src="https://github.com/Q0.png?size=48" width="48" height="48" style="border-radius: 50%; margin: 3px;" alt="Q0" /></a>
  <a href="https://github.com/Midas-sudo" title="Midas-sudo"><img src="https://github.com/Midas-sudo.png?size=48" width="48" height="48" style="border-radius: 50%; margin: 3px;" alt="Midas-sudo" /></a>
  <a href="https://github.com/Mehrdad-Farshi" title="Mehrdad-Farshi"><img src="https://github.com/Mehrdad-Farshi.png?size=48" width="48" height="48" style="border-radius: 50%; margin: 3px;" alt="Mehrdad-Farshi" /></a>
  <a href="https://github.com/leleobhz" title="leleobhz"><img src="https://github.com/leleobhz.png?size=48" width="48" height="48" style="border-radius: 50%; margin: 3px;" alt="leleobhz" /></a>
  <a href="https://github.com/eznix86" title="eznix86"><img src="https://github.com/eznix86.png?size=48" width="48" height="48" style="border-radius: 50%; margin: 3px;" alt="eznix86" /></a>
  <a href="https://github.com/OleksandrKucherenko" title="OleksandrKucherenko"><img src="https://github.com/OleksandrKucherenko.png?size=48" width="48" height="48" style="border-radius: 50%; margin: 3px;" alt="OleksandrKucherenko" /></a>
  <a href="https://github.com/gonsalvesc" title="gonsalvesc"><img src="https://github.com/gonsalvesc.png?size=48" width="48" height="48" style="border-radius: 50%; margin: 3px;" alt="gonsalvesc" /></a>
  <a href="https://github.com/levinion" title="levinion"><img src="https://github.com/levinion.png?size=48" width="48" height="48" style="border-radius: 50%; margin: 3px;" alt="levinion" /></a>
  <a href="https://github.com/gaoyifan" title="gaoyifan"><img src="https://github.com/gaoyifan.png?size=48" width="48" height="48" style="border-radius: 50%; margin: 3px;" alt="gaoyifan" /></a>
  <a href="https://github.com/k161196" title="k161196"><img src="https://github.com/k161196.png?size=48" width="48" height="48" style="border-radius: 50%; margin: 3px;" alt="k161196" /></a>
  <a href="https://github.com/shekel588" title="shekel588"><img src="https://github.com/shekel588.png?size=48" width="48" height="48" style="border-radius: 50%; margin: 3px;" alt="shekel588" /></a>
  <a href="https://github.com/vtmocanu" title="vtmocanu"><img src="https://github.com/vtmocanu.png?size=48" width="48" height="48" style="border-radius: 50%; margin: 3px;" alt="vtmocanu" /></a>
  <a href="https://github.com/davidszp" title="davidszp"><img src="https://github.com/davidszp.png?size=48" width="48" height="48" style="border-radius: 50%; margin: 3px;" alt="davidszp" /></a>
  <a href="https://github.com/maxadc" title="maxadc"><img src="https://github.com/maxadc.png?size=48" width="48" height="48" style="border-radius: 50%; margin: 3px;" alt="maxadc" /></a>
  <a href="https://github.com/franksl" title="franksl"><img src="https://github.com/franksl.png?size=48" width="48" height="48" style="border-radius: 50%; margin: 3px;" alt="franksl" /></a>
  <a href="https://github.com/mahyarmirrashed" title="mahyarmirrashed"><img src="https://github.com/mahyarmirrashed.png?size=48" width="48" height="48" style="border-radius: 50%; margin: 3px;" alt="mahyarmirrashed" /></a>
  <a href="https://github.com/vetash" title="vetash"><img src="https://github.com/vetash.png?size=48" width="48" height="48" style="border-radius: 50%; margin: 3px;" alt="vetash" /></a>
  <a href="https://github.com/piRGoif" title="piRGoif"><img src="https://github.com/piRGoif.png?size=48" width="48" height="48" style="border-radius: 50%; margin: 3px;" alt="piRGoif" /></a>
  <a href="https://github.com/pranav79" title="pranav79"><img src="https://github.com/pranav79.png?size=48" width="48" height="48" style="border-radius: 50%; margin: 3px;" alt="pranav79" /></a>
  <a href="https://github.com/mas-kon" title="mas-kon"><img src="https://github.com/mas-kon.png?size=48" width="48" height="48" style="border-radius: 50%; margin: 3px;" alt="mas-kon" /></a>
  <a href="https://github.com/0xkatana" title="0xkatana"><img src="https://github.com/0xkatana.png?size=48" width="48" height="48" style="border-radius: 50%; margin: 3px;" alt="0xkatana" /></a>
  <a href="https://github.com/arniom" title="arniom"><img src="https://github.com/arniom.png?size=48" width="48" height="48" style="border-radius: 50%; margin: 3px;" alt="arniom" /></a>
  <a href="https://github.com/leoncamel" title="leoncamel"><img src="https://github.com/leoncamel.png?size=48" width="48" height="48" style="border-radius: 50%; margin: 3px;" alt="leoncamel" /></a>
  <a href="https://github.com/breakersun" title="breakersun"><img src="https://github.com/breakersun.png?size=48" width="48" height="48" style="border-radius: 50%; margin: 3px;" alt="breakersun" /></a>
  <a href="https://github.com/OlalalalaO" title="OlalalalaO"><img src="https://github.com/OlalalalaO.png?size=48" width="48" height="48" style="border-radius: 50%; margin: 3px;" alt="OlalalalaO" /></a>
  <a href="https://github.com/manato-tajiri" title="manato-tajiri"><img src="https://github.com/manato-tajiri.png?size=48" width="48" height="48" style="border-radius: 50%; margin: 3px;" alt="manato-tajiri" /></a>
  <a href="https://github.com/komapro" title="komapro"><img src="https://github.com/komapro.png?size=48" width="48" height="48" style="border-radius: 50%; margin: 3px;" alt="komapro" /></a>
</p>
