# Papercuts

`papercuts` records workflow friction encountered while doing other work. Each capture appends a timestamped, severity-rated observation to a Markdown log.

Papercuts are not product bugs, feature requests, or repair tickets. Send those to the product’s issue tracker.

## Install

Releases publish standalone archives on [GitHub Releases](https://github.com/Whamp/papercuts/releases) for:

- Linux: `amd64`, `arm64`
- macOS: `amd64`, `arm64`
- Windows: `amd64`, `arm64`

Each release includes six archives and `checksums.txt`. Every archive contains `README.md`, `LICENSE`, and either `papercuts` on Unix or `papercuts.exe` on Windows.

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

Replace `amd64` with `arm64` on Windows ARM. Add `$HOME\bin` to the user `PATH` when it is not already present.

GitHub CLI users can also verify release provenance:

```sh
gh attestation verify papercuts_0.1.0_linux_amd64.tar.gz --repo Whamp/papercuts
```

Initial macOS and Windows binaries are unsigned. Gatekeeper or SmartScreen may warn. Checksums and GitHub attestations verify artifact integrity and provenance; they do not replace Apple Developer ID or Microsoft Authenticode signing. Do not disable platform security globally to run Papercuts.

### Go install

Developers with Go 1.24 or newer can install a pinned version:

```sh
go install github.com/Whamp/papercuts/cmd/papercuts@v0.1.0
```

## Start a project log

Run commands from the project root supplied to the agent session. Papercuts uses that exact working directory; it does not inspect Git or search parent directories.

```sh
papercuts init
```

This creates `./PAPERCUTS.md`. In a terminal, Papercuts shows the exact managed guidance and asks before changing `./AGENTS.md`. Non-terminal runs skip AGENTS.md integration.

Use explicit flags in scripts:

```sh
papercuts init --project --no-agents
papercuts init --project --agents
```

`--agents` manages one section bounded by `<!-- papercuts:begin -->` and `<!-- papercuts:end -->`. Papercuts preserves bytes outside that section. It rejects ambiguous or malformed markers instead of rewriting the file.

## Capture friction

Every capture requires one severity and one description:

```sh
papercuts capture --severity low \
  "The formatter required an undocumented working directory."
```

Use `--stdin` for multiline input. Papercuts reads stdin only when this flag is present.

```sh
cat <<'EOF' | papercuts capture --severity medium --stdin --reporter agent --model gpt-5-codex
The root-relative test path failed because the test runner started in apps/web.
Running the command from apps/web completed the task.
EOF
```

Severity meanings:

- `low`: avoidable detour; approach and confidence remained intact
- `medium`: meaningful rework, retries, workaround, changed approach, or reduced confidence
- `high`: blocked work, required intervention, or credible risk of a wrong, destructive, or insecure result

Use the highest severity that applies. Capture non-fatal friction and continue safe work. Do not start or continue unsafe work merely to record it. Never include secrets.

## Global log

Use global scope for friction outside the current project or in shared tooling:

```sh
papercuts init --global --no-agents
papercuts capture --global --severity medium "The shared tool returned stale documentation links."
```

The default global log is `~/.papercuts/PAPERCUTS.md`, using the operating system’s user-home directory. Override the file for one invocation or through the environment:

```sh
papercuts init --global --global-path /absolute/path/PAPERCUTS.md --no-agents
PAPERCUTS_GLOBAL_PATH=/absolute/path/PAPERCUTS.md papercuts capture --global --severity low "..."
```

`--global-path` takes precedence over `PAPERCUTS_GLOBAL_PATH`. Overrides must name an absolute file path. Papercuts does not expand `~` or environment references in override values.

## Storage and concurrency

Papercuts writes UTF-8 Markdown and preserves capture order. It locks the selected log itself, appends one complete entry, syncs it, and releases the lock. It rejects symbolic links and non-regular targets. It does not create lock sidecars.

Local filesystems supported by Linux, macOS, and Windows are the target. Locking and durability on network, FUSE, cloud-synced, or other unusual filesystems depend on that filesystem and are not guaranteed.

## Upgrade, rollback, and uninstall

Stop running Papercuts commands, verify the desired archive, and replace the executable to upgrade or roll back. Published versions are immutable; install an earlier named release to roll back.

To uninstall, delete only the executable. Papercuts leaves project and global logs in place. Remove those logs separately only when you intend to delete their contents.

## Develop

Requirements: Go 1.24 or newer.

```sh
go test ./...
go test -race ./...
go vet ./...
golangci-lint run ./...
go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.11
uvx zizmor==1.26.1 --pedantic .github/workflows
```

The last command uses [`uvx`](https://docs.astral.sh/uv/guides/tools/). CI enforces both workflow checks with immutable tool and action pins.

Build the command:

```sh
go build ./cmd/papercuts
```

The release pipeline also compiles `linux`, `darwin`, and `windows` for `amd64` and `arm64`. Native CI executes filesystem-lock and atomic-replacement tests on each operating system; cross-compilation alone does not prove those semantics.
