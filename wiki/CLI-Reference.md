# Command Line Usage

`neossh` provides an extensive set of command-line flags and subcommands for scripting, automation, direct connection, and headless operations.

```bash
neossh [filter] [flags]
neossh [command] [flags]
```

---

## Global Options & Flags

| Flag | Shorthand | Description | Default |
|---|:---:|---|:---:|
| `[filter]` | | Optional positional argument to pre-filter server list | `""` |
| `--filter <pattern>` | `-f` | Pre-filter server list by alias, hostname, or tag | `""` |
| `--connect` | `-c` | Connect directly to matching server without launching full TUI picker | `false` |
| `--import-known-hosts` | | Import newly discovered hosts from `known_hosts` into SSH config | `false` |
| `--known-hosts <path>` | | Specify custom path to `known_hosts` file | `~/.ssh/known_hosts` |
| `--git-ssh <key>` | | Configure Git SSH key for current repo (or globally) or view current setting | `""` |
| `--theme <mode>` | `-t` | Set color theme: `dark`, `light`, or `system` | stored pref or `dark` |
| `--lang <code>` | `-l` | Set interface language: `en`, `fr`, `zh-CN` | `en` |
| `--show-hidden` | `-H` | Display hidden servers in UI list | `false` |
| `--sftp <alias>` | | Quick-launch interactive SFTP session for server alias | `""` |
| `--file-manager <tool>` | | Specify file manager tool or command template for SFTP | `sftp` or settings |
| `--scp <alias>` | | Generate SCP upload/download command templates and copy to clipboard | `""` |
| `--sshfs <alias>` | | Generate SSHFS remote mount and unmount command templates and copy to clipboard | `""` |
| `--tunnel <alias>` | | Generate SSH port forwarding command templates and view saved profiles | `""` |
| `--forward <alias>` | | Alias for `--tunnel` | `""` |
| `--pre-connect <cmd>` | | Run local hook command before SSH connect (supports `%h`, `%p`, `%r`, `%n`) | `""` |
| `--default-key <path>` | | Get or set default SSH identity key prefilled for new server entries | `""` |
| `--password <pwd>` | `-P` | Password for automated `sshpass` authentication | `""` |
| `--ping-watch` | | Enable periodic background ping watch mode | `false` |
| `--ping-interval <sec>` | | Interval in seconds for periodic background ping watch mode | `60` |
| `--sshconfig <path>` | | Specify custom path to SSH config file | `~/.ssh/config` |
| `--readonly` | `-r` | Run in read-only / viewer mode (prevents writing or modifying config) | `false` |
| `--ssh-config-readonly` | | Alias for `--readonly` | `false` |
| `--exit-on-disconnect` | `-x` | Exit `neossh` immediately after SSH session terminates | `false` |
| `--auto-exit` | | Alias for `--exit-on-disconnect` | `false` |
| `--help` | `-h` | Display help message and available options | |
| `--version` | `-v` | Display version information | |

---

## Subcommands

### `export` / `backup`

Archives your primary `~/.ssh/config`, all recursively resolved `Include` files, and metadata (`metadata.json`, `settings.json`) into a compressed `.tar.gz` bundle.

```bash
# Basic export
neossh export

# Specify custom destination
neossh export --output ~/backups/my-ssh-bundle.tar.gz

# Sanitized export for dotfiles or team sharing
neossh export --sanitize
neossh export -s

# 'backup' is a direct alias for 'export'
neossh backup
```

### `import` / `restore`

Restores an exported bundle to `~/.ssh/` and `~/.neossh/` with automatic `.bak.<timestamp>` safety backups.

```bash
# Inspect changes without modifying anything
neossh import --dry-run my-ssh-bundle.tar.gz

# Restore bundle
neossh import my-ssh-bundle.tar.gz

# 'restore' is a direct alias for 'import'
neossh restore my-ssh-bundle.tar.gz
```

### `verify`

Validates bundle structure, schema version, and cryptographic SHA-256 checksums without extracting files.

```bash
neossh verify my-ssh-bundle.tar.gz
```

### `completion`

Generates shell completion scripts for your preferred shell.

```bash
neossh completion bash
neossh completion zsh
neossh completion fish
neossh completion powershell
```

See [[Shell Autocompletion|Shell-Completion]] for installation instructions.

---

## Environment Variables

| Variable | Description |
|---|---|
| `NEOSSH_LANG` | Default UI language code (`en`, `fr`, `zh-CN`). |
| `NEOSSH_DEFAULT_KEY` | Default SSH private key path for new host entries. |
| `NEOSSH_PASSWORD` | Password for automated `sshpass` connection. |
| `SSHPASS` | Standard fallback password variable for `sshpass`. |
| `XDG_CONFIG_HOME` | Overrides base configuration directory (default: `~/.config`). |
| `XDG_STATE_HOME` | Overrides base state/metadata directory. |

---

## Practical Examples

```bash
# 1. Launch normal interactive TUI
neossh

# 2. Direct connect to server 'web-prod'
neossh -c web-prod

# 3. Direct connect and exit terminal when SSH session ends
neossh -c -x web-prod

# 4. Quick launch SFTP file manager for server 'web-prod'
neossh --sftp web-prod

# 5. Quick launch external Yazi file manager
neossh --sftp web-prod --file-manager yazi

# 6. Pre-filter view to servers containing 'k8s'
neossh k8s
# or
neossh -f k8s

# 7. Safe read-only mode for production environments
neossh -r

# 8. Ping watch mode checking servers every 30 seconds
neossh --ping-watch --ping-interval 30

# 9. Configure default SSH key
neossh --default-key ~/.ssh/id_ed25519

# 10. Bootstrap from known_hosts
neossh --import-known-hosts

# 11. Custom SSH config path
neossh --sshconfig ~/.ssh/config_staging -r
```
