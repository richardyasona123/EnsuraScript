# Syntax Reference

Complete EnsuraScript language syntax reference.

## Program Structure

An EnsuraScript program consists of statements:

```ens
<statement>*
```

Statements include:
- Resource declarations
- Ensure statements
- On blocks
- Policy declarations
- Apply statements
- For-each loops
- Invariant blocks
- Violation handlers
- Assumptions

## Comments

```ens
# This is a comment
```

Line comments start with `#` and continue to end of line.

## Resource Declarations

### Basic Declaration

```ens
resource <type> "<path>"
```

**Types:** `file`, `directory`, `http`, `cron`, `database`, `service`, `process`

Examples:

```ens
resource file "/etc/app/config.yaml"
resource directory "/var/app/data"
resource http "https://api.example.com"
resource cron "backup_job"
```

### Named Resources (Aliases)

```ens
resource <type> "<path>" as <identifier>
```

Examples:

```ens
resource file "/var/app/secrets.db" as secrets
resource http "https://api.example.com/health" as api_health
```

Reference by alias:

```ens
ensure exists on secrets
ensure reachable on api_health
```

## Ensure Statements

### Full Syntax

```ens
ensure <condition> [on <resource>] [with <handler> <args>] [when <guard>] [requires <condition>] [after <resource>] [before <resource>]
```

All clauses except `<condition>` are optional.

### Basic Ensure

```ens
ensure <condition> on <resource>
```

Example:

```ens
ensure exists on file "config.yaml"
```

### With Handler

```ens
ensure <condition> with <handler> <key> "<value>" ... on <resource>
```

Example:

```ens
ensure encrypted with AES:256 key "env:SECRET_KEY" on file "secrets.db"
ensure permissions with posix mode "0600" on file "secrets.db"
```

### When Clause (Guards)

```ens
ensure <condition> when <identifier> <op> "<value>"
```

Operators: `==`, `!=`

Example:

```ens
ensure encrypted when environment == "prod"
ensure permissions with posix mode "0644" when environment != "prod"
```

### Requires Clause

```ens
ensure <condition-A> requires <condition-B>
```

Example:

```ens
ensure backed_up requires encrypted
```

### After/Before Clauses

```ens
ensure <condition> after <resource>
ensure <condition> before <resource>
```

Examples:

```ens
ensure startup on service "app" after file "config.yaml"
ensure exists on file "log.txt" before service "app"
```

### Implicit Subject

Inside an `on` block, `on <resource>` can be omitted:

```ens
on file "secrets.db" {
  ensure exists          # on file "secrets.db" is implicit
  ensure encrypted       # on file "secrets.db" is implicit
}
```

## On Blocks

Group multiple statements on one resource:

```ens
on <resource> {
  <statement>*
}
```

Example:

```ens
on file "secrets.db" {
  ensure exists
  ensure encrypted with AES:256 key "env:SECRET_KEY"
  ensure permissions with posix mode "0600"
}
```

## Policy Declarations

Define reusable guarantee templates:

```ens
policy <name>(<param1>, <param2>, ...) {
  <statement>*
}
```

Example:

```ens
policy secure_file(key_ref) {
  ensure encrypted with AES:256 key key_ref
  ensure permissions with posix mode "0600"
}
```

## Apply Statements

Apply a policy:

```ens
apply <name>(<arg1>, <arg2>, ...)
```

Example:

```ens
on file "database.db" {
  apply secure_file("env:DB_KEY")
}
```

## For-Each Loops

Iterate over collections:

```ens
for each <type> in <container> {
  <statement>*
}
```

Example:

```ens
for each file in directory "/secrets" {
  ensure encrypted with AES:256 key "env:SECRET_KEY"
}
```

## Invariant Blocks

High-priority guarantees:

```ens
invariant {
  <statement>*
}
```

Example:

```ens
invariant {
  ensure exists on file "/etc/app/license.key"
}
```

## Violation Handlers

