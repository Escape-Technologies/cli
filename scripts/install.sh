#!/bin/bash

set -euxo pipefail

# Get the latest version from the GitHub releases page
version="$(curl https://github.com/Escape-Technologies/cli/releases/latest -o /dev/null -v 2>&1 | grep '< location:' | awk -F'/v' '{ sub("\r$", ""); print $NF}')"

# infer os and platform
case "$(uname -s)" in
  "Darwin")
    os="darwin"
    ;;
  "Linux")
    os="linux"
    ;;
  *)
    echo "Unsupported OS: $(uname -s)"
    exit 1
    ;;
esac

case "$(uname -m)" in
  "x86_64")
    platform="amd64"
    ;;
  "arm64")
    platform="arm64"
    ;;
  *)
    echo "Unsupported platform: $(uname -m)"
    exit 1
    ;;
esac

tarball_name="cli_${version}_${os}_${platform}.tar.gz"

# create a temporary directory
tmpdir=$(mktemp -d)
mkdir -p "${tmpdir}"
cd "${tmpdir}"
pwd
trap 'rm -rf "${tmpdir}"' EXIT

# download the tarball
curl -f -s -L -o "${tarball_name}" "https://github.com/Escape-Technologies/cli/releases/download/v${version}/${tarball_name}"

# extract the tarball
tar -xzf "${tarball_name}"

# install the binary
install -m 755 escape-cli /usr/local/bin/escape-cli

echo "Done !"
