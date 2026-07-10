# Research: Reliable Cross-Platform Persistence in Go for `papercuts`

## Summary

Use the standard library for deterministic path resolution and race-safe creation, plus a pinned `github.com/gofrs/flock` dependency for one exclusive sidecar-lock protocol shared by `init` and `capture`. Resolve and lexically clean paths without expansion; reject symlinks and non-regular targets; build one UTF-8/LF entry in memory; under a bounded lock open the existing target with `O_WRONLY|O_APPEND`, revalidate the opened handle, perform one `Write`, `Sync`, and close before unlocking.

This is small and dependable for cooperating local processes on Linux, macOS, and Windows. It does **not** make a security boundary against hostile path replacement, hard-link aliases, non-cooperating writers, or filesystems whose advisory locks/durability semantics are weak or remote.

## Findings

1. **Resolve the settled paths literally and fail early.**
   - Project target: call `os.Getwd()` and `filepath.Join(wd, "PAPERCUTS.md")`. `Getwd` returns an absolute path; do not search parents.
   - Global target precedence: a *present* `--global-path`, then `os.LookupEnv("PAPERCUTS_GLOBAL_PATH")`, then `os.UserHomeDir()` joined with `.papercuts/PAPERCUTS.md`. `LookupEnv` distinguishes unset from explicitly empty. `UserHomeDir` uses `$HOME` on Unix/macOS and `%USERPROFILE%` on Windows and can fail; propagate that failure. For the default, also require the returned home to satisfy target-OS `filepath.IsAbs`, because these environment-derived values should not silently turn the global target into a relative path.
   - For either override, reject empty or `!filepath.IsAbs(value)`, then return `filepath.Clean(value)`. `Clean` is purely lexical and target-OS-aware; it does not access the filesystem. Do **not** call `ExpandEnv`, expand `~`, or turn a relative override absolute with `filepath.Abs`. Consequently, `C:\x` is not accepted as absolute by a Linux binary and `/x` is not interpreted using Windows rules by a Windows binary; validation is correctly performed for the running target OS. [Go `os` docs: `Getwd`, `LookupEnv`, `UserHomeDir`](https://pkg.go.dev/os#UserHomeDir) [Go `filepath` docs: `IsAbs`, `Clean`](https://pkg.go.dev/path/filepath#IsAbs)

2. **Use conservative creation permissions, but document that numeric Unix modes are not a Windows ACL policy.**
   - For global init, run `os.MkdirAll(filepath.Dir(target), 0o700)`. It is idempotent when the path is already a directory, uses the requested bits before umask for newly created parents, and does not tighten pre-existing directories.
   - Create target and sidecar lock files with `0o600`. On Unix, umask can remove requested permissions. On Windows, Go documents that only owner-write (`0o200`) affects the read-only attribute and other permission bits are unused; inherited Windows ACLs govern actual access. Thus `0700/0600` are privacy-preserving Unix defaults and sensible portable intent, not equivalent access-control guarantees on Windows.
   - Never `Chmod` an existing `PAPERCUTS.md` during idempotent init or capture. Project-file sharing policy can later choose `0o644`; `0o600` is the safer default absent such a contract. [Go `os.MkdirAll` and mode documentation](https://pkg.go.dev/os#MkdirAll) [Go `os.Chmod` Windows semantics](https://pkg.go.dev/os#Chmod)

3. **Make init idempotent and no-overwrite with `O_CREATE|O_EXCL`, not check-then-create or rename.**
   - After any required global `MkdirAll`, acquire the same sidecar lock used by captures, then call `os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)`.
   - `O_EXCL` closes the ordinary existence race. Go maps `O_CREATE|O_EXCL` to Windows `CREATE_NEW` and opens the reparse point rather than following it. If creation returns an error matching `fs.ErrExist`, use `Lstat`: return success unchanged only when the existing entry itself is regular; return a useful error for a symlink/reparse point, directory, device, or other type. Other errors are failures.
   - Prefer an empty initial file if the format permits. If init must write a header, build the complete UTF-8/LF header in memory, make one `Write`, verify the returned error, `Sync`, and close while still holding the lock. A crash after exclusive creation but before the header is fully durable can leave a partial existing file; idempotence forbids silently overwriting it on the next init. There is no standard-library, cross-platform “rename only if destination absent” replacement: `os.Rename` replaces an existing non-directory and is explicitly not atomic on non-Unix platforms. [Go `os.OpenFile`, `Write`, and `Rename` docs](https://pkg.go.dev/os#OpenFile) [Go Windows `Open`: `CREATE_NEW`, reparse-point handling](https://go.dev/src/syscall/syscall_windows.go#L434)

4. **Reject symlinks and validate both the path entry and opened object, while acknowledging portable TOCTOU limits.**
   - Before creating a lock file for capture, `os.Lstat(target)`: map not-exist to “not initialized; run `papercuts init`”; reject `ModeSymlink` (including Windows named-entity reparse points) and require `Mode().IsRegular()`.
   - Acquire the lock, repeat `Lstat`, then open without `O_CREATE`. Immediately call `f.Stat()` and again require `IsRegular()` before writing. The handle check prevents writing to a directory/device even when the pathname changes between checks.
   - `Lstat` deliberately does not follow symlinks; `Stat`/ordinary open do. These checks reduce accidents but do not atomically combine “no symlink” with open on every target OS. A hostile actor able to mutate the directory can swap entries between `Lstat` and `Open`; hard links and alternate Windows spellings can also bypass a lock keyed only by the cleaned pathname. Fully hostile-safe resolution requires OS-specific no-follow/open-by-handle code (or a constrained `os.Root` design) and is out of scope for a tiny portable CLI. [Go `os.Lstat`, `Stat`, `File.Stat`, and file modes](https://pkg.go.dev/os#Lstat) [Go `filepath.Clean` is lexical](https://pkg.go.dev/path/filepath#Clean)

5. **Serialize complete entries with a sidecar exclusive lock; do not treat `O_APPEND` as the whole concurrency design.**
   - Use a stable adjacent lock name such as `target + ".lock"`. Every `papercuts` command that mutates a target must use the same cleaned target and this same lock naming rule. Do not delete the lock file after use: the library closes/unlocks it but intentionally leaves it, avoiding races around deleting and recreating lock inodes.
   - Recommended dependency/API: `github.com/gofrs/flock`, pinned in `go.mod`; construct `flock.New(lockPath, flock.SetFlag(os.O_CREATE|os.O_RDWR), flock.SetPermissions(0o600))`. The explicit read/write flag avoids platform/filesystem surprises. The library uses `unix.Flock(... LOCK_EX|LOCK_NB)` on Linux and Darwin and Windows `LockFileEx` with `LOCKFILE_EXCLUSIVE_LOCK|LOCKFILE_FAIL_IMMEDIATELY`; its build tags cover Linux, Darwin, and Windows. [gofrs/flock Unix implementation](https://github.com/gofrs/flock/blob/master/flock_unix.go) [Windows implementation](https://github.com/gofrs/flock/blob/master/flock_windows.go) [options/defaults source](https://github.com/gofrs/flock/blob/master/flock.go)
   - Acquire with `TryLockContext`: e.g. a caller-derived context capped at **5 seconds** and a **25 ms** retry interval. The implementation retries non-blocking acquisition until success, lock error, or `ctx.Done()`, and returns `ctx.Err()` on timeout/cancellation. Report “timed out waiting for capture lock `<lockPath>`” while wrapping the context error. Always unlock in a defer, and preserve unlock errors (for example with `errors.Join` when an earlier operation also failed). [gofrs/flock `TryLockContext` source](https://github.com/gofrs/flock/blob/master/flock.go#L104) [package docs](https://pkg.go.dev/github.com/gofrs/flock#Flock.TryLockContext)
   - Build the entire Markdown record first, normalize CRLF and bare CR to `\n`, reject invalid UTF-8 with `utf8.Valid`, and append the exact framing/newline bytes in memory. Under the lock, use `os.OpenFile(target, os.O_WRONLY|os.O_APPEND, 0)`—without `O_CREATE`—and call `Write(entry)` once. `File.Write` reports a non-nil error when `n != len(entry)`; do not split one record over multiple writes.
   - `O_APPEND` is still useful: on Windows Go requests `FILE_APPEND_DATA`, and on Unix the OS append flag prevents a stale seek offset. However, the lock is what serializes a whole logical record across processes and makes behavior uniform; `os.File`’s goroutine safety alone does not coordinate separately opened files in separate processes. Non-cooperating writers remain outside the protocol. [Go `File.Write` and append example](https://pkg.go.dev/os#File.Write) [Go Windows `O_APPEND` mapping](https://go.dev/src/syscall/syscall_windows.go#L385) [Go `os` concurrency statement](https://pkg.go.dev/os#hdr-Concurrency)

6. **Make successful capture durable enough for a tiny CLI, with explicit limits.**
   - After the one successful write, call `f.Sync()` before close and before unlock; then check `Close`. Go defines `Sync` as committing current file contents to stable storage, and its Windows implementation calls `FlushFileBuffers`. This costs latency but is a reasonable dependable default for a tiny capture command.
   - `Sync` is not a promise against all hardware/controller failures. Syncing the new file does not portably guarantee that the parent-directory entry from init is durable; directory `fsync` is Unix-specific in practice and is not a uniform Go/Windows strategy. State init durability as best effort rather than claiming transactional crash consistency. [Go `File.Sync`](https://pkg.go.dev/os#File.Sync) [Go Windows `Fsync` source](https://go.dev/src/syscall/syscall_windows.go#L755)

7. **Wrap errors with action and quoted path while retaining machine-testable causes.**
   - Standard `os` path operations already return `*PathError` containing operation and path. Add domain context with `%w`, for example `fmt.Errorf("append capture to %q: %w", target, err)`, rather than flattening with `%v`.
   - Use `errors.Is` for `fs.ErrNotExist`, `fs.ErrExist`, `fs.ErrPermission`, `context.DeadlineExceeded`, and `context.Canceled`; use `errors.As` when path details are needed. Define stable domain sentinels only where the CLI/tests need classification (for example `ErrNotInitialized` and `ErrNotRegular`) and wrap both the sentinel and underlying cause only when useful. Never hide permission, lock, sync, close, or unlock failures behind a generic “capture failed.” [Go `errors` wrapping, `Is`, and `As`](https://pkg.go.dev/errors) [Go `os.PathError` behavior](https://pkg.go.dev/os#PathError)

8. **Test semantics with real subprocess contention, and distinguish cross-compilation from execution.**
   - Unit table tests: precedence (flag/env/default), explicitly empty overrides, absolute/relative paths per host, lexical cleaning, no expansion of `~`/`$VAR`/`%VAR%`, `UserHomeDir` failure injection through a resolver seam, UTF-8 rejection, and LF normalization.
   - Filesystem tests in `t.TempDir`: global parent creation; modes on Unix only; init twice leaves bytes unchanged; concurrent init processes all succeed while only one creation occurs; capture rejects absent, directory, FIFO/device where supported, and symlink/reparse point; existing permissions are unchanged; write/sync/close errors through a narrow filesystem seam.
   - Multi-process integration: spawn many helper processes against one target, then parse and assert exactly one intact record per process with no interleaving. Separately hold the sidecar lock to assert success-after-wait, timeout, and cancellation. Run these tests natively on Linux, macOS, and Windows CI because advisory-lock behavior cannot be validated by compilation.
   - Run `go test ./...` and `go test -race ./...` on each native runner. From one host, also compile every target to catch build-tag/API errors, for example `GOOS=linux GOARCH=amd64 go test -c -o /tmp/papercuts-linux.test ./...`, similarly Darwin, and `GOOS=windows ... -o ...exe`. `go test -c` compiles a test binary without running it; cross-compilation is not evidence that Windows/macOS filesystem semantics passed. [Go command `test -c` documentation](https://pkg.go.dev/cmd/go#hdr-Test_packages) [Official Go Windows cross-compilation wiki](https://go.dev/wiki/WindowsCrossCompiling)

## Concrete recommended protocol

```text
resolveTarget(scope, flagValue/presence, envValue/presence):
  project -> Getwd + /PAPERCUTS.md
  global override -> require IsAbs, then Clean; never expand
  global default -> UserHomeDir, require absolute, Join(.papercuts/PAPERCUTS.md)

init(global):
  MkdirAll(parent, 0700)                  # global only
  lock target + ".lock" (5 s / 25 ms)
  OpenFile(target, WRONLY|CREATE|EXCL, 0600)
  on ErrExist: Lstat; regular => success unchanged; otherwise error
  on create: write complete initial bytes once (or keep empty), Sync, Close
  unlock; retain lock file

capture:
  build and validate complete UTF-8/LF entry in memory
  Lstat target; require existing non-symlink regular file
  lock target + ".lock" (5 s / 25 ms)
  repeat Lstat validation
  OpenFile(target, WRONLY|APPEND, 0)       # never CREATE
  f.Stat; require regular
  Write(entry) exactly once; Sync; Close
  unlock; report every failure with operation and quoted path
```

Pin a reviewed `gofrs/flock` release compatible with the project’s Go floor. The first-party release page says v0.13.0 raises the minimum to Go 1.24; v0.12.1 is the preceding release, while the locking/options APIs used above originated no later than v0.11.0. Do not select a version implicitly via `latest`. [gofrs/flock releases](https://github.com/gofrs/flock/releases)

## Sources

- Kept: [Go `os` package documentation](https://pkg.go.dev/os) — authoritative API behavior for home/CWD, creation, modes, file types, append, sync, and path errors.
- Kept: [Go `path/filepath` package documentation](https://pkg.go.dev/path/filepath) — authoritative target-OS absolute-path and lexical-cleaning semantics.
- Kept: [Go Windows syscall source](https://go.dev/src/syscall/syscall_windows.go) — direct evidence for `FILE_APPEND_DATA`, `CREATE_NEW`, reparse-point handling, mode limitations, and `FlushFileBuffers`.
- Kept: [Go `errors` package documentation](https://pkg.go.dev/errors) — authoritative wrapping and classification guidance.
- Kept: [gofrs/flock package documentation](https://pkg.go.dev/github.com/gofrs/flock) — first-party public API and platform caveat.
- Kept: [gofrs/flock common source](https://github.com/gofrs/flock/blob/master/flock.go), [Unix source](https://github.com/gofrs/flock/blob/master/flock_unix.go), and [Windows source](https://github.com/gofrs/flock/blob/master/flock_windows.go) — direct evidence for retry/cancellation and OS lock primitives.
- Kept: [Go command documentation](https://pkg.go.dev/cmd/go#hdr-Test_packages) — authoritative `go test -c` behavior.
- Dropped: search-result summaries, tutorials, blogs, and third-party locking comparisons — excluded because the ticket requires official Go or first-party library sources.
- Dropped: `os.WriteFile` as an init strategy — it truncates existing files and can leave partial content after multi-call failure, violating no-overwrite init.
- Dropped: `os.Rename`-based publication — destination replacement and non-Unix atomicity caveats do not satisfy portable no-overwrite init.
- Dropped: in-process `sync.Mutex` alone — it cannot coordinate concurrent agent processes.

## Gaps

- Go’s portable APIs cannot provide an atomic no-symlink open plus pathname identity guarantee across Linux, macOS, and Windows. The recommended checks are robust against normal mistakes, not hostile directory mutation.
- Advisory lock behavior on network, FUSE, cloud-synced, or otherwise unusual filesystems remains filesystem-dependent. Native CI validates ordinary local filesystems only; a supported-filesystem policy is needed before claiming network-share support.
- The settled contracts do not specify initial file bytes or whether project files should be group/world-readable. This brief recommends an empty initial file and `0o600`; changing either should be an explicit product decision.
- Exact crash consistency for the newly created directory entry is not portable. File `Sync` improves content durability but is not a cross-platform transaction.
