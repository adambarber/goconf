package goconf

import (
	"reflect"
	"testing"
)

func TestDerefTargetTypeCoverage(t *testing.T) {
	v := reflect.ValueOf(new(*intExtrasDoc))
	_ = derefTargetType(v)

	var zero intExtrasDoc

	_ = derefTargetType(reflect.ValueOf(zero))
}

func TestDerefValuePointerChain(t *testing.T) {
	x := 1
	p := &x
	pp := &p
	v := reflect.ValueOf(&pp)
	_ = derefValue(v)
}

func TestPresenceOfCoverage(t *testing.T) {
	type doc struct {
		A resName `json:"a"`
	}

	_, err := presenceOf(SourceBytes(`{"a":"x"}`), FormatJSON, FilePath("f.json"), reflect.TypeFor[doc]())
	if err != nil {
		t.Fatal(err)
	}

	_, err = presenceOf(SourceBytes(`a = "x"`), FormatTOML, FilePath("f.toml"), reflect.TypeFor[doc]())
	if err != nil {
		t.Fatal(err)
	}

	_, err = presenceOf(SourceBytes("a: x\n"), FormatYAML, FilePath("f.yaml"), reflect.TypeFor[doc]())
	if err != nil {
		t.Fatal(err)
	}
}

func TestCandidatePathCoverage(t *testing.T) {
	_, err := candidatePath(DirPath("/tmp"), configName("app"), Extension(".json"))
	if err != nil {
		t.Fatal(err)
	}

	_, err = candidatePath(DirPath("/tmp"), configName("app"), Extension(".yaml"))
	if err != nil {
		t.Fatal(err)
	}

	_, err = candidatePath(DirPath("/tmp"), configName("app"), Extension(".toml"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestYAMLPopulateStructNodeAlias(t *testing.T) {
	type doc struct {
		Extra map[extraKey]extraVal `goconf:"extras" yaml:"-"`
		A     extraVal              `yaml:"a"`
	}

	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameYAML: SourceBytes("a: &v x\nb: *v\n")},
	})
	t.Chdir(string(root))

	got, _, err := Load[doc](Spec{Name: configName("app"), Format: FormatYAML},
		WithSearchPaths(DirPath(string(root))))
	if err != nil || got.A != "x" || got.Extra["b"] != "x" {
		t.Fatalf("got=%v err=%v", got, err)
	}
}
