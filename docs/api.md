# API Reference

This directory keeps reference-style documentation that is too detailed for the repository homepage.

## API Groups

- `client.API.AccountCartService`
- `client.API.BillingService`
- `client.API.CommunityService`
- `client.API.FamilyGroupsService`
- `client.API.LoyaltyRewardsService`
- `client.API.PlayerService`
- `client.API.SteamNews`
- `client.API.SteamUser`
- `client.API.SteamUserStats`

## Addons

- `addons/a2s`
- `addons/a2s/master`
- `addons/a2s/scanner`
- `addons/openid`
- [Addon usage notes](addons.md)

## Proxy Helpers

- `WithProxySelector(selector ProxySelector)`
- `NewStaticProxySelector(rawURL string)`
- `NewRoundRobinProxySelector(rawURLs ...string)`
- `NewRoutingProxySelector(routes ...ProxyRoute)`
- `NewHTTPClientWithProxySelector(selector ProxySelector, timeout time.Duration)`

## Examples

- `go run ./examples/a2s -server 1.2.3.4:27015 -query info`
- `go run ./examples/a2s -server 1.2.3.4:27015 -query players`
- `go run ./examples/a2s -server 1.2.3.4:27015 -query rules`
- `go run ./examples/openid`
- `go run ./examples/openid --proxy http://127.0.0.1:7897`
- `go run ./examples/proxy`
