# steam-go 工具包审计与改进路线总结

> 适用仓库：`github.com/GoFurry/steam-go`  
> 文档目的：把本轮审计中发现的改进点整理成可执行的工程清单，帮助 `steam-go` 从“可用 SDK”继续打磨为“值得他人依赖的开源 SDK”。

---

## 1. 当前总体评价

`steam-go` 目前已经具备一个正式 Go SDK 的基本形态：

- 使用根 `Client` 作为统一入口；
- 通过 `client.API.*` 暴露不同 Steam Web API 服务；
- 使用 Functional Options 配置 API key、access token、timeout、retry、rate limit、proxy；
- 区分 typed response 与 raw response；
- 将 A2S、OpenID 等能力放入 addons，避免核心 SDK 膨胀；
- 具备 README、中文文档、API 文档、CI、单元测试、示例程序等开源项目基础设施。

整体上，它已经不是一个简单练手仓库，而是一个可以继续正规化发布的轻量 Steam Web API SDK。

当前建议评分：**8 / 10**

扣分点主要不是功能不够，而是：

- public API 还需要进一步稳定；
- endpoint 稳定性需要标注；
- CI 检查还偏轻；
- rate limit/retry 能力还比较基础；
- live example 和 test 目录边界需要整理；
- 错误体与日志安全需要进一步打磨。

---

## 2. 当前优点

### 2.1 SDK 入口清晰

当前 `steam.NewClient(...)` + `client.API.*` 的设计是正确方向。

相比把 Store、Crawler、Server、Util 等能力混在一起，现在的结构更像一个真正的 SDK：

```go
client, err := steam.NewClient(
    steam.WithTimeout(10*time.Second),
    steam.WithRetry(2),
)

resp, err := client.API.SteamUser.GetPlayerSummaries(
    context.Background(),
    []string{"76561198370695025"},
)
```

建议继续保持这种结构，不要把无关功能重新塞回核心 Client。

---

### 2.2 认证模型比较专业

当前已经区分了：

- API key；
- access token；
- 静态凭证；
- 多凭证轮询；
- 自定义 provider；
- 方法级显式 key/accessToken；
- client-level 默认凭证。

这是 Steam SDK 很重要的一点。

建议后续文档进一步讲清楚：

- API Key 适合哪些接口；
- Access Token 适合哪些接口；
- OpenID 只用于确认用户 SteamID64，不等于 Web API 凭证；
- Publisher Key / 用户 Token / Public API Key 的边界。

---

### 2.3 addon 分层是正确的

当前 addons 包含：

- `addons/a2s`
- `addons/a2s/master`
- `addons/a2s/scanner`
- `addons/openid`

这个分层比较健康。

建议继续保持：

- 核心包只做 Steam Web API；
- OpenID 作为身份登录 addon；
- A2S 作为游戏服务器查询 addon；
- 不在核心 SDK 内部重复实现 A2S 逻辑。

---

### 2.4 错误模型已经有 SDK 风格

当前的错误分层：

```go
request_build
transport
http_status
decode
api_response
```

是比较好的设计。

这使用户可以：

```go
var apiErr *steam.APIError
if errors.As(err, &apiErr) {
    switch apiErr.Kind {
    case steam.ErrorKindHTTPStatus:
        // handle HTTP status
    case steam.ErrorKindDecode:
        // handle decode error
    }
}
```

建议保留这个方向，并继续强化错误文档。

---

### 2.5 代理能力保持克制

当前代理能力包括：

- static proxy；
- round-robin proxy；
- routing proxy；
- addon/standalone HTTP client proxy helper。

这是适合轻量 SDK 的方案。

建议不要默认加入复杂代理池、健康检查、熔断、IP 质量检测等能力，否则会让 SDK 变重。

---

## 3. 主要问题与风险

## P0：优先改进项

### 3.1 Rate Limit API 表达能力不足

当前 `WithRateLimit(requestsPerSecond int)` 只能表达“每秒多少请求”。

问题：

- Steam API 场景里经常需要表达每分钟、每小时、每 host、每 endpoint 的限速；
- 仅使用 RPS 对真实 Steam API 不够灵活；
- 当前 burst 默认等于 requestsPerSecond，用户不可控。

建议保留当前快捷方法，同时新增更灵活的配置：

```go
WithRateLimiter(limit rate.Limit, burst int)
WithMinInterval(d time.Duration)
WithPerHostRateLimit(map[string]RateLimitConfig)
```

可以设计为：

```go
type RateLimitConfig struct {
    Limit rate.Limit
    Burst int
}
```

短期最推荐：

```go
WithRateLimiter(limit rate.Limit, burst int)
```

这样既保持轻量，又不会把限流能力锁死。

---

### 3.2 CI 检查偏轻

当前 CI 主要是：

```bash
go test ./...
```

