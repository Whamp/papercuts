# Capture command prototype transcript

Question: Can one prompt-free command make severity, scope, descriptions, multiline input, success, and failures predictable for humans and agents?

Prototype: a throwaway Go parser with no persistence. It printed the complete parsed capture intent after each invocation. The source was deleted after the run.

## Single-line project capture

```console
$ papercuts capture --severity low 'The documented test path was relative to the repository root, but the test runner starts in apps/web.'
{
  "command": "capture",
  "scope": "project",
  "severity": "low",
  "description": "The documented test path was relative to the repository root, but the test runner starts in apps/web."
}
```

## Multiline global capture with attribution

```console
$ printf 'The first tool call failed.\nThe second attempt used the documented path and worked.\n' |
    papercuts capture --severity medium --global --stdin --reporter agent --model gpt-5-codex
{
  "command": "capture",
  "scope": "global",
  "severity": "medium",
  "description": "The first tool call failed.\nThe second attempt used the documented path and worked.",
  "reporter": "agent",
  "model": "gpt-5-codex"
}
```

## Explicit global path

```console
$ papercuts capture --global --global-path /tmp/papercuts.md --severity high \
    'The required service was unavailable, blocking verification.'
{
  "command": "capture",
  "scope": "global",
  "severity": "high",
  "description": "The required service was unavailable, blocking verification.",
  "globalPath": "/tmp/papercuts.md"
}
```

## Validation failures

```console
$ papercuts capture 'A description'
papercuts: capture: --severity is required (low|medium|high)

$ printf 'stdin text' | papercuts capture --severity low --stdin 'argument text'
papercuts: capture: provide either one description argument or --stdin, not both

$ papercuts capture --severity low --global-path /tmp/papercuts.md 'A description'
papercuts: capture: --global-path requires --global

$ printf '   \n' | papercuts capture --severity medium --stdin
papercuts: capture: description must not be empty
```

## Verdict

The command is readable without positional codes: `capture` names the action, severity is always explicit, project is the safe default, and global capture is visible at the call site. One quoted argument handles ordinary prose. `--stdin` makes multiline and generated input deliberate and prevents accidental reads from a pipe or terminal. Separate attribution flags avoid encoding metadata into prose. Usage errors can name the exact correction.
