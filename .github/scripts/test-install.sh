#!/usr/bin/env bash
set -euo pipefail

repository_root=$(git rev-parse --show-toplevel)
installer="$repository_root/install.sh"
temporary=$(mktemp -d)
trap 'rm -rf "$temporary"' EXIT

release_dir="$temporary/release"
fixture_dir="$temporary/fixture"
fake_bin="$temporary/fake-bin"
home_dir="$temporary/home"
install_dir="$home_dir/.local/bin"
mkdir -p "$release_dir" "$fixture_dir" "$fake_bin" "$home_dir"

case $(uname -s) in
  Linux) os=linux ;;
  Darwin) os=darwin ;;
  *) printf 'unsupported test operating system\n' >&2; exit 1 ;;
esac
case $(uname -m) in
  x86_64 | amd64) arch=amd64 ;;
  arm64 | aarch64) arch=arm64 ;;
  *) printf 'unsupported test architecture\n' >&2; exit 1 ;;
esac
archive_name="papercuts_1.2.3_${os}_${arch}.tar.gz"

cat >"$fixture_dir/papercuts" <<'EOF'
#!/bin/sh
printf '%s\n' 'papercuts v1.2.3 fixture'
EOF
chmod 0755 "$fixture_dir/papercuts"
tar -czf "$release_dir/$archive_name" -C "$fixture_dir" papercuts
(
  cd "$release_dir"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$archive_name" >checksums.txt
  else
    shasum --algorithm 256 "$archive_name" >checksums.txt
  fi
)

cat >"$fake_bin/curl" <<EOF
#!/bin/sh
output=
url=
while [ "\$#" -gt 0 ]; do
  case "\$1" in
    --output) output=\$2; shift 2 ;;
    -*) shift ;;
    *) url=\$1; shift ;;
  esac
done
printf '%s\\n' "\$url" >>'$temporary/curl.log'
case "\$url" in
  */checksums.txt) source_file='$release_dir/checksums.txt' ;;
  */$archive_name) source_file='$release_dir/$archive_name' ;;
  *) printf 'unexpected URL: %s\\n' "\$url" >&2; exit 1 ;;
esac
cp "\$source_file" "\$output"
EOF
chmod 0755 "$fake_bin/curl"

output=$(
  PATH="$fake_bin:/usr/bin:/bin" \
    HOME="$home_dir" \
    PAPERCUTS_INSTALL_DIR="$install_dir" \
    sh "$installer"
)

installed_version=$("$install_dir/papercuts")
if [[ "$installed_version" != 'papercuts v1.2.3 fixture' ]]; then
  printf 'installed binary returned %q, want fixture version\n' "$installed_version" >&2
  exit 1
fi
if [[ "$output" != *"export PATH='$install_dir':\$PATH"* ]]; then
  printf 'installer did not print the PATH command; output was:\n%s\n' "$output" >&2
  exit 1
fi
if find "$home_dir" -maxdepth 1 -type f | grep -q .; then
  printf 'installer unexpectedly edited a file in the home directory\n' >&2
  exit 1
fi

before_invalid=$("$install_dir/papercuts")
if PATH="$fake_bin:/usr/bin:/bin" \
  HOME="$home_dir" \
  PAPERCUTS_INSTALL_DIR="$install_dir" \
  PAPERCUTS_VERSION='v1/../../malicious' \
  sh "$installer" >/dev/null 2>&1; then
  printf 'installer accepted a malformed pinned version\n' >&2
  exit 1
fi
after_invalid=$("$install_dir/papercuts")
if [[ "$after_invalid" != "$before_invalid" ]]; then
  printf 'malformed pinned version replaced the installed binary\n' >&2
  exit 1
fi

cat >"$install_dir/papercuts" <<'EOF'
#!/bin/sh
printf '%s\n' old
EOF
chmod 0755 "$install_dir/papercuts"
: >"$temporary/curl.log"
PATH="$fake_bin:/usr/bin:/bin" \
  HOME="$home_dir" \
  PAPERCUTS_INSTALL_DIR="$install_dir" \
  PAPERCUTS_VERSION=v1.2.3 \
  sh "$installer" >/dev/null
if ! grep -Fxq 'https://github.com/Whamp/papercuts/releases/download/v1.2.3/checksums.txt' "$temporary/curl.log"; then
  printf 'pinned install did not use the pinned release URL\n' >&2
  exit 1
fi
if [[ $("$install_dir/papercuts") != 'papercuts v1.2.3 fixture' ]]; then
  printf 'pinned install did not replace the old binary\n' >&2
  exit 1
fi

printf '%064d  %s\n' 0 "$archive_name" >"$release_dir/checksums.txt"
cat >"$install_dir/papercuts" <<'EOF'
#!/bin/sh
printf '%s\n' preserved
EOF
chmod 0755 "$install_dir/papercuts"
if PATH="$fake_bin:/usr/bin:/bin" \
  HOME="$home_dir" \
  PAPERCUTS_INSTALL_DIR="$install_dir" \
  sh "$installer" >/dev/null 2>&1; then
  printf 'installer accepted an invalid archive checksum\n' >&2
  exit 1
fi
if [[ $("$install_dir/papercuts") != preserved ]]; then
  printf 'checksum failure replaced the existing binary\n' >&2
  exit 1
fi
