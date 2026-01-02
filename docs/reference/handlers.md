# Handlers Reference

Complete reference of all handlers and their arguments.

## Handler Overview

Handlers implement the check and enforce logic for conditions. Each handler supports specific conditions on specific resource types.

| Handler | Conditions | Resource Types | Arguments |
|---------|-----------|----------------|-----------|
| `fs.native` | exists, readable, writable, content, checksum | file, directory | content, checksum |
| `posix` | permissions | file, directory | mode (required) |
| `AES:256` | encrypted | file | key (required) |
| `http.get` | reachable, status_code, tls | http | expected_status |
| `cron.native` | scheduled | cron | schedule (required), command (required) |

---

## fs.native

Filesystem operations handler.

### Supported Conditions

#### exists

**Check:** `os.Stat(path)` succeeds  
**Enforce:** Create file with mode 0644 or directory with mode 0755

```ens
ensure exists on file "config.yaml"
ensure exists on directory "/var/app/data"
```

#### readable

**Check:** `os.Open(path)` succeeds  
**Enforce:** Cannot enforce (read-only)

```ens
ensure readable on file "data.txt"
```

#### writable

**Check:** `os.OpenFile(path, O_WRONLY)` succeeds  
**Enforce:** Cannot enforce (read-only)

```ens
ensure writable on file "output.log"
```

#### content

**Check:** Read file, exact string comparison  
**Enforce:** Write content to file

**Arguments:**
- `content` (required) - Exact file content

```ens
ensure content with fs.native content "app_name: MyApp\nversion: 1.0" on file "config.yaml"
```

#### checksum

**Check:** Compute SHA-256, compare with expected  
**Enforce:** Cannot enforce (read-only)

**Arguments:**
- `checksum` (required) - SHA-256 hex string (64 characters)

```ens
ensure checksum with fs.native checksum "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" on file "release.tar.gz"
```

---

## posix

POSIX permissions handler.

### Supported Conditions

#### permissions

**Check:** `os.Stat().Mode().Perm()` comparison  
**Enforce:** `os.Chmod(path, mode)`

**Arguments:**
- `mode` (required) - Octal permission string (e.g., "0644", "0755")

```ens
ensure permissions with posix mode "0600" on file "secrets.db"
ensure permissions with posix mode "0755" on directory "/var/app/bin"
```

**Permission Modes:**

| Mode | Binary | Meaning |
|------|--------|---------|
| 0600 | rw------- | Owner read/write |
| 0644 | rw-r--r-- | Owner read/write, others read |
| 0700 | rwx------ | Owner full access |
| 0755 | rwxr-xr-x | Owner full, others read/execute |
| 0777 | rwxrwxrwx | All full access (avoid) |

---

## AES:256

AES-256-GCM encryption handler.

### Supported Conditions

#### encrypted

**Check:** Look for magic header `ENSURA_AES256_V1` in file  
**Enforce:** Encrypt file with AES-256-GCM, prepend magic header

**Arguments:**
- `key` (required) - Key reference

**Key Reference Formats:**

1. **Environment Variable:**
   ```ens
   ensure encrypted with AES:256 key "env:SECRET_KEY"
   ```
   Reads from `os.Getenv("SECRET_KEY")`

2. **File:**
   ```ens
   ensure encrypted with AES:256 key "file:/etc/app/encryption.key"
   ```
   Reads key from file

3. **Literal:**
   ```ens
   ensure encrypted with AES:256 key "my-secret-key-123"
   ```
   Uses string directly

All keys are hashed to 32 bytes via SHA-256 before use.

**File Format:**

```
[ENSURA_AES256_V1 (16 bytes)][12-byte nonce + ciphertext]
```

**Algorithm:** AES-256-GCM (Galois/Counter Mode)

**Example:**

```ens
on file "secrets.db" {
  ensure encrypted with AES:256 key "env:DB_ENCRYPTION_KEY"
}
```

Before running:
```bash
export DB_ENCRYPTION_KEY="my-secure-key"
ensura run config.ens
```

---

## http.get

HTTP monitoring handler.

### Supported Conditions

#### reachable

**Check:** HTTP GET request returns 2xx or 3xx  
**Enforce:** Cannot enforce (read-only)

```ens
ensure reachable on http "https://api.example.com/health"
```

**Timeout:** 30 seconds

#### status_code

**Check:** HTTP GET request, compare status code  
**Enforce:** Cannot enforce (read-only)

**Arguments:**
- `expected_status` - HTTP status code string (default: "200")

```ens
ensure status_code with http.get expected_status "200" on http "https://example.com"
ensure status_code with http.get expected_status "204" on http "https://webhooks.example.com/ping"
ensure status_code with http.get expected_status "301" on http "https://redirect.example.com"
```

#### tls

**Check:** Verify TLS version ≥ 1.2  
**Enforce:** Cannot enforce (read-only)

```ens
ensure tls on http "https://secure.example.com"
```

**Minimum TLS Version:** TLS 1.2

---

## cron.native

Cron scheduling handler (Linux/macOS only).

### Supported Conditions

#### scheduled

**Check:** Scan `crontab -l` for job with marker `# EnsuraScript: <jobname>`  
**Enforce:** Add or update crontab entry with marker

**Arguments:**
- `schedule` (required) - Cron schedule string
- `command` (required) - Command to execute

**Schedule Format:** Standard cron format `minute hour day month weekday`

**Examples:**

```ens
# Daily at 2 AM
ensure scheduled with cron.native 
  schedule "0 2 * * *" 
  command "/usr/local/bin/backup.sh" 
  on cron "daily_backup"

# Every 15 minutes
ensure scheduled with cron.native 
  schedule "*/15 * * * *" 
  command "/usr/local/bin/health_check.sh" 
  on cron "health_monitor"

# Weekdays at 9 AM
ensure scheduled with cron.native 
  schedule "0 9 * * 1-5" 
  command "/usr/local/bin/report.sh" 
  on cron "daily_report"
```

**Crontab Entry Format:**

```
<schedule> <command> # EnsuraScript: <jobname>
```

**Platform Support:** Linux and macOS (not Windows)

---

## Default Handler Selection

If you omit `with <handler>`, EnsuraScript selects the default:

```ens
ensure exists on file "config.yaml"
# Uses fs.native by default

ensure encrypted on file "secrets.db"
# Uses AES:256 by default

ensure permissions with posix mode "0600"
# Must specify posix explicitly due to required argument
```

## Handler Error Handling

When a handler's enforce operation fails:

1. **Retry** - Up to `max_retries` times (default: 3)
2. **Re-check** - After each retry
3. **Violation Handler** - If all retries fail
4. **Mark Failed** - Status becomes FAILED

Example with retries:

```ens
ensure encrypted on file "secrets.db" with AES:256 key "env:SECRET_KEY"
on violation {
  retry 10
  notify "ops-team"
}
```

If encryption fails 10 times, ops-team is notified (when notification integration is implemented).

---

## Read-Only Handlers

Some conditions can only be checked, not enforced:

- `readable` - Can't make a file readable if it doesn't exist
- `writable` - Can't make a file writable if it doesn't exist
- `checksum` - Can't change file to match checksum
- `reachable` - Can't fix remote endpoints
- `status_code` - Can't fix remote endpoints
- `tls` - Can't fix remote TLS configuration

For these, enforcement always fails. Use them for monitoring and validation.
