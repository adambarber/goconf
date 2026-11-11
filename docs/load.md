# Load

`Load[T](spec Spec, opts ...Option) (T, []Provenance, error)` discovers, merges, and decodes one config into `T`, strictly.

`LoadWithMeta[T](spec Spec, opts ...Option) (T, []Provenance, Meta, error)` is the metadata-returning sibling: it preserves `Load`'s behavior and contracts and additionally returns per-occurrence key/value/location metadata. `Load` is equivalent to discarding `Meta`.

## Per-occurrence metadata (`LoadWithMeta`)

`Meta` is an ordered, queryable collection of `Occurrence` records. Records are emitted in discovery/load order by source, and in document order within each source, so every present key occurrence is observable (including repeated array/sequence elements and TOML array-of-tables entries).

Each `Occurrence` carries:

- `Path` (`OccurrencePath`): the indexed occurrence path identifying this specific occurrence (for example `servers[0].host`, `resource[2].name`, `env.FOO`). Indexed paths distinguish repeated elements/occurrences.
- `SchemaPath` (`SchemaPath`): the unindexed canonical schema path in the same vocabulary as presence and unknown-field handling (for example `servers.host`, `resource.name`). Array indices are not part of schema identity.
- `Key`: the leaf key or element name at this path segment (array elements use `[i]`).
- `Line`: the 1-based source line for the key/element occurrence.
- `File`: the source file path.
- `SourceIndex`: the index of the contributing source in load/provenance order.
- `Value`: the decoded value for this occurrence. For occurrences that contribute to the effective result, `Value` is enriched from the final loaded configuration after decode, layering/merge, and defaults when the path resolves in the effective config; for superseded occurrences, `Value` is the source-captured decoded value (and may remain source-derived when no effective value resolves at that indexed path).
- `Effective`: whether this occurrence contributes to the effective loaded configuration.

Query helpers: `Entries()`, `Len()`, `ByPath`/`Lookup` (indexed path), `BySchemaPath`/`LookupSchema` (schema path), `At`/`AtSchema` (zero-based occurrence position within matches), and `Effective()`.

Semantics and resolution:

- **Uniformity:** metadata is collected through goconf's format-specific decode/canonicalization paths and normalized to one occurrence model across TOML, JSON, and YAML. Line extraction is parser-derived for JSON (token offsets) and YAML (node lines); for TOML, BurntSushi/toml v1 metadata does not expose per-key source positions, so key/occurrence line attribution uses a hardened line-anchored, comment-aware heuristic (best-effort) while preserving uniform path/identity/ordering semantics.
- **Indexed paths, arrays, and array-of-tables:** array/sequence elements are recorded as indexed element occurrences (`items[i]`) with `Key="[i]"`, and their children under indexed paths (`items[i].name`). Container occurrences are recorded for array fields/paths where applicable, and ordering is consistent (container, then element, then element children, in document order). This includes JSON/YAML sequences, TOML array-of-tables `[[t]]` (element occurrences `t[i]`, including nested AOT indexed per parent element, with AOT container and element metadata normalized to the same array contract), and TOML inline arrays (scalar elements and inline table-array elements/containers/children) with best-effort line attribution per the TOML parser model.
- **Dynamic and extras:** keys inside `goconf:"dynamic"` maps and keys captured by `goconf:"extras"`/`goconf:"rest"` are representable as occurrences; unknown keys outside absorbing sites remain refused as before.
- **YAML merge keys (`<<`):** merged fields are emitted at their effective schema/occurrence paths (not under `<<`) so merged effective values are queryable at the paths they occupy in the loaded config. Explicit keys override merged keys; when `<<` merges a sequence of mappings, precedence follows YAML semantics (earlier merged mappings take precedence among merged sources, subject to explicit-key override). The literal `<<` key is not emitted as a separate occurrence.
- **Layering/merge and defaults:** all occurrences from all loaded sources are recorded. Effective contribution is determined per indexed occurrence path in load order: the last occurrence for a path that resolves in the final effective config is marked effective; earlier/superseded occurrences (including elements replaced by a later array value) are non-effective. Nested object merge and later-source override follow normal decode semantics. Defaults fill absent fields but do not synthesize source occurrences; effective values on contributing occurrences reflect the final post-merge/default config when resolvable.
- **Refusals and value-freedom:** metadata is structured API output, not refusal text. Refusal behavior, unknown-field handling, duplicate-key refusal, type-mismatch refusal, multi-document refusal, and refusal value-freedom guarantees are unchanged. On refusal, `LoadWithMeta` returns the zero config, nil provenance, zero `Meta`, and the refusal error (same as `Load`).

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

