package goconf

import (
	"reflect"

	"github.com/adambarber/gorefuse"
)

// checkValueContainerFormat refuses a Value field in a non-TOML source. The
// container is a TOML capability; falling back to JSON or YAML would decode
// only one of its two shapes and silently narrow the contract.
func checkValueContainerFormat(f Format, t reflect.Type, file FilePath) error {
	if f == FormatTOML || !hasValueType(t, map[reflect.Type]bool{}) {
		return nil
	}

	return refuseAt(ErrUnsupportedValueContainer, file, 0,
		gorefuse.Subject("a value container in a JSON or YAML config"),
		gorefuse.Why("declare the field as a plain value when this config is JSON or YAML."))
}
