# Welcome to the neossh Wiki

**neossh** is an interactive, keyboard-driven SSH manager for your terminal. With neossh, you can quickly navigate, connect, manage, and configure servers defined in your `~/.ssh/config` without remembering IP addresses or dealing with complex SSH commands.

> [!NOTE]
> **neossh is a direct fork of [lazyssh](https://github.com/Adembc/lazyssh)**, originally created by [Adembc](https://github.com/Adembc).
> All core credit for the foundational idea, design, and original implementation belongs to **Adembc**.

---

## ⚡ Quick Navigation

| Section | Description |
|---|---|
| **[[Installation]]** | Install via Homebrew, AUR, prebuilt binaries, Scoop, Go install, or source |
| **[[Quick Start|Quick-Start]]** | 2-minute guide to get productive with neossh |
| **[[Server Management|Server-Management]]** | Add, edit, delete, pin, hide, clone, and organize servers into groups |
| **[[Search & Navigation|Search-and-Navigation]]** | Fuzzy search, token filters (`tag:`, `user:`, `host:`), and panel shortcuts |
| **[[SSH Configuration & Safety|SSH-Configuration]]** | Non-destructive writes, atomic saves, rolling backups, and `Include` support |
| **[[Key Management & Git|Key-Management]]** | Autocomplete, default identity keys, Git SSH profiles, and FIDO2 badges |
| **[[SFTP & File Transfer|SFTP-and-File-Transfer]]** | Built-in dual-pane WinSCP-style manager and external file transfer tools |
| **[[Port Forwarding & Tunnels|Port-Forwarding-and-Tunnels]]** | Assistant for Local (`-L`), Remote (`-R`), and Dynamic SOCKS5 (`-D`) tunnels |
| **[[SSHFS Remote Mounts|SSHFS-Remote-Mounts]]** | Mount remote filesystems locally with full SSH options |
| **[[Backup & Export Bundle|Backup-and-Export]]** | Export configuration bundles with `--sanitize` for safe sharing |
| **[[Pre-Connect Hooks|Pre-Connect-Hooks]]** | Automated scripts before connecting (VPN, Wake-on-LAN, tokens) |
| **[[Password Authentication|Password-Authentication]]** | `sshpass` authentication backed by OS keyring or AES-256-GCM vault |
| **[[SSH Certificates|SSH-Certificates]]** | Certificate status, expiration detection, and on-demand renewal |
| **[[Tags & Comments|Tags-and-Config-Comments]]** | Synchronize metadata directly inside `~/.ssh/config` comments |
| **[[Internationalization|Internationalization]]** | Multilingual UI support (English, French, Simplified Chinese) |
| **[[CLI Flags & Options|CLI-Reference]]** | Complete command-line flags and options table |
| **[[Keybindings Reference|Keybindings-Reference]]** | Complete keyboard shortcuts table |
| **[[Configuration Files Layout|Configuration-Files]]** | Layout and purpose of `~/.ssh/config`, `metadata.json`, `settings.json`, `vault.json` |
| **[[Shell Autocompletion|Shell-Completion]]** | Dynamic autocompletion for Bash, Zsh, Fish, and PowerShell |
| **[[Security Policy|Security]]** | Credential security architecture and privacy design |

---

## 🚀 Quick Launch

```bash
# Launch interactive TUI
neossh

# Connect directly to a server
neossh -c my-server

# Quick SFTP session
neossh --sftp web-prod

# Pre-filter server list (great for screen sharing)
neossh prod
```
