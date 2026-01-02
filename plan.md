# EnsuraScript — Full Language Plan

EnsuraScript is an OSS, intent-first, “truth maintenance” language: you *declare* desired properties of systems; the runtime *satisfies and keeps them true*.

---

## 0. Design Goals

**Primary goals**

* **Intent over instruction:** users state outcomes (“ensure X”), not steps.
* **Deterministic:** no AI guessing. All inference is rule-based and inspectable.
* **Sequential enforcement:** the resolver executes a linear plan (one action at a time) derived from a dependency graph.
* **Continuous:** guarantees are re-checked; drift triggers repair.
* **Composable:** adapters/handlers implement enforcement; policies bundle reusable intent.
* **Readable:** Skript-like surface syntax.

**Non-goals (v1)**

* General-purpose programming (loops/vars/functions like Python).
* Arbitrary concurrency.
* Unbounded theorem-proving.

---

## 1. Core Concepts

### 1.1 Resources

A **resource** is something EnsuraScript can observe and/or change.
Examples: file, directory, service, process, database, http endpoint, cron schedule.

### 1.2 Guarantees

A **guarantee** is a constraint that must be true.
Written with `ensure`.

### 1.3 Handlers

A **handler** is a plugin implementation that can check/enforce a guarantee.
Selected via `with` (explicit) or by default resolution if unambiguous.

### 1.4 Implication

Guarantees can **imply prerequisites**.
Example: `ensure encrypted ...` implies `ensure exists ...`.

### 1.5 Context Binding

Statements inherit the **current subject** (resource) when unambiguous.

### 1.6 Deterministic Resolver

Source → AST → intent graph → implied expansions → de-dup → topo sort → **sequential** execution plan → enforcement loop.

---

## 2. File Format & Conventions

* Extension: `.ens`
* Encoding: UTF-8
* Comments:

  * Line comment: `# comment`
* Strings:

  * Double quotes: `"like this"`
* Identifiers:

  * `lower_snake_case` recommended

---

## 3. Keywords (v1)

### Top-level

* `resource`
* `ensure`
* `on`
* `with`
* `requires`
* `after`
* `before`
* `policy`
* `apply`
* `on violation`
* `retry`
* `notify`
* `assume`
* `when`
* `for each`
* `invariant`

### Block/Scope

* `{` `}`

---

## 4. Syntax Overview

### 4.1 Resource Declarations

**Named resource**

```ens
resource file "secrets.db"
resource database "main"
resource http "https://example.com/health"
```

Optional aliases (v1 supports `as`)

```ens
resource file "secrets.db" as secrets_db
```

### 4.2 Ensure Statements

Canonical form:

```ens
ensure <condition> on <resource_ref> [with <handler_spec>] [when <guard_expr>] [requires <dep_list>] [after <ref>] [before <ref>]
```

Examples:

```ens
ensure exists on file "secrets.db"
ensure encrypted on file "secrets.db" with AES:256 key "env:SECRET_KEY"
ensure reachable on http "https://example.com/health"
```

### 4.3 Context Carry-Forward (Skript-like)

Within an `on` block, the subject is fixed:

```ens
on file "secrets.db" {
  ensure exists
  ensure encrypted with AES:256 key "env:SECRET_KEY"
}
```

Outside blocks, last subject is inherited only if unambiguous:

```ens
ensure exists on file "secrets.db"
ensure encrypted with AES:256 key "env:SECRET_KEY"  # binds to file "secrets.db"
```

If ambiguous, compile error.

### 4.4 Policies

Define:

```ens
policy secure_file {
  ensure encrypted with AES:256 key "env:SECRET_KEY"
  ensure permissions with posix mode "0600"
}
```

Apply:

```ens
on file "secrets.db" {
  ensure exists
  apply secure_file
}
```

Policies can accept parameters (v1 supports positional params):

```ens
policy secure_file(key_ref) {
  ensure encrypted with AES:256 key key_ref
  ensure permissions with posix mode "0600"
}

on file "secrets.db" {
  apply secure_file("env:SECRET_KEY")
}
```

