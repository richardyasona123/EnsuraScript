# Dependencies

Control the order of guarantee execution using `requires`, `after`, and `before` clauses.

## Why Dependencies Matter

Sometimes guarantees must happen in a specific order:

```ens
# Wrong - might try to encrypt before file exists!
ensure encrypted on file "secrets.db"
ensure exists on file "secrets.db"
```

EnsuraScript's **implication system** handles many dependencies automatically (`encrypted` implies `exists`), but sometimes you need explicit control.

## Dependency Clauses

### requires - Logical Prerequisite

```ens
ensure <condition-A> requires <condition-B>
```

Condition A requires condition B to be satisfied first.

```ens
ensure backed_up requires encrypted
```

Before backing up, ensure encryption is complete.

### after - Temporal Ordering

```ens
ensure <condition> after <resource>
```

This guarantee runs after all guarantees on the specified resource.

```ens
ensure startup on service "app" after file "config.yaml" exists
```

Don't start the service until the config file is ready.

### before - Reverse Temporal Ordering

```ens
ensure <condition> before <resource>
```

This guarantee runs before all guarantees on the specified resource.

```ens
ensure exists on file "config.yaml" before service "app" startup
```

Ensure config exists before the service starts.

## Examples

### Requires - Logical Dependencies

```ens
on file "database.db" {
  ensure exists
  ensure encrypted with AES:256 key "env:DB_KEY"
  ensure backed_up requires encrypted
}
```

The backup won't start until encryption is verified.

### After - Service Startup

```ens
resource file "/etc/app/config.yaml" as config
resource service "myapp" as app

ensure exists on config
ensure readable on config

ensure running on app after config
```

(Note: `service` resource type is parsed but not yet implemented)

### Before - Preparation Steps

```ens
ensure exists on file "/var/log/app.log" before file "/usr/bin/app" startup
```

Create the log file before starting the application.

## Combining Dependencies

You can use multiple dependency clauses:

```ens
ensure backed_up requires encrypted after database "mydb" ready
```

This reads as: "Ensure backed_up, but only after:
1. `encrypted` is satisfied (requires)
2. All guarantees on database mydb are satisfied (after)

## Implication vs. Explicit Dependencies

### Automatic (Implication)

```ens
ensure encrypted on file "secrets.db"
# Automatically ensures: exists, readable, writable FIRST
```

### Explicit (Dependencies)

```ens
ensure backed_up requires encrypted
# YOU specify that backed_up depends on encrypted
```

Use implication when possible (less code), use explicit dependencies for business logic ordering.

## Dependency Graph

EnsuraScript builds a **dependency graph** from:
1. Implication rules (automatic)
2. `requires` clauses (explicit)
3. `after`/`before` clauses (explicit)

Then it performs **topological sorting** to determine execution order.

View the execution order:

```bash
ensura plan config.ens
```

View the dependency graph as DOT format:

```bash
ensura compile config.ens --graph
```

## Cycle Detection

If dependencies create a cycle, compilation fails:

```ens
ensure A requires B
ensure B requires C
ensure C requires A  # CYCLE!
```

Error:

```
Compilation error: Dependency cycle detected: A → B → C → A
```

## Next Steps

Continue to [Collections & Invariants](/learn/collections) to learn how to enforce guarantees across multiple resources.
