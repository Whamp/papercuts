# Research: Choose the Distribution and Update Path for `papercuts`

## Summary

Launch with **GitHub Releases as the canonical channel**, built by **GoReleaser in GitHub Actions** into six standalone archives, with SHA-256 checksums and GitHub release/build provenance attestations. The initial route needs no Apple or Microsoft signing credential and no runtime on user machines: users manually download, verify, extract, and place one executable on `PATH`; `go install` is a secondary developer convenience. Defer installer scripts, Homebrew, Scoop, WinGet, native code signing/notarization, and in-process self-update until usage justifies their extra trust, credentials, and maintenance surfaces.

## Decision

### Initial supported route

1. Create the public repository `github.com/Whamp/papercuts`, establish the Go module at that path, and protect release tags/workflows. The current repository has neither remote nor commits, so Releases and Actions cannot exist until this bootstrap is complete.
2. Publish these GoReleaser assets for each stable tag:

   | Target | Archive | Executable inside |
   |---|---|---|
   | Linux amd64 | `papercuts_0.1.0_linux_amd64.tar.gz` | `papercuts` |
   | Linux arm64 | `papercuts_0.1.0_linux_arm64.tar.gz` | `papercuts` |
   | macOS amd64 | `papercuts_0.1.0_darwin_amd64.tar.gz` | `papercuts` |
   | macOS arm64 | `papercuts_0.1.0_darwin_arm64.tar.gz` | `papercuts` |
   | Windows amd64 | `papercuts_0.1.0_windows_amd64.zip` | `papercuts.exe` |
   | Windows arm64 | `papercuts_0.1.0_windows_arm64.zip` | `papercuts.exe` |

   These six targets cover the two current desktop CPU families on all requested operating systems while keeping the matrix small. Go lists `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`, and `windows/arm64` as supported port combinations. Defer 32-bit Linux/Windows and Linux ARMv7 until requested. [Go supported systems](https://go.dev/doc/install/source#environment)
3. Use `.tar.gz` on Unix-like systems and `.zip` on Windows. Include `README.md` and `LICENSE` in every archive. GoReleaser supports per-OS archive formats and templated names; its archives documentation explicitly shows selecting `zip` for Windows. [GoReleaser archives](https://goreleaser.com/customization/archive/)
4. Publish `papercuts_0.1.0_checksums.txt`, generated with SHA-256, covering the six archives. GoReleaser's checksum pipe generates a checksums file and defaults to SHA-256. [GoReleaser checksums](https://goreleaser.com/customization/checksum/)
5. Enable **immutable releases**. Build/upload everything to a draft, attest the assets, and only then publish it. GitHub locks the tag and assets after publication and automatically creates a cryptographically verifiable release attestation containing the tag, commit, and assets; GitHub itself recommends draft → attach all assets → publish. [GitHub immutable releases](https://docs.github.com/en/code-security/concepts/supply-chain-security/immutable-releases)
6. Also run `actions/attest@v4` over the six archives and checksum file after GoReleaser builds them, with `id-token: write`, `contents: read`, and `attestations: write`. This records build provenance, not merely integrity. Users can verify an asset with `gh attestation verify FILE -R Whamp/papercuts`. Artifact attestations are available for public repositories on GitHub Free/Pro/Team. [GitHub artifact attestations](https://docs.github.com/actions/security-for-github-actions/using-artifact-attestations/using-artifact-attestations-to-establish-provenance-for-builds)
7. Do **not** add a project runtime, installer, background updater, daemon, registry entry, or privileged system modification. Installation is copying one executable; uninstall is deleting it.

### Why this is the dependable small-adoption optimum

- **GitHub Releases** provide one stable, versioned source for binaries and release notes, while immutable assets eliminate silent replacement. GitHub can verify both an immutable release (`gh release verify TAG`) and a downloaded asset (`gh release verify-asset TAG FILE`). [GitHub release integrity](https://docs.github.com/en/code-security/how-tos/secure-your-supply-chain/secure-your-dependencies/verify-release-integrity)
- **GoReleaser** centralizes cross-compilation, archive layout, checksums, changelog/release creation, and linker flags instead of maintaining six bespoke build commands. Its official Actions recipe uses full Git history and the repository `GITHUB_TOKEN`. [GoReleaser GitHub Actions](https://goreleaser.com/ci/actions/) [GoReleaser Go builds](https://goreleaser.com/customization/builds/go/)
- **Manual installation** is transparent and universally understandable for an early technical audience. It has no bootstrap script that executes network-fetched code and no package-index review latency. Its cost is a few documented commands and manual upgrades.
- The unsigned launch is useful but not frictionless: macOS Gatekeeper and Windows SmartScreen can warn about internet-downloaded unsigned software. Checksums and provenance prove release integrity/origin but do **not** substitute for platform code signing.

## Release contract

### Versions and tags

- Follow Semantic Versioning 2.0.0. Start at `v0.1.0`; while the CLI is `0.y.z`, increment minor for intentionally incompatible CLI/config/output changes and patch for backward-compatible fixes. SemVer states that `0.y.z` is initial development, normal versions have `MAJOR.MINOR.PATCH`, and released contents must not be modified. [Semantic Versioning](https://semver.org/)
- Stable tags are exactly `vMAJOR.MINOR.PATCH`; prereleases use `vMAJOR.MINOR.PATCH-rc.N` and are marked prerelease, never “latest.” Never retag or rebuild a published version.
- Archive filenames omit the tag's leading `v`; the program reports it (`v0.1.0`) so users can map directly to the Git tag.
- Generate release notes/changelog from commits since the previous tag, with an explicit breaking-changes section and installation/upgrade notes when behavior changes.

### Version injection and command

Provide `papercuts version` (and optionally `papercuts --version`) with stable, scriptable output:

```text
papercuts v0.1.0 (commit 1a2b3c4, built 2026-07-09T12:00:00Z)
```

- GoReleaser should inject `version`, full commit, and build date using `-ldflags -X`; its Go build pipeline exposes `.Version`, `.FullCommit`, and `.Date` templates and documents linker-flag version injection. [GoReleaser Go builds](https://goreleaser.com/customization/builds/go/)
- A local untagged build reports `papercuts devel` plus available commit metadata, never a misleading release version.
- For `go install ...@version`, fall back to `runtime/debug.ReadBuildInfo`: Go records the main module version in build information, and the API exposes `Main`, `Settings`, and VCS settings. This avoids falsely printing `devel` merely because `go install` did not receive GoReleaser linker flags. [Go `debug.ReadBuildInfo`](https://pkg.go.dev/runtime/debug#ReadBuildInfo)

### CI trigger and permissions

- Trigger the release workflow on a pushed tag matching `v*`, then fail immediately unless the tag strictly matches stable or approved prerelease SemVer. GitHub tag globs are not regexes, so the validation step is mandatory.
- Before release: checkout full history, set up the pinned Go version, run formatting checks, `go vet ./...`, `go test ./...`, and `goreleaser check`.
- Run GoReleaser with `--clean`, using the official `goreleaser/goreleaser-action`, and pin all third-party Actions to full commit SHAs. Use `fetch-depth: 0`, as in GoReleaser's first-party Actions guidance. [GoReleaser GitHub Actions](https://goreleaser.com/ci/actions/)
- Give the release job only `contents: write`, `id-token: write`, and `attestations: write`; do not use a personal access token. The repository-scoped `GITHUB_TOKEN` is enough for a public GitHub Release. Protect the environment/tag path so only reviewed commits are released.
- Create a draft release, upload all assets, attest them, smoke-test the downloaded assets, then publish the immutable release. A failed workflow leaves a draft, not a partially trusted public release.

## Install, upgrade, uninstall, and rollback

### Manual binary install — primary

Documentation should show all steps without `curl | sh`:

1. Select the archive from the GitHub Release for the user's OS/architecture.
2. Download the archive and `papercuts_0.1.0_checksums.txt`.
3. Verify SHA-256 (`sha256sum -c` / `shasum -a 256` on Unix, `Get-FileHash -Algorithm SHA256` on PowerShell), and optionally verify GitHub provenance/release integrity with `gh`.
4. Extract, run `./papercuts version` or `.\papercuts.exe version`, then move it into a directory already on user `PATH`.

Recommend a user-writable destination to avoid elevation: `~/.local/bin/papercuts` on Linux/macOS and a documented user bin directory on Windows. Mention `/usr/local/bin` only as an optional administrator-managed location. The executable has no project-runtime dependency.

- **Upgrade:** repeat verification and atomically replace the executable while it is not running. Never overwrite configuration or user data.
- **Uninstall:** delete the executable. Also list the exact optional config/cache paths, if the program creates any, and state that they are retained by default; users may delete them separately.
- **Rollback:** download and verify a named prior release, stop the running command, and replace the executable. Keep prior immutable releases available. For a bad release, mark it deprecated in release notes, stop calling it latest, and publish a new patch; do not mutate/reuse its tag or assets. Immutable release tags can be deleted only with the release and then cannot be reused, which reinforces “roll forward, preserve history.” [GitHub immutable releases](https://docs.github.com/en/code-security/concepts/supply-chain-security/immutable-releases)

### `go install` — secondary developer route

Document, after the binaries:

```sh
go install github.com/Whamp/papercuts/cmd/papercuts@latest
# or pin/rollback
go install github.com/Whamp/papercuts/cmd/papercuts@v0.1.0
```

(Use `github.com/Whamp/papercuts@...` instead if the command lives at module root.) A version suffix makes `go install` module-aware and installs the executable to `GOBIN`, or otherwise `$GOPATH/bin`. [Go command documentation](https://pkg.go.dev/cmd/go#hdr-Compile_and_install_packages_and_dependencies)

This route requires a Go toolchain **at install time**, may build with different toolchain/environment details than the attested release, and has no package-manager uninstall; uninstall is deleting the installed executable. It therefore does not meet the mainstream “no runtime/toolchain required” experience, but is useful for Go developers and exact-version rollback.

## Compared channels and deferral rationale

| Approach | Value | Initial decision |
|---|---|---|
| GitHub Releases + manual install | Universal standalone artifact, transparent versions, immutable assets, direct rollback | **Ship first; canonical** |
| GoReleaser | One declarative six-target build, OS-specific archive formats, SHA-256 list, linker metadata | **Ship first** |
| GitHub Actions attestations | Keyless provenance bound to repository/workflow; independently verifiable | **Ship first** |
| `go install` | Familiar to Go developers; easy pinning | **Document secondarily** |
| Shell/PowerShell installer | One-command OS/arch detection, download, verification, PATH setup | **Defer**: it adds remote-code bootstrap, shell variants, PATH mutation, error/rollback logic, and another asset to secure. When added, make scripts versioned/downloadable and readable, require checksum/provenance verification before replacement, and avoid `curl | sh` as the only documented path. |
| Homebrew | Excellent macOS/Linux upgrade/uninstall UX | **Defer until repeated demand**. A formula must carry stable URLs and SHA-256 values; `brew install`, `brew upgrade`, and `brew uninstall` then own lifecycle. Begin with a `Whamp/homebrew-tap`; seek `homebrew-core` only after the project meets its acceptance expectations. Homebrew requires checksums and discourages unstable artifacts. [Homebrew Formula Cookbook](https://docs.brew.sh/Formula-Cookbook) [Homebrew contribution process](https://docs.brew.sh/How-To-Open-a-Homebrew-Pull-Request) |
| Scoop | Lightweight Windows zip/portable workflow and straightforward manifests | **First deferred Windows channel** if Windows demand appears. Host a Whamp bucket initially; use manifest `architecture`, `url`, `hash`, and `bin`, and let Scoop own update/uninstall. Its official manifest docs define JSON manifests and architecture-specific downloads; autoupdate can detect versions and regenerate URLs/hashes. [Scoop manifests](https://github.com/ScoopInstaller/Scoop/wiki/App-Manifests) [Scoop autoupdate](https://github.com/ScoopInstaller/Scoop/wiki/App-Manifest-Autoupdate) |
| WinGet | Built-in broad Windows discovery, portable-package lifecycle | **Defer until identity/metadata and releases are stable**. Submission requires manifests and a pull request to `microsoft/winget-pkgs`; updates require new manifests/review. [WinGet manifests](https://learn.microsoft.com/en-us/windows/package-manager/package/manifest) [WinGet repository submission](https://learn.microsoft.com/en-us/windows/package-manager/package/repository) |
| In-process self-update | Potential `papercuts update` convenience | **Do not ship initially**. It duplicates package managers, must safely replace a running executable across three OSes, needs signature/checksum enforcement and failure recovery, and can conflict with read-only/admin/package-managed installs. `minio/selfupdate` supports checksum/signature validation and rollback on replacement failure, illustrating that this is security-sensitive lifecycle code rather than a free convenience. [minio/selfupdate](https://github.com/minio/selfupdate) |

When package channels are added, the docs must say: upgrade and uninstall with the same channel used to install. Never let `papercuts update` overwrite a Homebrew/Scoop/WinGet-managed binary.

## Platform trust and credentials

### macOS

Apple says software distributed outside the Mac App Store should be signed with a **Developer ID** certificate and notarized; Gatekeeper checks the Developer ID signature and notarization status. Notarization also requires hardened runtime and a secure timestamp. [Apple notarization](https://developer.apple.com/documentation/security/notarizing-macos-software-before-distribution) [Apple Developer ID](https://developer.apple.com/developer-id/)

- **Unsigned initial route:** viable for informed early CLI users via manual download plus checksum/attestation, but document that macOS may block or warn. Do not claim checksums bypass Gatekeeper and do not tell users to disable Gatekeeper globally. Link to Apple's user-facing safe-open guidance if an override is documented.
- **To remove friction later:** requires Apple Developer Program access, a Developer ID Application certificate/private key, notarization credentials (Apple ID/app-specific password or App Store Connect API key), and a macOS signing/notarization job using current `notarytool`. Sign each Mach-O before archiving, submit the supported archive/container, and verify with `codesign`, `spctl`, and notarization logs.

### Windows

Microsoft documents SmartScreen as reputation-based: reputation can attach to the file or a valid Authenticode signing certificate, and unsigned files cannot share publisher reputation across releases. Code signing proves integrity and publisher identity; it does not itself guarantee immediate reputation. [Microsoft SmartScreen reputation](https://learn.microsoft.com/en-us/windows/apps/package-and-deploy/smartscreen-reputation) [Microsoft code-signing options](https://learn.microsoft.com/en-us/windows/apps/package-and-deploy/code-signing-options)

- **Unsigned initial route:** viable for informed early users, but browsers/SmartScreen can warn. State this plainly and provide hashes/provenance; do not suggest disabling SmartScreen.
- **To remove friction later:** acquire an Authenticode certificate/private key or Microsoft Trusted Signing access/Azure credentials, sign and timestamp `papercuts.exe` before archiving, then verify with `Get-AuthenticodeSignature`/SignTool in CI. Preserve the same publisher identity across releases to build reputation.

No signing credential is required for the initial public GitHub/GoReleaser/attestation route. GitHub Actions OIDC creates provenance without a long-lived signing key. The required initial access is: the `Whamp` GitHub account, permission to create/configure the public repository, Actions enabled, immutable releases enabled, and permission to configure tag/ruleset/environment protections.

## Required documentation

Keep a single authoritative Install page/README section containing:

1. A prominent link to the latest stable release and the six-target support table.
2. OS/architecture detection commands and exact archive-selection examples.
3. Checksum verification first, provenance verification as a stronger optional step, and an explanation of what each proves.
4. Per-OS extract/install commands using user-writable `PATH` locations, followed by `papercuts version`.
5. Honest macOS Gatekeeper and Windows SmartScreen notes for unsigned builds.
6. Upgrade, exact-version rollback, and uninstall instructions for manual and `go install` routes.
7. A rule to use the originating package manager for future package-managed installs.
8. Release notes with supported platforms, checksums/provenance commands, known issues, and breaking changes.

## Validation and release acceptance

Before the first real tag:

1. `go test ./...` and `go vet ./...` pass; version-command tests cover linker-injected, Go build-info, and local `devel` cases.
2. `goreleaser check` passes and `goreleaser release --snapshot --clean` produces exactly six archives plus one SHA-256 file with the specified names and archive formats.
3. Inspect every archive: exactly one correctly named executable plus README/license; no source tree, secrets, or host-specific paths.
4. Recompute every SHA-256 independently and compare with the published list.
5. On native GitHub runners (`ubuntu`, `macos`, `windows`) for each available architecture, extract and execute `papercuts version` plus a non-destructive smoke command. Because cross-compiled arm64 executables cannot be assumed runnable on amd64 hosted runners, test arm64 on a real/self-hosted arm64 machine or explicitly record that execution coverage gap; at minimum inspect binary format and test cross-compilation.
6. On a release candidate, download assets from GitHub rather than testing only `dist/`; verify `gh release verify TAG`, `gh release verify-asset TAG FILE`, and `gh attestation verify FILE -R Whamp/papercuts`. [GitHub release integrity](https://docs.github.com/en/code-security/how-tos/secure-your-supply-chain/secure-your-dependencies/verify-release-integrity) [GitHub artifact attestations](https://docs.github.com/actions/security-for-github-actions/using-artifact-attestations/using-artifact-attestations-to-establish-provenance-for-builds)
7. Test clean install → version → upgrade → rollback → uninstall on one clean VM per OS. Confirm no Go installation is needed for release archives and no files remain except intentionally retained user config/cache.
8. For future signing, add clean-machine Gatekeeper and SmartScreen checks; do not infer platform trust merely from a successful signature command.

## Findings

1. **The smallest dependable release is a set of immutable, attested GitHub assets, not an installer ecosystem.** GitHub provides both immutable release attestations and explicit local-asset verification, while GoReleaser supplies deterministic naming, formats, checksums, and build metadata. [GitHub immutable releases](https://docs.github.com/en/code-security/concepts/supply-chain-security/immutable-releases) [GoReleaser archives](https://goreleaser.com/customization/archive/)
2. **Checksums, build provenance, and platform signing solve different problems.** SHA-256 detects changed bytes; GitHub attestations bind artifacts to a repository/workflow; Developer ID/Authenticode establish platform-recognized publisher identity and influence Gatekeeper/SmartScreen. None should be presented as a substitute for another. [GitHub artifact attestations](https://docs.github.com/actions/security-for-github-actions/using-artifact-attestations/using-artifact-attestations-to-establish-provenance-for-builds) [Apple Developer ID](https://developer.apple.com/developer-id/) [Microsoft SmartScreen reputation](https://learn.microsoft.com/en-us/windows/apps/package-and-deploy/smartscreen-reputation)
3. **Package managers improve lifecycle UX but multiply release maintenance.** Homebrew, Scoop, and WinGet each require channel-specific manifests/checksums/submission/update automation; they should consume the canonical GitHub assets after naming and cadence stabilize, not define the initial release. [Homebrew Formula Cookbook](https://docs.brew.sh/Formula-Cookbook) [Scoop manifests](https://github.com/ScoopInstaller/Scoop/wiki/App-Manifests) [WinGet manifests](https://learn.microsoft.com/en-us/windows/package-manager/package/manifest)
4. **Roll forward operationally while preserving exact rollback artifacts.** Immutable releases make previous bytes auditable and downloadable. A bad version should receive a superseding patch, while manual users can replace with a verified prior archive and Go users can install an exact module version. [GitHub immutable releases](https://docs.github.com/en/code-security/concepts/supply-chain-security/immutable-releases) [Go install](https://pkg.go.dev/cmd/go#hdr-Compile_and_install_packages_and_dependencies)

## Sources

### Kept

- [GitHub: Immutable releases](https://docs.github.com/en/code-security/concepts/supply-chain-security/immutable-releases) — canonical immutability, automatic release attestation, and draft-first publishing behavior.
- [GitHub: Artifact attestations](https://docs.github.com/actions/security-for-github-actions/using-artifact-attestations/using-artifact-attestations-to-establish-provenance-for-builds) — canonical permissions, Action, and verification command.
- [GitHub: Verify release integrity](https://docs.github.com/en/code-security/how-tos/secure-your-supply-chain/secure-your-dependencies/verify-release-integrity) — canonical immutable-release and local-asset verification.
- [GoReleaser: Go builds](https://goreleaser.com/customization/builds/go/), [archives](https://goreleaser.com/customization/archive/), [checksums](https://goreleaser.com/customization/checksum/), and [GitHub Actions](https://goreleaser.com/ci/actions/) — first-party behavior for version injection, packaging, hashes, and CI.
- [Go: supported systems](https://go.dev/doc/install/source#environment), [`go install`](https://pkg.go.dev/cmd/go#hdr-Compile_and_install_packages_and_dependencies), and [`ReadBuildInfo`](https://pkg.go.dev/runtime/debug#ReadBuildInfo) — authoritative target/toolchain/version behavior.
- [Semantic Versioning 2.0.0](https://semver.org/) — version/tag semantics.
- [Homebrew Formula Cookbook](https://docs.brew.sh/Formula-Cookbook) and [contribution process](https://docs.brew.sh/How-To-Open-a-Homebrew-Pull-Request) — first-party formula/checksum/channel maintenance.
- [Scoop app manifests](https://github.com/ScoopInstaller/Scoop/wiki/App-Manifests) and [autoupdate](https://github.com/ScoopInstaller/Scoop/wiki/App-Manifest-Autoupdate) — project-owned package schema and update automation.
- [Microsoft WinGet manifests](https://learn.microsoft.com/en-us/windows/package-manager/package/manifest) and [repository submission](https://learn.microsoft.com/en-us/windows/package-manager/package/repository) — first-party publishing requirements.
- [Apple notarization](https://developer.apple.com/documentation/security/notarizing-macos-software-before-distribution) and [Developer ID](https://developer.apple.com/developer-id/) — platform authority for Gatekeeper-facing distribution.
- [Microsoft SmartScreen reputation](https://learn.microsoft.com/en-us/windows/apps/package-and-deploy/smartscreen-reputation) and [code-signing options](https://learn.microsoft.com/en-us/windows/apps/package-and-deploy/code-signing-options) — platform authority for Windows trust.
- [minio/selfupdate](https://github.com/minio/selfupdate) — library owner's documentation of validation and rollback complexity.

### Dropped

- Third-party packaging tutorials and installer generators — excluded because the ticket requires primary first-party sources and they add no authoritative behavior.
- Community Q&A about SmartScreen reputation thresholds — excluded because Microsoft does not promise a fixed download/time threshold.
- Forks and package-index mirrors of self-update libraries — excluded in favor of the original project repository.
- Search-result summaries — used only to locate documentation; no recommendation relies on them as evidence.

## Gaps

- The final Go module/command layout is not yet present, so the exact `go install` import path must match the implemented location (`/cmd/papercuts` versus module root).
- GitHub Actions cannot be exercised and immutable releases cannot be enabled until `Whamp/papercuts` exists remotely with an initial commit.
- Hosted arm64 runner availability and real Gatekeeper/SmartScreen behavior must be tested at implementation time; cross-compilation alone is not execution validation.
- Apple and Windows signing are intentionally deferred. Their exact CI secret setup depends on credentials not currently available, but this does not block the useful unsigned initial release path.

```acceptance-report
{
  "criteriaSatisfied": [
    {
      "id": "criterion-1",
      "status": "satisfied",
      "evidence": "Created only the authoritative research brief at /tmp/papercuts-distribution-agent.md; no repository files were edited. The brief compares every requested channel/trust/update topic and makes an initial-versus-deferred decision."
    },
    {
      "id": "criterion-2",
      "status": "satisfied",
      "evidence": "The brief contains inline citations to primary GitHub, Go, GoReleaser, Apple, Microsoft, Homebrew, Scoop, SemVer, and original library documentation, plus an explicit validation checklist, gaps, and source disposition."
    }
  ],
  "changedFiles": [
    "/tmp/papercuts-distribution-agent.md"
  ],
  "testsAddedOrUpdated": [],
  "commandsRun": [
    {
      "command": "web_search: four-angle primary-source research (GitHub/GoReleaser/Go, package managers, platform signing, updates/rollback)",
      "result": "passed",
      "summary": "Located first-party documentation for all requested comparison areas."
    },
    {
      "command": "fetch_content/get_search_content: inspect selected first-party sources",
      "result": "passed",
      "summary": "Read authoritative details for attestation permissions, immutable-release behavior, formats/checksums, platform trust, package manifests, and verification."
    },
    {
      "command": "write /tmp/papercuts-distribution-agent.md",
      "result": "passed",
      "summary": "Wrote the sole requested research artifact to the runtime-authoritative path."
    }
  ],
  "validationOutput": [
    "All requested topics are addressed: installation, release, upgrade, uninstall, GitHub Releases/Actions attestations, GoReleaser, go install, manual install, scripts, Homebrew, Scoop, WinGet, macOS/Windows signing, self-update, and rollback.",
    "Artifact contract specifies six OS/architecture targets, names, formats, SHA-256, attestations, tags, CI trigger, version output, lifecycle docs, and release validation.",
    "Only primary first-party sources are retained; third-party/community material is explicitly dropped."
  ],
  "residualRisks": [
    "No remote repository exists yet, so the proposed Actions/Release flow remains a researched design rather than an executed release.",
    "Unsigned macOS and Windows binaries can trigger platform warnings until signing credentials are obtained.",
    "Arm64 binaries require native execution validation at implementation time."
  ],
  "noStagedFiles": true,
  "diffSummary": "Added one Markdown research brief outside the repository; no project source or configuration files changed.",
  "reviewFindings": [
    "no blockers"
  ],
  "manualNotes": "The runtime path override was honored instead of the ticket's repository-local destination. Repository staging state was not modified; the only write was under /tmp."
}
```