### 4.5 Dependencies & Ordering

Hard dependencies:

```ens
ensure backed_up on file "secrets.db" requires encrypted
```

Explicit resource dependency:

```ens
ensure backed_up on file "secrets.db" requires database "main" stable
```

Ordering constraints:

```ens
ensure backed_up on file "secrets.db" after database "main" stable
```

Notes:

* `requires` means “cannot proceed until satisfied.”
* `after/before` means temporal ordering among guarantees.

### 4.6 Guards (Conditional Activation)

```ens
ensure encrypted on file "secrets.db" when environment == "prod"
```

Guards control whether a guarantee exists in the active intent graph.

### 4.7 Quantifiers (Collections)

```ens
for each file in directory "/secrets" {
  ensure encrypted with AES:256 key "env:SECRET_KEY"
}
```

Rules:

* Expands to per-item guarantees.
* Enumeration is provided by an adapter capability (`directory.enumerate`).

### 4.8 Invariants (Global Assertions)

```ens
invariant {
  for each file in directory "/secrets" {
    ensure encrypted with AES:256 key "env:SECRET_KEY"
  }
}
```

Invariants are higher priority; violations are considered critical.

### 4.9 Violation Handling

Global handler block:

```ens
on violation {
  retry 2
  notify "ops"
}
```

Per-ensure overrides:

```ens
ensure encrypted on file "secrets.db" with AES:256 key "env:SECRET_KEY"
on violation {
  retry 5
  notify "security"
}
```

Semantics:

* If check fails, run repair plan, then re-check.
* If repair fails, apply retries.
* After retries exhausted: emit incident + mark guarantee unresolved.

### 4.10 Assumptions

```ens
assume environment == "dev"
assume filesystem reliable
```

Assumptions:

* Used to simplify enforcement.
* If an assumption is violated at runtime (when checkable), it becomes a violation.

---

## 5. Conditions (v1 Standard Library)

Conditions are nouns/adjectives checked on a resource. Each condition declares:

* applicable resource types
* implied prerequisites
* conflicts
* default handler (optional)

### Filesystem

* `exists` (file, directory)
* `encrypted` (file) — implies `exists`, `readable`, `writable`
* `permissions` (file) — implies `exists`
* `checksum` (file)
* `content` (file) (optional v1)

### Process/Service

* `running` (process)
* `listening` (service port)
* `healthy` (service)

### HTTP

* `reachable` (http endpoint)
* `status_code` (http endpoint)
* `tls` (http endpoint)

### Scheduling

* `scheduled` (cron)

### Backup (example domain)

* `backed_up` (file, database)

---

## 6. Handlers (Adapter System)

### 6.1 Handler Selection

* If `with <handler>` is specified: use that handler.
* Else: choose a default handler if exactly one matches.
* Else: compile error asking user to add `with`.

### 6.2 Handler Spec Syntax

```ens
with <handler_name> <arg_key> <arg_value> <arg_key> <arg_value> ...
```

Examples:

```ens
with AES:256 key "env:SECRET_KEY"
with posix mode "0600"
with http expected_status "200"
```

Args are strings or numbers (v1: strings only is fine).

### 6.3 Built-in Handlers (v1)

Filesystem:

* `fs.native` (exists, content, checksum)
* `posix` (permissions)

Encryption:

* `AES:256` (file encryption)

  * args: `key`, `mode` (optional), `salt` (optional)

HTTP:

* `http.get` (reachable, status_code)

Scheduler:

* `cron.native` (scheduled)

---

## 7. Implication Rules (Deterministic Inference)

### 7.1 Implication Expansion Pass

During compile:

1. Parse explicit ensures.
2. Resolve implicit subjects.
3. Expand implied prerequisites recursively.
4. De-duplicate identical guarantees.
5. Detect conflicts.

### 7.2 Examples

`ensure encrypted` implies:

* `ensure exists`
* `ensure readable`
* `ensure writable`

`ensure permissions` implies:

* `ensure exists`

### 7.3 Conflict Declaration

Conditions can declare mutual exclusivity.
Example (if defined): `encrypted` conflicts with `unencrypted`.

