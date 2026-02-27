# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in inframap-d2, please report it responsibly.

**Do not open a public issue.** Instead, use GitHub's [private security advisory](https://github.com/ThomasCrouzet/inframap-d2/security/advisories/new) feature to report the vulnerability.

Include:
- A description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

You should receive a response within 72 hours. We will work with you to understand and address the issue before any public disclosure.

## Scope

This policy covers:
- The inframap-d2 CLI tool
- Configuration file handling (inframap.yml)
- API interactions (Portainer, Proxmox VE)
- SSH command execution (systemd collector)
- File path handling and traversal

## Best Practices

- **Never commit secrets** in `inframap.yml`. Use environment variables instead:
  - `INFRAMAP_PORTAINER_API_KEY`
  - `INFRAMAP_PROXMOX_TOKEN_ID`
  - `INFRAMAP_PROXMOX_TOKEN`
- Config files created by `inframap-d2 init` are written with `0600` permissions
- Use `--insecure` for Proxmox only in trusted networks (it disables TLS verification)
