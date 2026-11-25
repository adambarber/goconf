package goconf

import (
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/adambarber/gorefuse"
)

// tomlUndecoded refuses keys present in the file but missing from the
// target: strictness matches JSON and YAML. Keys behind a Value are validated
// when that Value is materialized.
func tomlUndecoded(dst Target, data SourceBytes, md *toml.MetaData, file FilePath, absorbed valueKeySet) error {
	undec := []toml.Key{}

	for _, k := range md.Undecoded() {
		if !absorbed.absorbs(k) {
			undec = append(undec, k)
		}
	}

	if len(undec) == 0 {
		return nil
	}

	if !hasExtrasInType(dst.Value.Type()) {
		return refuseAt(ErrUnknownField, file, 0,
			gorefuse.Subject(string(joinKeys(undec))),
			gorefuse.Why("remove it from your config file, or the app must declare the field in its schema."))
	}

	return tomlRouteExtras(dst, data, undec, md, file)
}

// keyQuote wraps one key in quotes for a refusal message.
func keyQuote(k toml.Key) keyRender {
	return keyRender(`"` + k.String() + `"`)
}

// keyRender is one rendered key inside a refusal message.
type keyRender string

// joinKeys renders undecoded keys as a comma-separated list.
func joinKeys(keys []toml.Key) keyRender {
	parts := make([]keyRender, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, keyQuote(k))
	}

	return keyRender(strings.Join(toStrings(parts), ", "))
}

// toStrings adapts rendered keys to strings.Join.
func toStrings(parts []keyRender) []string {
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		out = append(out, string(p))
	}

	return out
}
