# Papercuts

`papercuts` gives coding agents a place to record workflow friction they would otherwise silently push through. Agents use the capture command while working; humans install the CLI and review its Markdown logs.

Papercuts is not a manual journaling tool. It is also not a product bug tracker or feature-request system. Send product bugs and feature requests to the product's issue tracker.

## Quick start

After installing `papercuts`, run this once from the project root:

```sh
papercuts init --project --agents
```

This command creates `PAPERCUTS.md` and adds capture instructions to `AGENTS.md`. An agent that follows `AGENTS.md` can then record friction as it works:

```sh
papercuts capture --severity low \
  "The formatter required an undocumented working directory."
```

Review the accumulated observations in `PAPERCUTS.md`. Each entry includes a UTC timestamp, severity, and description.

## Install

Tagged releases publish standalone archives on [GitHub Releases](https://github.com/Whamp/papercuts/releases) for Linux, macOS, and Windows on `amd64` and `arm64`. Each release includes six archives and `checksums.txt`. Each archive contains `README.md`, `LICENSE`, and the platform executable.

### Linux

This example installs `v0.1.0` for `amd64` into `~/.local/bin`:

```sh
version=0.1.0
archive="papercuts_${version}_linux_amd64.tar.gz"
base="https://github.com/Whamp/papercuts/releases/download/v${version}"
curl --fail --location --remote-name "$base/$archive"
curl --fail --location --remote-name "$base/checksums.txt"
grep "  $archive\$" checksums.txt | sha256sum --check -
tar -xzf "$archive"
mkdir -p "$HOME/.local/bin"
install -m 0755 papercuts "$HOME/.local/bin/papercuts"
"$HOME/.local/bin/papercuts" version
```

Replace `amd64` with `arm64` on ARM Linux.

### macOS

This example installs `v0.1.0` for Apple silicon into `~/.local/bin`:

```sh
version=0.1.0
archive="papercuts_${version}_darwin_arm64.tar.gz"
base="https://github.com/Whamp/papercuts/releases/download/v${version}"
curl --fail --location --remote-name "$base/$archive"
curl --fail --location --remote-name "$base/checksums.txt"
grep "  $archive\$" checksums.txt | shasum --algorithm 256 --check
tar -xzf "$archive"
mkdir -p "$HOME/.local/bin"
install -m 0755 papercuts "$HOME/.local/bin/papercuts"
"$HOME/.local/bin/papercuts" version
```

Replace `arm64` with `amd64` on an Intel Mac.

### Windows

This PowerShell example installs `v0.1.0` for `amd64` into `$HOME\bin`:

```powershell
$version = "0.1.0"
$archive = "papercuts_${version}_windows_amd64.zip"
$base = "https://github.com/Whamp/papercuts/releases/download/v${version}"
Invoke-WebRequest "$base/$archive" -OutFile $archive
Invoke-WebRequest "$base/checksums.txt" -OutFile checksums.txt
$expected = ((Select-String -Path checksums.txt -Pattern ([regex]::Escape($archive))).Line -split '\s+')[0]
$actual = (Get-FileHash -Algorithm SHA256 $archive).Hash.ToLowerInvariant()
if ($actual -ne $expected.ToLowerInvariant()) { throw "checksum mismatch" }
Expand-Archive -Path $archive -DestinationPath papercuts-release
$bin = Join-Path $HOME "bin"
New-Item -ItemType Directory -Force -Path $bin | Out-Null
Move-Item -Force papercuts-release\papercuts.exe $bin\papercuts.exe
& $bin\papercuts.exe version
```

Replace `amd64` with `arm64` on Windows ARM. Add `$HOME\bin` to the user `PATH` when needed.

GitHub CLI users can also verify release provenance:

```sh
gh attestation verify papercuts_0.1.0_linux_amd64.tar.gz --repo Whamp/papercuts
```

Initial macOS and Windows binaries are unsigned, so Gatekeeper or SmartScreen may warn. Checksums and GitHub attestations verify artifact integrity and provenance; they do not replace Apple Developer ID or Microsoft Authenticode signing. Do not disable platform security globally to run Papercuts.

### Go install

Developers with Go 1.24 or newer can install a pinned version:

```sh
go install github.com/Whamp/papercuts/cmd/papercuts@v0.1.0
```

## Project setup

Papercuts uses the command's exact working directory as project scope. It does not inspect Git or search parent directories.

For interactive setup, run:

```sh
papercuts init
```

In a terminal, Papercuts shows the exact agent guidance and asks before changing `AGENTS.md`. Non-terminal runs skip `AGENTS.md` integration unless you pass `--agents`. To create only the log, run:

```sh
papercuts init --project --no-agents
```

Papercuts manages one `AGENTS.md` section bounded by `<!-- papercuts:begin -->` and `<!-- papercuts:end -->`. It preserves bytes outside that section and rejects ambiguous or malformed markers.

## What agents capture

An agent captures one concrete instance of friction encountered while pursuing another task. Every capture requires a severity and a description:

```sh
papercuts capture --severity low \
  "The formatter required an undocumented working directory."
```

Use `--stdin` for multiline input. Papercuts reads stdin only when the agent passes this flag:

```sh
cat <<'EOF' | papercuts capture --severity medium --stdin --reporter agent --model gpt-5-codex
The root-relative test path failed because the test runner started in apps/web.
Running the command from apps/web completed the task.
EOF
```

Choose the highest severity that applies:

- `low`: an avoidable detour that did not change the approach or reduce confidence
- `medium`: meaningful rework, repeated attempts, a workaround, a changed approach, or reduced confidence
- `high`: blocked work, required intervention, or credible risk of a wrong, destructive, or insecure result

Agents should capture fatal and non-fatal friction, then continue safe work when possible. They should stop unsafe work rather than continue merely to record it. Captures must never contain secrets.

## Global log

Agents use global scope for friction outside the current project or in shared tooling:

```sh
papercuts init --global --no-agents
papercuts capture --global --severity medium \
  "The shared tool returned stale documentation links."
```

The default global log is `~/.papercuts/PAPERCUTS.md`, based on the operating system's user-home directory. Override it for one invocation or through the environment:

```sh
papercuts init --global --global-path /absolute/path/PAPERCUTS.md --no-agents
PAPERCUTS_GLOBAL_PATH=/absolute/path/PAPERCUTS.md \
  papercuts capture --global --severity low "The shared tool required a workaround."
```

`--global-path` takes precedence over `PAPERCUTS_GLOBAL_PATH`. Overrides must be absolute paths. Papercuts does not expand `~` or environment references in override values.

## Storage and concurrency

Papercuts writes UTF-8 Markdown and preserves capture order. It locks the selected log, appends one complete entry, syncs it, and releases the lock. It rejects symbolic links and non-regular targets. It creates no lock sidecars.

Papercuts targets local filesystems on Linux, macOS, and Windows. Locking and durability on network, FUSE, cloud-synced, and other unusual filesystems depend on that filesystem and are not guaranteed.

## Upgrade, rollback, and uninstall

Stop running Papercuts commands, verify the desired archive, and replace the executable to upgrade or roll back. Published versions are immutable; install an earlier named release to roll back.

To uninstall, delete the executable. Papercuts leaves project and global logs in place. Delete those logs separately only when you intend to delete their contents.

## Develop

Papercuts requires Go 1.24 or newer.

```sh
go test ./...
go test -race ./...
go vet ./...
golangci-lint run ./...
GOTOOLCHAIN=go1.26.5 go run golang.org/x/vuln/cmd/govulncheck@v1.6.0 ./...
go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.11
uvx zizmor==1.26.1 --pedantic .github/workflows
```

The vulnerability scan uses the patched release toolchain. Native and cross-compilation CI retain Go 1.24 compatibility coverage. The final command uses [`uvx`](https://docs.astral.sh/uv/guides/tools/). CI enforces both workflow checks with immutable tool and action pins.

Build the command:

```sh
go build ./cmd/papercuts
```

The release pipeline compiles Linux, macOS, and Windows binaries for `amd64` and `arm64`. Native CI executes filesystem-lock and atomic-replacement tests on each operating system; cross-compilation alone does not prove those semantics.
