#!/bin/sh
set -eu

repository_url=https://github.com/Whamp/papercuts
install_dir=${PAPERCUTS_INSTALL_DIR:-"$HOME/.local/bin"}
version=${PAPERCUTS_VERSION:-latest}

case $(uname -s) in
  Linux) os=linux ;;
  Darwin) os=darwin ;;
  *)
    printf 'papercuts: unsupported operating system: %s\n' "$(uname -s)" >&2
    exit 1
    ;;
esac

case $(uname -m) in
  x86_64 | amd64) arch=amd64 ;;
  arm64 | aarch64) arch=arm64 ;;
  *)
    printf 'papercuts: unsupported architecture: %s\n' "$(uname -m)" >&2
    exit 1
    ;;
esac

if [ "$version" = latest ]; then
  release_url="$repository_url/releases/latest/download"
elif printf '%s\n' "$version" | grep -Eq '^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-rc\.[1-9][0-9]*)?$'; then
  release_url="$repository_url/releases/download/$version"
else
  printf 'papercuts: PAPERCUTS_VERSION must be latest or a valid release tag\n' >&2
  exit 1
fi

temporary=$(mktemp -d)
trap 'rm -rf "$temporary"' EXIT HUP INT TERM
checksums="$temporary/checksums.txt"
curl --fail --silent --show-error --location --output "$checksums" "$release_url/checksums.txt"

archive_name=$(awk -v suffix="_${os}_${arch}.tar.gz" '$2 ~ suffix "$" { print $2 }' "$checksums")
case $archive_name in
  papercuts_*_"$os"_"$arch".tar.gz) ;;
  *)
    printf 'papercuts: release does not contain exactly one archive for %s/%s\n' "$os" "$arch" >&2
    exit 1
    ;;
esac

archive="$temporary/$archive_name"
curl --fail --silent --show-error --location --output "$archive" "$release_url/$archive_name"
expected_hash=$(awk -v name="$archive_name" '$2 == name { print $1 }' "$checksums")
if command -v sha256sum >/dev/null 2>&1; then
  actual_hash=$(sha256sum "$archive" | awk '{ print $1 }')
elif command -v shasum >/dev/null 2>&1; then
  actual_hash=$(shasum --algorithm 256 "$archive" | awk '{ print $1 }')
else
  printf 'papercuts: sha256sum or shasum is required to verify the release\n' >&2
  exit 1
fi
if [ "$actual_hash" != "$expected_hash" ]; then
  printf 'papercuts: checksum verification failed for %s\n' "$archive_name" >&2
  exit 1
fi

tar -xzf "$archive" -C "$temporary" papercuts
mkdir -p "$install_dir"
staged="$install_dir/.papercuts-install.$$"
install -m 0755 "$temporary/papercuts" "$staged"
mv -f "$staged" "$install_dir/papercuts"

printf 'Installed papercuts to %s\n' "$install_dir/papercuts"
case ":$PATH:" in
  *":$install_dir:"*) ;;
  *)
    quoted_install_dir=$(printf '%s' "$install_dir" | sed "s/'/'\\\\''/g")
    printf "Add papercuts to PATH for this shell:\n  export PATH='%s':\$PATH\n" "$quoted_install_dir"
    ;;
esac
