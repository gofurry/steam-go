# Fixture and Smoke Maintenance

`steam-go` keeps stable, sanitized fixtures under `testdata/fixtures` so payload drift can be caught by repeatable tests.

## Fixture Rules

- Use short synthetic or sanitized public payloads.
- Do not commit API keys, access tokens, refresh tokens, cookies, proxy credentials, or credential-bearing URLs.
- Keep fixtures small enough to review.
- Extra upstream fields are allowed when they help prove decode compatibility.
- Volatile nested payloads should stay as `json.RawMessage` unless the project intentionally commits to their structure.

## Golden Snapshots

Golden files live under `testdata/golden`.

Use them for stable helper output such as:

- asset URL generation
- redaction output
- generated coverage output

Do not use golden snapshots for live Steam payloads that are expected to drift.

## Live Smoke

Live smoke checks are opt-in:

```bash
STEAM_GO_LIVE=1 go test ./examples/live/...
```

Normal `go test ./...` must not require network access or Steam credentials. Live smoke output must avoid printing secrets; use redaction helpers before logging URLs or headers.
