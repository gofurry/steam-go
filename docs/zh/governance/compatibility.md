# 兼容性策略

本文档概述 `steam-go` v1.x 的兼容性策略。规范性要求以[公开兼容性契约](../../../contracts/compatibility.md)为准；上游假设以 [Steam 上游契约](../../../contracts/upstream.md)为准。

## 定位

`steam-go` 的定位是：

> 一个提供 Steam API、限定范围的只读 Web 能力和可选 addon 的稳定 Go SDK。

稳定承诺覆盖已记录的公开 Go API，不代表每个 Steam 托管的 endpoint 都有官方文档，也不代表其上游 payload 保持稳定。

## v1.x 稳定范围

除非另有说明，以下内容属于 `v1` 兼容性承诺范围：

- 根包 `steam`
- 公开的本地 `steamid` 包
- `NewClient(...)`
- 现有 `Option` 系统
- `Client` 与 `client.API.*` 分组访问模式
- `Client` 与 `client.Web.*` 分组访问模式，以及其导出的 request / response 结构体
- 现有导出的 service method 签名
- 现有导出的 request / response 结构体
- 代理相关公开 API
- traffic policy 相关公开 API
- 错误类型与错误分类
- URL 脱敏 helper
- 已记录的 addon import path

## 稳定行为预期

稳定范围内应保持：

- 导出名称
- 方法签名
- option 的行为语义
- 已记录的错误分类
- `client.API.*` 的分组模式
- `client.Web.*` 的分组模式

Bug 修复和内部调整应保持合法调用方的兼容性。安全或正确性修复可以收紧不安全或客观无效的行为，但必须按兼容性契约明确记录例外。

## 不属于 v1 承诺的范围

以下内容不属于 v1 兼容性承诺，除非文档明确另行说明：

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

## 非标准 Web surface

以下能力属于稳定的 Go API，但其上游仍然是高波动的 Web surface：

- `client.Web.Storefront`
- `client.Web.Community`
- `client.Web.Market`

上游 payload 漂移应通过防御性解码和增量模型处理，同时保留现有方法签名、导出字段类型、JSON tag 和原始字段。新增高波动子树可按需使用 `json.RawMessage`。

## 请求控制基础设施

以下已记录的配置 API 是当前 Web surface 使用的稳定基础设施：

- `TrafficClassPublicStorePage`
- `TrafficClassCommunityWeb`
- `TrafficClassMarketWeb`
- 公开商店页 header profile
- Referer 策略
- 短缓存与 block detection 基础设施
- 面向未来 TLS 定制或浏览器执行栈的 per-class transport hook

这些 root package 配置 API 按文档保持稳定，并作为现有 `client.Web.*` 的策略基础设施，但不代表未来所有 Steam Web 流程都会被产品化接入 SDK。

## 兼容演进

v1.x 可以继续以增量方式扩展：

- 新增官方 Steam Web API 方法
- 为稳定官方 payload 补充 typed 覆盖
- 新增具有明确范围和上游边界的可选 addon
