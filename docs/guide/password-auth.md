---
title: Password Authentication
---

# :material-form-textbox-password: Password Authentication via `sshpass`

For legacy servers or restricted environments that do not support SSH public key authentication, `neossh` provides automated password delivery via `sshpass` with end-to-end credential security.

## Security Architecture

`neossh` treats credential security and user privacy as paramount requirements:

- **Zero Plaintext Footprint in `~/.ssh/config`**: Passwords are **never** stored in plain text anywhere on disk — neither in `~/.ssh/config` nor in `metadata.json`. Your SSH configuration remains clean, portable, and completely safe to version-control in public or shared dotfiles.
- **Hardware-Backed Credential Store**: Passwords configured via the Add/Edit Server form under `▶ Password & Interactive` are securely stored in your operating system's native keyring (macOS Keychain, Linux Secret Service / DBus, Windows Credential Manager) with an authenticated AES-256-GCM local vault fallback (`~/.neossh/vault.json` with strict `0600` permissions).
- **Secure Process Invocation**: Uses `sshpass -e` with environment variable delivery (`SSHPASS`) rather than command-line arguments, completely eliminating visibility in system process tables (`ps aux`).
- **Strict File Permissions**: Config snapshots, backups, and vault files strictly enforce `0600` permissions.

!!! warning "Security Best Practice"
    Whenever possible, prefer SSH public key authentication or SSH certificates over password authentication. Password delivery via `sshpass` should primarily be used for legacy appliances, embedded devices, or temporary environments where key provisioning is restricted.

---

## Prerequisites

To use automated password authentication, `sshpass` must be installed on your system:

=== "macOS"

    ```bash
    # Via Homebrew (community tap or source)
    brew install esolitos/ipa/sshpass
    ```

=== "Ubuntu / Debian"

    ```bash
    sudo apt-get update && sudo apt-get install -y sshpass
    ```

=== "Arch Linux"

    ```bash
    sudo pacman -S sshpass
    ```

=== "Fedora / RHEL"

    ```bash
    sudo dnf install -y sshpass
    ```

---

## Configuring Passwords in the TUI

1. Highlight the target server and press ++e++ (or ++a++ when adding a new server).
2. Navigate to the **Authentication** tab or the `▶ Password & Interactive` section.
3. Enter your password in the secure password input field (characters are masked).
4. Save the form.
5. `neossh` securely persists the password into your OS keyring or encrypted local vault. The server entry in `~/.ssh/config` will NOT contain the password.

---

## CLI Usage & Environment Variables

You can supply passwords directly from the command line or via environment variables for one-shot connections or automated scripts:

### One-Shot CLI Flag

```bash
# Connect directly with one-shot password authentication
neossh -c srv-legacy -P "SuperSecretPassword"

# Or using long flag
neossh -c srv-legacy --password "SuperSecretPassword"
```

### Environment Variables

```bash
# Using NEOSSH_PASSWORD
export NEOSSH_PASSWORD="SuperSecretPassword"
neossh -c srv-legacy

# Standard SSHPASS is also honored
export SSHPASS="SuperSecretPassword"
neossh -c srv-legacy
```

---

## Export & Backup Safety

When exporting your configuration bundles (`neossh export`), password credentials are **never** included:

- The OS keyring contents are platform-specific and excluded.
- The AES-256-GCM vault (`~/.neossh/vault.json`) is explicitly omitted from export bundles.
- You can safely share exported bundles (especially with `--sanitize`) without leaking authentication credentials.
