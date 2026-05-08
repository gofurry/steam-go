# `v1.0.0-alpha-1` Release Notes

`v1.0.0-alpha-1` is the first alpha milestone for `steam-go`.

This release is focused on establishing a stable base Web API client, clarifying credential behavior, and shipping a broader first batch of typed Steam Web API coverage.

## Highlights

- unified root client with grouped service access under `client.API.*`
- functional options for API key, access token, timeout, retry, rate limit, and proxy selection
- typed responses by default, with raw response variants for the same endpoints
- addon support for `a2s` and `openid` without bloating the core client
- grouped real smoke-test entrypoints under `test/<service>`

## Included API Groups

- `AccountCartService`
- `BillingService`
- `CommunityService`
- `FamilyGroupsService`
- `LoyaltyRewardsService`
- `MobileNotificationService`
- `NewsService`
- `PlayerService`
- `SteamNews`
- `SteamUser`
- `SteamUserStats`

## PlayerService Coverage

This alpha includes a relatively broad first pass of `PlayerService`, including:

- owned games, recently played games, last played times
- badges, community badge progress, favorite badge, Steam level, Steam level distribution
- achievements progress and top achievements for games
- animated avatars, avatar frames, profile backgrounds, mini-profile backgrounds
- profile items equipped, profile items owned, profile customization, profile themes
- purchased profile customizations and purchased/upgraded customization summaries
- nickname list, player link details, and friends gameplay info

## Credential Behavior

Credential handling is now more explicit:

- `key` and `access_token` are treated as different credentials
- when a method explicitly accepts `key` or `accessToken`, the caller should pass the user-specific credential to that method
- client-level credentials remain useful as defaults for endpoints that do not require explicit per-method credentials

This makes it easier to distinguish between:

- developer credentials used for public/shared queries
- caller credentials required for user-specific private data

## Proxy and Network Support

- `WithProxySelector(...)` remains the main proxy extension point
- `NewStaticProxySelector(...)`, `NewRoundRobinProxySelector(...)`, and `NewRoutingProxySelector(...)` are available
- `NewHTTPClientWithProxySelector(...)` can be reused by addons or standalone HTTP flows

## Testing and Examples

- examples are available for `a2s`, `openid`, `proxy`, `steamuser`, `steamnews`, and more
- real smoke tests are now split by service, such as `test/steamuser` and `test/playerservice`
- the SDK test suite passes with `go test ./...`

## Alpha Notes

This is still an alpha release:

- API coverage is already useful, but not complete
- compatibility and ergonomics may continue to improve before a stable `v1`
- method signatures and typed models may still receive small adjustments as more endpoints are added

## Suggested Release Summary

> First public alpha of `steam-go`: a lightweight Steam Web API SDK for Go with grouped services, typed responses, explicit credential handling, proxy support, addons for `a2s` and `openid`, and an expanded first batch of `PlayerService` coverage.
