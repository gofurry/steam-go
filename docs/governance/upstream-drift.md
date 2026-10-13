# Steam Upstream Drift Ledger

This is the canonical record of confirmed Steam upstream drift that required a
compatibility fix. It records observations and SDK decisions, not guarantees
about Valve's intent or the continuing behavior of a live endpoint.

## Recording Rules

- Follow the [upstream contract](../../contracts/upstream.md).
- Record the observation date, surface, confirmed behavior, unknown cause,
  impact, compatibility decision, and deterministic regression protection.
- Attribute observations to their evidence. Do not turn an inferred cause or
  relationship into a confirmed fact.
- Keep payload fixtures small and sanitized. Never store credentials, cookies,
  proxy userinfo, or local configuration in incident records.
- Keep regression fixtures when live behavior later changes or recovers.
- Live Steam validation remains opt-in and is not a normal CI requirement.

## 2026-09-26 — Storefront AppDetails envelope identity drift

### Surface

`store.steampowered.com/api/appdetails`, exposed by
`client.Web.Storefront.GetAppDetails` and `GetAppDetailsRaw`.

### Confirmed

The maintainer's incident report dated 2026-09-26 recorded these responses:

| Requested AppID | Response envelope key | `data.steam_appid` | Application |
|---:|---|---:|---|
| 550 | `322070` | 550 | Left 4 Dead 2 |
| 620 | `323180` | 620 | Portal 2 |
| 1206410 | `4602510` | 1206410 | Kristala / 月影杀 |
| 10 | `10` | 10 | Counter-Strike, normal-key control |

- The first three key mismatches were observed with both `cc=CN&l=schinese`
  and `cc=US&l=english`.
- HTTP returned 200; raw JSON was valid and typed `AppDetailsEnvelope` decoding
  succeeded. The internal AppID still identified the requested application.
- At least AppID 550 reproduced outside steam-go using an ordinary Go HTTP
  client with its explicit HTTP proxy removed.

These are dated observations from the incident report, not a claim that every
current response reproduces the drift.

### Recovery observation — 2026-10-01

Steam Support reported that they could no longer reproduce the mismatch and
observed response keys matching the requested AppIDs. The maintainer then
re-tested AppIDs 550, 620, and 1206410 independently and observed the normal
shape again:

| Requested AppID | Response envelope key | `data.steam_appid` |
|---:|---|---:|
| 550 | `550` | 550 |
| 620 | `620` | 620 |
| 1206410 | `1206410` | 1206410 |

The incident is therefore currently considered recovered and likely transient
or limited in scope. One plausible but unverified explanation is temporary drift
in the Store backend product-ID graph or its envelope serialization: two
historical unexpected keys, `322070` and `323180`, are related applications
of the requested base apps. This remains a diagnostic hypothesis, not a
confirmed Valve root cause.

Regression coverage remains in place even though the current live behavior has
returned to the legacy key shape.

### Unknown / unverified

- Whether Valve intentionally changed the contract or introduced a regression.
- When the behavior began and its worldwide scope.
- Whether DLC, associated applications, caching, or another mechanism caused it.
- Valve's repair plans, if any.

No formal Valve contract change or application relationship has been confirmed.

### Impact

Consumers that assumed `envelope[requestedAppID]` could miss valid app data.
Storefront media discovery, free-package resolution, Doctor, and live examples
or smoke checks could then fail despite successful fetching and decoding.
Neither raw fetching nor typed JSON decoding was the underlying failure.

### Compatibility decision

The v1.3.12 hotfix is prepared under `Unreleased`:

- `GetAppDetailsRaw` preserves the response bytes unchanged.
- `GetAppDetails` preserves Steam's original typed envelope keys.
- `ResolveAppDetails` uses `data.steam_appid` as the requested identity anchor;
  `AppDetailsMatch.ResponseKey` preserves the original string key verbatim.
- `GetResolvedAppDetails` composes fetching and the same resolver for ordinary
  single-application consumption.
- No match, duplicate identity, unsuccessful identity matches, and invalid
  direct entries fail closed with existing error kinds. An alternate valid
  entry never overrides a direct failure, missing identity, or conflict.
- No first-entry fallback or DLC/parent/child inference is used.
- `GetAppDetailsBatch` keeps its existing structure and fetch-error semantics;
  callers explicitly resolve each successful item's envelope.
- Doctor reports safely resolved key drift as WARN, without increasing Fail or
  changing a successful exit code. Live smoke validates resolved identity and
  accepts either the legacy key or a drifted key.

### Regression protection

- [Minimal key-drift fixture](../../testdata/fixtures/web/storefront/GetAppDetails/app_550_key_drift.json)
  records only the `322070` → internal AppID 550 compatibility case.
- [Resolver and fidelity tests](../../web/storefront/appdetails_test.go) cover
  the complete identity boundary matrix, raw byte fidelity, original typed keys,
  unchanged batch semantics, and error propagation.
- [Assets tests](../../addons/assets/store_media_test.go) and
  [freeclaim tests](../../addons/freeclaim/package_test.go) protect consumer
  migration, including invalid direct entries alongside valid alternatives.
- [Doctor tests](../../examples/doctor/main_test.go) protect OK/WARN/FAIL,
  summary counts, and exit behavior. [Offline smoke tests](../../examples/live/appdetails_test.go)
  protect resolved identity without making live Steam access mandatory.

## 2026-09-04 — StoreBrowse assets.last_modified type drift

### Surface

`IStoreBrowseService/GetItems/v1`, specifically `assets.last_modified`.

### Confirmed

The v1.3.10 release record documents Steam returning this metadata as a JSON
number where the SDK's asset map expected string values. The existing
[regression tests](../../api/storebrowseservice/types_test.go) demonstrate the
compatibility behavior with synthetic data; their numeric values are not
presented as exact historical incident samples.

### Unknown / unverified

The exact introduction time, worldwide scope, and Valve's cause or intent were
not recorded. This ledger does not infer them retrospectively.

### Impact

Decoding asset metadata as a string map could fail and block otherwise valid
StoreBrowse asset discovery.

### Compatibility decision

Released in v1.3.10: defensive decoding retains string-valued asset metadata and
safely ignores non-string fields. The string-based `StoreItemAssets` public API
and asset URL discovery behavior remain unchanged. Raw fetching remains
available when callers need the complete upstream representation.

### Regression protection

`TestStoreItemAssetsUnmarshalJSON` covers numeric `last_modified` alongside
valid string filenames. Additional tests cover other non-string metadata and
malformed JSON; [asset discovery tests](../../addons/assets/store_item_assets_test.go)
exercise the consumer pipeline.
