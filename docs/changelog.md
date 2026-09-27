---
title: Changelog
---

# :material-history: Changelog

For the complete list of changes, see the [GitHub Releases](https://github.com/WhiteRoseLK/neossh/releases) page.

## What's New & Fixed vs. lazyssh

If you are coming from **lazyssh**, here is a summary of everything **neossh** adds, improves, and fixes.

### New Features & Enhancements

| Feature | Description |
| :--- | :--- |
| **Color Themes** | Dark, Light, and System auto-detection. Toggle at runtime with ++t++. |
| **Automatic Terminal Title** | Updates terminal window/tab titles to `ServerAlias (HostName)` on connection. |
| **Internationalization (i18n)** | Multilingual UI: English, French, Simplified Chinese. |
| **Import Known Hosts** | Bootstrap SSH config from `~/.ssh/known_hosts`. |
| **Hidden Hosts** | Hide jump hosts and internal nodes from the list. |
| **CLI Pre-filtering & Direct Connect** | `neossh prod` or `neossh -c <alias>`. |
| **Read-Only Mode** | Immutable viewer mode for production safety. |
| **Exit On Disconnect** | One-shot launcher with `-x`. |
| **Multi-Alias Support** | Preserves `Host web1 web2 staging` on writeback. |
| **SSH Config Tags** | Tags stored as comments for cross-machine sync. |
| **Wildcard Pattern Blocks** | `Host *.corp` displayed with `[wildcard]` badges. |
| **Diagnostic SSH Error Modals** | Clear error reason on connection failures. |
| **Portable Tilde Paths** | Normalizes to `~/.ssh/id_rsa` across platforms. |
| **Parallel Ping All** | Concurrent pings with colored latency badges. |
| **SSH Key Deployment** | One-touch `ssh-copy-id` from TUI. |
| **Copy SSH Command** | Copy full SSH command to clipboard. |
| **Pre-Connect Hooks** | Run scripts before connecting (VPN, WOL, tokens). |
| **Default Identity Key** | Global default key for new servers. |
| **SCP Command Generator** | Generate and copy SCP templates. |
| **Password Auth (sshpass)** | Secure password delivery via OS keyring. |
| **SSHFS Remote Mounts** | Mount remote filesystems locally. |
| **Port Forwarding & Tunnels** | Interactive tunnel assistant (Local/Remote/SOCKS5). |
| **Dual-Pane SFTP Manager** | WinSCP-style file manager in the TUI. |
| **External File Managers** | Launch yazi, ranger, filezilla, cyberduck, etc. |
| **SSH Certificate Support** | Status display, expiry detection, auto-renewal. |
| **Config Backup & Export** | Compressed bundles with sanitized sharing. |
| **Paste SSH Command** | Parse SSH command from clipboard into add form. |
| **Clone Server** | Duplicate configs with auto-deduplication. |
| **Advanced Search Filters** | `tag:`, `user:`, `host:`, `port:`, `status:`, `group:`. |
| **Shell Autocompletion** | Bash, Zsh, Fish, PowerShell with dynamic aliases. |
| **Periodic Ping Watch** | Background health checks at configurable intervals. |
| **SSH Key Type Badges** | Visual `[ED25519]`, `[RSA-4096]`, `[FIDO2]` badges. |
| **Active Sessions Panel** | Track running SSH sessions with process controls. |

### Bug Fixes & Stability Improvements

- :material-resize: **Terminal Resize & Dynamic Layout** — Fixed UI clipping and freezes on resize.
- :material-lightning-bolt: **Zero-Disk-I/O Search & Pings** — In-memory state eliminates redundant config re-reads.
- :material-shield-check: **SSH Configuration Validation** — Comprehensive client-side validation prevents config corruption.
- :material-lock: **Security Hardening** — Fixed subprocess injection and path traversal vulnerabilities.
- :material-keyboard: **TUI Key Traps & Navigation** — Fixed backspace, input modal, and cursor issues.
- :material-folder: **XDG Base Directory Compliance** — Respects `$XDG_CONFIG_HOME` and `$XDG_STATE_HOME`.
- :material-format-quote-close: **Quoted Host Alias Stripping** — Handles `Host "server"` correctly.
- :material-account: **Numeric Username Validation** — Supports usernames like `007admin`.
- :material-file-tree: **Include Globs Matching Directories** — Directories skipped like OpenSSH does.
- :material-rocket-launch: **Automated Multi-Arch Releases** — GoReleaser + Semantic Release Please for all platforms.
