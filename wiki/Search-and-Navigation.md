#  Search & Navigation

## Fuzzy Search

Press <kbd>slash</kbd> or <kbd>0</kbd> to activate the search bar. Type to fuzzy search across server aliases, hostnames, IP addresses, usernames, and tags.

## Advanced Filter Syntax

Combine structured filters with free-text search for precise results:

| Filter | Example | Description |
|--------|---------|-------------|
| `tag:<name>` | `tag:prod`, `tag:k8s` | Filter by server tag |
| `-tag:<name>` | `-tag:staging` | Exclude servers with tag |
| `user:<username>` | `user:root`, `user:ubuntu` | Filter by username |
| `-user:<username>` | `-user:deploy` | Exclude servers with user |
| `host:<hostname>` | `host:192.168.`, `host:*.aws.*` | Filter by host/IP |
| `port:<number>` | `port:2222` | Filter by SSH port |
| `status:<state>` | `status:up`, `status:down`, `status:unknown` | Filter by ping status |
| `group:<name>` | `group:production` | Filter by server folder/group |

> [!NOTE]
> **Example: Combining filters**
    ```
    tag:prod user:root status:up web
    ```
    This finds servers tagged `prod`, with user `root`, that are online, and contain "web" in the alias or hostname.

### Aliases

- `status:online` = `status:up`
- `status:offline` = `status:down`
- `tags:<name>` = `tag:<name>`

## Panel Navigation

neossh has four main panels:

| Panel | Focus Key | Description |
|-------|:---------:|-------------|
| Search | <kbd>0</kbd> | Fuzzy search bar |
| Servers | <kbd>1</kbd> | Server list |
| Active Sessions | <kbd>2</kbd> | Running SSH sessions |
| Details | <kbd>3</kbd> | Selected server details |

Use <kbd>tab</kbd> / ++shift+tab<kbd> to cycle focus between panels.

## Sorting

Press </kbd>s++ to toggle sort mode:

1. **Alias** (A→Z)
2. **Alias** (Z→A)
3. **Last SSH** (most recent first)
4. **Last SSH** (oldest first)

The sort preference is persisted across sessions.

## CLI Pre-filtering

Pre-filter the server list at launch to avoid exposing your entire fleet:

```bash
# Positional argument
neossh prod

# Flag
neossh -f prod
```

> [!TIP]
> **Screen sharing safety**
    Use pre-filtering during screen shares or demos to show only relevant servers.

## Direct Connect

Bypass the TUI entirely and connect directly:

```bash
neossh -c my-server
```

Combine with other flags:

```bash
# Direct connect + exit on disconnect (one-shot launcher)
neossh -c -x my-server

# Direct connect with password
neossh -c my-server -P "password"
```
