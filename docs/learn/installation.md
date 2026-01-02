# Installation

This guide will help you install EnsuraScript on your system.

## Prerequisites

EnsuraScript is written in Go and distributed as a single binary. You don't need Go installed to run it, but you will need:

- Linux, macOS, or WSL2 on Windows
- Basic command-line familiarity

## Installation Methods

### Option 1: Download Pre-built Binary (Recommended)

Download the latest release for your platform from the [GitHub Releases page](https://github.com/GustyCube/EnsuraScript/releases):

```bash
# macOS (ARM64)
curl -L https://github.com/GustyCube/EnsuraScript/releases/latest/download/ensura-darwin-arm64 -o ensura
chmod +x ensura
sudo mv ensura /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/GustyCube/EnsuraScript/releases/latest/download/ensura-darwin-amd64 -o ensura
chmod +x ensura
sudo mv ensura /usr/local/bin/

# Linux (x86_64)
curl -L https://github.com/GustyCube/EnsuraScript/releases/latest/download/ensura-linux-amd64 -o ensura
chmod +x ensura
sudo mv ensura /usr/local/bin/
```

### Option 2: Build from Source

If you have Go 1.21+ installed:

```bash
git clone https://github.com/GustyCube/EnsuraScript.git
cd EnsuraScript
go build -o ensura ./cmd/ensura
sudo mv ensura /usr/local/bin/
```

## Verify Installation

Check that EnsuraScript is installed correctly:

```bash
ensura --version
```

You should see output similar to:

```
EnsuraScript v1.0.0
```

## Next Steps

Now that you have EnsuraScript installed, continue to [Your First Guarantee](/learn/first-guarantee) to write and run your first program.

## Troubleshooting

### Command not found

If you get `command not found`, ensure `/usr/local/bin` is in your PATH:

```bash
echo $PATH | grep /usr/local/bin
```

If it's not there, add this to your `~/.bashrc` or `~/.zshrc`:

```bash
export PATH="/usr/local/bin:$PATH"
```

Then reload your shell:

```bash
source ~/.bashrc  # or ~/.zshrc
```

### Permission denied

If you get permission errors when running `ensura`, the binary may not be executable:

```bash
chmod +x /usr/local/bin/ensura
```
