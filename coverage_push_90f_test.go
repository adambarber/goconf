package goconf

import "testing"

func TestYAMLPopulateMapNodeRootScalarMismatch(t *testing.T) {
	assertYAMLPopulateMapRefused(t, SourceText("x\n"), "")
}
