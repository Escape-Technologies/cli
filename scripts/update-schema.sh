#!/usr/bin/env -S bash -exuo pipefail

# Retrieve project root directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
VERSIONS=(
    "v2"
)
TMP_DIR="${PROJECT_ROOT}/._openapi-generator/"

docker pull openapitools/openapi-generator-cli:latest

(
    cd "${PROJECT_ROOT}"
    for VERSION in "${VERSIONS[@]}"; do
        rm -rf "${TMP_DIR}/${VERSION}"
        mkdir -p "${TMP_DIR}/${VERSION}"
        curl -s -o "${TMP_DIR}/${VERSION}/openapi.json" "https://public.escape.tech/${VERSION}/openapi.json"

        docker run -u "$(id -u):$(id -g)" --rm -v "${TMP_DIR}/${VERSION}:/local" \
            openapitools/openapi-generator-cli:latest \
            generate \
            -i /local/openapi.json \
            -g go \
            --package-name "${VERSION}" \
            --type-mappings="integer=int" \
            -o /local

        rm -rf "${PROJECT_ROOT}/pkg/api/${VERSION}"
        mkdir -p "${PROJECT_ROOT}/pkg/api/${VERSION}"
        cp ${TMP_DIR}/${VERSION}/*.go "${PROJECT_ROOT}/pkg/api/${VERSION}/"
    done
)
exit $?
