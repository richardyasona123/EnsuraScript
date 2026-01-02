# Handlers

Handlers are pluggable implementations that know how to check and enforce conditions.

## Specifying Handlers

```ens
ensure <condition> with <handler> <key> "<value>" ...
```

### Example

```ens
ensure encrypted on file "secrets.db" with AES:256 key "env:SECRET_KEY"
ensure permissions on file "secrets.db" with posix mode "0600"
```

## Built-in Handlers

### fs.native

Filesystem operations for files and directories.

**Conditions:** exists, readable, writable, checksum, content

```ens
ensure exists on file "config.yaml"
ensure checksum on file "config.yaml" with fs.native expected "abc123..."
```

### posix

POSIX file permission management.

**Conditions:** permissions

**Arguments:**
- `mode` - Octal permission mode (e.g., "0600", "0755")

```ens
ensure permissions on file "secrets.db" with posix mode "0600"
ensure permissions on directory "/var/log" with posix mode "0755"
```

### AES:256

AES-256-GCM file encryption.

**Conditions:** encrypted

**Arguments:**
- `key` - Encryption key reference (env:VAR, file:/path, or literal)

```ens
ensure encrypted on file "secrets.db" with AES:256 key "env:SECRET_KEY"
ensure encrypted on file "creds.json" with AES:256 key "file:/etc/keys/master.key"
```

### http.get

HTTP endpoint checking.

**Conditions:** reachable, status_code, tls

**Arguments:**
- `expected_status` - Expected HTTP status code (default: 200)

```ens
ensure reachable on http "https://api.example.com/health"
ensure status_code on http "https://api.example.com/health" with http.get expected_status "200"
ensure tls on http "https://api.example.com/health"
```

### cron.native

Cron scheduling for Linux and macOS.

**Conditions:** scheduled

**Arguments:**
- `schedule` - Cron schedule expression (e.g., "0 2 * * *")
- `command` - Command to execute

```ens
ensure scheduled on cron "backup_job" with cron.native schedule "0 2 * * *" command "/usr/local/bin/backup.sh"
```

## Handler Selection

If you don't specify a handler with `with`, EnsuraScript uses the default handler for that condition:

| Condition | Default Handler |
|-----------|-----------------|
| exists | fs.native |
| readable | fs.native |
| writable | fs.native |
| encrypted | AES:256 |
| permissions | posix |
| checksum | fs.native |
| content | fs.native |
| reachable | http.get |
| status_code | http.get |
| tls | http.get |
| scheduled | cron.native |

## Secret References

Handler arguments support special prefixes for secrets:

- `env:VAR` - Read from environment variable
- `file:/path` - Read from file

```ens
# From environment variable
ensure encrypted with AES:256 key "env:SECRET_KEY"

# From file
ensure encrypted with AES:256 key "file:/etc/secrets/key"
```

**Security Note:** Secrets are never logged. Use `--redact` (default on) to ensure sensitive values are masked.
