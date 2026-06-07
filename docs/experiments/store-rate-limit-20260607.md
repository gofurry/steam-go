# Storefront AppDetails Rate-Limit Experiment - 2026-06-07

This note records a small empirical rate-limit experiment against Steam Storefront `appdetails`.

It is not an official Steam limit. Treat it as a reproducible observation for one network environment, one client profile, and one short time window. Use it to choose conservative production defaults, not to bypass upstream controls.

## Scope

- Endpoint: `store.steampowered.com/api/appdetails`
- SDK method: `client.Web.Storefront.GetAppDetailsRaw`
- Traffic class: `TrafficClassPublicStorePage`
- Date: 2026-06-07
- AppIDs: `440`, `570`, `730`
- Regions: `CN`, `US`, `HK`
- Languages: `schinese`, `english`
- Request cases per run: `3 appids * 3 regions * 2 languages * repeat 20 = 360`
- Workers: `10`
- Local Store interval: `0s`
- Burst: `0`
- Retry: `0`
- Local cooldown after block: `0s`
- Proxy: disabled
- Timeout: `15s`

The second run was started after roughly five minutes.

## Results

| Run ID | Total | OK | Failed | Blocked | HTTP 429 | HTTP 403 | Transport | Duration |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| `20260607-183228` | 360 | 228 | 132 | 105 | 31 | 74 | 27 | `1m35s` |
| `20260607-184100` | 360 | 228 | 132 | 132 | 34 | 98 | 0 | `20.55s` |

Both runs accepted exactly 228 successful Store appdetails requests and then failed the remaining 132 logical requests. The first run had 27 transport timeouts, while the second run returned explicit `429` / `403` responses for all failures.

## Timeline

| Run ID | First 429 | First 403 | First block-detected signal |
| --- | --- | --- | --- |
| `20260607-183228` | about 20s, seq `233` | about 22s, seq `267` | about 20s, seq `233` |
| `20260607-184100` | about 11s, seq `221` | about 12s, seq `263` | about 11s, seq `221` |

The second run reached the same request-count boundary faster because it did not spend time on the initial transport timeouts seen in the first run.

## Interpretation

- Storefront appdetails showed a repeatable request-count boundary in this environment.
- `429` started around 220-230 requests.
- Continuing after `429` quickly led to `403` / block-detected responses around the next few dozen requests.
- Waiting about five minutes was enough for a repeat test to reach a similar boundary again.
- The observed pattern matches a practical Store budget around `150-250` requests per five minutes per egress identity.

## Production Guidance

For production access to Storefront appdetails:

- Keep `TrafficClassPublicStorePage` separate from official API traffic.
- Start around `1 request / 2 seconds` for Store appdetails, with `burst=1`.
- Prefer queueing and caching over high worker counts.
- Treat `429`, `403`, and `BlockDetected=true` as hard slowdown signals.
- Keep a cooldown after Store failures instead of immediately retrying.
- Avoid using short benchmark peak throughput as a production target.

If each application detail collection needs three regional appdetails requests, a conservative `150 requests / 5 minutes` budget means roughly:

```text
150 Store requests / 5 minutes / 3 regional requests ~= 50 apps / 5 minutes
```

Leave additional room when Store events, reviews, or other Storefront helpers share the same traffic bucket.

## Official API Note

Short-term official API bursts are a different problem from Storefront page-like traffic. Even if short-term requests do not return `429`, production callers must still honor their daily key or application budget. If your developer key is budgeted at `10,000` calls per day, that daily budget is the controlling constraint for tasks such as current-player collection.
