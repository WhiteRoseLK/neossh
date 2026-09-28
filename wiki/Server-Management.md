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

## Multi-Alias Support

neossh preserves all space-separated aliases on a single `Host` line:

```ssh-config
Host web1 web2 staging
    HostName 192.168.1.10
    User deploy
```

All aliases are indexed for fuzzy search and connection. They are preserved verbatim on writeback.

## Active SSH Sessions Panel

Press <kbd>2</kbd> to focus the Active Sessions panel. This panel tracks running SSH and background sessions with:

- **Process inspection**: PID, forwarded ports, identity keys
- **One-touch termination**: Press <kbd>Shift+K</kbd> to kill a session
- **Config generation**: Press <kbd>a</kbd> to create a server entry from a running connection
