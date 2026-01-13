---
name: bootstrapping-with-bun
description: Set up projects to be fully bootstrapped with minimal prerequisites using Bun as the runtime. This allows easy onboarding and consistent environments across team members.
---

# Bootstrapping with Bun

> **Source:** Adapted from [murrayju/bun-strap](https://github.com/murrayju/bun-strap) (MIT License)

## Overview

Use Bun in your project without installing it globally. A bootstrap script acts as a local version manager, ensuring consistent Bun versions across your team.

## When to Use

- Setting up a new project that uses Bun
- Need version consistency across team members
- Multiple projects require different Bun versions
- Want to avoid global Bun installation conflicts

## Quick Start

1. Copy [bun](bun) into your project root
2. Make executable: `chmod +x bun`
3. Run via `./bun <command>` (e.g., `./bun install`)

## Version Pinning

Edit the `version` variable in the script, or override via environment:

```bash
BUN_VERSION="bun-v1.3.6" ./bun --version
```

See [bun](bun) for the full script. Available versions at [oven-sh/bun/tags](https://github.com/oven-sh/bun/tags).

## Using in package.json Scripts

Reference `./bun` directly in npm scripts:

```json
{
  "scripts": {
    "dev": "./bun src/index.ts",
    "check": "./bun run lint && ./bun run typecheck && ./bun test"
  }
}
```

Run with `npm run dev` or `./bun run dev` - both work.

## Using in GitHub Actions

The bootstrap script works in CI without any Bun setup action. See [check.yml](check.yml) for a complete example.

The first `./bun` invocation downloads Bun automatically. Subsequent calls use the cached version.

## Shell Alias (Optional)

Add to `.bashrc` or `.zshrc` to prefer local bootstrap when available:

```bash
bun() {
  if [ -x ./bun ]; then
    ./bun "$@"
  else
    command bun "$@"
  fi
}
```

## Git Configuration

Add `download/` to `.gitignore`. The bootstrap script itself should be committed.
