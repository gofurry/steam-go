# 兼容性策略

本文档定义 `steam-go` 从 `v1.0.0` 开始的兼容性承诺。

## 定位

`steam-go v1.0.0` 的定位是：

> 一个面向生产使用、专注于 Steam Web API 的稳定 Go SDK。

稳定承诺集中在官方 `api.steampowered.com` 客户端能力，以及根包已经公开的请求控制工具层。

## `v1.0.0` 稳定范围

除非另有说明，以下内容属于 `v1` 兼容性承诺范围：

- 根包 `steam`
- `NewClient(...)`
- 现有 `Option` 系统
- `Client` 与 `client.API.*` 分组访问模式
- 现有导出的 service method 签名
- 现有导出的 request / response 结构体
- 代理相关公开 API
- traffic policy 相关公开 API
- 错误类型与错误分类
- URL 脱敏 helper
- 已记录的 addon import path

## 稳定行为预期

稳定范围内应尽量保持：

- 导出名称
- 方法签名
- option 的行为语义
- 已记录的错误分类
- `client.API.*` 的分组模式

Bug 修复、校验收紧和内部实现调整是允许的，只要不破坏已记录的稳定面。

## 不属于 `v1.0.0` 承诺的范围

以下内容不属于 `v1.0.0` 兼容性承诺，除非未来文档另行说明：

- 未来新增的 Store / Community / CDN 抓取入口
- HTML 解析规则和页面结构假设
- 浏览器 fallback 的具体实现
- 未记录的 Web payload 结构
- 高波动 raw JSON 子树内部的细粒度结构
- 复杂外部代理池治理策略
- 未来 experimental package 或 addon

## Raw Payload 策略

`steam-go` 使用三类 payload 策略：

- 稳定官方 payload 优先使用强类型结构体
- 大型或快速变化的子树可以使用 `typed outer + json.RawMessage`
- 高波动 payload 在结构稳定前保持 raw

`json.RawMessage` 子树内部的精确结构不属于 `v1` 兼容性承诺，除非文档明确标记为稳定。

## Preview 与实验方向

以下能力可以存在于仓库中，但应理解为策略基础或 preview 方向，而不是完整产品化抓取 API：

- `TrafficClassPublicStorePage`
- 公开商店页 header profile
- Referer 策略
- 短缓存与 block detection 基础设施
- 面向未来 TLS 定制或浏览器执行栈的 per-class transport hook

这些 root package 配置 API 按文档保持稳定，但不代表 `steam-go v1.0.0` 已经提供完整公开商店页 SDK。
