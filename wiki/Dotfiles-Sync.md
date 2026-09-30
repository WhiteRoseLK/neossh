# Dotfiles Synchronization (chezmoi)

`neossh` features seamless, 1-click personal dotfiles synchronization to remote servers powered by [chezmoi](https://www.chezmoi.io/).

With this feature, your preferred shell configuration (`.bashrc`, `.zshrc`, `.profile`), terminal tools (`.tmux.conf`, `.vimrc`), and aliases are instantly deployed to any remote machine without manually cloning git repositories or configuring tooling remotely.

---

## ⚡ Key Highlights & Architecture

- **Zero Remote Dependencies**: The remote server **does not** need `chezmoi`, `git`, `curl`, or package managers installed. It only requires the POSIX-standard `tar` command, available on virtually every Unix, Linux, BSD, and macOS server.
- **Streaming Pipeline**: `neossh` generates an in-memory tarball on your local machine using `chezmoi archive` and streams it over the existing encrypted SSH connection directly into `tar -xf - -C ~`. No intermediate archive files are ever written to remote disk.
- **Seamless Authentication**: Works over all supported SSH authentication methods, including SSH keys, FIDO2 security keys, SSH agent, jump hosts (`ProxyJump`), and automated password authentication (`sshpass`).
- **Safety First**: Fully protected by read-only mode (`-r` / `--readonly`) and interactive confirmation modals to prevent accidental overwrites.

---

## 🚀 Manual Synchronization (<kbd>D</kbd>)

To synchronize your dotfiles to a server on demand:

1. In the **Servers** panel, highlight the desired server.
2. Press <kbd>D</kbd> (or <kbd>Shift+D</kbd>).
3. A confirmation dialog appears displaying the target server alias, hostname, and destination directory (`~`).
4. Press <kbd>Enter</kbd> to confirm.
5. `neossh` executes the streaming sync in the background and displays a real-time status notification upon completion:
   - `✓ Dotfiles successfully synchronized to <server>`

> [!NOTE]
> If `chezmoi` is not installed on your local workstation, `neossh` displays a friendly alert dialog prompting you to run `neossh --setup` to install `chezmoi` automatically via your native package manager.

---

## 🔄 Automatic Sync on Connect

You can configure specific servers to automatically receive your dotfiles every time you establish an SSH connection.

### Via the TUI (Add / Edit Form)

1. Highlight the server and press <kbd>e</kbd> (or press <kbd>a</kbd> when creating a new server).
2. Scroll to the **SyncDotfilesOnConnect** field.
3. Select `yes` from the dropdown list.
4. Save the form (<kbd>Ctrl+S</kbd> or submit button).

### Via `~/.ssh/config` Comment Tag

`neossh` stores this setting as a standard comment directive in your `~/.ssh/config`:

```ssh-config
# Inline on Host line:
Host dev-server # sync-dotfiles: true
    HostName 192.168.1.50
    User ubuntu

# Or inside the Host block:
Host prod-bastion
    HostName bastion.example.com
    User admin
    # sync-dotfiles: true
```

Whenever you connect to this server via <kbd>Enter</kbd> or `neossh -c dev-server`, `neossh` synchronizes your dotfiles immediately before launching the interactive terminal session.

---

## 📋 Prerequisites

| Component | Requirement | Notes |
|---|---|---|
| **Local Machine** | `chezmoi` binary in `$PATH` | Install via `neossh --setup` or package manager (`brew`, `apt`, `pacman`, etc.) |
| **Local Dotfiles** | Initialized chezmoi state | Run `chezmoi init` (and optionally `chezmoi add ~/.bashrc`, etc.) |
| **Remote Server** | `tar` binary in `$PATH` | Standard utility present on all standard Linux, BSD, and macOS distributions |

---

## 🔒 Security & Read-Only Protection

- **Read-Only Mode (`-r`)**: When `neossh` runs in read-only mode, the <kbd>D</kbd> keybinding is strictly blocked with an informative dialog to prevent unauthorized changes to remote hosts.
- **Sensitive Variables**: Because `chezmoi archive` is evaluated on your local trusted machine, chezmoi templates, `.chezmoiignore` rules, and machine-specific variables are resolved before streaming to the target.

---

## 🛠️ Troubleshooting

### Error: `chezmoi executable not found in PATH`
Install `chezmoi` on your local machine:
- Run `neossh --setup` to install it via your system's package manager.
- Or install manually: `brew install chezmoi` (macOS), `sudo apt install chezmoi` (Ubuntu/Debian), or `sh -c "$(curl -fsLS get.chezmoi.io)"`.

### Error: `remote tar extraction failed`
Ensure the user specified in your SSH configuration has write permissions to their remote home directory (`~`) and that `tar` is available in the remote `$PATH`.
