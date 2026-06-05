# Doctor 诊断

当你需要区分 SDK 问题、网络问题、代理问题、凭据问题或 Steam 上游问题时，可以运行 doctor 示例。

```bash
go run ./examples/doctor
```

也可以输出 JSON，方便脚本处理：

```bash
go run ./examples/doctor -json
```

## JSON 结构

JSON 输出适合本地脚本和发布诊断使用：

```json
{
  "summary": {
    "ok": 3,
    "warn": 1,
    "fail": 0
  },
  "checks": [
    {
      "category": "official_api",
      "name": "steamwebapiutil.get_server_info",
      "status": "OK",
      "message": "reachable",
      "detail": "optional redacted detail"
    }
  ]
}
```

`summary.ok`、`summary.warn`、`summary.fail` 分别表示三类状态的检查数量。`checks` 中每个条目包含：

- `category`：逻辑分组，例如 `environment`、`credential`、`official_api`、`web`、`proxy`。
- `name`：示例内部的稳定检查名。
- `status`：只会是 `OK`、`WARN`、`FAIL` 之一。
- `message`：简短的人类可读结果。
- `detail`：可选补充信息；如果存在，也应已经脱敏。

`examples/doctor` 仍然是 example，不是稳定 CLI 产品。可以把 JSON 当作实用诊断格式使用，但不要把它视为长期兼容契约。

## 退出码

- `0`：所有检查都是 `OK` 或 `WARN`。
- `1`：至少一个检查为 `FAIL`。
- `2`：flag、配置或输出渲染在完整诊断结果生成前失败。

`WARN` 不会让进程失败，通常表示某个检查被跳过或需要关注。

## 凭据

Doctor 优先读取环境变量：

- `STEAM_API_KEY`
- `STEAM_ACCESS_TOKEN`
- `STEAM_PROXY`
- `STEAM_PUBLIC_INVENTORY_ID`

如果环境变量为空，会兜底读取 `examples/live/key.txt`、`examples/live/access-token.txt` 和 `examples/live/proxy.txt`。

输出中不会打印 secret，包括 API key、access token、cookie、raw header、raw query string 或 proxy password。Proxy URL 会先脱敏再输出。

## 边界

Doctor 默认执行 live 网络检查，因为它用于本地诊断。普通测试不会自动运行 doctor；doctor 失败也不自动等同于 SDK 回归。
