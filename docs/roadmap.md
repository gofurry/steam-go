# steam-go Roadmap

> 目标：保持 `steam-go` 的版本节奏克制、边界清晰、可维护。

---

## Current Position

`steam-go` 已经具备稳定 SDK 形态：统一 `Client`、官方 Steam Web API typed/raw 调用、只读 Web helper、addons、request controls、proxy、cache、retry、rate limit、traffic policy、request observer、release checklist 和 coverage drift workflow。

下一阶段不做架构重建，也不追求 endpoint 数量清零；优先围绕现有 runtime 做小步兼容增强。

---

## Version Plan

### `v1.3.5` - Runtime Efficiency and Coverage Catch-up

**Status:** Planned

**Scope:** Performance / Observability / Stability / Official API / Documentation

**Goal:** 建立 runtime 性能基线，增强 cache、request control、proxy、retry 等热路径的可观测与可调能力，并补齐少量高价值官方 API。

#### Focus

- 建立 benchmark 基线，覆盖 cache、transport CookieJar、request control、proxy 和 raw HTTP 热路径。
- 新增只读 runtime stats，帮助使用者观察 cache、request control、proxy、retry 等运行状态。
- 增强 cache 可配置性与统计能力，并评估可选 singleflight 降低并发 miss 压力。
- 优化 context CookieJar 和 Retry-After 等小型高频路径。
- 补齐少量高价值 official API coverage，避免为了覆盖率数字盲目扩张。
- 在 addon 层补齐登录到 web session，再到显式单项 free claim 的产品流原子能力。

#### Non-goals

- 不做大规模目录重构或 request executor 重写。
- 不改变现有 public API 签名或默认语义。
- 不修改 `RequestObserver` 字段。
- 不一次性补完所有 missing endpoints。
- 不引入复杂外部依赖。
- 不把 Web helper 扩展成爬虫框架。
- 不扩大账号自动化、购买、出售、交易或批量领取能力。
- 不把登录、cookie、mobile confirmation 或 free claim 编排提升进 core；这些能力继续留在 addon / product-flow 层。
- 不提供一键批量领取、自动回答风险校验、自动绕过 Steam Guard 或托管 mobile secrets 的能力。

#### P0 Tasks - Runtime and Performance Foundation

- [x] 新增 benchmark 基线：cache hit/miss、transport context CookieJar、request control high cardinality、sticky proxy、raw HTTP read limit。
- [x] 新增 `Client.RuntimeStats()` 只读 snapshot，覆盖 cache、transport/request control、proxy 等非敏感运行状态。
- [x] 为 cache 增加 stats：entries、max entries、hit、miss、store、eviction、conditional hit。
- [x] 增加 cache max entries 可配置能力，保持现有默认容量与 `TrafficCachePolicy{TTL}` 兼容。
- [x] 评估并实现可选 GET cache miss singleflight，仅在 cache enabled 且 GET 请求下生效。
- [x] 优化 context CookieJar 热路径：context jar 与默认 jar 相同时跳过不必要的 client clone。
- [x] 为异常 `Retry-After` 增加最小 retry floor，覆盖负数、过去日期、无效格式和 context cancel。

#### P1 Tasks - Official API Coverage Catch-up

- [x] 补齐 `IAuthenticationService` 中高价值缺失方法，例如 auth-session info、risk info、mobile confirmation 等低层 API。
- [x] 新增 `api/contentserverdirectoryservice` package，优先覆盖 CDN、client update hosts、depot patch info、SteamPipe servers 等内容分发目录能力。
- [x] 更新 coverage snapshot 与 triage，把 broadcast、clientstats、用途不清晰或上报性质明显的端点标记为 deferred / won't add candidate。
- [x] 为新增 official API 增加 request path、method、query/body、raw method、typed decode 与输入校验测试。

#### P2 Tasks - Optional Enhancements

- [ ] 评估新增 `DoRawHTTPRequestStream`，用于 CDN/static/大响应场景，并明确调用方关闭 body 的责任。
- [ ] 评估新增 `HeaderProfileBuilder`，保持默认 header profile 稳定，同时允许用户显式配置 UA、language、accept 等字段。
- [ ] 根据 P0 benchmark 结果决定是否顺延复杂优化到 `v1.3.6` 或 `v1.4.0`。

#### Stage 2 Tasks - Websession and Free-claim Product Flow

- [ ] 为 `addons/websession` 新增保守的登录编排示例或 helper，串联 `StartWithCredentials`、`SubmitSteamGuardCode`、`Poll`、`RefreshTokenToWebCookies`、`ValidateWebCookies`，但不保存密码、不自动绕过确认。
- [ ] 为 `addons/websession` 评估并新增 QR 登录 helper，例如 `StartWithQR` / `PollQR`，复用 core `AuthenticationService.BeginAuthSessionViaQR`，由调用方负责展示二维码和人工确认。
- [ ] 为 mobile confirmation 明确 addon 边界：只允许调用方显式传入签名/确认结果，不托管 mobile secret；必要时新增薄封装或文档示例。
- [ ] 新增 web session 持久化/恢复能力，支持安全地保存、加载、校验 `WebCookieResult` / cookie jar，并明确本地存储权限、过期和 redaction 边界。
- [ ] 新增 free-claim 产品流示例，串联 `freeclaim.SearchPromotions`、`ResolveFreePackages`、`websession` 登录/session、`ClaimPackage`、`IsAppOwned`，保持单项显式领取，不做默认批量自动领取。
- [ ] 扩展 `addons/websession` 与 `addons/freeclaim` 文档，说明 authentication service、web session、cookie jar、free claim 之间的层级关系和安全边界。

#### Acceptance Criteria

- Public API 只做兼容性加法，不删除、不改签名、不改变默认请求语义。
- Runtime stats、observer、logs、examples 不暴露 query、header、body、cookie、token、API key 或 proxy password。
- Benchmark 命令和基线文档可复现，至少覆盖核心 runtime 热路径。
- Cache stats、capacity、TTL、validator、cookie/session/language dimension 和 optional singleflight 有测试覆盖。
- Retry-After floor 对负数、零值、过去日期、未来日期、无效格式和 cancel path 有测试覆盖。
- 新 official API 有 raw/typed 调用、文档说明和 coverage 更新。
- Stage 2 登录/free-claim 示例必须保持人工参与，不保存明文密码，不输出 token/cookie/sessionid，不默认批量领取。
- Web session 持久化必须有 redaction、安全存储权限和过期校验说明。
- `go test ./...`、`go test -race ./...`、`go vet ./...`、`staticcheck ./...`、`govulncheck ./...` 在发布前通过或有明确说明。

#### Suggested Minimum Delivery

如果时间有限，`v1.3.5` 最小可交付范围优先压缩为：

- benchmark baseline；
- `Client.RuntimeStats()`；
- cache stats + max entries；
- Retry-After floor；
- `IContentServerDirectoryService`。

---

## Release Gate

每个 release 前至少满足：

- [ ] `go test ./...`
- [ ] `go test -race ./...`
- [ ] `go vet ./...`
- [ ] `staticcheck ./...`
- [ ] `govulncheck ./...`
- [ ] API compatibility check
- [ ] staged secret scan clean
- [ ] README / docs / cookbook / release notes 中英文同步
- [ ] 新 public API 有 examples、cookbook 或 reference 说明
- [ ] 新 Web surface 标明 unofficial / volatile / read-only 边界
