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

### `v1.3.4` - TBD

**Status:** Planned / TBD

**Scope:** TBD
**Goal:** 待定；只在有明确用户需求、边界、测试计划和文档入口后进入实现。

#### Candidate Rules

- 候选必须是兼容性加法。
- 候选必须有明确用户场景，不为覆盖率或想象需求扩张。
- 新 Web helper 必须只读，并明确 unofficial / volatile 边界。
- 新 paginator / collector / batch helper 必须显式限制请求规模。
- 新 addon 必须保持独立边界，不污染 root `Client` 依赖。
- 不新增账号自动化、购买、交易、出售、browser fallback 或绕过 upstream 限制的能力。

#### Tasks

- [ ] 收集真实需求或维护痛点。
- [ ] 选择 1 个小主题进入 `v1.3.4`，或保持待定。
- [ ] 为选中主题补测试、文档和 release notes。

#### Acceptance Criteria

- [ ] 主题边界清楚。
- [ ] 测试覆盖核心行为和边界条件。
- [ ] README / docs / cookbook / release notes 中英文同步。
- [ ] 不引入 breaking change。
- [ ] 不引入未审视的重依赖。

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
