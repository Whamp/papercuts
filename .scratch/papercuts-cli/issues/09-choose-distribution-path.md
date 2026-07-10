# Choose the Distribution and Update Path

Type: research
Status: resolved

## Question

What installation, release, upgrade, and uninstall approach best delivers a standalone Go executable to the intended Linux, macOS, and Windows users while keeping adoption small and dependable?

## Answer

Use public GitHub Releases in `github.com/Whamp/papercuts` as the canonical channel. GoReleaser in GitHub Actions builds six standalone archives for `linux/{amd64,arm64}`, `darwin/{amd64,arm64}`, and `windows/{amd64,arm64}`. Unix archives are `.tar.gz`; Windows archives are `.zip`; each contains the executable, README, and license. Publish one SHA-256 checksum file.

Create releases from validated SemVer tags (`vMAJOR.MINOR.PATCH`, with `-rc.N` prereleases), first as drafts. Run the complete validation gate, smoke the archives, create GitHub artifact attestations, then publish with immutable releases enabled. Never rebuild, retag, or replace a published version. Start at `v0.1.0`; roll forward with a patch while retaining prior immutable assets for rollback.

Expose `papercuts version` and `papercuts --version` with version, commit, and build date injected by GoReleaser. Local builds report `devel`; `go install` builds fall back to `runtime/debug.ReadBuildInfo`. The release workflow runs formatting, vet, tests, race tests where native, `goreleaser check`, a snapshot build, archive inspection, independent checksum verification, cross-compilation, and native smoke execution on Linux, macOS, and Windows before publication.

Manual verified binary installation is primary: download an archive plus checksums, verify, extract, run `papercuts version`, and copy the executable to a user-writable directory already on `PATH`. Upgrade and rollback replace the stopped executable with a verified named version. Uninstall deletes the executable and leaves user logs untouched unless the user explicitly removes them. Document exact commands per OS and optional `gh attestation verify`.

Document version-pinned `go install` as a secondary developer route. Defer shell/PowerShell installers, Homebrew, Scoop, WinGet, platform code signing/notarization, and self-update until demand justifies their additional trust and maintenance surfaces. Unsigned initial macOS and Windows archives are viable for informed users, but documentation must state that Gatekeeper or SmartScreen may warn and must not advise disabling platform security globally. Checksums and provenance do not substitute for platform publisher signing.

The full primary-source comparison, credentials, rollout constraints, and acceptance checks are in [distribution research](../research/distribution.md).
