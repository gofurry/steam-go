# steam-go 架构设计与重构方案项目计划书

## 1. 文档信息

- 项目名称：`steam-go`
- 文档类型：架构设计与重构方案项目计划书
- 文档版本：v1.0
- 当前阶段：重构前设计评审
- 目标仓库定位：轻量、模块化、长期维护的 Go 语言 Steam Web API SDK

---

## 2. 项目背景

现有仓库 `gf-steam-sdk` 已经具备一定的 SDK 雏形与可复用价值，尤其体现在以下几个方面：

1. 已有统一入口思路，说明项目不是简单的方法堆砌，而是有 SDK 产品化意识。
2. 已有链式配置与请求复用设计，说明项目已经在考虑开发者体验与重复逻辑收敛。
3. 已有泛型或统一请求抽象方向，说明项目已经开始形成“请求构建—发送—解析—返回”的骨架。
4. 已有多层返回理念，说明项目已经意识到不同调用者对 typed result / raw response 的不同需求。

但原仓库同时存在明显的架构偏移问题：

- 核心定位本应是 Steam Web API SDK，但仓库逐步混入了 crawler、反爬、HTML 处理、重日志体系等非核心能力。
- 目录与依赖逐渐承载了过多不属于 SDK 核心职责的功能，导致仓库边界模糊。
- 一些能力虽然“能做”，但并不适合继续放在主仓库中长期维护。
- 由于 Steam API 本身存在官方命名体系，如果继续按主观业务域切分，很容易在未来接入更多 API 时出现归类混乱。

因此，本次重构并不是简单“重写一遍”，而是一次**目标收敛、边界重建、资产提纯**的工程化升级。

---

## 3. 重构目标

本次重构的核心目标不是追求功能更多，而是让项目变得更干净、更稳定、更适合长期维护。

### 3.1 主目标

1. 将仓库重新定义为：**Steam Web API SDK**。
2. 保留原项目最有价值的设计资产：
   - 统一入口
   - Option 配置体验
   - 请求抽象复用
   - 原始响应能力
3. 移除不适合留在核心仓库的重型依赖与偏离能力。
4. 建立可长期扩展的目录结构与服务划分方式。
5. 为未来的 OAuth、CLI、A2S addons、更多 API 接入留出空间，但不在首版强行实现。

### 3.2 重构后的目标形态

重构后的 `steam-go` 应满足以下特征：

- 用户一眼能看出它是 Go 语言的 Steam Web API SDK。
- 代码结构贴近 Steam 官方 API 命名体系，降低理解与维护成本。
- SDK 核心依赖尽可能轻量，不强绑日志、爬虫、HTML 处理方案。
- Transport、认证、代理等能力可插拔。
- 默认使用 typed response，raw response 作为补充能力，而不是方法爆炸式扩散。
- 仓库主线清晰，便于后续逐步补全官方 API。

---

## 4. 非目标

为了避免再次把核心边界打散，本次重构明确以下非目标：

1. **不内置 crawler 框架**
   - 不再把 `goquery`、`colly` 这类 HTML 抓取能力放入核心仓库。

2. **不把反爬作为主能力宣传**
   - 核心仓库不是采集框架，不负责复杂反爬策略。

3. **不集成重型日志体系**
   - 不再把 `zap`、`lumberjack` 作为核心依赖。
   - SDK 不替使用者决定落盘策略与日志框架。

4. **不在首版实现 OAuth**
   - OAuth 作为未来演进方向，仅在架构层面预留扩展点。

5. **不在首版把 A2S 做成核心模块**
   - A2S 作为 `addons` 提供二次封装，而不是与 Web API 并列为核心层。

6. **不做主观业务域大分层**
   - 不先做 `users / apps / store / community` 这类带主观判断的顶层架构。

---

## 5. 原仓库资产盘点

本节用于明确：哪些东西值得继承，哪些应该拆分，哪些应该彻底删除。

### 5.1 必须保留的核心资产

#### 5.1.1 统一入口思想
原仓库已经具备统一入口思路，这说明项目拥有较强的 SDK 设计意识。新仓库需要保留这一点，但将其职责收敛到 Steam Web API。

