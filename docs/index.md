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
      link: https://github.com/ensurascript/ensura

features:
  - icon: 🎯
    title: Intent Over Instruction
    details: Declare desired outcomes with 'ensure' statements. Let the runtime figure out how to achieve and maintain them.
  - icon: 🔄
    title: Continuous Enforcement
    details: Guarantees are continuously checked and automatically repaired when drift is detected.
  - icon: 🧩
    title: Composable Policies
    details: Bundle reusable guarantees into policies. Apply them consistently across your infrastructure.
  - icon: 🔍
    title: Deterministic & Inspectable
    details: No AI guessing. All inference is rule-based. Use 'ensura explain' to see exactly what will happen.
  - icon: 📊
    title: Dependency-Aware
    details: Automatic implication expansion and topological ordering ensures guarantees are satisfied in the right order.
  - icon: 🛡️
    title: Built-in Security
    details: File encryption, permission management, and secure secret handling out of the box.
---

## Quick Example

```ens
# Ensure a secrets file is encrypted and has restricted permissions
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

Run with:
```bash
ensura run config.ens
```
