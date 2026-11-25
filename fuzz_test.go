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
	seeds := fuzzGoldenSeeds()
	seeds = append(seeds,
		// duplicate keys in all three formats
		`{"region":"eu","region":"us"}`,
		"region: eu\nregion: us\n",
		"region = \"eu\"\nregion = \"us\"\n",
		// secret-shaped values: quoted, backquoted, multi-line
		`{"region":"`+fuzzSentinel+`"}`,
		"region: `"+fuzzSentinel+"`\n",
		"region = \""+fuzzSentinel+"\"\n",
		"region: |\n  "+fuzzSentinel+"\n  second line\n",
		// the marker as a key: the refusal must name it, not echo it
		fuzzSentinel+": 1\n",
		`{"`+fuzzSentinel+`":1}`,
		fuzzSentinel+" = 1\n",
		// multi-doc YAML, deep nesting, trailing data, escaped-quote keys
		"---\nregion: eu\n---\nregion: us\n",
		deep,
		`{"region":"eu"}{"region":"us"}`,
		`{"reg\"ion"="x"}`,
		// a marker used only as an array value: a false duplicate-key
		// refusal would echo it, so the twin oracle must catch that
		`{"tags":["`+fuzzSentinel+`","b","`+fuzzSentinel+`"]}`,
		// invalid UTF-8, null bytes, empty, huge scalar
		"\xff\xfe\xfa",
	)
	seeds = append(seeds, fuzzEdgeSeeds()...)
	seeds = append(seeds, "region = "+strings.Repeat("9", 4096)+"\n")

	addFuzzSeeds(f, seeds)

	f.Fuzz(func(t *testing.T, data []byte) {
		for _, format := range fuzzFormats() {
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

		refusal := fuzzRefusal(err)
		if refusal == "" {
			return ""
		}

		// the frame names the temp dir; it differs per run by design.
		_, tail, cut := strings.Cut(string(refusal), "/app"+string(format))
		if !cut {
			return refusal
		}

		return refusalText(tail)
	}

	assertFuzzRefusalOracle(t, data, errOf)
}

// stripQuotedFuzz drops every single-quoted span — a decoder's
// character-class echo and quoted key names — from one rendered
// refusal before the oracle compares the two sides.
func stripQuotedFuzz(msg refusalText) refusalText {
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
