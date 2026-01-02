# Your First Guarantee

Let's write your first EnsuraScript program to understand the fundamental concepts of the language.

## The Traditional Approach

In traditional scripting, you might write:

```bash
touch config.yaml
chmod 0644 config.yaml
```

This approach has problems:
- Runs once and forgets
- Doesn't detect drift
- Doesn't fix violations
- No way to know if it's still true

## The EnsuraScript Approach

With EnsuraScript, you declare what **must be true**:

```ens
on file "config.yaml" {
  ensure exists
  ensure permissions with posix mode "0644"
}
```

This is a **guarantee**. EnsuraScript will:
1. Check if it's true
2. Fix it if it's not
3. Keep checking forever
4. Re-fix if it breaks

## Writing Your First Program

Create a file called `first.ens`:

```ens
on file "config.yaml" {
  ensure exists
  ensure permissions with posix mode "0644"
}
```

### Understanding the Syntax

Let's break down each part:

**`on file "config.yaml"`** - This is a **resource declaration** with an **on block**. It says "for the file at path config.yaml, apply the following guarantees."

**`ensure exists`** - This is a **guarantee**. It states "this file must exist."

**`ensure permissions with posix mode "0644"`** - Another guarantee. It uses the `posix` **handler** with the argument `mode` set to `"0644"`.

## Planning Your Guarantees

Before running, let's see what EnsuraScript will do:

```bash
ensura plan first.ens
```

Output:

```
Execution Plan (2 steps):

1. [fs.native] ensure exists on file "config.yaml"
2. [posix] ensure permissions with posix mode "0644" on file "config.yaml"
```

This shows:
- The order of execution
- Which handler will be used for each guarantee
- The resource each guarantee applies to

## Running Your Program

Execute with continuous enforcement:

```bash
ensura run first.ens
```

Output:

```
[✓] ensure exists on file "config.yaml" - REPAIRED
[✓] ensure permissions with posix mode "0644" on file "config.yaml" - REPAIRED

All guarantees satisfied. Monitoring for drift...
```

The runtime:
1. Created the file (it didn't exist)
2. Set permissions to 0644
3. Now continuously monitors and will re-fix if changed

## Testing Drift Detection

While `ensura run` is running, try modifying the file permissions in another terminal:

```bash
chmod 0777 config.yaml
```

Watch the ensura output - within 30 seconds (the default check interval), you'll see:

```
[✓] ensure permissions with posix mode "0644" on file "config.yaml" - REPAIRED
```

EnsuraScript detected the drift and fixed it automatically!

## Dry Run (Check Mode)

To check guarantees without enforcing them:

```bash
ensura check first.ens
```

This runs once and reports violations without fixing them. Useful for validation.

## Understanding Implications

Here's something interesting - modify your program to:

```ens
on file "config.yaml" {
  ensure permissions with posix mode "0644"
}
```

Notice we removed `ensure exists`. Now run:

```bash
ensura explain first.ens
```

Output:

```
Guarantees (2 total, 1 implied):

1. [IMPLIED] [fs.native] ensure exists on file "config.yaml"
2. [posix] ensure permissions with posix mode "0644" on file "config.yaml"
```

EnsuraScript automatically added `ensure exists` because **permissions implies exists**. You can't set permissions on a file that doesn't exist, so the language infers the prerequisite for you.

This is called the **implication system** - we'll cover it in depth later in [Implication System](/learn/implications).

## What You Learned

In this tutorial, you:

- Wrote your first guarantee using `ensure`
- Understood resource declarations with `on`
- Used handlers with arguments (`posix mode "0644"`)
- Ran programs with `ensura run`, `ensura plan`, `ensura check`, and `ensura explain`
- Observed continuous drift detection
- Discovered implication expansion

## Next Steps

Continue to [Understanding Resources](/learn/resources) to learn about all the types of resources EnsuraScript can manage.

## Full Example

Here's a more complete example you can try:

```ens
# Application configuration file
on file "config.yaml" {
  ensure exists
  ensure permissions with posix mode "0644"
  ensure content with fs.native content "app_name: MyApp"
}

# Secrets file with encryption
on file "secrets.env" {
  ensure exists
  ensure encrypted with AES:256 key "env:SECRET_KEY"
  ensure permissions with posix mode "0600"
}
```

Before running, set a secret key:

```bash
export SECRET_KEY="my-encryption-key"
ensura run example.ens
```

Try this on your own to see encryption in action!
