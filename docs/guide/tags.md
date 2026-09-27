---
title: Tags & Config Comments
---

# :material-tag-multiple: Server Tags in SSH Config Comments

`neossh` allows assigning custom tags to servers (e.g., `prod`, `database`, `aws`, `k8s`) for rapid filtering and organization.

Unlike tools that store metadata exclusively in a local database, `neossh` persists tags directly inside `~/.ssh/config` as structured comments.

---

## Two-Way Synchronization

By saving tags as SSH configuration comments, your tags remain **completely portable and synchronizable** across multiple workstations using Git dotfiles, Nextcloud, or rsync without needing to export proprietary databases.

### Configuration Format

Tags can be formatted inline on the `Host` line or as dedicated comments inside the host block:

```ssh-config
# Inline on the Host line:
Host web-prod # tags: prod, web, us-east
    HostName 192.168.1.10
    User ubuntu

# Or inside the Host block:
Host db-primary
    # tags: prod, database
    HostName 192.168.1.20
    User postgres
```

- Comma-separated or space-separated lists of tags are supported.
- When edited via the TUI (++t++ or ++e++), tags are written directly back to the source file where the host is defined.

---

## Editing Tags in the TUI

1. Highlight any server in the list.
2. Press ++t++ to open the tag editor modal.
3. Add, remove, or modify tags separated by commas.
4. Press ++enter++ to save. The changes are immediately written to your `~/.ssh/config`.

---

## Filtering by Tags

In the search bar (++slash++ or ++0++):

```
tag:prod
```

Matches servers containing the `prod` tag.

```
-tag:staging
```

Excludes servers containing the `staging` tag.

Combine tags with other search criteria:
```
tag:prod user:root status:up web
```

---

## Other Supported Config Directives

`neossh` recognizes several other special comment directives within `~/.ssh/config`:

| Directive Comment | Purpose | Example |
|---|---|---|
| `# tags: <list>` | Server categorization tags | `# tags: prod, database` |
| `# pin` | Pin server to top of server list | `Host db # pin` |
| `# hidden` | Hide server from default view | `Host jumpbox # hidden` |
| `# pre-connect: <cmd>` | Command executed before SSH connection | `# pre-connect: vpn-up.sh %h` |
| `# hook: <cmd>` | Alias for pre-connect command | `# hook: wakeonlan %h` |
| `# certificate-command: <cmd>` | Certificate renewal command | `# certificate-command: step ssh login %u@%h` |
| `# cert-command: <cmd>` | Alias for certificate renewal command | `# cert-command: vault write ...` |
| `# cert-renew: <cmd>` | Alias for certificate renewal command | `# cert-renew: tsh login` |

All comment directives are preserved verbatim across read and write operations.
