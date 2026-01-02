# What is EnsuraScript?

EnsuraScript is an open-source, intent-first, "truth maintenance" language. Instead of writing procedural code that describes *how* to do things, you declare *what* properties you want your systems to have. The EnsuraScript runtime then satisfies those guarantees and keeps them true over time.

## Design Philosophy

### Intent Over Instruction

Traditional configuration management and scripting require you to specify exact steps:

```bash
# Traditional approach
if [ ! -f secrets.db ]; then
  touch secrets.db
fi
chmod 600 secrets.db
# Hope encryption is handled elsewhere...
```

With EnsuraScript, you declare the desired state:

```ens
on file "secrets.db" {
  ensure exists
  ensure encrypted with AES:256 key "env:SECRET_KEY"
  ensure permissions with posix mode "0600"
}
```

The runtime handles the "how" - creating the file if needed, encrypting it, setting permissions, and continuously monitoring for drift.

### Deterministic Inference

EnsuraScript uses rule-based inference, not AI or machine learning. When you write `ensure encrypted`, the system knows this implies `ensure exists` (a file must exist to be encrypted). You can always run `ensura explain` to see exactly what guarantees will be enforced and in what order.

### Continuous Enforcement

EnsuraScript doesn't just run once. It continuously monitors your system for drift and automatically repairs violations. If someone changes file permissions or an HTTP endpoint goes down, EnsuraScript detects and responds.

## Key Features

- **Resources**: Files, directories, HTTP endpoints, services, databases
- **Guarantees**: Declarative constraints that must be satisfied
- **Handlers**: Pluggable implementations for checking and enforcing conditions
- **Policies**: Reusable bundles of guarantees
- **Implications**: Automatic prerequisite expansion
- **Guards**: Conditional activation based on environment
- **Invariants**: High-priority, critical guarantees

## Non-Goals (v1)

EnsuraScript is not trying to be:

- A general-purpose programming language
- A replacement for Kubernetes or Terraform
- An AI-powered system that "figures things out"

It's a focused tool for declaring and maintaining system invariants.
