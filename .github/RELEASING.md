# Release Papercuts

The canonical release channel is [github.com/Whamp/papercuts/releases](https://github.com/Whamp/papercuts/releases). GitHub Actions builds each artifact from an annotated repository tag; maintainers do not upload locally built binaries.

## One-time repository setup

1. Commit the selected root `LICENSE`; GoReleaser requires it in every archive.
2. Change the existing `Whamp/papercuts` repository from private to public. Never publish it before the license is committed.
3. Verify immutable releases remain enabled. They were enabled through GitHub's repository API during private bootstrap.
4. Allow GitHub Actions to create attestations and write release contents.
5. Keep the workflow permission default read-only. `.github/workflows/release.yml` requests write permissions only for its release job graph.

Do not push a release tag until every setup step is complete.

To validate the complete release graph without creating a tag or release, run:

```sh
gh workflow run Release --ref master
gh run watch --repo Whamp/papercuts
```

A manual run builds snapshot archives, verifies their contract, and runs native archive lifecycle smoke on Linux, macOS, and Windows. Its publish job is skipped.

## Version contract

- The first stable version is `v0.1.0`.
- Stable tags use `vMAJOR.MINOR.PATCH`.
- Prerelease tags use `vMAJOR.MINOR.PATCH-rc.N`, where `N` starts at 1.
- Never move, delete, or rebuild a published tag.
- Fix a released defect with a new patch version.

## Publish

1. Confirm the branch passes CI.
2. Confirm `PATCHED_GO_VERSION` in both `.github/workflows/ci.yml` and `.github/workflows/release.yml` names the same supported, current security-patched release listed by [`go.dev/dl`](https://go.dev/dl/). Update and revalidate both workflows before tagging when it does not.
3. Confirm `git status --short` is empty.
4. Create and push an annotated tag:

   ```sh
   git tag -a v0.1.0 -m "papercuts v0.1.0"
   git push origin v0.1.0
   ```

5. Watch the Release workflow. It:
   - validates the tag and source;
   - runs actionlint, a pedantic zizmor security audit, and a reachable-code vulnerability scan;
   - runs native tests and race tests on Linux, macOS, and Windows;
   - builds six archives with GoReleaser;
   - inspects archive contents and independently verifies SHA-256 checksums;
   - runs install, capture, replacement, rollback, and uninstall smoke tests on all three operating systems;
   - creates a draft GitHub release;
   - creates GitHub artifact attestations;
   - publishes the release only after every gate passes.
6. Verify the published release:

   ```sh
   gh release verify v0.1.0 --repo Whamp/papercuts
   gh attestation verify papercuts_0.1.0_linux_amd64.tar.gz --repo Whamp/papercuts
   ```

7. Download one archive independently, verify it against `checksums.txt`, extract it, and run `papercuts version`.

A failed run may leave a draft release. Inspect it before retrying. Delete only an unpublished draft and its assets; never replace a published release.

## Roll back

Published artifacts remain available because releases are immutable. To roll back a client, stop Papercuts commands and install a verified archive from the earlier release. Do not retag the earlier commit. User logs are data and remain untouched.

## Platform signing

Initial macOS and Windows archives are unsigned and may trigger Gatekeeper or SmartScreen warnings. Checksums and GitHub attestations do not replace Developer ID or Authenticode signing. Do not tell users to disable platform security. Add signing and notarization only through a reviewed release change with protected credentials.
