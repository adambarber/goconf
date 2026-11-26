# Recipes

Each directory is a standalone `package main`. Copy `main.go` into your
project and replace the config shape, file names, or prefix. Every recipe is
compiled by `go test ./...` and was run with `go run ./examples/<name>`.

| Recipe | Shows | Output |
| --- | --- | --- |
| [`strict_load`](strict_load/main.go) | `LoadBytes`, strict decode, provenance | `eu 8080 1` |
| [`formats`](formats/main.go) | one schema decoded from YAML and TOML | `eu 8080` twice |
| [`layered_files`](layered_files/main.go) | global then local files, field-wise override | `local 1` |
| [`defaults`](defaults/main.go) | default fills absent, explicit null stays zero | `eu 8080` / `eu 0` |
| [`env_overlay`](env_overlay/main.go) | schema-bound `WithEnv` overlay | `eu 9090` |
| [`optional_absence`](optional_absence/main.go) | `WithOptionalConfig` plus defaults | warning line, then `eu 8080` |
| [`max_bytes`](max_bytes/main.go) | source-size refusal | `true` |
| [`metadata`](metadata/main.go) | `LoadWithMeta` occurrence path, line, survival | `port 3 true` |
| [`secrets`](secrets/main.go) | secret declaration and `Redact` | `{Token:[redacted]}` |
| [`extras_dynamic`](extras_dynamic/main.go) | `extras` and `dynamic` maps | `app team eu` |
| [`value`](value/main.go) | opaque TOML `Value` and `DecodeValue` | `42` |
| [`refusal_handling`](refusal_handling/main.go) | typed reason and frame from a refusal | `true true app 1` |

Run one locally:

```sh
go run ./examples/strict_load
```

The core scenarios are also published as executable `Example...` functions on
[pkg.go.dev](https://pkg.go.dev/github.com/adambarber/goconf).