#### 5.1.2 链式配置 / Option 体验
原仓库重视配置方式与调用体验，这一点需要保留，但实现方式应更贴近 Go 生态，使用 `Option` 模式而不是继续让配置层级膨胀。

#### 5.1.3 请求泛型复用 / 统一请求抽象
这是原仓库最有价值的技术资产之一。重构后的核心代码应围绕统一请求抽象展开，而不是围绕 endpoint 方法堆砌展开。

#### 5.1.4 多层数据返回理念
原仓库意识到了 typed result 与 raw response 的差异化需求，这是成熟方向。但需要改造为更收敛的 API 设计，避免变成每个方法都派生多套变体。

### 5.2 应拆分出去的能力

#### 5.2.1 A2S 能力
A2S 与 Steam Web API 不属于同一层级。它有价值，也值得做，但不应与 Web API 主仓库混为一体。

处理方案：
- 保留为 `addons/a2s`
- 做轻封装与接口适配
- 后续如有必要，可独立成单仓库或更成熟的扩展包

#### 5.2.2 代理池轮转
代理池轮转本身合理，但不应把复杂策略直接塞进核心。核心只暴露代理选择接口，具体轮转策略由扩展实现。

### 5.3 应彻底剔除的内容

#### 5.3.1 HTML 抓取能力
- `goquery`
- `colly`
- `goldmark`

这些依赖与 Steam Web API SDK 的核心定位不符，应从主仓库中移除。

#### 5.3.2 过重日志依赖
- `zap`
- `lumberjack`

SDK 只应支持注入轻量 logger 接口，不应强行绑定具体日志方案。

#### 5.3.3 `pkg/util` 之类兜底层
这类目录长期来看只会变成杂项堆放区。应改为职责清晰的内部模块，不再保留“大口袋 util”。

---

## 6. 核心设计原则

### 6.1 单一职责
主仓库只负责 Steam Web API SDK，不承载与此无关的框架性能力。

### 6.2 贴近上游 API 命名
服务分组优先按 Steam 官方 API 名称族组织，而不是开发者主观业务域判断。

### 6.3 轻量依赖
优先使用标准库与必要轻量工具，不把应用层方案强绑到 SDK 中。

### 6.4 扩展优先于内建复杂度
对于代理、认证、A2S 等能力，优先设计接口与扩展点，而不是首版内建重逻辑。

### 6.5 稳定优先于“好看但主观”的抽象
首版要优先保证结构稳定、容易补 API、容易维护，而不是追求抽象层次看起来更高级。

### 6.6 面向长期维护
所有核心目录与命名都应考虑未来新增 API 时的演进路径，避免频繁迁移与重构。

---

## 7. 总体架构设计

### 7.1 架构总览

```text
Client
 ├── API Services
 │     ├── SteamUser
 │     ├── PlayerService
 │     ├── SteamNews
 │     ├── SteamUserStats
 │     └── ...
 │
 ├── Core Components
 │     ├── Request Builder
 │     ├── Endpoint Registry
 │     ├── Response Decoder
 │     ├── Auth Injector
 │     └── Error Mapper
 │
 ├── Transport Layer
 │     ├── HTTP Client
 │     ├── Timeout
 │     ├── Retry
 │     ├── Rate Limit
 │     └── Proxy Selector
 │
 └── Addons
       └── A2S
```

### 7.2 架构分层说明

#### Client 层
负责统一入口与公共依赖持有。

#### API Service 层
负责对外暴露各个 Steam API 族的方法。

#### Core 层
负责统一请求构建、公共参数注入、响应解析、错误转换等通用逻辑。

#### Transport 层
负责实际请求发送与网络行为控制。

#### Addons 层
负责承载不属于核心 Web API 的扩展能力，例如 A2S。

---

## 8. 包与目录结构设计

建议目录结构如下：

