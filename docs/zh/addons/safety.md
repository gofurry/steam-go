# Addon 安全边界

Addon 用于把可选工作流放在核心 SDK 之外。它们应保持显式、收敛、易审计。

## `addons/openid`

- 验证基于浏览器的 Steam OpenID 回调。
- 不创建应用自己的用户系统。
- 不自动拉取玩家资料。
- 不管理长期 session 或 cookie。
- 调用方必须把 `state` 绑定到自己的 CSRF / session 模型。

## `addons/websession`

- 驱动一条手动 Steam Web 登录流程。
- 会处理 Steam 返回的 cookie jar，调用方必须像保护凭据一样保护这些 cookie。
- 不持久化密码、refresh token、access token 或 cookie。
- 不读取浏览器 cookie 或本地 Steam 客户端状态。
- 适合已有凭据处理策略的应用显式使用。

## `addons/freeclaim`

- 提供只读限免 / package 发现，以及一次显式 free license claim。
- 不自动领取所有免费 package。
- 不管理账号池或批量账号自动化。
- 不无限重试，也不绕过上游限制。
- Claim 模式应要求调用方显式决策，并妥善保护 refresh token。

## `addons/markup`

- 默认清洗生成的或调用方传入的 HTML。
- `WithSanitize(false)` 只适合可信输入和可信渲染目标。
- 不请求远程内容。
- 不替业务决定哪些清洗后的标签应该展示。

## `addons/vdf`

- 只解析调用方提供的文本 VDF / KeyValues 输入。
- 不解析 binary VDF 或 `shortcuts.vdf`。
- 不自动扫描本地 Steam 安装。
- 除非调用方显式传入路径，否则不读取用户目录。
- 不提取账号、token、cookie 或 session。

## `addons/assets`

- 构造并可选请求公开 Steam 资源 URL。
- 静态 URL 构造不需要网络；verify/read/download helper 会发起网络请求。
- 直接 URL helper 如果接收用户输入或其他不可信 URL，必须设置 `URLValidator`。
- Steam 可能新增、移动或删除资源，因此资源存在性不作永久保证。

## `addons/a2s`

- 查询游戏服务器和 master-server discovery endpoint。
- 不代理、隐藏或轮换调用方身份。
- 批量扫描应由调用方显式设置限制和 timeout。
- 桥接层应跟随上游 `a2s-go` 行为，不在本仓库重新实现协议逻辑。
