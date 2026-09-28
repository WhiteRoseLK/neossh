# Quick Start

Get productive with `neossh` in 2 minutes.

---

## 1. Launch

```bash
neossh
```

`neossh` reads your `~/.ssh/config` (and any `Include` files) and presents all configured servers in an interactive terminal user interface (TUI).

---

## 2. Navigate

| Key | Action |
|:---:|---|
| <kbd>j</kbd> / <kbd>k</kbd> or <kbd>↓</kbd> / <kbd>↑</kbd> | Move cursor up / down through the server list |
| <kbd>Tab</kbd> / <kbd>Shift+Tab</kbd> | Cycle focus between Search, Servers, Active, Details panels |
| <kbd>0</kbd> / <kbd>1</kbd> / <kbd>2</kbd> / <kbd>3</kbd> | Jump focus directly to Search / Servers / Active / Details |

---

## 3. Connect

Press <kbd>Enter</kbd> on any server to open an SSH session. `neossh` suspends the TUI and launches your system's native `ssh` binary. When the SSH session ends, the TUI seamlessly restores.

---

## 4. Search & Filter

Press <kbd>/</kbd> or <kbd>0</kbd> to activate fuzzy search. Type to filter by alias, hostname, IP, user, or tags.

```
tag:prod user:root status:up web
```

See **[[Search-and-Navigation]]** for the full filter syntax.

---

## 5. Quick Actions

| Key | Action |
|:---:|---|
| <kbd>a</kbd> | Add new server |
| <kbd>e</kbd> | Edit selected server |
| <kbd>d</kbd> | Delete selected server (with confirmation) |
| <kbd>c</kbd> | Copy SSH command to system clipboard |
| <kbd>o</kbd> | Open SCP command generator modal |
| <kbd>M</kbd> | Open SSHFS remote mount generator modal |
| <kbd>f</kbd> | Open SSH port forwarding & tunnel assistant |
| <kbd>F</kbd> | Launch SFTP session / configured external file manager |
| <kbd>Ctrl+F</kbd> | Open built-in dual-pane SFTP manager directly |
| <kbd>G</kbd> | Parallel ping all servers with color-coded latency badges |
| <kbd>t</kbd> | Edit tags stored in SSH config comments |
| <kbd>T</kbd> | Toggle color theme (Dark → Light → System) |

---

## 6. CLI Shortcuts

```bash
# Connect directly without opening the TUI picker
neossh -c my-server

# Pre-filtered view (great for screen shares)
neossh prod

# Quick SFTP session
neossh --sftp web-prod

# Generate SCP upload/download commands
neossh --scp web-prod

# Safe read-only mode for production environments
neossh -r

# Connect and exit immediately when session ends (one-shot)
neossh -c -x my-server
```

---

Next: read the comprehensive **[[Server-Management]]** guide!
