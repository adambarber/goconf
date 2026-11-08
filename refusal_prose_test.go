package goconf

import (
	"strings"
	"testing"
)

// goldenRefusalCases covers every user-facing refusal shape.
func goldenRefusalCases() []proseCase {
	return append(goldenVocabCases(), goldenStructuralCases()...)
}

// goldenVocabCases pins the unknown-field and type-mismatch shapes.
func goldenVocabCases() []proseCase {
	return []proseCase{
		unknownFieldCase("json unknown field", FormatJSON, SourceBytes(`{"region":"eu","extra":true}`)),
		unknownFieldCase("yaml unknown field", FormatYAML, SourceBytes("region: eu\nextra: 1\n")),
		unknownFieldCase("toml unknown field", FormatTOML, SourceBytes("region = \"eu\"\nextra = 1\n")),
		{
			name:     "json type mismatch",
			format:   FormatJSON,
			body:     SourceBytes(`{"retries": "not-a-number"}`),
			contains: []userSpan{typeMismatchSpan, typeMismatchKeySpan, expectedNumberFound, mismatchFixRetries},
			excludes: bannedVocab{"for a field of type"},
		},
		{
			name:     "yaml type mismatch",
			format:   FormatYAML,
			body:     SourceBytes("retries: not-a-number\n"),
			contains: []userSpan{typeMismatchSpan, typeMismatchKeySpan, expectedNumberFound, mismatchFixRetries},
			excludes: bannedVocab{"!!str", "unmarshal errors", "cannot unmarshal"},
		},
		{
			name:   "toml type mismatch",
			format: FormatTOML,
			body:   SourceBytes("retries = \"abc\"\n"),
			contains: []userSpan{
				typeMismatchSpan,
				typeMismatchKeySpan,
				expectedNumberFound,
				mismatchFixRetries,
			},
			excludes: bannedVocab{"destination has type", "incompatible types", banDecoderReported},
		},
	}
}

// goldenStructuralCases pins the syntax and shape refusal shapes.
func goldenStructuralCases() []proseCase {
	return []proseCase{
		{
			name:     "json syntax",
			format:   FormatJSON,
			body:     SourceBytes("{\n  \"region\": \"eu\"\n"),
			contains: []userSpan{"JSON syntax", syntaxFixJSON},
			excludes: bannedVocab{"JSON decode", banDecoderReported},
		},
		{
			name:   "toml syntax",
			format: FormatTOML,
			body:   SourceBytes("retries = zzq-secret-marker-8842\n"),
			contains: []userSpan{
				"TOML syntax",
				syntaxFixTOML,
			},
			excludes: bannedVocab{"TOML parse:", banDecoderReported, "zzq", "SE-CRET"},
		},
		{
			name:     "yaml scanner fault",
			format:   FormatYAML,
			body:     SourceBytes("region: eu\nretries: @x`broken\n"),
			contains: []userSpan{"YAML syntax", syntaxFixYAML},
			excludes: bannedVocab{"YAML decode", banDecoderReported},
		},
		{
			name:   "yaml not a mapping",
			format: FormatYAML,
			body:   SourceBytes("- a\n- b\n"),
			contains: []userSpan{
				"not a mapping",
				"hold key: value pairs at the top of your config file",
			},
			excludes: bannedVocab{"document root", "layered defaults"},
		},
		{
			name:   "yaml multi doc",
			format: FormatYAML,
			body:   SourceBytes("region: eu\n---\nregion: us\n"),
			contains: []userSpan{
				"multi-document stream",
				"hold one document per file",
			},
			excludes: bannedVocab{"first-wins"},
		},
		{
			name:   "json duplicate key",
			format: FormatJSON,
			body:   SourceBytes("{\n  \"region\": \"eu\",\n  \"region\": \"us\"\n}\n"),
			contains: []userSpan{
				`duplicate key: "region"`,
				`remove the second "region"; the last one would win silently`,
			},
			excludes: bannedVocab{"last-win:", "re-declared in the same object"},
		},
	}
}

// TestRefusalProseSpeaksToTheUser pins the owner's prose bar: every
// refusal names the user's objects, labels the fault once, and gives one
// fix — never goconf's internal vocabulary.
func TestRefusalProseSpeaksToTheUser(t *testing.T) {
	for i := range goldenRefusalCases() {
		runProseCase(t, &goldenRefusalCases()[i])
	}
}

// runProseCase renders one case's refusal through Load and judges it.
func runProseCase(t *testing.T, tc *proseCase) {
	t.Helper()

	t.Run(string(tc.name), func(t *testing.T) {
		refusal := loadProseRefusal(t, tc)

		for _, want := range tc.contains {
			if !strings.Contains(string(refusal), string(want)) {
				t.Fatalf("refusal %q must contain %q", refusal, want)
			}
		}

		for _, ban := range tc.excludes {
			if strings.Contains(string(refusal), ban) {
				t.Fatalf("refusal %q must not contain %q", refusal, ban)
			}
		}
	})
}

// loadProseRefusal loads one case's file and requires a refusal.
func loadProseRefusal(t *testing.T, tc *proseCase) refusalText {
	t.Helper()

	root := writeTree(t, map[searchDir]map[FileName]SourceBytes{
		"": {"app" + FileName(string(tc.format)): tc.body},
	})
	t.Chdir(string(root))

	_, _, err := Load[loadCfg](Spec{Name: configName("app"), Format: tc.format},
		WithSearchPaths(DirPath(string(root))))
	if err == nil {
		t.Fatalf("expected a refusal")
	}

	return refusalText(err.Error())
}
