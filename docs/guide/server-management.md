---
title: Server Management
---

# :material-server: Server Management

## Viewing Servers

neossh reads your `~/.ssh/config` (and all `Include` files) and displays servers in a clean, scrollable list. Wildcard patterns (`Host *.corp`, `Host *`) are displayed with `[wildcard]` badges and protected from accidental direct connections.

## Adding Servers

Press ++a++ to open the add server form. The form provides a comprehensive tabbed interface with all SSH configuration options:

- **Basic**: Alias, hostname, user, port, identity file
- **Connection**: Proxy settings, connection multiplexing, forwarding
- **Security**: Ciphers, MACs, key exchange algorithms
- **Advanced**: Pre-connect hooks, certificate commands, passwords

!!! tip "Default identity key"
    Configure a global default key with `neossh --default-key ~/.ssh/id_ed25519` to auto-prefill the identity file for new servers.

## Editing Servers

Press ++e++ on any server to open the edit form. All changes are written back to the original source file (respecting `Include` directives).

## Deleting Servers

Press ++d++ to delete with a confirmation dialog. The deletion modifies only the file that defines the host.

## Pin / Unpin

Press ++p++ to pin a server to the top of the list. Pinned servers are persisted in `metadata.json` and as `# pin` comments in SSH config.

## Hidden Hosts

Some hosts (jump hosts, proxy targets, internal nodes) are better kept out of the main list:

- Press ++m++ on a server to toggle hidden/visible
- Press ++shift+h++ to toggle showing hidden servers in the list
- Launch with `neossh -H` to start with hidden servers visible

Hidden status is stored as `# hidden` in your SSH config.

## Duplicate / Clone

Press ++y++ or ++shift+c++ to clone a server configuration. neossh automatically deduplicates the alias (`srv` → `srv_1` → `srv_2`).

## Paste SSH Command

Press ++v++ to parse an SSH command from your clipboard into the add server form. neossh intelligently extracts flags, identity keys, ports, jump hosts, and deduces an alias.

## Server Folders & Groups

Organize servers into collapsible folders using hierarchical aliases:

```
Host prod/database/primary
Host prod/database/replica
Host prod/web/frontend
Host staging/api
```

- Press ++space++ or ++enter++ on a group header to toggle collapse/expand
- Press ++m++ on a group header for the group menu:
    - **Tmux Connect All**: Launch tmux with panes connected to all servers in the group
    - **Collapse All / Expand All**: Batch toggle all groups

## Multi-Alias Support

neossh preserves all space-separated aliases on a single `Host` line:

```ssh-config
Host web1 web2 staging
    HostName 192.168.1.10
    User deploy
```

All aliases are indexed for fuzzy search and connection. They are preserved verbatim on writeback.

## Active SSH Sessions Panel

Press ++2++ to focus the Active Sessions panel. This panel tracks running SSH and background sessions with:

- **Process inspection**: PID, forwarded ports, identity keys
- **One-touch termination**: Press ++shift+k++ to kill a session
- **Config generation**: Press ++a++ to create a server entry from a running connection
