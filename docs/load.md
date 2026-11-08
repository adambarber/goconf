# Load

`Load[T](spec Spec, opts ...Option) (T, []Provenance, error)` discovers, merges, and decodes one config into `T`, strictly.

## The contract

The caller declares a config NAME and optionally a FORMAT. A format is
spelled only by its full extension — `FormatJSON`, `FormatYAML`,
`FormatTOML`; `.yml` is deliberately absent. An invented `Format` value is
refused at load, naming the accepted spellings.

Leaving the format undetermined (the zero value) is documented behavior:
discovery then demands that each directory hold exactly one candidate
extension, and each discovered file is decoded through its own format, so
search legs may legitimately mix formats (a global `.json` and a local
`.toml` load together, later legs overriding earlier ones).

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

## Error identity

A refusal is typed, not a wrapped sentence: every error `Load` returns
is a `gorefuse.Because[goconf.Reason]` — extract it with `errors.As`,
match it with `errors.Is` against the reason constants, or switch on
`because.Reason`. The constants (`ErrDuplicateKey`, `ErrUnknownField`,
`ErrAmbiguous`, `ErrNothingFound`, …) are one vocabulary across
formats: JSON's, TOML's, and YAML's unknown-field refusals share
`ErrUnknownField`; a YAML multi-document stream is `ErrMultiDoc` in
decode and in the presence walk alike.

The rendered text reads as one sentence: the `file:line` frame (when a
file was refused), the reason, what was refused, the decoder's detail
(when there is one), and the fix. A refusal raised before any file
exists — nothing found, ambiguous candidates, an invented format, an
escaping config name, a bad default — carries no frame and never
fabricates one.

## Refusals

Unknown fields are refused by every decode, in every format. Duplicate
keys are refused by every decode, in every format: a key re-declared in
the same object — a whole object re-opened, or one re-declaring part of a
struct that would otherwise silently merge — is refused naming the key
and, in JSON, the line of its first definition. Repeated field names
across different array elements are not duplication and decode. If a
directory holds candidates with different-format extensions and the
format does not narrow the search, the load is refused naming the
directory and both files. Ambiguity is never resolved by guessing;
declaring the format is the deterministic way out. A candidate whose
extension no format spells (e.g. `app.bak`) is refused by name. Refusal
text names keys, paths, types, and line numbers. Config values never
appear in refusal text, including fragments of values echoed by
third-party decoders — in every format, on every error path.
Unknown-field refusals name the offending key; JSON unknown-field
refusals also name the key's line.

## Defaults

`WithDefaults(value)` fills only the fields absent from every source.
A field present with an explicit zero value is kept; an explicit null is
a decision and is never overridden. A default whose value cannot marshal
fails loud at option construction: `WithDefaults` panics, a Must-style
contract like `regexp.MustCompile`.

The presence-path vocabulary is one across formats: a field's config key
is the first non-empty of its `json`, `yaml`, and `toml` tags, falling
back to the Go field name. Defaults fill exactly the fields no source —
any format — recorded as present.

## Multi-document streams

A YAML stream holding more than one document is refused, naming the line
of the separator: first-wins would silently resolve what must be
refused. JSON and TOML cannot express multi-document streams. YAML
anchors and aliases resolve in the presence walk, so an alias-resolved
value is never overridden by a default.

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
