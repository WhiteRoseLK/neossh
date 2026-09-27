---
title: SSHFS Remote Mounts
---

# :material-harddisk: SSHFS Remote Mounts

Mount remote server filesystems locally over SSH using SSHFS.

## TUI Usage

Press ++shift+m++ to open the SSHFS command generator modal. It generates ready-to-run mount and unmount commands with full SSH configuration:

- **Ports** and **identity files**
- **Jump proxies** (`ProxyJump`)
- **Auto-reconnect** options
- **Read-only** flags

The generated commands are copied to your clipboard.

## CLI Usage

```bash
neossh --sshfs web-prod
```

This generates and copies SSHFS mount/unmount command templates for the specified server alias.

## Example Output

```bash
# Mount
sshfs user@192.168.1.10:/remote/path /local/mountpoint \
  -p 22 \
  -o IdentityFile=~/.ssh/id_ed25519 \
  -o reconnect,ServerAliveInterval=15,ServerAliveCountMax=3

# Unmount
fusermount -u /local/mountpoint   # Linux
umount /local/mountpoint          # macOS
```

## Prerequisites

SSHFS must be installed on your system:

=== "macOS"

    ```bash
    brew install macfuse sshfs
    ```

=== "Linux (Debian/Ubuntu)"

    ```bash
    sudo apt install sshfs
    ```

=== "Linux (Arch)"

    ```bash
    sudo pacman -S sshfs
    ```
