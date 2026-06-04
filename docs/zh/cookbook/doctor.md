# Doctor 诊断

当你需要区分 SDK 问题、网络问题、代理问题、凭据问题或 Steam 上游问题时，可以运行 doctor 示例。

```bash
go run ./examples/doctor
```

也可以输出 JSON，方便脚本处理：

```bash
go run ./examples/doctor -json
```

出现 `FAIL` 时进程退出码为 `1`。`WARN` 不会让进程失败，通常表示某个检查被跳过或需要关注。

## 凭据

Doctor 优先读取环境变量：

- `STEAM_API_KEY`
- `STEAM_ACCESS_TOKEN`
- `STEAM_PROXY`
- `STEAM_PUBLIC_INVENTORY_ID`

如果环境变量为空，会兜底读取 `examples/live/key.txt`、`examples/live/access-token.txt` 和 `examples/live/proxy.txt`。

输出中不会打印 secret。Proxy URL 会先脱敏再输出。

## 边界

Doctor 默认执行 live 网络检查，因为它用于本地诊断。普通测试不会自动运行 doctor；doctor 失败也不自动等同于 SDK 回归。
