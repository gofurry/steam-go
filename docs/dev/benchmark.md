# Benchmark Guide

`v1.3.5` adds a small benchmark baseline for runtime hot paths. These benchmarks are intended for manual review before runtime, cache, proxy, or transport changes; they are not a hard CI gate yet.

Run the baseline with:

```bash
go test -bench='(Cache|Transport|RequestControl|Proxy|RawHTTP)' -benchmem ./...
```

Current benchmark groups:

- `BenchmarkCacheHitMiss`: in-memory cache hit/miss allocation and lookup cost.
- `BenchmarkTransportContextCookieJar`: per-request CookieJar override path.
- `BenchmarkRequestControlHighCardinality`: host/session control map growth and pruning pressure.
- `BenchmarkStickyProxySelector`: sticky proxy session lookup path.
- `BenchmarkRawHTTPReadLimit`: raw HTTP buffered response read path under a body limit.

Use benchmark output as a comparison point inside PR review. Do not optimize these numbers by weakening request safety, cache isolation, redaction, context cancellation, or proxy/cookie boundaries.
