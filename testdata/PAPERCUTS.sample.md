# Papercuts

## 2026-07-09T21:13:30.864Z — low

> The documented search example used an unquoted glob, so zsh rejected the command before `rg` ran. Quoting the glob worked.

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

## 2026-07-09T22:09:10.915Z — high

- Reporter: "Will Hampson"

> The deployment command selected the production account despite the documented staging flag, so I stopped before applying changes.
>
> ## This heading belongs to the description
>
> It remains inside the blockquote and cannot look like the next papercut.
