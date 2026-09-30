# Configuration Files Reference

`neossh` is designed around standard OpenSSH configurations while managing auxiliary state (stats, preferences, secrets) in dedicated, cleanly separated files.

---

## File Layout Overview

```
~/.ssh/
├── config                               # Primary OpenSSH configuration
├── config.original.backup               # Permanent snapshot created before first write
├── config-<timestamp>-neossh.backup     # Rolling backup (keeps 10 most recent)
└── config.d/                            # Included config files (if using Include)

~/.neossh/
├── metadata.json                        # Host usage metrics (SSH counts, last connected)
├── settings.json                        # User UI preferences, default keys, tunnel profiles
└── vault.json                           # AES-256-GCM local encrypted password vault (0600)
```

*(If `$XDG_CONFIG_HOME` and `$XDG_STATE_HOME` are set, `neossh` conforms to the XDG Base Directory Specification).*

---

## 1. `~/.ssh/config`

The single source of truth for host connections. `neossh` parses, edits, and writes back minimal modifications preserving:
- All standard OpenSSH options (`HostName`, `User`, `Port`, `IdentityFile`, `ProxyJump`, etc.)
- Indentation, spacing, and comments
- Top-level `Include` directives with glob expansions

### Embedded Directives

`neossh` embeds metadata within comments:

```ssh-config
Host prod-db # tags: prod, database # pin # sync-dotfiles: true
    HostName 10.0.1.5
    User postgres
    # pre-connect: /usr/local/bin/check-vpn.sh %h
    # certificate-command: step ssh login %u@%h
    # sync-dotfiles: true
```

---

## 2. `~/.neossh/settings.json`

Stores application preferences, configured tools, and saved tunnel profiles:

```json
{
  "theme": "dark",
  "file_manager": "yazi",
  "auto_ping": true,
  "default_identity_key": "~/.ssh/id_ed25519",
  "first_run_completed": true,
  "tunnel_profiles": {
    "prod-db": [
      {
        "name": "Local PostgreSQL",
        "type": "local",
        "local_port": 5432,
        "remote_host": "localhost",
        "remote_port": 5432
      }
    ]
  }
}
```

### Configurable Keys

| Setting | Type | Description | Default |
|---|---|---|---|
| `theme` | string | Color theme: `"dark"`, `"light"`, `"system"` | `"dark"` |
| `file_manager` | string | SFTP file manager tool (`"internal"`, `"sftp"`, `"yazi"`, `"ranger"`, `"filezilla"`, or custom template) | `"sftp"` |
| `auto_ping` | boolean | Automatically trigger parallel background pings at startup | `false` |
| `default_identity_key` | string | Path to default private key prefilled on new server forms | `""` |
| `first_run_completed` | boolean | Tracks whether the initial companion tools onboarding wizard has run | `false` |
| `tunnel_profiles` | object | Saved port forwarding profiles mapped by server alias | `{}` |
| `keybindings` | object | Custom keybinding overrides for actions (e.g. `{"add_server": "n", "clone_server": "c"}`) | `{}` |

#### Custom Keybindings Example (e.g. AZERTY / custom layouts):

```json
{
  "keybindings": {
    "add_server": "n",
    "clone_server": "c",
    "copy_command": "y",
    "quit": "x"
  }
}
```

---

## 3. `~/.neossh/metadata.json`

Tracks non-sensitive historical metrics:

```json
{
  "prod-db": {
    "ssh_count": 42,
    "last_ssh": "2026-09-27T17:30:00Z"
  }
}
```

---

## 4. `~/.neossh/vault.json`

When native OS credential keyrings (macOS Keychain, Linux Secret Service, Windows Credential Manager) are unavailable or fail, passwords configured for `sshpass` are encrypted with AES-256-GCM and stored in `vault.json`.

- File permissions are strictly enforced to `0600` (read/write only by the user).
- **Never included in bundle exports (`neossh export`)**.
- Does NOT store private keys — only encrypted password strings.

---

## 5. `~/.config/neossh/snippets.json`

Stores the command snippets library with parameter placeholders (`{{param}}` or `<param>`), descriptions, and categorization tags.

```json
[
  {
    "id": "snip-sys-uptime",
    "name": "System Load & Uptime",
    "command": "uptime && free -h",
    "description": "Inspect remote load average and memory",
    "tags": ["sysadmin", "monitoring"]
  },
  {
    "id": "snip-systemd-restart",
    "name": "Restart Systemd Service",
    "command": "sudo systemctl restart {{service}}",
    "description": "Prompt for service name and restart daemon",
    "tags": ["systemd", "ops"]
  }
]
```

- Permissions are strictly set to `0600`.
- Override with the `NEOSSH_SNIPPETS_FILE` environment variable.
- Defaults are automatically populated on first run if no snippets file exists.

---

## 6. Backups

To safeguard against data loss, `neossh` creates backups automatically before making any changes:

- **`config.original.backup`**: Created the first time `neossh` ever writes to your `~/.ssh/config`. It is permanently preserved and never modified.
- **`config-<timestamp>-neossh.backup`**: Created immediately before every write operation. `neossh` rotates these rolling backups and keeps the 10 most recent.

