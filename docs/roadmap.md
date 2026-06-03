# steam-go v1.2.x → v1.3.x 路线图

> 目标：把 `steam-go` 从“功能已经比较完整的个人开源 SDK”推进到“可信、可维护、容易采用、能够长期跟随 Steam 上游变化的 Go 工具包”。
>
> 范围：本路线图只规划 `v1.2.x` 到 `v1.3.x`，刻意放慢版本节奏，不把之前建议中的内容拆到过多 minor 版本。
>
> 建议放置路径：`docs/roadmap.md` 或 `docs/releases/roadmap-v1.2-v1.3.md`

---

## 1. 路线图结论

`steam-go` 当前的工程底座已经比较成熟：核心 `Client` 分成 `client.API.*` 和 `client.Web.*`，官方 Steam Web API 与非官方 Store / Community / Market JSON surface 有明确边界；请求层已有 timeout、retry、rate limit、proxy、traffic policy、body cap、block detection、short cache、credential rotation 等生产向能力；addon 体系也已经承接了 OpenID、websession、freeclaim、assets、A2S 等扩展场景。

下一阶段最值得做的事情不是快速堆 endpoint，也不是把 Store / Community 扩成大而全 SDK，而是优先做下面三类工作：

1. **可信度建设**：安全说明、贡献指南、CI 固化、release checklist、兼容性门禁、文档重构。
2. **维护自动化**：自动发现 Steam 官方 API 变化、生成 coverage diff、维护 fixture corpus、定期 live smoke。
3. **采用体验提升**：doctor/smoke 工具、pkg.go.dev 示例、README 收敛、常见场景 cookbook、可观测性 hook。

版本节奏建议：

- `v1.2.x`：质量、治理、文档、发布流程、兼容性安全网。
- `v1.3.x`：自动化维护工具、fixture/live smoke、doctor 工具、有限的高价值 helper。
- `v1.4+`：暂不纳入本路线图，只作为 backlog，不提前承诺。

---

## 2. 产品定位

### 2.1 推荐定位

`steam-go` 应继续定位为：

> A stable Go SDK for the official Steam Web API, with practical request controls and carefully scoped read-only Steam Web helpers.

中文可表述为：

> 面向官方 Steam Web API 的稳定 Go SDK，提供生产可用的请求控制能力，并谨慎扩展少量高价值、只读的 Steam Web 辅助能力。

这个定位有三个关键点：

- **官方 API 是核心**：`client.API.*` 是长期稳定主线。
- **Web surface 是谨慎扩展**：`client.Web.*` 只做高价值只读 JSON endpoint，不承诺完整 Store / Community SDK。
- **请求控制是差异化能力**：代理、限流、重试、block detection、traffic class、缓存、凭据轮换是这个库的重要护城河。

### 2.2 明确不做的事情

为了避免 scope creep，`v1.2.x → v1.3.x` 不建议做这些事情：

- 不把核心包扩成完整 Steam Store SDK。
- 不把核心包扩成完整 Steam Community SDK。
- 不在核心包内置浏览器自动化 fallback。
- 不做自动购买、自动出售、自动交易、批量领取等账号自动化能力。
- 不为了短期覆盖率把大量不稳定 payload 强行 typed 化。
- 不静默混入未文档化 scraping API。

---

## 3. 总体版本节奏

### 3.1 `v1.2.x`：稳定化与可信度阶段

`v1.2.x` 的目标不是新增大量功能，而是让项目更像一个可被外部用户放心采用的 SDK。

核心关键词：

- 仓库卫生
- 安全文档
- CI 稳定化
- 兼容性保护
- README 收敛
- 文档可采用性
- release 流程
- redaction / credential safety

建议拆分：

| 版本 | 主题 | 说明 |
|---|---|---|
| `v1.2.1` | Repository hygiene | 补基础开源文件、清理仓库、固定 CI 工具链、补 release checklist。 |
| `v1.2.2` | Documentation adoption | 重构 README，补 cookbook、pkg.go.dev 示例、常见场景文档。 |
| `v1.2.3` | Compatibility & safety | API diff 检查、redaction 增强、凭据安全文档、兼容性测试策略。 |
| `v1.2.x` 后续 patch | Bugfix-only | 根据用户反馈和 CI 问题修补，不主动扩大战线。 |

这些 patch 可以按实际工作量合并，不需要机械地每个主题都发一个版本。关键是：`v1.2.x` 不再承担“大功能扩张”的职责。

