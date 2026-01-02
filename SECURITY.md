# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in EnsuraScript, please report it privately:

1. **Do not** open a public issue
2. Email the maintainers or use GitHub's private vulnerability reporting
3. Include details about the vulnerability and steps to reproduce

We will respond within 48 hours and work with you to address the issue.

## Security Considerations

### Secrets Handling

EnsuraScript handles sensitive data like encryption keys. Best practices:

- Use `env:VARNAME` to reference secrets from environment variables
- Use `file:/path/to/key` for key files with restricted permissions
- Never commit secrets to version control
- Secrets are never logged (redacted by default)

### File Permissions

When using the `posix` handler, ensure appropriate permissions:

- `0600` for sensitive files (owner read/write only)
- `0700` for sensitive directories
- Avoid `0777` in production

### Encryption

The `AES:256` handler uses AES-256-GCM, a secure authenticated encryption mode. Keys are hashed to 32 bytes via SHA-256.

## Supported Versions

| Version | Supported |
|---------|-----------|
| 1.x     | ✅        |
