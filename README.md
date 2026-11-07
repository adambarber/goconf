# goconf

Strict, auditable config loading for Go.

```go
type Config struct {
	Region string `json:"region" yaml:"region" toml:"region"`
	Port   int    `json:"port" yaml:"port" toml:"port"`
}

cfg, prov, err := goconf.Load[Config](goconf.Spec{
	Name:   "myapp",
	Format: goconf.FormatJSON,
}, goconf.WithDefaults(Config{Region: "us-east", Port: 8080}))
```

`Load[T]` discovers, merges, and decodes one config into `T`, strictly:

- **Three formats, one API** — JSON, YAML, TOML, each decoded strictly:
  unknown fields, duplicate keys, and type mismatches are refused, naming
  the offending key and line. A format is spelled only by its full
  extension (`".json"`, `".yaml"`, `".toml"`); `.yml` is deliberately absent.
- **Layered discovery** — a global config (`os.UserConfigDir()/name/`)
  and a local one (the working directory); later layers override earlier
  ones field by field, nested objects deep merge. Legs may mix formats.
- **Defaults for absent fields only** — explicit null is a decision, not
  an absence: a null field is never overwritten by a default.
- **Provenance** — one entry per contributing source: the path, its search
  leg, and the sha256 of the exact bytes decoded.
- **Refusals that name things** — ambiguity (two formats in one dir),
  unknown candidate extensions, path-escaping names, symlink escapes:
  every refusal says what was refused and why, so misconfiguration fails
  loud at startup instead of haunting production.

Symlinks are confined: a config file that resolves outside its searched
directory is refused, in both discovery legs.

Zero dependencies beyond the two format decoders
([BurntSushi/toml](https://github.com/BurntSushi/toml),
[yaml.v3](https://gopkg.in/yaml.v3)); no reflection-based magic beyond
decoding into `T`.

## Usage

See [docs/load.md](docs/load.md) for the full contract: discovery order,
merge semantics, defaults, provenance, symlink confinement, and the
refusal shapes.

## Status

API is stable modulo the pre-1.0 conventions of an early library.
