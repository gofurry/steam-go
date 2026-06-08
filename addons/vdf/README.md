# steam-go/addons/vdf

`addons/vdf` bridges [`github.com/gofurry/vdf-go`](https://github.com/gofurry/vdf-go)
into the steam-go addon tree.

Use it when your steam-go-based tool also needs to parse local Valve Data Format
(VDF / KeyValues) text files such as:

- `libraryfolders.vdf`
- `appmanifest_*.acf`
- `config.vdf`
- `loginusers.vdf`

## Scope

This addon is a thin wrapper. It does not re-implement the parser.

It supports text VDF / KeyValues parsing, querying, small mutable edits, and
marshaling through the upstream `vdf-go` API.

It does not support:

- binary VDF
- `shortcuts.vdf`
- automatic Steam installation scanning
- account extraction
- reading local user directories automatically

## Example

```go
doc, err := vdf.ParseFile("steamapps/appmanifest_730.acf")
if err != nil {
	panic(err)
}

appid := doc.Path("AppState", "appid").AsString()
```
