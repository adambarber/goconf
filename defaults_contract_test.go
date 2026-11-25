package goconf

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type (
	defWord string
	defInt  int
	defKey  string

	defSrc struct {
		Second defWord `json:"second"`
		First  defWord `json:"first"`
	}
	defDoc struct {
		First  defWord `json:"first"`
		Second defWord `json:"second"`
	}
	shortDefault struct {
		First defWord `json:"first"`
	}
	wrongTypeDefault struct {
		Port defWord `json:"port"`
	}
	portDoc struct {
		Port defInt `json:"port"`
	}
	timeDoc struct {
		When time.Time `json:"when"`
	}
	opaqueValue struct {
		Hidden defWord `json:"hidden"`
	}
	opaqueDoc struct {
		Value opaqueValue `json:"value"`
	}
	aliasDoc struct {
		Labels map[defKey]defWord `json:"labels"`
	}
	extrasDoc struct {
		Extra map[extraKey]extraVal `goconf:"extras" json:"-"`
		Name  defWord               `json:"name"`
	}
)

const (
	firstDefault  defWord = "FIRST_VAL"
	secondDefault defWord = "SECOND_VAL"
)

// UnmarshalJSON makes opaqueValue a self-decoding leaf despite its field.
func (*opaqueValue) UnmarshalJSON(_ []byte) error { return nil }

// loadDefaulted loads src from a single local leg with the given options.
func loadDefaulted[T any](t *testing.T, src SourceBytes, opts ...Option) T {
	t.Helper()

	root := oneLegTree(t, map[FileName]SourceBytes{fileNameJSON: src})
	all := append([]Option{WithSearchPaths(legPaths(root, dirLocal)...)}, opts...)

	got, _, err := Load[T](Spec{Name: configName("app"), Format: FormatJSON}, all...)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	return got
}

func TestDefaultsMatchByCanonicalPathNotFieldIndex(t *testing.T) {
	got := loadDefaulted[defSrc](t, SourceBytes(`{}`),
		WithDefaults(defDoc{First: firstDefault, Second: secondDefault}))

	if got.First != firstDefault || got.Second != secondDefault {
		t.Fatalf("got first=%q second=%q, want FIRST_VAL/SECOND_VAL", got.First, got.Second)
	}
}

func TestShortDefaultRefusesAtLoad(t *testing.T) {
	root := oneLegTree(t, map[FileName]SourceBytes{fileNameJSON: SourceBytes(`{}`)})

	_, _, err := Load[defDoc](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(legPaths(root, dirLocal)...),
		WithDefaults(shortDefault{First: firstDefault}))

	if err == nil || !errors.Is(err, ErrBadDefaultShape) {
		t.Fatalf("err = %v, want ErrBadDefaultShape, no panic", err)
	}
}

func TestBadDefaultRefusesBeforeReadingSource(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{dirLocal: {}})

	candidate := filepath.Join(string(root), string(dirLocal), string(fileNameJSON))
	if err := os.Mkdir(candidate, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	_, _, err := Load[defDoc](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(legPaths(root, dirLocal)...),
		WithDefaults(shortDefault{First: firstDefault}))

	if !errors.Is(err, ErrBadDefaultShape) {
		t.Fatalf("err = %v, want ErrBadDefaultShape before the IO fault", err)
	}
}

func TestTypeMismatchedDefaultRefusesAtLoad(t *testing.T) {
	root := oneLegTree(t, map[FileName]SourceBytes{fileNameJSON: SourceBytes(`{}`)})

	_, _, err := Load[portDoc](Spec{Name: configName("app"), Format: FormatJSON},
		WithSearchPaths(legPaths(root, dirLocal)...),
		WithDefaults(wrongTypeDefault{Port: "8080"}))

	if err == nil || !errors.Is(err, ErrBadDefaultShape) {
		t.Fatalf("err = %v, want ErrBadDefaultShape, no panic", err)
	}
}

func TestOpaqueLeafDefaultFilled(t *testing.T) {
	when := time.Date(2020, time.March, 4, 5, 6, 7, 0, time.UTC)
	got := loadDefaulted[timeDoc](t, SourceBytes(`{}`), WithDefaults(timeDoc{When: when}))

	if !got.When.Equal(when) {
		t.Fatalf("When = %v, want %v", got.When, when)
	}
}

func TestSelfDecodingStructDefaultFilledAsLeaf(t *testing.T) {
	want := opaqueValue{Hidden: "opaque"}
	got := loadDefaulted[opaqueDoc](t, SourceBytes(`{}`), WithDefaults(opaqueDoc{Value: want}))

	if got.Value.Hidden != want.Hidden {
		t.Fatalf("Hidden = %q, want %q", got.Value.Hidden, want.Hidden)
	}
}

func TestDefaultCompositeNotAliasedAcrossLoads(t *testing.T) {
	def := aliasDoc{Labels: map[defKey]defWord{"k": "v"}}

	first := loadDefaulted[aliasDoc](t, SourceBytes(`{}`), WithDefaults(def))
	first.Labels["k"] = "MUTATED"

	second := loadDefaulted[aliasDoc](t, SourceBytes(`{}`), WithDefaults(def))
	if second.Labels["k"] != "v" {
		t.Fatalf("second load saw %q, want v", second.Labels["k"])
	}
}

func TestDefaultsPreserveExplicitExtras(t *testing.T) {
	got := loadDefaulted[extrasDoc](t, SourceBytes(`{"name":"r","provider":"aws"}`),
		WithDefaults(extrasDoc{Name: "r", Extra: map[extraKey]extraVal{"fallback": "x"}}))

	if len(got.Extra) != 1 || got.Extra["provider"] != "aws" {
		t.Fatalf("Extra = %v, want only provider=aws", got.Extra)
	}
}

func TestPointerTargetUsesDefaults(t *testing.T) {
	got := loadDefaulted[*portDoc](t, SourceBytes(`{}`), WithDefaults(portDoc{Port: 8080}))

	if got == nil || got.Port != 8080 {
		t.Fatalf("got %+v, want Port=8080", got)
	}
}

func TestPointerDefaultUsesDefaults(t *testing.T) {
	got := loadDefaulted[portDoc](t, SourceBytes(`{}`), WithDefaults(&portDoc{Port: 8080}))

	if got.Port != 8080 {
		t.Fatalf("Port = %d, want 8080", got.Port)
	}
}

func TestFirstDefaultLayerWins(t *testing.T) {
	got := loadDefaulted[portDoc](t, SourceBytes(`{}`),
		WithDefaults(portDoc{Port: 1}), WithDefaults(portDoc{Port: 2}))

	if got.Port != 1 {
		t.Fatalf("Port = %d, want 1", got.Port)
	}
}
