# Understanding Resources

Resources are the entities that EnsuraScript manages. Every guarantee applies to a resource.

## Resource Types

EnsuraScript supports several resource types:

### File Resources

```ens
resource file "/path/to/file.txt"
```

Represents a filesystem file. Supports conditions like:
- `exists`, `readable`, `writable`
- `encrypted`, `permissions`
- `content`, `checksum`

### Directory Resources

```ens
resource directory "/path/to/directory"
```

Represents a filesystem directory. Commonly used with:
- `exists`
- `permissions`
- Collection enforcement (`for each file in directory`)

### HTTP Resources

```ens
resource http "https://api.example.com/health"
```

Represents an HTTP/HTTPS endpoint. Supports:
- `reachable` - returns 2xx or 3xx status
- `status_code` - specific status code
- `tls` - valid TLS ≥ 1.2

### Cron Resources

```ens
resource cron "backup_job"
```

Represents a crontab entry. Use with:
- `scheduled` - ensures cron job exists with specific schedule and command

### Other Types (Parsed but Not Yet Implemented)

These are recognized by the parser but don't have handlers yet:

- `database` - database connections
- `service` - system services
- `process` - running processes

## Resource Declaration Syntax

### Inline Declaration

Declare resources directly in guarantees:

```ens
ensure exists on file "config.yaml"
```

### Named Resources (Aliases)

Declare once, reference many times:

```ens
resource file "/var/app/config.yaml" as config
resource file "/var/app/secrets.db" as secrets

ensure exists on config
ensure encrypted on secrets with AES:256 key "env:SECRET_KEY"
```

This is useful when you have many guarantees on the same resource.

### On Blocks

Group multiple guarantees on one resource:

```ens
on file "secrets.db" {
  ensure exists
  ensure encrypted with AES:256 key "env:SECRET_KEY"
  ensure permissions with posix mode "0600"
}
```

Inside an `on` block, the subject is implicit - you don't repeat `on file "secrets.db"` for each guarantee.

## Practical Examples

### Secure File Configuration

```ens
on file "/etc/app/database.conf" {
  ensure exists
  ensure permissions with posix mode "0600"
  ensure readable
  ensure writable
}
```

### API Health Monitoring

```ens
on http "https://api.production.com/health" {
  ensure reachable
  ensure tls
}

on violation {
  retry 3
  notify "ops-team"
}
```

### Scheduled Backup

```ens
on cron "daily_backup" {
  ensure scheduled with cron.native 
    schedule "0 2 * * *" 
    command "/usr/local/bin/backup.sh"
}
```

### Multiple Files with Aliases

```ens
resource file "/var/log/app.log" as applog
resource file "/var/log/error.log" as errlog
resource file "/var/log/access.log" as accesslog

ensure exists on applog
ensure exists on errlog
ensure exists on accesslog

ensure permissions with posix mode "0644" on applog
ensure permissions with posix mode "0644" on errlog
ensure permissions with posix mode "0600" on accesslog
```

## Next Steps

Now that you understand resources, learn about [Writing Guarantees](/learn/guarantees) to see all the ways you can use `ensure` statements.
