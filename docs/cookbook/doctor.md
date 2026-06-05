# Doctor Diagnostics

Use the doctor example when you need to separate SDK problems from network, proxy, credential, or Steam upstream issues.

```bash
go run ./examples/doctor
```

JSON output is available for scripts:

```bash
go run ./examples/doctor -json
```

## JSON Schema

The JSON form is intended for local scripts and release diagnostics:

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

`summary.ok`, `summary.warn`, and `summary.fail` count checks by final status. Each item in `checks` contains:

- `category`: logical group, such as `environment`, `credential`, `official_api`, `web`, or `proxy`.
- `name`: stable check name inside the example.
- `status`: one of `OK`, `WARN`, or `FAIL`.
- `message`: short human-readable result.
- `detail`: optional additional context, already redacted when present.

`examples/doctor` is still an example, not a stable CLI product. Treat this JSON as a practical diagnostic format rather than a long-term compatibility contract.

## Exit Codes

- `0`: all checks are `OK` or `WARN`.
- `1`: at least one check is `FAIL`.
- `2`: flag parsing, configuration, or output rendering failed before a complete diagnostic result could be produced.

`WARN` does not fail the process; it usually means a check was skipped or needs attention.

## Credentials

Doctor reads credentials from environment variables first:

- `STEAM_API_KEY`
- `STEAM_ACCESS_TOKEN`
- `STEAM_PROXY`
- `STEAM_PUBLIC_INVENTORY_ID`

It then falls back to `examples/live/key.txt`, `examples/live/access-token.txt`, and `examples/live/proxy.txt`.

Secrets are never printed. Output must not include API keys, access tokens, cookies, raw headers, raw query strings, or proxy passwords. Proxy URLs are redacted before output.

## Boundary

Doctor performs live network checks by default because it is meant for local diagnostics. It is not run by normal tests, and a doctor failure is not automatically an SDK regression.
