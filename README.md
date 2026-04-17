# Escape CLI

The official CLI for [Escape](https://escape.tech) — an AI-powered offensive security platform for APIs and web applications.

Two main use cases:

- **Security scanning** — trigger scans, manage vulnerabilities, and integrate Escape into CI/CD pipelines
- **Private Location** — run a self-hosted tunnel to reach APIs inside private networks

Full documentation: [docs.escape.tech/documentation/tooling/cli](https://docs.escape.tech/documentation/tooling/cli)

## Installation

**Binary** (Linux, macOS, Windows): download from [GitHub Releases](https://github.com/Escape-Technologies/cli/releases/latest).

**Docker:**

```bash
docker pull escapetech/cli:latest
```

**Helm** (for Private Location on Kubernetes):

```bash
helm install escape-location ./helm \
  --set ESCAPE_API_KEY=<your-api-key> \
  --set ESCAPE_PRIVATE_LOCATION=<location-name>
```

## Authentication

```bash
export ESCAPE_API_KEY=your_api_key_here
```

## Quick Start

```bash
# Start a scan and wait for results
escape-cli scans start <profile-id> --watch

# List emails received by a scan inbox
escape-cli emails list --email test.00000000-0000-0000-0000-000000000000@scan.escape.tech

# Review discovered issues
escape-cli issues list --severity CRITICAL,HIGH --status OPEN
```

## Private Location

Private Location is a self-hosted tunnel that lets Escape reach APIs that are not publicly accessible — behind firewalls, in private VPCs, or in Kubernetes clusters.

```bash
# Start a Private Location
escape-cli locations start <location-name>
```

For Kubernetes, use the provided Helm chart. See [`helm/values.yaml`](./helm/values.yaml) for configuration options including existing secrets and K8s integration.

## GitHub Actions

```yaml
- name: Escape scan
  uses: Escape-Technologies/cli@v1
  with:
    profile_id: ${{ secrets.ESCAPE_PROFILE_ID }}
    api_key: ${{ secrets.ESCAPE_API_KEY }}
    watch: 'true'
```

Optional inputs: `schema` (path or URL to update the API schema before the scan), `configuration_override` (partial JSON config).

## Release Checklist

Before cutting or validating a CLI release, make sure the GitHub Actions test secrets point to an org that can run the full e2e lifecycle.

- `E2E_API_KEY` must be allowed to create assets, create DAST WebApp profiles, start scans, cancel scans, delete profiles and assets, and create/delete tags.
- The target org must have DAST enabled. If the org has `DISABLE_DAST_SCANS`, the scan-start steps will fail.
- Re-run the GitHub workflows manually with `gh run rerun <run-id> --repo Escape-Technologies/cli`.
- If `scans start` fails with `400 Bad Request` and the `details` mention `FeatureUnavailableError`, `InvalidAuthenticationError`, or insufficient permissions, restore the missing org permission/feature flag instead of changing the CLI.

## Commands

```text
scans       Run and manage security scans
emails      List and read scan inbox emails
issues      Manage and track security vulnerabilities
problems    View scan problems and failures
profiles    Configure security testing profiles
assets      Manage your API inventory
locations   Manage Private Locations
custom-rules Manage custom security rules
tags        Organize resources with tags
upload      Upload schema files (OpenAPI, GraphQL, Postman)
audit       View audit logs
version     Display CLI version
```

Global flags: `-o pretty|json|yaml|schema`, `-v/-vv/-vvv`, and `--input-schema`.
