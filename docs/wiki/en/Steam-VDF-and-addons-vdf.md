# Steam VDF and `addons/vdf`

Steam uses Valve Data Format, also known as VDF or KeyValues, in many local text files.

These files are not Steam Web API endpoints. They are local configuration or metadata files used by the Steam client, Valve tools, Source games, and some game resources.

`addons/vdf` is a small bridge to `github.com/gofurry/vdf-go`. It lets projects that already use `steam-go` parse VDF / KeyValues text files through the addon tree.

## What VDF Looks Like

VDF is a tree-shaped key-value text format:

```vdf
"AppState"
{
    "appid"      "730"
    "name"       "Counter-Strike 2"
    "installdir" "Counter-Strike Global Offensive"
}
```

Common traits:

- keys and values are usually quoted
- `{}` denotes child nodes
- values are strings; numeric or boolean meaning belongs to the caller
- duplicate sibling keys may exist
- node order can matter
- VDF is not JSON

Because duplicate keys and order matter, VDF should not be modeled as a plain map.

## Common Steam Files

Common text VDF / KeyValues files include:

| File | Common use |
|---|---|
| `libraryfolders.vdf` | Steam library folder metadata |
| `appmanifest_*.acf` | local install manifest for installed apps |
| `config.vdf` | Steam client configuration |
| `loginusers.vdf` | locally known Steam users |
| `gameinfo.txt` | Source / Valve game or tool configuration |

Not every `.cfg` file is VDF. Many Source or console `.cfg` files are command-style files using `bind`, `alias`, or cvars. Those are outside the scope of `addons/vdf`.

## Role of `addons/vdf`

`addons/vdf` is intentionally thin.

It does:

- parse text VDF / KeyValues files
- expose the core `vdf-go` document and node model
- preserve duplicate keys and node order
- support read, query, edit, and marshal helpers from `vdf-go`

It does not:

- implement binary VDF
- parse `shortcuts.vdf`
- scan Steam installation directories automatically
- read user directories unless the caller passes a path
- extract accounts, tokens, cookies, or sessions
- provide strong `AppManifest` or `LibraryFolder` business models

If local Steam scanning or strong local Steam models are needed later, they should be designed separately instead of being mixed into `addons/vdf`.

## Import

If you only need a VDF parser, using `vdf-go` directly is fine:

```go
import vdf "github.com/gofurry/vdf-go"
```

If your project already depends on `steam-go`, use the addon entrypoint:

```go
import "github.com/gofurry/steam-go/addons/vdf"
```

## Parse an App Manifest

```go
doc, err := vdf.ParseFile("steamapps/appmanifest_730.acf")
if err != nil {
    panic(err)
}

appid := doc.Path("AppState", "appid").AsString()
name := doc.Path("AppState", "name").AsString()

fmt.Println(appid, name)
```

`appmanifest_*.acf` files may contain app ID, app name, install directory, build ID, install state, depot metadata, and update metadata.

`addons/vdf` only parses the VDF tree. It does not promise a typed app manifest model.

## Parse Library Folders

```go
doc, err := vdf.ParseFile("config/libraryfolders.vdf")
if err != nil {
    panic(err)
}

folders := doc.First("libraryfolders")
if folders == nil {
    return
}

for _, folder := range folders.Children {
    pathNode := folder.First("path")
    if pathNode == nil {
        continue
    }
    fmt.Println(folder.Key, pathNode.AsString())
}
```

Real Steam files can change across client versions. Treat missing or extra fields as normal.

## Duplicate Keys and Queries

`vdf-go` uses a slice-based document model so duplicate keys are not silently overwritten.

Useful helpers:

```go
first := node.First("item")
all := node.All("item")
leaf := doc.Path("AppState", "InstalledDepots", "731", "manifest")
```

- `First` returns the first matching node.
- `All` returns all matching nodes.
- `Path` walks a nested key path.

## Writing VDF

```go
doc := vdf.NewDocument(
    vdf.NewNode("AppState",
        vdf.NewValue("appid", "730"),
        vdf.NewValue("name", "Counter-Strike 2"),
    ),
)

text, err := vdf.MarshalString(doc)
if err != nil {
    panic(err)
}

fmt.Println(text)
```

Marshal output is stable and readable, but it does not preserve original comments, blank lines, or formatting style.

## Safety Notes

For untrusted or large input, use a size limit:

```go
doc, err := vdf.ParseReaderLimit(reader, 1<<20)
```

You can also set parser resource limits:

```go
doc, err := vdf.ParseReaderLimit(reader, 1<<20,
    vdf.WithMaxDepth(128),
    vdf.WithMaxTokenBytes(1<<20),
    vdf.WithMaxNodes(1_000_000),
)
```

`addons/vdf` does not execute `#include` or `#base`, and it does not automatically read additional local files. Known directives are ignored by default; callers can preserve them with `WithPreserveDirectives(true)`.

## Relationship With `steam-go`

`addons/vdf` is not part of the network request layer.

It does not use:

- `Client`
- `client.API.*`
- `client.Web.*`
- traffic policy
- proxy
- retry
- rate limit
- cookie jar

Recommended boundary:

| Capability | Location |
|---|---|
| Steam Web API requests | `client.API.*` |
| Storefront / Community / Market read-only JSON | `client.Web.*` |
| A2S queries | `addons/a2s` |
| VDF / ACF / KeyValues text parsing | `addons/vdf` |
| automatic local Steam scanning | design separately |

## Practical Advice

- Read local VDF files only when the caller explicitly provides a path.
- Do not scan user system directories by default.
- Do not treat `loginusers.vdf` as an authentication source.
- Do not commit real user paths, SteamIDs, tokens, or local Steam configuration into test data.
- Keep strong local Steam business models in application code unless a separate addon is intentionally designed.

## References

- [Valve Developer Community: KeyValues](https://developer.valvesoftware.com/wiki/KeyValues)
