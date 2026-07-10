# Papercuts CLI Implementation Specification

Status: draft; approval blocked only by [issue 12](issues/12-choose-repository-license.md)

## 1. Purpose

Papercuts is a small standalone CLI that records concrete workflow friction encountered while a human or agent pursues another task. It appends readable Markdown to a project or global log. It runs on Linux, macOS, and Windows without a project runtime.

A Papercut is an observation, not a product bug, feature request, repair ticket, or lifecycle-managed issue. Capture is the only record lifecycle.

The public evidence for the original tool establishes its purpose and a human-readable presentation shape. It does not establish an exact command, storage format, setup flow, or implementation. This specification owns those choices. See [original-tool research](research/original-tool-contract.md).

## 2. Normative language

`must`, `must not`, `should`, and `may` are normative. The reviewed artifacts linked below are byte-level contracts where this specification says “exact.”

## 3. Capture domain

Each capture contains:

1. A CLI-generated UTC timestamp formatted with RFC 3339 and fractional seconds when present. Callers cannot override it.
2. One required severity:
   - `low`: an avoidable detour that did not change the approach or confidence in the result;
   - `medium`: meaningful rework, repeated attempts, a workaround, a changed approach, or reduced confidence while the task remained safely completable;
   - `high`: blocked completion, required human or environment intervention, or credible risk of an incorrect, destructive, or insecure result.
3. Required non-empty UTF-8 prose. The CLI trims outer whitespace and preserves internal whitespace and line breaks. Guidance asks for one concrete friction instance: what the reporter attempted or expected, what happened, the impact, and the workaround or current state when known. The CLI validates presence rather than prose quality.
4. Optional caller-supplied `reporter` and `model` labels. The CLI trims them, requires valid UTF-8, rejects empty labels, line breaks, and NUL, and renders other escapable controls safely. It never infers attribution from Git, OS accounts, or environment state.

Use the highest applicable severity. One domain catalog must supply severity parsing, help summaries, and managed guidance so their meanings cannot drift.

Captures have no ID, title, status, owner, assignee, due date, priority separate from severity, category, tag, comment, duplicate relation, resolution flow, or automatic repair-ticket behavior. Scope belongs to the selected log and is not repeated in each entry.

## 4. Command interface

### 4.1 Root and version

The executable is `papercuts`. It supports:

```text
papercuts capture ...
papercuts init ...
papercuts version
papercuts --version
papercuts --help
```

The exact reviewed root, capture, and init help is [cli-help.txt](prototypes/cli-help.txt). No arguments, `-h`, and `--help` write root help to stdout and exit `0`. An unknown command writes one diagnostic to stderr and exits `2`. `version` and `--version` accept no trailing arguments; valid forms write only the version line to stdout and exit `0`, while trailing arguments produce one stderr usage diagnostic and exit `2`. Any stdout or stderr write failure exits `1`.

Release builds print:

```text
papercuts v0.1.0 (commit 1a2b3c4, built 2026-07-09T12:00:00Z)
```

GoReleaser linker values take precedence. A version discovered through `runtime/debug.ReadBuildInfo` supports version-pinned `go install`. A local build prints `papercuts devel`.

### 4.2 Capture

```text
papercuts capture --severity <low|medium|high> [options] <description>
papercuts capture --severity <low|medium|high> [options] --stdin
```

Options:

```text
--project
--global
--global-path <absolute-file-path>
--severity <low|medium|high>
--reporter <label>
--model <label>
--stdin
-h, --help
```

Requirements:

- `--severity` is mandatory. Accept only exact lowercase `low`, `medium`, or `high`.
- Accept exactly one description argument or explicit `--stdin`, never both.
- Read standard input only with `--stdin`. Never infer piped input.
- Project scope is the default. `--project` and `--global` conflict.
- `--global-path` requires `--global`.
- Reject duplicate options rather than silently applying first- or last-value semantics.
- Honor `--` as the end of options.
- Never prompt, launch an editor, or infer prose from unrelated arguments.

