# Getting Started

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/ensurascript/ensura.git
cd ensura

# Build
go build -o ensura ./cmd/ensura

# Install (optional)
sudo mv ensura /usr/local/bin/
```

### Verify Installation

```bash
ensura version
# ensura version 1.0.0
```

## Your First EnsuraScript

Create a file called `hello.ens`:

```ens
# Ensure a configuration file exists with specific permissions
resource file "config.yaml"

on file "config.yaml" {
  ensure exists
  ensure permissions with posix mode "0644"
}
```

### Compile and Validate

Check your script for errors:

```bash
ensura compile hello.ens
# Compilation successful!
#   Guarantees: 2
#   Dependencies: 1
#   Plan steps: 2
```

### See What Will Happen

Use `explain` to understand what EnsuraScript will do:

```bash
ensura explain hello.ens
# Guarantee Explanations
# ======================
#
# 1. Ensure exists on file "config.yaml"
#    Handler: fs.native
#
# 2. Ensure permissions on file "config.yaml"
#    Handler: posix
#    Arguments:
#      mode: 0644
```

### View the Execution Plan

See the ordered steps:

```bash
ensura plan hello.ens
# Execution Plan
# ==============
#
#   1. Ensure exists on file "config.yaml"
#       Handler: fs.native
#   2. Ensure permissions on file "config.yaml"
#       Handler: posix
#       Args:
#         mode: 0644
```

### Check Current State

Run a dry-run check without making changes:

```bash
ensura check hello.ens
```

### Run Continuous Enforcement

Start the enforcement loop:

```bash
ensura run hello.ens
# Starting enforcement loop (interval: 30s, retries: 3)
# Press Ctrl+C to stop
#
# [2024-01-15T10:30:00Z] Enforcement run completed in 45ms
#   Checks: 2, Repairs: 1, Failures: 0
#   Status: ALL SATISFIED
```

## Next Steps

- Learn about [Core Concepts](/guide/core-concepts)
- Explore [Resources](/guide/resources) and [Guarantees](/guide/guarantees)
- See more [Examples](/examples/)