```text
steam-go/
├── client.go
├── options.go
├── errors.go
├── logger.go
├── README.md
├── LICENSE
├── go.mod
│
├── api/
│   ├── steamuser/
│   │   ├── service.go
│   │   ├── methods.go
│   │   └── types.go
│   │
│   ├── playerservice/
│   │   ├── service.go
│   │   ├── methods.go
│   │   └── types.go
│   │
│   ├── steamnews/
│   │   ├── service.go
│   │   ├── methods.go
│   │   └── types.go
│   │
│   └── steamuserstats/
│       ├── service.go
│       ├── methods.go
│       └── types.go
│
├── internal/
│   ├── auth/
│   ├── endpoint/
│   ├── request/
│   ├── response/
│   ├── errors/
│   └── transport/
│
├── addons/
│   └── a2s/
│       ├── client.go
│       ├── options.go
│       └── types.go
│
└── examples/
    ├── steamuser/
    ├── playerservice/
    └── steamnews/
```

### 8.1 目录设计理由

#### `api/`
用于按 Steam 官方接口族组织对外服务与模型，便于查阅官方文档时做映射。

#### `internal/`
用于放置不会直接暴露给使用者的 SDK 核心实现。

#### `addons/`
用于放置不属于 Web API 核心但又与 Steam 生态相关的扩展能力。

#### `examples/`
用于放置最小可运行示例，不把示例逻辑混入 SDK 主流程。

---

## 9. API 分组策略

### 9.1 为什么不按业务域分组

不建议在首版直接采用：

- `users`
- `apps`
- `store`
- `community`

原因在于：

1. Steam 官方 API 自带固定命名体系。
2. 某些接口归属并不总是明显，强行按业务域拆分容易引入主观误判。
3. 未来新增 API 时，继续按业务域切分会越来越难保持一致。

### 9.2 推荐的分组方式

优先按官方接口名或接口族组织，例如：

- `steamuser`
- `playerservice`
- `steamnews`
- `steamuserstats`
- `steameconomy`
- `steamapps`

### 9.3 这种分组方式的好处

1. 与官方文档保持强映射，便于查阅与维护。
2. 新 API 接入成本低，不必先做复杂归类。
3. 结构稳定，适合长期扩展。

---

## 10. Client 设计

### 10.1 Client 职责

`Client` 是整个 SDK 的统一入口，负责：

- 持有基础配置
- 持有公共 transport
- 持有 logger
- 初始化各个 API Service

### 10.2 Client 草案

```go
type Client struct {
    apiKey     string
    httpClient *http.Client
    logger     Logger

    SteamUser      *steamuser.Service
    PlayerService  *playerservice.Service
    SteamNews      *steamnews.Service
    SteamUserStats *steamuserstats.Service
}
```

### 10.3 初始化方式

```go
client := steam.NewClient(
    steam.WithAPIKey("xxx"),
    steam.WithTimeout(10*time.Second),
    steam.WithRetry(2),
)
```

### 10.4 Client 初始化原则

1. 合理默认值优先。
2. 所有可调能力都通过 Option 注入。
3. 不要求用户先显式构建复杂配置对象。

---

## 11. Option 配置设计

### 11.1 设计目标

Option 模式用于替代重配置对象和链式膨胀配置，提升 Go 风格一致性。

### 11.2 Option 草案

```go
type Option func(*Client)

func WithAPIKey(key string) Option
func WithBaseURL(url string) Option
func WithHTTPClient(hc *http.Client) Option
func WithTimeout(d time.Duration) Option
func WithRetry(n int) Option
func WithRateLimit(limit int) Option
func WithProxySelector(selector ProxySelector) Option
func WithLogger(logger Logger) Option
```

### 11.3 配置设计要点

#### 11.3.1 必要配置
- API Key

#### 11.3.2 可选配置
- Timeout
- Retry
- Rate Limit
- Proxy Selector
- Logger
- BaseURL（用于测试或自定义环境）

#### 11.3.3 不建议的配置方式
- 不让使用者维护一个极重的嵌套配置对象
- 不把未来可能不用的高级能力全部暴露成首版大而全配置项

---

## 12. 请求抽象设计

### 12.1 目标

统一所有 API 方法的实现流程，避免每个方法重复处理：

- URL 构建
- Query 参数注入
- API Key 注入
- 请求发送
- 响应解码
- 错误转换

### 12.2 标准请求流程

```text
Service Method
  -> Build Endpoint
  -> Build Query Params
  -> Create Request
  -> Inject Auth
  -> Transport.Do
  -> Decode Response
  -> Map Error
  -> Return Typed Result
```