### Global

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
  notify "ops-team"
}
```

### Per-Ensure

Place immediately after an `ensure` statement:

```ens
ensure exists on file "critical.db"
on violation {
  retry 10
  notify "critical-alerts"
}
```

## Assumptions

Declare assumed values for guards:

```ens
assume <identifier> == "<value>"
assume <simple-statement>
```

Examples:

```ens
assume environment == "prod"
assume filesystem reliable
```

## Parallel Blocks (Parsed, Not Yet Executed)

```ens
parallel {
  <statement>*
}
```

Example:

```ens
parallel {
  ensure reachable on http "https://api1.example.com"
  ensure reachable on http "https://api2.example.com"
}
```

## Literals

### Strings

```ens
"double quoted strings"
```

Strings can contain spaces, special characters. No escape sequences currently.

### Numbers

```ens
123
456
```

Used in retry counts.

### Identifiers

```ens
myalias
environment
region
feature_flag
```

Used for resource aliases and guard identifiers.

## Keywords

Reserved keywords:

```
resource, ensure, on, with, requires, after, before
policy, apply, violation, retry, notify
assume, when, for, each, in, invariant, as
key, mode, directory, file, http, database
service, process, cron, environment, parallel
```

## Operators

- `==` - Equality (guards)
- `!=` - Inequality (guards)
- `:` - Handler separator (e.g., `AES:256`)

## Complete Example

```ens
# Global violation handler
on violation {
  retry 3
  notify "dev-team"
}

# Policy definition
policy secure_file(key_ref) {
  ensure encrypted with AES:256 key key_ref
  ensure permissions with posix mode "0600"
}

# Named resource
resource file "/var/app/database.db" as db

# Invariant with for-each
invariant {
  for each file in directory "/var/app/secrets" {
    apply secure_file("env:SECRET_KEY")
  }
}

# Regular guarantees with guards
on db {
  ensure exists
  apply secure_file("env:DB_KEY") when environment == "prod"
  ensure permissions with posix mode "0644" when environment != "prod"
}

# HTTP monitoring
on http "https://api.production.com/health" {
  ensure reachable
  ensure tls
}
on violation {
  retry 5
  notify "oncall"
}

# Cron scheduling
on cron "daily_backup" {
  ensure scheduled with cron.native 
    schedule "0 2 * * *" 
    command "/usr/local/bin/backup.sh"
}
```

## Grammar (EBNF)

```ebnf
program ::= statement*

statement ::= resource_decl
            | ensure_stmt
            | on_block
            | policy_decl
            | apply_stmt
            | for_each_stmt
            | invariant_block
            | on_violation_block
            | assume_stmt
            | parallel_block

resource_decl ::= "resource" resource_type string ["as" identifier]

ensure_stmt ::= "ensure" condition [ensure_clauses] [on_violation_block]

ensure_clauses ::= ["on" resource_ref] 
                   ["with" handler_spec] 
                   ["when" guard_expr]
                   ["requires" condition]
                   ["after" resource_ref]
                   ["before" resource_ref]

on_block ::= "on" resource_ref "{" statement* "}"

policy_decl ::= "policy" identifier "(" [param_list] ")" "{" statement* "}"

apply_stmt ::= "apply" identifier "(" [arg_list] ")"

for_each_stmt ::= "for" "each" resource_type "in" resource_ref "{" statement* "}"

invariant_block ::= "invariant" "{" statement* "}"

on_violation_block ::= "on" "violation" "{" violation_handler* "}"

violation_handler ::= "retry" number | "notify" string

assume_stmt ::= "assume" (guard_expr | identifier)

parallel_block ::= "parallel" "{" statement* "}"

resource_ref ::= resource_type string | identifier

resource_type ::= "file" | "directory" | "http" | "cron" | "database" | "service" | "process"

handler_spec ::= identifier [":" number] [handler_args]

handler_args ::= (identifier string)*

guard_expr ::= identifier ("==" | "!=") string

condition ::= identifier

param_list ::= identifier ("," identifier)*

arg_list ::= string ("," string)*
```
