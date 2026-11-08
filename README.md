# goconf

goconf loads configuration files strictly: unknown fields, duplicate keys,
and ambiguous discovery are refused, not silently resolved.

```sh
go get github.com/adambarber/goconf
```

Refusals are typed: every error is a `gorefuse.Because[goconf.Reason]`
you can `errors.As` out and switch on — `ErrDuplicateKey`,
`ErrUnknownField`, `ErrNothingFound`, and friends. See [Error
identity](#error-identity) below.

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
  fragments of values echoed by third-party decoders, in every format,
  on every error path.
- Discovery ambiguity is never resolved by guessing: if a directory holds
  candidates with different-format extensions and the format does not
  narrow the search, the load is refused naming both files; declaring the
  format is the deterministic way out.
- Symlinks are confined: a candidate whose symlink-resolved target leaves
  the searched directory is refused at read time, naming the file and
  where it points. This holds in both discovery legs.
- A broken default fails loud at option construction (`WithDefaults`
  panics, a Must-style contract like `regexp.MustCompile`), never
  silently at load.

### Error identity

Every refusal is a `gorefuse.Because[goconf.Reason]`: extract it with
`errors.As`, then switch on the reason or match it with `errors.Is`.

```go
var because gorefuse.Because[goconf.Reason]
if errors.As(err, &because) {
    switch because.Reason {
    case goconf.ErrDuplicateKey:
        // because.What.File names the file, because.What.Line the line
    case goconf.ErrUnknownField:
        // ...
    }
}
```

One reason per refusal semantics, across formats: `ErrUnknownField`
covers JSON's, TOML's, and YAML's unknown-field refusals;
`ErrTypeMismatch` covers a value that does not fit the field's declared
type, in every format; `ErrDecoderRefused` covers syntax and scanner
faults; `ErrIO` covers discovery and read faults — a directory that
could not be listed, a candidate that could not be stat-ed, a file that
could not be read, a symlink that could not be resolved. The full
constant list is in `refusal.go`. Refusals that happened before any file
exists (nothing found, ambiguity, an invented format, a bad option)
carry no file frame; refusals of a specific file carry the `file:line`
frame pointing at that file.

The full contract — discovery order, merge semantics, per-format
capabilities, refusal shapes — is in
[docs/load.md](docs/load.md).

## Dependencies

- [BurntSushi/toml](https://github.com/BurntSushi/toml)
- [yaml.v3](https://gopkg.in/yaml.v3)
- [gorefuse](https://github.com/adambarber/gorefuse)

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
