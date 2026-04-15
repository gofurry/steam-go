# steam-go Roadmap

## 代理规划（2026-04）

- 继续保留 `WithProxySelector(...)` 作为稳定扩展点
- 当前阶段暂不内建代理池轮转、健康检查、失败摘除或熔断策略
- 如调用方需要多代理轮换，先通过自定义 `ProxySelector` 在业务侧实现
- 等真实使用场景稳定后，再评估是否补内建 `RoundRobinProxySelector` 一类能力

## 当前状态

`steam-go` 的 Web API 主体重构已经基本完成。

- 已完成统一 `Client`
- 已完成 `Option`、认证注入、重试、限流、代理选择
- 已完成统一错误模型
- 已完成 `typed/raw` 双层返回
- 已完成核心 API 迁移：
  - `SteamUser`
  - `PlayerService`
  - `SteamNews`
  - `SteamUserStats`
  - `AccountCartService`
  - `BillingService`
  - `CommunityService`
  - `FamilyGroupsService`
  - `LoyaltyRewardsService`
- 已完成 `client.API.*` 二层入口重构
- 已完成 README、examples、CI、测试基线

按仓库最初的大路线看，当前整体完成度大约在 `80%` 左右。
如果只看 “轻量、聚焦、长期可维护的 Steam Web API SDK” 这个 v1 主目标，完成度已经在 `90%+`。

## 接下来的主线

接下来不建议再大改 Web API 核心结构，优先做三件事：

1. 继续补高价值 Web API 组，但保持现有 `client.API.*` 架构不变
2. 把 A2S 作为明确的 `addon` 能力接入，而不是继续塞回核心 Web API 主体
3. 为未来 OAuth / credential source 预留更清晰的扩展点

## 版本路线

### v2.x

目标：把 Web API 主体打磨到可长期发布的质量。

- 补更多高价值官方 API 组
- 补更完整的 examples
- 补更系统的集成测试和边界测试
- 完善 README 和发布说明

### v3

目标：引入扩展层，而不是继续膨胀核心 SDK。

- 接入 `addons/a2s`
- 明确 A2S 与 Web API 的边界
- 评估 OAuth 扩展设计

## A2S 规划

### 目标定位

A2S 不应该继续放进 Web API 主体里。

推荐结构：

```text
steam-go/
  addons/
    a2s/
      client.go
      options.go
      types.go
      errors.go
```

对外使用方式建议保持独立：

```go
a2sClient, err := a2s.NewClient("1.2.3.4:27015", ...)
info, err := a2sClient.QueryInfo(ctx)
players, err := a2sClient.QueryPlayers(ctx)
rules, err := a2sClient.QueryRules(ctx)
```

不要把它挂进 `steam.Client`，否则未来 Web API 和 UDP Query 两条线又会重新耦合。

### 建议的接入策略

优先建议：

1. 第一阶段基于现有成熟实现做一层轻封装
2. 第二阶段按真实使用情况决定是否 fork
3. 不建议一开始就完整自研

原因是 A2S 看起来协议不大，但兼容性细节并不少，真正麻烦的是不同游戏、不同服务端实现、challenge 刷新、分包、多包、超时和异常包处理。

### 对 `rumblefrog/go-a2s` 的评估

截至 `2026-04-15`，`rumblefrog/go-a2s` 的现状是：

- GitHub 仓库有 `72` stars、`19` forks、`1` 个 open issue
- README 明确说协议本身变化很少，所以更新不会频繁
- 最近一次 release 是 `v1.0.3`，发布日期是 `2026-03-24`
- 最近一次提交也是 `2026-03-24`
- 当前仍有一个 open issue：长期持有 client 时，服务端重新下发 challenge 可能导致 `unsupported protocol header`

来源：

