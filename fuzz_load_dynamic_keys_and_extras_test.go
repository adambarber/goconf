package goconf

import "testing"

type fuzzExtrasInner struct {
	Extra map[extraKey]extraVal `goconf:"extras" json:"-"    toml:"-"    yaml:"-"`
	Name  resName               `json:"name"     toml:"name" yaml:"name"`
}

type fuzzExtrasCfg struct {
	Extra    map[extraKey]extraVal `goconf:"extras"  json:"-"        toml:"-"        yaml:"-"`
	Dynamic  map[dynKey]nestedCfg  `goconf:"dynamic" json:"dynamic"  toml:"dynamic"  yaml:"dynamic"`
	Resource fuzzExtrasInner       `json:"resource"  toml:"resource" yaml:"resource"`
}

func FuzzLoadDynamicKeysAndExtras(f *testing.F) {
	seeds := fuzzSeeds{
		`{"resource":{"name":"r","x":"y"},"dynamic":{"k":{"host":"h","port":1}},"top":"v"}`,
		"resource:\n  name: r\n  x: y\ndynamic:\n  k:\n    host: h\n    port: 1\ntop: v\n",
		"[resource]\nname = \"r\"\nx = \"y\"\n[dynamic.k]\nhost = \"h\"\nport = 1\ntop = \"v\"\n",
		`{"resource":{"name":"r"},"dynamic":{},"top":"` + fuzzSentinel + `"}`,
		"resource:\n  name: r\n  " + fuzzSentinel + ": 1\ndynamic: {}\n",
		"[resource]\nname = \"r\"\n\"" + fuzzSentinel + "\" = 1\n[dynamic.k]\nhost = \"h\"\nport = 1\n",
		`{"resource":{"name":1},"dynamic":{"k":{"host":2,"port":"bad"}}}`,
		"resource:\n  name: [1]\ndynamic:\n  k:\n    host: h\n    port: [2]\n",
		`[1,2,3]`,
		"- a\n- b\n",
		`{"resource":{"name":"r","name":"dup"}}`,
		"resource:\n  name: r\n  name: r\n",
	}
	addFuzzSeeds(f, seeds)

	f.Fuzz(func(t *testing.T, data []byte) {
		for _, format := range fuzzFormats() {
			runFuzzDynamicExtrasLeg(t, data, format)
		}
	})
}

func runFuzzDynamicExtrasLeg(t *testing.T, data []byte, format Format) {
	t.Helper()

	errOf := func(bytes []byte) refusalText {
		t.Helper()

		_, _, err := LoadBytes[fuzzExtrasCfg](bytes, Spec{Name: configName("app"), Format: format})

		return fuzzRefusal(err)
	}

	assertFuzzRefusalOracle(t, data, errOf)
}
