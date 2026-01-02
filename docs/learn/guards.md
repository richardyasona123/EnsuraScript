# Guards & Conditions

Guards let you make guarantees conditional using `when` clauses.

## Guard Syntax

```ens
ensure <condition> when <identifier> == "<value>"
ensure <condition> when <identifier> != "<value>"
```

## Basic Example

```ens
ensure encrypted with AES:256 key "env:PROD_KEY" on file "secrets.db" when environment == "prod"
ensure permissions with posix mode "0644" on file "secrets.db" when environment != "prod"
```

In production (`environment == "prod"`):
- File is encrypted with production key

In development (`environment != "prod"`):
- File has relaxed permissions for easy access

## Guard Operators

Currently supported operators:

- `==` - Equals
- `!=` - Not equals

## Common Guard Patterns

### Environment-Based Configuration

```ens
on file "app.conf" {
  ensure content with fs.native content "debug=true" when environment == "dev"
  ensure content with fs.native content "debug=false" when environment == "prod"
}
```

### Region-Specific Settings

```ens
on http "https://api.us-east.example.com" {
  ensure reachable when region == "us-east"
}

on http "https://api.eu-west.example.com" {
  ensure reachable when region == "eu-west"
}
```

### Feature Flags

```ens
on file "/etc/app/experimental.conf" {
  ensure exists when feature_flag == "enabled"
}
```

## Guards in On Blocks

Guards work inside `on` blocks:

```ens
on file "database.conf" {
  ensure exists
  ensure encrypted with AES:256 key "env:DB_KEY" when environment == "prod"
  ensure permissions with posix mode "0600" when environment == "prod"
  ensure permissions with posix mode "0644" when environment != "prod"
}
```

## Guards with Policies

Combine guards with policies for maximum flexibility:

```ens
policy secure_file(key_ref) {
  ensure encrypted with AES:256 key key_ref
  ensure permissions with posix mode "0600"
}

policy dev_file {
  ensure permissions with posix mode "0644"
}

on file "secrets.db" {
  apply secure_file("env:SECRET_KEY") when environment == "prod"
  apply dev_file when environment == "dev"
}
```

## Guard Evaluation

::: warning Static Evaluation
Currently, guards are evaluated at **parse time**, not runtime. This means guard values must be known before execution begins.

Future versions may support runtime evaluation.
:::

## Setting Guard Values

Guards check against identifiers like `environment`, `region`, `feature_flag`. How do you set these?

Currently, you set them via **assumptions**:

```ens
assume environment == "prod"

ensure encrypted when environment == "prod"
```

The `assume` statement declares the value of an identifier for guard evaluation.

Future versions will support reading from environment variables and configuration files.

## Multiple Guards

You can have multiple guards on different guarantees:

```ens
ensure encrypted when environment == "prod"
ensure encrypted when security_level == "high"
ensure permissions with posix mode "0600" when tier == "critical"
```

## Next Steps

Continue to [Dependencies](/learn/dependencies) to learn how to control the order of guarantee execution.
