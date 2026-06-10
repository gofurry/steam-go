# steam-go Roadmap

> 目标：保持 `steam-go` 的版本节奏克制、边界清晰、可维护。
>
> 当前基线：`v1.3.3` 已完成 `addons/vdf`、bounded reviews collector、inventory description join，并继续保持只读 Web helper 与 addon 分层。

---

## 1. Product Boundary

`steam-go` 继续定位为：

> A stable Go SDK for the official Steam Web API, with practical request controls and carefully scoped read-only Steam Web helpers.

中文表述：

> 面向官方 Steam Web API 的稳定 Go SDK，提供生产可用的请求控制能力，并谨慎扩展少量高价值、只读的 Steam Web 辅助能力。

核心边界：

- `client.API.*` 是主稳定契约。
- `client.Web.*` 只做少量高价值只读 JSON helper。
- addons 承接不适合进入核心包的扩展能力。
- 维护自动化、fixture、doctor、redaction、traffic policy、request observer 是长期维护能力。
- 不为了 coverage 数字盲目补 endpoint。

明确不做：

- 不做完整 Store SDK 或完整 Community SDK。
- 不在核心包内置 browser automation / browser fallback。
- 不做自动购买、出售、交易、批量领取等账号自动化能力。
- 不强行 typed 化高波动 payload。
- 不让 live smoke 成为普通 CI 硬门禁。
- 不把 OpenTelemetry、Prometheus 等重依赖放入核心包。

---

## 2. Current Position

当前 `v1.3.x` 已经覆盖：

- 官方 Steam Web API 的稳定 typed service surface。
- 只读 Web helper：Storefront、Community inventory、Market JSON。
- 高价值 helper：reviews pagination / bounded collection、inventory description join、batch app details、batch market price。
- addon：OpenID、websession、freeclaim、assets、markup、vdf、A2S。
- 请求控制：timeout、retry、rate limit、proxy、traffic policy、request observer、safe defaults。
- 维护能力：API coverage automation、fixture corpus、opt-in live smoke、doctor、release checklist、bilingual docs。

下一阶段继续优先：

- 小步兼容增强。
- 文档和示例质量。
- 测试与 release hygiene。
- 明确边界的只读 helper。
- 不扩大账号自动化和高风险 Web surface。

---

## 3. Version Plan

### `v1.3.4` - Stabilization and Hygiene

**Status:** Planned

**Scope:** Stability / Security-Safety / Observability / Testing / Documentation / Maintenance
**Goal:** 提升 request、raw HTTP、observer、redaction、diagnostics 和生产使用文档的可靠性，不扩大高风险账号自动化或不稳定 Web surface。

#### Focus

- 收紧 retry、raw HTTP 和 redaction 的安全边界。
- 补齐 observer、live smoke、doctor JSON schema 等诊断能力。
- 建立 coverage drift 人工 triage 闭环，避免为了 coverage 数字盲目扩张 endpoint。
- 补强 production cookbook 和 addon safety 文档。

#### Non-goals

- 不为了 coverage 数字补齐所有 missing endpoint。
- 不把 Web helper 扩展成完整 Store SDK 或完整 Community SDK。
- 不在 core 中加入 browser fallback / browser automation。
- 不新增自动购买、自动出售、交易、批量账号操作或批量领取能力。
- 不让 live smoke 成为默认 CI 硬门禁。
- 不把 OpenTelemetry、Prometheus 等重依赖放入核心包。
- 不将高波动 HTML / Web payload 强行 typed 化为稳定契约。

#### P1 Tasks

- [x] 收紧 retry 幂等性语义：GET / HEAD / OPTIONS 默认可重试，POST / PUT / PATCH / DELETE 默认不自动重试，需要重试时显式标注。
- [x] 审视现有 POST / mutation-like request 的 retry 行为，并补充 GET 5xx retry、POST 5xx no retry、POST explicit retry、context canceled no retry、429 credential rotation 等测试。
- [x] 为 `DoRawHTTPRequest` 增加 raw HTTP host safety 文档，并评估兼容性的可选 host allowlist / Steam host policy。
- [x] 明确 raw HTTP 不应直接消费不可信 URL，addon 示例避免鼓励 SSRF 风险用法。

