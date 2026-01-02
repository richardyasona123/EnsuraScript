---
layout: home

hero:
  name: EnsuraScript
  text: Programming by guarantees, not instructions
  tagline: An intent-first, truth maintenance language for declaring and enforcing system guarantees
  actions:
    - theme: brand
      text: Get Started
      link: /guide/getting-started
    - theme: alt
      text: View on GitHub
      link: https://github.com/GustyCube/EnsuraScript

features:
  - icon:
      dark: /icons/target.svg
      light: /icons/target.svg
    title: Intent Over Instruction
    details: Stop writing procedural scripts. Declare what should be true, and EnsuraScript maintains it automatically.
  - icon:
      dark: /icons/refresh-cw.svg
      light: /icons/refresh-cw.svg
    title: Self-Healing Systems
    details: Continuous monitoring detects drift and automatically repairs violations. Your guarantees stay enforced, always.
  - icon:
      dark: /icons/layers.svg
      light: /icons/layers.svg
    title: Composable Policies
    details: Build reusable policy libraries. Apply complex configurations with a single ensure statement.
  - icon:
      dark: /icons/search.svg
      light: /icons/search.svg
    title: Zero Black Boxes
    details: Every action is explainable and predictable. No AI hallucinations—just deterministic rule-based inference.
  - icon:
      dark: /icons/git-branch.svg
      light: /icons/git-branch.svg
    title: Smart Dependencies
    details: Automatic dependency resolution and topological execution. No more ordering headaches.
  - icon:
      dark: /icons/shield.svg
      light: /icons/shield.svg
    title: Security First
    details: Encrypted secrets, permission enforcement, and audit trails built into the language itself.
---

## The Problem

Infrastructure drifts. Configuration files change. Permissions get modified. Security policies break. Traditional automation runs once and forgets. You're left manually checking, fixing, and re-running scripts.

**There's a better way.**

## Declare Guarantees, Not Steps

```ens
# Traditional approach: imperative, fragile, runs once
# touch secrets.db
# chmod 0600 secrets.db
# encrypt secrets.db --key=$SECRET_KEY

# EnsuraScript: declarative, self-healing, continuous
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

**What happens?**

- File doesn't exist → Created automatically
- Wrong permissions → Fixed immediately
- Encryption removed → Re-encrypted
- **Continuous enforcement** → Violations detected and repaired in real-time

## Stop Babysitting Your Infrastructure

```bash
# Run once, enforce forever
ensura run config.ens

# See exactly what will happen before it does
ensura plan config.ens

# Validate your configuration
ensura compile config.ens

# Check without enforcing (dry run)
ensura check config.ens
```

::: tip Truth Maintenance, Not Task Automation
Traditional scripts tell the computer **how** to do something. EnsuraScript tells it **what must be true**. The runtime maintains those truths automatically, forever.
:::
