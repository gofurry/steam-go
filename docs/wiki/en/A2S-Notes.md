# A2S Notes

`addons/a2s` is used for querying game servers directly.

It is different from Steam Web API calls.

## Web API vs A2S

| Steam Web API | A2S |
|---|---|
| HTTP-based | UDP-based game server query |
| Queries Steam platform data | Queries a specific game server |
| Uses SDK client and credentials when needed | Does not use Steam Web API key |
| Good for player profiles, achievements, app data | Good for server info, players, rules |

`client.API.GameServersService.GetServerList` is a Steam Web API discovery helper. It returns server candidates from Steam's Web API side; use A2S afterward when you need current server info, players, or rules.

## Supported Query Types

The A2S addon exposes common server query operations:

- info
- players
- rules

## Examples

Server info:

```bash
go run ./examples/a2s -server 1.2.3.4:27015 -query info
```

Players:

```bash
go run ./examples/a2s -server 1.2.3.4:27015 -query players
```

Rules:

```bash
go run ./examples/a2s -server 1.2.3.4:27015 -query rules
```

## Related Packages

- `addons/a2s`
- `addons/a2s/master`
- `addons/a2s/scanner`

## Practical Advice

- Treat A2S as network probing against game servers.
- Treat `GetServerList` results as candidates, not proof that every server still matches every filter.
- Expect timeouts and partial failures.
- Use bounded concurrency when scanning.
- Do not mix A2S error handling with Web API error handling.
- Keep A2S as an addon so the core Web API client stays small.