### 3.2 `v1.3.x`：维护自动化与采用工具阶段

`v1.3.x` 的目标是建立长期维护能力，让 Steam 上游变化能被及时发现、评估和跟进。

核心关键词：

- Steam API drift detection
- Endpoint coverage automation
- Fixture corpus
- Live smoke report
- Doctor command
- Paginator / batch helper
- Lightweight observability

建议拆分：

| 版本 | 主题 | 说明 |
|---|---|---|
| `v1.3.0` | API coverage automation | 引入 `steamapi-sync` 或等价工具，自动生成/更新 coverage diff。 |
| `v1.3.1` | Fixture & smoke baseline | 建立 fixture corpus、golden decode test、live smoke 报告。 |
| `v1.3.2` | Doctor command | 提供网络、凭据、代理、Store/Community 可用性诊断工具。 |
| `v1.3.3` | High-value helpers | 增加有限的 read-only paginator/batch helper 和可观测性 hook。 |
| `v1.3.x` 后续 patch | Stabilization | 修复自动化工具、fixture、doctor 在真实环境中的问题。 |

---

## 4. `v1.2.x` 详细计划

## 4.1 `v1.2.1`：Repository hygiene

**Status:** Completed  
**Scope:** CI/Release / Security/Safety / Documentation / Developer-facing  
**Goal:** 让仓库具备更完整的开源项目信任基础。

### 已完成

- [x] 新增 `SECURITY.md`，说明敏感凭据、cookie、proxy URL 和安全报告边界。
- [x] 新增 `CONTRIBUTING.md`，说明测试、endpoint、addon、兼容性和 secret safety 贡献要求。
- [x] 新增 bug report、feature request、question issue templates。
- [x] 新增 pull request template，覆盖兼容性、验证、文档和 secret safety。
- [x] 新增 `docs/releases/checklist.md`。
- [x] 扩展 `.gitignore`，覆盖 IDE/editor、凭据、日志、coverage 和构建产物。
- [x] 固定主线 CI 的 `staticcheck v0.7.0` 与 `govulncheck v1.3.0`。
- [x] 新增 latest toolchain scheduled advisory workflow。
- [x] 升级 `golang.org/x/net` 到当前 `govulncheck` 已修复版本，并固定 Go 1.26 CI lane 到 Go 1.26.4。
- [x] 新增中英文 `v1.2.1` release notes。

### 验收标准

- [x] 仓库有完整基础开源文件。
- [x] CI 主线更稳定，不因 latest 工具链随机变化而无故失败。
- [x] release 前有可执行检查清单。
- [x] 仓库中不再包含非必要 IDE 元数据。
- [x] `go test ./...`、`go test -race ./...`、`go vet ./...`、`staticcheck` 和 `govulncheck` 均已通过。

---

## 4.2 `v1.2.2`：Documentation adoption

**Status:** Completed  
**Scope:** Documentation / User-facing / Developer-facing  
**Goal:** 降低新用户理解和采用成本，让 README 从“全量说明”变成“入口页”。

### 本轮执行项

- [x] 保留 root `README.md` 顶部徽章、导航和 ASCII LOGO。
- [x] 将 root `README.md` 主体重构为入口页：定位、核心卖点、安装、三个 quick start、生产建议、稳定性边界、addon 简表和文档入口。
- [x] 将 `docs/zh/README.md` 同步重构为中文入口页，并保留顶部徽章、导航和 ASCII LOGO。
- [x] 新增英文 cookbook：basic API、OpenID、read-only Web、proxy/region、rate limit/retry、error handling、credential redaction、assets。
- [x] 新增中文 cookbook：basic API、OpenID、read-only Web、proxy/region、rate limit/retry、error handling、credential redaction、assets。
- [x] 更新 `docs/README.md` 和 `docs/zh/README.md` cookbook 索引。
- [x] 评估最低 Go 版本：代码语法本身接近 Go 1.22，但当前安全依赖链 `golang.org/x/net v0.55.0`、`x/term v0.43.0`、`x/sys v0.45.0` 要求 Go 1.25，因此本轮不降低 `go.mod` 的 `go 1.25.0`。
- [x] 为核心包和关键 addon 增加 pkg.go.dev `Example...` 测试。

### Deferred

- Go 1.22 支持暂不承诺。只有在安全依赖链允许且完整 CI 验证通过时，才重新评估降低最低 Go 版本。

### 验收标准