并且矩阵覆盖 Go 1.24、1.25。

建议补充：

```bash
go test ./...
go test -race ./...
go vet ./...
gofmt check
govulncheck ./...
staticcheck ./...
```

推荐 CI 结构：

```yaml
name: CI

on:
  push:
    branches:
      - main
      - dev
  pull_request:

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      fail-fast: false
      matrix:
        go-version:
          - "1.24"
          - "1.25"

    steps:
      - uses: actions/checkout@v6

      - uses: actions/setup-go@v6
        with:
          go-version: ${{ matrix.go-version }}
          cache: true

      - name: Check format
        run: |
          test -z "$(gofmt -l .)"

      - name: Vet
        run: go vet ./...

      - name: Test
        run: go test ./...

      - name: Race test
        run: go test -race ./...
```

`govulncheck` 和 `staticcheck` 可以单独加 job，避免每次矩阵重复安装。

---

### 3.3 `test/` 目录作为示例目录不够清晰

当前 README 里有类似：

```bash
go run ./test/steamuser
go run ./test/playerservice
go run ./test/wishlistservice
```

问题：

- Go 项目里 `test/` 容易被理解为测试辅助目录；
- 用户不容易判断这些是 live examples；
- 真实 API 示例可能需要 key/token，不适合和普通 test 混在一起。

建议迁移为：

```text
examples/live/steamuser
examples/live/playerservice
examples/live/wishlistservice
```

并在 README 里写清楚：

```bash
STEAM_API_KEY=xxx go run ./examples/live/steamuser
STEAM_ACCESS_TOKEN=xxx go run ./examples/live/playerservice
```

---

### 3.4 测试命名需要修正

当前测试中有一个语义不一致的小问题：

```go
TestNewClientRequiresAPIKey
```

但实际行为是 `steam.NewClient()` 不需要 API key 也可以创建成功。

建议改名为：

```go
TestNewClientWithoutAPIKey
```

这类问题虽然很小，但会影响维护者对 SDK 行为的理解。

---

## P1：重要改进项

### 3.5 Endpoint 稳定性需要标注

当前 SDK 覆盖的接口很多，包括：

- SteamUser；
- SteamUserStats；
- PlayerService；
- StoreService；
- WishlistService；
- SteamChartsService；
- SteamApps；
- SteamWebAPIUtil；
- UserReviewsService；
- UserStoreVisitService；
- StoreBrowseService；
- StoreCatalogService；
- StoreTopSellersService。

问题是 Steam 的接口稳定性并不一致：

- 有些是官方公开 Web API；
- 有些需要 API key；
- 有些需要 access token；
- 有些更像 Steam Store / Client 内部接口；
- 有些 payload 很大且变化频繁。

建议引入 endpoint 稳定等级：

```go
type EndpointStability string

const (
    StabilityOfficialPublic EndpointStability = "official_public"
    StabilityOfficialAuth   EndpointStability = "official_auth"
    StabilityStoreVolatile  EndpointStability = "store_volatile"
    StabilityAddon          EndpointStability = "addon"
)
```

文档中可以这样标注：

| Service | Method | Credential | Stability | Notes |
|---|---|---|---|---|
| SteamUser | GetPlayerSummaries | optional key | official_public | stable |
| SteamUserStats | GetNumberOfCurrentPlayers | none | official_public | stable |
| WishlistService | GetWishlistItemsOnSale | access token | store_volatile | store_item uses raw JSON |
| OpenID addon | Verify | none | addon | confirms identity only |

这样用户会更容易理解哪些接口适合生产依赖，哪些接口需要容忍 Steam 变动。

---

### 3.6 错误体需要增加安全使用方式

当前 `APIError` 中保存了：

```go
Body []byte
```

优点：

- 排查 HTTP status、decode、API response 问题很方便；
- 对 Steam 这种不稳定 payload 的接口很有帮助。

风险：

- 用户可能直接把 `Body` 打进生产日志；
- body 中可能包含用户资料、token 相关错误、私有信息；
- decode error 时可能记录过大的响应体。

建议新增：

```go
func (e *APIError) BodyPreview(max int) string
```

或者：

```go
func (e *APIError) SafeBodyPreview() string
```

建议文档中明确：

```go
var apiErr *steam.APIError
if errors.As(err, &apiErr) {
    log.Printf("steam api error: kind=%s status=%d message=%s body_preview=%s",
        apiErr.Kind,
        apiErr.StatusCode,
        apiErr.Message,
        apiErr.BodyPreview(1024),
    )
}
```

同时提醒：

> 不建议在生产环境直接记录完整 `apiErr.Body`。

---

### 3.7 文档需要拆分

当前 README 已经比较完整，但随着功能增多，建议拆分为：

