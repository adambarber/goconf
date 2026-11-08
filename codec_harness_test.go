package goconf

import (
	"reflect"
	"strings"
	"testing"
)

// codecCase is one strict-decode behavior: bytes in, value or refusal
// out. want is the fully-spelled expected config when wantErr is none.
type codecCase struct {
	data    SourceBytes
	name    caseName
	wantErr refusalNeed
	want    codecCfg
}

// codecCfg is the shared strict-decode target: one nested config all
// three format tests decode into.
type codecCfg struct {
	Region  region    `json:"region"`
	Nested  nestedCfg `json:"nested"`
	Retries int       `json:"retries"`
}

// codecWant composes a fully-spelled expected decode value; the region
// is always "eu" in these fixtures.
func codecWant(retries int, nested nestedCfg) codecCfg {
	return codecCfg{Nested: nested, Region: "eu", Retries: retries}
}

// zeroCodecWant is the value a refusal case decodes to.
func zeroCodecWant() codecCfg {
	return codecWant(0, nestedCfgWant("", 0))
}

// fileDecode adapts one strict decode to the file-carrying decoders:
// the refusal frame names the file the bytes are declared to come from.
func fileDecode(decode func(SourceBytes, FilePath, Target) error, file FileName) func(SourceBytes, Target) error {
	return func(data SourceBytes, dst Target) error {
		return decode(data, FilePath(string(file)), dst)
	}
}

// decodeCase runs one strict decode and pins the contract: a refusal
// case errors with the required fragment, a clean case decodes to the
// exact expected value.
func decodeCase(t *testing.T, tc *codecCase, decode func(SourceBytes, Target) error) {
	t.Helper()

	var got codecCfg

	err := decode(tc.data, targetOf(&got))
	assertDecoded(t, tc.wantErr, tc.want, got, err)
}

// targetOf builds the reflect Target for one test destination.
func targetOf(v *codecCfg) Target {
	return Target{Value: reflect.ValueOf(v).Elem()}
}

// assertDecoded pins the strict-decode contract for one decoded value:
// a refusal case errors with the required fragment, a clean case decodes
// to the exact expected value.
func assertDecoded(t *testing.T, wantErr refusalNeed, want, got codecCfg, err error) {
	t.Helper()

	if refusalSatisfied(t, wantErr, err) {
		return
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

// refusalSatisfied fails the test unless the case's error contract
// holds; it reports whether the check is conclusive either way.
func refusalSatisfied(t *testing.T, want refusalNeed, err error) bool {
	t.Helper()

	switch {
	case want == "" && err != nil:
		t.Fatalf("unexpected error: %v", err)
	case want != "" && err == nil:
		t.Fatalf("expected error containing %q, got nil", want)
	case want != "" && !strings.Contains(err.Error(), string(want)):
		t.Fatalf("error %q does not contain %q", err, want)
	}

	return want != ""
}

// keysPresent asserts the key set contains every wanted path and none
// of the explicitly-absent ones.
func keysPresent(t *testing.T, keys KeySet, want, absent []FieldPath) {
	t.Helper()

	for _, p := range want {
		if !keys.Has(p) {
			t.Errorf("expected key %q in %v", p, keys)
		}
	}

	for _, p := range absent {
		if keys.Has(p) {
			t.Errorf("%q is absent and must not be present", p)
		}
	}
}

// fileKeys adapts a file-carrying key walk to the one-argument shape
// presentKeysOK expects, naming the file the bytes are declared to come
// from.
func fileKeys(walk func(SourceBytes, FilePath) (KeySet, error), file FileName) func(SourceBytes) (KeySet, error) {
	return func(data SourceBytes) (KeySet, error) {
		return walk(data, FilePath(string(file)))
	}
}

// presentKeysOK walks keys from one source and fails the test on error.
func presentKeysOK(t *testing.T, walk func(SourceBytes) (KeySet, error), data SourceBytes) KeySet {
	t.Helper()

	keys, err := walk(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	return keys
}
