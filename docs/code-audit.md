# Go 代码审计报告

## 状态

本报告对应的是 `v1.0.0-alpha-3` 阶段新增能力的一轮专项审计。

当前状态：

- 所有本轮发现的问题均已修复
- 修复后的代码已通过：
  - `go test ./...`
  - `go vet ./...`
  - `go test -race ./...`

本文件保留为审计记录，同时用于说明本轮问题和修复方向。

## 审计范围

- `client.go`
- `traffic_policy.go`
- `proxy.go`
- `internal/request/cache.go`
- `internal/request/block.go`
- `internal/request/spec.go`
- `internal/transport/transport.go`
- `internal/transport/request_controls.go`

## 发现与处理结果

### P1-001 已修复

问题：

- alpha-3 新增的内存缓存、host/session 控制器、sticky proxy cache 都存在按高基数 key 持续增长的状态表
- 在长时间运行、URL 维度多、session key 多的真实环境里，存在内存膨胀和 DoS 风险

修复：

- 为短缓存增加了有界清理和容量上限
- 为 host/session 控制器增加了空闲 TTL 回收
- 为 sticky proxy cache 增加了空闲清理和容量上限
- 采用按访问次数触发的机会式 sweep，避免额外后台 goroutine

结果：

- 相关内存表不再无界增长
- 长跑环境下的状态残留风险明显降低

### P2-001 已修复

问题：

- host/session 并发槽位会在等待 rate-limit token 时被提前占用
- 低并发、低 RPS 配置下，容易放大队头阻塞并降低公平性

修复：

- 调整了 transport 执行顺序
- 现在先等待 host/session/class limiter，再申请 host/session 并发槽

结果：

- 请求不再“拿着并发槽睡觉”
- host/session 控制在真实抓取流量下更公平

### P2-002 已修复

问题：

- 开启 proxy selector 后，原始 `TransportHook` 拿到的通常是代理包装后的 `RoundTripper`
- 这会削弱 hook 在“代理池 + 自定义 TLS / 浏览器回退”场景中的实际可用性

修复：

- 保留原有 `TransportHook`
- 新增 `ProxyAwareTransportHook`
- 新增 `ProxyAwareTransportHookFunc`
- 新增 `WrapHTTPClientWithProxySelector(...)` helper
- 现在高级 hook 可以先拿到未包代理的 base client，自定义 TLS / transport 后再由 SDK helper 重新挂回 proxy selector

结果：

- 自定义 TLS 和未来 browser-backed client 的接入位更可用
- 保持了原有 API 的兼容性

### P2-003 已修复

问题：

- HTML block 检测原先对 `cloudflare`、`access denied` 这类宽松关键字过于敏感
- 真实环境下容易把普通 HTML 错误页、维护页误判成 challenge

修复：

- 将 challenge 检测拆成“强特征”和“弱特征”
- 强特征优先命中，例如：
  - `cf-browser-verification`
  - `g-recaptcha`
  - `hcaptcha`
  - `cf-chl-`
  - `/cdn-cgi/challenge-platform/`
  - `turnstile`
- 弱特征需要组合命中才会触发 block 判定
- 增加了负例测试，覆盖普通 HTML 页面中出现零散关键字但不应误报的场景

结果：

- block 检测误报率下降
- 商店公开页流量下的稳定性更好

### P3-001 已修复

问题：

- 缓存 key 原先通过 JSON 编码生成
- 每次请求都要构造对象、排序 cookie、再做 JSON marshal，热路径上有不必要的额外开销

修复：

- 改成更轻量的手工字符串 builder
- 继续保留原有隔离维度：
  - `TrafficClass`
  - URL
  - session key
  - `Accept-Language`
  - 显式 `Cookie`
  - `CookieJar` 视图

结果：

- 缓存 key 生成更轻量
- 在高频公开商店页请求路径上的额外分配有所下降

## 结论

`v1.0.0-alpha-3` 这一轮新增能力在修复后已经具备更好的真实环境可用性，尤其是：

- 内存状态表不再无界增长
- host/session 请求控制的公平性更合理
- 代理场景下的 transport 扩展点更实用
- block 检测误报风险更低
- 缓存热路径开销更轻

后续如果继续推进 `pre-v1`，更值得优先做的将是：

- `CHANGELOG.md`
- `docs/compatibility.md`
- endpoint stability / coverage 文档
- contract tests 与 live workflow 结构整理
