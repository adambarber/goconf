package goconf

import (
	"reflect"
	"testing"
)

func TestYAMLPopulateMapNodeRootMismatch(t *testing.T) {
	assertYAMLPopulateMapRefused(t, SourceText("- a\n"), "")
}

func TestYAMLWantKindMoreCoverage(t *testing.T) {
	_ = yamlWantKind(reflect.TypeFor[uint](), "k")
	_ = yamlWantKind(reflect.TypeFor[float64](), "k")
	_ = yamlWantKind(reflect.TypeFor[[2]int](), "k")
}
