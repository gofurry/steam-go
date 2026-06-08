# steam-go Roadmap

> 目标：在 `v1.3.0` 已完成治理、文档、诊断、API coverage automation、fixture/smoke、只读 Web helper 与 request observer 的基础上，继续放慢版本节奏，把维护体系稳定下来。
>
> 当前策略：`v1.3.2` 已作为小范围能力补丁落地 Store events 与 Steam 内容清洗；`v1.3.3` 固定为 Candidate D + `addons/vdf`，继续以兼容小步补丁推进，不提前承诺新的大版本。

---

## 1. Roadmap 判断

当前 roadmap 的方向是合理的：`steam-go` 已经完成从“API 封装库”到“有治理、诊断、自动化和安全边界的 Go SDK”的关键跃迁，下一步不应该继续快速堆 endpoint 或扩大 Web surface。

需要调整的是版本节奏和篇幅：

- 之前拆出的多个稳定化阶段都属于 `v1.3.0` 之后的同一类收尾工作，拆得过细会让版本号走太快。
- 下一阶段功能扩展前瞻仍然有价值，但不应现在固定成新的 minor 版本承诺。
- 旧 roadmap 中大量解释性内容、候选细节和重复原则可以压缩成可执行清单。

结论：

> `v1.3.1` 做稳定化与发布闭环，`v1.3.2` 补齐采集器急需的 Store events 与 markup 能力，`v1.3.3` 固定为 Candidate D + vdf addon，避免候选池继续发散。

---

## 2. 产品边界

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

## 3. 已完成基线

### `v1.2.x` - Trust and Adoption Foundation

**Status:** Completed

**Scope:** Documentation / Security / Compatibility / CI / Release

已完成：

- [x] 仓库治理文件：`SECURITY.md`、`CONTRIBUTING.md`、issue templates、PR template。
- [x] README 入口化，中英文 cookbook 补齐。
- [x] release checklist、兼容性策略、API diff 本地工具。
- [x] credential safety、redaction helper、error handling 文档。
- [x] pkg.go.dev examples。

### `v1.3.0` - Maintenance Automation and Adoption Helpers

**Status:** Completed

**Scope:** Developer-facing / Testing / Diagnostics / User-facing helpers / Documentation

已完成：

- [x] `internal/tools/steamapi-sync`：生成官方 API inventory、coverage Markdown/JSON 和 coverage diff。
- [x] scheduled coverage drift workflow：发现 drift 时开/更新 GitHub issue，不自动改代码。
- [x] fixture corpus、golden regression、opt-in live smoke baseline。
- [x] `examples/doctor`：网络、凭据、代理、official API、Storefront、Community、Market 诊断。
- [x] 高价值只读 helper：`ListAppReviews`、`ListInventory`、`GetAppDetailsBatch`、`GetPriceOverviewBatch`。
- [x] `WithRequestObserver`、`RequestObserverFunc`、脱敏 `RequestEvent`。
- [x] 中英文 docs、Web reference、cookbook、release checklist 同步。

---

## 4. `v1.3.1` - Stabilization and Release Closure

**Status:** In progress

**Scope:** Release / Documentation / Testing / Maintenance automation / Diagnostics
**Goal:** 不新增大功能，补齐 `v1.3.0` 的发布闭环、审计记录和维护自动化稳定化。

### Focus

- `v1.3.0` release closure
- coverage triage
- fixture / smoke / doctor 稳定化
- request observer 与 batch/paginator 使用边界
- 分发验证与文档一致性

### Tasks

