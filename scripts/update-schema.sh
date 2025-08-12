#!/usr/bin/env -S bash -exuo pipefail

# Retrieve project root directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
VERSIONS=(
    "v3"
)
TMP_DIR="${PROJECT_ROOT}/._openapi-generator/"

docker pull openapitools/openapi-generator-cli:latest
docker pull python:3.12.9

( 
    cd "${PROJECT_ROOT}"
    for VERSION in "${VERSIONS[@]}"; do
        rm -rf "${TMP_DIR}/${VERSION}"
        mkdir -p "${TMP_DIR}/${VERSION}"
        curl -s -o "${TMP_DIR}/${VERSION}/openapi-raw.json" "https://public.escape.tech/${VERSION}/openapi.json"

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