- [GitHub 仓库首页](https://github.com/rumblefrog/go-a2s)
- [提交历史](https://github.com/rumblefrog/go-a2s/commits/master/)
- [v1.0.3 release](https://github.com/rumblefrog/go-a2s/releases/tag/v1.0.3)
- [issue #13](https://github.com/rumblefrog/go-a2s/issues/13)

我的判断：

- 它不是“无人维护的死库”
- 但也不是“高频、强担保维护”的上游
- 作为 `addons/a2s` 的第一阶段底层依赖是可以接受的
- 如果你准备长期承接 A2S 质量责任，最终大概率还是会走向 fork 或替换实现

### 是否直接用 upstream

如果目标是尽快把 A2S 能力接回 `steam-go`，我建议：

- 第一版直接依赖 `github.com/rumblefrog/go-a2s`
- 在 `addons/a2s` 里做你自己的接口、错误包装、超时和批量能力
- 不要让业务代码直接依赖上游库类型

这样做的好处：

- 集成成本最低
- 可以快速验证 API 设计
- 后面切换到底层 fork 或自研时，对外 API 不需要推倒重来

### 是否 fork 一个版本自己维护

如果你准备把 A2S 当成长期卖点，而不是“顺带支持”，那我更推荐中期 fork。

比较适合 fork 的信号有：

- 你需要修 upstream 的 challenge 刷新问题
- 你需要为特定游戏做兼容修复
- 你需要更强的超时、重试、并发批量查询能力
- 你希望发布节奏、CI、Go 版本策略完全由自己掌控

建议不是“现在立刻 fork”，而是：

- `v3.0`: 先用 upstream 做 adapter
- `v3.1`: 如果发现需要 patch，就 fork 到你自己的命名空间
- `v3.2+`: 再决定继续维护 fork，还是逐步替换成自研实现

### 是否重新实现

重新实现不是做不到，但不建议作为第一步。

难度评估：

- 只做最基础的 `A2S_INFO / A2S_PLAYER / A2S_RULES`，并在少量常见服务器上可用：`中等`
- 做到可对外稳定发布、覆盖 challenge、分包、多包、异常包、长连接行为、各类游戏兼容：`中高`

粗略估算：

- 最小可用版：约 `3-5` 天
- 可在仓库里作为 addon 发布的稳定版：约 `1-2` 周
- 做到比较放心地长期维护：约 `3-6` 周，取决于你是否愿意搭真实服务器测试矩阵

真正麻烦的部分不是“协议字段怎么解析”，而是：

- challenge 重新下发
- 多包响应拼装
- 超规格服务端实现
- 老游戏和非标准服务端
- 真实网络超时和 UDP 丢包行为

### 当前推荐决策

推荐决策是：

1. `v3.0` 先做 `addons/a2s`
2. 底层先用 `rumblefrog/go-a2s`
3. 在 `addons/a2s` 外层建立自己的稳定 API
4. 增加真实服务器回归测试样例
5. 观察一到两个迭代，再决定是否 fork

不推荐当前采用的方案：

- 不推荐直接把第三方 A2S 类型暴露给 `steam-go` 用户
- 不推荐现在就完整自研
- 不推荐使用 `WoozyMasta/a2s` 作为直接依赖

最后这一点很重要：`WoozyMasta/a2s` 当前是 `AGPL-3.0`，和你现在仓库的 `MIT` 方向不匹配，不适合作为直接集成依赖。它可以参考，但不适合直接拿来接进主仓库。

来源：

- [WoozyMasta/a2s](https://github.com/WoozyMasta/a2s)

## A2S 的执行顺序

### Phase A

建立 `addons/a2s` 目录和最小 API：

- `NewClient`
- `QueryInfo`
- `QueryPlayers`
- `QueryRules`
- `Close`

### Phase B

补适配层能力：

- timeout option
- max packet size option
- 统一错误包装
- 基础 examples

### Phase C

补面向真实使用的能力：

- 批量查询
- context 支持
- 真实服务器手动测试
- 回归测试基线

### Phase D

决定底层策略：

- 保持 upstream
- fork upstream
- 替换为自研实现

## 一句话结论

`steam-go` 现在最合适的 A2S 路线不是“立刻重写”，而是：

> 先把 A2S 作为独立 addon 接回来，底层先吃 `rumblefrog/go-a2s`，对外暴露自己的稳定接口，等真实需求证明有必要，再 fork 或替换实现。