```text
docs/
  getting-started.md
  authentication.md
  rate-limit-and-retry.md
  endpoint-stability.md
  errors.md
  proxy.md
  live-examples.md
  addons.md
  api.md
```

各文档职责：

| 文档 | 作用 |
|---|---|
| getting-started.md | 安装、快速开始、最小示例 |
| authentication.md | key/token/OpenID 的区别 |
| rate-limit-and-retry.md | 限流、重试、key rotation |
| endpoint-stability.md | endpoint 稳定等级说明 |
| errors.md | APIError 使用方式 |
| proxy.md | static/round-robin/routing proxy |
| live-examples.md | 真实 Steam API 示例 |
| addons.md | A2S/OpenID 说明 |
| api.md | endpoint method reference |

README 只保留：

- 项目定位；
- 安装；
- Quick Start；
- 核心特性；
- 常用链接；
- 示例入口。

---

### 3.8 Live API 测试需要独立出来

不要把真实 Steam API 调用混进默认 CI。

建议新增：

```text
.github/workflows/live.yml
```

触发方式：

```yaml
on:
  workflow_dispatch:
```

环境变量：

```yaml
STEAM_API_KEY: ${{ secrets.STEAM_API_KEY }}
STEAM_ACCESS_TOKEN: ${{ secrets.STEAM_ACCESS_TOKEN }}
```

建议只测少量稳定接口：

```text
ISteamUser/GetPlayerSummaries
ISteamUserStats/GetNumberOfCurrentPlayers
ISteamWebAPIUtil/GetSupportedAPIList
```

这样既能验证 SDK 真的可用，又不会让 CI 被 Steam 网络、限流、权限变化拖垮。

---

## P2：中期优化项

### 3.9 增加 contract tests

建议建立一个更清晰的测试分层：

```text
tests/
  contract/
    steamuser_test.go
    steamuserstats_test.go
    wishlist_test.go
```

Contract test 的目标不是请求真实 Steam，而是用 `httptest` 固定：

- path；
- method；
- query；
- body；
- credential injection；
- response decode；
- error kind。

当前 `client_test.go` 已经有不少这类内容，后续可以逐步迁移、分组、拆文件。

---

### 3.10 增加 godoc 示例

建议给核心 API 增加 `ExampleXxx`：

```go
func ExampleNewClient() {
    client, err := steam.NewClient(
        steam.WithTimeout(10*time.Second),
    )
    if err != nil {
        panic(err)
    }
    defer client.Close()

    _ = client
}
```

给常用 service 增加：

```go
func ExampleService_GetPlayerSummaries()
func ExampleService_GetNumberOfCurrentPlayers()
func ExampleService_GetNewsForApp()
```

这样 pkg.go.dev 上会更友好。

---

### 3.11 给不稳定 payload 使用 RawMessage 策略

你已经在 `GetWishlistItemsOnSale` 里把 `store_item` 设计为 raw JSON，这是很正确的。

建议形成规则：

- 稳定字段：typed struct；
- 巨大且频繁变化字段：`json.RawMessage`；
- 不确定字段：先用 `map[string]json.RawMessage` 或 Raw 类型；
- 只在确认稳定后再升级为 typed struct。

可以在文档里说明：

> steam-go 优先为稳定接口提供 typed response；对 Steam Store 这类频繁变化的大型 payload，会保留 RawMessage 以避免 SDK 频繁破坏兼容性。

---

### 3.12 增加版本与兼容策略

建议新增 `docs/compatibility.md`：

```markdown
# Compatibility Policy

## Before v1

- Public API may change between minor versions.
- Breaking changes will be documented in CHANGELOG.
- Endpoint coverage may expand frequently.

## After v1

- Public API follows semantic versioning.
- Breaking changes require major version bump.
- Unstable Steam endpoints may still change if upstream payload changes.
```

并新增 `CHANGELOG.md`：

```markdown
# Changelog

## v0.1.0

- Initial public preview.
- Added typed Steam Web API client.
- Added OpenID addon.
- Added A2S addon bridge.
```

---

## P3：后续扩展项

### 3.13 继续补齐官方公开接口

在基础设施稳定之后，再继续扩接口。

优先级建议：

1. 官方公开、稳定、低权限接口；
2. 用户常用 API key 接口；
3. access token 接口；
4. store/client volatile 接口；
5. 特殊 addon 能力。

不要反过来先大量追 Steam Store 内部接口，否则维护成本会快速上升。

---

### 3.14 增加 endpoint coverage 表

建议生成一个文档：

```text
docs/coverage.md
```

内容类似：

| Interface | Method | Implemented | Raw | Typed | Stability |
|---|---:|---:|---:|---:|---|
| ISteamUser | GetPlayerSummaries | yes | yes | yes | official_public |
| ISteamUser | GetFriendList | yes | yes | yes | official_auth |
| ISteamUserStats | GetNumberOfCurrentPlayers | yes | yes | yes | official_public |
| IWishlistService | GetWishlistItemsOnSale | yes | yes | partial | store_volatile |

