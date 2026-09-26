# Endpoint 稳定性

本文档说明 `steam-go` v1.x 的 API 稳定性。规范性要求以[兼容性契约](../../../contracts/compatibility.md)和[上游契约](../../../contracts/upstream.md)为准。

## 稳定等级

### Stable

Stable 表示导出的 API 形状会纳入 `v1` 兼容性承诺。

包括：

- 当前 `client.API.*` 下的 typed service entrypoint
- `client.API.*` 和 `client.Web.*` 已记录的方法及导出 request / response 结构体
- 已记录的根包配置 API，例如 proxy、retry、traffic policy 和 request controls
- 公开的本地 `steamid` 包和已记录的 addon import path

### 请求控制基础设施

以下是现有 Web surface 使用的、已记录的稳定配置 API：

- `TrafficClassPublicStorePage`
- 公开商店页 header profile
- Referer selector
- 短缓存
- block detection
- per-class transport hook

这些基础能力不承诺提供通用抓取器或浏览器客户端。

### Experimental

只有被文档明确标记为 experimental 的能力，才因此不纳入 v1 稳定合同。使用非官方上游本身并不意味着已记录的 Go API 是实验功能。

### Out of Scope

SDK 不承诺：

- 完整覆盖 Steam endpoint
- 通用 Store / Community HTML 抓取和自动浏览器 fallback
- 非官方 Web surface 的上游结构或可用性保持稳定

## Steam Web API 服务

`client.API.*` 包括有官方文档的 API 和 observed Steam 服务。必须准确标注上游状态；仅由 `api.steampowered.com` 托管并不构成官方稳定性保证。

已记录的方法签名和导出类型受 v1 兼容性契约约束。新增覆盖默认采用增量方式；安全和正确性例外必须明确记录。

## Raw Payload 子树

部分官方响应包含高波动子树，因此故意建模为 `json.RawMessage`。

这不代表外层方法不稳定，只表示 raw 子树内部的精确 JSON 形状不作为 typed 稳定合同承诺。

## 非标准 Web Surface

`client.Web.*` 提供 `api.steampowered.com` 之外的只读 Storefront、Community 与 Market JSON 接口。

这些 Go 方法签名属于 `v1` 兼容性承诺范围。

Valve 不保证这些上游 payload 和可用性遵循稳定的官方 Web API 合同。应通过防御性解码和增量模型处理漂移，同时保留现有签名、字段类型、JSON tag 和原始字段。

## Addons

已记录的 addon import path 属于支持的仓库结构，但每个 addon 的行为仍以自身文档为准。

示例：

- `addons/openid`
- `addons/a2s`
- `addons/a2s/master`
- `addons/a2s/scanner`
- `addons/assets`
- `addons/markup`
- `addons/vdf`
- `addons/websession`
- `addons/freeclaim`

`addons/assets` 的资源发现和下载已属于现有支持能力。其他 addon 的认证与修改操作边界见 [addon 参考](../addons/reference.md)和[安全说明](../addons/safety.md)。
