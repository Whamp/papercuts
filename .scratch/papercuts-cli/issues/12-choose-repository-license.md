# Choose the Repository License

Type: grilling
Status: resolved
Blocked by: 09

## Question

Which license should govern the public `github.com/Whamp/papercuts` source and release archives?

## Context

The distribution contract requires a `LICENSE` file in each immutable release archive. This is a legal and product-direction decision, so implementation cannot choose it autonomously.

`Whamp/papercuts` now exists privately so hosted Linux, macOS, and Windows validation can run without publishing unlicensed source. It must remain private until this issue records the license choice and the exact `LICENSE` text is committed.

Will’s public GitHub repositories predominantly use MIT when they declare a license. Recommendation: **MIT**, matching projects such as `pi-codex-goal`, `pi-observational-memory`, `skills`, `napkin`, `pi-annotate`, and `pi-web-access`.

Answer with one of:

- `MIT` (recommended). This authorizes the canonical [SPDX MIT text](https://spdx.org/licenses/MIT.txt) with `Copyright (c) 2026 Will Hampson`, matching the public name on GitHub account `Whamp`.
- `Apache-2.0`
- another exact license identifier/text
- `all rights reserved` (public source with no open-source grant)

Specify a different copyright holder or year with the answer when desired; otherwise the `MIT` meaning above is complete and requires no follow-up.

## Decision History

Before Will answered, all independent work was complete through commit `b0b4f70`, the repository remained private without a root `LICENSE`, and no raw human session message authorized a license. That evidence justified pausing for this legal/product decision instead of inferring consent from the MIT recommendation.

## Answer

Will selected `MIT`. The repository and every release archive use the canonical SPDX MIT text with `Copyright (c) 2026 Will Hampson`; the exact text is committed at root [`LICENSE`](../../../LICENSE). No licensing question remains open.
