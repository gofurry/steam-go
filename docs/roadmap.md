# steam-go Roadmap

> 目标：在 `v1.3.0` 已完成治理、文档、诊断、API coverage automation、fixture/smoke、只读 Web helper 与 request observer 的基础上，继续放慢版本节奏，把维护体系稳定下来。
>
> 当前策略：`v1.3.1` 承接发布闭环与稳定化补丁；`v1.3.2` 只保留前瞻候选，不提前承诺新的大版本。

---

## 1. Roadmap 判断

当前 roadmap 的方向是合理的：`steam-go` 已经完成从“API 封装库”到“有治理、诊断、自动化和安全边界的 Go SDK”的关键跃迁，下一步不应该继续快速堆 endpoint 或扩大 Web surface。

需要调整的是版本节奏和篇幅：

- 之前拆出的多个稳定化阶段都属于 `v1.3.0` 之后的同一类收尾工作，拆得过细会让版本号走太快。
- 下一阶段功能扩展前瞻仍然有价值，但不应现在固定成新的 minor 版本承诺。
- 旧 roadmap 中大量解释性内容、候选细节和重复原则可以压缩成可执行清单。

结论：

> `v1.3.1` 做稳定化与发布闭环，`v1.3.2` 作为下一阶段前瞻候选。先让维护系统变稳，再决定是否扩功能。

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
- [ ] 新增 `docs/api/coverage-triage.md`，对 coverage diff 做第一版人工分类。
- [ ] 增强 coverage drift issue 输出：status counts、missing/version_mismatch/extra_sdk 摘要、triage checklist、artifact 链接。
- [ ] 为 coverage drift issue 自动打 `maintenance`、`steam-api-drift`、`needs-triage` 标签。
- [ ] 扩展 fixture corpus，优先覆盖 reviews cursor、inventory pagination、market `success=false`、Storefront 字段缺失/地区差异。
- [ ] 文档化 doctor JSON schema，明确字段、退出码、secret redaction 和脚本消费边界。
- [ ] 让 opt-in live smoke 输出可归档报告，包含 human summary、JSON report、skipped reason 和 redacted network info。
- [ ] 补 observability cookbook，说明同步 observer、异步 channel observer、panic-safe wrapper、metrics label 建议。
- [ ] 补 batch/paginator cookbook，强调 `MaxConcurrent` 不等于安全请求速率，建议配合 `WithSafeDefaults()` / `WithTrafficPolicy(...)`。
- [ ] 补 paginator edge-case tests：重复 cursor、空页、handler error、context cancellation、`MaxPages<0`。
- [ ] 增加 request observer 轻量 benchmark：no observer、no-op observer、counter observer。
- [x] 更新 release checklist，加入 GitHub Release、Go module proxy、pkg.go.dev 可见性检查。

### Acceptance Criteria

- [x] `v1.3.0` release notes 中英文可达。
- [x] `docs/code-audit.md` 有可审核内容，不再是占位文档。
- [ ] `coverage-triage.md` 至少覆盖 P1/P2 候选 endpoint。
- [ ] coverage drift issue 可以直接用于维护 triage。
- [ ] doctor JSON 输出有文档化结构。
- [ ] live smoke 仍为 opt-in，且输出不泄露 secret。
- [ ] observer、batch、paginator 的安全边界和性能边界写清楚。
- [ ] 不引入新的重依赖。
- [ ] 不引入 breaking change。

### Notes

- `v1.3.1` 可以包含兼容的文档、测试、维护工具和小补丁。
- 除非修复 `v1.3.0` 暴露出来的问题，否则不主动新增 endpoint。
- Go module / pkg.go.dev 分发验证应在 tag 发布后完成并记录。

---

## 5. `v1.3.2` - Forward Look

**Status:** Planned / Candidate

**Scope:** Feature selection / Addon candidate / Diagnostics candidate / Official endpoint candidate
**Goal:** 只在 `v1.3.1` 稳定后，从经过 triage 的候选中选择 1 到 2 个主题推进，不做大爆炸版本。

### Candidate A: 精选官方 Endpoint 扩展

适合条件：`coverage-triage.md` 已经成熟，有明确 P1 endpoint。

选择标准：

- 官方 Steam Web API。
- 只读或低风险。
- 通用性强，不是极窄游戏特定接口。
- 认证边界清楚。
- 可用本地 fixture 测试。
- payload 稳定，或能采用 typed outer + `json.RawMessage`。

不选：

- mutating endpoint。
- partner / publisher sensitive endpoint。
- 需要特殊权限且无法稳定测试的 endpoint。
- purchase、trade、sell、bulk automation 相关 endpoint。

### Candidate B: Observability Addon or Cookbook Adapter

适合条件：真实用户需要接入 metrics/tracing。

方向：

- 优先 cookbook adapter。
- 如果需要代码，考虑 `addons/otel` 或 `addons/prommetrics`。
- addon 只消费 sanitized `RequestEvent`。
- 不把重依赖放进核心包。
- 控制 metrics label 基数。

### Candidate C: Doctor 产品化

适合条件：用户反馈集中在网络、代理、凭据、区域和上游可用性排查。

方向：

- 保留 `examples/doctor` 作为学习入口。
- 评估是否新增 `cmd/steam-go-doctor`。
- 固化 JSON schema。
- 支持输出 redacted report。
- 不支持自动登录、cookie 刷新或 browser fallback。

### Candidate D: 小幅 Web Helper 增强

适合条件：现有 paginator/batch helper 在真实使用中暴露明确重复需求。

可考虑：

- Reviews collector helper，但必须要求显式 `MaxPages` 或 `MaxReviews`。
- Inventory asset/description join helper。
- Storefront app details 的稳定字段补 typed。
- App details batch 合并请求，前提是 upstream 多 appids 行为稳定且测试覆盖明确。

### Acceptance Criteria

- [ ] 只选择 1 到 2 个主题进入实现。
- [ ] 每个主题都有明确边界、测试计划和文档入口。
- [ ] 新 endpoint 必须来自 `coverage-triage.md`。
- [ ] 新 helper 不默认无限抓取。
- [ ] 新 addon 不污染核心依赖。
- [ ] 不做账号自动化、不绕过 upstream 限制。

---

## 6. Release Gate

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

## 7. 进入 `v1.3.2` 的条件

只有满足下面条件，才开始推进 `v1.3.2` 候选：

- [ ] `v1.3.1` release closure 完成。
- [ ] `coverage-triage.md` 有可执行分类。
- [ ] doctor JSON schema 和 live smoke 报告稳定。
- [ ] observer、batch、paginator 的真实使用边界已文档化。
- [ ] 没有未处理的兼容性或 secret safety 问题。

---

## 8. 最终建议

`steam-go` 当前路线是合理的，但应该进一步收敛版本节奏。

建议执行原则：

> `v1.3.1` 先固定发布闭环和维护体系；`v1.3.2` 再从 triage 结果中选择少量明确主题。不要急着进入新的 minor 版本。

这样可以让 `steam-go` 在不快速膨胀版本号的前提下，继续保持可信、可维护、边界清晰。
