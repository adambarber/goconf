package goconf

import "testing"

// fuzzSeeds is the corpus seed set, kept typed to satisfy no_untyped_value.
type fuzzSeeds []string

// fuzzFormats is the one format set every fuzz target drives.
func fuzzFormats() []Format {
	return []Format{FormatJSON, FormatYAML, FormatTOML}
}

// addFuzzSeeds registers the same seed set on every fuzz target.
func addFuzzSeeds(f *testing.F, seeds fuzzSeeds) {
	f.Helper()

	for _, seed := range seeds {
		f.Add([]byte(seed))
	}
}

// fuzzEdgeSeeds are the boundary seeds shared by the load fuzz targets.
func fuzzEdgeSeeds() fuzzSeeds {
	return fuzzSeeds{"\x00\x01", ""}
}

// fuzzGoldenSeeds are the shape seeds shared by the load fuzz targets.
func fuzzGoldenSeeds() fuzzSeeds {
	return fuzzSeeds{
		`{"region":"eu","nested":{"host":"h","port":1}}`,
		"region: eu\nnested:\n  host: h\n  port: 1\n",
		"region = \"eu\"\n\n[nested]\nhost = \"h\"\nport = 1\n",
	}
}
