# steam-go Roadmap

> Roadmap 只保留下一阶段可执行计划；已完成内容进入 release notes、文档和代码历史。

---

## Current Position

`steam-go` 当前处于 v1.x 稳定演进阶段：根包继续作为 `github.com/gofurry/steam-go` 的公开 SDK façade，`api/`、`web/`、`addons/` 提供分组能力，`internal/` 已承接请求执行、传输、流量分类、认证和端点等实现细节。

v1.3.7 的下一步重点不是新增能力，而是把根包中实现偏重的 raw HTTP host policy、request profile、proxy selector 逻辑收敛到 `internal/`，同时保持所有公开 API、默认行为和导入路径稳定。

已完成的版本工作不在 roadmap 中保留，避免后续计划被历史任务淹没。需要回看已完成内容时，请查看：

- `docs/releases/`
- `docs/zh/releases/`
- Git commit history

---

## Roadmap Strategy

优先做受控的内部重构，先降低根包维护成本，再继续扩展新的 read-only 能力。v1.3.7 必须是 refactor-only PR：不改变用户导入方式、不删除导出符号、不改变方法签名、不调整默认运行时行为、不引入新外部依赖。

---

## Version Plan

### v1.3.7 - Root Package Internal Extraction

**Status:** In progress

**Scope:** Architecture / Developer-facing / Testing / Documentation

**Goal:** 在保持根包公开 API 稳定的前提下，将实现细节迁入 focused `internal/` packages，降低后续维护和扩展风险。

#### Focus

- 保持 `package steam` 作为稳定 public façade
- 抽离 raw HTTP host policy、request profile、proxy selector 实现
- 用根包 adapter 保留公开类型所有权和兼容性
- 补强行为冻结测试和 API compatibility 检查

#### Tasks

- [x] 新增 `internal/rawhttp`，承接 raw HTTP host policy 实现、host/suffix normalization 和组合 policy 逻辑
- [x] 保留根包 `RawHTTPRequestOptions`、`RawHTTPHostPolicy`、`RawHTTPResult`、`DoRawHTTPRequest` 和 public constructors
- [x] 新增 `internal/requestprofile`，承接 header profile application、referer selector、referer context source 和 normalization 逻辑
- [x] 保留根包 `HeaderProfile`、`RefererSelector`、`RefererRoute`、默认 profile constructors 和 `WithRefererSource`
- [x] 新增 `internal/proxyselector`，承接 static、round-robin、health-checked、sticky、routing proxy selector 实现
- [x] 保留根包 `ProxySelector`、`ProxyHealthConfig`、`ProxyMetricsSnapshot`、`ProxyEndpointMetrics`、`ProxyRoute` 和 public constructors
- [x] 避免 public root exported types alias 到 internal exported types；需要转换时使用 root adapter
- [x] 确保 internal packages 不 import 根包，避免 import cycle
- [x] 保持 `ErrAllProxiesCoolingDown` 的 `errors.Is` 语义对用户稳定
- [x] 增加或保留 raw HTTP host policy、request profile、proxy selector 的行为冻结测试
- [x] 新增 `docs/dev/internal-architecture.md`，说明 root façade 与 `internal/` 边界
- [x] 完成 public API compatibility 检查并记录结果

#### Acceptance Criteria

- 根包导入路径仍为 `github.com/gofurry/steam-go`
- 没有导出的根包类型、函数、方法、常量或变量被删除、重命名或迁移到用户可见的新路径
- 所有公开方法签名和 option 行为保持不变
- retry、cache、proxy、cookie、traffic class、observer、request profile 和 raw HTTP 行为保持不变
- proxy credentials 继续在 metrics 和展示文本中被 redacted
- raw HTTP host policy 不放宽现有安全校验语义
- request profile 继续保持 explicit header / referer 优先
- `go test ./...` 通过
- `go test -race ./...` 通过，或记录明确的外部限制原因
- `go vet ./...`、`staticcheck ./...`、`govulncheck ./...` 通过，或记录明确的外部限制原因
- API compatibility check 不报告 incompatible public API changes

#### Notes

`proxyselector` 是本次风险最高的抽离点。实现时应先冻结现有 sticky session、health cooldown、metrics redaction、403 block-detection 判定和 all-cooling-down error 语义，再移动状态机代码。
