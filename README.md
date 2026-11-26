# goconf

Strict, auditable config loading for Go: decode JSON, YAML, or TOML into one typed struct, with layered defaults, sha256 provenance, and per-occurrence source metadata that declares its own line confidence.

`os.ReadFile` plus `json.Unmarshal` ignore unknown fields, resolve duplicates silently, and keep no record of which file contributed which value. goconf refuses what it cannot spell unambiguously and returns the evidence of what it loaded.

## Install

```sh
go get github.com/adambarber/goconf
```

Go 1.27 or later. The JSON codec uses `encoding/json/v2`.

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

`prov[0]` names the source and carries the sha256 of the exact decoded bytes. `LoadBytes` shares `Load`'s strict decode rules, but `spec.Format` must name a format because in-memory bytes have no extension to infer from. `Load` discovers files in the working directory and the user config directory, with later layers overriding earlier ones field by field. Use `LoadWithMeta` to also get `Meta`: ordered occurrences with indexed and schema paths, key, line, declared line confidence, file, source index, value, and whether each occurrence survived the merge.

## At a glance

- Use `Load[T]` or `LoadBytes[T]` for strict decoding into your struct: unknown fields, duplicate keys, and type mismatches are refused, not silently resolved.
- Use `WithDefaults` for "absent", not "zero": an explicit null is a decision and is never filled. Defaults are deep-copied and matched by canonical field path.
- Use `WithSearchPaths` or `WithGlobalDir` for layered discovery; `WithOptionalConfig` lets a fresh install run before any config file exists.
- Use `WithEnv("APP")` for a schema-bound overlay named `APP_PORT`, `APP_REGION`, and so on. Env values are the outermost layer and win over files and defaults.
- Use `WithMaxBytes(n)` to cap each source at `n` bytes; the default is 16 MiB, the hard maximum 64 MiB, and an out-of-range `n` panics at option construction.
- Use `goconf:"extras"` to capture unknown keys, `goconf:"dynamic"` for caller-chosen keys, and `goconf.Value` with `DecodeValue[T]` for opaque TOML values.
- Use `Redact` for output that leaves the process: fields tagged `secret:"true"` render as `[redacted]`.

## Why

The standard library parses one document and returns the result; it does not say what a typo costs. goconf adds three promises:

- **Fail loud.** An unknown key, duplicate key, ambiguous candidate, or symlink target outside the searched tree is refused by name. Symlink confinement is enforced on the descriptor that reads the bytes.
- **Never echo a value.** Refusal text names keys, paths, formats, and line numbers, but never a config value — including fragments a third-party decoder tried to echo. This is fuzz-pinned across all three formats.
- **Keep the evidence.** Every load returns the digest of every source, and `LoadWithMeta` returns the source line of each occurrence with `LineExact` for JSON/YAML or `LineHeuristic` for TOML.

goconf is not a registry: no reload, no file watching, no secret store, and no decision about what your application should do with a refusal. It loads one config, strictly.

## Secrets

A key whose name looks secret must be declared with a `secret` tag or the load is refused. Each word is stemmed, so plural and camelCase spellings cannot slip past: `password`, `pass`, `passwd`, `pwd`, `secret`, `token`, `credential`, `apikey`, `privatekey`, `api`, and `private`. `api` and `private` match on their own or followed by `key`.

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

`secret:"false"` overrides an inherited mark. Redaction is a rendering decision, not a secret store: the loaded struct still holds the real value.

## Error and refusal contract

- On any refusal, `Load`, `LoadBytes`, and their `WithMeta` siblings return the zero config, nil provenance, and zero `Meta`. There is no partial success.
- Every refusal is a `gorefuse.Because[goconf.Reason]`. Extract it with `errors.As` and switch on `Reason`, or use `errors.Is`. One reason covers one refusal semantics across formats: `ErrUnknownField`, `ErrTypeMismatch`, `ErrDuplicateKey`, `ErrDecoderRefused`, `ErrIO`, `ErrNothingFound`, `ErrSourceTooLarge`, `ErrNeedDeclaration`, and the rest are declared in `refusal.go`.
- Discovery ambiguity is never resolved by guessing: two candidate extensions in one directory are refused unless `Spec.Format` narrows the search.
- An absolute symlink target or a resolved target outside the searched directory is refused at read time, naming the file and where it points.
- `WithDefaults` and `WithMaxBytes` are Must-style: a broken default or an out-of-range cap panics at option construction. Ordinary bad input to `Load` is a refusal, not a panic.
- The refusal paths are fuzzed: arbitrary bytes through the real `Load` in all three formats assert that no config value reaches refusal text.

The full contract — discovery order, merge rules, per-format capabilities, metadata semantics — lives in [docs/load.md](docs/load.md).

## Comparison

- **vs `encoding/json`, `yaml.v3`, `BurntSushi/toml`:** goconf adds strict unknown-field refusal, layering, defaults, provenance, and a shared reason vocabulary on top of those decoders.
- **vs `spf13/viper`:** viper is a registry with flags, remote stores, aliases, and watching. goconf loads one typed config from disk and refuses ambiguity instead of merging everything. Its `WithEnv` is a schema-bound overlay, not a general env-config system.
- **vs `knadh/koanf`:** koanf is a key-value store with pluggable providers and parsers. goconf decodes directly into your struct and keeps per-occurrence source metadata.

## Caveats

- Formats are spelled only by their full extension: `".json"`, `".yaml"`, `".toml"`. `.yml` is deliberately absent.
- An explicit null marks a key present and decodes to zero; defaults never override it. TOML has no null: a key is either present with a value or absent.
- Repeated field names across different array elements are not duplication and decode fine.
- An empty file is refused for JSON and YAML but loads as no settings for TOML.
- TOML carries an exact source digest but heuristic line attribution: BurntSushi/toml v1 does not expose authoritative per-key positions. JSON lines come from token offsets and YAML lines from node lines. `Occurrence.LineConfidence` tells you which you are reading, so callers never have to guess from the format.
- Loading from an `fs.FS` is not in the public API yet. Sources are files discovered on disk or bytes passed to `LoadBytes`.

## Examples

[`examples/`](examples/README.md) has twelve standalone recipes: strict load, formats, layered files, defaults, env overlay, optional absence, source-size cap, metadata, secrets, extras/dynamic keys, opaque `Value`, and refusal handling. Each was run with `go run` and its output recorded; the core scenarios are also published as executable `Example...` functions on [pkg.go.dev](https://pkg.go.dev/github.com/adambarber/goconf).

## Dependencies

- [BurntSushi/toml](https://github.com/BurntSushi/toml)
- [yaml.v3](https://gopkg.in/yaml.v3)
- [gorefuse](https://github.com/adambarber/gorefuse) — the refusal carrier behind the error contract

## Status

Pre-1.0. Unreleased on `main`; the next tag will be `v0.5.0` and the last tag is `v0.4.1`. The pending release adds `LoadBytes`, `Value`, `WithMaxBytes`, secret classification, and `WithEnv`. Documented contracts are pinned by tests.

[![PkgGoDev](https://pkg.go.dev/badge/github.com/adambarber/goconf)](https://pkg.go.dev/github.com/adambarber/goconf) [![codecov](https://codecov.io/gh/adambarber/goconf/branch/main/graph/badge.svg)](https://app.codecov.io/gh/adambarber/goconf)

## License

[MIT](LICENSE)
