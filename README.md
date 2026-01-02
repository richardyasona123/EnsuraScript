# EnsuraScript

> Programming by guarantees, not instructions.

[![CI](https://github.com/GustyCube/EnsuraScript/actions/workflows/ci.yml/badge.svg)](https://github.com/GustyCube/EnsuraScript/actions/workflows/ci.yml)
[![Documentation](https://github.com/GustyCube/EnsuraScript/actions/workflows/deploy-docs.yml/badge.svg)](https://ensurascript.gustycube.com/)

EnsuraScript is an open-source, intent-first, "truth maintenance" language. You declare desired properties of your systems; the runtime satisfies and keeps them true.

## Features

- **Intent Over Instruction** - Declare outcomes with `ensure`, not procedural steps
- **Deterministic** - All inference is rule-based and inspectable, no AI guessing
- **Continuous Enforcement** - Guarantees are re-checked; drift triggers automatic repair
- **Composable** - Bundle reusable guarantees into policies
- **Skript-like Syntax** - Readable, declarative configuration

## Quick Start

### Installation

```bash
# Clone and build
git clone https://github.com/GustyCube/EnsuraScript.git
cd EnsuraScript
go build -o ensura ./cmd/ensura

# Optional: install to PATH
sudo mv ensura /usr/local/bin/
```

### Your First Script

Create `config.ens`:

```ens
# Ensure a secrets file is encrypted and secured
on file "secrets.db" {
  ensure exists
  ensure encrypted with AES:256 key "env:SECRET_KEY"
  ensure permissions with posix mode "0600"
}

on violation {
  retry 2
  notify "ops"
}
```

### Run It

```bash
# Check current state (dry run)
ensura check config.ens

# See what will happen
ensura explain config.ens

# Run continuous enforcement
ensura run config.ens
```

## Example Use Cases

### File Security

```ens
on file "credentials.json" {
  ensure exists
  ensure encrypted with AES:256 key "env:CRED_KEY"
  ensure permissions with posix mode "0600"
}
```

### HTTP Health Monitoring

```ens
ensure reachable on http "https://api.example.com/health"
ensure status_code on http "https://api.example.com/health" with http expected_status "200"
ensure tls on http "https://api.example.com/health"
```

### Reusable Policies

```ens
policy secure_file(key_ref) {
  ensure encrypted with AES:256 key key_ref
  ensure permissions with posix mode "0600"
}

on file "secrets.db" {
  apply secure_file("env:SECRET_KEY")
}
```

### Collection Enforcement

```ens
invariant {
  for each file in directory "/secrets" {
    ensure encrypted with AES:256 key "env:SECRET_KEY"
  }
}
```

## CLI Commands

| Command | Description |
|---------|-------------|
| `ensura compile <file>` | Validate and show resolved graph |
| `ensura explain <file>` | Show implied guarantees and handlers |
| `ensura plan <file>` | Show sequential execution plan |
| `ensura run <file>` | Run continuous enforcement loop |
| `ensura check <file>` | Check only (no repairs) |

## Documentation

Full documentation is available at [ensurascript.gustycube.com](https://ensurascript.gustycube.com/).

## Architecture

```
Source (.ens)
    ↓
  Lexer → Tokens
    ↓
  Parser → AST
    ↓
  Binder → Resolved subjects
    ↓
  Imply → Expanded guarantees
    ↓
  Graph → Dependency graph
    ↓
  Planner → Ordered execution plan
    ↓
  Runtime → Continuous enforcement
```

## Built-in Handlers

| Handler | Conditions | Description |
|---------|------------|-------------|
| `fs.native` | exists, readable, writable, checksum | Filesystem operations |
| `posix` | permissions | POSIX file permissions |
| `AES:256` | encrypted | AES-256 file encryption |
| `http.get` | reachable, status_code, tls | HTTP endpoint checks |

## Editor Support

### VS Code

Full-featured extension with syntax highlighting, snippets, and LSP support:

```bash
cd editors/vscode
npm install && npm run compile
npm run package
code --install-extension ensurascript-0.1.0.vsix
```

### Language Server

The `ensura-lsp` binary provides diagnostics, hover info, and completions for any LSP-capable editor:

```bash
go build -o bin/ensura-lsp ./cmd/ensura-lsp
```

See [editors/README.md](editors/README.md) for Neovim and Emacs configuration.

## Development

```bash
# Run tests
go test ./...

# Build
go build -o ensura ./cmd/ensura

# Build LSP server
go build -o bin/ensura-lsp ./cmd/ensura-lsp

# Run docs locally
cd docs && npm run docs:dev
```

## License

MIT License - see [LICENSE](LICENSE) for details.

## Contributing

Contributions are welcome! Please read our [contributing guidelines](CONTRIBUTING.md) before submitting PRs.