A durable capture prints exactly:

```text
Captured <severity> <project|global> papercut in <resolved-path>
```

It writes nothing to stderr and exits `0`. Usage and validation failures produce one `papercuts: capture:` diagnostic on stderr and exit `2`. Target, configuration, lock, and persistence failures produce one actionable stderr diagnostic and exit `1`. An indeterminate append tells the caller to inspect the log before retrying.

### 4.3 Initialization

```text
papercuts init [--project|--global] [--global-path <absolute-file-path>] [--agents|--no-agents]
```

Project scope is the default. Scope and global-path rules match capture. `--agents` and `--no-agents` conflict, and duplicate options are invalid.

Initialize the selected log before deciding whether to integrate guidance. Report exactly one of:

```text
Initialized <project|global> papercuts log at <path>
<project|global> papercuts log already exists at <path>
```

Then:

- `--agents` explicitly authorizes guidance integration;
- `--no-agents` skips it;
- no consent flag plus terminal stdin prints the exact proposed section and asks `Add Papercuts guidance to <path>? [y/N]`;
- accept only case-insensitive `y` or `yes`;
- no consent flag plus non-terminal stdin skips without reading stdin.

A skipped integration prints `Skipped AGENTS.md integration`. Created, changed, and unchanged integrations use the exact success lines in [cli-help.txt](prototypes/cli-help.txt).

If log initialization fails, do not prompt or touch `AGENTS.md`. If guidance integration fails after durable log initialization, retain the log and report the two-part outcome. Diagnostics distinguish unchanged, durably completed, and indeterminate effects so callers do not retry an operation that may already have committed.

Initialization and guidance success lines write to stdout. The proposed managed section and interactive prompt write to stderr so command substitution does not consume the prompt. A decline or non-terminal skip writes its skip line to stdout and exits `0`. Init help exits `0`; init option or validation failures exit `2`; filesystem, locking, consent-input, persistence, guidance, and stream failures exit `1`. Every non-stream failure writes one actionable diagnostic to stderr. A failed output stream returns `1`; no diagnostic is guaranteed when stderr itself is unwritable. A post-log guidance failure may follow an already-written initialization success line on stdout because the log is a committed first step.

## 5. Scope and target resolution

Resolve a selected target once per invocation and retain the configured display path.

### Project

Project scope targets:

```text
<exact-command-working-directory>/PAPERCUTS.md
```

Resolve the working directory once with `os.Getwd`. Do not inspect Git, traverse parents, search for markers, infer an agent root, or fall back to global scope.

Guidance always targets `<exact-command-working-directory>/AGENTS.md`, including when the selected log scope is global.

### Global

Resolve global targets in this order:

1. explicit `--global-path`;
2. a non-empty `PAPERCUTS_GLOBAL_PATH`;
3. `<os.UserHomeDir()>/.papercuts/PAPERCUTS.md`.

Overrides name files, not directories. Apply `filepath.Clean`, then require an absolute path. Do not expand `~`, environment references, XDG paths, AppData, or relative paths. An explicitly empty flag is invalid. An empty environment value is absent. No persistent configuration file exists.

A custom missing global log diagnostic must preserve the selected target and recommend `papercuts init --global --global-path "<path>"`; it must not recommend an init command that resolves a different path.

## 6. Markdown storage format

Initialize a new log as these exact UTF-8 bytes:

```text
# Papercuts\n
```

Append each entry in this shape:

```markdown

## 2026-07-09T22:05:18.605Z — medium

- Reporter: "agent"
- Model: "gpt-5-codex"

> One concrete friction instance.
>
> Its workaround or current state.
```

Rules:

