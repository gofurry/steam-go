# 路线图

本文档用于记录 `steam-go` 后续版本的推进顺序与工程改进主线。

## 当前状态

- `v1.0.0-alpha-1`：已完成基础 `Client`、`client.API.*` 分组与请求骨架
- `v1.0.0-alpha-2`：已发布，完成首批服务覆盖与请求层加固
- `v1.0.0-alpha-3`：已完成，补齐了公开商店页相关的策略隔离、请求控制、代理能力与稳定性骨架

## 下一阶段：pre-v1

重点方向：从“能力已具备”转向“可长期维护、可稳定依赖、可清晰发布”。

### 工程质量与测试体系

- 巩固 CI，持续执行 `test`、`race`、`vet`、`staticcheck`、`govulncheck`
- 建立更清晰的 contract tests 分层
- 区分常规测试与 live API workflow
- 收敛超大测试文件、命名不清晰和目录职责边界问题

### 文档与示例治理

- 整理 `examples/`、`test/`、未来 `examples/live/` 的职责边界
- 增加 endpoint stability 文档
- 增加 endpoint coverage 文档
- 增加 compatibility 文档
- 增加更系统的 godoc examples

### 发布与兼容性治理

- 增加 `CHANGELOG.md`
- 明确 public API 兼容性策略
- 对易变 payload、raw/typed 边界、addons 边界继续补文档
- 在 pre-v1 阶段尽量减少无必要的 public API 语义波动

### 接口覆盖扩展

- 优先补齐官方公开、稳定、低权限接口
- 再逐步推进更高权限或更易变接口
- 对公开商店页能力保持“明确隔离、渐进接入”的策略

## v1.0.0 发布门槛

- 核心 public API 命名与行为基本冻结
- README 与核心文档体系完整
- endpoint stability / coverage 文档可用
- live workflow 与示例结构稳定
- CI 与基础质量检查齐备
- 官方稳定接口覆盖达到可正式依赖水平
