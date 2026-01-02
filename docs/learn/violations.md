# Violation Handling

Configure what happens when guarantees can't be enforced using `on violation` blocks.

## What is a Violation?

A violation occurs when:
1. Check fails (guarantee not satisfied)
2. Enforce is attempted
3. Enforce fails (even after retries)

## Violation Handlers

### Global Violation Handler

Applies to all guarantees in the file:

```ens
on violation {
  retry 3
  notify "ops-team"
}

ensure exists on file "critical.db"
ensure encrypted on file "secrets.db"
```

If either guarantee fails after 3 retries, `ops-team` is notified.

### Per-Guarantee Violation Handler

Applies to a specific guarantee:

```ens
ensure exists on file "critical.db"
on violation {
  retry 10
  notify "critical-alerts"
}

ensure exists on file "logs.txt"
on violation {
  retry 2
  notify "dev-team"
}
```

The `critical.db` guarantee gets 10 retries and alerts the critical-alerts channel.

## Retry Syntax

```ens
retry <count>
```

Example:

```ens
on violation {
  retry 5
}
```

The runtime will:
1. Try to enforce
2. If it fails, wait (backoff)
3. Try again
4. Repeat up to 5 times total

## Notify Syntax

```ens
notify "<target>"
```

Example:

```ens
on violation {
  notify "ops"
}
```

::: warning Not Yet Implemented
The `notify` feature is parsed but not yet connected to notification channels (Slack, email, etc.). Currently, it's used for documentation and future integration.
:::

## Violation Handler Precedence

When a guarantee fails, the handler is chosen in this order:

1. **Per-guarantee handler** (highest priority)
2. **Global handler** (if no per-guarantee handler)
3. **CLI default** (from `--retries` flag or default 3)

## Practical Examples

### Critical Database File

```ens
ensure exists on file "/var/lib/db/primary.db"
on violation {
  retry 10
  notify "dba-team"
}

ensure encrypted on file "/var/lib/db/primary.db" with AES:256 key "env:DB_KEY"
on violation {
  retry 10
  notify "security-team"
}
```

### API Health Monitoring

```ens
on http "https://api.production.com/health" {
  ensure reachable
  ensure tls
}

on violation {
  retry 5
  notify "oncall"
}
```

### Mixed Criticality

```ens
# Global default for most guarantees
on violation {
  retry 3
  notify "dev-team"
}

# High-criticality override
ensure exists on file "/etc/app/license.key"
on violation {
  retry 20
  notify "critical-ops"
}

# Low-criticality uses global (3 retries, dev-team notification)
ensure exists on file "/var/log/debug.log"
```

## Violation Context

When a violation handler triggers, the runtime has access to:

- Which guarantee failed
- The resource it applies to
- The error message from the handler

Future versions will support templated notifications with this context.

## Backoff Strategy

::: warning Current Implementation
The current implementation uses a fixed backoff between retries. Future versions will support configurable backoff strategies (exponential, linear, etc.).
:::

## Best Practices

1. **Set realistic retry counts** - Don't retry forever, but give enough attempts for transient issues
2. **Notify appropriately** - Critical failures to oncall, low-priority to dev team
3. **Use per-guarantee handlers for critical resources** - Override global defaults where it matters
4. **Consider idempotency** - Ensure guarantees can be safely retried

## Example: Complete Application

```ens
# Global default
on violation {
  retry 3
  notify "dev-team"
}

# Critical license file
ensure exists on file "/etc/app/license.key"
on violation {
  retry 10
  notify "critical"
}

# Database must be encrypted
invariant {
  ensure encrypted on file "/var/lib/db/main.db" with AES:256 key "env:DB_KEY"
  on violation {
    retry 10
    notify "security-team"
  }
}

# HTTP health checks
for each endpoint in endpoints {
  ensure reachable
}
on violation {
  retry 5
  notify "ops"
}
```

## Next Steps

Continue to [Implication System](/learn/implications) for a deep dive into how guarantees automatically expand into prerequisites.