- The heading is `## <RFC3339-UTC> — <severity>`.
- Emit Reporter before Model; omit absent fields and the list when both are absent.
- Encode labels with Go `strconv.QuoteToGraphic`, producing one double-quoted Go string literal: printable graphic runes and spaces remain literal, quotes and backslashes use `\"` and `\\`, named controls use Go escapes such as `\t`, and other accepted non-graphic runes use Go hexadecimal or Unicode escapes.
- Normalize CRLF and CR description line endings to LF.
- Prefix every non-empty physical line with `> ` and every empty line with `>`.
- Preserve nested Markdown inside the blockquote.
- Build and validate the complete entry before locking.
- Begin the rendered entry with one blank line and end it with one LF.
- Under the same lock, add one LF before the entry when a non-empty existing file lacks a final LF.
- Capture never initializes or otherwise repairs existing content. An empty pre-existing file receives the rendered entry without an inserted header.
- Preserve capture order.

Do not add front matter, IDs, titles, scope, status, comments, a table of contents, or hidden schema markers. The exact representative log is [PAPERCUTS.sample.md](prototypes/PAPERCUTS.sample.md).

## 7. Initialization and persistence

### 7.1 Target validation

Reject symlinks and Windows reparse-point aliases, including aliases to regular files. Reject directories, sockets, named pipes, devices, dangling links, and other non-regular targets. Validate both the configured path with `Lstat` and the opened handle with `Stat`, and compare path/handle identity after lock acquisition.

This protocol protects ordinary local filesystem operation. It is not a security boundary against hostile parent-directory mutation or hard links. Network, FUSE, cloud-synced, and other filesystems with unverified advisory-lock semantics are unsupported.

### 7.2 Log initialization

For a missing global target only, create missing parent directories with `os.MkdirAll(..., 0o700)`. Do not create project parent directories.

Write `# Papercuts\n` to a mode-`0o600` same-directory temporary file, sync and close it, then atomically publish without replacing a concurrent winner:

- Linux and macOS use no-replace hard-link publication;
- Windows uses `MoveFileEx` with no replacement and write-through semantics.

A concurrent direct regular-file winner becomes idempotent `already exists`. An existing direct regular file remains byte-for-byte unchanged. Clean all temporary files. Unix modes are umask-filtered; Windows ACLs control effective access.

Linux syncs the parent directory after publication/replacement. Darwin does not sync the parent directory because Go file sync behavior and Darwin directory descriptors do not provide a safe portable equivalent. Windows uses write-through publication/replacement APIs.

### 7.3 Locking

Lock the selected target itself; never create persistent sidecar locks. `internal/filelock` opens the target read/write and acquires an exclusive whole-file lock:

- Linux/macOS: nonblocking `flock`;
- Windows: `LockFileEx`, with sharing flags required for atomic replacement.

Retry every 25 ms. Stop on caller cancellation or after the service's five-second cap. A caller deadline that expires sooner reports cancellation rather than falsely claiming five elapsed seconds. After the final sync or rollback attempt, unlock through the still-open handle and then close it. Preserve and join unlock and close failures with the primary operation error.

### 7.4 Capture transaction

Before locking, validate content and render the complete entry. Under the locked target handle:

1. revalidate target kind and path/handle identity;
2. inspect the final byte;
3. build the complete append payload, including missing-final-LF repair when required;
4. append the payload through the locked handle without `O_CREATE`;
5. sync;
6. unlock and close.

Capture requires readable and writable access. On partial write, attempt truncate-and-sync rollback to the original size. Return:

- `durable` only when the append is proven durable;
- `unchanged` only when no bytes committed or rollback is proven;
- `indeterminate` when durability or rollback cannot be proven.

Never retry an indeterminate append automatically.

### 7.5 Guidance transaction

The exact managed section is [AGENTS.sample.md](prototypes/AGENTS.sample.md), bounded by:

```text
<!-- papercuts:begin -->
<!-- papercuts:end -->
```

The section tells agents when to capture, defines severity and scope, describes argument and stdin forms, asks for one concrete instance and workaround/current state, permits safe continuation, forbids secrets, and routes product bugs/features to the product tracker. It excludes setup, storage internals, inferred attribution, ticket lifecycle, and repair instructions.

