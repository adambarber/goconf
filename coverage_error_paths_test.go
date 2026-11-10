package goconf

import (
	"encoding/json/jsontext"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestYAMLTerrorHelpersCoverage(t *testing.T) {
	_ = stripYAMLTypeLine(refusalText(""))
	_ = stripYAMLTypeLine(refusalText("line 1: cannot unmarshal !!str `x` into int"))
	_ = stripYAMLTypeLine(refusalText("line 1: !!str `x` cannot unmarshal"))
	_ = cutTypeSpan(terrorTail("!!str `x` more"))
	_ = cutTypeSpan(terrorTail("no type here"))

	err := errValueFree(&yaml.TypeError{Errors: []string{"line 1: cannot unmarshal !!str `x` into int"}})
	_ = err

	err = errValueFree(refusalError("plain \"error\""))
	_ = err
}

func TestTOMLWordAndFoundTermCoverage(t *testing.T) {
	_ = tomlWordKind(kindWord("true"))
	_ = tomlWordKind(kindWord("false"))
	_ = tomlWordKind(kindWord("1.2"))
	_ = tomlWordKind(kindWord("123"))
	_ = tomlWordKind(kindWord(`"x"`))
	_ = tomlWordKind(kindWord("[1]"))
	_ = tomlWordKind(kindWord("{a=1}"))
	_ = tomlWordKind(kindWord("other"))
	_ = tomlFoundTerm(tomlMessage(""))
	_ = tomlFoundTerm(tomlMessage("true"))
	_ = tomlFoundTerm(tomlMessage(`"hello"`))
	_ = tomlFoundTerm(tomlMessage("42"))
	_ = tomlFoundTerm(tomlMessage("expected string but found integer"))
}

func TestJSONExtrasRootShapeCoverage(t *testing.T) {
	if err := jsonExtrasRootShape(FilePath("f.json"), '[', reflect.ValueOf(struct{}{})); err == nil {
		t.Fatal("expected")
	}

	if err := jsonExtrasRootShape(FilePath("f.json"), '{', reflect.ValueOf([]int{})); err == nil {
		t.Fatal("expected")
	}

	if err := jsonExtrasRootShape(FilePath("f.json"), '{', reflect.ValueOf(struct{}{})); err != nil {
		t.Fatal(err)
	}

	if err := jsonExtrasRootShape(FilePath("f.json"), jsontext.Kind('"'), reflect.ValueOf(0)); err != nil {
		t.Fatal(err)
	}
}

func TestYAMLMapFieldDuplicateKeyRefused(t *testing.T) {
	type doc struct {
		Extra map[extraKey]extraVal `goconf:"extras" json:"-" toml:"-" yaml:"-"`
		M     map[extraKey]extraVal `json:"m"        toml:"m" yaml:"m"`
	}

	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameYAML: SourceBytes("m:\n  a: 1\n  a: 2\nx: y\n")},
	})
	t.Chdir(string(root))

	_, _, err := Load[doc](Spec{Name: configName("app"), Format: FormatYAML},
		WithSearchPaths(DirPath(string(root))))
	if err == nil {
		t.Fatal("expected refusal")
	}
}

func TestYAMLExtrasMapValueNodeMismatch(t *testing.T) {
	type mapVal map[extraKey]extraVal

	type doc struct {
		Extra map[extraKey]mapVal `goconf:"extras" json:"-"    toml:"-"    yaml:"-"`
		Name  resName             `json:"name"     toml:"name" yaml:"name"`
	}

	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {fileNameYAML: SourceBytes("name: r\nm: not-a-map\n")},
	})
	t.Chdir(string(root))

	_, _, err := Load[doc](Spec{Name: configName("app"), Format: FormatYAML},
		WithSearchPaths(DirPath(string(root))))
	if err == nil {
		t.Fatal("expected refusal")
	}
}

func TestYAMLTagKindCoverage(t *testing.T) {
	_ = yamlTagKind(kindWord("!!str"))
	_ = yamlTagKind(kindWord("!!int"))
	_ = yamlTagKind(kindWord("!!bool"))
	_ = yamlTagKind(kindWord("!!float"))
	_ = yamlTagKind(kindWord("!!null"))
	_ = yamlTagKind(kindWord("!!seq"))
	_ = yamlTagKind(kindWord("!!map"))
	_ = yamlTagKind(kindWord(""))
	_ = yamlTagKind(kindWord("!custom"))
}

func TestYAMLExtrasRootNotMappingDirect(t *testing.T) {
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte("x\n"), &doc); err != nil {
		t.Fatal(err)
	}

	root := &doc
	if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
		root = root.Content[0]
	}

	if err := yamlExtrasRootNotMapping(FilePath("f.yaml"), root); err == nil {
		t.Fatal("expected")
	}
}

func TestDerefValueAndAllocCoverage(t *testing.T) {
	v := reflect.ValueOf((*ptrInner)(nil))
	_ = derefValue(v)
	_ = derefAlloc(reflect.New(reflect.TypeFor[*ptrInner]()).Elem())
}
