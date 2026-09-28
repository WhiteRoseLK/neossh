# Pre-Connect Hooks

Run custom local commands or scripts automatically before connecting to an SSH host.

## Use Cases

-  Triggering corporate VPN connection scripts before dialing private IP ranges
-  Sending Wake-on-LAN (WOL) magic packets to spin up remote bare-metal hosts
-  Refreshing short-lived cloud credentials or MFA tokens (AWS SSM, Cloudflare Access, Okta)

## Configuration

### In SSH Config

Pre-connect hooks are configured directly in SSH config comments using `# pre-connect:` or `# hook:`:

```ssh-config
Host vpn-internal # pre-connect: /usr/local/bin/vpn-connect.sh %h
    HostName 10.10.0.50
    User devops

Host workstation-lab
    # hook: wakeonlan 00:11:22:33:44:55
    HostName 192.168.1.105
    User admin
```

### Via CLI

```bash
neossh --pre-connect "vpn-up.sh %h" -c my-server
```

### Via TUI

Edit a server (<kbd>e</kbd>) and configure the pre-connect command in the form.

## Token Expansion

Hook commands support token substitution:

| Token | Value |
|-------|-------|
| `%h` | Remote Hostname / IP address |
| `%p` | Remote SSH Port |
| `%r` | Remote User |
| `%n` | Server Alias |
| `%%` | Literal `%` |

## Environment Variables

neossh also sets context environment variables before running hooks:

| Variable | Value |
|----------|-------|
| `NEOSSH_ALIAS` | Server alias |
| `NEOSSH_HOST` | Remote hostname |
| `NEOSSH_PORT` | Remote SSH port |
| `NEOSSH_USER` | Remote username |

## Error Handling

If the pre-connect hook exits with a non-zero exit code, the SSH connection is **safely aborted** and the error output is reported in the TUI.

## Examples

```ssh-config
# VPN connection before accessing private network
Host private-server # pre-connect: /opt/vpn/connect.sh %h
    HostName 10.0.0.42
    User admin

# Wake-on-LAN for bare-metal host
Host homelab
    # hook: wakeonlan AA:BB:CC:DD:EE:FF && sleep 10
    HostName 192.168.1.100
    User root

# AWS SSO token refresh
Host aws-bastion # pre-connect: aws sso login --profile %n
    HostName bastion.aws.internal
    User ec2-user
```