- [x] 新增 `docs/releases/v1.3.0.md`。
- [x] 新增 `docs/zh/releases/v1.3.0.md`。
- [x] 更新 `docs/README.md` 与 `docs/zh/README.md` 的 release notes 索引。
- [x] 填充 `docs/code-audit.md`，记录 `v1.3.0` 审计范围、结论、风险和后续动作。
- [x] 新增 `docs/api/coverage-triage.md`，对 coverage diff 做第一版人工分类。
- [x] 增强 coverage drift issue 输出：status counts、missing/version_mismatch/extra_sdk 摘要、triage checklist、artifact 链接。
- [x] 为 coverage drift issue 自动打 `maintenance`、`steam-api-drift`、`needs-triage` 标签。
- [x] 扩展 fixture corpus，优先覆盖 reviews cursor、inventory pagination、market `success=false`、Storefront 字段缺失/地区差异。
- [x] 文档化 doctor JSON schema，明确字段、退出码、secret redaction 和脚本消费边界。
- [x] 让 opt-in live smoke 输出可归档报告，包含 human summary、JSON report、skipped reason 和 redacted network info。
- [x] 补 observability cookbook，说明同步 observer、异步 channel observer、panic-safe wrapper、metrics label 建议。
- [x] 补 batch/paginator cookbook，强调 `MaxConcurrent` 不等于安全请求速率，建议配合 `WithSafeDefaults()` / `WithTrafficPolicy(...)`。
- [x] 补 paginator edge-case tests：重复 cursor、空页、handler error、context cancellation、`MaxPages<0`。
- [x] 增加 request observer 轻量 benchmark：no observer、no-op observer、counter observer。
- [x] 更新 release checklist，加入 GitHub Release、Go module proxy、pkg.go.dev 可见性检查。

### Acceptance Criteria

- [x] `v1.3.0` release notes 中英文可达。
- [x] `docs/code-audit.md` 有可审核内容，不再是占位文档。
- [x] `coverage-triage.md` 至少覆盖 P1/P2 候选 endpoint。
- [x] coverage drift issue 可以直接用于维护 triage。
- [x] doctor JSON 输出有文档化结构。
- [x] live smoke 仍为 opt-in，且输出不泄露 secret。
- [x] observer、batch、paginator 的安全边界和性能边界写清楚。
- [x] 不引入新的重依赖。
- [x] 不引入 breaking change。

### Notes

- `v1.3.1` 可以包含兼容的文档、测试、维护工具和小补丁。
- 除非修复 `v1.3.0` 暴露出来的问题，否则不主动新增 endpoint。
- Go module / pkg.go.dev 分发验证应在 tag 发布后完成并记录。

---

## 5. `v1.3.2` - Store Events and Markup Utilities

**Status:** Completed

**Scope:** User-facing helper / Addon / Documentation / Testing
**Goal:** 为采集器和下游内容处理补齐 Steam Store events 读取、BBCode/HTML 清洗、纯文本摘要能力，同时保持核心 SDK 边界克制。

### Focus

- Store events JSON endpoint
- Steam BBCode / HTML content cleaning
- raw payload preservation for volatile fields
- bilingual docs and release notes

### Tasks

- [x] 新增 `Web.Storefront.GetAdjacentPartnerEvents` / `GetAdjacentPartnerEventsRaw`。
- [x] 为 Store events 提供稳定 typed subset，并为 event 与 announcement 保留 `json.RawMessage`。
- [x] 新增 `addons/markup`，支持 Steam BBCode 转 HTML、HTML sanitize、plain text、summary helper。
- [x] 支持 `{STEAM_CLAN_IMAGE}`、Steam list item closure `[/*]`、Steam escaped URL/text 等真实内容格式。
- [x] 新增 `examples/markup` 可运行示例。
- [x] 新增中英文 v1.3.2 release notes，并更新 docs 索引。
- [x] 增加单元测试和跨平台 golden 换行归一化。

### Acceptance Criteria

- [x] `go test ./...` 通过。
- [x] `go run ./examples/markup` 通过。
- [x] 新 public API 有文档入口。
- [x] 新 Web surface 仍保持 read-only，并保留 raw payload 应对上游字段波动。
- [x] `addons/markup` 不污染核心 client API。
- [x] 不引入账号自动化、购买、交易、绕过 upstream 限制等能力。

### Notes

