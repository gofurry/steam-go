# 路线图

## 当前阶段

`steam-go` 已完成 `v1.0.0-alpha-1` 到 `v1.0.0-alpha-3` 的核心能力建设：

- 统一 root `Client` 与 `client.API.*` 服务分组
- 生产导向的 retry、rate limit、safe defaults、proxy、traffic policy、request controls
- `addons/openid` 与 `addons/a2s` 的 addon 边界验证
- 已落地的 CI 基线：`go test`、`go test -race`、`go vet`、`staticcheck`、`govulncheck`

从这个阶段开始，到 `v1.0.0` 正式版之前，**暂停扩充新的 API 覆盖**。  
接下来的重点不再是继续接接口，而是明确稳定边界、冻结 public surface、整理发布文档，并完成 RC 阶段的验证与收尾。

## 路线策略

`v1.0.0` 前的推进顺序：

1. 冻结稳定 public API
2. 明确 compatibility 与 experimental 边界
3. 完成发布导向文档与 examples 收口
4. 只修 bug、补测试、补文档、做发布治理
5. `v1.1.0+` 再恢复官方 Steam Web API 覆盖扩展

短期目标：

- `v1.0.0-rc.1`：稳定面冻结与发布治理文档
- `v1.0.0-rc.2`：硬化、回归验证、发布收尾
- `v1.0.0`：首个正式稳定版

中期目标：

- `v1.1.0+`：在稳定边界明确之后，恢复官方 Steam Web API 扩展

长期目标：

- Store / Community / CDN 等高波动方向继续保持在 `v1` 兼容承诺之外，必要时通过 experimental 包或 addon 推进

## 版本计划

### v1.0.0-rc.1 - 稳定面冻结

**状态：** Current  
**范围：** Stability / Documentation / Release Governance  
**目标：** 冻结 `v1.0.0` 的稳定 public surface，明确兼容性承诺与非承诺范围。

#### 重点

- 兼容性策略
- 发布治理文档
- endpoint 覆盖与稳定性说明
- README / 中文文档口径收口

#### 任务

- [ ] 增加 `docs/compatibility.md`
- [ ] 增加 `docs/release-v1.md`
- [ ] 增加 `CHANGELOG.md`
- [ ] 增加 `docs/endpoint-stability.md`
- [ ] 增加 `docs/endpoint-coverage.md`
- [ ] 复查 `README.md` 与 `docs/README_zh.md`，统一 RC 口径
- [ ] 明确 `v1.0.0` 前不再扩充新的 API 覆盖
- [ ] 明确 Store / Community / CDN / browser fallback 不属于 `v1.0.0` 稳定承诺

#### 验收标准

- `v1.0.0` 稳定面有明确书面定义
- 用户能区分 stable、preview、experimental 与 out-of-scope
- 核心文档之间口径一致
- roadmap 不再把新 API 扩展写入 `v1.0.0` 前阶段

---

### v1.0.0-rc.2 - 硬化与验证

**状态：** Planned  
**范围：** Testing / Stability / Documentation  
**目标：** 围绕已冻结的稳定面做回归验证与发布前硬化，只接受 bug fix 和文档修订。

#### 重点

- 回归保护
- examples 可用性
- 测试结构收口
- 发布前文档打磨

#### 任务

- [ ] 补充稳定 public behavior 的 contract-style 覆盖
  - retry / backoff
  - rate limit / safe defaults
  - proxy selection / proxy health
  - traffic policy isolation
  - OpenID verification 边界行为
- [ ] 复查并收口测试组织、命名和超大测试文件
- [ ] 确认 examples、README、docs 与冻结后的 public API 一致
- [ ] 清理 `test` / `race` / `vet` / `staticcheck` / `govulncheck` 暴露的问题
- [ ] 只修 bug、补测试、补文档，不引入新的 API 覆盖或新的架构主线

#### 验收标准

- 核心 public API 有足够回归保护
- examples 可运行且与文档一致
- CI 全绿且没有 blocker 级问题
- `rc.2` 没有引入新的 API 覆盖扩展

---

### v1.0.0 - 首个正式稳定版

**状态：** Planned  
**范围：** Release / Documentation / Stability  
**目标：** 正式发布首个稳定版，定位为面向生产使用的 Steam Web API Go SDK。

#### 重点

- 正式稳定版发布
- 兼容性承诺
- 发布说明

#### 任务

- [ ] 完成最终 release notes
- [ ] 确认稳定 public surface 与 RC 阶段承诺一致
- [ ] 确认 README、核心 docs、compatibility 说明、examples 已完整
- [ ] 打 tag 并发布 `v1.0.0`

#### 验收标准

- public API 命名与行为基本冻结
- 文档已明确 stable scope 与非承诺范围
- 发布时 CI 与验证全绿
- 可以清晰定位为 “A stable Go SDK for Steam Web API with production-oriented request controls.”

---

### v1.1.0 - 恢复官方 API 扩展

**状态：** Deferred  
**范围：** User-facing / Documentation / Testing  
**目标：** 在 `v1.0.0` 边界明确之后，再恢复官方 Steam Web API 覆盖扩展。

#### 重点

- 官方 API 扩展
- typed response 提升
- 新增 examples 与测试

#### 任务

- [ ] 恢复对更多 `api.steampowered.com` 官方 endpoint 的覆盖
- [ ] 对稳定官方 payload 继续提升 typed 覆盖
- [ ] 为新增 method 补文档、examples 和测试

#### 验收标准

- 新增 API 保持向后兼容
- 每个新增 public method 都有文档与测试
- API 扩展不会模糊 stable 与 experimental 的边界

## 发布前检查

- [ ] `v1.0.0` 前不再安排新的 API 扩展
- [ ] `v1.0.0-rc.1` 只做稳定面冻结与边界定义
- [ ] `v1.0.0-rc.2` 只做硬化、回归与发布收尾
- [ ] `v1.0.0` 保留给首个正式稳定版
- [ ] `v1.1.0+` 再恢复官方 Steam Web API 的向后兼容扩展
