#!/usr/bin/env bash

set -euo pipefail

output_dir=${1:-.cross-compile}
rm -rf "$output_dir"
mkdir -p "$output_dir"

packages_output=$(go list ./...)
mapfile -t packages <<<"$packages_output"

for target in \
  linux/amd64 linux/arm64 \
  darwin/amd64 darwin/arm64 \
  windows/amd64 windows/arm64; do
  goos=${target%/*}
  goarch=${target#*/}
  suffix=
  if [[ "$goos" == windows ]]; then
    suffix=.exe
  fi

  GOOS=$goos GOARCH=$goarch go build \
    -o "$output_dir/papercuts-${goos}-${goarch}${suffix}" \
    ./cmd/papercuts

  for package in "${packages[@]}"; do
    package_name=${package//\//_}
    GOOS=$goos GOARCH=$goarch go test -c \
      -o "$output_dir/${package_name}-${goos}-${goarch}.test${suffix}" \
      "$package"
  done
done
