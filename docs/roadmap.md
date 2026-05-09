# 路线图

这个文档用于记录 `steam-go` 的版本演进方向与工程改进主线。它是主阅读入口，强调里程碑、长期方向和发布成熟度；更细的审计问题拆解与改进依据保留在独立审计路线图文档中。

## v1.0.0-alpha-1

- 建立根级 `Client` 与 `client.API.*` 分组服务结构
- 保持请求执行层轻量、可复用
- 支持 API key、access token、timeout、retry、rate limit、proxy 注入
- 将 addon 能力与核心 Web API 客户端解耦

## v1.0.0-alpha-2

- 扩展 `PlayerService` 覆盖范围
- 增加 `WishlistService`，支持愿望单列表、数量、打折条目查询
- 强化请求层的重试、响应体大小限制、代理 helper 默认行为
- 补充新增服务与请求层行为的针对性测试

## v1.0.0-alpha-3

重点方向：为后续接入部分 Steam 商店公开页面做稳定性与抗风控准备。这类流量更接近浏览器访问行为，而不是官方开放 Web API 的标准调用模式。

### 代理与路由能力

- 增加代理健康检查、失败评分与临时冷却能力
- 支持混合质量代理池的加权轮转
- 支持会话粘性路由，保证同一会话尽量固定走同一代理
- 按 API host、商店 host、OpenID 流程、公开页面 path 做路由级代理选择
- 提供每个代理的成功率、延迟、重试次数、拦截率等观测指标

### 请求画像与会话能力

- 官方 API 流量继续保持简单的 API key / access token 轮转
- 面向商店公开页面，增加浏览器式请求画像能力
- 提供少量稳定的 header profile 预设，而不是随机化单个 header
- 支持固定、路由式、上下文生成式 `Referer` 策略
- 支持会话级 cookie jar 与基础会话保持能力

### 稳定性与抗风控能力

- 按 host 和会话维度做自适应限速
- 针对高批量列表/详情抓取做有界并发控制
- 优化带抖动重试、冷却与 block 检测策略
- 在上游支持时使用 `ETag` / `If-Modified-Since` 条件请求
- 为重复公开页面请求增加短时缓存
- 为未来可能接入的自定义 TLS / 浏览器回退抓取策略预留 transport 扩展点

### alpha-3 明确不做的事情

- 不内置重量级通用爬虫框架
- 不内置挑战页绕过逻辑
- 不引入不透明的默认行为去干扰官方 API 调用语义
- 不把整个 SDK 绑定到单一反爬服务商或浏览器自动化方案上

## pre-v1 成熟路径

### 核心能力稳定化

- 增强限流接口表达能力，支持比单一 RPS 更灵活的配置方式
- 持续打磨 retry / backoff 语义与限流配合方式
- 为错误体提供安全预览能力，降低生产日志误用风险
- 为不同接口标注稳定性等级，区分官方公开、认证接口、易变商店接口和 addon 能力

### 工程质量增强

- 将 CI 提升到 `test`、`race`、`vet`、`gofmt`、`govulncheck`、`staticcheck` 级别
- 逐步建立更清晰的 contract tests 分层
- 将 live API 验证从常规测试路径中拆分出来，保留手动或独立 workflow 入口
- 持续收敛超大测试文件与边界不清晰的测试组织方式

### 文档与可信度增强

- 拆分认证、限流与重试、错误、代理、live examples、兼容性等主题文档
- 增加 endpoint stability 与 endpoint coverage 文档
- 增加 godoc examples，提高 pkg.go.dev 可读性
- 明确 OpenID、API key、access token 以及不同上游接口类型之间的边界

### 示例与目录结构整理

- 将 `test/` 中偏真实调用的示例逐步迁移到 `examples/live/`
- 保持基础示例、认证示例、代理示例、addon 示例、live 示例职责分离
- 为需要真实凭证的示例提供更清晰的运行方式说明

### Payload 与兼容性策略

- 对稳定 payload 优先提供 typed response
- 对字段庞大且高频变化的 payload 使用 `json.RawMessage` 或类似保守策略
- 增加 `CHANGELOG.md` 与兼容性策略文档
- 在 pre-v1 阶段尽量减少无必要的 public API 语义波动

## v1.0.0 发布门槛

- CI 覆盖 `test`、`race`、`vet`、`gofmt`，并补齐基础漏洞与静态检查能力
- README 与核心文档结构清晰，认证、错误、限流、代理、addons、live examples 有独立说明
- endpoint stability 与 endpoint coverage 文档可用
- `examples/live/` 目录和真实调用 workflow 结构稳定
- `CHANGELOG` 与兼容性策略文档可用
- 核心 public API 命名和行为基本冻结，可作为正式依赖使用
