package goconf

import (
	"reflect"

	"github.com/adambarber/gorefuse"
)

// envExpected states the accepted shape in user terms, never in Go type names.
func envExpected(k reflect.Kind) gorefuse.Why {
	if k == reflect.Bool {
		return gorefuse.Why("true or false")
	}

	if k >= reflect.Int && k <= reflect.Int64 {
		return gorefuse.Why("a whole number")
	}

	if k >= reflect.Uint && k <= reflect.Uintptr {
		return gorefuse.Why("a non-negative whole number")
	}

	if k == reflect.Float32 || k == reflect.Float64 {
		return gorefuse.Why("a number")
	}

	return envExpectedOther(k)
}

func envExpectedOther(k reflect.Kind) gorefuse.Why {
	if k == reflect.String {
		return gorefuse.Why("text")
	}

	if k == reflect.Slice || k == reflect.Array {
		return gorefuse.Why("JSON for a list")
	}

	if k == reflect.Map {
		return gorefuse.Why("JSON for a map")
	}

	if k == reflect.Struct {
		return gorefuse.Why("JSON for an object")
	}

	return gorefuse.Why("a supported value")
}

func envParseRefusal(name EnvName, t reflect.Type) error {
	return refuseAt(ErrEnvParse, FilePath(name), 1,
		gorefuse.Subject(string(name)),
		gorefuse.Why("set it to "+string(envExpected(t.Kind()))+"."))
}

func envAmbiguousRefusal(name EnvName, other, path FieldPath) error {
	return refuseAt(ErrEnvAmbiguous, FilePath(name), 1,
		gorefuse.Subject(string(name)),
		gorefuse.Why("rename one of the fields. "+string(other)+" and "+string(path)+" share this name."))
}
