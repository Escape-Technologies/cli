#!/usr/bin/env bash
# Rewrite a CLI mirror branch so public GitHub history does not expose
# internal monorepo commit messages or author metadata.
set -euo pipefail

ANON_MESSAGE='feat: sync code from internal repository'
ANON_NAME='Escape Technologies'
ANON_EMAIL='bot@escape.tech'
ANON_BRANCH='cli-mirror-anonymized'

self_check() {
  local root script
  root=$(mktemp -d)
  script=$(cd "$(dirname "$0")" && pwd)/$(basename "$0")
  trap '[[ -n "${root:-}" ]] && rm -rf "$root"' EXIT

  git -C "$root" init -q
  git -C "$root" config user.email 'internal@escape.tech'
  git -C "$root" config user.name 'internal-bot'

  echo secret >"$root/file"
  git -C "$root" add file
  git -C "$root" commit -q -m 'fix(PLA-659): internal only (MR !24243)

Co-authored-by: internal <internal@escape.tech>'

  echo more >>"$root/file"
  git -C "$root" add file
  git -C "$root" commit -q -m 'feat(SAA-1358): also internal'

  local source_ref anonymized_ref
  source_ref=$(git -C "$root" rev-parse HEAD)
  anonymized_ref=$(
    cd "$root" && "$script" "$source_ref"
  )

  while IFS= read -r line; do
    [[ "$line" == "$ANON_MESSAGE" ]] || {
      echo "self-check: unexpected commit message: $line" >&2
      exit 1
    }
  done < <(
    git -C "$root" log --format=%s "$anonymized_ref"
  )

  while IFS= read -r line; do
    [[ "$line" == "${ANON_NAME} <${ANON_EMAIL}>" ]] || {
      echo "self-check: unexpected author: $line" >&2
      exit 1
    }
  done < <(
    git -C "$root" log --format='%an <%ae>' "$anonymized_ref"
  )

  echo 'anonymize-mirror-history self-check ok'
}

anonymize() {
  local source_ref="$1"

  git branch -D "$ANON_BRANCH" 2>/dev/null || true
  git branch "$ANON_BRANCH" "$source_ref"

  FILTER_BRANCH_SQUELCH_WARNING=1 git filter-branch -f \
    --msg-filter "printf '%s\n' '${ANON_MESSAGE}'" \
    --env-filter "
      export GIT_AUTHOR_NAME='${ANON_NAME}'
      export GIT_AUTHOR_EMAIL='${ANON_EMAIL}'
      export GIT_COMMITTER_NAME='${ANON_NAME}'
      export GIT_COMMITTER_EMAIL='${ANON_EMAIL}'
    " \
    "$ANON_BRANCH" >&2

  git rev-parse "$ANON_BRANCH"
}

case "${1:-}" in
  --self-check)
    self_check
    ;;
  --help | -h)
    echo "usage: $0 <commit-ref>" >&2
    echo "       $0 --self-check" >&2
    exit 0
    ;;
  '')
    echo "missing commit ref" >&2
    exit 1
    ;;
  *)
    anonymize "$1"
    ;;
esac
