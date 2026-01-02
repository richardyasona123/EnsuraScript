# Policies

Policies let you bundle reusable guarantees and apply them consistently across your infrastructure.

## Defining Policies

```ens
policy <name>[(<params>)] {
  <statements>
}
```

### Simple Policy

```ens
policy secure_config {
  ensure permissions with posix mode "0644"
}
```

### Parameterized Policy

```ens
policy secure_file(key_ref) {
  ensure encrypted with AES:256 key key_ref
  ensure permissions with posix mode "0600"
}
```

## Applying Policies

```ens
apply <policy_name>[(<args>)]
```

### In Context Blocks

```ens
on file "secrets.db" {
  ensure exists
  apply secure_file("env:SECRET_KEY")
}
```

### Multiple Files

```ens
policy secure_file(key_ref) {
  ensure encrypted with AES:256 key key_ref
  ensure permissions with posix mode "0600"
}

on file "secrets.db" {
  ensure exists
  apply secure_file("env:DB_KEY")
}

on file "credentials.json" {
  ensure exists
  apply secure_file("env:CRED_KEY")
}

on file "tokens.yaml" {
  ensure exists
  apply secure_file("env:TOKEN_KEY")
}
```

## How Policies Work

When you apply a policy, EnsuraScript:

1. Looks up the policy definition
2. Substitutes parameters with provided arguments
3. Expands the policy statements in place
4. Binds the current subject to each statement

### Before Expansion

```ens
policy secure_file(key_ref) {
  ensure encrypted with AES:256 key key_ref
  ensure permissions with posix mode "0600"
}

on file "secrets.db" {
  ensure exists
  apply secure_file("env:SECRET_KEY")
}
```

### After Expansion

```ens
on file "secrets.db" {
  ensure exists
  ensure encrypted with AES:256 key "env:SECRET_KEY"
  ensure permissions with posix mode "0600"
}
```

## Best Practices

### Name Policies Clearly

```ens
# Good
policy secure_secrets_file(key) { ... }
policy public_config_file { ... }

# Avoid
policy p1(k) { ... }
```

### Keep Policies Focused

Each policy should have a single purpose:

```ens
# Good - focused policies
policy encrypted_file(key) {
  ensure encrypted with AES:256 key key
}

policy restricted_permissions {
  ensure permissions with posix mode "0600"
}

# Apply both
on file "secrets.db" {
  apply encrypted_file("env:KEY")
  apply restricted_permissions
}
```

### Document Parameters

Use comments to explain what parameters mean:

```ens
# Secures a file with AES-256 encryption
# key_ref: Environment variable or file path containing the encryption key
#          Examples: "env:SECRET_KEY", "file:/etc/keys/master.key"
policy secure_file(key_ref) {
  ensure encrypted with AES:256 key key_ref
  ensure permissions with posix mode "0600"
}
```
