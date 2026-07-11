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

### Linux and macOS

```sh
curl --proto '=https' --tlsv1.2 --fail --silent --show-error --location https://raw.githubusercontent.com/Whamp/papercuts/master/install.sh | sh
```

The installer detects the operating system and architecture, verifies the release checksum, and installs the latest stable version to `~/.local/bin`. When that directory is not on `PATH`, it prints the command needed for the current shell.

To install a specific version:

```sh
curl --proto '=https' --tlsv1.2 --fail --silent --show-error --location https://raw.githubusercontent.com/Whamp/papercuts/master/install.sh | PAPERCUTS_VERSION=v0.1.0 sh
```

### Windows

Run in PowerShell:

```powershell
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; irm https://raw.githubusercontent.com/Whamp/papercuts/master/install.ps1 | iex
```

The installer verifies the release checksum and installs the latest stable version to `$HOME\bin`. When that directory is not on the user `PATH`, it prints a command that adds it.

To install a specific version:

```powershell
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; $env:PAPERCUTS_VERSION = 'v0.1.0'; irm https://raw.githubusercontent.com/Whamp/papercuts/master/install.ps1 | iex
```

Rerun the installer to upgrade, or set `PAPERCUTS_VERSION` to install or restore a specific release. Set `PAPERCUTS_INSTALL_DIR` to override the installation directory. The installers do not modify shell profiles or `PATH`.

### Alternative installation methods

Download a platform archive and `checksums.txt` from [GitHub Releases](https://github.com/Whamp/papercuts/releases) for a fully manual installation. Verify the selected archive with `sha256sum` on Linux, `shasum --algorithm 256` on macOS, or `Get-FileHash -Algorithm SHA256` on Windows before extracting and copying the executable to a directory on `PATH`.

GitHub CLI users can additionally verify release provenance:

```sh
gh attestation verify papercuts_0.1.0_linux_amd64.tar.gz --repo Whamp/papercuts
```

Initial macOS and Windows binaries are unsigned, so Gatekeeper or SmartScreen may warn. Checksums and GitHub attestations verify artifact integrity and provenance; they do not replace Apple Developer ID or Microsoft Authenticode signing. Do not disable platform security globally to run Papercuts.

Developers with Go 1.24 or newer can install a pinned version directly:

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

Stop running Papercuts commands, then rerun the installer to upgrade. Set `PAPERCUTS_VERSION` to a previously published version to roll back. The installer verifies the selected archive before replacing the executable, and published versions are immutable.

To uninstall, delete the executable. Papercuts leaves project and global logs in place. Delete those logs separately only when you intend to delete their contents.

## Develop

Papercuts requires Go 1.24 or newer.

```sh
go test ./...
go test -race ./...
go vet ./...
GOTOOLCHAIN=go1.26.5 go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 run ./...
GOTOOLCHAIN=go1.26.5 go run golang.org/x/vuln/cmd/govulncheck@v1.6.0 ./...
go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.11
uvx zizmor==1.26.1 --pedantic .github/workflows
```

The pinned `go run` commands use the patched release toolchain and keep local golangci-lint behavior aligned with CI without requiring a separate installation. Native and cross-compilation CI retain Go 1.24 compatibility coverage. The final command uses [`uvx`](https://docs.astral.sh/uv/guides/tools/). CI enforces both workflow checks with immutable tool and action pins.

Build the command:

```sh
go build ./cmd/papercuts
```

The release pipeline compiles Linux, macOS, and Windows binaries for `amd64` and `arm64`. Native CI executes filesystem-lock and atomic-replacement tests on each operating system; cross-compilation alone does not prove those semantics.