#### P2 Tasks

- [x] 为 conditional cache 304 refresh 补 request observer 事件，至少保证 cache hit / conditional refresh 可被统计。
- [x] 简化 raw retry 的 `GetBody` 生命周期：有 `GetBody` 时不提前打开 unused body，每次执行前再获取新的 body。
- [x] 为 malformed URL redaction 增加 best-effort fallback，覆盖 API key、access token、refresh token、session id、Steam cookie / token 风格字段。
- [x] 统一 SteamID64 校验内部 helper，覆盖空值、空格、非数字、超大数和合法 SteamID64。
- [x] 新增或完善 `docs/api/coverage-triage.md`，分类 missing、version_mismatch、extra_sdk，并标注 P1/P2 candidate、deferred、won't add。
- [x] 新增 doctor JSON schema 文档，说明字段、status enum、exit code、redaction guarantee 和示例输出。
- [x] 为 opt-in live smoke 增加 redacted summary report 能力或文档，明确 skipped reason、网络依赖和 upstream 波动边界。

#### P3 Tasks

- [x] 新增 observer cookbook，说明同步回调边界、异步转发建议、统计维度和 secret-safe 事件字段。
- [x] 新增 traffic policy cookbook，覆盖 official API、Storefront Web、Community Web、Market Web、代理、host/session concurrency 和 cache TTL 边界。
- [x] 扩展 bounded Web helper cookbook，说明 reviews collector、inventory pagination、batch helpers、market helper 的请求规模限制建议。
- [x] 新增或扩展 addon safety 文档，明确 openid、websession、freeclaim、markup、vdf、assets 的安全边界。

#### Acceptance Criteria

- [ ] 不引入 breaking change；新增行为保持内部、兼容性加法或 opt-in。
- [ ] Retry 默认行为对非幂等请求更保守，并有单元测试覆盖核心分支。
- [ ] Raw HTTP SSRF / untrusted URL 风险有明确文档和可测试的可选限制策略。
- [ ] Redaction 在 URL parse 失败时不会原样泄露明显 token。
- [ ] Observer 事件不包含 query、header、body、cookie、credential 等敏感信息。
- [ ] Coverage triage 文档能指导后续 endpoint 选择，而不是追逐 coverage 数字。
- [ ] Doctor、live smoke、cookbook、addon safety 文档与 release notes 同步。
- [ ] `go test ./...`、`go test -race ./...`、`go vet ./...`、`staticcheck ./...`、`govulncheck ./...` 和 API compatibility check 在发布前通过或有明确说明。

---

## 4. Release Gate

每个 release 前至少满足：

- [ ] `go test ./...`
- [ ] `go test -race ./...`
- [ ] `go vet ./...`
- [ ] `go run honnef.co/go/tools/cmd/staticcheck@v0.7.0 ./...`
- [ ] `go1.26.4 run golang.org/x/vuln/cmd/govulncheck@v1.3.0 ./...`
- [ ] `go run ./internal/tools/apidiffcheck -base <previous-tag> -incompatible`
- [ ] docs markdown links resolve
- [ ] staged secret scan clean
- [ ] README / docs / cookbook / release notes 中英文同步
- [ ] 新 public API 有 examples 或 cookbook
- [ ] 新 Web surface 标明 unofficial / volatile / read-only 边界
- [ ] doctor、observer、logs、examples 不打印 secret

发布后额外验证：

- [ ] GitHub tag 已推送。
- [ ] GitHub Release 已发布。
- [ ] `go list -m -versions github.com/gofurry/steam-go` 能看到新版本。
- [ ] `go get github.com/gofurry/steam-go@<version>` 正常。
- [ ] pkg.go.dev 展示最新版本。
