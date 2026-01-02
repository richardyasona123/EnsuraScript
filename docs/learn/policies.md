# Creating Policies

Policies are reusable templates for guarantees. They let you define patterns once and apply them everywhere.

## Why Policies?

Without policies, you repeat yourself:

```ens
on file "database.db" {
  ensure encrypted with AES:256 key "env:DB_KEY"
  ensure permissions with posix mode "0600"
}

on file "secrets.env" {
  ensure encrypted with AES:256 key "env:SECRET_KEY"
  ensure permissions with posix mode "0600"
}

on file "api_keys.json" {
  ensure encrypted with AES:256 key "env:API_KEY"
  ensure permissions with posix mode "0600"
}
```

With policies, you define once and apply everywhere:

```ens
policy secure_file(key_ref) {
  ensure encrypted with AES:256 key key_ref
  ensure permissions with posix mode "0600"
}

on file "database.db" {
  apply secure_file("env:DB_KEY")
}

on file "secrets.env" {
  apply secure_file("env:SECRET_KEY")
}

on file "api_keys.json" {
  apply secure_file("env:API_KEY")
}
```

## Policy Syntax

### Defining a Policy

```ens
policy <name>(<param1>, <param2>, ...) {
  <statements>
}
```

Example:

```ens
policy secure_file(key_ref) {
  ensure exists
  ensure encrypted with AES:256 key key_ref
  ensure permissions with posix mode "0600"
}
```

### Applying a Policy

```ens
apply <name>(<arg1>, <arg2>, ...)
```

Inside an `on` block:

```ens
on file "secrets.db" {
  apply secure_file("env:SECRET_KEY")
}
```

## Parameters

Policies can take parameters that get substituted when applied.

### Single Parameter

```ens
policy set_mode(mode_value) {
  ensure permissions with posix mode mode_value
}

on file "script.sh" {
  apply set_mode("0755")
}

on file "data.txt" {
  apply set_mode("0644")
}
```

### Multiple Parameters

```ens
policy encrypted_file(key_ref, mode_value) {
  ensure encrypted with AES:256 key key_ref
  ensure permissions with posix mode mode_value
}

on file "public_data.enc" {
  apply encrypted_file("env:PUBLIC_KEY", "0644")
}

on file "private_data.enc" {
  apply encrypted_file("env:PRIVATE_KEY", "0600")
}
```

## Policy Expansion

When you apply a policy, EnsuraScript expands it into the actual guarantees during the binding phase.

Given:

```ens
policy secure(key_ref) {
  ensure encrypted with AES:256 key key_ref
  ensure permissions with posix mode "0600"
}

on file "secrets.db" {
  apply secure("env:SECRET_KEY")
}
```

After expansion (what EnsuraScript actually executes):

```ens
on file "secrets.db" {
  ensure encrypted with AES:256 key "env:SECRET_KEY"
  ensure permissions with posix mode "0600"
}
```

Use `ensura explain` to see expanded policies:

```bash
ensura explain config.ens
```

## Practical Examples

### Web Application Security

```ens
policy secure_web_file(mode_value) {
  ensure exists
  ensure permissions with posix mode mode_value
  ensure readable
}

policy secure_secret_file(key_ref) {
  ensure exists
  ensure encrypted with AES:256 key key_ref
  ensure permissions with posix mode "0600"
}

# Public web assets
on file "/var/www/index.html" {
  apply secure_web_file("0644")
}

on file "/var/www/app.js" {
  apply secure_web_file("0644")
}

# Secret configuration
on file "/etc/app/database.conf" {
  apply secure_secret_file("env:DB_CONFIG_KEY")
}

on file "/etc/app/api_keys.env" {
  apply secure_secret_file("env:API_KEY")
}
```

### API Health Monitoring

```ens
policy monitor_api(expected_code) {
  ensure reachable
  ensure status_code with http.get expected_status expected_code
  ensure tls
}

on http "https://api.production.com/health" {
  apply monitor_api("200")
}

on http "https://api.production.com/ready" {
  apply monitor_api("200")
}

on http "https://webhooks.production.com/ping" {
  apply monitor_api("204")
}
```

### Scheduled Jobs

```ens
policy daily_job(job_name, hour, command_path) {
  ensure scheduled with cron.native 
    schedule ("0 " + hour + " * * *")
    command command_path
}

on cron "backup" {
  apply daily_job("backup", "2", "/usr/local/bin/backup.sh")
}

on cron "cleanup" {
  apply daily_job("cleanup", "3", "/usr/local/bin/cleanup.sh")
}
```

Note: String concatenation is not currently supported - this example shows the intended design.

## Policy Best Practices

1. **Make policies generic** - Use parameters for values that change
2. **Name descriptively** - `secure_file` is better than `sf`
3. **Group related guarantees** - Policies should represent coherent patterns
4. **Document parameters** - Use comments to explain what parameters do

Example with documentation:

```ens
# Secures a file with encryption and restrictive permissions
# Parameters:
#   key_ref - Key reference (env:VAR, file:/path, or literal)
policy secure_file(key_ref) {
  ensure exists
  ensure encrypted with AES:256 key key_ref
  ensure permissions with posix mode "0600"
}
```

## Policies vs. Implication

Policies and implication serve different purposes:

- **Implication** - Automatic prerequisite inference (e.g., `encrypted` → `exists`)
- **Policies** - Explicit reusable patterns you define

Both work together:

```ens
policy secure(key_ref) {
  ensure encrypted with AES:256 key key_ref  # encrypted implies exists, readable, writable
  ensure permissions with posix mode "0600"
}
```

When applied, both policy expansion AND implication happen, giving you all necessary guarantees.

## Next Steps

Continue to [Guards & Conditions](/learn/guards) to learn how to make guarantees conditional.
