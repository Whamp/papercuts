# Design Initialization and AGENTS.md Integration

Type: prototype
Status: resolved
Blocked by: 03, 04

## Question

How should interactive and non-interactive initialization create the project log, request explicit permission to modify an existing or new `AGENTS.md`, and manage that optional integration without duplication or destructive edits?

## Answer

Use this command:

```text
papercuts init [--project|--global] [--global-path <absolute-path>] [--agents|--no-agents]
```

Project is the default. Scope and global-path behavior match `capture`. Initialize the selected log first and print whether it was created or already existed.

### Consent

- `--agents` is explicit consent to create or update `<working-directory>/AGENTS.md`. This path is exact; never search parent directories. It applies for either project or global log initialization, because `AGENTS.md` is project guidance even when it directs captures to a global log.
- `--no-agents` explicitly skips integration. Reject it when combined with `--agents`.
- With neither flag, if stdin is a terminal, print the target path and exact proposed managed section, then ask `Add Papercuts guidance to <path>? [y/N]`. Accept only case-insensitive `y` or `yes`. Empty input, EOF, and every other response decline without error.
- With neither flag and non-terminal stdin, skip integration without reading input. Never let a general confirmation flag imply AGENTS consent.

### Managed section

Papercuts owns only the inclusive region bounded by marker lines:

```markdown
<!-- papercuts:begin -->
## Papercuts
<canonical guidance finalized by “Finalize the Agent-Facing Guidance”>
<!-- papercuts:end -->
```

- Missing `AGENTS.md`: after consent, create a regular file containing only the managed section, with mode `0o644` before umask.
- Existing regular file with no markers or Papercuts heading: append one blank line as needed, then the section. Preserve every existing byte and use the file’s existing CRLF style only when it is consistently CRLF; otherwise use LF for inserted text.
- Existing regular file with exactly one ordered marker pair: replace only the inclusive managed region. Preserve all surrounding bytes, file mode, and newline style. Identical content is a no-op.
- A Papercuts Markdown heading without markers, one missing marker, reversed/nested markers, duplicate markers, a symlink/reparse point, or a non-regular target is an actionable error. Leave `AGENTS.md` byte-for-byte unchanged rather than guessing ownership.
- Publish a changed file through a same-directory durable temporary file and platform-safe atomic replacement. Clean up the temporary file on failure. “Choose the Implementation Architecture” selects the replacement adapter and tests its Windows behavior.

Log initialization and AGENTS integration are two reported outcomes. If log initialization fails, do not prompt or touch `AGENTS.md`. If integration fails after successful log initialization, keep the valid log, report `log initialized; AGENTS.md unchanged`, and exit `1`. Repeating a successful command is idempotent and never duplicates guidance.

The consent/file-state matrix is preserved in the [initialization integration prototype](../prototypes/init-integration-transcript.md).
