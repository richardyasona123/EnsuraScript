# Implication System

The implication system is EnsuraScript's secret weapon - it automatically infers prerequisite guarantees so you don't have to write them.

## The Problem

In traditional automation, you must think about every step:

```bash
# Must remember ALL the steps
touch secrets.db           # 1. Create file
chmod +rw secrets.db       # 2. Make it readable/writable
encrypt secrets.db         # 3. Encrypt it
chmod 0600 secrets.db      # 4. Lock down permissions
```

Miss a step? Your automation breaks.

## The EnsuraScript Solution

Just declare the end state:

```ens
ensure encrypted on file "secrets.db" with AES:256 key "env:KEY"
```

EnsuraScript automatically infers:
1. File must exist (can't encrypt nothing)
2. File must be readable (can't read to encrypt)
3. File must be writable (can't write encrypted data)

## How It Works

### Implication Rules

The language has built-in implication rules:

```
encrypted → exists, readable, writable
readable → exists
writable → exists
permissions → exists
checksum → exists, readable
content → exists
```

The `→` means "implies". So `encrypted → exists` reads as "encrypted implies exists".

### Expansion Process

When you write:

```ens
ensure encrypted on file "secrets.db"
```

The **imply phase** expands it to:

```
1. ensure exists on file "secrets.db"      [IMPLIED]
2. ensure readable on file "secrets.db"    [IMPLIED]
3. ensure writable on file "secrets.db"    [IMPLIED]
4. ensure encrypted on file "secrets.db"   [ORIGINAL]
```

View this with:

```bash
ensura explain config.ens
```

Output:

```
Guarantees (4 total, 3 implied):

1. [IMPLIED] [fs.native] ensure exists on file "secrets.db"
2. [IMPLIED] [fs.native] ensure readable on file "secrets.db"
3. [IMPLIED] [fs.native] ensure writable on file "secrets.db"
4. [AES:256] ensure encrypted on file "secrets.db" with AES:256 key "env:KEY"
```

## Complete Implication Rules

From the codebase analysis:

| Condition | Implies |
|-----------|---------|
| `encrypted` | `exists`, `readable`, `writable` |
| `readable` | `exists` |
| `writable` | `exists` |
| `permissions` | `exists` |
| `checksum` | `exists`, `readable` |
| `content` | `exists` |
| `listening` | `running` |
| `healthy` | `running` |
| `status_code` | `reachable` |
| `tls` | `reachable` |
| `backed_up` | `exists` |

## Transitive Implications

Implications are transitive:

```ens
ensure checksum
```

Expands to:

```
1. ensure exists    [IMPLIED from readable]
2. ensure readable  [IMPLIED from checksum]
3. ensure checksum  [ORIGINAL]
```

Because `checksum → readable` and `readable → exists`.

## Conflict Detection

The implication system detects conflicting conditions:

```ens
ensure encrypted on file "data.txt"
ensure unencrypted on file "data.txt"  # CONFLICT!
```

Compilation error:

```
Error: Conflicting conditions on file "data.txt": encrypted vs unencrypted
```

Other conflicts:
- `running` vs `stopped`
- `encrypted` vs `unencrypted`

## Deduplication

If you manually write implied guarantees, they're deduplicated:

```ens
ensure exists on file "data.txt"
ensure readable on file "data.txt"
ensure encrypted on file "data.txt"
```

After implication expansion (with deduplication):

```
1. ensure exists       [MANUAL + IMPLIED]
2. ensure readable     [MANUAL + IMPLIED]
3. ensure writable     [IMPLIED from encrypted]
4. ensure encrypted    [ORIGINAL]
```

No duplicates - `exists` and `readable` aren't added twice.

## Execution Order

Implied guarantees are inserted **before** the original in the dependency graph:

```ens
ensure encrypted on file "secrets.db"
ensure backed_up on file "secrets.db"
```

Execution order:

```
1. exists (implied by encrypted)
2. readable (implied by encrypted)
3. writable (implied by encrypted)
4. encrypted (original)
5. backed_up (original)
```

This ensures prerequisites are satisfied before the main guarantee.

## Why This Matters

### Less Code

Without implications:

```ens
ensure exists on file "secrets.db"
ensure readable on file "secrets.db"
ensure writable on file "secrets.db"
ensure encrypted on file "secrets.db"
ensure permissions on file "secrets.db" with posix mode "0600"
```

With implications:

```ens
ensure encrypted on file "secrets.db"
ensure permissions on file "secrets.db" with posix mode "0600"
```

### Fewer Bugs

You can't forget a prerequisite - the language infers it.

### Better Intent Expression

`ensure encrypted` says what you want, not how to achieve it. The language figures out the "how".

## Implementation Details

The implication phase happens in the compilation pipeline:

```
Source → Lexer → Parser → Binder → Imply → Graph → Planner → Runtime
                                      ↑
                                  We are here
```

See the code in `pkg/imply/imply.go`.

The process:
1. Traverse the bound AST
2. For each `ensure` statement, look up implications
3. Generate new `ensure` statements for implied conditions
4. Insert them before the original in the AST
5. Continue to dependency graph building

## Next Steps

Continue to [Execution Model](/learn/execution) to understand topological sorting and continuous enforcement.