For a missing `AGENTS.md`, create a mode-`0o644` file containing only the managed section and publish it without replacement. For an existing regular file:

- no markers and no Papercuts Markdown heading: append one separated managed section;
- exactly one ordered marker pair: replace only the inclusive region;
- identical canonical content: no-op;
- malformed, missing, reversed, nested, or duplicate markers, or any unowned ATX/setext Papercuts heading: fail unchanged.

Preserve every byte outside the managed region, full Unix file mode including setuid/setgid/sticky bits, and consistent CRLF convention; otherwise insert LF. Hold the existing target lock through publication, revalidate identity, and retry when another process replaces the file before lock acquisition or publication.

Write and sync a same-directory temporary file, then atomically replace:

- Unix: `rename`;
- Windows: `ReplaceFileW`.

Never use remove-then-rename or copy/truncate fallbacks.

## 8. Implementation architecture

Use Go 1.24+ and module `github.com/Whamp/papercuts`.

```text
cmd/papercuts/       process composition only
internal/cli/        parsing, terminal/stdin interaction, help, diagnostics, exit codes
internal/papercuts/  deep synchronous application module
internal/buildinfo/  linker and Go build metadata
internal/filelock/   native target-handle lock adapter
internal/atomicfile/ native no-replace publication and replacement adapter
```

`internal/papercuts.Service` is concrete. The CLI owns this narrow consumer interface:

```go
type operations interface {
    Capture(context.Context, papercuts.CaptureRequest) (papercuts.CaptureResult, error)
    InitializeLog(context.Context, papercuts.InitializeRequest) (papercuts.InitializeResult, error)
    IntegrateGuidance(context.Context, papercuts.GuidanceRequest) (papercuts.GuidanceResult, error)
}
```

Requests preserve typed scope and explicit global-path presence. Results include selected scope/path and `unchanged`, `durable`, or `indeterminate` effect. `InitializeLog` commits before CLI consent handling, forming an explicit two-step saga.

Use a purpose-built long-option parser and pure package-private render/merge functions. Inject only package-private working-directory, environment, user-home, and clock sources for tests. `cli.Run` is the executable seam; `main` wires OS streams, TTY detection, service, and build info.

Only `internal/filelock` and `internal/atomicfile` may contain build-tagged production platform code. Pin `golang.org/x/sys`; use `golang.org/x/term` only for TTY detection.

Do not add Cobra, a generic filesystem or repository abstraction, public storage interfaces, `util`/`common`, background goroutines, persistent lock files, compatibility shims, or self-update modules.

## 9. Distribution

The canonical channel is public GitHub Releases at `github.com/Whamp/papercuts`. Start at `v0.1.0`. Accept only `vMAJOR.MINOR.PATCH` and `vMAJOR.MINOR.PATCH-rc.N` tags.

GoReleaser builds:

```text
papercuts_<version>_linux_amd64.tar.gz
papercuts_<version>_linux_arm64.tar.gz
papercuts_<version>_darwin_amd64.tar.gz
papercuts_<version>_darwin_arm64.tar.gz
papercuts_<version>_windows_amd64.zip
papercuts_<version>_windows_arm64.zip
checksums.txt
```

Every archive contains the executable, `README.md`, and the repository `LICENSE`. The license text cannot be specified or published until issue 12 records Will's legal/product choice. This is the only unresolved specification field.

The release workflow must:

1. validate the exact SemVer tag and run the complete checked-out source quality gate;
2. run formatting, vet, lint, native tests, and native race tests on Linux, macOS, and Windows;
3. run GoReleaser validation and a clean snapshot build;
4. inspect all archive names and contents;
5. recompute and verify every SHA-256 checksum independently;
6. cross-compile all six commands and test binaries;
7. smoke clean install, capture, stopped-executable replacement, rollback, and uninstall on Linux, macOS, and Windows while proving logs survive uninstall;
8. create a draft release;
9. create GitHub artifact attestations;
10. publish only after every gate succeeds.

