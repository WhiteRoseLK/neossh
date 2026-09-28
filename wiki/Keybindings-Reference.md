# ⌨️ Keybindings Reference

This reference documents all keyboard shortcuts available inside the `neossh` interactive terminal user interface (TUI).

---

## Complete Keybindings Table

| Key | Action | Read-Only Safe? | Description |
|:---:|---|:---:|---|
| <kbd>Enter</kbd> | **SSH Connect** | ✅ Yes | Open native SSH connection to selected server |
| <kbd>/</kbd> | **Search** | ✅ Yes | Open fuzzy search input bar |
| <kbd>a</kbd> | **Add Server** | 🔒 Blocked | Open modal form to add a new server entry |
| <kbd>e</kbd> | **Edit Server** | 🔒 Blocked | Open modal form to edit selected server configuration |
| <kbd>d</kbd> | **Delete Server** | 🔒 Blocked | Safely delete selected server with confirmation dialog |
| <kbd>i</kbd> | **Import Known Hosts** | 🔒 Blocked | Import discovered hosts from `~/.ssh/known_hosts` |
| <kbd>Shift+I</kbd> | **Known Hosts Manager** | 🔒 Blocked | Inspect, search, and delete host key entries from `~/.ssh/known_hosts` |
| <kbd>Space</kbd> | **Toggle Group** | ✅ Yes | Expand or collapse server folder group |
| <kbd>m</kbd> | **Context Menu** | 🔒 Blocked* | Toggle hidden server (on server) / Group tmux menu (on group) |
| <kbd>Shift+H</kbd> | **Toggle Hidden** | ✅ Yes | Toggle visibility of hidden servers in the list |
| <kbd>p</kbd> | **Pin / Unpin** | 🔒 Blocked | Pin or unpin server to top of list |
| <kbd>Shift+P</kbd> / <kbd>Ctrl+G</kbd> | **Git SSH Profiles** | 🔒 Blocked | Open Git SSH Key Configuration & Profile Switcher |
| <kbd>t</kbd> | **Edit Tags** | 🔒 Blocked | Edit tags stored in SSH config comments |
| <kbd>Shift+T</kbd> | **Toggle Theme** | ✅ Yes | Cycle themes: Dark → Light → System |
| <kbd>c</kbd> | **Copy SSH Command** | ✅ Yes | Copy full connection command to system clipboard |
| <kbd>o</kbd> | **SCP Modal** | ✅ Yes | Open SCP command generator to copy upload/download commands |
| <kbd>Shift+M</kbd> | **SSHFS Modal** | ✅ Yes | Open SSHFS remote mount generator to copy mount commands |
| <kbd>Shift+C</kbd> | **Key Comment** | 🔒 Blocked | Inspect and edit SSH public/private key comment |
| <kbd>l</kbd> | **Agent Load** | 🔒 Blocked | Load server identity key into system `ssh-agent` |
| <kbd>u</kbd> | **Agent Unload** | 🔒 Blocked | Unload server identity key from `ssh-agent` |
| <kbd>v</kbd> | **Paste SSH Command** | 🔒 Blocked | Parse SSH command from clipboard into Add form |
| <kbd>y</kbd> | **Clone Server** | 🔒 Blocked | Duplicate server configuration with automatic alias deduplication |
| <kbd>Shift+K</kbd> | **Kill / Deploy Key** | 🔒 Blocked | Kill session (Active panel) / Run `ssh-copy-id` (Servers panel) |
| <kbd>f</kbd> | **Port Forwarding** | ✅ Yes | Open interactive SSH port forwarding & tunnel assistant |
| <kbd>Shift+F</kbd> | **Launch SFTP** | ✅ Yes | Launch configured SFTP / external file manager |
| <kbd>Ctrl+F</kbd> | **Dual-Pane SFTP** | ✅ Yes | Open built-in dual-pane SFTP manager directly |
| <kbd>s</kbd> | **Toggle Sort** | ✅ Yes | Cycle sort mode (Alias A-Z, Z-A, Last SSH) |
| <kbd>g</kbd> | **Ping Server** | ✅ Yes | Ping selected server and measure round-trip latency |
| <kbd>Shift+G</kbd> | **Ping All** | ✅ Yes | Parallel ping all servers with color-coded latency badges |
| <kbd>Shift+W</kbd> / <kbd>Ctrl+P</kbd> | **Ping Watch** | ✅ Yes | Toggle periodic background ping watch mode (default: 60s) |
| <kbd>Tab</kbd> / <kbd>Shift+Tab</kbd> | **Cycle Panels** | ✅ Yes | Move focus through Search, Servers, Active, Details panels |
| <kbd>0</kbd> | **Focus Search** | ✅ Yes | Jump focus directly to search bar |
| <kbd>1</kbd> | **Focus Servers** | ✅ Yes | Jump focus directly to servers list |
| <kbd>2</kbd> | **Focus Active** | ✅ Yes | Jump focus directly to active sessions list |
| <kbd>3</kbd> | **Focus Details** | ✅ Yes | Jump focus directly to details panel |
| <kbd>j</kbd> / <kbd>k</kbd> or <kbd>↓</kbd> / <kbd>↑</kbd> | **Navigate** | ✅ Yes | Move cursor up / down in server or session lists |
| <kbd>q</kbd> / <kbd>Ctrl+C</kbd> | **Quit** | ✅ Yes | Exit `neossh` |

---

## Read-Only / Viewer Mode (`-r`)

When launched with `--readonly` or `-r`:

- All modifying actions (<kbd>a</kbd>, <kbd>e</kbd>, <kbd>d</kbd>, <kbd>y</kbd>, <kbd>v</kbd>, <kbd>t</kbd>, <kbd>Shift+C</kbd>, <kbd>Shift+K</kbd>, <kbd>i</kbd>, <kbd>p</kbd>, <kbd>l</kbd>, <kbd>u</kbd>, <kbd>Shift+P</kbd>) are safely blocked.
- Attempting any restricted keybinding displays a clear informational dialog explaining that read-only mode is active.
- Safe exploration, search, SSH connecting, command copying, and monitoring operations remain fully accessible.
