package goconf

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// assertYAMLPopulateStructRefused pins one struct-node mismatch path.
func assertYAMLPopulateStructRefused(t *testing.T, data SourceText, key FieldPath) {
	t.Helper()

	root := yamlContentNode(t, data)

	dst := reflect.New(reflect.TypeFor[intExtrasDoc]()).Elem()
	if err := yamlPopulateStructNode(root, dst, FilePath("f.yaml"), key); err == nil {
		t.Fatal("expected")
	}
}

// assertYAMLPopulateMapRefused pins one map-node mismatch path.
func assertYAMLPopulateMapRefused(t *testing.T, data SourceText, key FieldPath) {
	t.Helper()

	root := yamlContentNode(t, data)

	dst := reflect.MakeMap(reflect.MapOf(reflect.TypeFor[extraKey](), reflect.TypeFor[extraVal]()))
	if err := yamlPopulateMapNode(root, dst, FilePath("f.yaml"), key); err == nil {
		t.Fatal("expected")
	}
}

// assertJSONObjectEnd pins one consumeJSONObjectEnd outcome.
func assertJSONObjectEnd(t *testing.T, data SourceBytes, wantErr bool) {
	t.Helper()

	dec := jsonObjectDecoder(t, data)
	if _, err := decodeJSONObjectKey(dec, data, FilePath("f.json")); err != nil {
		t.Fatal(err)
	}

	if _, err := dec.ReadToken(); err != nil {
		t.Fatal(err)
	}

	err := consumeJSONObjectEnd(dec, data, FilePath("f.json"))
	if (err != nil) != wantErr {
		t.Fatalf("err=%v wantErr=%v", err, wantErr)
	}
}

// notFoundRefusal discovers from one empty root and returns the refusal.
func notFoundRefusal(t *testing.T) error {
	t.Helper()

	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{})
	t.Chdir(string(root))

	_, err := discover(Discovery{Search: []DirPath{DirPath(string(root))}, Name: configName("app"), Format: FormatJSON})
	if err == nil {
		t.Fatal("empty discovery must be refused")
	}

	return err
}

// emptyDiscoveryRefusal loads an empty search tree and pins that discovery refuses.
func emptyDiscoveryRefusal(t *testing.T, dirs ...searchDir) (error, []DirPath) {
	t.Helper()

	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{})
	paths := legPaths(root, dirs...)

	_, err := discover(Discovery{Search: paths, Name: configName("app"), Format: FormatJSON})
	if err == nil {
		t.Fatal("empty discovery must be refused")
	}

	return err, paths
}

// refusalOracle renders one input as refusal text for the fuzz oracle.
type refusalOracle func([]byte) refusalText

// assertFuzzRefusalOracle runs the shared substitution-stability check.
func assertFuzzRefusalOracle(t *testing.T, data []byte, errOf refusalOracle) {
	t.Helper()

	refusal := errOf(data)
	if refusal == "" {
		return
	}

	twinRefusal := errOf([]byte(strings.ReplaceAll(string(data), fuzzSentinel, fuzzTwin)))
	assertFuzzSubstitutionStable(t, data, refusal, twinRefusal)
}

// fuzzRefusal converts one load error into the fuzz oracle's refusal text.
func fuzzRefusal(err error) refusalText {
	if err == nil {
		return ""
	}

	return refusalText(err.Error())
}

// assertFuzzSubstitutionStable pins that swapping the marker changes only the marker.
func assertFuzzSubstitutionStable(t *testing.T, data []byte, refusal, twin refusalText) {
	t.Helper()

	if !strings.Contains(string(data), fuzzSentinel) {
		return
	}

	placeholder := func(msg refusalText) refusalText {
		return refusalText(strings.ReplaceAll(strings.ReplaceAll(string(msg), fuzzSentinel, "#"), fuzzTwin, "#"))
	}

	if string(placeholder(stripQuotedFuzz(refusal))) != string(placeholder(stripQuotedFuzz(twin))) {
		t.Fatalf("refusal is not substitution-stable (value echo):\n got: %s\ntwin: %s", refusal, twin)
	}
}

// assertResourceExtras pins the fixed fields and captured extra on one resource.
func assertResourceExtras(t *testing.T, got resourceDoc) {
	t.Helper()

	if got.Resource.Name != "r" || got.Resource.Kind != "k" {
		t.Fatalf("fixed: %#v", got.Resource)
	}

	if got.Resource.Extra[extraKey("provider")] != extraVal(testProvider) {
		t.Fatalf("extra provider = %#v", got.Resource.Extra[extraKey("provider")])
	}
}

// unlistableDir creates a directory whose listing faults, and cleans it up.
func unlistableDir(t *testing.T, root DirPath) DirPath {
	t.Helper()

	path := filepath.Join(string(root), "unlistable")
	if err := os.MkdirAll(path, 0); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	if err := os.Chmod(path, 0); err != nil {
		t.Fatalf("chmod: %v", err)
	}

	t.Cleanup(func() {
		if err := os.Remove(path); err != nil {
			t.Logf("cleanup remove: %v", err)
		}
	})

	return DirPath(path)
}
