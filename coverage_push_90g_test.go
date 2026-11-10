package goconf

import "testing"

func TestTOMLFoundTermEdgeCoverage(t *testing.T) {
	_ = tomlFoundTerm(tomlMessage("expected datetime but found string"))
	_ = tomlFoundTerm(tomlMessage("expected  but found "))
	_ = tomlFoundTerm(tomlMessage("no marker"))
}

func TestLineAtByteMiddle(t *testing.T) {
	if n := lineAtByte(SourceBytes("a\nb\nc"), 2); n != 2 {
		t.Fatalf("n=%d", n)
	}

	if n := lineAtByte(SourceBytes("a\nb\nc"), 4); n != 3 {
		t.Fatalf("n=%d", n)
	}
}
