# goconf

Strict, auditable config loading for Go: decode JSON, YAML, or TOML into one typed struct, with layered defaults, sha256 provenance, and per-occurrence source metadata.

`os.ReadFile` plus `json.Unmarshal` ignore unknown fields, resolve duplicates silently, and keep no record of which file contributed which value. goconf refuses what it cannot spell unambiguously and returns the evidence of what it loaded.

## Install

```sh
go get github.com/adambarber/goconf
```

Go 1.27 or later is required. The JSON codec uses `encoding/json/v2`.

## Example

```go
package main

import (
	"fmt"

	"github.com/adambarber/goconf"
)

type Config struct {
	Region string `json:"region"`
	Port   int    `json:"port"`
}

func main() {
	cfg, prov, err := goconf.LoadBytes[Config](
		[]byte(`{"region":"eu"}`),
		goconf.Spec{Name: "app", Format: goconf.FormatJSON},
		goconf.WithDefaults(Config{Port: 8080}),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(cfg.Region, cfg.Port, len(prov))
}
// Output:
// eu 8080 1
```

`prov[0]` names the source and carries the sha256 of the exact decoded bytes. `LoadBytes` decodes in-memory bytes under the same strict decode rules as `Load`, but `spec.Format` must name a format because in-memory bytes have no file extension to infer from; `Load` discovers files in the working directory and the user config directory, and later layers override earlier ones field by field. Use `LoadWithMeta` instead of `Load` to also get `Meta`: ordered occurrences with indexed and schema paths, key, line, file, source index, value, and whether each occurrence survived the merge.

## At a glance

- `Load[T]` / `LoadBytes[T]`: strict decoding into your struct — unknown fields, duplicate keys, and type mismatches are refused, not silently resolved.
- `WithDefaults`: defaults mean "absent", not "zero". An explicit null is a decision and is never filled; the stored default is deep-copied so later caller mutations cannot change a load.
- `WithSearchPaths` / `WithGlobalDir`: layered discovery — global config first, local config next, later layers overriding earlier ones with nested-object deep merge.
- `WithOptionalConfig`: forgive true absence so a fresh install can run before any config file exists. An IO fault or a present-but-invalid file is still refused, and one warning line says when nothing was loaded.
- `WithEnv("APP")`: overlay process variables named `APP_PORT`, `APP_REGION`, and so on. The path comes from the schema, values parse to the field's shape, and env values are the outermost layer: they beat file values and defaults. Unknown variables are ignored. With `APP_PORT=9090` the example above prints `eu 9090`.
- `WithMaxBytes(n)`: cap each source at `n` bytes, default 16 MiB and hard maximum 64 MiB. An out-of-range `n` is a programmer error and panics; an oversized source is refused with `ErrSourceTooLarge` before decode.
- `Provenance`: one entry per contributing source — path plus sha256 of the bytes decoded.
- `LoadWithMeta` / `Meta`: answer "which line set this value, and did it survive the merge" across JSON, YAML, and TOML, including dynamic keys, extras, and repeated array elements.
- `goconf:"extras"` (or `goconf:"rest"`) captures unknown keys into a map; `goconf:"dynamic"` accepts arbitrary user-chosen keys in a map instead of fixed schema fields.
- `goconf.Value` holds an opaque TOML value for later `DecodeValue[T]`; it is refused in JSON and YAML sources.
- `Redact` renders a loaded config the way `fmt.Sprintf("%+v", cfg)` would, except that fields tagged `secret:"true"` render as `[redacted]`.

## Why

The standard library parses one document and returns the result; it does not say what a typo costs. goconf is built around three extra promises:

- **Fail loud.** An unknown key, a duplicate key, an ambiguous candidate, or a symlink whose target leaves the searched directory is refused by name.
- **Never echo a value.** Refusal text names keys, paths, formats, and line numbers, but never a config value — including fragments a third-party decoder tried to echo. This is fuzz-pinned across all three formats.
- **Keep the evidence.** Every load returns the digest of every source, and `LoadWithMeta` returns the source line of each occurrence.

goconf is not a registry: no reload or file watching, no secret store, and no decision about what your application should do with a refusal. It loads one config, strictly.

## Secrets

A key whose name looks secret must be declared with a `secret` tag on the schema field, or the load is refused. Each word is stemmed before matching, so plural and camelCase spellings cannot slip past: `password`, `pass`, `passwd`, `pwd`, `secret`, `token`, `credential`, `apikey`, `privatekey`, `api`, and `private`. `api` and `private` match on their own or followed by `key`. The refusal names the path, never the value:

