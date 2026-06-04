# 新增官方 Endpoint

官方 Steam Web API 覆盖范围通过本地维护工具追踪：

```bash
go run ./internal/tools/steamapi-sync -output-dir docs/api
```

生成的报告包括：

- [coverage.generated.md](../../api/coverage.generated.md)：完整官方 inventory 与 SDK 覆盖范围。
- [coverage-diff.md](../../api/coverage-diff.md)：缺失、版本不一致或仅 SDK 中存在的条目。
- [coverage.generated.json](../../api/coverage.generated.json)：稳定的机器可读快照。

## 覆盖状态

- `covered`：Steam 列出了该 endpoint，SDK 也暴露了相同 interface、method、version 的 typed 或 raw 方法。
- `missing`：Steam 列出了该 endpoint，但 SDK 没有暴露相同 endpoint。
- `version_mismatch`：Steam 与 SDK 都知道该 interface 和 method，但 version 不一致。
- `extra_sdk`：SDK 暴露了该 endpoint，但它当前不在 Steam 公开 `GetSupportedAPIList` 响应中。

`extra_sdk` 是 drift 信号，不是自动删除要求。有些实用 Steam endpoint 不一定总会出现在公开 inventory 中。

## 新增流程

1. 从 `coverage-diff.md` 选择一个 endpoint，并确认它符合项目边界。
2. 新增或更新 internal endpoint path 常量。
3. 在对应 `api/*` package 中新增 service method 和 `Raw` method。
4. 补请求校验、响应类型，以及基于本地 HTTP fixture 的单元测试。
5. 如果是新的 API group，以 additive public field 的方式接入 `client.API`。
6. 运行 `go run ./internal/tools/steamapi-sync -output-dir docs/api` 并 review 生成报告。
7. 如果影响用户使用方式，同步更新 API 文档、示例或 cookbook。

不要在没有单独设计讨论的情况下加入写操作、账号自动化、购买、出售、交易或 browser-backed 行为。

## Release 检查

发版前运行 sync 工具并 review drift。报告变化不自动构成 release blocker，但无法解释的 drift 应记录在 release notes 或 roadmap 中。
