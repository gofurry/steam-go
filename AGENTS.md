# AGENTS.md

`steam-go` is a stable Go SDK for Steam APIs, carefully scoped read-only Steam Web surfaces, and optional higher-level addons.

This file is the shortest maintenance entrypoint. Read only the document relevant to the current task.

## Repository Map

- `client.go`, root package — client construction, options, retry/rate-limit/proxy/traffic policy, shared public infrastructure.
- `api/` — Steam Web API service bindings exposed through `client.API.*`.
- `web/` — read-only Storefront / Community / Market surfaces exposed through `client.Web.*`.
- `addons/` — optional higher-level features composed from SDK capabilities.
- `internal/` — implementation details; never part of the public import surface.
- `examples/` — minimal public usage examples.
- `docs/` — user-facing documentation.

Architecture details: `.agents/architecture.md`  
Common change workflows: `.agents/playbook.md`

## Hard Contracts

Before changing public behavior, read:

- `contracts/compatibility.md`
- `contracts/upstream.md`

`contracts/*` is normative. Descriptive docs must not contradict it.

## Core Rules

- Preserve v1 compatibility; prefer additive changes.
- Keep raw Steam fields when adding normalized or derived views.
- Unknown upstream values must degrade safely instead of corrupting meaning.
- Do not invent Steam data or infer historical facts from current state.
- Prefer authoritative Steam-returned URLs over guessed CDN paths.
- Keep unofficial Web surfaces read-only.
- Do not move addon behavior into core API packages without a strong boundary reason.
- Never expose API keys, access tokens, cookies, proxy credentials, or secrets in logs, errors, fixtures, or examples.

## Change Routing

| Change | Default location |
|---|---|
| New Steam Web API endpoint | `api/<service>/` |
| Storefront / Community / Market JSON surface | `web/<surface>/` |
| Normalization of an existing Steam field | Package owning that field |
| Resource discovery / verify / read / download | `addons/assets/` |
| Request / retry / proxy / traffic infrastructure | root package or `internal/` |
| Public usage example | `examples/` |
| User-facing docs | `docs/` |

## Validation

For code changes, run at minimum:

```bash
go test ./...
go test -race ./...
go vet ./...
```

Run `staticcheck ./...` and `govulncheck ./...` when available.

## Documentation

Public behavior changes should update:

- relevant reference or example docs
- `CHANGELOG.md`

Do not rewrite unrelated documentation.
