#  Keybindings Reference

This reference documents all keyboard shortcuts available inside the `neossh` interactive terminal user interface (TUI).

---

## Complete Keybindings Table

| Key | Action | Read-Only Safe? | Description |
|:---:|---|:---:|---|
| <kbd>enter</kbd> | **SSH Connect** |  | Open native SSH connection to selected server |
| <kbd>slash</kbd> | **Search** |  | Open fuzzy search input bar |
| <kbd>a</kbd> | **Add Server** |  | Open modal form to add a new server entry |
| <kbd>e</kbd> | **Edit Server** |  | Open modal form to edit selected server configuration |
| <kbd>d</kbd> | **Delete Server** |  | Safely delete selected server with confirmation dialog |
| <kbd>i</kbd> | **Import Known Hosts** |  | Import discovered hosts from `~/.ssh/known_hosts` |
| <kbd>space</kbd> | **Toggle Group** |  | Expand or collapse server folder group |
| <kbd>m</kbd> | **Context Menu** | * | Toggle hidden server (on server) / Group tmux menu (on group) |
| ++shift+h<kbd> | **Toggle Hidden** |  | Toggle visibility of hidden servers in the list |
| </kbd>p<kbd> | **Pin / Unpin** |  | Pin or unpin server to top of list |
| </kbd>shift+p<kbd> / </kbd>ctrl+g<kbd> | **Git SSH Profiles** |  | Open Git SSH Key Configuration & Profile Switcher |
| </kbd>t<kbd> | **Edit Tags** |  | Edit tags stored in SSH config comments |
| </kbd>shift+t<kbd> | **Toggle Theme** |  | Cycle themes: Dark → Light → System |
| </kbd>c<kbd> | **Copy SSH Command** |  | Copy full connection command to system clipboard |
| </kbd>o<kbd> | **SCP Modal** |  | Open SCP command generator to copy upload/download commands |
| </kbd>shift+m<kbd> | **SSHFS Modal** |  | Open SSHFS remote mount generator to copy mount commands |
| </kbd>shift+c<kbd> | **Key Comment** |  | Inspect and edit SSH public/private key comment |
| </kbd>l<kbd> | **Agent Load** |  | Load server identity key into system `ssh-agent` |
| </kbd>u<kbd> | **Agent Unload** |  | Unload server identity key from `ssh-agent` |
| </kbd>v<kbd> | **Paste SSH Command** |  | Parse SSH command from clipboard into Add form |
| </kbd>y<kbd> | **Clone Server** |  | Duplicate server configuration with automatic alias deduplication |
| </kbd>shift+k<kbd> | **Kill / Deploy Key** |  | Kill session (Active panel) / Run `ssh-copy-id` (Servers panel) |
| </kbd>f<kbd> | **Port Forwarding** |  | Open interactive SSH port forwarding & tunnel assistant |
| </kbd>shift+f<kbd> | **Launch SFTP** |  | Launch configured SFTP / external file manager |
| </kbd>ctrl+f<kbd> | **Dual-Pane SFTP** |  | Open built-in dual-pane SFTP manager directly |
| </kbd>s<kbd> | **Toggle Sort** |  | Cycle sort mode (Alias A-Z, Z-A, Last SSH) |
| </kbd>g<kbd> | **Ping Server** |  | Ping selected server and measure round-trip latency |
| </kbd>shift+g<kbd> | **Ping All** |  | Parallel ping all servers with color-coded latency badges |
| </kbd>shift+w<kbd> / </kbd>ctrl+p<kbd> | **Ping Watch** |  | Toggle periodic background ping watch mode (default: 60s) |
| </kbd>tab<kbd> / </kbd>shift+tab<kbd> | **Cycle Panels** |  | Move focus through Search, Servers, Active, Details panels |
| </kbd>0<kbd> | **Focus Search** |  | Jump focus directly to search bar |
| </kbd>1<kbd> | **Focus Servers** |  | Jump focus directly to servers list |
| </kbd>2<kbd> | **Focus Active** |  | Jump focus directly to active sessions list |
| </kbd>3<kbd> | **Focus Details** |  | Jump focus directly to details panel |
| </kbd>j<kbd> / </kbd>k<kbd> or </kbd>arrow-down<kbd> / </kbd>arrow-up<kbd> | **Navigate** |  | Move cursor up / down in server or session lists |
| </kbd>q<kbd> / </kbd>ctrl+c<kbd> | **Quit** |  | Exit `neossh` |

---

## Read-Only / Viewer Mode (`-r`)

When launched with `--readonly` or `-r`:

- All modifying actions (</kbd>a<kbd>, </kbd>e<kbd>, </kbd>d<kbd>, </kbd>y<kbd>, </kbd>v<kbd>, </kbd>t<kbd>, </kbd>shift+c<kbd>, </kbd>shift+k<kbd>, </kbd>i<kbd>, </kbd>p<kbd>, </kbd>l<kbd>, </kbd>u<kbd>, </kbd>shift+p++) are safely blocked.
- Attempting any restricted keybinding displays a clear informational dialog explaining that read-only mode is active.
- Safe exploration, search, SSH connecting, command copying, and monitoring operations remain fully accessible.
