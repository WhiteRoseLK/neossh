# Security

neossh treats credential security and user privacy as paramount requirements.

## Core Security Principles

### No Plaintext Passwords on Disk

Passwords are **never** stored in plain text anywhere on disk — neither in `~/.ssh/config` nor in `metadata.json`. Your SSH configuration remains clean, portable, and completely safe to version-control in dotfiles.

### Native OS Keyring & Hardware Security

Server passwords configured for automated `sshpass` login are managed by the operating system's native secure credential store:

| Platform | Credential Store |
|----------|-----------------|
| **macOS** | Keychain |
| **Linux** | Secret Service / DBus |
| **Windows** | Credential Manager |

If the OS keyring is unavailable, neossh falls back to an authenticated **AES-256-GCM local vault** (`~/.neossh/vault.json`) with strict `0600` file permissions.

### Process Table Protection

When connecting via `sshpass`, neossh delivers passwords through the `SSHPASS` environment variable (`sshpass -e`) rather than command-line arguments, preventing password exposure to other users via `ps aux`.

### OpenSSH Native Execution

All SSH connections run through your system's native `ssh` binary (OpenSSH). Your existing `IdentityFile` paths, passphrases, and `ssh-agent` integrations work exactly as before. neossh never implements its own SSH protocol layer.

### Strict File Permissions

Config snapshots, backups, and vault files strictly enforce `0600` permissions, ensuring only the file owner has read/write access.

## Export & Sharing Safety

When using `neossh export`, the following items are **never** included in bundles:

| Item | Reason |
|------|--------|
| Private key files | Only `IdentityFile` path references appear in SSH config |
| `vault.json` | Contains encrypted passwords |
| OS keyring contents | Platform-specific, non-portable |

> [!TIP]
> **Sanitized exports for teams**
> Use `neossh export --sanitize` to additionally strip `IdentityFile` paths and sensitive comments (`# password:`, `# token:`) from the bundle, making it safe for team distribution.

## Config Safety: Non-destructive Writes

- **Minimal edits**: neossh only writes the minimal required changes to your `~/.ssh/config`. Comments, spacing, indentation, and untouched settings remain intact.
- **Atomic writes**: Updates are written to a temporary file and atomically renamed over the original to prevent corruption.
- **Backups**:
    - *One-time snapshot*: Before neossh makes its first change, it creates `config.original.backup`. This file is never overwritten.
    - *Rolling backups*: On each save, neossh creates a timestamped backup (`~/.ssh/config-<timestamp>-neossh.backup`), keeping the 10 most recent.

## Reporting Security Issues

If you discover a security vulnerability, please report it responsibly by opening a [GitHub Issue](https://github.com/WhiteRoseLK/neossh/issues) or contacting the maintainers directly.
