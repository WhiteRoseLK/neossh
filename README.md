<div align="center">
  <h1>🚀 neossh</h1>
  <p><b>An actively maintained fork and continuation of <a href="https://github.com/Adembc/lazyssh">lazyssh</a></b></p>
  <p><i>Created by <a href="https://github.com/Adembc">Adembc</a> • Maintained & developed by <a href="https://github.com/WhiteRoseLK">WhiteRoseLK</a> & the community</i></p>
</div>

<div align="center">

[![GitHub release](https://img.shields.io/github/v/release/WhiteRoseLK/neossh?style=flat-square)](https://github.com/WhiteRoseLK/neossh/releases)
[![License](https://img.shields.io/github/license/WhiteRoseLK/neossh?style=flat-square)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/WhiteRoseLK/neossh?style=flat-square)](go.mod)
[![Fork of](https://img.shields.io/badge/fork%20of-Adembc%2Flazyssh-blue?style=flat-square)](https://github.com/Adembc/lazyssh)

</div>

---

> [!NOTE]
> **neossh is a direct fork of [lazyssh](https://github.com/Adembc/lazyssh)**, originally created by [Adembc](https://github.com/Adembc).
> All core credit for the foundational idea, design, and original implementation belongs to **Adembc**.
> This project exists solely because the original repository became unmaintained while having numerous valuable open PRs and issues. Rather than letting that work gather dust, **neossh** continues development, integrates community contributions, and provides ongoing maintenance.

## 💡 About neossh

**neossh** is an interactive, keyboard-driven SSH manager for your terminal. With neossh, you can quickly navigate, connect, manage, and configure servers defined in your `~/.ssh/config` without remembering IP addresses or dealing with complex SSH commands.

### Why neossh?

The original [lazyssh](https://github.com/Adembc/lazyssh) repository had not seen merged changes in over a year despite dozens of open issues and community-contributed pull requests. `neossh` was born to pick up the torch and give these contributions an actively maintained home.

---

## ⚡ What's New & Fixed vs. lazyssh?

If you are coming from **lazyssh**, here is a concrete summary of everything **neossh** adds, improves, and fixes:

### 🌟 New Features & Enhancements

| Feature | Description | Shortcut / Usage |
| :--- | :--- | :---: |
| **Color Themes (Dark, Light, System)** | Customize TUI appearance with dedicated Dark and Light palettes or automatic OS system appearance detection across macOS, Linux, and Windows, toggleable at runtime. | <kbd>T</kbd> / `--theme` |
| **Automatic Terminal / Tab Title** | Automatically updates terminal window and tab titles to `ServerAlias (HostName)` on SSH connection across modern terminal emulators (iTerm2, Alacritty, Kitty, Konsole, Windows Terminal) and restores it upon exit. | *Automatic* |
| **Internationalization (i18n)** | Multilingual UI support with runtime localization (English, French, Simplified Chinese). Seamlessly configured via CLI flag (`--lang`), `NEOSSH_LANG`, or compile-time defaults. | `--lang` / `-l` |
| **Import Known Hosts** | Quickly bootstrap your SSH config by discovering and importing unconfigured hosts from `~/.ssh/known_hosts` with automatic deduplication, standard port parsing (`[host]:port`), and safe skipping of hashed entries. | `--import-known-hosts` / <kbd>i</kbd> |
| **Hidden Hosts Support** | Hide jump hosts, proxy targets, or internal nodes from the primary server list (<kbd>m</kbd> or form), reveal on demand with <kbd>H</kbd>, or launch with hidden servers visible. | <kbd>m</kbd> / <kbd>H</kbd> / `-H` |
| **CLI Pre-filtering & Direct Connect** | Launch pre-filtered (`neossh prod` or `-f prod`) to prevent exposing your entire server fleet during screen shares, or connect directly (`neossh -c <alias>`). | `neossh <filter>` / `-c` |
| **Read-Only / Viewer Mode** | Protect production files with an immutable viewer mode. Blocks add, edit, delete, clone, paste, and key installs with an interactive indicator and notification. | `--readonly` / `-r` |
| **Exit On Disconnect** | Automatically exits `neossh` when your SSH session terminates, providing a seamless one-shot terminal launcher experience. | `-x` / `--exit-on-disconnect` |
| **Multi-Alias Directive Support** | Preserves and indexes all space-separated aliases on a single `Host` line (`Host web1 web2 staging`). Supports fuzzy search and connection by any defined alias without dropping them on writeback. | *Automatic* |
| **SSH Config Tag Comments** | Store and sync tags directly in `~/.ssh/config` comments (`# tags: prod, db` on the `Host` line or inside the block), keeping tags in sync across machines without relying solely on local `metadata.json`. | *Automatic* / <kbd>t</kbd> |
| **Wildcard Pattern Blocks** | Accurately reads and preserves wildcard configurations (`Host *.corp`, `Host *`) across edits, displays `[wildcard]` badges, and guards against accidental direct connections. | *Automatic* |
| **Diagnostic SSH Error Modals** | Intercepts SSH subprocess `stderr` on connection failures (`connection refused`, `host unreachable`, `permission denied`, timeouts) and displays the exact reason in a clear UI modal dialog. | *Automatic* |
| **Portable Tilde Paths (`~`)** | Normalizes absolute paths to portable relative tilde paths (`~/.ssh/id_rsa`) across Linux, macOS, and Windows. | *Automatic* |
| **Parallel Ping All** | Concurrently pings all configured servers in the background with real-time colored latency badges in the list: `[<50ms]` (green), `[<150ms]` (yellow), `[>150ms]` (red), or `[DOWN]` (red). | <kbd>G</kbd> |
| **One-Touch SSH Key Deployment** | Automatically pushes your public SSH key to the remote host using native `ssh-copy-id` directly from the TUI. | <kbd>K</kbd> |
| **Copy SSH Command** | Copies the full SSH connection command directly to your system clipboard. | <kbd>c</kbd> |
| **Pre-Connect Command Hooks** | Run automated local scripts/hooks (VPN bring-up, Wake-on-LAN, bastion tunnels, token refresh) before connecting. Supports token expansions (`%h`, `%p`, `%r`, `%n`), environment variables, and config comment persistence (`# pre-connect: ...`). | `--pre-connect <cmd>` / UI form |
| **Default SSH Identity Key** | Configure a global default private key (via CLI `--default-key <path>`, Git & SSH Keys Setup dialog, or `NEOSSH_DEFAULT_KEY`), automatically prefilling the identity file on new servers while allowing manual overrides. | `--default-key <path>` / UI form |
| **SCP Command Generator** | Generates and copies ready-to-use `scp` upload and download command templates with port, identity key, and proxy jump arguments directly to your clipboard. | <kbd>o</kbd> / `--scp <alias>` |
| **Secure Password Auth (sshpass)** | Automated password delivery for legacy hosts using `sshpass` backed by native OS keyring (macOS Keychain, Linux Secret Service, Windows Credential Manager) or AES-256-GCM vault—never written in plain text to `~/.ssh/config` or `metadata.json`. | `--password` / `-P` / UI form |
| **SSHFS Remote Mounts** | Mount remote server filesystems locally with full SSH configuration (ports, identity files, jump proxies, auto-reconnect, and read-only flags) and copy ready-to-run mount/unmount commands. | <kbd>M</kbd> / `--sshfs <alias>` |
| **SSH Port Forwarding & Tunnel Assistant** | Interactive port forwarding assistant for Local (`-L`), Remote (`-R`), and Dynamic SOCKS5 (`-D`) proxy tunnels. Supports saving favorite tunnel profiles per host in `~/.neossh/settings.json`, real-time command preview, background daemon launch, and one-touch command copy to clipboard. | <kbd>f</kbd> / `--tunnel <alias>` |
| **Built-in Dual-Pane SFTP File Manager** | Interactive full-screen dual-pane file manager (WinSCP / FileZilla style) inside the TUI. Browse local and remote directories over the SSH connection via the SFTP subsystem with directory navigation, sorting (name, size, modification date), non-blocking streaming transfers (<kbd>u</kbd> for upload, <kbd>d</kbd> for download), transfer speed and progress indicators, and overwrite confirmation dialogs. | <kbd>F</kbd> / <kbd>Ctrl+F</kbd> |
| **SSH Certificate Status & Expiry Detection** | Full support for OpenSSH certificates via `CertificateFile` or implicit `<IdentityFile>-cert.pub`. Server details display live validity status (`✓ Valid`, `⚠ Expiring soon`, `✗ Expired`), remaining time, principals, key ID, and validity window. Proportional "expiring soon" warning scaled dynamically to the certificate's lifetime (e.g. 6m for 1h certs, 24h for 30d certs). SSH failure dialog automatically detects and highlights expired certificates as the probable root cause. | *Automatic* / UI form |
| **On-Demand SSH Certificate Renewal** | Automatically renews short-lived SSH certificates before connecting via a configurable per-host command (`step ssh login`, `vault write`, `tsh login`, or custom script). If the certificate is already valid, connection starts immediately without prompts; if missing, expired, or expiring soon, the command runs once, verifies the certificate is now valid on disk, and proceeds. Configurable via UI form and SSH config comments (`# certificate-command:`). | UI form / `# certificate-command:` |
| **Paste SSH Command** | Parses any SSH command from system clipboard (flags, identity keys, ports, jump hosts) into an add-server modal with intelligent alias deduction and deduplication. | <kbd>v</kbd> |
| **Duplicate / Clone Server** | Instantly clones any existing server configuration into the Add form with automatic alias deduplication (`srv_1`, `srv_2`), eliminating manual re-typing. | <kbd>y</kbd> / <kbd>C</kbd> |
| **Zero-Friction Migration** | Automatically detects and migrates your favorites, tags, and connection history from `~/.lazyssh` to `~/.neossh`. | *Automatic* |
| **Custom Config Path** | Loads any alternative SSH config file without modifying `~/.ssh/config`. | `--sshconfig <path>` |
| **Windows Scoop Manifest** | Native Scoop package recipe allowing effortless installation and updates on Windows without administrator rights. | `scoop install ...` |
| **Active SSH Sessions Panel** | Dedicated live panel tracking running SSH and background sessions with process inspection (PID, forwarded ports, identity keys), one-touch process termination (<kbd>K</kbd>), and instant configuration generation (<kbd>a</kbd>) directly from running connections. | <kbd>2</kbd> / <kbd>K</kbd> |
| **Focus Borders & UI Navigation** | Distinct focus borders highlight the currently active panel (Search, Servers, Active Sessions, Details), with smooth <kbd>Tab</kbd> / <kbd>Shift+Tab</kbd> cycling across panels and form fields, active field highlights, and robust destructive confirmation dialogs. | <kbd>Tab</kbd> / <kbd>Shift+Tab</kbd> |
| **In-Memory Server State & Fast Search** | Server configuration hierarchy is preloaded and managed reactively in-memory with write-through persistence. Search-as-you-type and async background pings run instantaneously with 0 redundant disk I/O or parser allocations. | *Core Performance* |
| **Advanced Search Filters** | Token-based search syntax in the search bar: filter by `tag:`, `user:`, `host:`, `port:`, `status:`, and `group:` with negative exclusions (e.g. `-tag:staging`). | `/` or <kbd>0</kbd> |
| **Shell Autocompletion** | Full autocompletion for Bash, Zsh, Fish, and PowerShell with dynamic server alias completion (`neossh <TAB>`, `neossh -c <TAB>`, `--scp <TAB>`, `--sshfs <TAB>`), flags, themes, and language codes. | `neossh completion` |
| **Periodic Ping Watch Mode** | Continuous background health checks for all servers at configurable intervals (default: 60s), with real-time latency badges and a live countdown timer in the status bar. | <kbd>W</kbd> / <kbd>Ctrl+P</kbd> / `--ping-watch` |
| **SSH Key Type Badges & FIDO2** | Visual badges in server details showing key algorithm (`[ED25519]`, `[RSA-4096]`, `[ECDSA]`), FIDO2/security key indicator (`[FIDO2]` for YubiKey, SoloKey), file existence check (✓ found / ⚠ missing), detected from public key headers. | *Automatic* |
| **Quick Panel Jump** | Instant focus switching between Search, Servers, Active Sessions, and Details panels using numeric keys. | <kbd>0</kbd> / <kbd>1</kbd> / <kbd>2</kbd> / <kbd>3</kbd> |
| **Modern Toolchain & Deps** | Fully upgraded to latest upstream packages (`tview v0.42`, `tcell/v2 v2.13`, `cobra v1.10`, `zap v1.28`, `go-runewidth v0.0.30`), with Go race detection and `golangci-lint` v2. | *Core* |

### 🛠️ Bug Fixes & Stability Improvements

- 📐 **Terminal Resize & Dynamic Layout**: Completely fixed UI clipping and freezes when resizing the terminal. Server table columns and latency badges dynamically recalculate widths without text overflow.
- ⚡ **Zero-Disk-I/O Search & Pings**: Eliminated repetitive config file re-reads during search filtering and parallel pings by maintaining thread-safe in-memory state with automatic write-through persistence, cutting latency from ~9ms to <0.5ms and removing massive garbage collector churn.
- 🛡️ **SSH Configuration Validation**: Comprehensive client-side validation prevents config corruption (bounds check on port numbers `1–65535`, hostname/IP validation, format checks on dynamic forwardings and escape characters).
- 🔒 **Security Hardening**: Fixed G204 subprocess variable injection risks and G703 path traversal vulnerabilities; all commands run with verified parameters.
- ⌨️ **TUI Key Traps & Navigation**: Fixed backspace key issues, input modal traps, and cursor glitches across terminal emulators.
- 🗂️ **XDG Base Directory Compliance**: Standardized config and state paths respecting `$XDG_CONFIG_HOME` and `$XDG_STATE_HOME`.
- 🏷️ **Quoted Host Alias Stripping**: Automatically sanitizes and strips enclosing quotes from `Host` lines in SSH configs (e.g. `Host "server"`), avoiding invalid hostname rejection during connection.
- 👤 **Numeric Username Validation**: Adjusted username validation rules to support usernames starting with a number or underscore (e.g. `007admin`), fully supporting Linux service and UID-based accounts.
- 📂 **Include Globs Matching Directories**: Fixed a startup failure that left the server list empty when a top-level `Include` glob (e.g. `Include config.d/*`) matched a subdirectory. Directories are now skipped like OpenSSH does, and `Include` directives are preserved verbatim on writeback.
- 🚀 **Automated Multi-Arch Releases**: Continuous delivery via GoReleaser and Semantic Release Please providing prebuilt binaries for macOS (Intel & Apple Silicon), Linux (x86_64, ARM64), and Windows, alongside an official Homebrew tap.

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

## ✨ Features

### Server Management
- 📜 Read & display servers from your `~/.ssh/config` in a clean, scrollable list.
- ➕ Add a new server from the UI with comprehensive SSH configuration options.
- ✏ Edit existing server entries directly from the UI with a tabbed interface.
- 🗑 Delete server entries safely.
- 📌 Pin / unpin servers to keep favorites at the top.
- 🏓 Ping server to check reachability.

### Server Folders & Groups
- 📁 Organize servers into collapsible folders/groups with nested hierarchies (e.g. `prod/database`, `infra/k8s`).
- 📂 Expand or collapse groups with `Space` or `Enter` for a tidy, clutter-free server list.
- 🪟 Tmux multi-pane integration: Launch and connect to all servers in a group simultaneously using tmux panes (<kbd>m</kbd> context menu).

### Quick Server Navigation & Advanced Search
- 🔍 Fuzzy search by alias, hostname, IP, user, or tags (`/` or `0`).
- 🎯 **Advanced Filter Syntax**:
  - `tag:<name>` / `tags:<name>` (e.g. `tag:prod`, `tag:k8s`) — filter by server tag.
  - `-tag:<name>` (e.g. `-tag:staging`) — exclude servers with tag.
  - `user:<username>` (e.g. `user:root`, `user:ubuntu`) — filter by username.
  - `-user:<username>` — exclude servers with user.
  - `host:<hostname>` (e.g. `host:192.168.`, `host:*.aws.*`) — filter by host/IP.
  - `port:<number>` (e.g. `port:2222`) — filter by SSH port.
  - `status:up` / `status:down` / `status:unknown` (or `online`/`offline`) — filter by ping status.
  - `group:<name>` (e.g. `group:production`) — filter by server folder/group.
  - Combine filters and search terms: `tag:prod user:root status:up web`
- 🖥 One‑keypress SSH into the selected server (`Enter`).
- 🏷 Tag servers (e.g., `prod`, `dev`, `test`) stored directly as SSH config comments for quick filtering and cross-machine synchronization.
- ↕️ Sort by alias or last SSH (toggle + reverse).
- 🔢 Jump focus between panels using numeric shortcuts (`0` Search, `1` Servers, `2` Active, `3` Details).

### Advanced SSH Configuration
- 🔗 Port forwarding (`LocalForward`, `RemoteForward`, `DynamicForward`).
- 🚀 Connection multiplexing for instant subsequent connections.
- 🔐 Advanced authentication options (public key, password, agent forwarding).
- 🔑 Automated password authentication via `sshpass` for legacy or restricted hosts.
- 🔒 Security settings (ciphers, MACs, key exchange algorithms).
- 🌐 Proxy settings (`ProxyJump`, `ProxyCommand`).
- ⚙️ Full SSH config options organized in a tabbed interface.

### Key Management & Git SSH Profiles
- 🔑 SSH key autocomplete with automatic detection of available keys in `~/.ssh/` and SSH config.
- 📝 Smart key selection with support for multiple identity files.
- 🐙 **Git SSH key configuration & profile switcher** (<kbd>P</kbd> or <kbd>Ctrl+G</kbd>): configure per-repository or global Git SSH keys via `core.sshCommand` for GitHub, GitLab, and Bitbucket.
- 💬 **SSH key comment editor** (<kbd>C</kbd>): inspect and directly edit public/private key comments.
- ⚡ **SSH Agent integration** (<kbd>l</kbd> / <kbd>u</kbd>): load or unload server identity keys directly into/from `ssh-agent`.
- 🗝️ **Configurable Default Identity Key**: designate a default SSH private key (via CLI `--default-key <path>`, Git & SSH Keys Setup dialog, or `NEOSSH_DEFAULT_KEY`), automatically prefilling new servers.
- 🏷️ **SSH Key Type Badges & FIDO2 Indicator**: visual badges displaying the key algorithm (`[ED25519]`, `[RSA-4096]`, `[ECDSA]`), FIDO2/security key detection (`[FIDO2]` for YubiKey, SoloKey), and file existence status (✓ found / ⚠ missing) in server details.
- 📜 **SSH Certificate Status & Expiry Detection**: first-class support for OpenSSH short-lived certificates via `CertificateFile` or implicit `<IdentityFile>-cert.pub`. Server details display live validity status badges (`✓ Valid`, `⚠ Expiring soon`, `✗ Expired`), remaining time, principals, key ID, and validity window. Expiration warnings scale proportionally to certificate lifetime, and SSH connection failures explicitly flag expired certificates.
- 🔄 **On-Demand SSH Certificate Renewal**: eliminate annoying repeated MFA/OIDC logins with smart on-demand certificate renewal. Specify a per-host certificate renewal command (e.g. `step ssh login %u@%h`, `vault write ...`, `tsh login`). Neossh checks certificate validity before connecting: if the certificate is already valid, connection starts immediately with zero overhead; if missing or expired, the command runs once, verifies on disk that a valid certificate was produced, and seamlessly establishes the connection. Stored in SSH config comments (`# certificate-command:`, `# cert-command:`, `# cert-renew:`).

### Remote Filesystem Mounts (SSHFS) & SFTP File Manager
- 📂 **SSHFS Remote Mounts** (<kbd>M</kbd>): mount remote server filesystems locally with full SSH configuration (ports, identity files, jump proxies, auto-reconnect, and read-only flags) and copy ready-to-run mount/unmount commands.
- 📁 **Dual-Pane SFTP File Manager** (<kbd>F</kbd> or <kbd>Ctrl+F</kbd>): interactive WinSCP / FileZilla style file manager with local (left) and remote (right) panels, directory navigation, sorting by name/size/date (<kbd>s</kbd>), seamless uploads (<kbd>u</kbd>) and downloads (<kbd>d</kbd>), real-time streaming transfer progress, and overwrite confirmation dialogs.

### Internationalization & Localization (i18n)
- 🌐 Multilingual user interface with native support for English (`en`), French (`fr`), and Simplified Chinese (`zh-CN`).
- 🔄 Dynamic language selection via CLI flag `--lang` / `-l` (e.g. `neossh --lang fr`) or `NEOSSH_LANG` environment variable.
- 🔤 Graceful fallback mechanism: missing translations automatically fall back to standard English keys.

---

## 🚀 Installation

### Option 1: Homebrew (macOS & Linux) — Official Tap

Install `neossh` using the official Homebrew tap:

```bash
brew tap WhiteRoseLK/tap
brew trust WhiteRoseLK/tap
brew install neossh
```

*(If you previously had `lazyssh` installed, Homebrew will seamlessly prompt to replace it while preserving your server configs and metadata).*

> [!NOTE]
> **Why an external tap instead of `brew install neossh` directly?**  
> Homebrew Core requires new packages to meet a community adoption threshold (typically 50–75 GitHub stars) before being accepted into the central registry.
> 
> Because `neossh` was recently established as an independent continuation of `lazyssh`, it is currently distributed via this official tap. Once the project meets Homebrew's notoriety criteria, we will submit a formula to `homebrew/core` so everyone can simply run `brew install neossh`.
> 
> ⭐ **[Star the repository](https://github.com/WhiteRoseLK/neossh)** to help us reach the threshold for Homebrew Core inclusion!

---

### Option 2: Arch Linux (AUR)

`neossh` is available in the Arch User Repository ([AUR/neossh](https://aur.archlinux.org/packages/neossh)):

```bash
# Using yay:
yay -S neossh

# Using paru:
paru -S neossh
```

---

### Option 3: Pre-compiled Binaries (Direct Download)

Ready-to-run binaries are available for **macOS**, **Linux**, and **Windows** on the [Releases page](https://github.com/WhiteRoseLK/neossh/releases/latest).

#### One-liner for macOS & Linux:

```bash
# Automatically downloads and extracts the latest binary for your OS and architecture:
OS="$(uname -s)"
ARCH="$(uname -m)"
[ "$ARCH" = "x86_64" ] && ARCH="x86_64" || ARCH="arm64"

curl -sL "https://github.com/WhiteRoseLK/neossh/releases/latest/download/neossh_${OS}_${ARCH}.tar.gz" | tar -xz

# Move binary to PATH:
sudo mv neossh /usr/local/bin/
```

#### Windows:

- **Via Scoop (Recommended)**:
  ```powershell
  # Install directly via repository manifest:
  scoop install https://raw.githubusercontent.com/WhiteRoseLK/neossh/main/scoop/neossh.json

  # Or add the official scoop bucket:
  scoop bucket add neossh https://github.com/WhiteRoseLK/scoop-bucket
  scoop install neossh
  ```
- **Manual Zip Download**:
  1. Download the `.zip` archive from the [Releases page](https://github.com/WhiteRoseLK/neossh/releases/latest).
  2. Extract `neossh.exe` to a folder in your `PATH` (e.g. `C:\Windows\System32` or a dedicated tools directory).

---

### Option 4: Go Install

If you have Go installed:

```bash
go install github.com/WhiteRoseLK/neossh/cmd@latest
```

*(Make sure `$GOPATH/bin` or `~/go/bin` is in your `$PATH`).*

---

### Option 5: Build from Source

**Prerequisites**: [Go](https://go.dev/) 1.22+ and `git` (and optionally `make`).

```bash
# 1. Clone the repository
git clone https://github.com/WhiteRoseLK/neossh.git
cd neossh

# 2. Build using Make
make build

# 3. Install the binary into your PATH
sudo cp bin/neossh /usr/local/bin/
```

*Or build manually without Make:*

```bash
go build -ldflags "-X main.version=v1.0.0" -o neossh ./cmd/main.go
sudo mv neossh /usr/local/bin/
```

---

## 💻 Command Line Usage

`neossh` provides command line flags for automation, alternative configurations, and scripting:

```bash
neossh [filter] [flags]
```

### Options & Flags

| Flag | Shorthand | Description | Default |
| :--- | :---: | :--- | :---: |
| `[filter]` | | Optional positional argument to pre-filter server list | `""` |
| `--filter <pattern>` | `-f` | Pre-filter server list by alias, hostname, or tag | `""` |
| `--connect` | `-c` | Connect directly to matching server without launching full TUI picker | `false` |
| `--import-known-hosts` | | Import newly discovered hosts from `known_hosts` into SSH config | `false` |
| `--known-hosts <path>` | | Specify custom path to `known_hosts` file | `~/.ssh/known_hosts` |
| `--git-ssh <key>` | | Configure Git SSH key for current repo (or globally) or view current setting | `""` |
| `--theme <mode>` | `-t` | Set color theme: `dark`, `light`, or `system` | `""` *(stored preference or dark)* |
| `--lang <code>` | `-l` | Set interface language: `en`, `fr`, `zh-CN` (or via `NEOSSH_LANG`) | `""` *(English default)* |
| `--show-hidden` | `-H` | Display hidden servers in UI list | `false` |
| `--scp <alias>` | | Generate SCP upload/download command templates for a server alias and copy to clipboard | `""` |
| `--sshfs <alias>` | | Generate SSHFS remote mount and unmount command templates for a server alias and copy to clipboard | `""` |
| `--tunnel <alias>`, `--forward <alias>` | | Generate SSH port forwarding / tunnel command templates for a server alias and copy to clipboard | `""` |
| `--pre-connect <cmd>` | | Run local hook command before SSH connect (supports `%h`, `%p`, `%r`, `%n`) | `""` |
| `--default-key <path>` | | Get or set default SSH identity key prefilled for new server entries | `""` |
| `--password <pwd>` | `-P` | Password for automated `sshpass` authentication | `""` |
| `--ping-watch` | | Enable periodic background ping watch mode | `false` |
| `--ping-interval <sec>` | | Interval in seconds for periodic background ping watch mode | `60` |
| `--sshconfig <path>` | | Specify custom path to SSH config file | `~/.ssh/config` |
| `--readonly`, `--ssh-config-readonly` | `-r` | Run in read-only / viewer mode (prevents writing or modifying SSH configuration) | `false` |
| `--exit-on-disconnect`, `--auto-exit` | `-x` | Exit `neossh` immediately after SSH session terminates (one-shot launcher) | `false` |
| `--help` | `-h` | Display help message and available options | |

#### Examples:
```bash
# Launch normal interactive TUI:
neossh

# Launch with continuous background ping watch every 30 seconds:
neossh --ping-watch --ping-interval 30

# Set default SSH identity key for new servers:
neossh --default-key ~/.ssh/id_ed25519

# Check current default SSH identity key:
neossh --default-key ""

# Generate SCP command templates for a server alias:
neossh --scp web-prod

# Generate SSHFS remote filesystem mount commands:
neossh --sshfs web-prod

# Generate SSH port forwarding command templates and view saved tunnel profiles:
neossh --tunnel web-prod

# Launch TUI with French localization:
neossh --lang fr

# Launch TUI with Simplified Chinese localization:
neossh -l zh-CN

# Configure Git SSH key for the current repository:
neossh --git-ssh ~/.ssh/id_ed25519_work

# Check current Git SSH configuration for the current directory:
neossh --git-ssh ""

# Launch with light theme:
neossh --theme light

# Launch following OS system appearance (auto dark/light):
neossh --theme system

# Launch TUI revealing all hidden hosts:
neossh -H

# Bootstrap SSH config by importing discovered hosts from ~/.ssh/known_hosts:
neossh --import-known-hosts

# Import from a custom known_hosts file into a specific SSH config:
neossh --import-known-hosts --known-hosts ~/.ssh/known_hosts_work --sshconfig ~/.ssh/config_work

# Launch pre-filtered to "prod" servers (avoids exposing other servers on screen shares):
neossh prod
# or using flag:
neossh -f prod

# Connect directly via SSH to a specific server alias without opening the picker:
neossh -c my-server

# Combine direct connect with exit-on-disconnect:
neossh -c -x my-server

# Connect directly with one-shot password authentication via sshpass:
neossh -c my-server -P "secretpassword"

# Open in safe read-only viewer mode (modifications disabled):
neossh -r

# Connect and exit automatically when the SSH session ends:
neossh -x

# Load a dedicated work or staging SSH config file in read-only mode:
neossh --sshconfig ~/.ssh/config_work -r
```

### 🐚 Shell Autocompletion

`neossh` provides built-in shell autocompletion for **Bash**, **Zsh**, **Fish**, and **PowerShell**, complete with dynamic server alias suggestions directly parsed from your SSH configuration.

- **Dynamic Server Alias Completion**: Press <kbd>Tab</kbd> after `neossh`, `neossh -c`, `neossh --scp`, `neossh --sshfs`, or `neossh --tunnel` to view matching host aliases alongside their `user@host:port` descriptions.
- **Flag & Option Completion**: Flags like `--theme` suggest available themes (`dark`, `light`, `system`), `--lang` suggests supported locales (`en`, `fr`, `zh-CN`), and `--sshconfig` / `--known-hosts` trigger path completions.

#### Installation Instructions

<details>
<summary><b>Bash</b></summary>

```bash
# To load completions for the current session:
source <(neossh completion bash)

# To load completions automatically for each session:
# Linux:
neossh completion bash | sudo tee /etc/bash_completion.d/neossh > /dev/null

# macOS (Homebrew bash-completion):
neossh completion bash > $(brew --prefix)/etc/bash_completion.d/neossh
```
</details>

<details>
<summary><b>Zsh</b></summary>

```bash
# If shell completion is not already enabled in your environment, add to ~/.zshrc:
echo "autoload -U compinit; compinit" >> ~/.zshrc

# Generate and save completions to your zsh completion folder:
neossh completion zsh > "${fpath[1]}/_neossh"

# Restart your zsh shell or reload completions:
exec zsh
```
</details>

<details>
<summary><b>Fish</b></summary>

```fish
# To load completions for the current session:
neossh completion fish | source

# To install completions permanently:
mkdir -p ~/.config/fish/completions
neossh completion fish > ~/.config/fish/completions/neossh.fish
```
</details>

<details>
<summary><b>PowerShell</b></summary>

```powershell
# To load completions for the current session:
neossh completion powershell | Out-String | Invoke-Expression

# To load completions automatically in your profile:
neossh completion powershell > "$HOME\Documents\PowerShell\neossh.ps1"
# And add ". $HOME\Documents\PowerShell\neossh.ps1" to your $PROFILE
```
</details>

---

## ⌨️ Keybindings

| Key | Action |
|:---:|--------|
| `Enter` | SSH into selected server |
| `/` | Fuzzy search by alias, IP, or tag |
| `a` | Add new server *(disabled in read-only mode)* |
| `e` | Edit selected server *(disabled in read-only mode)* |
| `d` | Delete selected server *(disabled in read-only mode)* |
| `i` | Import discovered hosts from `known_hosts` *(disabled in read-only mode)* |
| `Space` | Toggle collapse / expand group *(when on group header)* |
| `m` | Mark server hidden/visible *(on server)* / Group menu: Tmux connect all, collapse/expand all *(on group header)* |
| `H` | Toggle displaying hidden servers in the list |
| `p` | Pin / unpin server |
| `P` / `Ctrl+G` | Open Git SSH Key Configuration & Profile Switcher dialog *(disabled in read-only mode)* |
| `t` | Edit tags *(disabled in read-only mode)* |
| `T` | Toggle color theme (Dark → Light → System) |
| `c` | Copy SSH connection command to clipboard |
| `o` | Open SCP command generator modal to configure and copy upload/download commands |
| `M` | Open SSHFS remote mount command generator modal to configure and copy mount/unmount commands |
| `C` | Edit SSH key comment *(on server with identity file)* *(disabled in read-only mode)* |
| `l` | Load selected server's key into `ssh-agent` *(disabled in read-only mode)* |
| `u` | Unload selected server's key from `ssh-agent` *(disabled in read-only mode)* |
| `v` | Paste SSH command from clipboard *(disabled in read-only mode)* |
| `y` | Duplicate / clone selected server entry *(disabled in read-only mode)* |
| `K` | Terminate active SSH session (when on Active Sessions) / Push SSH key via `ssh-copy-id` (when on Servers) *(disabled in read-only mode)* |
| `f` | Open interactive SSH port forwarding and tunnel assistant (Local `-L`, Remote `-R`, Dynamic SOCKS5 `-D`, saved favorite profiles, live command preview, and clipboard copy) |
| `F` / `Ctrl+F` | Open built-in dual-pane SFTP file manager (WinSCP / FileZilla style: local & remote browsing, file upload/download with real-time transfer progress and overwrite confirmation) |
| `s` | Toggle sort mode (alias, last SSH, reverse) |
| `g` | Ping selected server |
| `G` | Ping all servers (parallel check with latency badges) |
| `W` / `Ctrl+P` | Toggle periodic background ping watch mode (default: 60s) |
| `Tab` / `Shift+Tab` | Cycle focus between Search, Servers, Active Sessions, and Details panels |
| `0` / `1` / `2` / `3` | Focus Search (`0`) / Servers (`1`) / Active Sessions (`2`) / Details (`3`) |
| `j` / `k` or `↓` / `↑` | Navigate server / active session list |
| `q` / `Ctrl+C` | Quit |

> [!NOTE]
> When launched with `--readonly` / `-r`, all modifying operations (`a`, `e`, `d`, `y`, `C`, `v`, `t`, `K`, `i`, `m`, `l`, `u`, `P`) are locked with clear informational dialogs, making it completely safe for shared or production environments.


---

## 🔐 Security Notice

`neossh` treats credential security and user privacy as paramount requirements:

- **No Plaintext Passwords on Disk**: Passwords are **never** stored in plain text anywhere on disk—neither in `~/.ssh/config` nor in `metadata.json`. Your SSH configuration remains clean, portable, and completely safe to version-control in dotfiles.
- **Native OS Keyring & Hardware Security**: Server passwords configured for automated `sshpass` login are managed by the operating system's native secure credential store (macOS Keychain, Linux Secret Service / DBus, Windows Credential Manager) with an authenticated AES-256-GCM local vault fallback (`0600` permissions).
- **Process Table Protection**: When connecting via `sshpass`, `neossh` delivers passwords through the `SSHPASS` environment variable (`sshpass -e`) rather than command-line arguments, preventing password exposure to other users in `ps aux`.
- **OpenSSH Native Execution**: All SSH connections run through your system's native `ssh` binary (OpenSSH). Your existing `IdentityFile` paths, passphrases, and `ssh-agent` integrations work exactly as before.
- **Strict File Permissions**: Config snapshots, backups, and vault files strictly enforce `0600` permissions.

---

## 🛡️ Config Safety: Non‑destructive writes and backups

- **Non‑destructive edits**: `neossh` only writes the minimal required changes to your `~/.ssh/config`. Comments, spacing, indentation, and untouched settings remain intact.
- **Atomic writes**: Updates are written to a temporary file and atomically renamed over the original to prevent corruption.
- **Backups**:
  - *One‑time snapshot*: Before `neossh` makes its first change, it creates a snapshot named `config.original.backup`. This file is never overwritten.
  - *Rolling backups*: On each save, `neossh` creates a timestamped backup (`~/.ssh/config-<timestamp>-neossh.backup`), keeping the 10 most recent.

---

## 📂 SSH Config `Include` Support

`neossh` honours top-level `Include` directives in your `~/.ssh/config`:

- **Reads**: All included files are parsed in OpenSSH precedence order.
- **Globs**: Wildcard patterns are expanded like OpenSSH; matches that are directories are skipped instead of aborting startup.
- **Writes route back to source**: Editing or deleting a host modifies the file that actually defines it. Other files are never touched.
- **Ambiguity modal**: If the same alias is defined in multiple files, a prompt asks which file to update. Your choice is remembered in `metadata.json`.

---

## 🏷️ Server Tags in SSH Config Comments

`neossh` stores server tags directly within your `~/.ssh/config` file as comments, ensuring your tags stay in sync across machines (e.g. via dotfiles or Git) without relying exclusively on a local machine-specific `metadata.json`:

```ssh
# Inline on the Host line:
Host web-prod # tags: prod, web, us-east
    HostName 192.168.1.10
    User ubuntu

# Or inside the Host block:
Host db-primary
    # tags: prod, database
    HostName 192.168.1.20
    User postgres
```

- **Two-Way Synchronization**: Tags edited in the TUI (via <kbd>t</kbd> or the full edit form <kbd>e</kbd>) are written directly to your SSH config.
- **Backwards Compatibility**: Any existing tags in `metadata.json` (or migrated from `~/.lazyssh`) are seamlessly loaded on startup and will be saved directly into `~/.ssh/config` upon your next edit.

---

## 🪝 Pre-Connect Command Hooks

`neossh` allows executing custom local commands or scripts right before connecting to an SSH host. This is particularly useful for:
- 🛡️ Triggering corporate VPN connection scripts before dialing private IP ranges.
- ⚡ Sending Wake-on-LAN (WOL) magic packets to spin up remote bare-metal hosts.
- 🔑 Refreshing short-lived cloud credentials or MFA tokens (AWS SSM, Cloudflare Access, Okta).

#### Configuration in `~/.ssh/config`
Pre-connect hooks can be configured directly in SSH config comments using `# pre-connect:` or `# hook:`:

```ssh-config
Host vpn-internal # pre-connect: /usr/local/bin/vpn-connect.sh %h
    HostName 10.10.0.50
    User devops

Host workstation-lab
    # hook: wakeonlan 00:11:22:33:44:55
    HostName 192.168.1.105
    User admin
```

#### Token Expansion & Context Environment Variables
Hook commands support tokens and environment variables:
- `%h`: Remote Hostname / IP address
- `%p`: Remote SSH Port
- `%r`: Remote User
- `%n`: Server Alias
- `%%`: Literal `%`
- Environment variables: `NEOSSH_ALIAS`, `NEOSSH_HOST`, `NEOSSH_PORT`, `NEOSSH_USER`

If the pre-connect hook command exits with a non-zero exit code, the SSH connection is safely aborted and the error output is reported.

---

## 🔑 Password Authentication via `sshpass`

For legacy servers or restricted environments that do not support SSH public key authentication, `neossh` provides automated password authentication via `sshpass` with end-to-end credential security:

- **Zero Plaintext Footprint in `~/.ssh/config`**: Unlike insecure approaches, `neossh` **never** stores passwords in `~/.ssh/config` or `metadata.json`. Your SSH configuration remains clean, portable, and completely safe to version-control in public or shared dotfiles.
- **Hardware-Backed Credential Store**: Passwords configured via the Add/Edit Server form under `▶ Password & Interactive` are securely stored in your operating system's native keyring (macOS Keychain, Linux Secret Service / DBus, Windows Credential Manager) with an authenticated AES-256-GCM local vault fallback.
- **Secure Process Invocation**: Uses `sshpass -e` with environment variable delivery (`SSHPASS`) rather than CLI flags, completely eliminating visibility in system process tables (`ps aux`).
- **One-Shot CLI Flag & Environment Variable**: Pass one-shot passwords via `--password <pwd>` (or `-P <pwd>`) or through the `NEOSSH_PASSWORD` / `SSHPASS` environment variables.

---

## 🤝 Contributing

Contributions are welcome! Feel free to open an [Issue](https://github.com/WhiteRoseLK/neossh/issues) or submit a Pull Request.

---

## 📄 License & Attribution

This project is licensed under the [Apache-2.0 License](LICENSE).

### Credits & Acknowledgments

- **[Adembc](https://github.com/Adembc)**: Original author and creator of [lazyssh](https://github.com/Adembc/lazyssh). Without his architectural work, `neossh` would not exist.
- **Community contributors**: Full credit to all contributors from the upstream repository whose ideas and pull requests made this release possible:
  - `@DelphicOkami`, `@malaiwah`, `@aabichou`, `@barthofu` — SSH `Include` support
  - `@omani` — `--sshconfig` custom config flag
  - `@yaronuliel` — `--ssh-config-readonly` mode
  - `@natefabian18` — `--exit-on-disconnect` session behavior
  - `@Ferdyverse` — Multi-alias `Host` lines support & SSH config tags comments
  - `@Q0`, `@Midas-sudo` — Server folders, nested grouping, and tmux session integration
  - `@Midas-sudo` — Wildcard pattern blocks & pre-connect command hooks
  - `@Mehrdad-Farshi` — SSH error diagnostics display
  - `@leleobhz` — CLI filter, direct connect options, and Arch Linux AUR package maintenance
  - `@eznix86` — Import hosts from `~/.ssh/known_hosts` (CLI flag & bootstrap)
  - `@OleksandrKucherenko` — Git SSH key configuration, profile switcher, and SSH key management
  - `@gonsalvesc` — XDG base directory specification support & Windows Scoop package manifest
  - `@levinion` — Copy SSH command shortcut
  - `@gaoyifan` — Persistent sort mode
  - `@k161196` — Panel focus shortcuts, active background SSH sessions panel, and process controls
  - `@shekel588` — Keyboard navigation improvements, focus borders, active field styling, and confirmation dialogs
  - `@vtmocanu` — Hidden hosts support, visibility toggling, and filtering
  - `@davidszp` — Dark, Light, and System color theme support with runtime toggle
  - `@maxadc` — Internationalization framework and localization support (English, French, Chinese)
  - `@franksl` — Quoted `Host` aliases sanitization
  - `@mahyarmirrashed` — Numeric username validation support
  - `@vetash` — Automatic terminal and tab title integration
  - `@piRGoif` — SCP command generator modal and CLI helper
  - `@pranav79` — Configurable default identity SSH key for new servers
  - `@mas-kon` — SSHFS remote filesystem mount integration
  - `@0xkatana` — Password authentication integration via `sshpass`
  - `@arniom`, `@leoncamel`, `@breakersun`, `@OlalalalaO`, `@manato-tajiri`, `@komapro` — Bug fixes & documentation improvements
