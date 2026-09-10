# SteamID parsing and conversion

Import `github.com/gofurry/steam-go/steamid` for local identity parsing and conversion. It uses only the Go standard library, needs no credentials, and never makes network requests. `Valid()` checks the structure of an ID; it does not check whether an account exists.

[中文](zh/steamid.md) · [Offline example](../examples/steamid/main.go)

## Quick start

```go
import (
    "fmt"

    "github.com/gofurry/steam-go/steamid"
)

id, err := steamid.Parse("12345")
if err != nil {
    return err
}
fmt.Println(id.String())    // 76561197960278073
fmt.Println(id.AccountID()) // 12345

steam2, err := id.Steam2()
if err != nil {
    return err
}
steam3, err := id.Steam3()
if err != nil {
    return err
}
fmt.Println(steam2) // STEAM_1:1:6172
fmt.Println(steam3) // [U:1:12345]
```

Run the complete example with `go run ./examples/steamid`.

## Public API

| API | Meaning |
|---|---|
| `type ID uint64` | Canonical packed SteamID64 value; zero is invalid. |
| `type Universe uint8`, `type AccountType uint8` | Typed Steamworks component enums. |
| `Parse(string) (ID, error)` | Parse AccountID, SteamID64, Steam2, or Steam3; does not accept URLs. |
| `ParseSteamID64(string) (ID, error)` | Parse decimal packed SteamID64 and validate its structure. |
| `ParseSteam2(string) (ID, error)` | Parse `STEAM_X:Y:Z`. |
| `ParseSteam3(string) (ID, error)` | Parse `[type:universe:account[:instance]]`. |
| `ParseCommunityURL(string) (ID, error)` | Parse a numeric Steam Community Individual profile URL. |
| `New(uint32, Universe, AccountType, uint32) (ID, error)` | Construct from AccountID, universe, account type, and instance. |
| `NewIndividual(uint32) (ID, error)` | Construct a Public Individual Desktop ID from an AccountID. |
| `ID.Uint64() uint64`, `ID.String() string` | Raw packed value and canonical decimal text. |
| `ID.AccountID() uint32`, `ID.Instance() uint32` | Low 32-bit AccountID and 20-bit instance. |
| `ID.Universe() Universe`, `ID.AccountType() AccountType` | 8-bit universe and 4-bit account type. |
| `ID.Valid() bool` | Independent structural validation, including values created by direct casts. |
| `ID.Steam2() (string, error)`, `ID.Steam3() (string, error)` | Lossless conversion where the target format can represent the ID. |

The bit layout is AccountID at bits 0–31, Instance at 32–51, AccountType at 52–55, and Universe at 56–63. `New` rejects instances above `0xFFFFF`; it never masks an overflowing input into a different ID.

Universe constants are `UniverseInvalid=0`, `UniversePublic=1`, `UniverseBeta=2`, `UniverseInternal=3`, and `UniverseDev=4`. Account type constants use the `AccountType` prefix:

| Suffix | Value | Steam3 character |
|---|---:|---|
| Invalid | 0 | Unsupported |
| Individual | 1 | `U` |
| Multiseat | 2 | `M` |
| GameServer | 3 | `G` |
| AnonGameServer | 4 | `A` |
| Pending | 5 | `P` |
| ContentServer | 6 | `C` |
| Clan | 7 | `g` |
| Chat | 8 | `T`, `c` (clan chat), `L` (lobby chat) |
| ConsoleUser | 9 | Unsupported; no invented character |
| AnonUser | 10 | `a` |

Common instances are `InstanceAll=0`, `InstanceDesktop=1`, `InstanceConsole=2`, and `InstanceWeb=4` (all `uint32`).

## Parsing and normalization

All parsers trim surrounding whitespace. Decimal text must contain ASCII digits only: no signs, hexadecimal, floats, or exponents. Leading zeros are accepted; `String()` emits canonical decimal text.

