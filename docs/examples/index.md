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

[Full Example →](/examples/file-security)

## HTTP Health Monitoring

Monitor HTTP endpoints for availability:

```ens
resource http "https://api.example.com/health" as api_health

ensure reachable on http "https://api.example.com/health"
ensure status_code on http "https://api.example.com/health" with http expected_status "200"
ensure tls on http "https://api.example.com/health"

on violation {
  retry 5
  notify "oncall"
}
```

[Full Example →](/examples/http-monitoring)

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

[Full Example →](/examples/policy-reuse)

## Collection Enforcement

Enforce guarantees across all files in a directory:

```ens
invariant {
  for each file in directory "/secrets" {
    ensure encrypted with AES:256 key "env:SECRET_KEY"
  }
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
