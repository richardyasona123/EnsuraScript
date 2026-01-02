# Core Concepts

## Resources

A **resource** is something EnsuraScript can observe and/or modify. Resources are the subjects of your guarantees.

```ens
resource file "secrets.db"
resource http "https://api.example.com/health"
resource directory "/var/log/myapp"
```

### Resource Types

| Type | Description |
|------|-------------|
| `file` | A file on the filesystem |
| `directory` | A directory on the filesystem |
| `http` | An HTTP/HTTPS endpoint |
| `database` | A database connection |
| `service` | A system service |
| `process` | A running process |
| `cron` | A scheduled task |

### Aliases

Give resources memorable names:

```ens
resource file "secrets.db" as secrets
resource http "https://api.example.com/health" as api_health
```

## Guarantees

A **guarantee** is a constraint that must be true. Written with `ensure`.

```ens
ensure exists on file "secrets.db"
ensure reachable on http "https://api.example.com/health"
```

### The Ensure Statement

```ens
ensure <condition> on <resource> [with <handler>] [when <guard>]
```

- **condition**: What should be true (exists, encrypted, reachable, etc.)
- **resource**: What to check/modify
- **handler**: How to check/enforce (optional, uses default if unambiguous)
- **guard**: When this guarantee is active (optional)

## Handlers

A **handler** is a plugin that knows how to check and enforce a condition.

```ens
ensure encrypted on file "secrets.db" with AES:256 key "env:SECRET_KEY"
ensure permissions on file "secrets.db" with posix mode "0600"
```

### Built-in Handlers

| Handler | Conditions | Description |
|---------|------------|-------------|
| `fs.native` | exists, readable, writable, checksum | Filesystem operations |
| `posix` | permissions | POSIX file permissions |
| `AES:256` | encrypted | AES-256 file encryption |
| `http.get` | reachable, status_code, tls | HTTP endpoint checks |

## Context Blocks

Use `on` blocks to group guarantees for a resource:

```ens
on file "secrets.db" {
  ensure exists
  ensure encrypted with AES:256 key "env:SECRET_KEY"
  ensure permissions with posix mode "0600"
}
```

Statements inside the block automatically inherit the subject.

## Implications

Certain conditions automatically imply others. For example:

- `encrypted` implies `exists`, `readable`, `writable`
- `permissions` implies `exists`
- `status_code` implies `reachable`

EnsuraScript automatically expands these implications and orders them correctly.

## The Execution Model

1. **Compile**: Parse, bind, expand implications, build dependency graph
2. **Plan**: Topologically sort guarantees into an ordered execution plan
3. **Run**: Sequentially check each guarantee; repair violations
4. **Loop**: Sleep, then repeat

```
Source (.ens)
    ↓
  Parse → AST
    ↓
  Bind → Resolve implicit subjects
    ↓
  Expand → Add implied guarantees
    ↓
  Graph → Build dependencies
    ↓
  Plan → Topological sort
    ↓
  Execute → Check/Repair loop
```
