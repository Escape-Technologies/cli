#!/bin/bash

# Retrieve project root directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

(
    cd "${PROJECT_ROOT}"

    # Retrieve the schema from the public API
    curl -s https://public.escape.tech/v1/openapi.json | jq > "${PROJECT_ROOT}/assets/public-api.json"

    # Generate the code binding on openAPI schema
    go generate ./...
)
