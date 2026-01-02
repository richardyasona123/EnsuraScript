# Syntax Reference

## File Format

- Extension: `.ens`
- Encoding: UTF-8
- Comments: Line comments with `#`

```ens
# This is a comment
resource file "example.txt"  # Inline comment
```

## Strings

Strings use double quotes:

```ens
"path/to/file"
"https://example.com"
"env:SECRET_KEY"
```

## Identifiers

Use `lower_snake_case` for identifiers:

```ens
resource file "secrets.db" as my_secrets
policy secure_file { ... }
```

## Resource Declarations

```ens
resource <type> "<path>" [as <alias>]
```

Examples:
```ens
resource file "secrets.db"
resource http "https://example.com/health" as api_health
resource directory "/var/log/myapp"
```

## Ensure Statements

```ens
ensure <condition> [on <resource>] [with <handler> <args>] [when <guard>] [requires <deps>]
```

Examples:
```ens
ensure exists on file "secrets.db"
ensure encrypted on file "secrets.db" with AES:256 key "env:SECRET_KEY"
ensure reachable on http "https://example.com" when environment == "prod"
ensure backed_up on file "secrets.db" requires encrypted
```

## Context Blocks

```ens
on <resource> {
  <statements>
}
```

Example:
```ens
on file "secrets.db" {
  ensure exists
  ensure encrypted with AES:256 key "env:SECRET_KEY"
}
```

## Policies

```ens
policy <name>[(<params>)] {
  <statements>
}

apply <name>[(<args>)]
```

Example:
```ens
policy secure_file(key_ref) {
  ensure encrypted with AES:256 key key_ref
  ensure permissions with posix mode "0600"
}

on file "secrets.db" {
  apply secure_file("env:SECRET_KEY")
}
```

## For Each Loops

```ens
for each <type> in <container> {
  <statements>
}
```

Example:
```ens
for each file in directory "/secrets" {
  ensure encrypted with AES:256 key "env:SECRET_KEY"
}
```

## Invariants

```ens
invariant {
  <statements>
}
```

Example:
```ens
invariant {
  for each file in directory "/secrets" {
    ensure encrypted with AES:256 key "env:SECRET_KEY"
  }
}
```

## Violation Handling

```ens
on violation {
  retry <count>
  notify "<target>"
}
```

Example:
```ens
on violation {
  retry 3
  notify "ops"
  notify "security"
}
```

## Guards

```ens
ensure <condition> when <var> == "<value>"
ensure <condition> when <var> != "<value>"
```

Example:
```ens
ensure encrypted on file "secrets.db" when environment == "prod"
```

## Assumptions

```ens
assume <var> == "<value>"
assume <simple_statement>
```

Example:
```ens
assume environment == "prod"
assume filesystem reliable
```
