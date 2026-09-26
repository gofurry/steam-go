# Endpoint 覆盖范围

本文档列出 `steam-go` v1.x 当前公开的 Steam Web API 服务分组，其中包括有官方文档的 API 和 observed Steam 服务；两者的区别见[上游契约](../../../contracts/upstream.md)。

## 覆盖策略

v1.x 可以兼容地新增 endpoint，但应先审查上游状态、认证要求、职责范围和可测试性。

Endpoint 级别的信息以以下文档为入口：

- [API 参考](../api/reference.md)：服务用法和边界
- [Generated 覆盖报告](../../api/coverage.generated.md)：已记录的上游 inventory 与 SDK 对比
- [Coverage triage](../../api/coverage-triage.md)：缺口和 drift 的处理决策

## 当前服务分组

当前仓库在 `client.API.*` 下暴露这些服务分组：

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

## 覆盖审查

SDK 不宣称完整覆盖 Steam。审查时应遵循：

- 根据用户需要和项目边界评估缺失的 endpoint
- inventory 中存在缺口本身不构成发布阻塞
- 新增 endpoint 必须保持公开兼容性契约，并记录上游状态

`GameServersService.GetServerList` 会被文档化为服务器发现接口。如果 generated coverage 将它标记为 `extra_sdk`，这是该实用 endpoint 没有稳定出现在公开 inventory 中的预期 drift，不是自动删除信号。

## 只读 Web surface

仓库通过独立的 `client.Web.*` 层提供限定范围的只读 Storefront、Community 与 Market JSON 接口。

这些入口单独记录在 [Web 参考](../web/reference.md)中，以保证 `client.API.*` 的边界清晰。

## 非覆盖范围

以下内容不属于当前 endpoint 覆盖范围：

- 未来新增的、未文档化的公开 Steam Store 页面抓取 API
- 超出当前 `client.Web.*` 已记录 JSON 接口范围的 Steam Community 页面抓取 API
- CDN 或静态资源 helper API
- 超出当前 `client.Web.*` 已记录方法范围的未文档化网页 JSON endpoint

`addons/assets` 已提供的资源 helper 记录在 [addon 参考](../addons/reference.md)中。更广泛的 Web 扩展应单独确定范围和编写文档，与 Web API endpoint 覆盖区分。
