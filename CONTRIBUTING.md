# Contributing to Escape CLI

## Development workflow

The Escape CLI source of truth lives in the **Escape Technologies product monorepo** (private, GitLab). This GitHub repository is a **read-only mirror** — it exists for public auditability, install scripts, and GitHub Releases via goreleaser.

**Pull requests opened on this GitHub repository will not be reviewed or merged.**

If you found a bug or want to suggest a feature, please [open an issue](https://github.com/Escape-Technologies/cli/issues) and the team will handle it internally.

---

## For Escape team members

All CLI development happens inside the monorepo under `packages/cli/`. Changes are automatically mirrored to this GitHub repository on every merge to `prod`.

### Common commands (run from the monorepo root)

```bash
# Regenerate Go models after a public-api spec change
just cli-sync

# Run unit tests
just cli-test

# Lint
just cli-lint

# Build binary locally
just cli-build
```

### Releasing a new version

```bash
# 1. Bump version and commit (patch or minor)
just cli-release patch

# 2. Open MR, get review, merge to prod
#    → CI automatically mirrors the code to GitHub

# 3. Push the version tag to GitHub to trigger goreleaser
just cli-ci-mirror-tag v<new-version>
```

goreleaser will then publish the GitHub Release, Docker image, and binaries automatically.
