---
title: Backup & Export
---

# :material-package-variant: Backup & Export

## Export Configuration Bundle

Archive your SSH configuration into a single compressed `.tar.gz` bundle:

```bash
neossh export
# or
neossh backup
```

### What's Included

| Item | Description |
|------|-------------|
| `~/.ssh/config` | Main SSH configuration file |
| All `Include` files | Recursively resolved included config files |
| `~/.neossh/metadata.json` | Usage statistics (SSHCount, LastSSH) |
| `~/.neossh/settings.json` | UI preferences (theme, file_manager, etc.) |

### Options

```bash
# Specify output path
neossh export --output ~/backups/ssh-config.tar.gz

# Sanitized export (safe for sharing)
neossh export --sanitize
neossh export -s
```

## Sanitized Sharing

The `--sanitize` flag strips sensitive data for safe team distribution:

- Removes `IdentityFile` private key path references
- Strips sensitive comments (`# password:`, `# token:`)
- Produces a clean bundle safe for dotfiles or team sharing

!!! warning "What is NEVER included (even without sanitize)"
    - **Private key files** — only `IdentityFile` path references appear in config
    - **vault.json** — encrypted password vault
    - **OS keyring contents** — platform-specific credentials

## Bundle Verification

Validate a bundle's integrity without modifying any files:

```bash
# Verify command
neossh verify backup.tar.gz

# Or dry-run import
neossh import --dry-run backup.tar.gz
```

Verification checks:

- Bundle integrity
- Schema version compatibility
- SHA-256 checksums for each file

## Import / Restore

Restore configuration from a bundle:

```bash
neossh import backup.tar.gz
# or
neossh restore backup.tar.gz
```

### Safety Features

- **Automatic backups**: Existing files are backed up with `.bak.<timestamp>` extensions before being overwritten
- **Custom directories**: Restore to alternative locations if needed
- **Dry-run preview**: Use `--dry-run` to preview changes without modifying files
