# Release Checklist

Use this checklist before publishing a `steam-go` release.

## Version Scope

- [ ] Confirm the target version and release type.
- [ ] Confirm the release matches the roadmap scope.
- [ ] Confirm no unintended breaking change is included.
- [ ] Confirm new public API additions are compatible with `docs/governance/compatibility.md`.
- [ ] Confirm new official endpoints update `docs/governance/endpoint-coverage.md`.
- [ ] Confirm new Web surfaces are documented as unofficial or volatile.
- [ ] Confirm new addons document what they do and what they do not do.

## Local Validation

- [ ] `go test ./...`
- [ ] `go test -race ./...`
- [ ] `go vet ./...`
- [ ] `staticcheck ./...`
- [ ] `govulncheck ./...`
- [ ] `go run ./internal/tools/apidiffcheck -base <previous-tag> -incompatible`
- [ ] `go run ./internal/tools/steamapi-sync -output-dir docs/api`
- [ ] Review generated API coverage drift under `docs/api/coverage-diff.md`.
- [ ] Examples compile or the release notes explain why they were not checked.
- [ ] Live smoke examples only run through explicit opt-in.

## Documentation

- [ ] README examples still compile or remain copyable.
- [ ] `docs/api/reference.md` is updated when official API behavior changes.
- [ ] `docs/web/reference.md` is updated when `client.Web.*` behavior changes.
- [ ] `docs/addons/reference.md` is updated when addon behavior changes.
- [ ] Governance documents are updated when compatibility, coverage, or stability expectations change.
- [ ] Release notes are added under `docs/releases/`.
- [ ] Chinese documentation is updated when the change affects user-facing guidance.

## Security and Secrets

- [ ] No real API keys, access tokens, refresh tokens, cookies, proxy passwords, or credential-bearing URLs are committed.
- [ ] New logs, examples, errors, and diagnostics avoid printing secrets.
- [ ] URL examples use `steam.RedactSensitiveURL(...)` or clearly show redaction.
- [ ] Error body previews remain bounded.
- [ ] Examples that need secrets use environment variables or hidden terminal prompts.
- [ ] Mutating addon behavior remains explicit and opt-in.

## CI and Tooling

- [ ] Mainline CI passes.
- [ ] Pinned `staticcheck` and `govulncheck` versions are still intentional.
- [ ] Latest toolchain advisory failures, if any, are reviewed and either fixed or documented.
- [ ] Dependency updates are reviewed for compatibility and security impact.

## Release Notes

- [ ] Summarize user-facing changes.
- [ ] Summarize developer-facing changes.
- [ ] List any compatibility notes.
- [ ] List validation commands that passed.
- [ ] Mention known limitations or deferred work.
- [ ] Tag and publish only after the checklist is complete.