Enable immutable releases in repository settings. Never rebuild, move, or retag a published version; roll forward with a patch. Upgrade and rollback replace a stopped executable with a verified named version. Uninstall removes only the executable.

Manual checksum-verified binary installation is primary. Version-pinned `go install` is secondary. Document exact commands for each OS and optional `gh attestation verify`.

Initial macOS and Windows archives are unsigned and may trigger Gatekeeper or SmartScreen. Do not advise globally disabling platform protections. Checksums and attestations do not replace Developer ID or Authenticode signing.

Defer shell/PowerShell installers, Homebrew, Scoop, WinGet, platform signing/notarization, and self-update. These are explicit non-goals, not hidden implementation hooks.

## 10. Acceptance gate

Tests use standard `testing`, `t.TempDir`, real files, and subprocesses. Do not mock the filesystem. Narrow test seams may force short writes, sync, close, unlock, rollback, and indeterminate outcomes that ordinary files cannot produce.

Approval requires fresh evidence for:

- exact golden matches for [AGENTS.sample.md](prototypes/AGENTS.sample.md), [PAPERCUTS.sample.md](prototypes/PAPERCUTS.sample.md), and the complete [cli-help.txt](prototypes/cli-help.txt);
- table and fuzz tests for parsing, validation, serialization, entry-boundary safety, and managed-region idempotence;
- path precedence, no expansion/fallback, exact cwd behavior, and custom-global diagnostics;
- idempotent and concurrent initialization;
- symlink, reparse-point, non-regular target, and permission rejection;
- global parent/log permissions and existing mode preservation;
- missing-final-LF framing and no implicit empty-file initialization;
- AGENTS byte, newline, complete mode, malformed-marker, heading, idempotence, and concurrent-replacement behavior;
- native target-lock contention, timeout, caller cancellation, hard-link aliasing, no-replace publication, and atomic replacement;
- real multi-process capture serialization;
- unchanged/durable/indeterminate diagnostics and rollback behavior;
- `gofmt`, `go vet`, configured `golangci-lint`, `go test ./...`, and native `go test -race ./...`;
- six-target command and test-binary cross-compilation;
- real executable project, global, stdin, interactive-consent, and recovery-command smoke tests;
- `goreleaser check`, six-archive snapshot inspection, and independent checksum verification;
- native Linux, macOS, and Windows CI and release-archive lifecycle smoke tests.

Cross-compilation proves build-tag and API compatibility only. It does not prove native lock, replacement, permission, or durability behavior.

## 11. Explicit non-goals

Papercuts does not:

- repair friction automatically;
- create, update, triage, or review product tickets;
- detect or redact secrets automatically;
- discover projects through Git or parent traversal;
- infer reporter/model identity;
- provide issue lifecycle fields;
- support unusual filesystems without separate verification;
- maintain persistent configuration;
- update itself;
- ship package-manager channels or installers in the initial release;
- claim platform publisher signing through checksums or provenance.

## 12. Approval state

Issues 01–10 supply all technical and product behavior above. The implementation and reviewed local validation exist through commit `0e76a24`. Private repository `Whamp/papercuts` preserves the unpublished history while the license is unresolved. Hosted CI run [`29066692111`](https://github.com/Whamp/papercuts/actions/runs/29066692111) passed formatting, vet, golangci-lint, reachable-vulnerability scanning, checked workflow scripts, exact six-target command/test cross-compilation, and native tests plus race tests on Linux, macOS, and Windows. Isolated tagged builds additionally passed exact snapshot/release archive, checksum-manifest, and distinct-version lifecycle validation. GitHub's repository API confirms immutable releases are enabled.

Issue 12 remains the sole high-level decision: the exact repository and archive license. Until Will selects it, this document remains a draft, issue 11 remains open, release archives cannot satisfy their content contract, and the repository must remain private.

After the license decision, approval requires adding its exact text as root `LICENSE`, replacing this draft status with approved status, resolving issues 12 and 11, and making the repository public. No other product or technical question remains open.