### 12.3 请求抽象设计建议

可以在 `internal/request` 中定义统一请求描述结构，例如：

```go
type RequestSpec struct {
    Method      string
    Path        string
    Query       url.Values
    Headers     http.Header
}
```

然后由统一执行层负责：

- 拼接基础 URL
- 注入鉴权参数
- 发送请求
- 读取响应
- 解析目标类型

### 12.4 请求抽象的收益

1. 减少 endpoint 方法样板代码。
2. 便于后续批量补全更多 API。
3. 便于统一测试和错误处理。

---

## 13. Endpoint Registry 设计

为了避免路径字符串散落在各处，建议在 `internal/endpoint` 中集中管理 endpoint 元数据。

### 13.1 示例形式

```go
const (
    SteamUserGetPlayerSummaries = "/ISteamUser/GetPlayerSummaries/v2/"
    SteamUserResolveVanityURL   = "/ISteamUser/ResolveVanityURL/v1/"
)
```

### 13.2 设计收益

1. 减少 magic string。
2. 便于版本切换或路径修复。
3. 便于统一审查接口覆盖率。

---

## 14. 多层数据返回设计

### 14.1 原仓库问题

原仓库中的多层返回理念是正确的，但如果每个 API 都衍生出：

- Typed
- Brief
- RawModel
- RawBytes

会导致方法数量膨胀，维护成本急剧上升。

### 14.2 重构方案

首版建议保留两层：

1. **Typed Response**：默认返回形式
2. **Raw Response**：补充能力

### 14.3 调用示例

```go
resp, err := client.SteamUser.GetPlayerSummaries(ctx, ids)
raw, err := client.SteamUser.GetPlayerSummariesRaw(ctx, ids)
```

### 14.4 为什么不保留 Brief 为核心返回层

`Brief` 更像是展示层或便捷层，不应成为 SDK 核心服务层的一等职责。

建议将 `Brief` 能力下沉为：

- helper
- formatter
- example
- 或调用方自行处理

---

## 15. 类型模型设计

### 15.1 设计原则

1. 模型应贴近官方返回结构。
2. 尽量避免过早做“业务化 DTO 重写”。
3. 每个 API 组维护自己的 `types.go`，减少跨模块耦合。

### 15.2 类型组织建议

- 每个 API 组下维护与该组强相关的 request / response struct。
- 公共字段可抽出到少量 shared type，但避免过度抽象。

### 15.3 为什么不做一个全局大 models 包

全局 `models` 包虽然看起来统一，但长期很容易膨胀成跨领域混合模型仓库，降低维护清晰度。

---

## 16. 错误处理设计

### 16.1 设计目标

统一对 HTTP 错误、解析错误、Steam API 响应异常进行分类封装。

### 16.2 错误类型草案

```go
type APIError struct {
    StatusCode int
    Code       string
    Message    string
    RawBody    []byte
}
```

### 16.3 错误分类

建议至少区分：

- 请求构建错误
- 网络发送错误
- HTTP 状态异常
- 响应解码错误
- Steam API 业务错误

### 16.4 设计收益

1. 让用户能更稳定地判断错误来源。
2. 降低上层调用方对字符串匹配的依赖。
3. 便于测试和日志记录。

---

## 17. 日志设计

### 17.1 目标

让 SDK 支持日志，但不绑死用户的技术选型。

### 17.2 方案

只提供一个极轻量 logger 接口：

```go
type Logger interface {
    Debug(msg string, args ...any)
    Error(msg string, args ...any)
}
```

### 17.3 默认策略

- 默认可以不输出日志
- 或默认使用最轻量的 no-op logger

### 17.4 不做的事情

- 不默认引入 `zap`
- 不默认引入 `lumberjack`
- 不处理文件滚动
- 不替开发者决定日志落地方式

---

## 18. Transport 设计

### 18.1 职责

Transport 负责：

- HTTP 请求发送
- timeout
- retry
- rate limit
- proxy 注入

### 18.2 Proxy 设计

核心只保留接口：

```go
type ProxySelector interface {
    Next(req *http.Request) (*url.URL, error)
}
```