```go
type Config struct {
	Password string `json:"password"`
}

_, _, err := goconf.LoadBytes[Config](
	[]byte(`{"password":"hunter2"}`),
	goconf.Spec{Name: "app", Format: goconf.FormatJSON},
)
fmt.Println(err)
// Output:
// Error: app:1: undeclared secret: "password". add a secret tag to the field that holds it.
```

Tag the field to accept it, then use `Redact` for any output that leaves the process:

```go
type Config struct {
	Token string `json:"token" secret:"true"`
}

cfg, _, err := goconf.LoadBytes[Config](
	[]byte(`{"token":"hunter2"}`),
	goconf.Spec{Name: "app", Format: goconf.FormatJSON},
)
if err != nil {
	panic(err)
}
fmt.Println(goconf.Redact(cfg))
// Output:
// {Token:[redacted]}
```

`secret:"false"` overrides an inherited mark. Redaction is a rendering decision, not a secret store: it hides marked fields in `Redact` output, but the loaded struct still holds the real value.

## Error and refusal contract

- On any refusal, `Load`, `LoadBytes`, and their `WithMeta` siblings return the zero config, nil provenance, and zero `Meta`. There is no partial success.
- Every refusal is a `gorefuse.Because[goconf.Reason]`. Extract it with `errors.As` and switch on `Reason`, or use `errors.Is`. One reason covers one refusal semantics across formats: `ErrUnknownField`, `ErrTypeMismatch`, `ErrDuplicateKey`, `ErrDecoderRefused`, `ErrIO`, `ErrNothingFound`, `ErrSourceTooLarge`, `ErrNeedDeclaration`, and the rest are declared in `refusal.go`.

```go
var because gorefuse.Because[goconf.Reason]
if errors.As(err, &because) && because.Reason == goconf.ErrNeedDeclaration {
	// because.What.File names the file, because.What.Line the line
}
```

- Discovery ambiguity is never resolved by guessing: two candidate extensions in one directory are refused unless `Spec.Format` narrows the search.
- A candidate whose symlink-resolved target leaves the searched directory is refused at read time, naming the file and its target.
- `WithDefaults` and `WithMaxBytes` are Must-style: a broken default or an out-of-range cap panics at option construction. Ordinary bad input to `Load` is a refusal, not a panic.
- The refusal paths are fuzzed: arbitrary bytes through the real `Load` in all three formats assert that no config value reaches refusal text.

The full contract — discovery order, merge rules, per-format capabilities, metadata semantics — lives in [docs/load.md](docs/load.md).

## Comparison

- **vs `encoding/json`, `yaml.v3`, `BurntSushi/toml`:** goconf adds strict unknown-field refusal, layering, defaults, provenance, and a shared reason vocabulary on top of those decoders.
- **vs `spf13/viper`:** viper is a registry with env vars, flags, remote stores, and watching. goconf loads one typed config from disk and refuses ambiguity instead of merging everything. Its `WithEnv` is schema-bound overlay, not a general env-config system.
- **vs `knadh/koanf`:** koanf is a key-value store with pluggable providers and parsers. goconf decodes directly into your struct and keeps per-occurrence source metadata.

## Caveats

- Formats are spelled only by their full extension: `".json"`, `".yaml"`, `".toml"`. `.yml` is deliberately absent, so which decoder runs is never guessed.
- An explicit null marks a key present and decodes to the zero value; defaults never override it. TOML has no null: a key is either present with a value or absent.
- Repeated field names across different array elements are not duplication and decode fine.
- An empty file is refused for JSON and YAML but loads as no settings for TOML; goconf does not paper over the formats' own null semantics.
- TOML carries an exact source digest but best-effort line attribution: BurntSushi/toml v1 does not expose authoritative per-key positions. JSON lines come from token offsets and YAML lines from node lines.
- Loading from an `fs.FS` is not in the public API yet. Sources are files discovered on disk or bytes passed to `LoadBytes`.

## Dependencies

- [BurntSushi/toml](https://github.com/BurntSushi/toml)
- [yaml.v3](https://gopkg.in/yaml.v3)
- [gorefuse](https://github.com/adambarber/gorefuse) — the refusal carrier behind the error contract

## Status

Pre-1.0. The current release is `v0.5.0`; it includes `LoadBytes`, `Value`, `WithMaxBytes`, secret classification, and `WithEnv`. Documented contracts are pinned by tests.

[![PkgGoDev](https://pkg.go.dev/badge/github.com/adambarber/goconf)](https://pkg.go.dev/github.com/adambarber/goconf)

## License

[MIT](LICENSE)
