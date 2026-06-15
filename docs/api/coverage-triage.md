# Steam Web API Coverage Triage

This is the first manual triage pass for `docs/api/coverage-diff.md`.

It is intentionally maintained by humans. Do not generate this file from `steamapi-sync`; use it to decide which official endpoints are worth implementing after reviewing stability, authentication boundaries, abuse risk, testability, and general usefulness.

Current tracked counts:

- `covered=27`
- `extra_sdk=74`
- `missing=34`
- `version_mismatch=2`

## Triage Rules

- Prefer official, read-only, broadly useful endpoints.
- Prefer endpoints with clear authentication requirements and fixture-testable payloads.
- Prefer stable outer structs with `json.RawMessage` for volatile nested payloads.
- Do not chase coverage percentage for its own sake.
- Do not add purchase, sale, trade, login automation, bulk account automation, or publisher-sensitive workflows.
- Treat `version_mismatch` as a manual compatibility question before adding or changing code.

## P1 Candidates

These are reasonable candidates for future endpoint work after `v1.3.4` stabilization if users need more official endpoint coverage.

| Endpoint | Why it is interesting | Initial boundary |
|---|---|---|
| `ISteamUserStats/GetGlobalStatsForGame/v1` | General public stats endpoint, read-only, adjacent to existing `steamuserstats` coverage. | Add fixture-backed typed outer response plus raw method if payload shape is not fully stable. |
| `ISteamDirectory/GetCMList/v1` | Read-only directory endpoint, adjacent to existing `ISteamDirectory/GetCMListForConnect/v1`. | Confirm parameter behavior and whether existing directory service should expose both variants. |
| `ISteamRemoteStorage/GetPublishedFileDetails/v1` | Broad Workshop/public file lookup use case, common in Steam tooling. | Keep read-only; avoid mutating UGC or voting workflows. |
| `ISteamRemoteStorage/GetCollectionDetails/v1` | Companion to published file details, useful for Workshop collection metadata. | Keep fixture coverage small and avoid assuming game-specific nested schema. |

## P2 Candidates

These may be useful, but need more context before they should become implementation work.

| Endpoint | Why it may matter | Reason not P1 |
|---|---|---|
| `IStoreService/GetRecommendedTagsForUser/v1` | Store tagging information can be useful for discovery tooling. | Needs review of regional behavior and whether it depends on user personalization. |
| `IPublishedFileService/GetUserVoteSummary/v1` | Could help read-only Workshop metadata workflows. | User-specific semantics need authentication review. |
| `ISteamUserOAuth/GetTokenDetails/v1` | Useful for token diagnostics. | Requires access token handling and stricter credential-safety examples. |

## Deferred

These are official gaps, but they are not good near-term expansion targets.

| Group | Examples | Reason |
|---|---|---|
| Broadcast/session/stat reporting endpoints | `IBroadcastService/PostGameDataFrameRTMP`, `ISteamBroadcast/PlayerStats`, `ViewerHeartbeat`, `IClientStats_1046930/ReportEvent` | Session/token-sensitive or client telemetry/reporting behavior; not aligned with the stable core SDK boundary. |
| Game notification endpoints | `IGameNotificationsService/UserCreateSession`, `UserDeleteSession`, `UserUpdateSession` | Session-like behavior and user interaction state; defer until a clear read-only use case exists. |
| Help request log endpoints | `IHelpRequestLogsService/GetApplicationLogDemand`, `UploadUserApplicationLog` | Upload/log workflows can expose local or user data; not suitable for a quick coverage addition. |
| Game-specific GC/version endpoints | `IGCVersion_*`, `ITFSystem_440`, `IPortal2Leaderboards_620` | Narrow game-specific coverage; not a good general SDK priority. |
| Narrow content-server picker | `IContentServerDirectoryService/PickSingleContentServer` | The broader `GetServersForSteamPipe` helper now covers the main read-only directory use case; defer this until a user needs the exact single-server picker semantics. |

## Implemented Low-level Boundary

These were implemented after manual safety review, but remain low-level helpers rather than productized workflows.

| Group | Endpoints | Boundary |
|---|---|---|
| Authentication session helpers | `IAuthenticationService/GetAuthSessionInfo`, `GetAuthSessionRiskInfo`, `NotifyRiskQuizResults`, `UpdateAuthSessionWithMobileConfirmation` | Low-level typed/raw API coverage only. The SDK does not complete login flows, store user passwords, bypass Steam Guard, or answer risk checks automatically. |
| Content server directory helpers | `IContentServerDirectoryService/GetCDNForVideo`, `GetClientUpdateHosts`, `GetDepotPatchInfo`, `GetServersForSteamPipe` | Low-level read-only directory metadata only. The SDK does not become a CDN downloader, depot patcher, or SteamPipe client. |

## Version Mismatch Review

These require manual compatibility review before code changes.

| Endpoint | SDK path | Current decision |
|---|---|---|
| `ISteamNews/GetNewsForApp/v1` | `/ISteamNews/GetNewsForApp/v2/` | Keep current v2 coverage unless users need v1 specifically. Do not downgrade existing behavior. |
| `ISteamUserStats/GetGlobalAchievementPercentagesForApp/v1` | `/ISteamUserStats/GetGlobalAchievementPercentagesForApp/v2/` | Keep current v2 coverage unless upstream v1 provides materially different behavior. |

## Extra SDK Review

`extra_sdk` means the SDK exposes endpoints not present in the current public `GetSupportedAPIList` snapshot. This can happen because Steam's public inventory is incomplete, regional, or has changed.

Do not remove extra SDK endpoints just because they are absent from the snapshot. Review them only when:

- live smoke or fixture tests show a real failure;
- docs claim an endpoint is official but upstream no longer lists or serves it;
- an endpoint has unclear credentials or mutating behavior;
- users report compatibility problems.

## Next Actions

- Use P1/P2 candidates as input for `v1.4` or later endpoint planning, not as automatic work.
- Reference this triage before adding new official endpoint coverage.
- Add fixtures before adding typed responses.
- Prefer one small endpoint group at a time.
- Keep generated reports and this triage file in sync during release review.