### 18.3 为什么只保留接口

代理池轮转的策略非常多样：

- round robin
- random
- health-based
- sticky

这些策略不应在首版核心中强行内建。核心只定义能力边界，复杂实现由扩展承担。

### 18.4 Retry 设计建议

首版可支持简单重试：

- 指定最大重试次数
- 对临时网络错误或 5xx 做有限重试
- 避免在核心中过度复杂化重试策略

### 18.5 Rate Limit 设计建议

首版可支持基础限流能力，但不追求复杂令牌桶实现暴露给用户。可以通过 option 注入轻量策略或内部实现。

---

## 19. 认证设计

### 19.1 首版认证目标

首版主要支持 API Key 注入。

### 19.2 实现建议

在 `internal/auth` 维护统一注入逻辑，例如：

- query 参数注入
- header 注入（若未来需要）

### 19.3 OAuth 的处理方式

首版不实现完整 OAuth，但架构上应避免将 API Key 写死到所有方法逻辑中。

建议预留：

- auth provider
- token source
- credential injector

这样后续接入 OAuth 时不需要推倒重来。

---

## 20. A2S Addons 设计

### 20.1 设计定位

A2S 是 Steam 生态里有价值的扩展能力，但不应直接和 Web API 主体混编。

### 20.2 目录建议

```text
addons/
  └── a2s/
      ├── client.go
      ├── options.go
      └── types.go
```

### 20.3 职责边界

- 对社区推荐的 A2S 请求库做二次封装
- 提供更贴合 `steam-go` 风格的配置与调用体验
- 与主 `Client` 分离，避免边界混乱

### 20.4 后续演进

如果未来你参与上游 A2S 库维护，`addons/a2s` 可以保持轻封装角色，减少重复造轮子。

---

## 21. 示例与文档策略

### 21.1 示例目录

示例建议按 API 组组织：

- `examples/steamuser`
- `examples/playerservice`
- `examples/steamnews`

### 21.2 示例设计原则

- 示例必须最小、清晰、可直接运行
- 不把复杂 demo 混入 SDK 核心
- 示例优先展示最常见调用方式

### 21.3 README 策略

README 首版只强调：

- 这是 Steam Web API SDK
- 支持 Option 配置
- 支持 typed response + raw response
- 支持可插拔 transport 扩展
- A2S 为 addons

不要再把 crawler、反爬、HTML 处理写成核心卖点。

---

## 22. 测试策略

### 22.1 测试目标

重构后的 SDK 应优先保证：

- 请求构建正确
- endpoint 映射正确
- 响应解析正确
- 错误分类正确
- option 生效正确

### 22.2 测试层次

#### 单元测试
- request builder
- option 应用
- response decoder
- error mapper

#### 集成测试
- 针对少量核心 API 做真实或 mock 请求验证

#### 回归测试
- 对重构前后保留功能做行为对照

### 22.3 Mock 建议

首版可基于 `httptest` 进行 mock，不要引入过重测试框架。

---

## 23. 迁移与重构执行计划

本节为项目计划书的执行部分，明确重构节奏，而不是停留在架构概念层。

### Phase 0：设计冻结

目标：把边界、目录、接口命名、保留内容全部定下来。

输出：
- 本架构文档
- 初版目录草案
- API 分组原则

### Phase 1：仓库初始化

目标：新建 `steam-go` 仓库骨架。

任务：
1. 初始化 `go.mod`
2. 建立基础目录结构
3. 落地 `Client`
4. 落地 `Option`
5. 落地 `Logger`
6. 落地内部 request / response / transport 骨架

输出：
- 能编译通过的空骨架
- 最小 README

### Phase 2：核心请求链路打通

目标：先打通“一个 API 从请求到返回”的最小闭环。

任务：
1. 设计 endpoint registry
2. 设计 request spec
3. 设计统一执行器
4. 设计 typed response 解码
5. 设计 raw response 返回能力
6. 设计基础错误处理

输出：
- 一个可工作的 API 组
- 完整请求闭环

### Phase 3：首批 API 迁移

目标：挑选最常用的 API 组完成首轮迁移。

