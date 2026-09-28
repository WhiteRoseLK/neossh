# SFTP & File Transfer

## Built-in Dual-Pane SFTP Manager

Press <kbd>Ctrl+F</kbd> (or <kbd>Shift+F</kbd> to use the configured file manager) to open the interactive dual-pane file manager:

- **WinSCP / FileZilla style** with local (left) and remote (right) panels
- Directory navigation with arrow keys
- Sort by name, size, or date with <kbd>s</kbd>
- Upload files with <kbd>u</kbd>, download with <kbd>d</kbd>
- Real-time streaming transfer progress indicators
- Overwrite confirmation dialogs

## Quick-Launch from CLI

```bash
# Launch with configured file manager (default: sftp)
neossh --sftp web-prod

# Override the file manager tool
neossh --sftp web-prod --file-manager yazi
neossh --sftp web-prod --file-manager filezilla
```

## Configurable File Managers

Configure your preferred file manager in `~/.neossh/settings.json`:

```json
{
  "file_manager": "yazi"
}
```

### Supported Tools

| Tool | Type | Description |
|------|------|-------------|
| `sftp` | Terminal | Standard OpenSSH SFTP client (default) |
| `internal` | TUI | Built-in dual-pane manager |
| `yazi` | Terminal | Modern terminal file manager |
| `ranger` | Terminal | VIM-inspired file manager |
| `filezilla` | GUI | FileZilla SFTP client |
| `cyberduck` | GUI | Cyberduck file transfer |
| `nautilus` | GUI | GNOME file manager |
| `dolphin` | GUI | KDE file manager |
| Custom template | Any | Your own command with placeholders |

### Custom Command Templates

Use placeholders in custom commands:

| Placeholder | Value |
|-------------|-------|
| `%a` | Server alias |
| `%h` | Hostname |
| `%u` | Username |
| `%p` | Port |
| `%url` | `sftp://[user@]host[:port]/` URL |
| `%fish_url` | `fish://[user@]host[:port]/` URL |
| `%c` | SSH config file path |

Example custom template:

```json
{
  "file_manager": "my-sftp-tool --host %h --user %u --port %p"
}
```

## TUI Shortcuts

| Key | Action |
|:---:|--------|
| <kbd>Shift+F</kbd> | Launch configured file manager (external or internal) |
| <kbd>Ctrl+F</kbd> | Open built-in dual-pane SFTP manager directly |

> [!TIP]
> **Terminal vs GUI tools**
> Terminal tools (sftp, yazi, ranger) suspend the TUI during use. GUI tools (filezilla, cyberduck, nautilus, dolphin) launch in the background while the TUI stays active.

## Protocol URLs

neossh automatically generates standard protocol URLs:

- **SFTP**: `sftp://[user@]host[:port]/`
- **Fish** (KDE Dolphin): `fish://[user@]host[:port]/`
