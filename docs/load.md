# Load

`Load[T](spec Spec, opts ...Option) (T, []Provenance, error)` discovers, merges, and decodes one config into `T`, strictly.

## The contract

The caller declares a config NAME and a required FORMAT. The format is
spelled only by its full extension — `FormatJSON`, `FormatYAML`,
`FormatTOML`; `.yml` is deliberately absent. An invented `Format` value is
refused at load, naming the accepted spellings.

## Discovery

Default search paths, in order:

1. `os.UserConfigDir()/name/` — the global layer
2. `./` — the local file of the same name in the working directory

Later sources override earlier ones field by field; nested objects deep
merge. `WithSearchPaths(...)` replaces the list; `WithGlobalDir(...)`
replaces only the global layer.

## Symlink confinement

A candidate file whose symlink-resolved target leaves the directory that
was searched is refused at read time: the refusal names the file and where
it points. A symlink whose target stays inside the searched directory is a
legitimate candidate and loads. Confinement lives in the read path
(`readSources`), the one chokepoint every candidate's bytes pass through,
so the window between check and read is minimal.

## Refusals

Unknown fields are refused by every decode, in every format. If a
directory holds candidates with different-format extensions and the
format does not narrow the search, the load is refused naming the
directory and both files. Ambiguity is never resolved by guessing;
declaring the format is the deterministic way out.

## Defaults

`WithDefaults(value)` fills only the fields absent from every source.
A field present with an explicit zero value is kept; an explicit null is
a decision and is never overridden. A default whose value cannot marshal
fails loud at option construction.

## Per-format capabilities

JSON: explicit null exists; a null marks the key present and decodes to
the zero value.

YAML: explicit null exists, same semantics as JSON.

TOML: explicit null does not exist. A TOML key is either present with a
value or absent; there is no way to write a decision to clear a field.

## Provenance

Every successful load returns one `Provenance` per contributing source,
in discovery order: the file path and the hex sha256 digest of the bytes
that were decoded.
