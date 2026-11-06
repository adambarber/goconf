# goconf Architecture Plan

## Purpose

goconf is a shared Go library for **config validation at load**, extracted from the shared behavior found across control-point, hardstand, and birdcage. It owns one question: "given a hand-authored config file, is it valid, and what are its typed values?" It answers by refusing invalid input at load time and returning a typed, refusal-vocabulary result.

## Architecture

goconf is a **leaf package**: it imports nothing from the three consuming projects, and nothing from them imports goconf's internals. It will export exactly the chokepoints:

| Chokepoint | What it owns | Derived from |
|---|---|---|
| Config loading | strict decode, unknown-field refusal, defaults only for absent fields, digest stamping | control-point `internal/policy/`, hardstand `config/`, birdcage `bpl/` |
| Refusal vocabulary | `Refusal{File, Line, Subject, Why}`, `At()` constructor, `Describable` interface, generic `Because[R]` carrier | hardstand `refusal/` |
| Boundary types | named string kinds for account, model, address; no bare `string` at domain positions | control-point `internal/domain/` |

The loader chokepoint ships as `Load[T]` (not `LoadBytes`), with file discovery, per-file format handling, layered defaults (`WithDefaults`), provenance with digest stamping, and symlink confinement. What remains of the chokepoints: wiring the loader's refusals into the gorefuse refusal vocabulary, and the boundary newtypes.

## Feature Set (Phase 1 — MVP)

1. **Strict config loader** — *shipped as `Load[T]`*: strict JSON/YAML/TOML decode with unknown-field refusal, defaults for absent fields only (`WithDefaults`), explicit-vs-absent via presence KeySets (not pointer sentinels), digest stamped per source in `Provenance`.
2. **Typed refusal carrier** (`Because[R]`): generic, compile-time-incompatible reason types per package; `Frame`, `Reason`, `Subject`, `Tail`, `Cause`; `Error()` renders prose for people; `Describe()` renders `Refusal` for machines. — *not yet implemented*
3. **Refusal verdict** (`Refusal{File, Line, Subject, Why}`): `At(file, line, subject, why)` single constructor; `Describable` interface. — *not yet implemented*
4. **Domain newtypes** (`Account`, `Model`, `Addr`): named string types with `AddrOf`/`SplitProvider` as the only join/split points. — *not yet implemented*

## Scaffolding Already Landed

| File | Contents |
|---|---|
| `go.mod` | `module github.com/adambarber/goconf`, `go 1.27` |
| `policy/policy.bpl` | Birdcage policy with `no_clones` rule (`clone(lines: 3, tokens: 20)`, `lang: any`), canonical alphabetical field order |
| `.golangci.yml` | Strictest merged lint gate from all three repos (all linters enabled in any repo, strictest settings where they differ; `interfacebloat max: 5`, `funlen 60/40`, `gocognit 10`, `cyclop 10`, gofumpt with `group-params`+`clothe-returns`) |

## Migration Path for Each Repo

- **hardstand**: lift `refusal/carrier.go` into goconf as the `Because[R]` carrier; hardstand imports goconf for the refusal vocabulary; hardstand's own `Refusal` type either becomes a thin alias or is removed.
- **control-point**: `internal/policy/` loader imports goconf for strict-decode defaults and refusal rendering; `internal/domain/` newtypes derive from goconf's boundary types.
- **birdcage**: `bpl/` compiler refusals use goconf's `Because[R]` carrier; birdcage's `no_clones` rule (already in policy.bpl) validates goconf's own tree via the shared lint gate.

## Lint Config Derivation

The `.golangci.yml` merges all three repos' enable lists (default: none, explicit enable only). Where settings differ, the stricter value wins: `interfacebloat max: 5` (hardstand) over birdcage's 8; `funlen 60/40`, `gocognit 10`, `cyclop 10`, `maintidx under: 20` consistent across all three. Formatters: gofumpt + goimports with `local-prefixes: github.com/adambarber/goconf`.

## Next Steps

1. Implement the `refusal` package (`carrier.go` shape from hardstand) in goconf.
2. Wire `Load`'s refusal sentinels into the refusal vocabulary (see Backlog item 2).
3. Implement domain newtypes.
4. Have each target repo import goconf and remove its inline copy.

## Backlog (from the parity review, 2026-11-06)

Each item lands as its own TDD cluster after the current fix program and reviewer cycles close.

2. **Public error-identity contract**: export the load-refusal sentinels (`ErrAmbiguous`, `ErrUnknownFormat`, `ErrUnmarshalable`, `ErrEscapingName`, `ErrOutsideSearch`) so callers branch on `errors.Is`, never string-matching. Settled as part of the goconf→gorefuse wiring step.
3. **Product limit on source size**: a public-interface option (`WithMaxBytes`) with a reasonable default value and a hard upper maximum — values above the maximum are refused outright; a source exceeding the limit refuses with a named sentinel.
4. **Env-var overlay** (documented progression path, not built now): a third source that slots into the existing layer chain as an env layer; the layer architecture keeps the seam open for it.
5. **Watch/reload is not a feature**: Load stays one-shot; watch/reload orchestration belongs to the caller. The README carries a sample watch/reload implementation built on `Load` + `Provenance` digests, so consumers have a worked starting point.

## Fix Program Status (2026-11-06)

Clusters 1-4 landed: path-escape chokepoint (f0bea95), leaf presence (eb530fa), default shapes + name boundary + empty-discovery refusal (c53f95a), symlink confinement (868ca94). Formatter gate (f87afca) and policy command ledger (80c309e) landed.

Cluster 5 landed:

- **m5 — narrow() surfaces non-IsNotExist stat errors**: only `fs.ErrNotExist` is absence; any other stat failure is refused naming the candidate path.
- **m1 — unknown candidate extensions refused by name**: `formatOf`'s silent undetermined arm removed; onePer refuses files the name boundary selected but no format spells, naming the file.
- **m2 — ReadFile refusal names path + leg**: discovery stamps each `FoundFile` with its search leg; read failures name the full path and the leg.
- **m3 — dead merge layer deleted**: `Layers`, `Layers.Present`, `layerSources`, `decodeMerged` were referenced only by their own tests (the Load path composes presence + decode inline); deleted together with those tests, their behaviors remaining pinned through Load.
- **Seam test**: the real `os.UserConfigDir` branch of `defaultSearchPaths` is pinned hermetically (HOME pinned to scratch; failure branch falls to the local leg).
