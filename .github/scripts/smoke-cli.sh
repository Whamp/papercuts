#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo 'usage: smoke-cli.sh <papercuts-binary>' >&2
  exit 2
fi
binary_directory=$(cd "$(dirname "$1")" && pwd)
binary="$binary_directory/$(basename "$1")"
if [[ ! -x $binary ]]; then
  printf 'papercuts binary is not executable: %s\n' "$binary" >&2
  exit 1
fi

root=$(mktemp -d)
trap 'rm -rf "$root"' EXIT

project="$root/project"
mkdir "$project"
cd "$project"
if "$binary" capture --severity low 'recovery smoke' >capture.out 2>capture.err; then
  echo 'capture unexpectedly succeeded before project initialization' >&2
  exit 1
fi
# The recovery command is literal output, not shell syntax.
# shellcheck disable=SC2016
grep -F 'run `papercuts init --project`' capture.err
"$binary" init --no-agents
"$binary" capture --severity low 'project smoke'
grep -F '> project smoke' PAPERCUTS.md

global_log="$root/global/PAPERCUTS.md"
if "$binary" capture --global --global-path "$global_log" --severity low 'global recovery smoke' >global.out 2>global.err; then
  echo 'capture unexpectedly succeeded before global initialization' >&2
  exit 1
fi
grep -F 'papercuts init --global --global-path' global.err
"$binary" init --global --global-path "$global_log" --no-agents
printf 'global line one\nglobal line two\n' | "$binary" capture --global --global-path "$global_log" --severity medium --stdin
grep -F '> global line one' "$global_log"
grep -F '> global line two' "$global_log"

agents_project="$root/agents"
mkdir "$agents_project"
cd "$agents_project"
"$binary" init --agents
grep -F '<!-- papercuts:begin -->' AGENTS.md
grep -F '<!-- papercuts:end -->' AGENTS.md

if [[ $(uname -s) == Linux ]]; then
  command -v script >/dev/null
  interactive_project="$root/interactive"
  mkdir "$interactive_project"
  printf 'y\n' | script -qec "cd '$interactive_project' && '$binary' init" /dev/null >/dev/null
  grep -F '<!-- papercuts:begin -->' "$interactive_project/AGENTS.md"
fi