---

## 8. Execution Model

### 8.1 Phases

1. **Compile**

   * Lex/parse `.ens` to AST
   * Bind context
   * Expand implications
   * Build intent graph
   * Topological sort into an ordered list
2. **Run**

   * Sequentially evaluate guarantees in order
   * For first violated guarantee: attempt repair
   * Re-check; if satisfied, continue
   * After full pass with all satisfied: sleep interval

### 8.2 Sequential Guarantee Resolution

* Exactly one enforcement action at a time (v1).
* A guarantee must be satisfied before proceeding to later guarantees.

### 8.3 Drift Handling

* If any satisfied guarantee becomes unsatisfied later, the loop restarts at the earliest violated guarantee.

### 8.4 Observability

* Each run produces:

  * evaluation logs
  * enforcement logs
  * final satisfaction report

---

## 9. CLI Commands (v1)

* `ensura compile <file.ens>`

  * validates + prints resolved graph
* `ensura explain <file.ens>`

  * prints implied guarantees + chosen handlers
* `ensura plan <file.ens>`

  * prints the deterministic sequential plan
* `ensura run <file.ens>`

  * runs the continuous resolver
* `ensura check <file.ens>`

  * checks only (no repairs)

---

## 10. Tech Specs (Implementation)

### 10.1 Language Implementation

* **Go** (v1) for core engine + daemon
* Modules:

  * `parser` (lexer + AST)
  * `binder` (implicit subject resolution)
  * `imply` (implication expansion)
  * `graph` (dependency + ordering)
  * `planner` (topo sort + linear plan)
  * `runtime` (enforcement loop)
  * `adapters` (built-in handlers)

### 10.2 Parsing Approach

* Start with a small hand-written recursive descent parser or PEG.
* Output AST with source spans for good errors.

### 10.3 Data Model

* ResourceRef: `{type, name/alias, literal}`
* Guarantee: `{condition, subject, handler, params, guards, deps, ordering, violation_policy}`

### 10.4 Plugin System

Two options (v1 pick one):

* **In-process Go plugins** via registry + build tags (simple OSS dev)
* **Out-of-process adapters** via gRPC/stdin-json RPC (more secure)

Recommended v1: **in-process registry** (fast iteration).

### 10.5 State & Caching

* Cache check results per cycle.
* Detect no-op repairs.

### 10.6 Security Model

* Secrets referenced as `env:VAR` or `file:/path`.
* Never print secret values in logs.
* Add `--redact` default on.

### 10.7 Determinism Guarantees

* Stable sorting for topo ties.
* No random ordering.
* All inference rules are declared in condition metadata.

---

## 11. Error Rules (Strictness)

Compile-time errors:

* Ambiguous implicit subject
* Unknown keyword/condition/handler
* No handler matches a condition
* Multiple handlers match without `with`
* Cyclic dependencies in graph
* Declared conflicts (mutually exclusive conditions)

Runtime errors:

* Handler enforcement failure
* Unreachable resources
* Permission issues

---

## 12. Canonical Examples

### Example A — Simple file security

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

### Example B — Policy reuse

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

### Example C — Collection enforcement

```ens
invariant {
  for each file in directory "/secrets" {
    ensure encrypted with AES:256 key "env:SECRET_KEY"
  }
}
```

---

## 13. Roadmap

### v1 (MVP)

* Resources: file, directory, http
* Conditions: exists, encrypted, permissions, reachable/status
* Handlers: fs.native, AES:256, posix, http.get
* Blocks: `on <resource> { ... }`
* Implications + `ensura explain`
* Sequential runtime loop

### v1.5

* Scheduling resource + `scheduled`
* Per-ensure `on violation`
* Better guards + environment variables

### v2 - implement this too!

* Out-of-process adapters
* Parallel groups (`parallel {}`)
* Richer constraint conflicts
* Remote targets (ssh)

---

## 14. Naming & Branding

* Language name: **EnsuraScript**
* Tooling binary name: `ensura`
* Tagline: **Programming by guarantees, not instructions.**