- [x] README 变短，入口更清晰。
- [x] 至少 6 篇 cookbook 完成。
- [x] pkg.go.dev 能展示核心示例。
- [x] 最低 Go 版本策略有明确说明。

---

## 4.3 `v1.2.3`：Compatibility & safety

**Status:** Completed  
**Scope:** Compatibility / Security/Safety / Documentation / CI/Release  
**Goal:** 建立 `v1` 兼容性保护和凭据安全默认路径。

### 已完成

- [x] 新增 `internal/tools/apidiffcheck`，用于 release 前对比 base ref 与当前工作树的 public API drift。
- [x] 将 API diff 本地检查加入 `docs/releases/checklist.md`。
- [x] 扩展 `RedactSensitiveURL(...)` 覆盖 `refresh_token`、`steamLoginSecure`、`sessionid`、`webapi_token`、`loyalty_webapi_token` 等 query key。
- [x] 新增 `RedactSensitiveHeaderValue(...)` 和 `RedactSensitiveHeaders(...)`。
- [x] 补充 redaction tests，覆盖 query、proxy userinfo、redirect final URL、header/cookie redaction 和 clone 不变性。
- [x] 新增 `docs/security/credentials.md` 与 `docs/zh/security/credentials.md`。
- [x] 增强中英文 error handling cookbook，明确 retry 与重新登录/换 token 的边界。
- [x] 更新文档索引和 API reference 中的 redaction 指引。

### 验收标准

- [x] release 前能检查 public API drift。
- [x] redaction 覆盖 query、proxy、cookie/header 场景。
- [x] 凭据安全有独立文档。
- [x] 错误处理有 cookbook 示例。

---

# 5. `v1.3.x` 详细计划

## 5.1 `v1.3.0`：API coverage automation

### 目标

把 endpoint 维护从“人工记忆”升级为“自动发现、自动 diff、自动生成报告”。

### 建议工具：`steamapi-sync`

可以放在：

- `cmd/steamapi-sync`
- 或 `internal/tools/steamapi-sync`
- 或 `tools/steamapi-sync`

推荐先作为仓库维护工具，不急着承诺为稳定 public API。

### 功能设计

`steamapi-sync` 做四件事：

1. 调用官方 `ISteamWebAPIUtil/GetSupportedAPIList`。
2. 生成 Steam 官方 API inventory。
3. 与仓库当前 `client.API.*` coverage 对比。
4. 输出 markdown / JSON diff 报告。

输出文件建议：

- `docs/api/coverage.generated.md`
- `docs/api/coverage.generated.json`
- `docs/api/coverage-diff.md`

每个 endpoint 记录：

- interface/service name
- method name
- version
- HTTP method
- parameters
- auth requirement
- SDK coverage status
- typed/raw status
- stability level
- docs/example link
- live smoke status

### GitHub Actions 集成

新增 scheduled job：

- 定期跑 `steamapi-sync`。
- 若发现新增/变化 endpoint，生成 artifact。
- 可选：自动开 issue 或创建 PR。
- 不建议一开始自动改代码，先只生成报告。

### 验收标准

- 能生成当前官方 API coverage 报告。
- 能发现新增、删除、参数变化或版本变化。
- 报告可读，可用于手工规划后续 endpoint。
- 不因 Steam 偶发请求失败导致主线 CI 失败。

---

## 5.2 `v1.3.1`：Fixture corpus & smoke baseline

### 目标

让 payload drift 能被测试发现，而不是用户运行时报错后才知道。

### Fixture 目录建议

```text
testdata/
  fixtures/
    official/
      ISteamUser/GetPlayerSummaries/v2/public.json
      ISteamUserStats/GetGlobalAchievementPercentagesForApp/v2/public.json
    web/
      storefront/GetAppDetails/app_550_en.json
      storefront/GetAppReviews/app_550_en.json
      market/GetPriceOverview/app_730_item_key.json
    addons/
      assets/store_media_app_550.json
```

### 测试类型

#### 5.2.1 Decode compatibility tests

对 typed response 做 decode test：

- fixture 必须能 decode 到当前 response struct。
- 可接受额外字段。
- 不接受已承诺稳定字段 decode 失败。

#### 5.2.2 Raw subtree stability tests

对 `json.RawMessage` 字段做最小检查：

- raw 字段存在时必须是合法 JSON。
- 文档说明该字段为什么 raw。
- 不强制检查 raw 内部所有字段。

#### 5.2.3 Golden snapshot tests

对关键 helper 输出做 golden test：

