package goconf

import (
	"strings"
	"testing"
)

// TestNotFoundProseSpeaksToTheUser pins the not-found refusal shape.
func TestNotFoundProseSpeaksToTheUser(t *testing.T) {
	err := notFoundRefusal(t)

	for _, want := range []userSpan{
		"no config file found",
		"the app looked for app.json in:",
		"create your config at one of those locations",
	} {
		if !strings.Contains(err.Error(), string(want)) {
			t.Fatalf("refusal %q must contain %q", err, want)
		}
	}

	for _, ban := range []string{"widen the search", "some format spells"} {
		if strings.Contains(err.Error(), ban) {
			t.Fatalf("refusal %q must not contain %q", err, ban)
		}
	}
}

// TestNotFoundLabelsTheFaultOnce pins that the not-found refusal does not
// double-label: the Reason already says "no config file found", so the
// Subject carries only the looked-for names, the paths, and the action.
func TestNotFoundLabelsTheFaultOnce(t *testing.T) {
	err := notFoundRefusal(t)

	if got := strings.Count(err.Error(), "no config file"); got != 1 {
		t.Fatalf("refusal %q says \"no config file\" %d times; one label", err, got)
	}

	if strings.Contains(err.Error(), "was not found") {
		t.Fatalf("refusal %q repeats the reason's label in the subject", err)
	}

	for _, want := range []string{"the app looked for app.json in:", "create your config at one of those locations"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("refusal %q must contain %q", err, want)
		}
	}
}
