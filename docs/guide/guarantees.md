# Guarantees

Guarantees are the core of EnsuraScript - declarative constraints that must be satisfied.

## The Ensure Statement

```ens
ensure <condition> [on <resource>] [with <handler> <args>] [when <guard>]
```

### Basic Guarantee

```ens
ensure exists on file "config.yaml"
```

### With Handler

```ens
ensure encrypted on file "secrets.db" with AES:256 key "env:SECRET_KEY"
```

### With Guard

```ens
ensure encrypted on file "secrets.db" when environment == "prod"
```

## Conditions

Conditions describe what should be true about a resource.

### Filesystem Conditions

| Condition | Description | Implies |
|-----------|-------------|---------|
| `exists` | Resource exists | - |
| `readable` | File is readable | exists |
| `writable` | File is writable | exists |
| `encrypted` | File is encrypted | exists, readable, writable |
| `permissions` | File has specific permissions | exists |
| `checksum` | File has specific checksum | exists, readable |
| `content` | File has specific content | exists |

### HTTP Conditions

| Condition | Description | Implies |
|-----------|-------------|---------|
| `reachable` | Endpoint is reachable | - |
| `status_code` | Returns specific status | reachable |
| `tls` | Uses valid TLS | reachable |

### Service Conditions

| Condition | Description | Implies |
|-----------|-------------|---------|
| `running` | Service is running | - |
| `stopped` | Service is stopped | - |
| `listening` | Service is listening | running |
| `healthy` | Service is healthy | running |

## Dependencies

### Requires

Specify that one guarantee depends on another:

```ens
ensure backed_up on file "secrets.db" requires encrypted
```

### After/Before

Specify ordering without hard dependencies:

```ens
ensure backed_up on file "secrets.db" after database "main" stable
```

## Context Blocks

Group guarantees for a single resource:

```ens
on file "secrets.db" {
  ensure exists
  ensure encrypted with AES:256 key "env:SECRET_KEY"
  ensure permissions with posix mode "0600"
}
```

Statements inside the block inherit the subject.

## How Guarantees Are Processed

1. **Parse**: Read the ensure statements
2. **Bind**: Resolve implicit subjects
3. **Expand**: Add implied prerequisites
4. **Order**: Topologically sort by dependencies
5. **Execute**: Check and repair in order
6. **Loop**: Continuously monitor for drift
