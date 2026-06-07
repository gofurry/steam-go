# Storefront AppDetails 限流实验 - 2026-06-07

本文记录一次针对 Steam Storefront `appdetails` 的小规模经验性限流实验。

这不是 Steam 官方限流声明。它只代表一个网络环境、一个客户端配置和一个短时间窗口下的可复现实验观察。它适合用于选择保守生产默认值，不适合用于绕过上游限制。

## 实验范围

- Endpoint：`store.steampowered.com/api/appdetails`
- SDK 方法：`client.Web.Storefront.GetAppDetailsRaw`
- Traffic class：`TrafficClassPublicStorePage`
- 日期：2026-06-07
- AppIDs：`440`、`570`、`730`
- 地区：`CN`、`US`、`HK`
- 语言：`schinese`、`english`
- 每轮请求数：`3 appids * 3 regions * 2 languages * repeat 20 = 360`
- Workers：`10`
- 本地 Store interval：`0s`
- Burst：`0`
- Retry：`0`
- 本地 block cooldown：`0s`
- Proxy：关闭
- Timeout：`15s`

第二轮在等待约 5 分钟后开始。

## 实验结果

| Run ID | Total | OK | Failed | Blocked | HTTP 429 | HTTP 403 | Transport | Duration |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| `20260607-183228` | 360 | 228 | 132 | 105 | 31 | 74 | 27 | `1m35s` |
| `20260607-184100` | 360 | 228 | 132 | 132 | 34 | 98 | 0 | `20.55s` |

两轮都恰好成功了 228 次 Store appdetails 请求，随后剩余 132 个逻辑请求失败。第一轮有 27 个 transport timeout，第二轮所有失败都明确表现为 `429` / `403`。

## 时间线

| Run ID | 首个 429 | 首个 403 | 首个 block-detected 信号 |
| --- | --- | --- | --- |
| `20260607-183228` | 约 20s，seq `233` | 约 22s，seq `267` | 约 20s，seq `233` |
| `20260607-184100` | 约 11s，seq `221` | 约 12s，seq `263` | 约 11s，seq `221` |

第二轮更早在时间上触发，是因为没有第一轮开头那些 transport timeout，因此更快到达相近的请求数量边界。

## 结论解释

- Storefront appdetails 在该环境下表现出可复现的请求数量边界。
- `429` 大约从 220-230 次请求附近开始出现。
- 继续请求后，接下来几十个请求内会快速转为 `403` / block-detected。
- 等待约 5 分钟后，复测可以再次到达相近边界。
- 这个模式与 Store 约 `150-250` requests / 5 minutes / egress identity 的经验规则吻合。

## 生产建议

生产环境访问 Storefront appdetails 时：

- Store 页面类流量必须和 official API 流量分 bucket。
- Store appdetails 可以先按约 `1 request / 2 seconds`、`burst=1` 设计。
- 优先使用队列和缓存，不要靠高 worker 数堆吞吐。
- 把 `429`、`403` 和 `BlockDetected=true` 视为明确降速信号。
- Store 失败后应保留 cooldown，不要立即重试轰炸。
- 不要把短时压测峰值当作生产目标。

如果一次游戏详情采集需要 3 个地区的 appdetails 请求，那么保守的 `150 requests / 5 minutes` 预算大约等于：

```text
150 Store requests / 5 minutes / 3 regional requests ~= 50 apps / 5 minutes
```

如果 Store events、reviews 或其他 Storefront helper 共用同一个 traffic bucket，还需要继续预留余量。

## Official API 说明

短时 official API 突发能力和 Store 页面类流量是两个问题。即使短时请求没有返回 `429`，生产调用方仍然必须遵守 key 或应用的每日预算。如果某个 developer key 的业务预算是每天 `10,000` 次，那么当前在线人数这类任务应优先按每日预算调度，而不是按短时 QPS 上限调度。
