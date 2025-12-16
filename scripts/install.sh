#!/bin/bash

set -euo pipefail

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
  "arm64"|"aarch64")
    platform="arm64"
    ;;
  *)
    echo "Unsupported platform: $(uname -m)"
    exit 1
    ;;
esac

echo "Installing escape-cli v${version} for ${os} ${platform}..."

tarball_name="cli_${version}_${os}_${platform}.tar.gz"

# create a temporary directory
tmpdir=$(mktemp -d)
mkdir -p "${tmpdir}"
cd "${tmpdir}"
trap 'rm -rf "${tmpdir}"' EXIT

# download the tarball
curl -f -s -L -o "${tarball_name}" "https://github.com/Escape-Technologies/cli/releases/download/v${version}/${tarball_name}"

# extract the tarball
tar -xzf "${tarball_name}"

# install the binary
install -m 755 escape-cli /usr/local/bin/escape-cli

cat <<EOF
Done ! You can now use the escape-cli command.

If /usr/local/bin/ is not in your PATH, you can add this in your ~/.zshrc or ~/.bashrc file:

export PATH="/usr/local/bin:\$PATH"

Then, go to https://app.escape.tech/user/profile/ copy your API key and add the following to your ~/.zshrc or ~/.bashrc file:

export ESCAPE_API_KEY="your-api-key"

Then, you should be able to run:

escape-cli profiles list
EOF