## Optional config

`WithOptionalConfig()` forgives true absence: when every search

directory is readable and zero candidates exist, `Load` behaves as if
zero keys were present — defaults fill, zero values remain, provenance
is empty, and one warning is printed through the standard `log`
package:

    [warn] No config file found. Using default values.

That line is the only signal: it fires exactly once per Load call, and
only when the caller opted in AND no file was found. A config file that
was found loads normally with no warning.

The boundary is fault versus absence:

- True absence — every search directory readable, zero candidates — is
  the only case the option forgives. With `WithSearchPaths`, absence in
  the overridden paths is still absence.
- IO faults still refuse with `ErrIO` even with the option set: an
  unlistable directory, an unreadable file, a permission error. A fault
  is not absence; a load that could not look did not look.
- Present-but-invalid config still refuses even with the option set:
  unknown fields, type mismatches, duplicate keys, syntax errors — all
  unchanged.
- Without the option, absence keeps the `ErrNothingFound` refusal.

With the option and no defaults, a zero-value struct loads fine: the
caller opted in.

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
`ErrTypeMismatch`, `ErrDecoderRefused`, `ErrIO`, `ErrAmbiguous`,
`ErrNothingFound`, …) are one vocabulary across formats: JSON's, TOML's,
and YAML's unknown-field refusals share `ErrUnknownField`; a value that
does not fit a field's declared type is `ErrTypeMismatch` in every
format; syntax and scanner faults are `ErrDecoderRefused`; discovery and
read faults are `ErrIO`; a YAML multi-document stream is `ErrMultiDoc`
in decode and in the presence walk alike.

The rendered text reads as one sentence: the `file:line` frame (when a
file was refused), the reason, what was refused, and the fix. The text
speaks to the config file's author: the file's objects, the value's
shape, one labeling of the fault, and one fix. A type-mismatch refusal
names the key and the user's value-shape ("the value of "port" is text,
not a whole number"), synthesized from the target field's declared kind —
never a Go type name. A refusal raised before
any file exists — nothing found, ambiguous candidates, an invented
format, an escaping config name, a bad default — carries no frame and
never fabricates one.

## Refusals

Unknown fields are refused by every decode, in every format, unless the
containing struct declares an extras-capture field (`goconf:"extras"` or
`goconf:"rest"`). When extras are declared, keys that do not match a
fixed field of that struct are captured into the extras map instead of
being refused; fixed fields are never captured as extras. Unknown fields
in structs without extras are still refused, and duplicate keys, type
mismatches, and root-shape errors are still refused.

Duplicate keys are refused by every decode, in every format: a key
re-declared in the same object — a whole object re-opened, or one
re-declaring part of a struct that would otherwise silently merge — is
refused naming the key and, in JSON, the line of its first definition.
Repeated field names across different array elements are not duplication
and decode. If a directory holds candidates with different-format
extensions and the format does not narrow the search, the load is refused
naming the directory and both files. Ambiguity is never resolved by
guessing; declaring the format is the deterministic way out. A candidate
whose extension no format spells (e.g. `app.bak`) is refused by name.
Refusal text names keys, paths, types, and line numbers. Config values
never appear in refusal text, including fragments of values echoed by
third-party decoders — in every format, on every error path.
Unknown-field refusals name the offending key; JSON and YAML
unknown-field refusals also name the key's line.

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

