# Complete Project and Global Scope Behavior

Type: grilling
Status: resolved

## Question

What exact initialization, working-directory, global-default, override, configuration, and failure behavior completes the agreed project/global scope model on Linux, macOS, and Windows?

## Answer

### Scope and target selection

- Project scope targets `<working-directory>/PAPERCUTS.md`. Resolve the process working directory once at command start. Do not inspect Git, search parents, follow a project marker, or silently switch scope.
- Global scope defaults to `<user-home>/.papercuts/PAPERCUTS.md`, using Go’s cross-platform user-home lookup. Do not substitute XDG, AppData, or another platform config directory.
- Project scope is the default for capture. A global selector opts into global scope. Explicit project and global selectors are mutually exclusive.
- A per-invocation `--global-path` value overrides the global target. Otherwise use non-empty `PAPERCUTS_GLOBAL_PATH`; otherwise use the default. The flag and environment variable name a file, not a directory. Both must be absolute after lexical cleaning. Do not expand `~` or environment references. The flag is invalid for project operations.
- The selected target remains fixed for the invocation. No failure may fall back from one scope or path to another.

### Initialization

- Initialization always states or determines one scope before writing. Project initialization uses the exact working directory. Global initialization uses the same override precedence as global capture.
- Create a missing parent directory only for the global target. Create the log atomically without overwriting an entry created concurrently.
- If the target already exists directly as a regular file, initialization succeeds without changing its contents and reports that it already exists. Reject symlinks and Windows reparse-point aliases, including links to regular files. If the target is a directory, non-regular file, link, or otherwise unusable, fail without replacing it.
- Initialization creates only the selected log plus its missing global parent directories. `AGENTS.md` integration is a separate, explicitly consented operation defined by “Design Initialization and AGENTS.md Integration.”
- Initialization is idempotent: repeating it leaves the filesystem unchanged and exits successfully.

### Capture prerequisites and failures

- Capture requires the selected log to exist as a regular readable and writable file. Read access is required to verify the final byte under the same locked handle before exact entry framing; write-only logs are rejected with an actionable access error. A missing project log tells the caller to initialize that exact working directory. A missing global log tells the caller to initialize the resolved global target.
- Fail before appending when the working directory or user home cannot be resolved, an override is empty or relative, selectors conflict, a project operation receives a global override, the target is missing or not a regular file, or required directory/file access fails.
- Report the selected scope, resolved path, operation, and underlying operating-system error where available. Write diagnostics to stderr and return a non-zero exit status. Never create or modify another path as recovery.
- Reject symlinks and Windows reparse-point aliases rather than following them; path aliases can bypass a lock keyed by the configured path. Display the configured path rather than claiming a canonical path. Cross-platform persistence defines race handling and permissions.

No persistent config file is needed. The explicit flag and environment variable form the complete global-path configuration interface.
