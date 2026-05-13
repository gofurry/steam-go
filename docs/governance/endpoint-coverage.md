# Endpoint Coverage

This document lists the current official Steam Web API coverage that is already present in the repository as `steam-go` moves toward `v1.0.0`.

## Coverage policy before v1.0.0

Before `v1.0.0`, the project is not planning to expand API coverage further.

The release priority is:

- freeze and document the stable surface
- verify tests and examples
- complete compatibility and release documentation

Additional official API coverage is deferred to `v1.1.0+`.

## Current service groups

The repository currently exposes these grouped services under `client.API.*`:

- `AccountCartService`
- `BillingService`
- `CommunityService`
- `FamilyGroupsService`
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

## Current release interpretation

For `v1.0.0`, the important point is not to claim "complete Steam coverage".

The practical interpretation is:

- the current official service groups are already broad enough to justify a stable `v1.0.0`
- missing official endpoints do not block `v1.0.0`
- new official endpoints can be added compatibly in later `v1.x` releases

## Non-coverage areas

The following are intentionally not part of current endpoint coverage:

- new public Steam Store page fetch APIs
- Steam Community page scraping APIs
- CDN or static asset helper APIs
- undocumented web page JSON endpoints

If these appear later, they should be introduced deliberately and documented separately instead of being quietly mixed into the first stable release.
