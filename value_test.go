package goconf

import (
	"errors"
	"testing"
)

type (
	valueKey  string
	valueWord string
	envTable  struct {
		File valueWord `toml:"file"`
		Mode int       `toml:"mode"`
	}
	valueDoc struct {
		Env map[valueKey]Value `json:"env" toml:"env" yaml:"env"`
	}
)

func loadValueDoc(t *testing.T, src SourceBytes) valueDoc {
	t.Helper()

	root := oneLegTree(t, map[FileName]SourceBytes{fileNameTOML: src})

	got, _, err := Load[valueDoc](Spec{Name: configName("app"), Format: FormatTOML},
		WithSearchPaths(legPaths(root, dirLocal)...))
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	return got
}

func TestValueDecodesStringAndTableInOnePass(t *testing.T) {
	got := loadValueDoc(t, SourceBytes("env.KEY = \"literal\"\nenv.OTHER = {file = \"p\", mode = 1}\n"))

	literal, err := DecodeValue[valueWord](got.Env["KEY"])
	if err != nil {
		t.Fatalf("DecodeValue string: %v", err)
	}

	if literal != "literal" {
		t.Fatalf("literal = %q, want literal", literal)
	}

	table, err := DecodeValue[envTable](got.Env["OTHER"])
	if err != nil {
		t.Fatalf("DecodeValue table: %v", err)
	}

	if table.File != "p" || table.Mode != 1 {
		t.Fatalf("table = %+v, want file=p mode=1", table)
	}
}

func TestDecodeValueRefusesUnknownKeyInTable(t *testing.T) {
	got := loadValueDoc(t, SourceBytes("env.OTHER = {file = \"p\", nope = 1}\n"))

	_, err := DecodeValue[envTable](got.Env["OTHER"])
	if !errors.Is(err, ErrUnknownField) {
		t.Fatalf("err = %v, want ErrUnknownField", err)
	}
}

func TestValueFieldsRefusedOutsideTOML(t *testing.T) {
	for _, tc := range []struct {
		format Format
		name   FileName
		src    SourceBytes
	}{
		{FormatJSON, fileNameJSON, SourceBytes(`{"env":{"KEY":"literal"}}`)},
		{FormatYAML, fileNameYAML, SourceBytes("env:\n  KEY: literal\n")},
	} {
		root := oneLegTree(t, map[FileName]SourceBytes{tc.name: tc.src})

		_, _, err := Load[valueDoc](Spec{Name: configName("app"), Format: tc.format},
			WithSearchPaths(legPaths(root, dirLocal)...))
		if !errors.Is(err, ErrUnsupportedValueContainer) {
			t.Fatalf("%s: err = %v, want ErrUnsupportedValueContainer", tc.format, err)
		}
	}
}

type valueOccurrence struct {
	path OccurrencePath
	key  OccurrenceKey
	line int
}

func TestValuePreservesOccurrenceMetadataForBothShapes(t *testing.T) {
	src := SourceBytes("env.KEY = \"literal\"\nenv.OTHER = {file = \"p\", mode = 1}\n")
	root := oneLegTree(t, map[FileName]SourceBytes{fileNameTOML: src})

	_, _, meta, err := LoadWithMeta[valueDoc](Spec{Name: configName("app"), Format: FormatTOML},
		WithSearchPaths(legPaths(root, dirLocal)...))
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	entries := meta.Entries()
	got := make([]valueOccurrence, 0, len(entries))

	for _, e := range entries {
		got = append(got, valueOccurrence{e.Path, e.Key, e.Line})
	}

	want := []valueOccurrence{
		{"env.KEY", "KEY", 1},
		{"env.OTHER", "OTHER", 2},
		{"env.OTHER.file", "file", 2},
		{"env.OTHER.mode", "mode", 2},
	}

	if len(got) != len(want) {
		t.Fatalf("occurrences = %v, want %v", got, want)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("occurrence %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestDecodeValueOnZeroRefuses(t *testing.T) {
	var empty Value

	_, err := DecodeValue[valueWord](empty)

	if !errors.Is(err, ErrDecoderRefused) {
		t.Fatalf("err = %v, want ErrDecoderRefused", err)
	}
}