- asset URL generation
- redaction output
- coverage generated output
- doctor diagnostic formatting

#### 5.2.4 Live smoke opt-in

保留 live smoke，但必须 opt-in：

```bash
STEAM_GO_LIVE=1 go test ./examples/live/...
```

或：

```bash
go run ./cmd/steam-go-smoke
```

live smoke 输出应避免泄露 secret。

### 验收标准

- 至少 10 个代表性 fixture。
- 核心 typed endpoint 有 decode regression test。
- Web volatile payload 有 raw subtree 检查。
- live smoke 不默认跑，不影响普通贡献者。

---

## 5.3 `v1.3.2`：Doctor command

### 目标

提供一个用户可以直接运行的诊断工具，快速判断问题来自 key/token、网络、代理、Steam 上游、cookie 还是 SDK 配置。

### 建议命令

```bash
go run ./cmd/steam-go-doctor
```

或作为 example：

```bash
go run ./examples/doctor
```

一开始建议放 `examples/doctor`，成熟后再考虑 `cmd/steam-go-doctor`。

### 检查项

#### 5.3.1 环境与版本

- Go version
- steam-go module version，若可获取
- OS / arch
- timeout / retry / rate limit 配置摘要

#### 5.3.2 官方 API 可达性

- `SteamWebAPIUtil.GetServerInfo`
- `SteamUser.GetPlayerSummaries`，使用公开 SteamID 示例
- API key 是否被注入
- 401 / 403 / 429 / 5xx 分类展示

#### 5.3.3 Web surface 可达性

- Storefront app details
- App reviews
- Community inventory public sample，谨慎选择公开库存示例
- Market price overview

#### 5.3.4 Proxy 检查

- direct mode 是否成功
- static proxy 是否成功
- routing proxy 是否按 host/path 生效
- sticky proxy session key 是否保持选择
- health-checked proxy 是否出现 cooldown

#### 5.3.5 Credential 检查

- API key 是否存在，但不打印。
- access token 是否存在，但不打印。
- cookie jar 是否配置。
- refresh token 不应通过 flag 明文传入。

#### 5.3.6 输出格式

默认人类可读：

```text
[OK] official API reachable
[OK] Storefront reachable
[WARN] Community inventory requires cookie or public profile
[FAIL] API key unauthorized: status=401
```

可选 JSON：

```bash
go run ./examples/doctor -json
```

### 验收标准

- 用户能用 doctor 快速定位常见网络/代理/凭据问题。
- 输出默认不泄露 secret。
- doctor failure 不等同于 SDK failure，文档要解释诊断边界。

---

## 5.4 `v1.3.3`：High-value helpers

### 目标

只增加少量“非常高频、低风险、只读”的 helper，避免 API 膨胀。

### 推荐 helper

#### 5.4.1 Reviews paginator

场景：Store reviews cursor 翻页很常见，适合 SDK 提供 helper。

建议形式：

- `GetAppReviewsPage`
- `AppReviewsIterator`
- 或 `ListAppReviews(ctx, appID, opts, handler)`

优先选择不一次性拉爆内存的 handler/iterator 模式。

#### 5.4.2 Inventory paginator

场景：Community inventory 经常需要翻页，且可能需要 cookie。

建议明确：

- 只读。
- 不管理登录。
- 不刷新 cookie。
- 不保证 private inventory 可访问。

#### 5.4.3 Batch app details / price overview

场景：批量查询 app details、market price overview 很常见。

要求：

- 内置并发限制。
- 尊重 client rate limit。
- 返回 per-item error，不因单个失败丢掉整个批次。
- 支持 context cancellation。

#### 5.4.4 Lightweight observability hook

先提供低依赖 hook，不直接引入 OpenTelemetry：

```go
type RequestObserver interface {
    OnRequestStart(...)
    OnRequestDone(...)
}
```

或更简单：

```go
type RequestObserverFunc func(event RequestEvent)
```

事件字段建议：

- traffic class
- method
- host
- path template，避免 raw query
- status code
- error kind
- retry attempt
- cache hit
- block detected
- proxy used，脱敏
- duration

### 验收标准

- helper 只覆盖高频只读场景。
- 所有 helper 都有 context、rate limit、body cap、error handling 示例。
- 不引入账号自动化行为。
- 可观测性默认无依赖、无输出。

---

# 6. GitHub Milestones 建议

可以在 GitHub 上创建以下 milestones。

## Milestone: `v1.2.1 Repository hygiene`（Completed）

建议 issues：

