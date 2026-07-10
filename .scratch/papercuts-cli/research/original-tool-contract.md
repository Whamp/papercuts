# Research: Steve Ruiz’s original Papercuts CLI observable contract

## Summary

The only located first-party evidence is Steve Ruiz’s July 9, 2026 X post and its attached screenshot. The post directly verifies a small CLI named **“papercuts”**, intended for agents to record frustrations encountered during work; the screenshot shows four human-readable records containing an ISO-8601 UTC timestamp, model identifier, user identifier, and free-form narrative. Neither source reveals command names or syntax, the on-disk filename, a serialization specification, how metadata is obtained, setup/install behavior, or the exact instructions given to agents. [Post](https://x.com/steveruizok/status/2075303919664734295) · [Screenshot](https://pbs.twimg.com/media/HMz1tvqWoAA6wh2?format=png&name=large)

## Scope and evidence labels

- **Directly verified** means stated in Steve Ruiz’s post.
- **Screenshot-derived observation** means visible in the attached first-party image, transcribed from the image; it does not establish hidden implementation behavior.
- **Reasonable inference** means a plausible interpretation of those observations, explicitly not part of the verified contract.
- **Unknown** means the public primary sources located do not answer the question.

This report makes no product decisions for a new implementation.

## Findings

### 1. Purpose and existence — directly verified

1. Steve Ruiz says: **“I added a tiny ‘papercuts’ cli tool”**. This directly verifies the tool name, that it is a CLI, and Ruiz’s characterization of it as small. [Post](https://x.com/steveruizok/status/2075303919664734295)
2. Ruiz says the tool is one **“that agents can use to complain about bullshit they encountered during work, like dead-end tool calls, broken links, or other frustrations.”** This directly verifies the intended user (agents), the action (record/express complaints), the timing/context (during work), and example complaint classes. It does not define a schema or exhaustive taxonomy. [Post](https://x.com/steveruizok/status/2075303919664734295)
3. Ruiz explains the motivation: **“The models would usually just push through without mentioning any problems.”** The observable purpose is therefore to make otherwise-unreported agent friction visible. The post does not say what later consumes or triages the records. [Post](https://x.com/steveruizok/status/2075303919664734295)

### 2. Command surface

#### Directly verified

- The executable/tool is called **`papercuts`** in prose and is a CLI. No invocation is visible in the post text. [Post](https://x.com/steveruizok/status/2075303919664734295)

#### Screenshot-derived observations

- The screenshot displays four formatted complaint records, but no shell prompt, executable invocation, subcommand, flags, positional arguments, help text, version output, or exit status is visible. [Screenshot](https://pbs.twimg.com/media/HMz1tvqWoAA6wh2?format=png&name=large)

#### Reasonable inference (not verified)

- Because the screenshot presents several records together, some operation probably writes or displays accumulated complaints. The image does **not** distinguish among CLI output, a file opened in a terminal, or a diff, so neither a “list” command nor any command name can be inferred safely. [Screenshot](https://pbs.twimg.com/media/HMz1tvqWoAA6wh2?format=png&name=large)

#### Unknown

- Exact binary spelling/casing at the shell level.
- Subcommands (for example, add/list/setup), aliases, flags, positional arguments, stdin behavior, interactive prompts, output modes, help/version commands, exit codes, validation, duplicate handling, and error messages.

### 3. Record/file format

#### Screenshot-derived observations

The image shows four records with this repeated *rendered* shape:

```text
2026-07-08T21:13:30.864Z - gpt-5-codex - stephenruiz

While investigating PRD-4202, I used unquoted zsh ** globs for rg and got 'no matches found'
before rg could run. Quoting globs or using rg --files avoids this shell-level miss.
```

The other visible headers are:

```text
2026-07-09T02:05:18.605Z - gpt-5-codex - stephenruiz
2026-07-09T02:09:10.915Z - gpt-5-codex - stephenruiz
2026-07-09T06:20:04.413Z - gpt-5-codex - stephenruiz
```

The header fields are separated visually by ` - `; a blank line separates each header from a free-form prose body; records are separated by whitespace. Timestamps include milliseconds and the `Z` UTC designator. [Screenshot](https://pbs.twimg.com/media/HMz1tvqWoAA6wh2?format=png&name=large)

The four bodies describe concrete work friction and include both what failed and useful context/remediation:

- unquoted zsh `**` globs prevented `rg` from running; quoting or `rg --files` was named as the remedy;
- a root-relative Vitest path failed because the workspace test CWD was `apps/web`;
- a tab-indented `deploy.yml` block caused `yarn format` YAML parsing to fail;
- no `yarn supabase` helper was available and a global Supabase CLI warned that it was old. [Screenshot](https://pbs.twimg.com/media/HMz1tvqWoAA6wh2?format=png&name=large)

#### Reasonable inferences (not verified)

- The visible layout is Markdown-compatible plain text, but it could also be rendered CLI output or diff content. It is therefore reasonable only to call it a **human-readable presentation format**, not a verified storage format. [Screenshot](https://pbs.twimg.com/media/HMz1tvqWoAA6wh2?format=png&name=large)
- The bodies suggest the tool accepts unconstrained prose rather than requiring one of the example categories, but no input grammar is shown. [Screenshot](https://pbs.twimg.com/media/HMz1tvqWoAA6wh2?format=png&name=large)
- Leading `+` marks visible at the left edge are consistent with added lines in a diff, but the screenshot does not expose enough surrounding UI to prove that interpretation. They should not be treated as record delimiters or literal file content. [Screenshot](https://pbs.twimg.com/media/HMz1tvqWoAA6wh2?format=png&name=large)

#### Unknown

- Filename and location; whether storage is project-local or global.
- Whether the backing representation is Markdown, JSON, JSONL, a database, or another format.
- Formal field names, escaping, multiline rules, ordering guarantees, encoding, locking/concurrency, append semantics, rotation, retention, and schema/version markers.
- Whether the displayed order is insertion order, chronological order, or a sorted view.

### 4. Metadata acquisition

#### Screenshot-derived observations

- Every visible record carries three header values: timestamp, `gpt-5-codex`, and `stephenruiz`. [Screenshot](https://pbs.twimg.com/media/HMz1tvqWoAA6wh2?format=png&name=large)
- Across all four records, model and user are constant while timestamps differ. [Screenshot](https://pbs.twimg.com/media/HMz1tvqWoAA6wh2?format=png&name=large)

#### Reasonable inferences (not verified)

- `gpt-5-codex` is plausibly a model identifier and `stephenruiz` is plausibly a user/operator identifier, based on their positions and values. The screenshot contains no labels, so their semantic field names are inferred rather than stated. [Screenshot](https://pbs.twimg.com/media/HMz1tvqWoAA6wh2?format=png&name=large)
- The timestamps are plausibly generated at record creation, but the sources do not establish whether they are generated by the CLI, supplied by the agent, or added by another layer. [Screenshot](https://pbs.twimg.com/media/HMz1tvqWoAA6wh2?format=png&name=large)

#### Unknown

- How timestamp, model, and user values are acquired: CLI arguments, environment variables, agent-runtime metadata, Git configuration, OS account, config file, or prompt-generated text.
- Whether metadata is optional, validated, overrideable, authenticated, or extensible.
- Whether repository, branch, commit, task/PRD, working directory, agent/session ID, tool call, command, exit code, or links are captured structurally. `PRD-4202` appears in prose bodies, not visibly as a separate metadata field. [Screenshot](https://pbs.twimg.com/media/HMz1tvqWoAA6wh2?format=png&name=large)

### 5. Setup and installation behavior

#### Directly verified

- Nothing beyond Ruiz saying he “added” the CLI is disclosed. [Post](https://x.com/steveruizok/status/2075303919664734295)

#### Public-package/repository checks

- The unscoped npm name `papercuts` returns **“package 'papercuts' not found”** on npm’s official package page. This only rules out that exact public npm package name at research time; it does not rule out a scoped, private, unpublished, differently named, or non-npm tool. [npm](https://www.npmjs.com/package/papercuts)
- The candidate scoped npm page `@steveruizok/papercuts` did not expose package metadata and redirected to sign-in during retrieval, so it cannot be cited as evidence that such a package exists. [npm candidate](https://www.npmjs.com/package/@steveruizok/papercuts)
- Focused searches of Steve Ruiz’s public GitHub surface did not locate a first-party Papercuts repository or source file. His public GitHub profile is the authoritative repository index checked, but absence from search is not proof that no source exists. [GitHub profile](https://github.com/steveruizok)

#### Unknown

- Package manager, package/repository name, language/runtime, supported operating systems, install command, global versus local installation, version, license, release history, dependencies, update mechanism, and uninstall behavior.
- Whether “setup” creates a file, modifies `AGENTS.md` or another agent instruction file, installs hooks, adds PATH entries, detects an agent runtime, or is idempotent.
- Behavior in an existing repository, monorepo/worktree behavior, permissions, and failure/rollback behavior.

### 6. Agent guidance

#### Directly verified

- Agents are intended to complain about encountered friction, with explicit examples: **“dead-end tool calls, broken links, or other frustrations.”** [Post](https://x.com/steveruizok/status/2075303919664734295)
- The motivation is to counter models silently pushing through problems without reporting them. [Post](https://x.com/steveruizok/status/2075303919664734295)

#### Screenshot-derived observations

- Each visible complaint is specific and contextual: it names the task (`PRD-4202`), attempted operation, observed failure, and often a workaround or desired improvement. [Screenshot](https://pbs.twimg.com/media/HMz1tvqWoAA6wh2?format=png&name=large)
- The records cover friction even when work ultimately succeeded (for example, “The migration path worked, but…”), so screenshot examples are not limited to fatal blockers. [Screenshot](https://pbs.twimg.com/media/HMz1tvqWoAA6wh2?format=png&name=large)

#### Reasonable inference (not verified)

- The examples suggest useful agent guidance would encourage recording non-fatal workflow friction and enough context to reproduce or improve it. That is an observed pattern in four examples, not a recovered instruction text or normative requirement. [Screenshot](https://pbs.twimg.com/media/HMz1tvqWoAA6wh2?format=png&name=large)

#### Unknown

- Exact prompt/instruction wording and where it is installed.
- When an agent should report, what threshold applies, whether it must continue work afterward, whether it should avoid duplicates, and whether user approval is needed.
- Required body fields, tone, length, categories, privacy/redaction rules, security constraints, and whether agents may report user or secret data.
- Whether agents are instructed to inspect existing papercuts before adding one, summarize at session end, or notify a human.

## Observable contract matrix

| Area | Verified/observed minimum | Not established |
|---|---|---|
| Command surface | A CLI referred to as `papercuts` exists for agents. [Post](https://x.com/steveruizok/status/2075303919664734295) | Any command, flag, argument, help, output, or exit-code contract. |
| Record presentation | Repeated human-readable records show timestamp, apparent model, apparent user, blank line, and multiline prose. [Screenshot](https://pbs.twimg.com/media/HMz1tvqWoAA6wh2?format=png&name=large) | Backing file type, filename, schema, delimiters, escaping, ordering, persistence. |
| Metadata | Values resembling UTC timestamp, model ID, and user ID are displayed. [Screenshot](https://pbs.twimg.com/media/HMz1tvqWoAA6wh2?format=png&name=large) | Acquisition source, field labels, override/validation rules, additional metadata. |
| Setup | No setup behavior is shown. | Install/setup commands, files modified, idempotence, runtime/platform. |
| Agent guidance | Record work frustrations that models would otherwise silently push through; examples include dead ends and broken links. [Post](https://x.com/steveruizok/status/2075303919664734295) | Exact instruction text, trigger/threshold, privacy, deduplication, lifecycle. |

## Sources

### Kept

- [Steve Ruiz X post, status 2075303919664734295](https://x.com/steveruizok/status/2075303919664734295) — first-party announcement; establishes purpose, intended user, and examples.
- [Attached screenshot, `HMz1tvqWoAA6wh2`](https://pbs.twimg.com/media/HMz1tvqWoAA6wh2?format=png&name=large) — first-party visual evidence for the displayed record shape and examples.
- [npm: `papercuts`](https://www.npmjs.com/package/papercuts) — official registry evidence that the exact unscoped name was not a public npm package at research time.
- [npm candidate: `@steveruizok/papercuts`](https://www.npmjs.com/package/@steveruizok/papercuts) — retained only to document an inconclusive candidate check; it exposed no package metadata.
- [Steve Ruiz GitHub profile](https://github.com/steveruizok) — first-party public repository index checked; no matching source was located.

### Dropped

- Unrelated projects named Papercut/Papercuts (image tools, SMTP/print-management software, and AI testing services) — name collision, no connection to Steve Ruiz’s tool.
- Third-party repost/index pages — no additional first-party evidence and no recoverable command or source details.
- Search-result guesses about possible package names — excluded because no package metadata or source code verified them.

## Gaps and next steps

No public primary source located in this pass exposes source code, repository history, package metadata, README/help output, release artifacts, or the agent instruction text. The exact command surface, durable file contract, metadata acquisition mechanism, setup side effects, and normative agent guidance therefore remain unknown.

The strongest next evidence would be, in order: (1) a first-party repository or gist from Steve Ruiz; (2) the tool’s `--help` and setup output; (3) the generated file plus any `AGENTS.md`/agent-config diff; (4) package metadata or a release artifact; or (5) a follow-up first-party post. Until one appears, the screenshot’s presentation should not be promoted into an implementation specification.

## Research method note

The screenshot was inspected through its original first-party media URL and OCR was used to aid transcription. Ambiguous OCR tokens were reconciled against visible terminal conventions (notably `rg --files`); only clearly recoverable text is quoted, and hidden behavior is not inferred from OCR.
