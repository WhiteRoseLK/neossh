---
title: SSH Configuration
---

# :material-cog: SSH Configuration

## Config Include Support

neossh honours top-level `Include` directives in your `~/.ssh/config`:

- **Reads**: All included files are parsed in OpenSSH precedence order.
- **Globs**: Wildcard patterns are expanded like OpenSSH; matches that are directories are skipped instead of aborting startup.
- **Writes route back to source**: Editing or deleting a host modifies the file that actually defines it. Other files are never touched.
- **Ambiguity modal**: If the same alias is defined in multiple files, a prompt asks which file to update. Your choice is remembered in `metadata.json`.

```ssh-config
# ~/.ssh/config
Include config.d/*
Include config_work

Host personal-server
    HostName 192.168.1.10
    User me
```

## Config Safety

### Non-destructive Writes

neossh only writes the minimal required changes to your `~/.ssh/config`. Comments, spacing, indentation, and untouched settings remain intact.

### Atomic Writes

Updates are written to a temporary file and atomically renamed over the original to prevent corruption on crashes or power loss.

### Automatic Backups

| Backup Type | File | Description |
|-------------|------|-------------|
| One-time snapshot | `config.original.backup` | Created before first change. Never overwritten. |
| Rolling backups | `~/.ssh/config-<timestamp>-neossh.backup` | On each save. Keeps 10 most recent. |

## Advanced SSH Options

The add/edit form provides a comprehensive tabbed interface:

### Connection

- Port forwarding (`LocalForward`, `RemoteForward`, `DynamicForward`)
- Connection multiplexing for instant subsequent connections
- Keep-alive settings

### Authentication

- Public key authentication
- Password authentication (via `sshpass`)
- Agent forwarding
- FIDO2 / security keys

### Security

- Ciphers
- MACs (message authentication codes)
- Key exchange algorithms
- Host key algorithms

### Proxy

- `ProxyJump` — Modern multi-hop proxy
- `ProxyCommand` — Custom proxy commands

## Portable Tilde Paths

neossh automatically normalizes absolute paths to portable relative tilde paths:

```
/Users/john/.ssh/id_rsa  →  ~/.ssh/id_rsa
```

This ensures SSH configs remain portable across machines with different usernames.

## Quoted Host Alias Stripping

neossh automatically sanitizes enclosing quotes from `Host` lines:

```ssh-config
Host "server"  →  Host server
```

## Numeric Username Support

neossh supports usernames starting with a number or underscore (e.g. `007admin`), fully supporting Linux service and UID-based accounts.

## Custom Config Path

Load any alternative SSH config file:

```bash
neossh --sshconfig ~/.ssh/config_work
```

This is useful for separating personal and work configurations.
