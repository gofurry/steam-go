# SteamID Model

SteamID is the identity model used across the Steam ecosystem.

It is not limited to player accounts. A SteamID can identify individuals, clans, game servers, chat objects, anonymous accounts, and other Steam account types.

`steam-go` exposes local SteamID parsing and conversion through:

```go
import "github.com/gofurry/steam-go/steamid"
```

The package is entirely local. It does not require a Steam Web API key and does not make network requests.

## The 64-bit Identity

The canonical Steam identity is a packed 64-bit value:

```text
63                 56 55      52 51                 32 31                 0
┌────────────────────┬───────────┬─────────────────────┬────────────────────┐
│ Universe   8 bits  │ Type 4bit │ Instance    20 bits │ AccountID  32 bits │
└────────────────────┴───────────┴─────────────────────┴────────────────────┘
```

Conceptually:

```text
SteamID
=
Universe
+ AccountType
+ Instance
+ AccountID
```

These components give a SteamID more meaning than a plain numeric user ID.

For example, the same 32-bit AccountID could theoretically belong to identities with different account types or instances.

## Common Representations

Steam identities appear in several common forms.

| Representation | Example | Typical use |
|---|---|---|
| AccountID | `12345` | Low 32-bit account identifier |
| SteamID64 | `76561197960278073` | Steam Web API and Community profiles |
| Steam2 | `STEAM_1:1:6172` | Source / GoldSrc server tooling |
| Steam3 | `[U:1:12345]` | Modern typed Steam identity representation |

These are not four unrelated identifiers.

They are different representations of Steam identity data.

For a normal Public Individual account:

```text
AccountID: 12345

        ↓

SteamID64:
76561197960278073

Steam2:
STEAM_1:1:6172

Steam3:
[U:1:12345]
```

## AccountID

`AccountID` is only the low 32 bits of a SteamID.

By itself, it does not contain:

- Universe
- AccountType
- Instance

Because of this, converting an AccountID into a full SteamID requires additional assumptions.

`steam-go` interprets a standalone AccountID through `steamid.Parse` as a Public Individual Desktop identity.

For explicit construction, use:

```go
id, err := steamid.NewIndividual(12345)
```

or provide all components with `steamid.New`.

## SteamID64

SteamID64 is the packed 64-bit representation commonly seen in Steam Web APIs.

For example:

```text
76561197960278073
```

A numeric Steam Community profile URL also contains this form:

```text
https://steamcommunity.com/profiles/76561197960278073
```

In `steam-go`:

```go
id, err := steamid.ParseSteamID64("76561197960278073")
```

Use `ParseCommunityURL` when parsing the URL form.

## Steam2

Steam2 IDs use the familiar Source-style representation:

```text
STEAM_X:Y:Z
```

Example:

```text
STEAM_1:1:6172
```

They are mainly meaningful for Individual accounts and are common in Source / GoldSrc server administration, ban lists, and older tooling.

Historical software may expose Public universe IDs as:

```text
STEAM_0:...
```

while the canonical Public universe is `1`.

`steam-go` accepts both `STEAM_0` and `STEAM_1` for Public IDs and emits `STEAM_1` as its canonical representation.

Steam2 cannot represent every possible Steam identity without losing information. `steam-go` therefore returns an error instead of inventing a lossy conversion.

## Steam3

Steam3 explicitly carries an account type character:

```text
[U:1:12345]
```

Common examples include:

| Character | Meaning |
|---|---|
| `U` | Individual |
| `G` | Game server |
| `A` | Anonymous game server |
| `g` | Clan |
| `T` | Chat |
| `c` | Clan chat |
| `L` | Lobby |
| `a` | Anonymous user |

Steam3 may also contain an explicit instance:

```text
[U:1:12345:2]
```

The account type is important. `[U:...]` and `[g:...]`, for example, are not interchangeable even when their AccountID component is the same.

## Community URLs and Vanity Names

Steam Community commonly uses two profile URL forms:

```text
/profiles/<SteamID64>
```

and:

```text
/id/<vanity>
```

A numeric `/profiles/` URL can be parsed locally:

```go
id, err := steamid.ParseCommunityURL(
    "https://steamcommunity.com/profiles/76561197960278073",
)
```

A vanity URL such as:

```text
https://steamcommunity.com/id/example-user
```

does not contain the SteamID.

Resolving it requires a Steam network request. `steam-go` deliberately keeps this separate: `steamid.ParseCommunityURL` reports a vanity reference, while the existing `client.API.SteamUser.ResolveVanityURL` capability performs network resolution.

## Parsing with steam-go

For general local input:

```go
id, err := steamid.Parse("STEAM_1:1:6172")
if err != nil {
    return err
}

fmt.Println(id.String())
fmt.Println(id.AccountID())
fmt.Println(id.Universe())
fmt.Println(id.AccountType())
fmt.Println(id.Instance())
```

`Parse` accepts:

- AccountID
- SteamID64
- Steam2
- Steam3

It does not automatically resolve vanity names or fetch profile data.

## Practical Advice

- Use SteamID64 when interacting with most Steam Web APIs.
- Use Steam2 when interoperating with Source / GoldSrc tooling that expects it.
- Prefer Steam3 when the account type and instance are important.
- Do not treat AccountID as a complete Steam identity without knowing the missing components.
- Do not assume every numeric `uint64` is a structurally valid SteamID.
- Do not perform lossy conversion between representations silently.
- Treat vanity names as references that require explicit network resolution.

For the complete `steam-go` API, validation rules, conversion behavior, and error model, see [`docs/steamid.md`](https://github.com/gofurry/steam-go/blob/main/docs/steamid.md).