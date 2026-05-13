# OpenID 说明

`addons/openid` 用于浏览器中的 Steam 登录验证。

它确认用户的 Steam 身份，但不是通用的 Web API 凭证系统。

## 它做什么

OpenID addon 帮助完成：

- 构造 Steam OpenID 登录 URL
- 验证 Steam callback
- 调用 `check_authentication`
- 获取 `SteamID64`
- 保留并校验 `state`

## 它不做什么

它不负责：

- 替代 Steam Web API key
- 替代 access token
- 自动获取用户资料
- 管理你的应用 session
- 把用户写入数据库
- 处理前端 UI

## 典型流程

```text
1. 生成随机 state
2. 把 state 存到安全 cookie 或服务端 session
3. 重定向用户到 Steam OpenID 登录 URL
4. Steam 回调你的 callback URL
5. 使用 addons/openid 验证 callback
6. 比对返回的 state 和已保存 state
7. 创建你自己的应用 session
```

## 示例

```bash
go run ./examples/openid
```

带代理：

```bash
go run ./examples/openid --proxy http://127.0.0.1:7897
```

## 安全说明

- 一定要校验 `state`。
- 不要要求用户输入 Steam 用户名或密码。
- OpenID 只证明身份，授权决策仍属于你的应用。
- 生产环境使用 HTTPS。
- 应用 session 应和 Steam OpenID 验证流程分离。

