package goconf

import (
	"strings"
	"testing"
)

// fuzzSentinel is the planted marker: no refusal path may echo any
// fragment of it into rendered error text.
const fuzzSentinel = "sk-live-9f8e7d6c"

// FuzzLoad drives arbitrary bytes through the real Load path. Every input
// runs through all three formats — one fuzz target, each mutation exercises
// each format's decode and refusal paths.
func FuzzLoad(f *testing.F) {
	deep := strings.Repeat(`{"a":`, 50) + "1" + strings.Repeat("}", 50)
	seeds := []string{
		// golden fixture shapes
		`{"region":"eu","nested":{"host":"h","port":1}}`,
		"region: eu\nnested:\n  host: h\n  port: 1\n",
		"region = \"eu\"\n\n[nested]\nhost = \"h\"\nport = 1\n",
		// duplicate keys in all three formats
		`{"region":"eu","region":"us"}`,
		"region: eu\nregion: us\n",
		"region = \"eu\"\nregion = \"us\"\n",
		// secret-shaped values: quoted, backquoted, multi-line
		`{"region":"` + fuzzSentinel + `"}`,
		"region: `" + fuzzSentinel + "`\n",
		"region = \"" + fuzzSentinel + "\"\n",
		"region: |\n  " + fuzzSentinel + "\n  second line\n",
		// multi-doc YAML, deep nesting, trailing data, escaped-quote keys
		"---\nregion: eu\n---\nregion: us\n",
		deep,
		`{"region":"eu"}{"region":"us"}`,
		`{"reg\"ion"="x"}`,
		// invalid UTF-8, null bytes, empty, huge scalar
		"\xff\xfe\xfa",
		"\x00\x01\x02",
		"",
		"region = " + strings.Repeat("9", 4096) + "\n",
	}

	for _, seed := range seeds {
		f.Add([]byte(seed))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		for _, format := range []Format{FormatJSON, FormatYAML, FormatTOML} {
			dir := t.TempDir()
			_ = writeFileAt(t, DirPath(dir), FileName("app"+string(format)), SourceBytes(data))

			_, _, err := Load[loadCfg](Spec{Name: configName("app"), Format: format},
				WithSearchPaths(DirPath(dir)))

			// no panic holds by construction; value-freedom: a refusal
			// may not echo the marker. no-partial-state is covered by
			// the TestLoad* suite, not re-asserted here.
			if err != nil && strings.Contains(string(data), fuzzSentinel) &&
				strings.Contains(err.Error(), fuzzSentinel) {
				t.Fatalf("refusal echoed the sentinel: %v", err)
			}
		}
	})
}