## Extras and dynamic keys

Two struct tags extend the schema beyond fixed, known fields:

- `goconf:"dynamic"` on a map field declares a dynamic-key map. Any key
  present at that map position is accepted as a map entry and decoded
  into the map, rather than treated as a fixed schema field name. This is
  the way to load named instances, environment-specific blocks, or other
  user-chosen keys.

  ```go
  type Config struct {
  	Region string            `json:"region" toml:"region" yaml:"region"`
  	Env    map[string]string `goconf:"dynamic" json:"env" toml:"env" yaml:"env"`
  }

  // {"region":"eu","env":{"FOO":"1","BAR":"2"}}
  // region == "eu", env["FOO"] == "1", env["BAR"] == "2"
  ```

- `goconf:"extras"` (or `goconf:"rest"`) on a map field captures unknown
  keys from the containing struct's object. Keys that do not match a
  fixed field are stored in the extras map instead of being refused;
  fixed fields are decoded normally and are never captured as extras. The
  extras field itself is not a config key (it is typically tagged
  `json:"-" toml:"-" yaml:"-"`).

  ```go
  type Resource struct {
  	Name  string            `json:"name" toml:"name" yaml:"name"`
  	Kind  string            `json:"kind" toml:"kind" yaml:"kind"`
  	Extra map[string]string `goconf:"extras" json:"-" toml:"-" yaml:"-"`
  }

  // {"name":"r","kind":"k","provider":"aws"}
  // Name == "r", Kind == "k", Extra["provider"] == "aws"
  ```

Behavior (all formats): extras and dynamic keys are supported in JSON,
YAML, and TOML. Captured extra values must be compatible with the extras
map's element type; a value that cannot be assigned or converted to that
type is refused as a type mismatch. Unknown fields inside a struct that
does not declare extras remain refused. Duplicate keys, type mismatches,
and non-mapping root documents remain refused according to the usual
rules.

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
An empty TOML file loads as no settings — JSON's and YAML's decoders
refuse an empty file naming the fix, an asymmetry of the underlying
decoders.

## Fuzzing

The refusal paths are fuzzed: `FuzzLoad` (fuzz_test.go) drives arbitrary
bytes through the real `Load` path in all three formats, seeded with the
golden fixture shapes and the hostile ones — duplicate keys per format,
secret-shaped values, backquoted and multi-line strings, multi-document
streams, deep nesting, invalid UTF-8, null bytes, empty files, and huge
scalars. Its oracle is differential: the input and a twin where the
planted marker is replaced by a neutral token must render the same
refusal after both sides' markers are normalized to one placeholder.
Structural references survive — a key names itself, by the contract
above — so key-position echoes normalize; the oracle detects an echo
whose spelling the substitution does NOT account for (a leak that
transforms differently from the twin's). It is a tripwire, not the
primary leak guard: the deterministic table in
TestNoValueFragmentSurvivesAnyRefusal and the corpus pins in
testdata/fuzz/FuzzLoad/ assert directly that no value fragment renders,
and those are the guarantee. A decoder may name a character class by
one character (JSON's "invalid character
's'"), which is not a value echo: one byte reconstructs nothing; such
spans are dropped before comparison. Any failing input lands in
testdata/fuzz/FuzzLoad/ and replays in plain `go test` as a permanent
pinned regression.

Two of the yaml terror-strip helpers (codec_yaml_terror.go) are
unreachable through the public API by construction: every
`*yaml.TypeError` is classified into the shared vocabulary before the
strip pass runs, and the empty-issue TypeError that would classify as
`ErrDecoderRefused` is one yaml.v3 never emits. They stay as the strip
pass's own unit surface.

## Provenance

Every successful load returns one `Provenance` per contributing source,
in discovery order: the file path and the hex sha256 digest of the bytes
that were decoded.
