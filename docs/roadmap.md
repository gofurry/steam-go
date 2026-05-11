# 路线图

这个文档用于记录 `steam-go` 后续版本的推进顺序与工程改进主线。当前 `v1.0.0-alpha-2` 已经发布，因此后续规划统一从 `v1.0.0-alpha-3` 开始编排；更细的审计问题拆解与改进依据保留在独立审计路线图文档中。

## 当前状态

- `v1.0.0-alpha-1`：已完成基础 `Client`、`client.API.*` 分组与请求骨架
- `v1.0.0-alpha-2`：已发布，完成首批服务扩展与请求层加固
- 下一阶段主线：`v1.0.0-alpha-3`

## v1.0.0-alpha-3

重点方向：先补齐会直接影响下一阶段可执行性的基础能力，再推进公开商店页相关稳定性与抗风控能力。这类流量更接近浏览器访问行为，因此必须建立在稳定的请求控制、会话、代理和错误处理基础上。

### 基础前置能力

- [已完成] 增强 rate limit API 表达能力，避免只依赖单一 RPS 语义
- [已完成] 打磨 retry / backoff / cooldown 语义与限流配合方式
- [待开始] 增加 host / session 级并发与限速控制的扩展能力
- [已完成] 为错误体提供安全预览能力，降低生产日志误用风险
- [待开始] 明确易变 payload 的 `json.RawMessage` 使用策略

### 请求会话与代理基础

- [已完成] 支持会话级 cookie jar 与基础会话保持能力
- [已完成] 支持会话粘性代理，保证同一会话尽量固定走同一代理
- [已完成] 增加代理健康检查、失败评分与临时冷却能力
- [已完成] 增加健康代理池基础观测指标，如选中次数、成功/失败次数、冷却次数与当前冷却状态
- [已完成] 保持官方 API 流量与公开商店页流量在策略上的边界分离

### 商店公开页稳定性能力

- [待开始] 增加少量稳定的 header profile 预设
- [待开始] 支持固定、路由式、上下文生成式 `Referer` 策略
- [待开始] 增加条件请求与短时缓存能力
- [待开始] 增加 block 检测、429/403 识别与恢复策略
- [待开始] 为自定义 TLS / 浏览器回退方案预留 transport 扩展点

## pre-v1 收敛阶段

重点方向：在 `alpha-3` 的基础能力成型后，集中做工程治理、文档体系、示例结构和兼容性收敛，提升外部用户的依赖信心。

### 工程质量与测试体系

- 将 CI 升级到 `test`、`race`、`vet`、`gofmt`、`govulncheck`、`staticcheck` 级别
- 建立更清晰的 contract tests 分层
- 拆分 live API workflow 与常规测试路径
- 收敛测试命名、超大测试文件和目录职责边界

### 示例、目录与文档治理

- 将 `test/` 中偏真实调用的内容逐步迁移到 `examples/live/`
- 拆分认证、错误、代理、限流与重试、live examples、compatibility 等主题文档
- 增加 endpoint stability 与 endpoint coverage 文档
- 增加 godoc examples，提高 pkg.go.dev 可读性
- 明确 OpenID、API key、access token、addon 与不同接口类型之间的边界

### 兼容性与发布治理

- 增加 `CHANGELOG.md`
- 增加兼容性策略文档
- 在 pre-v1 阶段尽量减少无必要的 public API 语义波动
- 优先扩展官方公开、稳定、低权限接口，再逐步推进更高权限或更易变接口

## v1.0.0 发布门槛

- 核心 public API 命名和行为基本冻结
- README 与核心文档体系完整，认证、错误、限流、代理、addons、live examples 有独立说明
- endpoint stability 与 endpoint coverage 文档可用
- `examples/live/` 与 live workflow 结构稳定
- CI 与基础静态检查、漏洞检查能力齐备
- 官方稳定接口覆盖达到可正式依赖水平
