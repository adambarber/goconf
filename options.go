package goconf

import (
	"reflect"

	"github.com/adambarber/gorefuse"
)

// Name is the caller-declared name of a config: files are searched as
// NAME.EXT in each directory.
type Name = configName

// Spec declares one config load: the config name and the required format.
type Spec struct {
	Format Format
	Name   Name
}

// Option customizes one Load call.
type Option interface {
	apply(s *settings)
}

// valueOption is the one Option implementation: it stores a typed setting
// and the setter that folds it into a load's settings.
type valueOption[T any] struct {
	value T
	set   func(*settings, T)
}

func (o valueOption[T]) apply(s *settings) { o.set(s, o.value) }

func newValueOption[T any](value T, set func(*settings, T)) *valueOption[T] {
	return &valueOption[T]{value: value, set: set}
}

// settings is the mutable configuration of one Load call.
type settings struct {
	search    []DirPath
	defaults  []layeredDefault
	globalDir DirPath
	maxBytes  byteCap
	optional  bool
	envPrefix EnvPrefix
}

// layeredDefault is one defaults layer whose filled fields are exactly
// the fields absent from every source. Unexported: it is built only by
// WithDefaults and carries no constructable public surface.
type layeredDefault struct {
	value reflect.Value
}

// searchPathsOption overrides the default search path list.
type searchPathsOption struct {
	paths []DirPath
}

func (o searchPathsOption) apply(s *settings) {
	s.search = o.paths
}

// globalDirOption overrides the user config directory for one load.
type globalDirOption struct {
	dir DirPath
}

func (o globalDirOption) apply(s *settings) {
	s.globalDir = o.dir
}

// defaultsOption appends one defaults layer.
type defaultsOption struct {
	value reflect.Value
}

func (o defaultsOption) apply(s *settings) {
	s.defaults = append(s.defaults, layeredDefault{value: o.value})
}

// WithEnv overlays process variables named PREFIX_FIELD over the decoded
// config. Dots in a field path become underscores and the path is upper-cased;
// env values are the outermost layer and win over file values and defaults.
func WithEnv(prefix EnvPrefix) *valueOption[EnvPrefix] {
	if prefix == "" {
		panic("goconf: WithEnv prefix must not be empty")
	}

	return newValueOption(prefix, func(s *settings, p EnvPrefix) { s.envPrefix = p })
}

// optionalConfigOption forgives true absence of config for one load.
type optionalConfigOption struct{}

func (o optionalConfigOption) apply(s *settings) {
	s.optional = true
}

// WithOptionalConfig forgives true absence of config: every search
// directory readable and zero candidates found. Load then behaves as if
// zero keys were present; a fault is not absence.
func WithOptionalConfig() *optionalConfigOption {
	return &optionalConfigOption{}
}

// WithSearchPaths replaces the default discovery paths (user config
// directory then the working directory) with the given list.
func WithSearchPaths(paths ...DirPath) *searchPathsOption {
	return &searchPathsOption{paths: paths}
}

// WithGlobalDir replaces the user config directory the default discovery
// starts from; the working directory still follows as the local layer.
func WithGlobalDir(dir DirPath) *globalDirOption {
	return &globalDirOption{dir: dir}
}

// WithDefaults appends a struct-shaped defaults layer (a struct or pointer to
// one) and PANICS on a broken value — a Must-style contract. The stored value
// is a deep copy, so later caller mutations cannot alter a load.
func WithDefaults[T any](value T) *defaultsOption {
	v := reflect.ValueOf(value)
	if err := marshalable(v); err != nil {
		panic(err)
	}

	structValue, err := checkDefaultShape(v)
	if err != nil {
		panic(err)
	}

	return &defaultsOption{value: cloneValue(structValue)}
}

// checkDefaultShape refuses a default value that is not struct-shaped,
// returning the struct its fields will be matched from.
func checkDefaultShape(v reflect.Value) (reflect.Value, error) {
	return normalizeDefault(v)
}

// marshalable refuses a default whose value JSON cannot represent: a
// channel or func field would be silently dropped at decode time.
func marshalable(v reflect.Value) error {
	return walkMarshalable(v, "")
}

// walkMarshalable refuses channel and func fields anywhere in the value.
func walkMarshalable(v reflect.Value, path FieldPath) error {
	if v.Kind() == reflect.Chan || v.Kind() == reflect.Func {
		return refuseWhy(ErrUnmarshalable,
			gorefuse.Subject("a default with a "+v.Kind().String()+" field"),
			gorefuse.Why("a default the format cannot represent would be silently dropped at decode time."))
	}

	if v.Kind() == reflect.Struct {
		return walkStruct(v, path)
	}
	// every other kind marshals faithfully

	return nil
}

// walkStruct walks each field of a struct value.
func walkStruct(v reflect.Value, prefix FieldPath) error {
	t := v.Type()
	for i := range t.NumField() {
		p := childPath(prefix, FieldPath(t.Field(i).Name))
		if err := walkMarshalable(v.Field(i), p); err != nil {
			return err
		}
	}

	return nil
}
