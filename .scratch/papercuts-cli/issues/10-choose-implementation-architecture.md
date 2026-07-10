# Choose the Implementation Architecture

Type: grilling
Status: resolved
Blocked by: 04, 05, 06, 07, 09

## Question

What deep module boundaries, configuration interface, storage interface, platform seams, and test strategy should the implementation use after the user-facing and filesystem contracts are settled?

## Answer

Use Go 1.24+ with module path `github.com/Whamp/papercuts`. Keep four conceptual modules:

```text
cmd/papercuts/        process composition only
internal/cli/         parsing, stdin/terminal interaction, help, diagnostics, exit codes
internal/papercuts/   deep domain/application module: targets, validation, formatting, init, capture, guidance
internal/buildinfo/   release and Go build metadata
internal/filelock/    native target-handle locking (real platform seam)
internal/atomicfile/  native no-replace publication and replacement (real platform seam)
```

Do not add Cobra, a generic filesystem, repositories, `util`/`common`, public storage interfaces, background goroutines, or a self-update module.

### Deep application interface

`internal/papercuts.Service` is concrete and synchronous. The CLI’s consumer-owned internal interface has three methods:

```go
type operations interface {
    Capture(context.Context, papercuts.CaptureRequest) (papercuts.CaptureResult, error)
    InitializeLog(context.Context, papercuts.InitializeRequest) (papercuts.InitializeResult, error)
    IntegrateGuidance(context.Context, papercuts.GuidanceRequest) (papercuts.GuidanceResult, error)
}
```

Requests carry typed scope selection, global-path presence/value, capture content, and attribution. Results always carry the selected scope/path and an effect: `unchanged`, `durable`, or `indeterminate`. Return a populated result with an error when bytes may have changed. The CLI never automatically retries an indeterminate append; it tells the user to inspect the log first.

`InitializeLog` resolves the invocation working directory once, returns the exact `AGENTS.md` path, and commits the log before the CLI seeks consent. The CLI then skips, prompts, or calls `IntegrateGuidance`. This explicit two-step saga preserves the required partial outcome without making terminal I/O part of the application module.

### Configuration and domain ownership

The service owns flag/environment/home precedence after parsing has preserved whether `--global-path` was present. Inject only package-private `Getwd`, `LookupEnv`, `UserHomeDir`, and clock functions for deterministic tests. One severity catalog supplies validation, generated AGENTS guidance, and help summaries. Pure package-private functions render entries and merge managed guidance; callers never assemble bytes.

Use a small purpose-built long-option parser so the reviewed command surface, `--` handling, explicit-empty values, and one-description-versus-stdin rule do not inherit a framework’s output or ordering behavior. `cli.Run(ctx, args, IO, operations, buildinfo)` is the executable test seam; `main` only wires OS streams, TTY detection, the service, and build metadata.

### Persistence

Reject persistent sidecar locks: they add undeclared worktree artifacts and aliases can bypass path-keyed locks. `internal/filelock` opens the actual target read/write with delete sharing where Windows requires it, then acquires a bounded whole-file exclusive lock using nonblocking `flock` on Linux/macOS or `LockFileEx` on Windows. Retry every 25 ms until a five-second deadline or caller cancellation. Write and unlock through the same handle.

Initialize a missing log by writing `# Papercuts\n` to a synced same-directory temporary file, then atomically publish it without replacing a winner. Unix uses no-replace hard-link publication; Windows uses a no-replace move with write-through semantics. A concurrent direct regular winner becomes idempotent `already exists`; aliases and other file kinds fail. This strengthens the researched direct `O_EXCL` sequence by ensuring a visible target is never an empty or partial header.

Capture validates and renders before opening the target. Under the target lock, revalidate path/handle identity, inspect the final byte, append one complete payload, and `Sync`. On a partial write, attempt truncate-and-sync rollback to the original size. Classify any failure that cannot prove rollback or durability as indeterminate. Require readable and writable access, as recorded in the corrected scope contract.

Guidance transformation is pure. For a change, write and sync a same-directory temporary file, preserve the existing mode/newline convention, revalidate source identity, then atomically replace. Unix uses rename; Windows uses `ReplaceFileW`. New `AGENTS.md` uses the no-replace publisher. Keep an existing target handle locked through publication; waiters compare path/handle identity after locking and retry on replacement. Never fall back to remove-then-rename or copy/truncate.

`internal/atomicfile` and `internal/filelock` are the only build-tagged production seams. Pin `golang.org/x/sys`; use `golang.org/x/term` only for terminal detection. Ordinary local filesystems are supported; network, FUSE, and cloud-synced filesystem guarantees remain explicitly unsupported.

### Test seams and gate

The pre-agreed TDD seams are:

1. `cli.Run` for parsing, streams, help, consent, diagnostics, success text, and exit codes;
2. the three `Service` operations for target resolution and complete domain transactions;
3. native `filelock` and `atomicfile` adapters for true OS variation;
4. package-private pure render/merge functions for golden and property/fuzz tests; and
5. the built executable for end-to-end and subprocess contention tests.

Use standard `testing`, real `t.TempDir` files, and subprocesses. Inject narrow opened-handle fakes only for short-write, sync, close, unlock, rollback, and indeterminate-effect paths that real files cannot force. Do not mock the filesystem.

Required evidence: exact golden matches for all three reviewed prototype artifacts; table and fuzz tests for validation/serialization/managed-region preservation; idempotent init and byte/mode preservation; native multi-process contention, timeout, cancellation, alias, publish, and replacement tests on Linux, macOS, and Windows; `gofmt`, `go vet`, `golangci-lint`, `go test`, native `go test -race`, six-target cross-compilation, CLI smoke tests, `goreleaser check`, snapshot archive inspection, and independent checksum verification. Cross-compilation is not native filesystem evidence.

Build metadata lives only in `internal/buildinfo`: GoReleaser linker values win, `runtime/debug.ReadBuildInfo` supports `go install`, and local builds report `devel`.