`Parse` treats positive decimal values through `4294967295` as AccountIDs, filling in Public + Individual + Desktop. Larger numbers are interpreted as packed SteamID64. `ParseSteamID64("12345")` returns `ErrInvalidID`, while `Parse("12345")` succeeds. AccountID `0` is invalid for an Individual.

Steam2 prefixes are case-insensitive. `STEAM_0` and `STEAM_1` both mean Public; canonical output is `STEAM_1`. The parity digit `Y` must be `0` or `1`, and `Z*2+Y` must fit in uint32. Universes 2–4 retain their meaning. Parsing sets Individual + Desktop. Formatting any other type or instance returns `ErrUnsupportedConversion` rather than discarding information.

Steam3 characters are case-sensitive: `g` is Clan and `G` is GameServer. Both `[U:1:12345:2]` and the legacy `[U:1:12345(2)]` are accepted; output always uses colons. Missing instances default to 0 for `g/T/c/L` and 1 for other supported characters. Steam3 universe 0 is invalid; only Steam2 has the historical universe-0 alias.

For `c` and `L`, parsing ORs the instance with the clan flag `0x80000` or lobby flag `0x40000`. Other instance bits are preserved. Formatting omits the instance only when it equals the character's default including its flag; Multiseat and AnonGameServer always include it. Extra bits use the full instance value, for example `[c:1:12345:524291]`. If both flags exist, output uses `c` with the full instance. For every representable valid ID, parsing the formatted Steam2 or Steam3 string returns the same ID.

## Structural validation and errors

`Valid()` accepts universes 1–4 and account types 1–10. Individual requires a nonzero AccountID and instance 0–4, including 3. Clan requires a nonzero AccountID and instance 0. GameServer requires a nonzero AccountID. Other known types receive only the common structural checks, so some valid non-Individual IDs have AccountID 0.

Every successful parser or constructor returns a valid ID. A direct `ID(raw)` cast can create an invalid value; `String()` and component access still work. Formatting an invalid ID returns `ErrInvalidID`.

Use `errors.Is`, not error-message comparisons:

| Sentinel | Meaning |
|---|---|
| `ErrInvalidFormat` | Malformed text, unknown Steam3 character, or unsupported URL shape/host. |
| `ErrInvalidID` | Numeric overflow or invalid component combination. |
| `ErrUnsupportedConversion` | Valid ID cannot be expressed losslessly in the requested format. |
| `ErrVanityReference` | Community vanity URL requires explicit network resolution. |

Failures return zero ID (or an empty formatted string). Error messages do not echo the input URL or its credentials.

## Community URLs and vanity resolution

`ParseCommunityURL` accepts `http` or `https` on the exact hostname `steamcommunity.com` or `www.steamcommunity.com`. `/profiles/<SteamID64>` must identify a structurally valid Individual. One trailing slash, query, and fragment are allowed. Similar-looking hosts, userinfo, extra path segments, group/workshop URLs, and `steam://` URIs are rejected.

`/id/<vanity>` returns `ErrVanityReference`. The local parser never calls Steam. For network resolution, pass the vanity **token**, such as `example-user`, to the existing `client.API.SteamUser.ResolveVanityURL(ctx, token, nil)` capability. Handle that API's response before parsing any returned SteamID64. See the [API reference](api/reference.md).

The public `steamid` package is opt-in. Existing SDK endpoints and `internal/steamid.ValidateSteamID64` retain their permissive decimal uint64 request validation. They do not automatically adopt these stricter identity rules. This addition does not include profile fetching, friend/invite codes, authentication, or PICS.

## Semantic references

- [Steamworks CSteamID, EUniverse and EAccountType](https://partner.steamgames.com/doc/api/steam_api): component enums, including `ConsoleUser=9`.
- [SteamKit SteamID](https://github.com/SteamRE/SteamKit/blob/master/SteamKit2/SteamKit2/Types/SteamID.cs): structural validity and Steam3 chat flags. This package additionally preserves non-default instance bits when formatting, and uses `STEAM_1` for canonical Public Steam2 output.