1. [x] Add `SECURITY.md`.
2. [x] Add `CONTRIBUTING.md`.
3. [x] Add issue templates and PR template.
4. [x] Add release checklist.
5. [x] Pin staticcheck and govulncheck versions in CI.
6. [x] Add scheduled latest-toolchain CI.
7. [x] Remove IDE metadata from repository if unnecessary.
8. [ ] Add Dependabot or Renovate config.

## Milestone: `v1.2.2 Documentation adoption`（Completed）

建议 issues：

1. [x] Restructure README into concise entry page.
2. [x] Add cookbook: basic official API usage.
3. [x] Add cookbook: OpenID login.
4. [x] Add cookbook: Web read-only endpoints.
5. [x] Add cookbook: proxy and region network setup.
6. [x] Add cookbook: retry and rate limit.
7. [x] Add cookbook: credential redaction.
8. [x] Add pkg.go.dev examples for core package.
9. [x] Evaluate minimum Go version.

## Milestone: `v1.2.3 Compatibility and safety`（Completed）

建议 issues：

1. [x] Add exported API diff check.
2. [x] Add redaction tests for sensitive query fields.
3. [x] Add redaction tests for proxy userinfo.
4. [x] Add credential safety document.
5. [x] Add error handling cookbook.
6. [x] Add compatibility policy checklist to release process.

## Milestone: `v1.3.0 API coverage automation`

建议 issues：

1. Implement `steamapi-sync` prototype.
2. Generate official API inventory JSON.
3. Generate coverage markdown table.
4. Compare inventory with existing SDK services.
5. Add scheduled coverage drift workflow.
6. Document how to add a new official endpoint.

## Milestone: `v1.3.1 Fixture and smoke baseline`

建议 issues：

1. Create fixture directory structure.
2. Add decode tests for core official endpoints.
3. Add fixture tests for Web storefront endpoints.
4. Add raw subtree validation tests.
5. Add golden tests for redaction and asset URL helpers.
6. Add opt-in live smoke documentation.
7. Add live smoke report template.

## Milestone: `v1.3.2 Doctor command`

建议 issues：

1. Add `examples/doctor` skeleton.
2. Add official API reachability checks.
3. Add Storefront/Community/Market checks.
4. Add proxy diagnostics.
5. Add credential presence checks without printing secrets.
6. Add JSON output mode.
7. Add doctor cookbook.

## Milestone: `v1.3.3 High-value helpers`

建议 issues：

1. Add reviews paginator.
2. Add inventory paginator.
3. Add batch app details helper.
4. Add batch market price overview helper.
5. Add lightweight request observer hook.
6. Add examples for batch and paginator usage.
7. Add tests for cancellation and partial failures.

---

# 7. Release gating 标准

每个 release 前建议满足以下标准。

## 7.1 通用标准

- `go test ./...` 通过。
- `go test -race ./...` 通过。
- `go vet ./...` 通过。
- `staticcheck ./...` 通过。
- `govulncheck ./...` 通过。
- README 示例能复制运行。
- 新增 public API 有示例或文档。
- 新增 Web surface 有 volatile/unofficial 边界说明。
- 新增 addon 有 “what it does / what it does not do”。
- 不打印 secret。
- release notes 已更新。

## 7.2 `v1.2.x` release gate

额外要求：

- 不引入不必要 breaking change。
- 兼容性策略文档更新。
- 开源基础文件齐全。
- CI 不依赖非固定 latest 工具链作为主线阻塞项。
- redaction 示例覆盖生产日志场景。

## 7.3 `v1.3.x` release gate

额外要求：

- 自动生成 coverage 报告。
- fixture decode test 覆盖核心 endpoint。
- live smoke 明确 opt-in。
- doctor 输出不泄露 secret。
- helper 不引入账号自动化行为。
- batch/paginator helper 支持 context cancellation。

---

# 8. 风险清单与应对策略

## 8.1 Steam 上游 payload 漂移

风险：typed struct decode 失败或字段含义变化。

应对：

- fixture corpus。
- live smoke。
- 对 volatile subtree 使用 `json.RawMessage`。
- coverage/drift report。

## 8.2 Web surface 不稳定

风险：Store / Community / Market endpoint 行为变化、区域差异、登录态差异、反爬或 block。

应对：

- 坚持只读、小面、明确 volatile。
- block detection 不伪装成成功。
- doctor 工具帮助用户判断网络与代理问题。
- 不把 browser fallback 放进核心稳定面。

