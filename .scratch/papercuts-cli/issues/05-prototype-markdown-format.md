# Prototype the Markdown Log Format

Type: prototype
Status: resolved
Blocked by: 01, 02

## Question

What concise Markdown entry format is append-friendly, readable by humans and fixing agents, robust for multiline descriptions, and capable of carrying the agreed capture data without unnecessary structure?

## Answer

Initialize a new UTF-8 log as:

```markdown
# Papercuts
```

Append each papercut in this exact shape:

```markdown

## 2026-07-09T22:05:18.605Z — medium

- Reporter: "agent"
- Model: "gpt-5-codex"

> While verifying the project, `go test ./...` failed because the fixture path assumed a repository-root working directory.
>
> The workaround was:
>
> ```sh
> cd internal/capture && go test ./...
> ```
```

- The level-two heading is `## <RFC3339 UTC timestamp> — <severity>`. Timestamp and severity are generated/validated values, so they need no Markdown escaping.
- Emit attribution as a fixed-order list: Reporter, then Model. Omit absent fields and omit the list entirely when both are absent. Encode labels as quoted strings with escapes for quotes, backslashes, and controls; labels must be valid UTF-8 and contain no newline or NUL.
- Normalize CRLF and CR description line endings to LF. Prefix each non-empty physical line with `> ` and each empty line with `>`. This preserves multiline prose and nested Markdown while preventing a description heading from imitating an entry boundary.
- Build the complete entry in memory, beginning with one blank line and ending with one LF. Persistence appends that byte slice as one locked operation. If a pre-existing log lacks a final LF, add one before the entry while holding the same lock.
- Keep capture order. Do not add IDs, titles, scope fields, schema front matter, HTML comments, status fields, or a table of contents.

The [rendered sample log](../prototypes/PAPERCUTS.sample.md) covers entries with no attribution, both attribution fields, paragraphs, a code fence, and a heading inside a description. The blockquoted shape was selected over raw prose, which can imitate entry headings, and repeated front matter, which renders poorly and adds unnecessary schema.
