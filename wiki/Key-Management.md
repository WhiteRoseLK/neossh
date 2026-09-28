#  Key Management

## SSH Key Autocomplete

When adding or editing a server, neossh automatically detects available keys in `~/.ssh/` and provides autocomplete suggestions for the `IdentityFile` field.

## Default SSH Identity Key

Configure a global default private key that auto-prefills the identity file on new servers:

```bash
# Set default key
neossh --default-key ~/.ssh/id_ed25519

# Check current default
neossh --default-key ""
```

Also configurable via:

- **Git & SSH Keys Setup dialog** (++shift+p<kbd> or </kbd>ctrl+g<kbd>)
- **Environment variable**: `NEOSSH_DEFAULT_KEY`
- **Settings file**: `~/.neossh/settings.json` → `"default_identity_key"`

## Git SSH Key Profiles

Press </kbd>shift+p<kbd> or </kbd>ctrl+g<kbd> to open the Git SSH Key Configuration & Profile Switcher:

- Configure **per-repository** or **global** Git SSH keys via `core.sshCommand`
- Supports GitHub, GitLab, and Bitbucket
- Switch between SSH key profiles for different Git hosting services

```bash
# CLI: configure Git SSH key for current repo
neossh --git-ssh ~/.ssh/id_ed25519_work

# Check current Git SSH configuration
neossh --git-ssh ""
```

## SSH Key Comment Editor

Press </kbd>shift+c<kbd> to inspect and directly edit public/private key comments on the selected server's identity file.

## SSH Agent Integration

| Key | Action |
|:---:|--------|
| </kbd>l<kbd> | Load selected server's key into `ssh-agent` |
| </kbd>u<kbd> | Unload selected server's key from `ssh-agent` |

## SSH Key Type Badges & FIDO2

Server details display visual badges for key information:

| Badge | Meaning |
|-------|---------|
| `[ED25519]` | Ed25519 key algorithm |
| `[RSA-4096]` | RSA key with 4096-bit modulus |
| `[ECDSA]` | ECDSA key algorithm |
| `[FIDO2]` | FIDO2/security key (YubiKey, SoloKey) |
| ✓ found | Key file exists on disk |
| ⚠ missing | Key file not found |

Badges are detected from public key file headers.

## One-Touch SSH Key Deployment

Press </kbd>shift+k++ (on the Servers panel) to automatically push your public SSH key to the remote host using `ssh-copy-id`.

> [!NOTE]
    This action is disabled in read-only mode (`--readonly` / `-r`).
