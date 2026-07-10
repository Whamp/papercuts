# Research Reliable Cross-Platform Persistence in Go

Type: research
Status: resolved
Blocked by: 03, 05

## Question

Which Go filesystem and configuration strategies safely support concurrent agent captures, directory and file creation, path overrides, permissions, and useful failures across Linux, macOS, and Windows?

## Answer

The portable protocol is:

- Resolve project paths with `os.Getwd`; resolve global paths through explicit flag, `os.LookupEnv`, then `os.UserHomeDir`. Require absolute overrides and apply only `filepath.Clean`—never expansion or implicit absolutization.
- Create missing global directories with `os.MkdirAll(..., 0o700)`. Create logs with `os.OpenFile` using `O_WRONLY|O_CREATE|O_EXCL` and `0o600`; write the complete `# Papercuts\n` header once, `Sync`, and close. Existing regular files remain unchanged. Unix modes are filtered by umask; Windows ACLs, not these bits, control access.
- Reject symlinks/reparse points and non-regular targets with `Lstat`, then validate the opened handle with `File.Stat`. This prevents ordinary aliasing and device/directory mistakes but is not a security boundary against hostile directory mutation or hard links.
- Serialize every init and capture for a target with one cross-process exclusive-lock protocol. Go has no unified standard-library lock. The smallest off-the-shelf bounded strategy is pinned `github.com/gofrs/flock` with an adjacent persistent sidecar, `TryLockContext`, a 5-second cap, and 25 ms retries. A target-handle platform adapter may avoid the sidecar, but it must provide the same bounded Linux/Darwin `flock` and Windows `LockFileEx` semantics and pass the same native tests; “Choose the Implementation Architecture” owns that tradeoff.
- Build and validate the complete UTF-8/LF entry before locking. Under the lock, revalidate, open the log with `O_WRONLY|O_APPEND` and no `O_CREATE`, add a missing final LF if needed, perform the complete append, `Sync`, close, then unlock. Preserve write, sync, close, and unlock errors with operation and quoted-path context.
- Treat unusual network, FUSE, and cloud-synced filesystems as unsupported unless their advisory-lock behavior is separately verified. Neither `O_APPEND`, an in-process mutex, nor cross-compilation alone proves multi-process safety.

Verification must include table tests for path precedence/validation and serialization, temporary-filesystem tests for creation/types/permissions, real subprocess contention and timeout tests, `go test` plus `go test -race` on native Linux, macOS, and Windows runners, and `go test -c` cross-compilation for all three targets. Native execution—not compilation—proves lock behavior.

The full primary-source analysis, caveats, and links are in [cross-platform persistence research](../research/cross-platform-persistence.md).
