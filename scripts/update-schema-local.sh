#!/usr/bin/env -S bash -exuo pipefail

# Non-Docker version of update-schema.sh.
# Uses python3 and openapi-generator-cli (pip-installed) instead of Docker.
# Used in CI (toolkit image has openapi-generator-cli via pip) and local monorepo dev.

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
VERSIONS=(
    "v3"
)
TMP_DIR="${PROJECT_ROOT}/._openapi-generator/"

# Find the monorepo spec: walk up from packages/cli until services/public-api/v3.openapi.json is found
MONOREPO_ROOT="${PROJECT_ROOT}"
while [[ ! -f "${MONOREPO_ROOT}/services/public-api/v3.openapi.json" ]] && [[ "${MONOREPO_ROOT}" != "/" ]]; do
    MONOREPO_ROOT="$(dirname "${MONOREPO_ROOT}")"
done

if [[ ! -f "${MONOREPO_ROOT}/services/public-api/v3.openapi.json" ]]; then
    echo "ERROR: could not find services/public-api/v3.openapi.json in any parent directory"
    exit 1
fi

LOCAL_SPEC="${MONOREPO_ROOT}/services/public-api/v3.openapi.json"
echo "Using local spec: ${LOCAL_SPEC}"

(
    cd "${PROJECT_ROOT}"
    for VERSION in "${VERSIONS[@]}"; do
        rm -rf "${TMP_DIR}/${VERSION}"
        mkdir -p "${TMP_DIR}/${VERSION}"

        # Transform spec with Python (no Docker)
        python3 "${PROJECT_ROOT}/scripts/convert.py" "${LOCAL_SPEC}" "${TMP_DIR}/${VERSION}/openapi.json"

        # Generate Go client with openapi-generator-cli (pip-installed, includes bundled JDK)
        openapi-generator-cli generate \
            -i "${TMP_DIR}/${VERSION}/openapi.json" \
            -c "${PROJECT_ROOT}/scripts/openapi-generator.yaml" \
            -g go \
            --package-name "${VERSION}" \
            --type-mappings="integer=int" \
            -o "${TMP_DIR}/${VERSION}"

        rm -rf "${PROJECT_ROOT}/pkg/api/${VERSION}"
        mkdir -p "${PROJECT_ROOT}/pkg/api/${VERSION}"
        cp "${TMP_DIR}/${VERSION}"/*.go "${PROJECT_ROOT}/pkg/api/${VERSION}/"
    done
)