建议优先级：
1. `steamuser`
2. `playerservice`
3. `steamnews`
4. `steamuserstats`

任务：
- 迁移方法定义
- 迁移模型
- 为每个 API 组补示例与测试

输出：
- 首版可用 SDK 雏形

### Phase 4：清理遗留设计

目标：确保新仓库没有旧包袱残留。

任务：
1. 确认不再引入 `colly`、`goquery`、`goldmark`
2. 确认不再引入 `zap`、`lumberjack`
3. 确认没有 `pkg/util` 类兜底目录
4. 确认 README 不再出现 crawler 主叙事

输出：
- 真正“干净”的 v1 核心结构

### Phase 5：A2S Addons 接入

目标：在核心稳定后再提供 A2S 扩展封装。

任务：
1. 设计 `addons/a2s` 结构
2. 保持与主 Client 风格一致
3. 明确其不是 Web API 核心组成部分

输出：
- 第一个 addons 模块

### Phase 6：发布前收尾

目标：对外发布前完成开源工程化整理。

任务：
1. 整理 README
2. 增加示例
3. 增加基础测试
4. 增加 License
5. 增加 GitHub Topics
6. 首个 Release 版本规划

输出：
- 可对外公开的 v0.x / v1.0 版本

---

## 24. 里程碑建议

### 里程碑 M1：骨架完成

完成标准：
- 新仓库可编译
- Client/Option/Transport 骨架完成
- 无旧依赖残留

### 里程碑 M2：最小 API 可用

完成标准：
- 至少一个 API 组可工作
- 支持 typed response 与 raw response
- 基础错误处理完成

### 里程碑 M3：首版对外可用

完成标准：
- 至少 3~4 组常用 API 已迁移
- 示例完整
- 测试覆盖核心逻辑
- README 与开源形象完成

### 里程碑 M4：扩展准备完成

完成标准：
- A2S addons 原型完成
- OAuth 扩展点预留完成
- 目录结构已验证可继续增长

---

## 25. 风险与注意事项

### 25.1 风险一：旧思路回流

在重构中最容易发生的问题是“因为旧代码现成，所以又把原来的能力顺手搬回来”。

处理原则：
- 每迁移一个模块都先问：它是不是 Steam Web API SDK 的核心职责？

### 25.2 风险二：过早抽象

为了“优雅”而把层级做得过多，会拖慢首版落地。

处理原则：
- 只抽象那些已经明显重复、且未来会继续重复的部分。

### 25.3 风险三：命名再次主观化

如果半途改回 `users/apps/store/community` 这类分组，未来扩展时仍会遇到归类冲突。

处理原则：
- 以官方 API 名称族为准。

### 25.4 风险四：首版功能欲望过强

SDK 重构最容易失控的地方不是代码难，而是想一次把 OAuth、CLI、代理池、A2S 全做完。

处理原则：
- 首版只解决最核心的 Web API SDK 重构。

---

## 26. 首版交付范围建议

建议首版交付范围控制在以下内容：

### 必须交付
- `Client`
- `Option`
- `Logger`
- `Request Core`
- `Response Decoder`
- `Error Model`
- `steamuser`
- `playerservice`
- `steamnews`
- 基础示例
- 基础测试

### 可后置交付
- `steamuserstats`
- 更多 API 组
- `addons/a2s`
- 更完整的 rate limit 策略
- OAuth
- CLI

---

## 27. 最终结论

本次重构不是把旧项目“整理一下”，而是一次明确方向的再出发。

### 应保留的，是旧仓库里最值钱的部分
- 统一入口
- Option 配置
- 请求抽象
- raw response 能力

### 应舍弃的，是让仓库失焦的部分
- crawler
- HTML 抓取依赖
- 重日志体系
- util 大杂烩
- 主观业务分层

### 新仓库真正要解决的问题只有一个

> **做一个干净、稳定、长期可维护的 Go 语言 Steam Web API SDK。**

只要这个主线不偏，后续的 A2S、OAuth、CLI、更多 API 接入都可以自然演进；
但如果首版主线继续发散，仓库很容易再次回到“大而杂”的状态。

因此，`steam-go` 的第一优先级不是“做更多”，而是：

> **先把最核心的结构做对。**

