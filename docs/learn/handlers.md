# Using Handlers

Handlers are the built-in modules that check and enforce guarantees. This tutorial covers all available handlers.

## What is a Handler?

A handler has two responsibilities:

1. **Check** - Determine if a guarantee is currently satisfied
2. **Enforce** - Make the guarantee true if it's violated

## Available Handlers

### fs.native - Filesystem Operations

**Conditions:** `exists`, `readable`, `writable`, `content`, `checksum`

#### exists

```ens
ensure exists on file "config.yaml"
```

- **Check**: Calls `os.Stat()` to see if path exists
- **Enforce**: Creates file with mode 0644 or directory with mode 0755

#### readable

```ens
ensure readable on file "data.txt"
```

- **Check**: Attempts to open file for reading
- **Enforce**: Cannot enforce (read-only check)

#### writable

```ens
ensure writable on file "output.log"
```

- **Check**: Attempts to open file for writing
- **Enforce**: Cannot enforce (read-only check)

#### content

```ens
ensure content with fs.native content "app_name: MyApp\nversion: 1.0" on file "config.yaml"
```

- **Check**: Reads file and compares exact string match
- **Enforce**: Writes content to file

#### checksum

```ens
ensure checksum with fs.native checksum "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" on file "release.tar.gz"
```

- **Check**: Computes SHA-256 and compares
- **Enforce**: Cannot enforce (read-only check)

---

### posix - POSIX Permissions

**Conditions:** `permissions`

```ens
ensure permissions with posix mode "0600" on file "secrets.db"
ensure permissions with posix mode "0755" on file "script.sh"
ensure permissions with posix mode "0644" on file "config.yaml"
```

**Arguments:**
- `mode` - Octal permission string (e.g., "0644", "0755", "0600")

**Operations:**
- **Check**: Compares `os.Stat().Mode().Perm()` with desired mode
- **Enforce**: Calls `os.Chmod(path, mode)`

---

### AES:256 - Encryption

**Conditions:** `encrypted`

```ens
ensure encrypted with AES:256 key "env:SECRET_KEY" on file "secrets.db"
```

**Arguments:**
- `key` - Key reference (required)

**Key Reference Formats:**
- `env:VARNAME` - Read from environment variable
- `file:/path/to/key` - Read from file
- `"literal"` - Use string directly

All keys are hashed to 32 bytes via SHA-256.

**Implementation Details:**
- Uses AES-256-GCM (Galois/Counter Mode)
- Magic header: `ENSURA_AES256_V1` (16 bytes)
- File format: `[MAGIC_HEADER][nonce + ciphertext]`

**Operations:**
- **Check**: Looks for magic header in file
- **Enforce**: Encrypts plaintext, prepends magic header

---

### http.get - HTTP Monitoring

**Conditions:** `reachable`, `status_code`, `tls`

#### reachable

```ens
ensure reachable on http "https://api.example.com/health"
```

- **Check**: HTTP GET request, succeeds if 2xx or 3xx status
- **Enforce**: Read-only (cannot fix unreachable endpoints)

#### status_code

```ens
ensure status_code with http.get expected_status "200" on http "https://example.com"
```

**Arguments:**
- `expected_status` - Expected HTTP status code (default: "200")

- **Check**: HTTP GET request, compares status code
- **Enforce**: Read-only

#### tls

```ens
ensure tls on http "https://secure.example.com"
```

- **Check**: Verifies TLS version ≥ 1.2
- **Enforce**: Read-only

**Configuration:**
- 30-second timeout
- Minimum TLS 1.2

---

### cron.native - Cron Scheduling

**Conditions:** `scheduled`

```ens
ensure scheduled with cron.native 
  schedule "0 2 * * *" 
  command "/usr/local/bin/backup.sh" 
  on cron "backup_job"
```

**Arguments:**
- `schedule` - Cron schedule string (e.g., "0 2 * * *")
- `command` - Command to execute

**Operations:**
- **Check**: Scans `crontab -l` for job marker
- **Enforce**: Adds/updates crontab entry with marker `# EnsuraScript: <jobname>`

**Platform Support:** Linux and macOS only

---

## Default Handler Selection

If you omit `with <handler>`, EnsuraScript chooses automatically:

| Condition | Default Handler |
|-----------|----------------|
| `exists` | `fs.native` |
| `readable` | `fs.native` |
| `writable` | `fs.native` |
| `content` | `fs.native` |
| `checksum` | `fs.native` |
| `permissions` | `posix` |
| `encrypted` | `AES:256` |
| `reachable` | `http.get` |
| `status_code` | `http.get` |
| `tls` | `http.get` |
| `scheduled` | `cron.native` |

## Handler Argument Syntax

Arguments are space-separated key-value pairs:

```ens
ensure <condition> with <handler> key1 "value1" key2 "value2"
```

Examples:

```ens
# Single argument
ensure permissions with posix mode "0755"

# Multiple arguments
ensure scheduled with cron.native schedule "0 2 * * *" command "/backup.sh"

# Environment variable in value
ensure encrypted with AES:256 key "env:SECRET_KEY"
```

## Next Steps

Continue to [Creating Policies](/learn/policies) to learn how to create reusable guarantee templates.
