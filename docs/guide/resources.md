# Resources

Resources are the subjects of your guarantees - the things EnsuraScript can observe and modify.

## Declaring Resources

```ens
resource <type> "<path>" [as <alias>]
```

## Resource Types

### file

Represents a file on the filesystem.

```ens
resource file "secrets.db"
resource file "/etc/myapp/config.yaml" as app_config
```

**Applicable Conditions:** exists, readable, writable, encrypted, permissions, checksum, content

### directory

Represents a directory on the filesystem.

```ens
resource directory "/var/log/myapp"
resource directory "/secrets" as secrets_dir
```

**Applicable Conditions:** exists, permissions

### http

Represents an HTTP/HTTPS endpoint.

```ens
resource http "https://api.example.com/health"
resource http "https://auth.example.com/status" as auth_health
```

**Applicable Conditions:** reachable, status_code, tls

### database

Represents a database connection.

```ens
resource database "postgres://localhost/mydb"
```

**Applicable Conditions:** stable

### service

Represents a system service.

```ens
resource service "nginx"
```

**Applicable Conditions:** running, listening, healthy

### process

Represents a running process.

```ens
resource process "myapp"
```

**Applicable Conditions:** running, stopped

### cron

Represents a scheduled task.

```ens
resource cron "backup-job"
```

**Applicable Conditions:** scheduled

## Inline vs Declared Resources

You can reference resources inline without declaring them first:

```ens
# Inline reference
ensure exists on file "secrets.db"

# Or declare first
resource file "secrets.db" as secrets
ensure exists on secrets
```

Declaring resources with aliases is useful when:
- You reference the same resource multiple times
- You want more readable code
- You need to share resources across policies

## Resource Aliases

Use `as` to give resources memorable names:

```ens
resource file "/etc/myapp/production.yaml" as prod_config
resource http "https://api.prod.example.com/health" as prod_api

on prod_config {
  ensure exists
  ensure permissions with posix mode "0644"
}

ensure reachable on prod_api
```
