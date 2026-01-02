# EnsuraScript for Visual Studio Code

Language support for [EnsuraScript](https://github.com/ensurascript/ensura) - programming by guarantees, not instructions.

## Features

- **Syntax Highlighting** - Full TextMate grammar for `.ens` files
- **Snippets** - Quick templates for common patterns
- **Hover Documentation** - Inline docs for keywords and handlers
- **Outline View** - Navigate policies and resource blocks

## Snippets

| Prefix | Description |
|--------|-------------|
| `on` | Resource context block |
| `ensure-exists` | Ensure resource exists |
| `ensure-encrypted` | Ensure AES-256 encryption |
| `ensure-permissions` | Ensure POSIX permissions |
| `policy` | Define a policy |
| `apply` | Apply a policy |
| `file` | File with common guarantees |
| `secure-file` | Encrypted file with restricted permissions |
| `http-health` | HTTP health check |
| `service` | Service monitoring |

## Example

```ens
policy secure_secrets(key_ref) {
  ensure encrypted with AES:256 key key_ref
  ensure permissions with posix mode "0600"
}

on file "secrets.db" {
  ensure exists
  apply secure_secrets("env:SECRET_KEY")
}

on http "https://api.example.com/health" {
  ensure reachable
  ensure tls
}
```

## Installation

### From VSIX (Local)

```bash
cd editors/vscode
npm install
npm run package
code --install-extension ensurascript-0.1.0.vsix
```

### From Marketplace

Coming soon!

## Requirements

- VS Code 1.75.0 or later

## License

MIT
