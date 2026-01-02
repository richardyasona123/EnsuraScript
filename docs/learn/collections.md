# Collections & Invariants

Enforce guarantees across multiple resources using `for each` loops and `invariant` blocks.

## For Each Loops

Apply guarantees to all items in a collection.

### Syntax

```ens
for each <type> in <container> {
  <statements>
}
```

### Directory Iteration

```ens
for each file in directory "/secrets" {
  ensure encrypted with AES:256 key "env:SECRET_KEY"
  ensure permissions with posix mode "0600"
}
```

This ensures **every file** in `/secrets` is encrypted and has 0600 permissions.

## Invariants

Invariant blocks declare guarantees with **higher priority** than regular guarantees.

### Syntax

```ens
invariant {
  <statements>
}
```

### Why Invariants?

Invariants are executed **first** in the topological order. Use them for critical guarantees that must be established before anything else.

```ens
invariant {
  ensure exists on file "/etc/app/license.key"
  ensure readable on file "/etc/app/license.key"
}

# Regular guarantees follow
ensure running on service "app"
```

The license key is guaranteed to exist before the service starts.

## Combining For Each and Invariant

The most powerful pattern - enforce critical guarantees across collections:

```ens
invariant {
  for each file in directory "/secrets" {
    ensure encrypted with AES:256 key "env:SECRET_KEY"
    ensure permissions with posix mode "0600"
  }
}
```

This ensures:
1. All files in `/secrets` are encrypted (highest priority)
2. Executed before any other guarantees

## Practical Examples

### Secure Secrets Directory

```ens
invariant {
  # Ensure the directory itself is secure
  ensure exists on directory "/var/app/secrets"
  ensure permissions with posix mode "0700" on directory "/var/app/secrets"

  # Ensure all files within are encrypted and restricted
  for each file in directory "/var/app/secrets" {
    ensure encrypted with AES:256 key "env:SECRET_KEY"
    ensure permissions with posix mode "0600"
  }
}
```

### Web Assets Validation

```ens
for each file in directory "/var/www/html" {
  ensure exists
  ensure readable
  ensure permissions with posix mode "0644"
}
```

### Log Files

```ens
for each file in directory "/var/log/app" {
  ensure exists
  ensure writable
  ensure permissions with posix mode "0640"
}
```

## For Each with Policies

Combine collection iteration with policies:

```ens
policy secure_file(key_ref) {
  ensure encrypted with AES:256 key key_ref
  ensure permissions with posix mode "0600"
}

for each file in directory "/secrets" {
  apply secure_file("env:SECRET_KEY")
}
```

## Execution Order

Invariants have priority +1000 in topological sorting:

```ens
# Priority 1000
invariant {
  ensure exists on file "critical.db"
}

# Priority 0
ensure backed_up on file "critical.db"
```

The `exists` guarantee runs first, even though `backed_up` might logically depend on it.

## Invariant Best Practices

1. **Use sparingly** - Only for truly critical guarantees
2. **Foundation guarantees** - License files, critical configs, security policies
3. **System prerequisites** - Things that must exist before the system can function

## Implementation Note

The `for each` loop expands during the binding phase:

```ens
for each file in directory "/secrets" {
  ensure encrypted
}
```

If `/secrets` contains `a.txt` and `b.txt`, this expands to:

```ens
ensure encrypted on file "/secrets/a.txt"
ensure encrypted on file "/secrets/b.txt"
```

Each guarantee is then processed through implication expansion and added to the dependency graph.

## Next Steps

Continue to [Violation Handling](/learn/violations) to learn how to configure retry logic and notifications when guarantees fail.
