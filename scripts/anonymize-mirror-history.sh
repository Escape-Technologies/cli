#!/usr/bin/env bash
# Build one anonymized commit from a packages/cli tree for the public GitHub mirror.
#
# Public history must not expose monorepo messages or authors. We used to
# `git subtree split` the whole repo and `git filter-branch` the result.
# That walks ~80k monorepo commits, exceeds the 25m CI timeout, and prints
# nothing because the split ran with `-q`. One `commit-tree` on the current
# tree is enough: messages stay dummy, GitHub keeps a linear sync history.
set -euo pipefail

ANON_MESSAGE='feat: sync code from internal repository'
ANON_NAME='Escape Technologies'
ANON_EMAIL='bot@escape.tech'

log() {
  printf '[cli-mirror %s] %s\n' "$(date -u +%H:%M:%S)" "$*" >&2
}

die() {
  log "error: $*"
  exit 1
}

commit_tree() {
  local tree="$1"
  shift
  GIT_AUTHOR_NAME="$ANON_NAME" \
    GIT_AUTHOR_EMAIL="$ANON_EMAIL" \
    GIT_COMMITTER_NAME="$ANON_NAME" \
    GIT_COMMITTER_EMAIL="$ANON_EMAIL" \
    git commit-tree "$tree" "$@" -m "$ANON_MESSAGE"
}

# Print the anonymized commit SHA for tree_ish, parented on onto_ref when given.
anonymize() {
  local tree_ish="$1"
  local onto_ref="${2:-}"

  log "resolving tree from ${tree_ish}"
  local object tree
  # Peel after resolving. `HEAD:path^{tree}` is parsed as a path, not a peel.
  object=$(git rev-parse --verify "${tree_ish}") || die "cannot resolve ${tree_ish}"
  tree=$(git rev-parse "${object}^{tree}") || die "cannot peel tree from ${object}"
  log "object=${object} tree=${tree}"

  if [[ -n "$onto_ref" ]]; then
    log "resolving onto ${onto_ref}"
    local onto_commit onto_tree
    onto_commit=$(git rev-parse "$onto_ref")
    onto_tree=$(git rev-parse "${onto_commit}^{tree}")
    log "onto=${onto_commit} onto_tree=${onto_tree}"
    if [[ "$tree" == "$onto_tree" ]]; then
      log "tree matches onto, reusing ${onto_commit}"
      printf '%s\n' "$onto_commit"
      return
    fi
    log "creating anonymized commit with parent ${onto_commit}"
    commit_tree "$tree" -p "$onto_commit"
    return
  fi

  log "creating orphan anonymized commit"
  commit_tree "$tree"
}

assert_anon_commit() {
  local repo="$1"
  local commit="$2"
  local expected_tree="$3"
  local expected_parent="${4:-}"
  local subject author parent

  subject=$(git -C "$repo" log -1 --format=%s "$commit")
  author=$(git -C "$repo" log -1 --format='%an <%ae>' "$commit")
  parent=$(git -C "$repo" log -1 --format=%P "$commit")

  [[ "$subject" == "$ANON_MESSAGE" ]] || die "unexpected commit message: ${subject}"
  [[ "$author" == "${ANON_NAME} <${ANON_EMAIL}>" ]] || die "unexpected author: ${author}"
  [[ "$(git -C "$repo" rev-parse "${commit}^{tree}")" == "$expected_tree" ]] || die "unexpected tree on ${commit}"
  [[ "$parent" == "$expected_parent" ]] || die "unexpected parent on ${commit}: ${parent}"
}

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

  local first_tree first_anon reused
  first_tree=$(git -C "$root" rev-parse 'HEAD^{tree}')
  first_anon=$(cd "$root" && "$script" HEAD)
  assert_anon_commit "$root" "$first_anon" "$first_tree"

  reused=$(cd "$root" && "$script" HEAD "$first_anon")
  [[ "$reused" == "$first_anon" ]] || die "same tree should reuse ${first_anon}, got ${reused}"

  mkdir -p "$root/packages/cli"
  echo nested >"$root/packages/cli/file"
  git -C "$root" add packages/cli/file
  git -C "$root" commit -q -m 'feat: add nested tree'

  local nested_tree nested_anon
  nested_tree=$(git -C "$root" rev-parse 'HEAD:packages/cli')
  nested_anon=$(cd "$root" && "$script" 'HEAD:packages/cli')
  assert_anon_commit "$root" "$nested_anon" "$nested_tree"
  reused=$(cd "$root" && "$script" 'HEAD:packages/cli' "$nested_anon")
  [[ "$reused" == "$nested_anon" ]] || die "nested tree should reuse ${nested_anon}, got ${reused}"

  echo extra >>"$root/file"
  git -C "$root" add file
  git -C "$root" commit -q -m 'fix(PLA-1): more internal'

  local second_tree second_anon
  second_tree=$(git -C "$root" rev-parse 'HEAD^{tree}')
  second_anon=$(cd "$root" && "$script" HEAD "$first_anon")
  assert_anon_commit "$root" "$second_anon" "$second_tree" "$first_anon"
  [[ "$second_anon" != "$first_anon" ]] || die "changed tree should create a new commit"

  echo 'anonymize-mirror-history self-check ok'
}

case "${1:-}" in
  --self-check)
    self_check
    ;;
  --help | -h)
    echo "usage: $0 <tree-ish> [onto-ref]" >&2
    echo "       $0 --self-check" >&2
    exit 0
    ;;
  '')
    echo "missing tree-ish" >&2
    exit 1
    ;;
  *)
    anonymize "$1" "${2:-}"
    ;;
esac
