# Server Management

## Viewing Servers

neossh reads your `~/.ssh/config` (and all `Include` files) and displays servers in a clean, scrollable list. Wildcard patterns (`Host *.corp`, `Host *`) are displayed with `[wildcard]` badges and protected from accidental direct connections.

## Adding Servers

Press <kbd>a</kbd> to open the add server form. The form provides a comprehensive tabbed interface with all SSH configuration options:

- **Basic**: Alias, hostname, user, port, identity file
- **Connection**: Proxy settings, connection multiplexing, forwarding
- **Security**: Ciphers, MACs, key exchange algorithms
- **Advanced**: Pre-connect hooks, certificate commands, passwords

> [!TIP]
> **Default identity key**
> Configure a global default key with `neossh --default-key ~/.ssh/id_ed25519` to auto-prefill the identity file for new servers.

## Editing Servers

Press <kbd>e</kbd> on any server to open the edit form. All changes are written back to the original source file (respecting `Include` directives).

## Deleting Servers

Press <kbd>d</kbd> to delete with a confirmation dialog. The deletion modifies only the file that defines the host.

## Pin / Unpin

Press <kbd>p</kbd> to pin a server to the top of the list. Pinned servers are persisted in `metadata.json` and as `# pin` comments in SSH config.

## Hidden Hosts

Some hosts (jump hosts, proxy targets, internal nodes) are better kept out of the main list:

- Press <kbd>m</kbd> on a server to toggle hidden/visible
- Press <kbd>Shift+H</kbd> to toggle showing hidden servers in the list
- Launch with `neossh -H` to start with hidden servers visible

Hidden status is stored as `# hidden` in your SSH config.

## Duplicate / Clone

Press <kbd>y</kbd> or <kbd>Shift+C</kbd> to clone a server configuration. neossh automatically deduplicates the alias (`srv` → `srv_1` → `srv_2`).

## Paste SSH Command

Press <kbd>v</kbd> to parse an SSH command from your clipboard into the add server form. neossh intelligently extracts flags, identity keys, ports, jump hosts, and deduces an alias.

## Server Folders & Groups

Organize servers into collapsible folders using hierarchical aliases:

```
Host prod/database/primary
Host prod/database/replica
Host prod/web/frontend
Host staging/api
```

- Press <kbd>Space</kbd> or <kbd>Enter</kbd> on a group header to toggle collapse/expand
- Press <kbd>m</kbd> on a group header for the group menu:
    - **Tmux Connect All**: Launch tmux with panes connected to all servers in the group
    - **Collapse All / Expand All**: Batch toggle all groups

## Multi-Server Selection & Bulk Operations

Select multiple servers in the main server list to perform operations in bulk:

- **Checkbox Toggle**: Press <kbd>Space</kbd> on any focused server to toggle selection (`[✓]` / `[ ]`).
- **Select All**: Press <kbd>Ctrl+A</kbd> or <kbd>*</kbd> to select all servers in the current view, or toggle selection across all servers in a focused group.
- **Clear Selection**: Press <kbd>Esc</kbd> to deselect all servers (pressing <kbd>Esc</kbd> again returns focus to the search bar).
- **Status Counter**: A dynamic badge in the status bar displays the count of selected servers (e.g. `[3 servers selected]`).
- **Bulk Ping**: Press <kbd>Shift+G</kbd> to ping only the selected servers concurrently (falls back to pinging all servers when none are selected).
- **Bulk Tagging**: Press <kbd>t</kbd> when multiple servers are selected to open the bulk tag modal and add/remove tags across all selected servers simultaneously.
- **Multi-Session Engine & Tabs**: Press <kbd>Enter</kbd> when multiple servers are selected (or run command snippets via <kbd>X</kbd>) to open the interactive Multi-Session Dashboard:
  - **Tab 0 (Overview Dashboard)**: Aggregated status table with real-time execution progress (`⏳ Running`, `✓ Success`, `✗ Failed`), durations, and output summaries.
  - **Interactive Drill-Down**: Press <kbd>Enter</kbd> on any single server row to jump immediately to its dedicated session tab with scrollable output and live interactive shell capability (<kbd>i</kbd>).
  - **Split View Tab**: Check 2 to 4 servers with <kbd>Space</kbd> and press <kbd>Enter</kbd> to open a multi-pane split view (side-by-side or 2x2 grid) with <kbd>Tab</kbd> navigation.
  - **Tab Navigation**: Jump directly to tabs using <kbd>Alt+0</kbd> (Dashboard) and <kbd>Alt+1..9</kbd> (individual session tabs).

## Multi-Alias Support

neossh preserves all space-separated aliases on a single `Host` line:

```ssh-config
Host web1 web2 staging
    HostName 192.168.1.10
    User deploy
```

All aliases are indexed for fuzzy search and connection. They are preserved verbatim on writeback.

## Dotfiles Synchronization (chezmoi)

Deploy your personal configuration files and shell environment (`.bashrc`, `.zshrc`, `.tmux.conf`, `.vimrc`, etc.) to remote servers seamlessly with zero remote installation:

- **1-Click Sync (<kbd>D</kbd>)**: Highlight any server in the Servers panel and press <kbd>D</kbd> (or <kbd>Shift+D</kbd>). A confirmation dialog displays the target server alias and details, streaming your local `chezmoi archive` directly over SSH into `tar -xf - -C ~`.
- **Automatic Sync on Connect**: In the Add/Edit server form, set `SyncDotfilesOnConnect` to `yes` (persisted as `# sync-dotfiles: true` in your SSH configuration). Whenever you connect (<kbd>Enter</kbd>), neossh synchronizes your dotfiles before starting the interactive shell.
- **Zero Remote Dependencies**: The remote host only needs the standard `tar` command (available on every Unix-like system). `chezmoi` is only required locally.

For complete details, see [[Dotfiles Sync (chezmoi)|Dotfiles-Sync]].

## Active SSH Sessions Panel

Press <kbd>2</kbd> to focus the Active Sessions panel. This panel tracks running SSH and background sessions with:

- **Process inspection**: PID, forwarded ports, identity keys
- **One-touch termination**: Press <kbd>Shift+K</kbd> to kill a session
- **Config generation**: Press <kbd>a</kbd> to create a server entry from a running connection
