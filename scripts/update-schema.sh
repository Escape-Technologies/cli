#!/usr/bin/env -S bash -exuo pipefail

# Retrieve project root directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
VERSIONS=(
    "v3"
)
TMP_DIR="${PROJECT_ROOT}/._openapi-generator/"

# Local spec paths to check (relative to PROJECT_ROOT), in order of preference.
# Falls back to fetching from public URL if none found.
LOCAL_SPEC_CANDIDATES=(
    "../product/services/public-api/v3.openapi.json"
    "../../product/services/public-api/v3.openapi.json"
)

docker pull openapitools/openapi-generator-cli:latest
docker pull python:3.12.9

(
    cd "${PROJECT_ROOT}"
    for VERSION in "${VERSIONS[@]}"; do
        rm -rf "${TMP_DIR}/${VERSION}"
        mkdir -p "${TMP_DIR}/${VERSION}"

        # Determine spec source: local file or remote URL
        LOCAL_SPEC=""
        for candidate in "${LOCAL_SPEC_CANDIDATES[@]}"; do
            resolved="${PROJECT_ROOT}/${candidate}"
            if [[ -f "${resolved}" ]]; then
                LOCAL_SPEC="${resolved}"
                break
            fi
        done

        if [[ -n "${LOCAL_SPEC}" ]]; then
            echo "Using local spec: ${LOCAL_SPEC}"
            cp "${LOCAL_SPEC}" "${TMP_DIR}/${VERSION}/openapi-raw.json"
        else
            echo "Fetching spec from https://public.escape.tech/${VERSION}/openapi.json"
            curl -s -o "${TMP_DIR}/${VERSION}/openapi-raw.json" "https://public.escape.tech/${VERSION}/openapi.json"
        fi

        docker run -u "$(id -u):$(id -g)" --rm -v "${TMP_DIR}/${VERSION}:/local" -v "${PROJECT_ROOT}/scripts:/scripts" \
            python:3.12.9 \
            python /scripts/convert.py /local/openapi-raw.json /local/openapi.json

        docker run -u "$(id -u):$(id -g)" --rm -v "${TMP_DIR}/${VERSION}:/local" -v "${PROJECT_ROOT}/scripts:/scripts" \
            openapitools/openapi-generator-cli:latest \
            generate \
            -i /local/openapi.json \
            -c /scripts/openapi-generator.yaml \
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
