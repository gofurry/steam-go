# 路线图

## 当前状态

`steam-go` 已完成 `v1.0.0` 正式版发布前的主要准备工作，并在 `v1.1.0` 阶段补齐了 `client.Web.*` 非官方只读 Web JSON 能力：

- 稳定 public API 边界已定义
- 兼容性、endpoint 覆盖、endpoint 稳定性文档已整理
- 真实 live smoke 入口已统一到 `examples/live/`
- `client.API.*` 继续承载官方 Steam Web API
- `client.Web.*` 已承载 Storefront、Community、Market 的只读 Web JSON 能力

下一阶段计划吸收 `steam-auth-web-session-poc-roadmap.md` 中已验证闭环的认证会话与免费领取经验，但仍保持主线 SDK、Web 会话 addon、免费领取 addon 的边界清晰。

## 路线图策略

`v1.2.0` 的优先级是先把可复用的官方认证原子接口放入主线，再把登录态编排和免费领取链路放入 addon。这样可以让 `client.API.*` 保持官方 API SDK 定位，同时让更脆弱、更业务化的网页登录与领取流程以可选模块演进。

实现时需要参考 `steam-auth-web-session-poc-roadmap.md` 的设计草案和接入经验，尤其是 protobuf form 编码、Steam Guard 分支、refresh token 换 Web Cookie、Store/Community session 校验、免费包解析和领取成功判定。

## 版本计划

### v1.2.0 - Auth Web Session and Free Claim Addons

**Status:** In progress
**Scope:** User-facing / Developer-facing / Stability / Testing / Documentation  
**Goal:** 将已验证的认证会话与免费领取闭环拆成可维护的主线原子能力和可选 addon 能力。

#### Focus

- 官方 `IAuthenticationService` 原子接口
- Web session 登录态编排
- refresh token 到 Store/Community Web Cookie 的转换
- 免费促销搜索、SubID 解析与单个免费 license 领取
- 错误模型、流量策略、示例和文档

#### Tasks

- [x] 新增 `api/authenticationservice`，接入 `client.API.AuthenticationService`
- [x] 实现 `GetPasswordRSAPublicKey`
- [x] 实现 `BeginAuthSessionViaCredentials`
- [x] 实现 `BeginAuthSessionViaQR`
- [x] 实现 `UpdateAuthSessionWithSteamGuardCode`
- [x] 实现 `PollAuthSessionStatus`
- [x] 在 `api/authenticationservice` 内部实现最小 protobuf form 编码/解码 helper，暂不引入公共 protobuf 抽象
- [x] 增加 RSA 密码加密 helper，并明确其只辅助认证原子接口，不保存密码、不编排完整登录
- [x] 增加 `EResultError` 或等价可 `errors.As` 的错误类型，用于识别 `DuplicateRequest`、`AccountLogonDenied`、`InvalidLoginAuthCode`、`ExpiredLoginAuthCode`、`RateLimitExceeded`
- [x] 新增 `addons/websession`，基于 `client.API.AuthenticationService` 编排 credentials 登录流程
- [x] 在 `addons/websession` 中支持手机批准、手机令牌、邮箱验证码三类确认路径
- [x] 在 `addons/websession` 中实现 refresh token 到 Web Cookie 的转换，复用同一个 `CookieJar` 完成 `finalizelogin` 与 transfer 请求
- [x] 在 `addons/websession` 中实现 Store 与 Community session 校验，避免只依赖 `/account/licenses/`
- [ ] 新增 `addons/freeclaim`，只依赖调用方提供的 Web Cookie / CookieJar，不管理账号密码或 refresh token
- [ ] 在 `addons/freeclaim` 中实现限时免费候选搜索，先按 POC 经验解析 Store search HTML 片段
- [ ] 在 `addons/freeclaim` 中复用 `client.Web.Storefront.GetAppDetails`，从 `package_groups` raw payload 中解析免费 package/subid
- [ ] 在 `addons/freeclaim` 中实现单个 package 的免费 license 领取，优先从 app 页面表单解析隐藏字段
- [ ] 在 `addons/freeclaim` 中区分领取成功、已拥有、登录失效、疑似限流等结果
- [ ] 增加面向认证、websession、freeclaim 的单元测试和 fixture，覆盖请求构造、protobuf 字段、错误映射、HTML 解析和成功判定
- [ ] 增加 live smoke 示例，但默认不在 CI 中执行
- [ ] 更新英文与中文文档，说明主线 API、Web session addon、freeclaim addon 的边界、限制和安全注意事项
- [ ] 明确不支持读取浏览器 Cookie、读取 Steam 客户端本地 token、自动全部领取、无限重试和本地账号数据库

#### Acceptance Criteria

- `client.API.AuthenticationService` 提供 5 个认证原子接口，并有请求构造、protobuf 编码和响应解析测试
- `addons/websession` 能完成 credentials 登录编排、Steam Guard 分支处理、refresh token 换 Web Cookie、Store/Community session 校验
- `addons/freeclaim` 能搜索免费候选、解析 package/subid、领取单个免费 license，并能区分主要结果状态
- 所有新增能力复用现有 `WithHTTPClient`、`WithCookieJar`、`WithProxySelector`、`WithTrafficPolicy`、`WithRequestSessionKey` 和 traffic class 体系
- 不在主线 SDK 中保存密码、refresh token 或长期登录态
- `go test ./...` 通过，新增 live smoke 示例可由调用方显式配置后手动执行
- README、`docs/` 和中英文文档都说明能力边界、风险和推荐用法

#### Notes

- `steam-auth-web-session-poc-roadmap.md` 是本里程碑的实现参考，应在开发任务开始前逐项核对。
- 免费促销搜索和领取流程依赖 Store 页面行为，默认放在 `addons/freeclaim`，不进入 `client.API.*` 或 `client.Web.*` 稳定主线。
- `addons/websession` 不读取浏览器 Cookie，也不读取 Steam 客户端本地登录态。
- `addons/freeclaim` 首版只提供单个领取能力，不提供自动全部领取能力。

## 维护规则

- `roadmap.md` 只记录未来计划和当前阶段状态，不再承担历史 changelog 职责。
- 已发布版本的说明以 GitHub Releases 和 `docs/releases/` 为准。
- 没有明确排期时，保持本文档简洁，不预先写入猜测性的 `v1.x` 任务。
- 新增 roadmap 项目前，应先确认目标、范围、验收标准和是否会影响稳定 API。
