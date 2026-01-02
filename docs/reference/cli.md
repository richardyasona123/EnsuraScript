# CLI Reference

Complete reference for all `ensura` CLI commands.

## ensura compile

Validate syntax and semantics of an EnsuraScript file.

### Usage

```bash
ensura compile <file.ens> [options]
```

### Options

- `--json` - Output compilation result as JSON
- `--graph` - Output dependency graph in DOT format

### Output

Shows compilation statistics:

```
Compilation successful.
Resources: 5
Guarantees: 12 (4 implied)
Policies: 2
```

### Examples

```bash
# Basic compilation
ensura compile config.ens

# JSON output
ensura compile config.ens --json

# Generate dependency graph
ensura compile config.ens --graph > graph.dot
dot -Tpng graph.dot -o graph.png
```

### Exit Codes

- `0` - Compilation successful
- `1` - Compilation error (syntax or semantic)

---

## ensura explain

Show all guarantees including implied ones.

### Usage

```bash
ensura explain <file.ens> [options]
```

### Options

- `--json` - Output as JSON

### Output

Lists all guarantees with:
- Implied guarantees marked `[IMPLIED]`
- Invariant guarantees marked `[INVARIANT]`
- Handler and arguments for each

```
Guarantees (6 total, 3 implied):

1. [IMPLIED] [fs.native] ensure exists on file "secrets.db"
2. [IMPLIED] [fs.native] ensure readable on file "secrets.db"
3. [IMPLIED] [fs.native] ensure writable on file "secrets.db"
4. [AES:256] ensure encrypted on file "secrets.db" with AES:256 key "env:SECRET_KEY"
5. [posix] ensure permissions on file "secrets.db" with posix mode "0600"
6. [fs.native] ensure backed_up on file "secrets.db"
```

### Examples

```bash
# Show expanded guarantees
ensura explain config.ens

# JSON output
ensura explain config.ens --json
```

---

## ensura plan

Show deterministic execution order.

### Usage

```bash
ensura plan <file.ens> [options]
```

### Options

- `--json` - Output as JSON

### Output

Shows sequential execution plan:

```
Execution Plan (6 steps):

1. [fs.native] ensure exists on file "secrets.db"
2. [fs.native] ensure readable on file "secrets.db"
3. [fs.native] ensure writable on file "secrets.db"
4. [AES:256] ensure encrypted on file "secrets.db" with AES:256 key "env:SECRET_KEY"
5. [posix] ensure permissions on file "secrets.db" with posix mode "0600"
6. [fs.native] ensure backed_up on file "secrets.db"
```

### Examples

```bash
# Show execution plan
ensura plan config.ens

# JSON output
ensura plan config.ens --json
```

---

## ensura run

Continuous enforcement loop.

### Usage

```bash
ensura run <file.ens> [options]
```

### Options

- `--interval <duration>` - Time between checks (default: 30s)
- `--retries <count>` - Max retry attempts (default: 3)
- `--verbose` - Verbose output

### Duration Format

- `30s` - 30 seconds
- `1m` - 1 minute
- `5m` - 5 minutes
- `1h` - 1 hour

### Output

Shows check and enforce results:

```
[✓] ensure exists on file "secrets.db" - SATISFIED
[✓] ensure encrypted on file "secrets.db" - SATISFIED
[✓] ensure permissions on file "secrets.db" - SATISFIED

All guarantees satisfied. Monitoring for drift...

(30 seconds later...)

[✓] ensure permissions on file "secrets.db" - REPAIRED

All guarantees satisfied. Monitoring for drift...
```

**Status Codes:**
- `SATISFIED` - Check passed, already true
- `REPAIRED` - Check failed, enforce succeeded
- `VIOLATED` - Dry-run mode, check failed
- `FAILED` - Enforce failed after all retries

### Examples

```bash
# Run with defaults (30s interval, 3 retries)
ensura run config.ens

# Custom interval
ensura run config.ens --interval 1m

# Custom retry count
ensura run config.ens --retries 5

# Verbose output
ensura run config.ens --verbose

# Combined
ensura run config.ens --interval 5m --retries 10 --verbose
```

### Stopping

Press `Ctrl+C` to stop gracefully. The runtime will finish the current loop and exit.

---

## ensura check

Single dry-run check without enforcement.

### Usage

```bash
ensura check <file.ens> [options]
```

### Options

- `--json` - Output as JSON

### Behavior

Runs one iteration of the check loop:
- Does NOT enforce
- Reports violations
- Exits after one pass

### Output

```
[✓] ensure exists on file "secrets.db" - SATISFIED
[✗] ensure encrypted on file "secrets.db" - VIOLATED
[✓] ensure permissions on file "secrets.db" - SATISFIED

2/3 guarantees satisfied. 1 violation detected.
```

### Exit Codes

- `0` - All guarantees satisfied
- `1` - At least one violation detected

### Examples

```bash
# Check without enforcing
ensura check config.ens

# JSON output
ensura check config.ens --json

# Use in CI/CD
if ensura check production.ens; then
  echo "Production configuration valid"
else
  echo "Production configuration has violations!"
  exit 1
fi
```

---

## ensura --version

Show version information.

### Usage

```bash
ensura --version
```

### Output

```
EnsuraScript v1.0.0
```

---

## ensura --help

Show help information.

### Usage

```bash
ensura --help
ensura <command> --help
```

### Output

Shows usage and available commands.

---

## Environment Variables

### SECRET_KEY

Example usage with encryption:

```bash
export SECRET_KEY="my-encryption-key"
ensura run config.ens
```

Any environment variable can be referenced via `env:VARNAME` in key references.

---

## Exit Code Summary

| Command | Success | Failure |
|---------|---------|---------|
| `compile` | 0 | 1 (syntax/semantic error) |
| `explain` | 0 | 1 (compilation error) |
| `plan` | 0 | 1 (compilation error) |
| `run` | 0 (if stopped with Ctrl+C) | 1 (if critical error) |
| `check` | 0 (all satisfied) | 1 (violations detected) |

---

## Common Workflows

### Development

```bash
# Validate syntax
ensura compile config.ens

# See what will happen
ensura plan config.ens

# Run once to check
ensura check config.ens

# Run with enforcement
ensura run config.ens
```

### Production

```bash
# Validate before deployment
ensura compile production.ens || exit 1

# Run with long interval
ensura run production.ens --interval 5m --retries 10
```

### CI/CD

```bash
#!/bin/bash
set -e

# Validate configuration
ensura compile config.ens

# Check all guarantees (no enforcement)
ensura check config.ens

# If we get here, configuration is valid
echo "Configuration validated successfully"
```

### Debugging

```bash
# See expanded guarantees
ensura explain config.ens

# See execution order
ensura plan config.ens

# Generate dependency graph
ensura compile config.ens --graph | dot -Tpng > graph.png

# Run with verbose output
ensura run config.ens --verbose
```
