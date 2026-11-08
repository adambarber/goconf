package goconf

import (
	"reflect"
	"strings"
	"testing"
)

// ptrCfg targets a pointer-typed field: the mismatch classification must
// dereference indirection before naming the shape.
type ptrCfg struct {
	Port *int `json:"port"`
}

// mixedCfg targets one field of every scalar kind, plus a list: each
// decoder's found-word vocabulary must classify into the shared kinds.
type mixedCfg struct {
	Nested nestedCfg `json:"nested"`
	Tags   []tagWord `json:"tags"`
	Count  int       `json:"count"`
	Ratio  float64   `json:"ratio"`
	Flag   bool      `json:"flag"`
}

// foundShapeCase is one found-value row: bytes, the decode target's
// constructor, and the user's word the refusal must render.
type foundShapeCase struct {
	target func() Target
	name   caseName
	found  userTerm
	body   SourceBytes
}

// shapeTable is the user's word for a table-valued config node.
const shapeTable = userTerm("a table")

// zeroMixed is the zero-merged mixed-kind decode destination.
func zeroMixed() Target {
	return Target{Value: reflect.ValueOf(&mixedCfg{Nested: nestedCfgWant("", 0), Tags: nil, Count: 0, Ratio: 0, Flag: false}).Elem()}
}

// zeroPtr is the zero-merged pointer-field decode destination.
func zeroPtr() Target {
	return Target{Value: reflect.ValueOf(&ptrCfg{Port: nil}).Elem()}
}

// jsonFoundShapeCases pins JSON's found-word vocabulary through decode.
func jsonFoundShapeCases() []foundShapeCase {
	return []foundShapeCase{
		{zeroPtr, caseName("object found"), shapeTable, SourceBytes(`{"port": {"a": 1}}`)},
		{zeroMixed, caseName("number found"), "a decimal number", SourceBytes(`{"count": 1.5}`)},
		{zeroMixed, caseName("bool found"), "yes/no", SourceBytes(`{"ratio": true}`)},
		{zeroMixed, caseName("array found"), "a list", SourceBytes(`{"flag": [1]}`)},
	}
}

// yamlFoundShapeCases pins YAML's echoed-tag vocabulary through decode.
// A float into an int field decodes (yaml.v3 coerces), so only the tags
// that genuinely refuse are pinned.
func yamlFoundShapeCases() []foundShapeCase {
	return []foundShapeCase{
		{zeroMixed, caseName("seq found"), "a list", SourceBytes("ratio: [1]\n")},
		{zeroMixed, caseName("map found"), shapeTable, SourceBytes("flag:\n  a: 1\n")},
		{zeroPtr, caseName("map into pointer"), shapeTable, SourceBytes("port:\n  a: 1\n")},
	}
}

// tomlFoundShapeCases pins TOML's found-word vocabulary through decode:
// BurntSushi renders "TOML value has type" only when the found value is
// a string, so that is the reachable row.
func tomlFoundShapeCases() []foundShapeCase {
	return []foundShapeCase{
		{zeroMixed, caseName("string found"), "text", SourceBytes("flag = \"x\"\n")},
	}
}

// assertFoundShape runs one vocabulary row: the refusal must say the
// found shape, naming it in the user's word.
func assertFoundShape(t *testing.T, tc foundShapeCase, decode func(SourceBytes, Target) error) {
	t.Helper()

	err := decode(tc.body, tc.target())
	if err == nil {
		t.Fatalf("mismatch in %s must be refused", tc.body)
	}

	want := "is " + string(tc.found) + ", not"
	if !strings.Contains(err.Error(), want) {
		t.Errorf("refusal %q must say %q", err, want)
	}
}

// TestFoundWordVocabularyCoversEveryShape pins the one value vocabulary
// across formats: every decoder's found-word or echoed-tag word renders
// as the user's word for the shape, per format.
func TestFoundWordVocabularyCoversEveryShape(t *testing.T) {
	suites := []struct {
		decode func(SourceBytes, Target) error
		cases  []foundShapeCase
	}{
		{fileDecode(decodeJSONStrict, fileNameJSON), jsonFoundShapeCases()},
		{fileDecode(decodeYAMLStrict, fileNameYAML), yamlFoundShapeCases()},
		{fileDecode(decodeTOMLStrict, fileNameTOML), tomlFoundShapeCases()},
	}

	for _, suite := range suites {
		for _, tc := range suite.cases {
			t.Run(string(tc.name), func(t *testing.T) {
				assertFoundShape(t, tc, suite.decode)
			})
		}
	}
}

// TestMismatchIntoPointerFieldNamesTheBaseKind pins that pointer
// indirection never leaks into the refusal's shape vocabulary.
func TestMismatchIntoPointerFieldNamesTheBaseKind(t *testing.T) {
	err := decodeYAMLStrict(SourceBytes("port: x\n"), FilePath(string(fileNameYAML)), zeroPtr())
	if err == nil {
		t.Fatalf("text into a pointer-int field must be refused")
	}

	if !strings.Contains(err.Error(), "not a whole number") {
		t.Fatalf("refusal %q must name the base kind", err)
	}
}

// TestJSONKeyLineAtDeepOffset pins the byte-offset to line translation:
// a key on a later line is named with that line in the refusal.
func TestJSONKeyLineAtDeepOffset(t *testing.T) {
	body := SourceBytes("{\n\n\n  \"region\": \"eu\",\n  \"nonesuch\": 1\n}")

	var got codecCfg

	err := decodeJSONStrict(body, FilePath(string(fileNameJSON)), targetOf(&got))
	if err == nil {
		t.Fatalf("unknown field must be refused")
	}

	for _, need := range []string{"app.json:5:", `"nonesuch"`} {
		if !strings.Contains(err.Error(), need) {
			t.Errorf("refusal %q must contain %q", err, need)
		}
	}
}

// TestYAMLTypeErrorCarriesEveryIssue pins the TypeError arm of the
// value-freedom pass: a two-fault document classifies by its first
// issue and echoes no config value.
func TestYAMLTypeErrorCarriesEveryIssue(t *testing.T) {
	body := SourceBytes("count: not-a-number\nratio: also-text\n")

	err := decodeYAMLStrict(body, FilePath(string(fileNameYAML)), zeroMixed())
	if err == nil {
		t.Fatalf("two type faults must be refused")
	}

	if !strings.Contains(err.Error(), "not a whole number") {
		t.Errorf("refusal %q must classify the first issue", err)
	}

	for _, leak := range []string{"not-a-number", "also-text"} {
		if strings.Contains(err.Error(), leak) {
			t.Errorf("refusal %q echoed a config value", err)
		}
	}
}
