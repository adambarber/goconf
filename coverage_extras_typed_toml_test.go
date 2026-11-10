package goconf

import (
	"errors"
	"testing"
)

func TestExtrasTypedIntTOML(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameTOML: SourceBytes("name = \"r\"\nn = 42\n")},
	})
	t.Chdir(string(root))

	got, _, err := Load[intExtrasDoc](Spec{Name: configName("app"), Format: FormatTOML},
		WithSearchPaths(DirPath(string(root))))
	if err != nil || got.Extra["n"] != 42 {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestExtrasTypedBoolTOML(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameTOML: SourceBytes("name = \"r\"\nflag = true\n")},
	})
	t.Chdir(string(root))

	got, _, err := Load[boolExtrasDoc](Spec{Name: configName("app"), Format: FormatTOML},
		WithSearchPaths(DirPath(string(root))))
	if err != nil || !bool(got.Extra["flag"]) {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestExtrasTypedFloatTOML(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameTOML: SourceBytes("name = \"r\"\nf = 1.5\n")},
	})
	t.Chdir(string(root))

	got, _, err := Load[floatExtrasDoc](Spec{Name: configName("app"), Format: FormatTOML},
		WithSearchPaths(DirPath(string(root))))
	if err != nil || got.Extra["f"] != 1.5 {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestExtrasTypedNarrowIntTOML(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameTOML: SourceBytes("name = \"r\"\nn = 42\n")},
	})
	t.Chdir(string(root))

	got, _, err := Load[int8ExtrasDoc](Spec{Name: configName("app"), Format: FormatTOML},
		WithSearchPaths(DirPath(string(root))))
	if err != nil || got.Extra["n"] != 42 {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestExtrasTypedUintTOMLRefused(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameTOML: SourceBytes("name = \"r\"\nu = 7\n")},
	})
	t.Chdir(string(root))

	_, _, err := Load[uintExtrasDoc](Spec{Name: configName("app"), Format: FormatTOML},
		WithSearchPaths(DirPath(string(root))))
	if err == nil {
		t.Fatal("expected refusal")
	}

	if !errors.Is(err, ErrTypeMismatch) {
		t.Fatalf("err=%v", err)
	}
}

func TestExtrasTypedInt8OverflowTOML(t *testing.T) {
	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameTOML: SourceBytes("name = \"r\"\nn = 999999\n")},
	})
	t.Chdir(string(root))

	_, _, err := Load[int8ExtrasDoc](Spec{Name: configName("app"), Format: FormatTOML},
		WithSearchPaths(DirPath(string(root))))
	if err == nil {
		t.Fatal("expected refusal")
	}

	if !errors.Is(err, ErrTypeMismatch) {
		t.Fatalf("err=%v", err)
	}
}
