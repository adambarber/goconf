# goconf

goconf loads configuration files strictly: unknown fields, duplicate keys,
and ambiguous discovery are refused, not silently resolved.

```sh
go get github.com/adambarber/goconf
```

Given `app.json` in the working directory:

```json
{"region": "eu"}
```

```go
type Config struct {
	Region string `json:"region"`
	Port   int    `json:"port"`
}

cfg, prov, err := goconf.Load[Config](
	goconf.Spec{Name: "app", Format: goconf.FormatJSON},
	goconf.WithDefaults(Config{Port: 8080}),
)
// cfg.Region == "eu", cfg.Port == 8080 (filled from defaults)
// prov[0].Path names the file; prov[0].Digest is its sha256.
// err is nil.
```

With `{"region": "eu", "tls": true}`, the load is refused: `tls` is not a
field of `Config`, and the error names the offending key.

## At a glance

- Use goconf if you want **strict decoding** into your struct —
  `Load[T]` refuses unknown fields, duplicate keys, and type mismatches,
  naming the offending key and line.
- Use it if you want **layered discovery** — a global layer under
  `os.UserConfigDir()/name/` and a local file in the working directory,
  later layers overriding earlier ones field by field, nested objects
  deep merging (`WithSearchPaths`, `WithGlobalDir` to override).
- Use it if you want **defaults that mean absent, not zero** —
  `WithDefaults` fills only fields absent from every source; an explicit
  null is never overridden by a default.
- Use it if you want **provenance** — every load returns one `Provenance`
  per contributing source: the path and the sha256 of the bytes decoded.
- Use it if you want **fail-loud discovery** — two format candidates in
  one directory, an unknown extension, a path-escaping name, or a symlink
  leaving the searched directory is refused by name, never guessed.

## Why

`os.ReadFile` + `json.Unmarshal` gives you none of this: unknown fields
slip through, duplicate keys are resolved silently by whichever decoder
wins, one file only, no layering, no record of which file contributed
which bytes, and a zero value in the config is indistinguishable from a
missing one — so defaults may or may not win, depending on the file's
whitespace. goconf refuses what it cannot spell unambiguously, layers
sources deterministically, and returns the evidence of what it loaded.

What goconf does not do: no environment-variable overlay, no
reload/watch (define your own loop around `Load[T]`), no secrets
management. It loads one config, strictly.

## Error and refusal contract

- `Load` never partially succeeds: on any refusal the returned config is
  the zero value and provenance is nil.
- Every refusal names what was refused — the key, the file, the line, or
  the directory — and why. Refusal text names keys, paths, types, and
  line numbers. Config values never appear in refusal text, including
  fragments of values echoed by third-party decoders.
- Discovery ambiguity is never resolved by guessing: if a directory holds
  candidates with different-format extensions and the format does not
  narrow the search, the load is refused naming both files; declaring the
  format is the deterministic way out.
- Symlinks are confined: a candidate whose symlink-resolved target leaves
  the searched directory is refused at read time, naming the file and
  where it points. This holds in both discovery legs.
- A broken default fails loud at option construction
  (`WithDefaults`), never silently at load.
- Errors are wrapped `fmt.Errorf`/`%w` today; sentinel-error identity via
  `errors.Is` is not yet part of the public API.

The full contract — discovery order, merge semantics, per-format
capabilities, refusal shapes — is in
[docs/load.md](docs/load.md).

## Dependencies

Two, both format decoders:
[BurntSushi/toml](https://github.com/BurntSushi/toml) and
[yaml.v3](https://gopkg.in/yaml.v3). JSON uses the standard library.
Nothing else.

## Caveats

- Formats are spelled only by their full extension — `FormatJSON` is
  `".json"`. `.yml` is deliberately absent: one spelling per format,
  no ambiguity about which decoder runs.
- An explicit null marks a key present and decodes to the zero value;
  it is a decision, and defaults never override it. (TOML has no null:
  a TOML key is either present with a value or absent.)
- Repeated field names across different array elements are not
  duplication and decode fine.

## Status

API is stable modulo the pre-1.0 conventions of an early library.

[![PkgGoDev](https://pkg.go.dev/badge/github.com/adambarber/goconf)](https://pkg.go.dev/github.com/adambarber/goconf)
[![Go Report Card](https://goreportcard.com/badge/github.com/adambarber/goconf)](https://goreportcard.com/report/github.com/adambarber/goconf)
