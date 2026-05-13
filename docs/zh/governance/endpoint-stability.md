# Endpoint 稳定性

本文档说明 `steam-go v1.0.0` 如何理解 API 稳定性。

## 稳定等级

### Stable

Stable 表示导出的 API 形状会纳入 `v1` 兼容性承诺。

包括：

- 当前 `client.API.*` 下的 typed service entrypoint
- 官方 Steam Web API 方法对应的导出 request / response 结构体
- 已记录的根包配置 API，例如 proxy、retry、traffic policy 和 request controls

### Preview

Preview 表示公开配置面已经存在并有文档，但项目尚未承诺完整产品化抓取 API。

当前包括：

- `TrafficClassPublicStorePage`
- 公开商店页 header profile
- Referer selector
- 短缓存
- block detection
- per-class transport hook

这些是有效基础能力，但不应理解为 `steam-go v1.0.0` 已经提供完整公开商店页客户端。

### Experimental

Experimental 表示未来可能通过 addon 或显式 experimental package 推进的方向，不属于 `v1.0.0` 稳定合同。

示例：

- Steam Store 页面抓取入口
- Steam Community 页面 helper
- CDN 衍生资源 helper
- 浏览器 fallback 实现

### Out of Scope

Out of Scope 表示当前不承诺。

包括：

- 在 `v1.0.0` 前继续扩充新的 API 覆盖
- 把非官方 Store / Community / CDN 抓取 API 纳入首个稳定面

## 官方 Web API 服务

当前通过 `client.API.*` 暴露的官方 `api.steampowered.com` 服务，是 `v1.0.0` 的主要稳定面。

除非存在 blocker 级问题，否则方法签名和导出的 typed response model 应保持稳定。

## Raw Payload 子树

部分官方响应包含高波动子树，因此故意建模为 `json.RawMessage`。

这不代表外层方法不稳定，只表示 raw 子树内部的精确 JSON 形状不作为 typed 稳定合同承诺。

## Addons

已记录的 addon import path 属于支持的仓库结构，但每个 addon 的行为仍以自身文档为准。

示例：

- `addons/openid`
- `addons/a2s`
- `addons/a2s/master`
- `addons/a2s/scanner`
