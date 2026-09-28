#  Port Forwarding & Tunnels

## Interactive Tunnel Assistant

Press <kbd>f</kbd> to open the port forwarding assistant. It supports three tunnel types:

### Local Forwarding (`-L`)

Forward a local port to a service on the remote network:

```
Local port 3306 → remote-db:3306 via ssh-server
```

> [!NOTE]
> **Example: Access a remote database locally**
    Forward local port `3306` to `localhost:3306` on the remote server, then connect with `mysql -h 127.0.0.1 -P 3306`.

### Remote Forwarding (`-R`)

Expose a local service on the remote host:

```
Remote port 8080 → localhost:3000
```

### Dynamic SOCKS5 Proxy (`-D`)

Create a SOCKS5 proxy to tunnel all traffic through the SSH connection:

```
Local SOCKS5 proxy on port 1080
```

## Tunnel Profiles

Save frequently used tunnel configurations per host. Saved profiles are stored in `~/.neossh/settings.json` and can be quickly recalled from the tunnel assistant.

## Features

- **Real-time command preview** — See the complete SSH command before executing
- **Background daemon launch** — Run tunnels in the background
- **Clipboard copy** — Copy the generated command with one keypress

## CLI Usage

```bash
# Generate tunnel command templates and view saved profiles
neossh --tunnel web-prod

# Alias
neossh --forward web-prod
```

## Examples

```bash
# Local forwarding: access remote MySQL on local port 3306
ssh -L 3306:localhost:3306 web-prod

# Remote forwarding: expose local dev server to remote
ssh -R 8080:localhost:3000 web-prod

# Dynamic SOCKS5 proxy
ssh -D 1080 web-prod

# Background tunnel
ssh -fN -L 5432:localhost:5432 db-prod
```
