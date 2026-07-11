<!-- papercuts:begin -->
## Papercuts

Capture a **papercut**—one concrete friction instance encountered while doing other work—when it occurs:

```text
papercuts capture --severity <level> "<attempt; friction; impact; workaround or current state>"
```

Severity: `low` = detour only; `medium` = rework, retries, workaround, changed approach, or reduced confidence; `high` = blocked work, intervention, or credible wrong/destructive/insecure risk. Choose the highest applicable level.

Project captures target `PAPERCUTS.md` in the exact working directory; run from the directory containing that log. Use `--global` only for shared tooling or environment friction. Run `papercuts capture --help` for multiline input and other options. When a log is missing, surface the suggested init command and preserve the intended scope.

A capture is complete when the CLI confirms it and the description records one causal instance. Product bugs and feature requests belong in the product issue tracker. Refer to secrets by role, never value. Resume only safe work.
<!-- papercuts:end -->
