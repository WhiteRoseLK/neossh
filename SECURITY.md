# Security Policy

`neossh` takes the security of its users and their systems seriously. We appreciate your efforts to responsibly disclose any security vulnerabilities.

---

## Supported Versions

Only the latest major/minor releases of `neossh` receive active security patches:

| Version | Supported          |
| ------- | ------------------ |
| 2.x.x   | :white_check_mark: |
| 1.x.x   | :x:                |
| < 1.0   | :x:                |

We strongly advise all users to upgrade to the latest released version.

---

## Reporting a Vulnerability

If you discover a security vulnerability in `neossh`:

1. **Do not create a public GitHub issue.**
2. Report the vulnerability privately via **[GitHub Private Vulnerability Reporting](https://github.com/WhiteRoseLK/neossh/security/advisories/new)**.
3. Alternatively, send an email to [whiteroselk@users.noreply.github.com](mailto:whiteroselk@users.noreply.github.com) with:
   - A detailed description of the vulnerability.
   - Steps or proof of concept to reproduce the issue.
   - Potential impact of exploitation.
   - Any suggested remediations or mitigations.

### Response Timeline

- **Initial Acknowledgement**: Within 48 hours.
- **Triage & Assessment**: Within 7 business days.
- **Patch & Release**: Priority fixes are released as a patch update as quickly as possible.

We will credit you in the security advisory release notes (unless you prefer to remain anonymous).

---

## Security Architecture & Design Principles

`neossh` is built with a defense-in-depth security model:

1. **Zero Plaintext Passwords on Disk**: Server passwords configured for `sshpass` are managed via the operating system's native secure keyring (macOS Keychain, Linux Secret Service / DBus, Windows Credential Manager). When unavailable, an authenticated **AES-256-GCM** local vault is used (`~/.neossh/vault.json` with strict `0600` permissions). Passwords are never written to `~/.ssh/config` or `metadata.json`.
2. **Process Table Protection**: Password delivery to `sshpass` uses the `SSHPASS` environment variable (`sshpass -e`) rather than command-line arguments, preventing password leakage in process listings (`ps aux`).
3. **OpenSSH Native Execution**: Connections are mediated exclusively through the system's native `ssh` binary. All host key verification, cryptographic negotiation, and agent integrations remain under OpenSSH control.
4. **Non-Destructive Writes & Backups**: Config updates are performed atomically via temporary files and rename operations. Automatic snapshots (`config.original.backup`) and timestamped rolling backups are created before edits.
5. **Sanitized Sharing**: The `neossh export --sanitize` command strips `IdentityFile` paths and sensitive comments, ensuring that configuration bundles can be shared without exposing private key locations.