这会明显提升仓库专业度。

---

### 3.15 增加 examples 分类

建议最终形成：

```text
examples/
  basic/
    player_summaries/
    current_players/
  auth/
    friend_list/
    player_achievements/
  proxy/
  openid/
  a2s/
  live/
    steamuser/
    playerservice/
    wishlistservice/
```

其中：

- `basic/` 尽量不依赖 key；
- `auth/` 明确依赖 key/token；
- `live/` 是真实环境调试工具；
- `openid/` 是浏览器登录示例；
- `a2s/` 是服务器查询示例。

---

## 4. 推荐路线图

## v0.1.x：当前预览版

目标：可以被早期用户试用。

建议完成：

- 保持当前 README 和 docs；
- 打 tag；
- 确认 `go test ./...` 通过；
- 确认 pkg.go.dev 可正常展示；
- 明确标注 pre-v1。

---

## v0.2.x：基础能力稳定版

目标：稳定 SDK 核心 API。

建议完成：

- 改进 RateLimit API；
- 修正测试命名；
- 完善 CI；
- 增加错误体 preview；
- 迁移 `test/` 到 `examples/live/`；
- 增加 `CHANGELOG.md`；
- 增加 `compatibility.md`。

---

## v0.3.x：文档与可信度增强版

目标：让外部用户更敢依赖。

建议完成：

- 增加 endpoint stability 文档；
- 增加 endpoint coverage 表；
- 拆分认证、限流、错误、代理文档；
- 增加 live API workflow；
- 增加 godoc examples。

---

## v0.4.x：覆盖面增强版

目标：继续扩展官方稳定接口。

建议完成：

- 优先补齐官方公开接口；
- 对高频接口增加 typed response；
- 对 volatile payload 保持 raw JSON；
- 增加更多 contract tests。

---

## v1.0.0：稳定发布版

目标：public API 冻结，外部项目可以正式依赖。

v1 发布前建议最低门槛：

- CI 包含 test / race / vet / gofmt / govulncheck；
- README 简洁清晰；
- docs 文档拆分完成；
- endpoint stability 完成；
- examples/live 完成；
- CHANGELOG 完成；
- compatibility policy 完成；
- public API 命名基本冻结；
- Go Report Card 正常；
- pkg.go.dev 展示正常。

---

## 5. 推荐优先级清单

### P0：立即处理

- [ ] 修正 `TestNewClientRequiresAPIKey` 命名；
- [ ] 扩展 RateLimit API；
- [ ] CI 增加 `go vet ./...`；
- [ ] CI 增加 gofmt 检查；
- [ ] CI 增加 `go test -race ./...`；
- [ ] 将 `go run ./test/...` 迁移到 `examples/live/...`。

---

### P1：近期处理

- [ ] 增加 endpoint stability 分类；
- [ ] 增加 endpoint coverage 文档；
- [ ] 增加 `docs/authentication.md`；
- [ ] 增加 `docs/rate-limit-and-retry.md`；
- [ ] 增加 `docs/errors.md`；
- [ ] 增加错误体 preview 方法；
- [ ] 增加 live API 手动 workflow。

---

### P2：中期处理

- [ ] 拆分 `client_test.go` 中过大的测试文件；
- [ ] 建立 `tests/contract/`；
- [ ] 增加 godoc examples；
- [ ] 增加 `CHANGELOG.md`；
- [ ] 增加 `docs/compatibility.md`；
- [ ] 给 volatile payload 制定 RawMessage 规则。

---

### P3：长期处理

- [ ] 继续补官方公开 API；
- [ ] 增加更多 typed response；
- [ ] 增加更多 live examples；
- [ ] 为不同 credential 场景补充文档；
- [ ] 根据用户反馈决定是否增加更高级的 retry/backoff 策略。

---

## 6. 建议仓库定位描述

可以使用下面这段作为 GitHub repository description：

> A lightweight, typed Go SDK for Steam Web API, with optional addons for Steam OpenID and A2S server queries.

中文定位可以写成：

> 一个轻量、类型化的 Steam Web API Go SDK，并通过 addons 提供 Steam OpenID 与 A2S 服务器查询能力。

---

## 7. 最终建议

`steam-go` 当前最重要的任务不是继续疯狂加接口，而是把已有能力打磨到别人敢依赖。

推荐主线：

```text
先稳定 SDK 核心 API
再完善测试和 CI
再标注 endpoint 稳定性
再补文档与示例
最后继续扩展接口覆盖
```

只要按这个节奏推进，`steam-go` 会从“个人可用工具包”逐步变成一个真正可以发布、可以被其他 Go 开发者使用的 Steam SDK。
