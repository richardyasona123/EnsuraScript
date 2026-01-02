# Conditions Reference

Complete reference of all conditions and their implications.

## Condition Table

| Condition | Applies To | Implies | Default Handler | Description |
|-----------|------------|---------|-----------------|-------------|
| `exists` | file, directory | - | `fs.native` | Resource exists |
| `readable` | file | `exists` | `fs.native` | File can be read |
| `writable` | file | `exists` | `fs.native` | File can be written |
| `encrypted` | file | `exists`, `readable`, `writable` | `AES:256` | File is encrypted with AES-256-GCM |
| `permissions` | file, directory | `exists` | `posix` | POSIX permissions match specified mode |
| `content` | file | `exists` | `fs.native` | File content matches exact string |
| `checksum` | file | `exists`, `readable` | `fs.native` | File SHA-256 checksum matches |
| `reachable` | http | - | `http.get` | HTTP endpoint returns 2xx or 3xx |
| `status_code` | http | `reachable` | `http.get` | HTTP endpoint returns specific status |
| `tls` | http | `reachable` | `http.get` | HTTPS endpoint uses TLS ≥ 1.2 |
| `scheduled` | cron | - | `cron.native` | Cron job exists with specified schedule |
| `backed_up` | file | `exists` | - | File has backup (parsed, no handler) |
| `running` | service, process | - | - | Service/process is running (parsed, no handler) |
| `stopped` | service, process | - | - | Service/process is stopped (parsed, no handler) |
| `listening` | service | `running` | - | Service is listening (parsed, no handler) |
| `healthy` | service | `running` | - | Service is healthy (parsed, no handler) |

## Filesystem Conditions

### exists

File or directory exists on the filesystem.

**Syntax:**
```ens
ensure exists on file "path/to/file.txt"
ensure exists on directory "path/to/dir"
```

**Check:** `os.Stat(path)` succeeds  
**Enforce:** Create file with mode 0644 or directory with mode 0755  
**Implies:** -

---

### readable

File can be opened for reading.

**Syntax:**
```ens
ensure readable on file "data.txt"
```

**Check:** `os.Open(path)` succeeds  
**Enforce:** Cannot enforce (read-only check)  
**Implies:** `exists`

---

### writable

File can be opened for writing.

**Syntax:**
```ens
ensure writable on file "output.log"
```

**Check:** `os.OpenFile(path, O_WRONLY)` succeeds  
**Enforce:** Cannot enforce (read-only check)  
**Implies:** `exists`

---

### encrypted

File is encrypted with AES-256-GCM.

**Syntax:**
```ens
ensure encrypted with AES:256 key "env:SECRET_KEY" on file "secrets.db"
```

**Arguments:** `key` (required) - Key reference  
**Check:** File contains magic header `ENSURA_AES256_V1`  
**Enforce:** Encrypt file contents with AES-256-GCM  
**Implies:** `exists`, `readable`, `writable`

**Key Formats:**
- `env:VARNAME` - From environment variable
- `file:/path/to/key` - From file
- `"literal"` - Direct string (hashed to 32 bytes)

---

### permissions

File or directory has specific POSIX permissions.

**Syntax:**
```ens
ensure permissions with posix mode "0600" on file "secrets.db"
ensure permissions with posix mode "0755" on directory "/var/app"
```

**Arguments:** `mode` (required) - Octal permission string  
**Check:** `os.Stat().Mode().Perm()` matches mode  
**Enforce:** `os.Chmod(path, mode)`  
**Implies:** `exists`

**Common Modes:**
- `0600` - Owner read/write only
- `0644` - Owner read/write, others read
- `0700` - Owner full access only
- `0755` - Owner full, others read/execute

---

### content

File contains exact string content.

**Syntax:**
```ens
ensure content with fs.native content "key: value\napp: myapp" on file "config.yaml"
```

**Arguments:** `content` (required) - Exact file content  
**Check:** Read file, compare with exact string match  
**Enforce:** Write content to file  
**Implies:** `exists`

---

### checksum

File SHA-256 checksum matches.

**Syntax:**
```ens
ensure checksum with fs.native checksum "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" on file "release.tar.gz"
```

**Arguments:** `checksum` (required) - SHA-256 hex string  
**Check:** Compute SHA-256, compare with expected  
**Enforce:** Cannot enforce (read-only check)  
**Implies:** `exists`, `readable`

---

## HTTP Conditions

### reachable

HTTP/HTTPS endpoint is reachable and returns 2xx or 3xx.

**Syntax:**
```ens
ensure reachable on http "https://api.example.com"
```

**Check:** HTTP GET request, status 200-399  
**Enforce:** Cannot enforce (read-only check)  
**Implies:** -

**Timeout:** 30 seconds

---

### status_code

HTTP endpoint returns specific status code.

**Syntax:**
```ens
ensure status_code with http.get expected_status "200" on http "https://example.com"
ensure status_code with http.get expected_status "204" on http "https://webhooks.example.com/ping"
```

**Arguments:** `expected_status` - HTTP status code (default: "200")  
**Check:** HTTP GET request, compare status code  
**Enforce:** Cannot enforce (read-only check)  
**Implies:** `reachable`

---

### tls

HTTPS endpoint uses valid TLS ≥ 1.2.

**Syntax:**
```ens
ensure tls on http "https://secure.example.com"
```

**Check:** Verify TLS version ≥ 1.2  
**Enforce:** Cannot enforce (read-only check)  
**Implies:** `reachable`

---

## Cron Conditions

### scheduled

Cron job exists with specified schedule and command.

**Syntax:**
```ens
ensure scheduled with cron.native 
  schedule "0 2 * * *" 
  command "/usr/local/bin/backup.sh" 
  on cron "backup_job"
```

**Arguments:**
- `schedule` (required) - Cron schedule string
- `command` (required) - Command to execute

**Check:** Scan `crontab -l` for job marker  
**Enforce:** Add/update crontab entry  
**Implies:** -

**Platform:** Linux and macOS only  
**Marker Format:** `# EnsuraScript: <jobname>`

---

## Service Conditions (Parsed, Not Implemented)

These conditions are recognized by the parser but don't have handlers yet.

### running

Service or process is running.

**Syntax:**
```ens
ensure running on service "nginx"
ensure running on process "myapp"
```

**Status:** No handler implemented

---

### stopped

Service or process is stopped.

**Syntax:**
```ens
ensure stopped on service "nginx"
```

**Status:** No handler implemented

---

### listening

Service is listening on a port.

**Syntax:**
```ens
ensure listening on service "nginx"
```

**Implies:** `running`  
**Status:** No handler implemented

---

### healthy

Service is healthy (passes health check).

**Syntax:**
```ens
ensure healthy on service "myapp"
```

**Implies:** `running`  
**Status:** No handler implemented

---

## Other Conditions (Parsed, Not Implemented)

### backed_up

File has a backup.

**Syntax:**
```ens
ensure backed_up on file "database.db"
```

**Implies:** `exists`  
**Status:** No handler implemented

---

## Implication Chain Reference

Understanding transitive implications:

```
encrypted → exists, readable, writable
checksum → exists, readable
status_code → reachable
tls → reachable
listening → running
healthy → running
```

Example:

```ens
ensure status_code with http.get expected_status "200"
```

Expands to:

```
1. ensure reachable (implied)
2. ensure status_code (original)
```

---

## Conflict Detection

These conditions conflict and cannot both be true:

- `encrypted` ↔ `unencrypted`
- `running` ↔ `stopped`

If both are specified on the same resource, compilation fails.
