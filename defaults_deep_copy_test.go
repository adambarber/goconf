package goconf

import "testing"

type deepWord string

type deepKey string

// deepDecodable is a named interface so the default clone path is exercised
// without an untyped any in a field position.
type deepDecodable interface {
	decode() error
}

type deepText struct {
	Value deepWord
}

func (d deepText) decode() error { return nil }

type deepInner struct {
	Name deepWord `json:"name"`
}

type deepDoc struct {
	Map    map[deepKey]deepWord `json:"map"`
	Nested *deepInner           `json:"nested"`
	Slice  []deepWord           `json:"slice"`
}

type ifaceDoc struct {
	Value deepDecodable `json:"value"`
}

const (
	deepProbeName deepWord = "original"
	deepMutated   deepWord = "mutated"
	deepChanged   deepWord = "changed"
)

func assertDeepDoc(t *testing.T, got deepDoc) {
	t.Helper()

	if got.Nested == nil || got.Nested.Name != deepProbeName {
		t.Fatalf("nested = %+v, want name %q", got.Nested, deepProbeName)
	}

	if len(got.Slice) != 1 || got.Slice[0] != deepProbeName {
		t.Fatalf("slice = %v, want [%q]", got.Slice, deepProbeName)
	}

	if got.Map["key"] != deepProbeName {
		t.Fatalf("map = %v, want key %q", got.Map, deepProbeName)
	}
}

func assertIfaceDoc(t *testing.T, got ifaceDoc) {
	t.Helper()

	decoded, ok := got.Value.(deepText)
	if !ok || decoded.Value != deepProbeName {
		t.Fatalf("interface value = %v, want %q", got.Value, deepProbeName)
	}
}

// TestDefaultsDeepCopyIsolatesLoads pins that a stored default shares no
// backing store with the caller's value or with a previously loaded config.
func TestDefaultsDeepCopyIsolatesLoads(t *testing.T) {
	def := deepDoc{
		Slice:  []deepWord{deepProbeName},
		Map:    map[deepKey]deepWord{"key": deepProbeName},
		Nested: &deepInner{Name: deepProbeName},
	}

	opt := WithDefaults(def)
	def.Nested.Name = deepMutated
	def.Slice[0] = deepMutated
	def.Map["key"] = deepMutated

	first := loadDefaulted[deepDoc](t, SourceBytes(`{}`), opt)
	assertDeepDoc(t, first)

	first.Nested.Name = deepChanged
	first.Slice[0] = deepChanged
	first.Map["key"] = deepChanged

	second := loadDefaulted[deepDoc](t, SourceBytes(`{}`), opt)
	assertDeepDoc(t, second)
}

// TestDefaultsDeepCopyInterface pins that a named-interface default is copied
// without sharing its captured value.
func TestDefaultsDeepCopyInterface(t *testing.T) {
	def := ifaceDoc{Value: deepText{Value: deepProbeName}}
	opt := WithDefaults(def)

	first := loadDefaulted[ifaceDoc](t, SourceBytes(`{}`), opt)
	assertIfaceDoc(t, first)

	second := loadDefaulted[ifaceDoc](t, SourceBytes(`{}`), opt)
	assertIfaceDoc(t, second)
}
