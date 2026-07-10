#!/usr/bin/env bash
set -euo pipefail

repository_root=$(git rev-parse --show-toplevel)
validator="$repository_root/.github/scripts/validate-release-tag.sh"
temporary=$(mktemp -d)
trap 'rm -rf "$temporary"' EXIT

git -C "$temporary" init --quiet
git -C "$temporary" config user.name 'Papercuts Test'
git -C "$temporary" config user.email 'papercuts@example.invalid'
git -C "$temporary" commit --quiet --allow-empty -m initial

git -C "$temporary" tag v1.2.3
if (cd "$temporary" && "$validator" v1.2.3 >/dev/null 2>&1); then
  echo 'validator accepted a lightweight release tag' >&2
  exit 1
fi

git -C "$temporary" tag -a v1.2.4 -m v1.2.4
(cd "$temporary" && "$validator" v1.2.4)

if (cd "$temporary" && "$validator" v01.2.4 >/dev/null 2>&1); then
  echo 'validator accepted an invalid release tag' >&2
  exit 1
fi
