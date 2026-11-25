package goconf

import (
	"fmt"
	"testing"
	"time"
)

type redactTimeCfg struct {
	When time.Time
}

func TestRedactMatchesGoStringForOpaqueLeaf(t *testing.T) {
	cfg := redactTimeCfg{When: time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)}
	want := fmt.Sprintf("%+v", cfg)

	if got := string(Redact(cfg)); got != want {
		t.Fatalf("Redact = %q, want %q", got, want)
	}
}
