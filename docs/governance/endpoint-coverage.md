# Endpoint Coverage

This document lists the current Steam Web API service groups exposed by `steam-go` v1.x. They include both documented official APIs and observed Steam services; see [the upstream contract](../../contracts/upstream.md) for the distinction.

## Coverage policy

New endpoints may be added compatibly in v1.x after reviewing upstream status, authentication, scope, and testability.

Use these sources for endpoint-level detail:

- [API reference](../api/reference.md) for service usage and boundaries
- [Generated coverage](../api/coverage.generated.md) for the tracked upstream inventory and SDK comparison
- [Coverage triage](../api/coverage-triage.md) for decisions about gaps and drift

## Current service groups

The repository currently exposes these grouped services under `client.API.*`:

- `AccountCartService`
- `AuthenticationService`
- `BillingService`
- `CommunityService`
- `ContentServerDirectoryService`
- `FamilyGroupsService`
- `GameServersService`
- `LoyaltyRewardsService`
- `MobileNotificationService`
- `NewsService`
- `PlayerService`
- `QuestService`
- `SaleFeatureService`
- `SteamApps`
- `SteamChartsService`
- `SteamDirectory`
- `SteamNews`
- `SteamNotificationService`
- `SteamUser`
- `SteamUserOAuth`
- `SteamUserStats`
- `SteamWebAPIUtil`
- `StoreBrowseService`
- `StoreCatalogService`
- `StorePreferencesService`
- `StoreService`
- `StoreTopSellersService`
- `UserAccountService`
- `UserReviewsService`
- `UserStoreVisitService`
- `WishlistService`

## Coverage review

The SDK does not claim complete Steam coverage.

The practical interpretation is:

- missing endpoints are reviewed according to user needs and the project boundary
- an inventory gap alone is not a release blocker
- new endpoints must preserve the public compatibility contract and document their upstream status

`GameServersService.GetServerList` is intentionally documented as server discovery. If the generated coverage reports mark it as `extra_sdk`, treat that as expected drift for this useful endpoint rather than an automatic removal signal.

## Read-only web surfaces

The repository exposes a separate `client.Web.*` layer for scoped read-only Storefront, Community, and Market JSON endpoints.

These entrypoints are documented separately in the [Web reference](../web/reference.md) so the `client.API.*` boundary remains clear.

## Non-coverage areas

The following are intentionally not part of current endpoint coverage:

- future undocumented or broader Steam Store page fetch APIs
- future Steam Community page scraping APIs beyond the documented `client.Web.*` JSON endpoints
- CDN or static asset helper APIs
- undocumented or unstable web page JSON endpoints beyond the documented `client.Web.*` methods

Resource helpers already provided by `addons/assets` are documented in the [addon reference](../addons/reference.md). Any broader Web expansion should be scoped and documented separately from Web API endpoint coverage.
