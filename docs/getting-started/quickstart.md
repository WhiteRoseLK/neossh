---
title: Quick Start
---

# :material-rocket-launch: Quick Start

Get productive with neossh in 2 minutes.

## Launch

```bash
neossh
```

neossh reads your `~/.ssh/config` and presents all configured servers in an interactive TUI.

## Navigate

| Key | Action |
|:---:|--------|
| ++j++ / ++k++ or ++arrow-down++ / ++arrow-up++ | Move through the server list |
| ++tab++ / ++shift+tab++ | Cycle focus between panels |
| ++0++ / ++1++ / ++2++ / ++3++ | Jump to Search / Servers / Active Sessions / Details |

## Connect

Press ++enter++ on any server to open an SSH session. neossh suspends the TUI and starts your system's native `ssh` binary.

## Search

Press ++slash++ or ++0++ to activate fuzzy search. Type to filter by alias, hostname, IP, user, or tags.

```
tag:prod user:root status:up web
```

See [Search & Navigation](../guide/search-navigation.md) for advanced filter syntax.

## Quick Actions

| Key | Action |
|:---:|--------|
| ++a++ | Add new server |
| ++e++ | Edit selected server |
| ++d++ | Delete server |
| ++c++ | Copy SSH command to clipboard |
| ++o++ | Generate SCP commands |
| ++m++ | SSHFS remote mount |
| ++f++ | Port forwarding assistant |
| ++shift+f++ | Launch SFTP file manager |
| ++ctrl+f++ | Built-in dual-pane SFTP |
| ++shift+g++ | Ping all servers |
| ++t++ | Edit tags |
| ++shift+t++ | Toggle theme |

## CLI Shortcuts

```bash
# Connect directly without TUI picker
neossh -c my-server

# Pre-filtered view (great for screen shares)
neossh prod

# Quick SFTP session
neossh --sftp web-prod

# Generate SCP commands
neossh --scp web-prod

# Read-only mode (safe for production)
neossh -r

# Exit after SSH disconnect (one-shot)
neossh -x

# Launch with French interface
neossh --lang fr
```

## What's Next?

- **[Server Management](../guide/server-management.md)** — Add, edit, organize, and manage servers
- **[CLI Reference](../reference/cli.md)** — Complete flag and option reference
- **[Keybindings](../reference/keybindings.md)** — Full keyboard shortcut reference
- **[SFTP & File Transfer](../guide/sftp-file-transfer.md)** — Built-in and external file managers
- **[Port Forwarding](../guide/port-forwarding.md)** — Tunnel assistant
- **[Security](../security.md)** — How neossh handles credentials
