# Doctor Diagnostics

Use the doctor example when you need to separate SDK problems from network, proxy, credential, or Steam upstream issues.

```bash
go run ./examples/doctor
```

JSON output is available for scripts:

```bash
go run ./examples/doctor -json
```

`FAIL` exits with code `1`. `WARN` does not fail the process; it usually means a check was skipped or needs attention.

## Credentials

Doctor reads credentials from environment variables first:

- `STEAM_API_KEY`
- `STEAM_ACCESS_TOKEN`
- `STEAM_PROXY`
- `STEAM_PUBLIC_INVENTORY_ID`

It then falls back to `examples/live/key.txt`, `examples/live/access-token.txt`, and `examples/live/proxy.txt`.

Secrets are never printed. Proxy URLs are redacted before output.

## Boundary

Doctor performs live network checks by default because it is meant for local diagnostics. It is not run by normal tests, and a doctor failure is not automatically an SDK regression.
