# 凭据安全

Steam 凭据可能是长期有效、账号绑定或 session 绑定的敏感值。即使它们出现在 URL、redirect、cookie、示例或 debug 输出里，也应该按 secret 处理。

## 凭据类型

| 值 | 常见用途 | 安全说明 |
|---|---|---|
| Steam Web API key | 需要 `key` 的官方 Web API 方法 | 只应放在后端。泄露后应重置。 |
| Publisher / partner key | Steamworks 合作伙伴后端 API | 权限高于普通 key，不能分发给客户端。 |
| `access_token` | token-backed 用户或 Web 流程 | 通常较短期，但仍然绑定账号。 |
| `refresh_token` | 交换新的 web/session token | 高风险。不要通过命令行参数传递。 |
| `steamLoginSecure` / `sessionid` | Store / Community Web cookie | 绑定 session，应保存在受控 cookie jar 中。 |
| Proxy URL userinfo | 代理认证 | 记录日志前必须脱敏。 |

OpenID 不同：它只验证 Steam 身份并返回 SteamID64，不能替代 API key、access token 或应用 session。

## 为什么 URL 有风险

Steam API 经常把凭据放在 query 参数中：

```text
https://api.steampowered.com/.../?key=SECRET&access_token=TOKEN
```

这些 URL 可能通过下面渠道泄露：

- 应用日志
- HTTP trace
- metrics label
- redirect final URL
- panic 输出
- 截图
- bug report

记录 URL 前应使用 `steam.RedactSensitiveURL(...)`。

## Header 与 Cookie 脱敏

记录 request 或 response header 前，使用 `steam.RedactSensitiveHeaders(...)`：

```go
safeHeaders := steam.RedactSensitiveHeaders(req.Header)
```

`Authorization`、`Proxy-Authorization`、`Cookie`、`Set-Cookie`、`X-WebAPI-Key`、`X-API-Key` 等敏感 header 会被替换为 `[REDACTED]`。

`RedactSensitiveHeaders(...)` 返回 clone，不会修改原始 header map。

## Cookie Jar 生命周期

- cookie jar 应限定在一个用户 session 或一个显式 workflow 内。
- 不要跨无关用户复用已认证 Web cookie。
- 不要把 cookie jar 序列化到日志或公开 artifact。
- 用户退出登录或轮换凭据后，应清理或替换 jar。
- inventory 和 web-session 流程可能需要调用方提供 cookie，但核心 SDK 不替你托管长期登录态。

## 示例与 CLI 安全

项目示例会使用环境变量或隐藏终端输入读取敏感值。避免用命令行参数传递密码、refresh token 和 cookie，因为 shell history 和进程列表可能泄露它们。

优先使用：

- 本地 demo 使用环境变量
- 交互示例使用隐藏输入
- 自动化使用 CI/CD secrets
- 生产环境使用专用 secret manager

## 如果 Secret 泄露

1. 撤销或轮换泄露的凭据。
2. 尽可能从日志、issue、trace 和 artifact 中移除该值。
3. 检查 CI 变量和部署配置。
4. 审计该凭据最近的使用情况。
5. 为泄露路径补充或加强脱敏测试。
