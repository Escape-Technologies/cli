# Escape CLI

[Escape](https://escape.tech) is a comprehensive API security platform with powerful API security testing capabilities.
This repository holds the code for the `escape-cli` tool to interact with the Escape API.

## GitHub Action

Run Escape directly in GitHub Actions. The action pulls the [`escapetech/cli`](https://hub.docker.com/r/escapetech/cli) image and runs it via `docker` on the runner (supported on GitHub-hosted `ubuntu-latest` and other environments where Docker is available). Pin a version tag of this repository (for example `v1`) in `uses:`.

We recommend providing these values as [encrypted secrets](https://docs.github.com/en/actions/security-guides/encrypted-secrets).

### Required inputs

- `profile_id`: The Escape profile to scan
- `api_key`: Your Escape API key

### Optional inputs

- `watch`: Wait for scan completion (`"true"` / `"false"`)
- `configuration_override`: JSON override of the scan configuration
- `schema`: Local path or public URL to update the application schema before the scan

### Example workflow

```yaml
name: Escape

on:
  push:
    branches:
      - main

jobs:
  escape:
    runs-on: ubuntu-latest
    steps:
      - name: Escape scan
        uses: Escape-Technologies/cli@v1
        with:
          profile_id: ${{ secrets.ESCAPE_PROFILE_ID }}
          api_key: ${{ secrets.ESCAPE_API_KEY }}
```