## 8.3 凭据泄露

风险：Steam credentials 常出现在 query、cookie、proxy URL、日志、错误 body 中。

应对：

- redaction helpers。
- credential safety docs。
- examples 使用 env 或 hidden prompt。
- tests 覆盖敏感字段。
- doctor 默认不打印 secret。

## 8.4 CI 不稳定

风险：latest staticcheck/govulncheck 或 Go 工具链变化导致随机失败。

应对：

- 主线工具版本 pin。
- scheduled latest job 只做预警。
- release checklist 固化。

## 8.5 版本节奏过快

风险：用户对 minor 版本感到频繁变化，维护压力变大。

应对：

- `v1.2.x` 聚焦质量。
- `v1.3.x` 聚焦自动化和工具。
- 不把每个功能拆成新的 minor。
- 大部分改进通过 patch release 交付。

---

# 9. 建议的工作顺序

## 第一阶段：1 到 2 个周末可完成

优先做：

1. `SECURITY.md`
2. `CONTRIBUTING.md`
3. issue templates / PR template
4. release checklist
5. CI 工具 pin
6. README 收敛第一版
7. `.idea` / `.gitignore` 检查

这个阶段产出 `v1.2.1`。

## 第二阶段：2 到 4 个周末可完成

优先做：

1. cookbook 文档。
2. pkg.go.dev examples。
3. redaction 增强。
4. credential safety 文档。
5. API diff 检查原型。
6. Go 最低版本评估。

这个阶段产出 `v1.2.2` 和 `v1.2.3`。

## 第三阶段：较完整的 `v1.3.0`

优先做：

1. `steamapi-sync` 原型。
2. coverage generated markdown。
3. coverage generated JSON。
4. scheduled drift check。
5. “如何新增 endpoint” 文档。

这个阶段产出 `v1.3.0`。

## 第四阶段：`v1.3.x` 深化

优先做：

1. fixture corpus。
2. decode regression tests。
3. live smoke report。
4. doctor command。
5. reviews/inventory paginator。
6. batch app details / market price helper。
7. lightweight observability hook。

这个阶段分多个 `v1.3.x` patch/minor patch 完成，不急着进入 `v1.4`。

---

# 10. 不急着做的 backlog

这些可以记录，但不建议放进 `v1.2.x → v1.3.x` 承诺范围：

- 完整 Store SDK。
- 完整 Community SDK。
- Browser-backed fallback。
- OpenTelemetry addon。
- Prometheus adapter。
- 更复杂的外部 proxy pool manager。
- 大规模 scraping 工具。
- 自动购买、出售、交易、批量领取。
- SteamGridDB 集成。
- GUI / Web dashboard。

---

# 11. 成功指标

可以用以下指标判断路线图是否有效。

## 11.1 工程指标

- CI 稳定通过率。
- release 前 checklist 完成率。
- public API accidental breaking change 数量。
- govulncheck/staticcheck 问题响应时间。
- fixture decode regression 数量。

## 11.2 维护指标

- Steam API drift 被发现的时间。
- coverage report 更新频率。
- 新 endpoint 从发现到 issue 的平均时间。
- live smoke 失败分类是否清晰。

## 11.3 用户采用指标

- README 到 Quick Start 的路径是否清晰。
- pkg.go.dev 示例数量。
- issue 中“如何使用”的重复问题是否下降。
- doctor 输出是否能减少代理/网络/凭据排查成本。
- stars / forks / external imports 是否增长。

---

# 12. 最终建议

`steam-go` 接下来最应该强化的是“长期可信维护能力”。

建议把路线图浓缩成一句执行原则：

> `v1.2.x` 让项目可信，`v1.3.x` 让项目可持续维护。

更具体地说：

- `v1.2.x` 不急着做新 endpoint，先把仓库治理、CI、文档、安全、兼容性保护做好。
- `v1.3.x` 不急着做大而全 Store/Community SDK，先做 API drift detection、fixture corpus、doctor command 和少量高频只读 helper。
- 所有账号相关、交易相关、购买相关、浏览器自动化相关能力都应保持 addon 化、显式 opt-in、强边界文档。
- 继续坚持 typed outer + raw volatile subtree 的 payload 策略。
- 继续把 request-control layer 作为核心差异化能力，而不是隐藏在内部实现里。

这条路线能让 `steam-go` 在不快速膨胀版本号的前提下，从 `v1.2.x` 稳健推进到 `v1.3.x`，同时为后续更大的功能扩展打好基础。
