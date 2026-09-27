---
title: Keybindings Reference
---

# :material-keyboard: Keybindings Reference

This reference documents all keyboard shortcuts available inside the `neossh` interactive terminal user interface (TUI).

---

## Complete Keybindings Table

| Key | Action | Read-Only Safe? | Description |
|:---:|---|:---:|---|
| ++enter++ | **SSH Connect** | :material-check:{ .green } | Open native SSH connection to selected server |
| ++slash++ | **Search** | :material-check:{ .green } | Open fuzzy search input bar |
| ++a++ | **Add Server** | :material-close:{ .red } | Open modal form to add a new server entry |
| ++e++ | **Edit Server** | :material-close:{ .red } | Open modal form to edit selected server configuration |
| ++d++ | **Delete Server** | :material-close:{ .red } | Safely delete selected server with confirmation dialog |
| ++i++ | **Import Known Hosts** | :material-close:{ .red } | Import discovered hosts from `~/.ssh/known_hosts` |
| ++space++ | **Toggle Group** | :material-check:{ .green } | Expand or collapse server folder group |
| ++m++ | **Context Menu** | :material-close:{ .red }* | Toggle hidden server (on server) / Group tmux menu (on group) |
| ++shift+h++ | **Toggle Hidden** | :material-check:{ .green } | Toggle visibility of hidden servers in the list |
| ++p++ | **Pin / Unpin** | :material-close:{ .red } | Pin or unpin server to top of list |
| ++shift+p++ / ++ctrl+g++ | **Git SSH Profiles** | :material-close:{ .red } | Open Git SSH Key Configuration & Profile Switcher |
| ++t++ | **Edit Tags** | :material-close:{ .red } | Edit tags stored in SSH config comments |
| ++shift+t++ | **Toggle Theme** | :material-check:{ .green } | Cycle themes: Dark :octicons-arrow-right-24: Light :octicons-arrow-right-24: System |
| ++c++ | **Copy SSH Command** | :material-check:{ .green } | Copy full connection command to system clipboard |
| ++o++ | **SCP Modal** | :material-check:{ .green } | Open SCP command generator to copy upload/download commands |
| ++shift+m++ | **SSHFS Modal** | :material-check:{ .green } | Open SSHFS remote mount generator to copy mount commands |
| ++shift+c++ | **Key Comment** | :material-close:{ .red } | Inspect and edit SSH public/private key comment |
| ++l++ | **Agent Load** | :material-close:{ .red } | Load server identity key into system `ssh-agent` |
| ++u++ | **Agent Unload** | :material-close:{ .red } | Unload server identity key from `ssh-agent` |
| ++v++ | **Paste SSH Command** | :material-close:{ .red } | Parse SSH command from clipboard into Add form |
| ++y++ | **Clone Server** | :material-close:{ .red } | Duplicate server configuration with automatic alias deduplication |
| ++shift+k++ | **Kill / Deploy Key** | :material-close:{ .red } | Kill session (Active panel) / Run `ssh-copy-id` (Servers panel) |
| ++f++ | **Port Forwarding** | :material-check:{ .green } | Open interactive SSH port forwarding & tunnel assistant |
| ++shift+f++ | **Launch SFTP** | :material-check:{ .green } | Launch configured SFTP / external file manager |
| ++ctrl+f++ | **Dual-Pane SFTP** | :material-check:{ .green } | Open built-in dual-pane SFTP manager directly |
| ++s++ | **Toggle Sort** | :material-check:{ .green } | Cycle sort mode (Alias A-Z, Z-A, Last SSH) |
| ++g++ | **Ping Server** | :material-check:{ .green } | Ping selected server and measure round-trip latency |
| ++shift+g++ | **Ping All** | :material-check:{ .green } | Parallel ping all servers with color-coded latency badges |
| ++shift+w++ / ++ctrl+p++ | **Ping Watch** | :material-check:{ .green } | Toggle periodic background ping watch mode (default: 60s) |
| ++tab++ / ++shift+tab++ | **Cycle Panels** | :material-check:{ .green } | Move focus through Search, Servers, Active, Details panels |
| ++0++ | **Focus Search** | :material-check:{ .green } | Jump focus directly to search bar |
| ++1++ | **Focus Servers** | :material-check:{ .green } | Jump focus directly to servers list |
| ++2++ | **Focus Active** | :material-check:{ .green } | Jump focus directly to active sessions list |
| ++3++ | **Focus Details** | :material-check:{ .green } | Jump focus directly to details panel |
| ++j++ / ++k++ or ++arrow-down++ / ++arrow-up++ | **Navigate** | :material-check:{ .green } | Move cursor up / down in server or session lists |
| ++q++ / ++ctrl+c++ | **Quit** | :material-check:{ .green } | Exit `neossh` |

---

## Read-Only / Viewer Mode (`-r`)

When launched with `--readonly` or `-r`:

- All modifying actions (++a++, ++e++, ++d++, ++y++, ++v++, ++t++, ++shift+c++, ++shift+k++, ++i++, ++p++, ++l++, ++u++, ++shift+p++) are safely blocked.
- Attempting any restricted keybinding displays a clear informational dialog explaining that read-only mode is active.
- Safe exploration, search, SSH connecting, command copying, and monitoring operations remain fully accessible.