- `v1.3.2` 是兼容性加法版本，不包含 breaking change。
- 原 roadmap 中的 forward-look 候选计划顺延到 `v1.3.3`。

---

## 6. `v1.3.3` - Candidate D + VDF Addon

**Status:** In progress

**Scope:** Addon / User-facing helper / Documentation / Testing
**Goal:** 固定 `v1.3.3` 为一个克制的兼容补丁：新增 `addons/vdf` 轻桥接，并只保留 Candidate D 的小幅 Web helper 增强空间。

### Focus

- `addons/vdf` text VDF / KeyValues bridge
- Candidate D 小幅 Web helper 增强
- addon scope / non-goals 文档化
- 最小但可验证的测试和示例

### Tasks

- [x] 引入 `github.com/gofurry/vdf-go v1.0.0`。
- [x] 新增 `addons/vdf` 包，re-export `vdf-go` 的核心类型、parse option、encode option、parse/marshal/write 函数。
- [x] 为 `addons/vdf` 增加 GoDoc，明确只处理文本 VDF / KeyValues。
- [x] 增加 `addons/vdf` bridge tests 和 example tests。
- [x] 增加 `examples/vdf` 可运行示例。
- [x] 更新 root README、中文 README 和中英文 addon reference。
- [ ] 评估 Candidate D 的唯一小幅 Web helper 主题，优先从 reviews collector、inventory join、Storefront typed field 补齐中选择一个。
- [ ] Candidate D helper 必须显式限制翻页或批量范围，不允许默认无限抓取。
- [ ] 新增或更新 `docs/releases/v1.3.3.md` 与 `docs/zh/releases/v1.3.3.md`。

### Acceptance Criteria

- [x] `go test ./...` 通过。
- [x] `go run ./examples/vdf -file <caller-provided-vdf> -key <top-level-key>` 可运行。
- [x] `addons/vdf` 不重新实现 parser，只桥接 `vdf-go`。
- [x] `addons/vdf` 不自动扫描 Steam 安装目录。
- [x] `addons/vdf` 不读取账号、token、cookie、session。
- [x] `addons/vdf` 不承诺 binary VDF 或 `shortcuts.vdf`。
- [ ] Candidate D helper 若进入本版本，必须有测试、文档和明确上限。
- [ ] 不新增账号自动化、购买、交易、绕过 upstream 限制等能力。

### Notes

- 原 Candidate A/B/C 不进入 `v1.3.3`，后续如果有真实需求再重新排期。
- `addons/vdf` 是对独立稳定库的轻桥接，不把 `steam-go` 扩展成本地 Steam 扫描工具。
- 如果 Candidate D 在实现前没有足够明确的真实需求，`v1.3.3` 可以只发布 `addons/vdf` 与相关文档。

---

## 7. Release Gate

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

---

## 8. `v1.3.3` 执行约束

`v1.3.3` 已进入 Candidate D + vdf addon 执行态，后续只按下面约束继续推进：

- [x] 候选范围从 A/B/C/D 收敛为 Candidate D + `addons/vdf`。
- [x] vdf 能力必须保持 addon 化，不进入 root `Client` 或 `client.Web.*`。
- [x] vdf 能力必须复用 `vdf-go`，不在本仓库重写 parser。
- [ ] Candidate D 只能选择一个小幅 Web helper 主题，不能重新打开大候选池。
- [ ] Candidate D helper 必须有显式分页、批量或请求上限。
- [ ] 没有未处理的兼容性或 secret safety 问题。

---

## 9. 最终建议

`steam-go` 当前路线是合理的，但应该进一步收敛版本节奏。

建议执行原则：

> `v1.3.2` 先完成 Store events 与 markup 这类明确、低风险、采集器依赖的补丁；`v1.3.3` 固定为 Candidate D + vdf addon。不要重新打开大候选池，也不要急着进入新的 minor 版本。

这样可以让 `steam-go` 在不快速膨胀版本号的前提下，继续保持可信、可维护、边界清晰。
