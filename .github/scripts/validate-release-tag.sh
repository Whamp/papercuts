#!/usr/bin/env bash
set -euo pipefail

tag=${1-}
if [[ ! $tag =~ ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-rc\.[1-9][0-9]*)?$ ]]; then
  printf 'invalid release tag: %s\n' "$tag" >&2
  exit 1
fi

object_type=$(git cat-file -t "refs/tags/$tag" 2>/dev/null || true)
if [[ $object_type != tag ]]; then
  printf 'release tag %s must be annotated\n' "$tag" >&2
  exit 1
fi
