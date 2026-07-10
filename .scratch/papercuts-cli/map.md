# Specify the Papercuts CLI

## Destination

A tiny standalone `papercuts` CLI for Linux, macOS, and Windows, accompanied by an approved implementation specification. It captures project-scoped or global workflow friction, supports an overridable global location, and can optionally advertise itself in `AGENTS.md`.

## Notes

- Read [`CONTEXT.md`](../../CONTEXT.md) before resolving a ticket; use `domain-modeling` whenever the language changes.
- Use each ticket’s corresponding `research`, `prototype`, or `grilling` skill. Use `writing-great-skills` for every agent-facing instruction, help message, and error. Use `codebase-design` for implementation seams and `tdd` for the acceptance strategy.
- Source material: [Steve Ruiz’s post](https://x.com/steveruizok/status/2075303919664734295) and [screenshot](https://pbs.twimg.com/media/HMz1tvqWoAA6wh2?format=png&name=large).
- The executable will be written in Go and support Linux, macOS, and Windows without depending on a project runtime.
- Capture is the only lifecycle in scope. Every capture requires the reporting model to choose `low`, `medium`, or `high` severity.
- Project scope uses the agent session’s project root and never discovers it through Git or parent traversal. Project commands operate on their exact working directory.
- Global scope defaults to `~/.papercuts/PAPERCUTS.md` (the user-home equivalent on each platform) and supports a user override.
- Setup supports humans and agents: human-first interactive guidance plus deterministic non-interactive operation. Editing `AGENTS.md` is optional and requires explicit consent.
- Agent guidance will say not to include secrets; automatic secret detection is outside the destination.

## Decisions so far

<!-- Closed tickets are indexed here. -->

- [Recover the Original Tool’s Observable Contract](issues/01-recover-original-contract.md) — Public evidence verifies the purpose and a human-readable presentation shape; command, storage, metadata acquisition, setup, and exact guidance remain unknown.
- [Define What a Capture Must Contain](issues/02-define-capture-content.md) — A capture requires generated UTC time, severity, and concrete prose; explicit reporter/model attribution is optional, and issue-lifecycle fields are excluded.
- [Complete Project and Global Scope Behavior](issues/03-complete-scope-configuration.md) — Project scope is the exact working directory; global path precedence is flag, environment, then user-home default, with explicit idempotent initialization and no fallback.
- [Prototype the Capture Command](issues/04-prototype-capture-command.md) — `capture` requires explicit severity and either one description argument or deliberate stdin, defaults to project scope, and returns stable success, diagnostic, and exit-status behavior.
- [Prototype the Markdown Log Format](issues/05-prototype-markdown-format.md) — Each entry is a timestamp/severity heading, optional quoted attribution list, and blockquoted multiline observation appended to a minimal `# Papercuts` log.
- [Research Reliable Cross-Platform Persistence in Go](issues/06-research-cross-platform-persistence.md) — Literal path resolution, exclusive creation, regular-file checks, bounded cross-process locking, append-plus-sync, and native subprocess tests form the portable persistence protocol.
- [Design Initialization and AGENTS.md Integration](issues/07-design-initialization-integration.md) — Init is independently idempotent; AGENTS integration requires `--agents` or terminal consent and updates only one marked section without guessing at malformed files.
- [Finalize the Agent-Facing Guidance](issues/08-finalize-agent-guidance.md) — Reviewed canonical AGENTS guidance and CLI help cover capture criteria, severity, scope, multiline input, safe continuation, secrets, and actionable failures without setup or storage detail.
- [Choose the Distribution and Update Path](issues/09-choose-distribution-path.md) — Immutable, attested GitHub Releases built by GoReleaser are canonical; manual verified binaries lead, `go install` is secondary, and package channels, signing, installers, and self-update are deferred.
- [Choose the Implementation Architecture](issues/10-choose-implementation-architecture.md) — A three-operation deep application module sits behind the CLI, with target-handle locks and atomic publication/replacement as the only platform seams and explicit indeterminate persistence outcomes.
- [Choose the Repository License](issues/12-choose-repository-license.md) — Will selected canonical SPDX MIT text with `Copyright (c) 2026 Will Hampson` for the repository and every release archive.

## Not yet specified

- [Draft and Approve the Implementation Specification](issues/11-approve-implementation-specification.md) — The reviewed draft now has every product and technical field complete; final approval and public publication remain.

## Out of scope

- Automated repair, ticket generation, implementation, or code review. A future configurable loop may compose `to-spec → to-tickets → implement → code-review`.
- Product bug tracking or feature requests; `PAPERCUTS.md` is not a second issue tracker.
- Automatic secret detection or redaction.
- Git-based project discovery, parent-directory traversal, or implicit project-root guessing.
