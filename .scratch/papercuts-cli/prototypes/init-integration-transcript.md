# Initialization integration prototype

Question: How should initialization remain deterministic for agents while requiring explicit human consent before creating or changing `AGENTS.md`?

A throwaway Go state prototype exercised terminal and non-terminal calls against missing, plain, managed, and malformed `AGENTS.md` states. It performed no filesystem writes. The source was deleted after the run.

## Observed state transitions

| Invocation state | AGENTS.md state | Result |
| --- | --- | --- |
| Terminal, no integration flag, response is no/EOF | Existing | Initialize the log; leave `AGENTS.md` untouched. |
| Terminal, no integration flag, response is yes | Missing | Initialize the log; create `AGENTS.md` with the managed section. |
| Non-terminal, no integration flag | Existing | Initialize the log; skip integration without prompting. |
| `--agents` | Existing without markers | Initialize the log; append the managed section. |
| `--agents` | Existing valid managed section | Initialize the log; replace only the managed section. |
| `--agents` | Malformed or duplicate markers | Keep the initialized log; leave `AGENTS.md` unchanged; report integration failure. |
| `--no-agents` | Any | Initialize the log; leave `AGENTS.md` untouched. |

## Verdict

`--agents` is explicit non-interactive consent. `--no-agents` is an explicit prompt suppressor. Without either flag, only a terminal session offers the exact managed section and asks `Add Papercuts guidance to <path>? [y/N]`; every response except `y` or `yes` declines. Non-terminal execution skips integration. This makes unattended initialization deterministic and makes an `AGENTS.md` edit impossible without an affirmative act.

The log is initialized before integration. An integration failure does not delete a valid log; the diagnostic reports the partial outcome. Managed markers make refresh idempotent and confine replacement to bytes owned by Papercuts.
