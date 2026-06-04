# Fixture 与 Smoke 维护

`steam-go` 将稳定、脱敏的 fixture 放在 `testdata/fixtures`，用于通过可重复测试发现 payload drift。

## Fixture 规则

- 使用短小的合成 payload 或脱敏公开 payload。
- 不提交 API key、access token、refresh token、cookie、proxy 凭据或携带凭据的 URL。
- fixture 应保持容易 review。
- 可以保留额外上游字段，用来证明当前 decode 对新增字段兼容。
- 高波动 nested payload 应保持为 `json.RawMessage`，除非项目明确承诺它的结构。

## Golden 快照

Golden 文件放在 `testdata/golden`。

适合覆盖稳定 helper 输出，例如：

- asset URL generation
- redaction output
- generated coverage output

不要把预期会漂移的 live Steam payload 做成 golden snapshot。

## Live Smoke

Live smoke 必须显式 opt-in：

```bash
STEAM_GO_LIVE=1 go test ./examples/live/...
```

普通 `go test ./...` 不应要求网络或 Steam 凭据。Live smoke 输出不应打印 secret；记录 URL 或 header 前应使用 redaction helpers。
