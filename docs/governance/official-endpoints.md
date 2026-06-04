# Adding Official Endpoints

Official Steam Web API coverage is tracked by the local maintenance tool:

```bash
go run ./internal/tools/steamapi-sync -output-dir docs/api
```

The generated reports are:

- [coverage.generated.md](../api/coverage.generated.md): full official inventory and SDK coverage.
- [coverage-diff.md](../api/coverage-diff.md): missing, version-mismatched, or SDK-only entries.
- [coverage.generated.json](../api/coverage.generated.json): stable machine-readable snapshot.

## Coverage Status

- `covered`: Steam lists the endpoint and the SDK exposes a typed or raw method for the same interface, method, and version.
- `missing`: Steam lists the endpoint, but the SDK does not expose that exact endpoint.
- `version_mismatch`: Steam and the SDK both know the interface and method, but the version differs.
- `extra_sdk`: the SDK exposes an endpoint that is not currently present in Steam's public `GetSupportedAPIList` response.

`extra_sdk` is a drift signal, not an automatic removal request. Some useful Steam endpoints are not always listed by the public inventory.

## Endpoint Addition Flow

1. Pick one endpoint from `coverage-diff.md` and confirm it fits the project boundary.
2. Add or update the internal endpoint path constant.
3. Add the service method and `Raw` method under the matching `api/*` package.
4. Add request validation, response types, and unit tests with local HTTP fixtures.
5. If this is a new API group, wire the service into `client.API` as an additive public field.
6. Run `go run ./internal/tools/steamapi-sync -output-dir docs/api` and review the generated diff.
7. Update API docs, examples, or cookbook pages when the endpoint changes user-facing behavior.

Do not add mutating, account-automation, purchase, sale, trade, or browser-backed behavior without a separate design discussion.

## Release Checks

Before release, run the sync tool and review any drift. A changed report is not automatically a release blocker, but unexplained drift should be documented in the release notes or roadmap.
