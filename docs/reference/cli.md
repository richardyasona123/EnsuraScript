# CLI Reference

## Commands

### ensura compile

Validates an EnsuraScript file and shows compilation statistics.

```bash
ensura compile <file.ens> [options]
```

**Options:**
- `-json` - Output in JSON format
- `-graph` - Output dependency graph in DOT format

**Examples:**
```bash
ensura compile config.ens
ensura compile config.ens -json
ensura compile config.ens -graph > graph.dot
```

### ensura explain

Shows implied guarantees and chosen handlers.

```bash
ensura explain <file.ens> [options]
```

**Options:**
- `-json` - Output in JSON format

**Examples:**
```bash
ensura explain config.ens
ensura explain config.ens -json
```

### ensura plan

Prints the deterministic sequential execution plan.

```bash
ensura plan <file.ens> [options]
```

**Options:**
- `-json` - Output in JSON format

**Examples:**
```bash
ensura plan config.ens
ensura plan config.ens -json
```

### ensura run

Runs the continuous enforcement loop.

```bash
ensura run <file.ens> [options]
```

**Options:**
- `-interval <duration>` - Interval between enforcement loops (default: 30s)
- `-retries <count>` - Maximum retries per step (default: 3)

**Examples:**
```bash
ensura run config.ens
ensura run config.ens -interval 60s
ensura run config.ens -interval 10s -retries 5
```

### ensura check

Runs a single check pass without enforcement (dry run).

```bash
ensura check <file.ens> [options]
```

**Options:**
- `-json` - Output in JSON format

**Examples:**
```bash
ensura check config.ens
ensura check config.ens -json
```

### ensura version

Prints version information.

```bash
ensura version
```

### ensura help

Shows help message.

```bash
ensura help
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error (compilation, runtime, or violations detected) |

## Environment Variables

EnsuraScript supports referencing environment variables in handler arguments:

```ens
ensure encrypted with AES:256 key "env:SECRET_KEY"
```

The `env:` prefix tells EnsuraScript to read the value from the `SECRET_KEY` environment variable.

## File References

You can also reference files for secrets:

```ens
ensure encrypted with AES:256 key "file:/path/to/keyfile"
```
