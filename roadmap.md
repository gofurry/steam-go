# steam-go Roadmap

## 1. 当前状态概览

`steam-go` 当前已经完成从旧仓库 `gf-steam-sdk` 向“轻量、聚焦的 Steam Web API SDK”方向的主体重构。

从 [`steam-go-architecture-refactor-plan.md`](./steam-go-architecture-refactor-plan.md) 的执行阶段来看，可以这样判断：

- `Phase 0 设计冻结`：已完成
- `Phase 1 仓库初始化`：已完成
- `Phase 2 核心请求链路打通`：已完成
- `Phase 3 首批 API 迁移`：已完成，并且已经补上了原计划里可后置的 `steamuserstats`
- `Phase 4 清理遗留设计`：基本完成
- `Phase 5 A2S Addons 接入`：未开始
- `Phase 6 发布前收尾`：部分完成

## 2. 完成度判断

### 2.1 按“原始重构总计划”估算

如果按整份计划书来算，包括 `addons/a2s`、OAuth 扩展预留、更多发布工程化内容在内：

> **整体完成度约为 80% 左右。**

这个判断的依据是：

- 核心架构、目录、请求链路、错误模型、Option 体系已经落地
- 核心 API 组已经覆盖 `steamuser`、`playerservice`、`steamnews`、`steamuserstats`
- 测试、README、examples 已经具备可用雏形
- 但 `addons/a2s`、OAuth 扩展点、CLI、更多 API 组、发布工程化收尾还没有系统推进

### 2.2 按“v1 核心目标”估算

如果只看计划里最核心的 v1 目标，也就是“做一个干净、稳定、长期可维护的 Steam Web API SDK”：

> **v1 核心目标完成度约为 90% 以上。**

原因是：

- 核心定位已经明确并落地
- 非目标内容没有继续回流进主仓库
- 目录结构和 API 分组已经与 Steam 官方命名体系对齐
- typed/raw 双层返回、轻量 transport、可选/轮转 key、基础 failover 都已经具备

## 3. 已完成内容

### 3.1 核心架构

- 统一 `Client` 入口已完成
- `Option` 配置体系已完成
- 轻量 `Logger` 接口已完成
- `internal/request`、`internal/transport`、`internal/auth`、`internal/errors`、`internal/endpoint` 已完成
- typed response + raw response 双层返回已完成

### 3.2 首批核心 API

- `SteamUser`
  - `GetPlayerSummaries`
- `PlayerService`
  - `GetOwnedGames`
- `SteamNews`
  - `GetNewsForApp`
- `SteamUserStats`
  - `GetPlayerAchievements`

### 3.3 认证与请求稳定性

- API key 现在支持“不填写 key”
- 支持 `WithAPIKey(...)`
- 支持 `WithAPIKeys(...)` 做轮转
- 支持自定义 `APIKeyProvider`
- 已支持 `401/429` 自动切换下一个 key 并重试
- `5xx` 与网络错误重试已具备

### 3.4 文档与测试

- README 已经成型
- 每个核心 API 组已有 example
- `go test ./...` 通过
- 已覆盖请求构建、query 编码、错误映射、key failover、raw/typed 解码等关键路径

## 4. 未完成内容

### 4.1 计划中尚未推进

- `addons/a2s`
- OAuth 扩展点的正式设计与实现
- CLI
- 更多 Steam Web API 组
- 更完整的 rate limit / retry 策略能力
- 发布前工程化收尾
  - changelog / release notes 规范
  - GitHub Topics
  - 更完整的版本发布策略

### 4.2 尚可继续优化的点

- 将 `README` 从“可用”提升到“对外发布级别”
- 补更多真实文档示例和错误处理示例
- 增加更多回归测试和边界测试
- 梳理未来 OAuth / token source / credential injector 的演进接口

## 5. 后续路线图

## v2 已完成

- 补齐 `steamuserstats`
- 完成 key failover
- 保持现有公开 API 基本兼容

## v2.1 核心增强

目标：继续补强 Steam Web API 主线，而不扩散边界。

- 增加 1~2 组高价值官方 API
  - 候选：`steamapps`
  - 候选：`steameconomy`
- 补更多 endpoint registry 常量
- 扩充 typed/raw 测试覆盖
- 增加对 `401/429/5xx` 重试行为的文档说明

## v2.2 认证与传输增强

目标：让 SDK 在复杂环境下更稳，但仍保持轻量。

- 抽象更清晰的 auth provider 演进接口
- 为未来 OAuth 预留 token source / credential injector
- 细化 retry policy 与 rate limit policy
- 补充更多自定义 `ProxySelector` / `APIKeyProvider` 示例

## v3 扩展能力

目标：在核心稳定后，再进入非核心扩展。

- 设计并接入 `addons/a2s`
- 明确其与 Web API 主体的边界
- 评估是否拆为独立扩展模块

## 发布前里程碑

### M1 核心 SDK 稳定

- 现有 4 组 API 行为稳定
- 文档与测试补齐到可对外说明

### M2 首个公开可用版本

- README 达到对外发布标准
- 版本策略明确
- Release 说明和基础工程化补齐

### M3 扩展准备完成

- OAuth 演进方向明确
- A2S 扩展边界明确

## 6. 推荐下一步

如果继续沿着“最符合当前仓库主线”的方向推进，建议优先级如下：

1. 先把 `README`、错误处理说明、key failover 说明补成对外发布级别
2. 再补 1~2 组高价值官方 API，继续增强 Web API 覆盖
3. 等主线稳定后，再进入 `addons/a2s` 和 OAuth 扩展点设计

## 7. 一句话结论

`steam-go` 的**核心重构其实已经基本成功**：

- 如果按“原始总计划”算，**约完成 80%**
- 如果按“v1 核心目标”算，**约完成 90%+**

现在最适合的推进方式，不是重新大改结构，而是：

> **继续沿当前架构补核心 API、补文档、补发布质量，然后再做扩展层。**
