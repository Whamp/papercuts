# Prototype the Capture Command

Type: prototype
Status: resolved
Blocked by: 01, 02, 03

## Question

What non-interactive command shape makes required severity, project/global selection, descriptions, multiline input, successful output, and actionable failures equally clear to agents and humans?

## Answer

Use one prompt-free command:

```text
papercuts capture --severity <low|medium|high> [options] <description>
papercuts capture --severity <low|medium|high> [options] --stdin
```

Options:

```text
--project              capture in the current project (the default)
--global               capture in the global log
--global-path <path>   override the global log for this invocation; requires --global
--reporter <label>     attach an optional reporter label
--model <label>        attach an optional model label
--stdin                read the complete description from standard input
```

- Require exactly one quoted description argument or `--stdin`, never both. Read stdin only when `--stdin` is present. Trim the description’s outer whitespace and preserve internal newlines.
- Require `--severity` on every capture. Do not accept abbreviations, numeric levels, defaults, or case variants.
- Project scope is the default. `--project` exists for explicit scripts. Reject `--project` with `--global`, and reject `--global-path` without `--global`.
- Treat empty descriptions and attribution labels that become empty after trimming as invalid input; omitted attribution remains valid.
- Do not prompt, open an editor, infer prose from unrelated arguments, or read piped input implicitly.

On success, write one line to stdout and nothing to stderr:

```text
Captured <severity> <project|global> papercut in <resolved-path>
```

On failure, write one actionable diagnostic to stderr prefixed by `papercuts: capture:`. Include the rejected option or value for usage errors. Include the selected scope and resolved path for filesystem errors, plus an initialization command when the log is missing. Do not print a stack trace or partial success message.

Exit `0` after a durable append, `2` for command-line/validation errors, and `1` for path, configuration, locking, or persistence failures. `papercuts capture --help` writes usage to stdout and exits `0`.

The exercised no-persistence prototype and failure examples are preserved in the [capture command transcript](../prototypes/capture-command-transcript.md); the throwaway source was deleted.
