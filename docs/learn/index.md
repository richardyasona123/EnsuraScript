# Learn EnsuraScript

Welcome to the EnsuraScript learning path! This tutorial series will teach you everything you need to know to write powerful declarative system automation.

## What You'll Learn

This tutorial is organized into progressive sections that build on each other:

### Getting Started

Start here if you're new to EnsuraScript.

- **[Installation](/learn/installation)** - Set up EnsuraScript on your system
- **[Your First Guarantee](/learn/first-guarantee)** - Write and run your first EnsuraScript program

### Core Concepts

Learn the fundamental building blocks of the language.

- **[Understanding Resources](/learn/resources)** - Files, HTTP endpoints, and other system entities
- **[Writing Guarantees](/learn/guarantees)** - The `ensure` statement and guarantee syntax
- **[Using Handlers](/learn/handlers)** - Built-in handlers for encryption, permissions, and more

### Advanced Features

Master the powerful features that make EnsuraScript unique.

- **[Creating Policies](/learn/policies)** - Reusable guarantee templates with parameters
- **[Guards & Conditions](/learn/guards)** - Conditional guarantees with `when` clauses
- **[Dependencies](/learn/dependencies)** - Ordering guarantees with `requires`, `after`, and `before`
- **[Collections & Invariants](/learn/collections)** - Enforce guarantees across multiple resources
- **[Violation Handling](/learn/violations)** - Retry logic and notifications when guarantees fail

### Deep Dives

Understand how EnsuraScript works under the hood.

- **[Implication System](/learn/implications)** - How guarantees automatically expand into prerequisites
- **[Execution Model](/learn/execution)** - Topological sorting and continuous enforcement

## Learning Path

We recommend following the tutorials in order, especially if you're new to declarative programming. Each tutorial builds on concepts from previous ones.

::: tip Not Just Another Scripting Language
EnsuraScript is fundamentally different from traditional automation tools. Instead of writing **procedures** (step-by-step instructions), you write **guarantees** (statements of truth). The runtime maintains those truths automatically through continuous monitoring and enforcement.
:::

## Quick Examples

### Secure a File

```ens
on file "secrets.db" {
  ensure exists
  ensure encrypted with AES:256 key "env:SECRET_KEY"
  ensure permissions with posix mode "0600"
}
```

### Monitor an API

```ens
on http "https://api.example.com/health" {
  ensure reachable
  ensure status_code with http.get expected_status "200"
  ensure tls
}
```

### Create Reusable Policies

```ens
policy secure_file(key_ref) {
  ensure encrypted with AES:256 key key_ref
  ensure permissions with posix mode "0600"
}

on file "database.db" {
  apply secure_file("env:DB_KEY")
}
```

## Ready to Start?

Begin with [Installation](/learn/installation) to set up EnsuraScript on your system, then move on to [Your First Guarantee](/learn/first-guarantee) to write your first program.

If you prefer to jump straight to specific topics, use the sidebar to navigate to any section.
