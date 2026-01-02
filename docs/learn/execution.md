# Execution Model

Understand how EnsuraScript orders guarantees and enforces them continuously.

## The Execution Pipeline

When you run `ensura run config.ens`, this happens:

```
1. Lexer       → Tokenize source code
2. Parser      → Build AST
3. Binder      → Resolve implicit subjects, expand policies
4. Imply       → Add implied guarantees
5. Graph       → Build dependency graph
6. Topo Sort   → Determine execution order
7. Planner     → Create sequential execution plan
8. Runtime     → Execute plan in continuous loop
```

## Dependency Graph

### Graph Construction

The graph builder collects all guarantees and creates edges for:

1. **Implication dependencies** - Implied conditions → original
2. **Explicit requires** - `ensure A requires B` creates edge B → A
3. **Temporal ordering** - `ensure A after B` creates edge B → A
4. **Priorities** - Invariants get +1000 priority

Each guarantee gets a unique ID:

```
<condition>:<subject>@<position>
```

Example:

```
exists:file("/var/app/config.yaml")@line12
encrypted:file("/var/app/secrets.db")@line15
```

### Graph Edges

Given:

```ens
ensure encrypted on file "secrets.db"  # Line 5
ensure backed_up on file "secrets.db" requires encrypted  # Line 6
```

Graph edges:

```
exists:file("secrets.db")@5 → encrypted:file("secrets.db")@5
readable:file("secrets.db")@5 → encrypted:file("secrets.db")@5
writable:file("secrets.db")@5 → encrypted:file("secrets.db")@5
encrypted:file("secrets.db")@5 → backed_up:file("secrets.db")@6
```

## Topological Sort

### Kahn's Algorithm

EnsuraScript uses Kahn's algorithm for topological sorting:

1. Start with all nodes that have no incoming edges
2. Remove a node and add to output
3. Remove all edges from that node
4. Repeat until graph is empty
5. If graph is not empty → cycle detected → ERROR

### Priority-Based Tie-Breaking

When multiple nodes have no incoming edges, choose by priority:

- Invariants: priority 1000
- Regular guarantees: priority 0

Higher priority = executed first.

### Cycle Detection

If dependencies create a cycle:

```ens
ensure A requires B
ensure B requires C
ensure C requires A
```

Compilation fails:

```
Error: Dependency cycle detected: A → B → C → A
```

## Execution Plan

The planner takes the topologically sorted guarantees and creates a sequential plan:

```bash
ensura plan config.ens
```

Output:

```
Execution Plan (6 steps):

1. [fs.native] ensure exists on file "secrets.db"
2. [fs.native] ensure readable on file "secrets.db"
3. [fs.native] ensure writable on file "secrets.db"
4. [AES:256] ensure encrypted on file "secrets.db" with AES:256 key "env:KEY"
5. [posix] ensure permissions on file "secrets.db" with posix mode "0600"
6. [fs.native] ensure backed_up on file "secrets.db"
```

This is the exact order the runtime executes.

## Runtime Loop

### Continuous Enforcement

```
loop forever:
  for each guarantee in plan:
    check()
    if not satisfied:
      if dry_run:
        mark VIOLATED
      else:
        enforce()
        recheck()
        if still not satisfied:
          retry up to max_retries
          if still failed:
            trigger violation handler
            mark FAILED
        else:
          mark REPAIRED
    else:
      mark SATISFIED
  
  sleep(interval)
```

### Configuration

- `--interval` - Time between loops (default: 30s)
- `--retries` - Max retry attempts (default: 3)
- `--dry-run` - Check only, no enforcement

### Check vs. Enforce

**Check**:
- Read-only operation
- Returns true if guarantee is satisfied
- Example: Does file exist? Is it encrypted?

**Enforce**:
- Makes the guarantee true
- Can modify the system
- Example: Create file, encrypt it

**Re-check**:
- After enforcement, verify it worked
- If not, retry or trigger violation handler

## Example Execution

Given:

```ens
on file "secrets.db" {
  ensure exists
  ensure encrypted with AES:256 key "env:SECRET_KEY"
  ensure permissions with posix mode "0600"
}
```

### First Loop (file doesn't exist)

```
Step 1: ensure exists
  Check: os.Stat("secrets.db") → NOT FOUND
  Enforce: os.Create("secrets.db")
  Re-check: os.Stat("secrets.db") → FOUND
  Status: REPAIRED

Step 2: ensure encrypted
  Check: Read file, look for magic header → NOT FOUND
  Enforce: Read plaintext, encrypt with AES-256-GCM, write with magic header
  Re-check: Read file, look for magic header → FOUND
  Status: REPAIRED

Step 3: ensure permissions
  Check: os.Stat().Mode() → 0644 (default from create)
  Enforce: os.Chmod("secrets.db", 0600)
  Re-check: os.Stat().Mode() → 0600
  Status: REPAIRED

Sleep 30 seconds...
```

### Second Loop (file exists, all satisfied)

```
Step 1: ensure exists
  Check: os.Stat("secrets.db") → FOUND
  Status: SATISFIED

Step 2: ensure encrypted
  Check: Read file, look for magic header → FOUND
  Status: SATISFIED

Step 3: ensure permissions
  Check: os.Stat().Mode() → 0600
  Status: SATISFIED

Sleep 30 seconds...
```

### Third Loop (user changed permissions)

```
Step 1: ensure exists
  Check: SATISFIED

Step 2: ensure encrypted
  Check: SATISFIED

Step 3: ensure permissions
  Check: os.Stat().Mode() → 0777 (user changed it!)
  Enforce: os.Chmod("secrets.db", 0600)
  Re-check: os.Stat().Mode() → 0600
  Status: REPAIRED

Sleep 30 seconds...
```

This is **drift detection** and **automatic remediation**.

## Parallel Execution (Future)

Currently, guarantees execute sequentially in topological order.

The `parallel` block is parsed but not yet executed in parallel:

```ens
parallel {
  ensure reachable on http "https://api1.example.com"
  ensure reachable on http "https://api2.example.com"
  ensure reachable on http "https://api3.example.com"
}
```

Future versions will execute independent guarantees concurrently.

## Performance Considerations

- **Graph building** - O(V + E) where V = guarantees, E = dependencies
- **Topological sort** - O(V + E)
- **Runtime loop** - O(V) per iteration
- **Check operations** - Typically O(1) (file stat, HTTP GET, etc.)
- **Enforce operations** - Varies (file creation is fast, encryption is slower)

For 100 guarantees with 200 dependencies:
- Compilation: < 100ms
- Runtime loop: ~1-5s depending on handlers

## Observability

View what's happening:

```bash
# See execution order
ensura plan config.ens

# See expanded guarantees
ensura explain config.ens

# See dependency graph as DOT
ensura compile config.ens --graph

# Run with verbose output
ensura run config.ens --verbose
```

## Next Steps

You've completed the Learn section! Continue to the [Reference](/reference/syntax) section for complete syntax documentation.
