# Writing Guarantees

The `ensure` statement is the core of EnsuraScript. This tutorial covers all its forms and features.

## Basic Syntax

```ens
ensure <condition> on <resource>
```

Example:

```ens
ensure exists on file "config.yaml"
```

## Implicit Subjects in On Blocks

When inside an `on` block, the subject is inherited:

```ens
on file "config.yaml" {
  ensure exists          # subject is "file config.yaml"
  ensure readable        # subject is "file config.yaml"
  ensure writable        # subject is "file config.yaml"
}
```

## Using Handlers

Handlers perform the check and enforcement logic. Specify with `with`:

```ens
ensure encrypted with AES:256 key "env:SECRET_KEY"
ensure permissions with posix mode "0600"
ensure status_code with http.get expected_status "200"
```

### Handler Arguments

Handlers take key-value arguments:

```ens
ensure <condition> with <handler> <key1> "<value1>" <key2> "<value2>"
```

Examples:

```ens
# Single argument
ensure permissions with posix mode "0755"

# Multiple arguments
ensure scheduled with cron.native schedule "0 2 * * *" command "/backup.sh"
```

## Default Handlers

If you omit `with <handler>`, EnsuraScript selects a default:

```ens
ensure exists         # uses fs.native
ensure encrypted      # uses AES:256
ensure permissions    # uses posix
ensure reachable      # uses http.get
```

## Common Conditions

### Filesystem Conditions

```ens
ensure exists on file "config.yaml"
ensure readable on file "data.txt"
ensure writable on file "output.log"
ensure permissions with posix mode "0644" on file "public.txt"
ensure permissions with posix mode "0600" on file "secrets.txt"
ensure content with fs.native content "key: value" on file "config.yaml"
ensure checksum with fs.native checksum "abc123..." on file "release.tar.gz"
```

### Encryption

```ens
ensure encrypted with AES:256 key "env:SECRET_KEY" on file "secrets.db"
```

Key reference formats:
- `env:VAR_NAME` - from environment variable
- `file:/path/to/key` - from file
- `"literal"` - direct string (hashed to 32 bytes)

### HTTP Conditions

```ens
ensure reachable on http "https://api.example.com"
ensure status_code with http.get expected_status "200" on http "https://example.com"
ensure tls on http "https://secure.example.com"
```

### Cron Scheduling

```ens
ensure scheduled with cron.native 
  schedule "0 2 * * *" 
  command "/usr/local/bin/backup.sh" 
  on cron "backup_job"
```

## Full Ensure Syntax

The complete `ensure` statement supports several clauses:

```ens
ensure <condition> [on <resource>] [with <handler> <args>] [when <guard>] [requires <condition>] [after <resource>] [before <resource>]
```

We'll cover `when`, `requires`, `after`, and `before` in later tutorials.

## Chaining Guarantees

Multiple guarantees are enforced in order:

```ens
on file "app.db" {
  ensure exists                    # 1. Create if needed
  ensure permissions with posix mode "0600"  # 2. Set permissions
  ensure encrypted with AES:256 key "env:DB_KEY"  # 3. Encrypt
}
```

Thanks to the implication system, you could also write just:

```ens
on file "app.db" {
  ensure encrypted with AES:256 key "env:DB_KEY"
}
```

And EnsuraScript would automatically ensure `exists`, `readable`, and `writable` first.

## Next Steps

Continue to [Using Handlers](/learn/handlers) to learn about all available handlers and their arguments.
