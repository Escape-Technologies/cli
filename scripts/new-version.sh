#!/bin/bash

# Retrieve project root directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

old_version=$(cat $PROJECT_ROOT/version.txt)
SEMVER_REGEX="^([0-9]+)\.([0-9]+)\.([0-9]+)$"

if [[ "$old_version" =~ $SEMVER_REGEX ]]; then
    _major=${BASH_REMATCH[1]}
    _minor=${BASH_REMATCH[2]}
    _patch=${BASH_REMATCH[3]}
else
    echo "Invalid version in version.txt: $old_version"
    exit 1
fi

case $1 in
  patch)
    _patch=$(($_patch + 1))
    ;;
  minor)
    _minor=$(($_minor + 1))
    _patch=0
    ;;
  *)
    echo "Usage: $0 <minor|patch>"
    exit 1
    ;;
esac

new_version="${_major}.${_minor}.${_patch}"
echo "Bump done: $old_version -> $new_version"

echo "${new_version}" > $PROJECT_ROOT/version.txt

git add $PROJECT_ROOT/version.txt
git commit -m "v${new_version}"
git tag -a "v${new_version}" -m "v${new_version}"
git push
git push --tags

GOPROXY=proxy.golang.org go list -m "github.com/Escape-Technologies/cli@v${new_version}"

echo "Done !"
