# Go 代码审计报告

## 审计范围

本轮回访聚焦于 2026-05-15 当天新增的认证与网页登录相关能力，以及 2026-05-17 对审计项的修复结果，主要范围包括：

- `api/authenticationservice`
- `addons/websession`
- `addons/freeclaim`
- `examples/websession`
- `examples/freeclaim`

## 当前结论

- P0：0
- P1：0
- P2：0
- P3：0

当前未发现会直接导致远程代码执行、敏感 token 明文落盘、任意请求伪造之类的高危问题。上轮审计中关于示例 secrets 输入、Store 登录态弱校验、示例轮询取消响应的 3 个问题已完成修复；本轮又补上了 addon Web 流量与根 SDK traffic-policy 执行链的统一接入，因此当前没有未修复的 P2/P3 项。

## 已修复项

### 1. 示例程序不再通过命令行参数直接接收高敏感 secrets

已完成修复。

- `examples/websession` 已移除 `-password`、`-guard-code`
- `examples/freeclaim` 已移除 `-refresh-token`
- 高敏感值现在统一优先读取环境变量，环境变量缺失时仅在交互终端里使用隐藏输入

修复效果：

- 避免把密码、Guard Code、Refresh Token 暴露到 shell history、进程列表和终端录屏中
- 保留了本地手动调试的可用性，不强制调用方只走环境变量

### 2. `ValidateStoreSession` 不再依赖正文包含 `"login"` 的弱启发式

已完成修复。

- `addons/websession/validate.go` 现在只依赖：
  - 响应状态为 2xx
  - 最终跳转目标没有落到 `/login` 相关路径

修复效果：

- 避免把“login history”“login guard”之类的正常页面文案误判成未登录
- 降低对 Store HTML 文案和页面实验的耦合

### 3. `examples/websession` 轮询等待已支持上下文取消

已完成修复。

- `examples/websession` 不再直接 `time.Sleep(...)`
- 轮询等待窗口现在会及时响应 `ctx.Done()`

修复效果：

- 手动调试时中断更及时
- 轮询间隔较长时不会出现“明明取消了但还要等一整个 sleep 周期”的体验问题

## 本轮补充修复

### 4. `websession` 与 `freeclaim` 已统一接入 SDK traffic-policy 执行链

已完成修复。

- 根包新增了 `(*Client).DoRawHTTPRequest(...)`
- 新增公开类型：
  - `RawHTTPRequestOptions`
  - `RawHTTPBlockResult`
  - `RawHTTPResult`
- `addons/websession.NewClientFromSteamClient(...)` 与 `addons/freeclaim.NewClientFromSteamClient(...)` 已接入根 SDK 的按类别执行链
- `WithHTTPClient(...)` 在 `FromSteamClient` 路径上会直接报配置错误，避免手动模式与 SDK 模式混用
- `internal/request` 的短缓存现在会保存完整 raw 响应元数据，因此 addon raw HTTP 流量同样可以复用 short cache / revalidation

修复效果：

- addon Web 流量可以继承主线 SDK 的 per-class retry / backoff、block detection、header profile、referer selector、short cache / revalidation
- `websession` 的 Community / Store 链路与 `freeclaim` 的 Store 链路可以和 `client.Web.*` 一样按 traffic class 统一调优
- 文档边界也同步收敛成两种模式：
  - `NewClientFromSteamClient(...)`：推荐方式，复用根 SDK 执行链
  - 旧 `NewClient(...)`：手动模式，保留给调用方自己管理 `http.Client`

## 验证结果

本轮修复后已通过：

- `go test ./examples/websession ./examples/freeclaim ./addons/websession ./addons/freeclaim`
- `go test ./...`
- `go vet ./...`

## 结论

这轮修复已经把今天审计范围内的关键问题都收掉了：高敏感 secrets 输入方式更安全，Store 登录态校验更稳，示例轮询的取消响应更合理，addon Web 流量也已经纳入根 SDK 的按类别执行链。当前审计结论可以收敛到 `P0=0 / P1=0 / P2=0 / P3=0`。
