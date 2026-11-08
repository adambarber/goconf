package goconf

import (
	"strings"
	"testing"
)

// fuzzSentinel is the planted marker and fuzzTwin its stand-in; the
// oracle is differential (docs/load.md, "Fuzzing").
const (
	fuzzSentinel = "sk-live-9f8e7d6c"
	fuzzTwin     = "zz-neutral-00"
)

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
		// the marker as a key: the refusal must name it, not echo it
		fuzzSentinel + ": 1\n",
		`{"` + fuzzSentinel + `":1}`,
		fuzzSentinel + " = 1\n",
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
			runFuzzLeg(t, data, format)
		}
	})
}

// runFuzzLeg loads data through one format and twins it; the oracle is
// differential (see the docs/load.md fuzzing note).
func runFuzzLeg(t *testing.T, data []byte, format Format) {
	t.Helper()

	errOf := func(bytes []byte) refusalText {
		t.Helper()

		dir := t.TempDir()
		_ = writeFileAt(t, DirPath(dir), FileName("app"+string(format)), SourceBytes(bytes))

		_, _, err := Load[loadCfg](Spec{Name: configName("app"), Format: format},
			WithSearchPaths(DirPath(dir)))
		if err == nil {
			return ""
		}

		// the frame names the temp dir; it differs per run by design.
		_, tail, cut := strings.Cut(err.Error(), "/app"+string(format))
		if !cut {
			return refusalText(err.Error())
		}

		return refusalText(tail)
	}

	refusal := errOf(data)

	if refusal == "" {
		return
	}

	twinRefusal := errOf([]byte(strings.ReplaceAll(string(data), fuzzSentinel, fuzzTwin)))

	// twinRefusal for a no-marker input is the same bytes: the oracle
	// stands down; a decoder's single-character class echo ("'s'") is
	// dropped from both sides before comparing.
	if !strings.Contains(string(data), fuzzSentinel) {
		return
	}

	got := stripSingleCharSpans(refusal)

	want := stripSingleCharSpans(twinRefusal)

	if strings.ReplaceAll(string(got), fuzzSentinel, fuzzTwin) != string(want) {
		t.Fatalf("refusal is not substitution-stable (value echo):\n got: %s\ntwin: %s", refusal, twinRefusal)
	}
}

// stripSingleCharSpans drops every one-character single-quoted span
// ("'s'"), a decoder's character-class echo, from one rendered refusal.
func stripSingleCharSpans(msg refusalText) refusalText {
	text := string(msg)

	for {
		open := strings.IndexByte(text, '\'')
		if open < 0 || open+2 >= len(text) {
			return refusalText(text)
		}

		if text[open+2] != '\'' {
			// keep this quote; continue after it
			next := strings.IndexByte(text[open+1:], '\'')
			if next < 0 {
				return refusalText(text)
			}

			text = text[open+1+next+1:]

			continue
		}

		text = text[:open] + text[open+3:]
	}
}
