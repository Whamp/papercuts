<!-- papercuts:begin -->
## Papercuts

Capture concrete workflow friction encountered while pursuing another task:

```text
papercuts capture --severity <low|medium|high> "<what you tried, what happened, and the impact>"
papercuts capture --severity <low|medium|high> --stdin
```

- `low`: an avoidable detour that did not change your approach or confidence in the result.
- `medium`: meaningful rework, repeated attempts, a workaround, a changed approach, or reduced confidence while the task remained safely completable.
- `high`: blocked completion, required human or environment intervention, or created a credible risk of an incorrect, destructive, or insecure result.
- When several definitions apply, use the highest severity.

Project scope is the default. Add `--global` only for friction outside this project or in shared tooling. Supply exactly one quoted description, or pipe a multiline description with `--stdin`. Describe one instance and include a workaround or current state when known. Capture fatal and non-fatal friction; continue safe work, but do not run or continue unsafe work. Never include secrets. Use the product issue tracker for product bugs and feature requests.
<!-- papercuts:end -->
