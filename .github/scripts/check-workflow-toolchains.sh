#!/usr/bin/env bash

set -euo pipefail

workflow_files=(
  .github/workflows/ci.yml
  .github/workflows/release.yml
)
expected_patched_consumers=(1 4)
expected_baseline_consumers=(2 0)
versions=()

for index in "${!workflow_files[@]}"; do
  workflow_file=${workflow_files[$index]}
  mapfile -t matches < <(grep -E '^  PATCHED_GO_VERSION: "[0-9]+\.[0-9]+\.[0-9]+"$' "$workflow_file" || true)
  if [[ ${#matches[@]} -ne 1 ]]; then
    printf '%s must define PATCHED_GO_VERSION exactly once\n' "$workflow_file" >&2
    exit 1
  fi
  version=${matches[0]#*\"}
  versions+=("${version%\"}")

  setup_count=$(grep -Fc 'uses: actions/setup-go@' "$workflow_file" || true)
  # The GitHub expression is source text here, not a shell expansion.
  # shellcheck disable=SC2016
  patched_count=$(grep -Fc 'go-version: ${{ env.PATCHED_GO_VERSION }}' "$workflow_file" || true)
  baseline_count=$(grep -Fc 'go-version-file: go.mod' "$workflow_file" || true)
  if [[ $patched_count -ne ${expected_patched_consumers[$index]} ]] ||
    [[ $baseline_count -ne ${expected_baseline_consumers[$index]} ]] ||
    [[ $setup_count -ne $((patched_count + baseline_count)) ]]; then
    printf '%s has unexpected setup-go toolchain consumers\n' "$workflow_file" >&2
    exit 1
  fi
done

if [[ ${versions[0]} != "${versions[1]}" ]]; then
  printf 'patched Go versions differ: CI=%s release=%s\n' "${versions[0]}" "${versions[1]}" >&2
  exit 1
fi

if ! grep -Fq "GOTOOLCHAIN=go${versions[0]} go run golang.org/x/vuln/cmd/govulncheck@v1.6.0 ./..." README.md; then
  printf 'README vulnerability command does not use Go %s\n' "${versions[0]}" >&2
  exit 1
fi
