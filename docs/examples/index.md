# Examples

This section contains practical examples of using EnsuraScript for common use cases.

## File Security

Ensure files are encrypted and have proper permissions:

```ens
resource file "secrets.db"

on file "secrets.db" {
  ensure exists
  ensure encrypted with AES:256 key "env:SECRET_KEY"
  ensure permissions with posix mode "0600"
}

on violation {
  retry 2
  notify "ops"
}
```

## HTTP Health Monitoring

Monitor HTTP endpoints for availability:

```ens
resource http "https://api.example.com/health" as api_health

on http "https://api.example.com/health" {
  ensure reachable
  ensure status_code with http.get expected_status "200"
  ensure tls
}

on violation {
  retry 5
  notify "oncall"
}
```

## Policy Reuse

Define reusable security policies:

```ens
policy secure_file(key_ref) {
  ensure encrypted with AES:256 key key_ref
  ensure permissions with posix mode "0600"
}

on file "secrets.db" {
  ensure exists
  apply secure_file("env:SECRET_KEY")
}

on file "credentials.json" {
  ensure exists
  apply secure_file("env:CRED_KEY")
}
```

## Collection Enforcement

Enforce guarantees across all files in a directory:

```ens
invariant {
  for each file in directory "/secrets" {
    ensure encrypted with AES:256 key "env:SECRET_KEY"
  }
}
```

## Scheduled Tasks

Set up automated cron jobs:

```ens
# Daily backup at 2 AM
on cron "daily_backup" {
  ensure scheduled with cron.native schedule "0 2 * * *" command "/usr/local/bin/backup.sh"
}

# Health check every 15 minutes
on cron "health_monitor" {
  ensure scheduled with cron.native schedule "*/15 * * * *" command "/usr/local/bin/health_check.sh"
}
```

## Conditional Guarantees

Apply different guarantees based on environment:

```ens
assume environment == "prod"

on file "secrets.db" {
  ensure exists
  ensure encrypted with AES:256 key "env:SECRET_KEY" when environment == "prod"
  ensure permissions with posix mode "0600" when environment == "prod"
  ensure permissions with posix mode "0644" when environment == "dev"
}
```
