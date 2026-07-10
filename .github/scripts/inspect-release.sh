#!/usr/bin/env bash

set -euo pipefail

dist_dir=${1:-dist}
metadata_file="$dist_dir/metadata.json"
checksum_file="$dist_dir/checksums.txt"
repository_root=$(git rev-parse --show-toplevel)
extracted=$(mktemp -d)
trap 'rm -rf "$extracted"' EXIT

version=$(jq -er '.version | select(type == "string" and length > 0)' "$metadata_file")
expected_names=(
  "papercuts_${version}_darwin_amd64.tar.gz"
  "papercuts_${version}_darwin_arm64.tar.gz"
  "papercuts_${version}_linux_amd64.tar.gz"
  "papercuts_${version}_linux_arm64.tar.gz"
  "papercuts_${version}_windows_amd64.zip"
  "papercuts_${version}_windows_arm64.zip"
)

mapfile -t archives < <(find "$dist_dir" -maxdepth 1 -type f \( -name '*.tar.gz' -o -name '*.zip' \) | sort)
mapfile -t archive_names < <(printf '%s\n' "${archives[@]##*/}" | sort)
if [[ ${#archive_names[@]} -ne 6 ]] ||
  [[ "$(printf '%s\n' "${archive_names[@]}")" != "$(printf '%s\n' "${expected_names[@]}")" ]]; then
  printf 'archive names do not match the release contract:\n' >&2
  printf '%s\n' "${archive_names[@]}" >&2
  exit 1
fi

mapfile -t checksum_names < <(awk 'NF == 2 { print $2 }' "$checksum_file" | sort)
if [[ ${#checksum_names[@]} -ne 6 ]] ||
  [[ "$(printf '%s\n' "${checksum_names[@]}")" != "$(printf '%s\n' "${expected_names[@]}")" ]]; then
  printf 'checksum manifest does not cover exactly the release archives:\n' >&2
  printf '%s\n' "${checksum_names[@]}" >&2
  exit 1
fi

for archive in "${archives[@]}"; do
  case "$archive" in
    *.tar.gz)
      listing=$(tar -tzf "$archive" | sort)
      executable=papercuts
      tar -xOzf "$archive" LICENSE >"$extracted/LICENSE"
      tar -xOzf "$archive" README.md >"$extracted/README.md"
      ;;
    *.zip)
      listing=$(unzip -Z1 "$archive" | sort)
      executable=papercuts.exe
      unzip -p "$archive" LICENSE >"$extracted/LICENSE"
      unzip -p "$archive" README.md >"$extracted/README.md"
      ;;
  esac
  expected=$(printf '%s\n' LICENSE README.md "$executable" | sort)
  if [[ "$listing" != "$expected" ]]; then
    printf 'unexpected contents in %s:\n%s\n' "$archive" "$listing" >&2
    exit 1
  fi
  for source_file in LICENSE README.md; do
    if ! cmp -s "$repository_root/$source_file" "$extracted/$source_file"; then
      printf '%s in %s does not match the checked-out source file\n' "$source_file" "$archive" >&2
      exit 1
    fi
  done
done

(
  cd "$dist_dir"
  sha256sum --check --strict checksums.txt
)
