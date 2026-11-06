package goconf

import (
	"errors"
	"fmt"
	"reflect"
)

// errUnmarshalable is the refusal for a default JSON cannot represent.
var errUnmarshalable = errors.New("default cannot marshal")

// Spec declares one config load: the config name and the required format.
type Spec struct {
	Format Format
	Name   configName
}

// Option customizes one Load call.
type Option interface {
	apply(s *settings)
}

// settings is the mutable configuration of one Load call.
type settings struct {
	search    []DirPath
	globalDir DirPath
	defaults  []LayeredDefault
}

// LayeredDefault is one defaults layer whose filled fields are exactly
// the fields absent from every source.
type LayeredDefault struct {
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
	s.defaults = append(s.defaults, LayeredDefault{value: o.value})
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

// WithDefaults appends one defaults layer. The value must be marshalable
// by the spec's format: a broken default fails loud here, at option
// construction, and is never silently stored.
func WithDefaults[T any](value T) *defaultsOption {
	v := reflect.ValueOf(value)
	if err := marshalable(v); err != nil {
		panic(err)
	}

	return &defaultsOption{value: v}
}

// marshalable refuses a default whose value JSON cannot represent: a
// channel or func field would be silently dropped at decode time.
func marshalable(v reflect.Value) error {
	return walkMarshalable(v, "")
}

// walkMarshalable refuses channel and func fields anywhere in the value.
func walkMarshalable(v reflect.Value, path FieldPath) error {
	if v.Kind() == reflect.Chan || v.Kind() == reflect.Func {
		return fmt.Errorf("WithDefaults: %w", errUnmarshalable)
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
